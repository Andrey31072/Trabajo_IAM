<!-- RESUMEN-EJECUTIVO
agente: Jesús Ariel González Bonilla (PM + Arquitecto)
capacidad: contrato de evento / mensajería async
fase: diseño
estado: accepted
dependencias_entrada: 09-microservices/event-catalog.md, la BD de monitoring-service
consumidores_siguientes: notification-worker, audit-service (pendiente reconciliación de nombre de evento, ver halts)
tldr: Worker que evalúa seguimiento de ficha y umbrales de KPI, e inserta alertas generadas.
decisiones_clave: Conteo de alertas activas derivado (no columna redundante, ver normalization-assessment.md); idempotencia por event_id; DLQ alert-worker.dlq.
halts_registrados: academic.enrollment_ficha.status_changed, monitoring.kpi.tick y monitoring.alert.generated no están en event-catalog.md — el catálogo define `monitoring.alert.triggered`, no `monitoring.alert.generated` (pendiente reconciliación)
-->

# Contrato — alert-worker

> Estado: 🟡 En progreso | Última actualización: 2026-08-06
> Autor: Jesús Ariel González Bonilla (PM + Arquitecto) | Equipo: Arquitectura

> **Fuente de verdad:** el inventario de eventos lo gobierna [`event-catalog.md`](../../../../event-catalog.md) y el envelope estándar [`_template/service/events.md`](../../../../_template/service/events.md). Ante discrepancia, esos documentos prevalecen sobre este contrato.

> **Diseño previsto — no implementado.** Contrato de mensajería, agnóstico de lenguaje.

## Eventos consumidos
| Evento | Origen | Acción |
|--------|--------|--------|
| `academic.enrollment_ficha.status_changed` | academic | Recalcular seguimiento de ficha |
| `scheduling.schedule.published` | scheduling | Evaluar cobertura/KPIs de horario |
| `monitoring.kpi.tick` | scheduler | Evaluación periódica de umbrales |

## Eventos producidos
| Evento | Descripción |
|--------|-------------|
| `monitoring.alert.generated` | Se creó un `generated_alert` (nivel de riesgo, KPI, ficha) |

## Envelope (estándar)
```json
{
  "event_id": "uuid-v4",
  "event_type": "monitoring.alert.generated",
  "version": "1.0",
  "timestamp": "2026-08-01T14:00:00Z",
  "source_service": "monitoring-service",
  "correlation_id": "uuid-v4",
  "payload": { "alert_id": "uuid", "risk_level": "HIGH", "kpi": "coverage", "ficha_id": "uuid" }
}
```

## Efecto
Inserta `generated_alert` y actualiza `ficha_tracking`; el conteo de alertas activas se mantiene de forma consistente (derivado/trigger, no columna redundante — ver [normalization-assessment.md](../../../../../06-data/normalization-assessment.md)).

## Idempotencia
`event_id` único; misma condición evaluada no crea alertas duplicadas.

## Política de reintentos
Fallo persistente → reintento con backoff; agotados los reintentos → **DLQ** `alert-worker.dlq`.
