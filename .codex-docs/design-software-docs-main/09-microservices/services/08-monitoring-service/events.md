# Eventos — monitoring-service

> Estado: 🟢 Estable | Última actualización: 2026-06-20

---

## Eventos publicados

Topic: `monitoring-events` | Consumidor conocido de los tres eventos: `audit-service`

---

### `monitoring.alert.triggered`

Publicado cuando el `alert-worker` persiste una nueva `generated_alert` cuyo
`risk_level` es LOW o superior. Se emite una vez por alerta; no se re-emite si la
alerta ya existe (idempotencia por `source_event_id`).

```json
{
  "event_id": "f3a1bc20-0d4e-4c2a-8e91-2b5d7f0a3e88",
  "event_type": "monitoring.alert.triggered",
  "version": "1.0",
  "timestamp": "2026-06-20T09:14:32Z",
  "source_service": "monitoring-service",
  "correlation_id": "a0e2d91c-11fc-4b78-b563-4f8c2e7d5a01",
  "payload": {
    "alert_id": "9c4e2b1a-3f7d-4a0e-b812-6d5c9f2e0a77",
    "ficha_tracking_id": "b2d3e4f5-6a7b-8c9d-0e1f-2a3b4c5d6e7f",
    "ficha_id": "1a2b3c4d-5e6f-7a8b-9c0d-e1f2a3b4c5d6",
    "alert_type_code": "LOW_ATTENDANCE",
    "risk_level_code": "HIGH",
    "triggered_value": 0.7200,
    "threshold_value": 0.8000,
    "affected_entity_type": "Learner",
    "affected_entity_id": "d4e5f6a7-b8c9-0d1e-2f3a-4b5c6d7e8f9a",
    "generated_at": "2026-06-20T09:14:31Z"
  }
}
```

| Campo | Tipo | Descripción |
|-------|------|-------------|
| `alert_id` | UUID | PK de `generated_alert` |
| `ficha_tracking_id` | UUID | FK a `ficha_tracking` donde se originó la alerta |
| `ficha_id` | UUID | Referencia externa a `academic-management-service` |
| `alert_type_code` | string | Código del catálogo `alert_type` (ej. `LOW_ATTENDANCE`, `HIGH_DROPOUT_RISK`) |
| `risk_level_code` | string | `INFO` / `LOW` / `MEDIUM` / `HIGH` / `CRITICAL` |
| `triggered_value` | decimal | Valor medido que superó el umbral |
| `threshold_value` | decimal | Umbral configurado en el catálogo al momento de la alerta |
| `affected_entity_type` | string | Tipo de entidad afectada (ej. `Learner`, `Instructor`, `Ficha`) |
| `affected_entity_id` | UUID | ID de la entidad afectada; null si aplica a la ficha completa |
| `generated_at` | timestamptz | Momento en que se creó la alerta |

---

### `monitoring.kpi.threshold_breached`

Publicado inmediatamente después de insertar un registro `kpi_tracking` cuyo
`kpi_status` resulta `AT_RISK` o `CRITICAL`. Permite a consumidores externos
reaccionar a cambios de KPI sin consultar la API del monitoring-service.

```json
{
  "event_id": "e7b2d4a1-9c3f-4e0b-a215-8f6d1c9e3b47",
  "event_type": "monitoring.kpi.threshold_breached",
  "version": "1.0",
  "timestamp": "2026-06-20T09:14:28Z",
  "source_service": "monitoring-service",
  "correlation_id": "a0e2d91c-11fc-4b78-b563-4f8c2e7d5a01",
  "payload": {
    "kpi_tracking_id": "c3d4e5f6-7a8b-9c0d-1e2f-3a4b5c6d7e8f",
    "ficha_tracking_id": "b2d3e4f5-6a7b-8c9d-0e1f-2a3b4c5d6e7f",
    "kpi_type_code": "ATTENDANCE",
    "current_value": 0.7200,
    "threshold_value": 0.8000,
    "period_start": "2026-06-01",
    "period_end": "2026-06-20",
    "measured_at": "2026-06-20T09:14:27Z"
  }
}
```

| Campo | Tipo | Descripción |
|-------|------|-------------|
| `kpi_tracking_id` | UUID | PK del registro `kpi_tracking` recién creado |
| `ficha_tracking_id` | UUID | FK a `ficha_tracking` |
| `kpi_type_code` | string | Código del catálogo `kpi_type` (ej. `ATTENDANCE`, `CURRICULUM_PROGRESS`) |
| `current_value` | decimal | Valor medido en este período |
| `threshold_value` | decimal | Umbral que fue superado |
| `period_start` | date | Inicio del período de medición |
| `period_end` | date | Fin del período de medición |
| `measured_at` | timestamptz | Timestamp de la medición |

---

### `monitoring.notification.sent`

Publicado por el `notification-worker` tras actualizar `sent_notification.send_status`
a `SENT`. No se publica si el envío queda en `FAILED`; el reintento lo gestiona la
política DLQ (ver sección Política de reintentos).

```json
{
  "event_id": "a1b2c3d4-e5f6-7a8b-9c0d-e1f2a3b4c5d6",
  "event_type": "monitoring.notification.sent",
  "version": "1.0",
  "timestamp": "2026-06-20T09:14:45Z",
  "source_service": "monitoring-service",
  "correlation_id": "a0e2d91c-11fc-4b78-b563-4f8c2e7d5a01",
  "payload": {
    "notification_id": "d5e6f7a8-b9c0-1d2e-3f4a-5b6c7d8e9f0a",
    "alert_id": "9c4e2b1a-3f7d-4a0e-b812-6d5c9f2e0a77",
    "recipient_id": "e6f7a8b9-c0d1-2e3f-4a5b-6c7d8e9f0a1b",
    "channel": "EMAIL",
    "subject": "Alerta de asistencia — Ficha 2345678",
    "send_status": "SENT",
    "sent_at": "2026-06-20T09:14:44Z"
  }
}
```

| Campo | Tipo | Descripción |
|-------|------|-------------|
| `notification_id` | UUID | PK de `sent_notification` |
| `alert_id` | UUID | FK a `generated_alert` que originó la notificación |
| `recipient_id` | UUID | ID del usuario en `iam-service` |
| `channel` | string | `EMAIL` o `IN_APP` |
| `subject` | string | Asunto de la notificación enviada |
| `send_status` | string | Siempre `SENT` en este evento |
| `sent_at` | timestamptz | Momento en que el canal confirmó la entrega |

---

## Eventos consumidos

| Evento | Servicio fuente | Acción en monitoring-service |
|--------|----------------|------------------------------|
| `academic.ficha.opened` | `academic-management-service` | Crea registro `ficha_tracking` con `overall_status = ON_TRACK` y `active_alert_count = 0`. Punto de entrada del ciclo de seguimiento de la ficha. |
| `academic.ficha.closed` | `academic-management-service` | Cierra el `ficha_tracking` activo: marca `last_tracking_date`, inhabilita nuevas mediciones de KPI y resuelve automáticamente las alertas pendientes que no requieren acción manual. |
| `scheduling.schedule.published` | `scheduling-service` | Actualiza `ficha_tracking.next_tracking_date` con la fecha de la próxima sesión programada. Permite saber cuándo se espera el siguiente corte de KPI sin polling al scheduling-service. |
| `scheduling.class_session.cancelled` | `scheduling-service` | Dispara una re-evaluación inmediata de KPIs de asistencia y avance curricular para la ficha afectada. Si la cancelación supera el umbral configurado en `alert_type` (`CURRICULUM_DELAY`), se genera una nueva `generated_alert`. |
| `actors.company_visit.completed` | `actors-service` | Actualiza el KPI `PRODUCTIVE_STAGE_PROGRESS` para el aprendiz afectado: inserta un registro en `kpi_tracking` con el nuevo porcentaje de avance y evalúa si se debe resolver la alerta `PRODUCTIVE_STAGE_DELAY` activa, si la hubiera. |

---

## Formato de envelope

Todos los eventos publicados por monitoring-service usan el envelope estándar del ecosistema:

```json
{
  "event_id": "uuid-v4",
  "event_type": "<servicio>.<entidad>.<accion>",
  "version": "1.0",
  "timestamp": "2026-01-01T00:00:00Z",
  "source_service": "monitoring-service",
  "correlation_id": "uuid-v4",
  "payload": {}
}
```

| Campo | Obligatorio | Descripción |
|-------|-------------|-------------|
| `event_id` | Sí | UUID v4 único por mensaje |
| `event_type` | Sí | Nombre del evento en formato `<servicio>.<entidad>.<accion>` |
| `version` | Sí | Versión del schema del payload; actualmente `"1.0"` |
| `timestamp` | Sí | ISO 8601 con zona horaria; momento de publicación |
| `source_service` | Sí | Siempre `"monitoring-service"` para eventos de este servicio |
| `correlation_id` | Sí | ID heredado del evento que originó la cadena; permite rastrear la trazabilidad entre servicios |
| `payload` | Sí | Cuerpo específico del evento; schema documentado en cada sección |

---

## Política de reintentos

**Dead Letter Queue:** `monitoring-events.dlq`

Los mensajes que no pueden procesarse tras agotar los reintentos se mueven a
`monitoring-events.dlq` para inspección manual y reinyección controlada.

| Parámetro | Valor |
|-----------|-------|
| Reintentos máximos | 3 |
| Backoff | Exponencial: 5 s → 25 s → 125 s |
| DLQ | `monitoring-events.dlq` |

**Idempotencia en el procesamiento de alertas**

El `alert-worker` identifica mensajes duplicados usando `source_event_id` (campo
`generated_alert.source_event_id`). Antes de crear una nueva alerta, verifica si ya
existe un registro en `generated_alert` con ese `source_event_id`. Si existe, descarta
el mensaje sin error. Esto garantiza que un reintento o reinyección desde la DLQ no
duplique alertas ni notificaciones.

El mismo principio aplica al `notification-worker`: antes de enviar, verifica si
`sent_notification` ya tiene un registro `SENT` asociado al `generated_alert_id`
procesado.

---

## Flujo de alerta

Describe el recorrido completo desde la detección de un KPI hasta la confirmación
de entrega de la notificación:

```
evento externo recibido
        │
        ▼
  monitoring-api evalúa KPI
  (inserta kpi_tracking)
        │
        ▼
  ¿valor supera threshold?
        │
      SÍ │
        ▼
  monitoring-api crea generated_alert
  (alert_type, risk_level, source_event_id)
        │
        ├─── publica monitoring.kpi.threshold_breached
        │
        ▼
  alert-worker lee generated_alert
  (consumer del bus interno)
        │
        ▼
  alert-worker publica monitoring.alert.triggered
        │
        ▼
  notification-worker recibe monitoring.alert.triggered
        │
        ▼
  notification-worker crea sent_notification (PENDING)
  y envía por canal (EMAIL / IN_APP)
        │
        ▼
  canal confirma entrega
  → sent_notification.send_status = SENT
        │
        ▼
  notification-worker publica monitoring.notification.sent
```

**Notas sobre el flujo:**

- El `monitoring-api` y el `alert-worker` son componentes internos del mismo servicio; se comunican a través de la base de datos (`generated_alert`), no por el bus de mensajes externo.
- `monitoring.kpi.threshold_breached` y `monitoring.alert.triggered` pueden publicarse casi en simultáneo para el mismo incidente; el orden no está garantizado entre ellos, pero cada consumidor solo necesita uno de los dos.
- Si el canal de notificación falla, `sent_notification.send_status` queda en `FAILED` y el mensaje se mueve a `monitoring-events.dlq`. No se publica `monitoring.notification.sent` hasta que el reintento resulte exitoso.
- El `alert-worker` verifica `source_event_id` antes de procesar para garantizar que una reinyección desde la DLQ no genere una alerta duplicada.
