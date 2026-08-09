# monitoring-service — Seguimiento, KPIs y alertas

> Estado: 🟡 En progreso | Última actualización: 2026-08-01
> Autor: Jesús Ariel González Bonilla (PM + Arquitecto) | Equipo: Arquitectura

> **Diseño previsto — no implementado.** Este documento describe el diseño acordado
> del servicio. La capa de aplicación aún no se construye ni se ha elegido su lenguaje;
> los contratos se especifican a nivel de protocolo (REST, eventos, esquema de BD).

## Responsabilidad

Consolida el seguimiento integral del proceso formativo: mide KPIs de las fichas en
ejecución, evalúa umbrales y genera alertas automáticas, gestiona sesiones de seguimiento
y planes de mejoramiento, y despacha notificaciones a los actores correspondientes.
Se activa principalmente por **eventos** de otros servicios; **no** llama síncronamente a
los servicios de dominio.

## Bounded context

Entidades propias del servicio. Ningún otro servicio escribe directamente en estas tablas.
Detalle completo en [data-model.md](./data-model.md).

| Entidad | Descripción |
|---------|-------------|
| `kpi_type` | Catálogo configurable de tipos de KPI medibles (ATTENDANCE, CURRICULUM_PROGRESS, DROPOUT_RISK, PRODUCTIVE_STAGE_PROGRESS) |
| `alert_type` | Catálogo configurable de tipos de alerta con umbral por defecto |
| `risk_level` | Catálogo de niveles de severidad (INFO → LOW → MEDIUM → HIGH → CRITICAL) |
| `kpi_status` | Catálogo de estados de KPI (ON_TRACK / AT_RISK / CRITICAL) |
| `ficha_tracking` | Estado de seguimiento consolidado de una ficha en ejecución |
| `kpi_tracking` | Medición puntual de un KPI para una ficha (serie temporal) |
| `tracking_session` | Sesión de seguimiento periódico registrada por instructor o coordinador |
| `generated_alert` | Alerta disparada cuando un KPI supera un umbral o se detecta una condición crítica |
| `improvement_plan` | Plan de mejoramiento a nivel de ficha generado a partir de una alerta |
| `sent_notification` | Registro append-only de notificaciones despachadas a los actores |

## Módulo de origen

M9 — Seguimiento y Analítica

## Dependencias

| Servicio | Tipo | Motivo |
|----------|------|--------|
| `iam-service` | auth | Validación del JWT en cada request de `monitoring-api` |
| `academic-management-service` | async | Consume `academic.ficha.opened` / `academic.ficha.closed` |
| `scheduling-service` | async | Consume `scheduling.schedule.published`, `scheduling.class_session.cancelled` |
| `actors-service` | async | Consume `actors.company_visit.completed` y afines |

> Las dependencias de dominio son **asíncronas** (bus de eventos AMQP, ADR-001). La única
> dependencia síncrona es `iam-service` para autenticación en la API de consulta.

## Componentes desplegables

| Componente | Sufijo | Descripción |
|------------|--------|-------------|
| [`monitoring-api`](./components/monitoring-api/README.md) | `-api` | REST de consulta sobre KPIs, alertas, seguimiento de fichas y planes de mejora |
| [`alert-worker`](./components/alert-worker/README.md) | `-worker` | Consume eventos de dominio, evalúa umbrales y genera `generated_alert` |
| [`notification-worker`](./components/notification-worker/README.md) | `-worker` | Consume alertas y despacha notificaciones salientes (EMAIL / IN_APP) |

## Base de datos

- Nombre lógico: `monitoring_db`
- Motor: **PostgreSQL 16**
- Esquema: ver [data-model.md](./data-model.md) — vigente 🟢 Estable
- Notas de modelado: `kpi_tracking` y `sent_notification` se particionan por RANGE mensual
  (serie temporal y retención); `active_alert_count` en `ficha_tracking` es un contador
  desnormalizado mantenido por triggers.

## Eventos

Ver [events.md](./events.md).

- **Publica** (topic `monitoring-events`): `monitoring.alert.triggered`,
  `monitoring.kpi.threshold_breached`, `monitoring.notification.sent`.
- **Consume**: `academic.ficha.opened`, `academic.ficha.closed`,
  `scheduling.schedule.published`, `scheduling.class_session.cancelled`,
  `actors.company_visit.completed`.

## Links

- Repo: (pendiente)
- Data model: [data-model.md](./data-model.md)
- Eventos: [events.md](./events.md)
- Runbook: [runbook.md](./runbook.md)
- Decisiones internas: [decisions.md](./decisions.md)
- Catálogo global de eventos: [event-catalog.md](../../event-catalog.md)
- ADR-001 (broker de mensajes): [ADR-001](../../../05-architecture/decisions/records/ADR-001-message-broker.md)
- ADR-004 (estados y auditoría): [ADR-004](../../../05-architecture/decisions/records/ADR-004-status-parametrization-and-audit-standard.md)
