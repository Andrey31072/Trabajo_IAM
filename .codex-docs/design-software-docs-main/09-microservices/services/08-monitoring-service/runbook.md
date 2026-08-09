# Runbook — monitoring-service

> Estado: 🟢 Estable | Última actualización: 2026-06-20

---

## Healthcheck

El servicio expone tres componentes desplegables. El healthcheck primario es el endpoint HTTP de la API; los workers se verifican por heartbeat en la cola.

| Componente | Verificación | Respuesta esperada | SLO |
|------------|-------------|-------------------|-----|
| `monitoring-api` | `GET http://localhost:8008/health` | `200 { "status": "ok" }` | < 200 ms, 99% |
| `alert-worker` | Heartbeat en cola `monitoring-internal` | Latido cada 30 s | Lag < 5 min |
| `notification-worker` | Heartbeat en cola `monitoring-internal` | Latido cada 30 s | Lag < 5 min |

Verificación manual de la API:

```bash
curl -sf http://localhost:8008/health | jq .
```

Verificación de lag del `alert-worker` (reemplazar `<broker>` y `<consumer-group>`):

```bash
# Kafka
kafka-consumer-groups.sh --bootstrap-server <broker>:9092 \
  --group monitoring-alert-worker --describe

# RabbitMQ / AMQP — revisar profundidad de la cola interna
rabbitmqctl list_queues name messages consumers
```

---

## Variables de entorno requeridas

Las siguientes variables deben estar presentes en todos los componentes del servicio. Una variable faltante impide el arranque.

### Comunes a los tres componentes

| Variable | Descripción | Ejemplo |
|----------|-------------|---------|
| `DATABASE_URL` | Cadena de conexión a `monitoring_db` | `postgresql://user:pass@host:5432/monitoring_db` |
| `MESSAGE_BROKER_URL` | URL del broker de mensajes | `amqp://user:pass@broker:5672` |
| `LOG_LEVEL` | Nivel de log (`debug`, `info`, `warn`, `error`) | `info` |
| `SERVICE_ENV` | Entorno de ejecución (`production`, `staging`) | `production` |

### Solo `monitoring-api`

| Variable | Descripción | Ejemplo |
|----------|-------------|---------|
| `PORT` | Puerto de escucha HTTP | `8008` |
| `IAM_SERVICE_URL` | URL base del `iam-service` para validar JWT | `http://iam-service:8001` |
| `JWT_SECRET` | Clave para verificar tokens (o usar IAM introspection) | — |

### Solo `notification-worker`

| Variable | Descripción | Ejemplo |
|----------|-------------|---------|
| `EMAIL_SMTP_HOST` | Host del servidor SMTP | `smtp.example.com` |
| `EMAIL_SMTP_PORT` | Puerto SMTP | `587` |
| `EMAIL_SMTP_USER` | Usuario SMTP | `notifications@example.com` |
| `EMAIL_SMTP_PASSWORD` | Contraseña SMTP | — |
| `EMAIL_FROM_ADDRESS` | Dirección remitente | `no-reply@example.com` |
| `IN_APP_PUSH_ENDPOINT` | Endpoint del servicio de push in-app | `http://push-service:9000` |

---

## Alertas críticas

### P0 — Base de datos no responde

**Condición:** timeout de conexión a `monitoring_db` > 5 s, o connection pool exhausto.

**Impacto:** La `monitoring-api` devuelve 503. El `alert-worker` no puede persistir alertas; el `notification-worker` no puede actualizar `send_status`. Los eventos de entrada siguen siendo consumidos del broker y el lag se detiene en el punto de fallo, pero ningún dato nuevo se escribe.

**Acción inmediata:**
1. Verificar estado del pod/contenedor de PostgreSQL: `docker ps` o `kubectl get pods`.
2. Intentar conexión directa: `psql $DATABASE_URL -c "SELECT 1"`.
3. Revisar logs de la BD por errores de disco, OOM o lock contention.
4. Si la BD está sana y el problema es de red, reiniciar los pods del servicio para forzar reconexión del pool.
5. Escalar con DBA si la BD no responde en 10 min.

---

### P1 — Alert processing lag > 5 min (KPIs demorados)

**Condición:** el lag del consumer group del `alert-worker` sobre el topic interno supera 300 s, o la cola de alertas pendientes en `generated_alert` crece continuamente sin procesar.

**Impacto:** Las alertas se generan tarde. Los instructores y coordinadores no reciben notificaciones a tiempo. Los KPIs en pantalla quedan desactualizados.

**Acción inmediata:**
1. Verificar que el pod del `alert-worker` está corriendo: `docker ps | grep alert-worker` o `kubectl get pods`.
2. Revisar logs del worker por errores de procesamiento o bucles de reintento.
3. Si el worker está caído, reiniciarlo — procesa el backlog automáticamente al arrancar.
4. Si el worker está activo pero lento, verificar carga de la BD y latencia de escritura.
5. Si el backlog supera 1000 mensajes, escalar instancias del `alert-worker` (el worker es stateless).
6. Verificar que la DLQ `monitoring-events.dlq` no esté acumulando mensajes fallidos.

```bash
# Ver alertas pendientes sin procesar (generadas pero sin notificación enviada)
psql $DATABASE_URL -c "
  SELECT count(*) AS pending_alerts
  FROM generated_alert ga
  LEFT JOIN sent_notification sn ON sn.generated_alert_id = ga.id
  WHERE sn.id IS NULL
    AND ga.created_at < NOW() - INTERVAL '5 minutes';
"
```

---

### P1 — Notification failure rate alto

**Condición:** más del 10% de los registros en `sent_notification` tienen `send_status = 'FAILED'` en la última hora, o la DLQ `monitoring-events.dlq` tiene más de 50 mensajes.

**Impacto:** Los actores (instructores, coordinadores) no reciben alertas. Las alertas están generadas correctamente en BD pero el canal de entrega falla.

**Acción inmediata:**
1. Revisar logs del `notification-worker` por errores SMTP o de conexión al servicio de push.
2. Verificar credenciales SMTP y conectividad: `telnet $EMAIL_SMTP_HOST $EMAIL_SMTP_PORT`.
3. Consultar el conteo de fallos recientes:

```bash
psql $DATABASE_URL -c "
  SELECT send_status, count(*) AS total
  FROM sent_notification
  WHERE created_at > NOW() - INTERVAL '1 hour'
  GROUP BY send_status
  ORDER BY total DESC;
"
```

4. Si el problema es SMTP transitorio, los mensajes en DLQ se pueden reinyectar cuando el canal recupere (ver procedimiento de reenvío abajo).
5. Si el canal está permanentemente caído, escalar con el equipo de infraestructura.

---

### P2 — Cola de eventos creciendo (queue depth alto)

**Condición:** profundidad de la cola de entrada al `alert-worker` > 500 mensajes sostenida por más de 10 min.

**Impacto:** Degradación del SLO de tiempo de respuesta de alertas. No hay pérdida de datos; el broker retiene los mensajes.

**Acción inmediata:**
1. Escalar el número de réplicas del `alert-worker` (es stateless, escala horizontalmente).
2. Verificar que no haya un evento de alta carga puntual (ej. cierre masivo de fichas).
3. Monitorear que el lag empiece a bajar tras el escalado.

---

## Procedimientos comunes

### Resolver manualmente una alerta

Usado cuando una alerta fue generada por error de datos o la situación se resolvió fuera de banda y el sistema no la cerró automáticamente.

```sql
-- 1. Identificar la alerta
SELECT id, alert_type_id, risk_level_id, resolution_status, created_at
FROM generated_alert
WHERE ficha_tracking_id = '<ficha_tracking_uuid>'
  AND resolution_status = 'OPEN'
ORDER BY created_at DESC;

-- 2. Resolver la alerta manualmente (registrar quien resuelve en notas)
UPDATE generated_alert
SET resolution_status = 'MANUALLY_RESOLVED',
    resolved_at = NOW(),
    resolution_notes = 'Resuelto manualmente por <nombre-operador>: <motivo>'
WHERE id = '<alert_uuid>'
  AND resolution_status = 'OPEN';
```

> No modificar otros campos. El log de resolucion es auditado por `audit-service`.

---

### Recalcular KPIs para una ficha

Usado cuando se corrigen datos en servicios upstream (ej. asistencia recalculada en `scheduling-service`) y los KPIs de `monitoring-service` quedaron desactualizados.

**Opcion 1 — via API (preferida):**

```bash
# Forzar recalculo de todos los KPIs de una ficha
curl -X POST http://localhost:8008/api/v1/ficha-tracking/<ficha_tracking_id>/recalculate \
  -H "Authorization: Bearer $ADMIN_TOKEN" \
  -H "Content-Type: application/json"
```

**Opcion 2 — manual en BD (solo si la API no está disponible):**

```sql
-- Ver el estado actual de los KPIs de la ficha
SELECT kt.id, kpt.code AS kpi, kt.current_value, kt.kpi_status, kt.measured_at
FROM kpi_tracking kt
JOIN kpi_type kpt ON kpt.id = kt.kpi_type_id
WHERE kt.ficha_tracking_id = '<ficha_tracking_uuid>'
ORDER BY kt.measured_at DESC;

-- Insertar nuevo registro de KPI con valor corregido
-- NOTA: nunca modificar registros existentes; solo insertar nuevos
INSERT INTO kpi_tracking (id, ficha_tracking_id, kpi_type_id, current_value,
  threshold_value, kpi_status, period_start, period_end, measured_at)
VALUES (gen_random_uuid(), '<ficha_tracking_uuid>', '<kpi_type_id>',
  <nuevo_valor>, <threshold>, '<NORMAL|AT_RISK|CRITICAL>',
  '<fecha_inicio>', '<fecha_fin>', NOW());
```

---

### Reenviar una notificacion fallida

Usado cuando un envio quedo en estado `FAILED` y el canal de entrega ya esta disponible.

**Paso 1 — Identificar notificaciones fallidas:**

```sql
SELECT sn.id, sn.generated_alert_id, sn.recipient_id, sn.channel,
       sn.send_status, sn.created_at, sn.error_detail
FROM sent_notification sn
WHERE sn.send_status = 'FAILED'
  AND sn.created_at > NOW() - INTERVAL '24 hours'
ORDER BY sn.created_at DESC;
```

**Paso 2 — Reinyectar desde la DLQ (metodo recomendado):**

```bash
# Kafka — mover mensajes de la DLQ al topic original
kafka-console-consumer.sh \
  --bootstrap-server <broker>:9092 \
  --topic monitoring-events.dlq \
  --from-beginning \
  --max-messages 100 | \
kafka-console-producer.sh \
  --bootstrap-server <broker>:9092 \
  --topic monitoring-events
```

**Paso 3 — Verificar resultado:**

```sql
-- Confirmar que el estado cambi a SENT
SELECT send_status, count(*)
FROM sent_notification
WHERE id IN ('<notification_uuid_1>', '<notification_uuid_2>')
GROUP BY send_status;
```

> El `notification-worker` es idempotente: verifica si ya existe un registro `SENT` antes de enviar. Reinyectar no genera duplicados.

---

## Escenarios de falla

### Escenario 1 — alert-worker caido durante mantenimiento

**Sintoma:** lag de alertas creciendo; no se generan notificaciones; `monitoring-api` sigue respondiendo normalmente.

**Comportamiento esperado:** los eventos se acumulan en la cola del broker sin perderse. La API sigue sirviendo datos historicos de KPIs y alertas ya generadas.

**Recuperacion:**
1. Reiniciar el `alert-worker`.
2. El worker procesa el backlog automaticamente (FIFO, con idempotencia por `source_event_id`).
3. Monitorear que el lag baje a cero.
4. Verificar que la DLQ no acumulo mensajes durante la caida.

**Tiempo esperado de recuperacion (RTO):** < 10 min para backlog < 1000 eventos.

---

### Escenario 2 — notification-worker caido; SMTP no disponible

**Sintoma:** alertas generadas correctamente en BD pero `send_status` queda en `PENDING` o `FAILED`; la DLQ crece.

**Recuperacion:**
1. Resolver el problema del canal (SMTP, push service).
2. Reiniciar el `notification-worker`.
3. Reinyectar mensajes de la DLQ (ver procedimiento arriba).
4. Confirmar que `sent_notification` actualiza a `SENT`.

---

### Escenario 3 — KPIs inconsistentes tras correccion de datos upstream

**Sintoma:** los valores de KPI en `monitoring-service` no coinciden con los datos corregidos en `scheduling-service` u otro servicio upstream.

**Causa tipica:** el evento de correccion no fue publicado, o el `alert-worker` lo proceso antes de que la correccion se aplicara.

**Recuperacion:**
1. Identificar la ficha afectada y los KPIs incorrectos.
2. Usar el endpoint de recalculo (ver procedimiento arriba) o insertar nuevos registros de KPI.
3. Evaluar si las alertas generadas con datos incorrectos deben resolverse manualmente.

---

### Escenario 4 — Datos duplicados en generated_alert

**Sintoma:** multiples alertas con el mismo `source_event_id` en la tabla `generated_alert`.

**Causa:** fallo en la verificacion de idempotencia del `alert-worker` (bug de version o race condition).

**Investigacion:**

```sql
SELECT source_event_id, count(*) AS duplicados
FROM generated_alert
GROUP BY source_event_id
HAVING count(*) > 1;
```

**Accion:** escalar inmediatamente al equipo de desarrollo. No resolver manualmente sin analisis de causa raiz.

---

## Escalamiento

| Nivel | Condicion | Responsable | Accion |
|-------|-----------|-------------|--------|
| L1 — Operaciones | Cualquier alerta P1/P2 | On-call de turno | Ejecutar procedimientos de este runbook |
| L2 — Infraestructura | BD no responde, broker inestable | Equipo de infraestructura | Diagnóstico de capa de datos y red |
| L3 — Desarrollo | Bug confirmado (duplicados, idempotencia rota), comportamiento inesperado del worker | Equipo de desarrollo del servicio | Hotfix y deploy |
| L4 — Producto | Alertas incorrectas afectan decisiones pedagógicas | Product owner + equipo pedagógico | Revision de umbrales y tipos de alerta en catalogo |

**Contactos de escalamiento:** definidos en el directorio interno del equipo (pendiente de completar con nombres y canales).

**SLO del servicio:** 99% de disponibilidad mensual. Budget de error: 7.2 h/mes.

**Canal de incidentes:** canal de Slack/Teams definido por el equipo (pendiente de configurar).
