<!-- RESUMEN-EJECUTIVO
agente: Jesús Ariel González Bonilla (PM + Arquitecto)
capacidad: contrato de evento / mensajería async
fase: diseño
estado: accepted
dependencias_entrada: 09-microservices/event-catalog.md, la BD de monitoring-service
consumidores_siguientes: audit-service
tldr: Worker que resuelve destinatario/canal y envía la notificación derivada de una alerta generada.
decisiones_clave: Un envío ya SENT no se repite; contenido va al canal, no al log (privacidad); DLQ notification-worker.dlq.
halts_registrados: monitoring.alert.generated no está en event-catalog.md — el catálogo define `monitoring.alert.triggered` (pendiente reconciliación)
-->

# Contrato — notification-worker

> Estado: 🟡 En progreso | Última actualización: 2026-08-06
> Autor: Jesús Ariel González Bonilla (PM + Arquitecto) | Equipo: Arquitectura

> **Fuente de verdad:** el inventario de eventos lo gobierna [`event-catalog.md`](../../../../event-catalog.md) y el envelope estándar [`_template/service/events.md`](../../../../_template/service/events.md). Ante discrepancia, esos documentos prevalecen sobre este contrato.

> **Diseño previsto — no implementado.** Contrato de mensajería, agnóstico de lenguaje.

## Eventos consumidos
| Evento | Origen | Acción |
|--------|--------|--------|
| `monitoring.alert.generated` | alert-worker | Enviar notificación al destinatario |

## Eventos producidos (opcional)
| Evento | Descripción |
|--------|-------------|
| `monitoring.notification.sent` | Notificación entregada (para trazabilidad/auditoría) |

## Envelope (estándar)
```json
{
  "event_id": "uuid-v4",
  "event_type": "monitoring.notification.sent",
  "version": "1.0",
  "timestamp": "2026-08-01T14:00:00Z",
  "source_service": "monitoring-service",
  "correlation_id": "uuid-v4",
  "payload": { "notification_id": "uuid", "channel": "email", "status": "SENT" }
}
```

## Efecto
1. Resuelve destinatario y canal.
2. Crea `sent_notification` en estado `PENDING`.
3. Invoca el proveedor saliente; actualiza a `SENT` o `FAILED`.

## Idempotencia
`event_id` único; un envío ya `SENT` no se repite.

## Política de reintentos
Fallos → reintento con backoff; agotado → `FAILED` + **DLQ** `notification-worker.dlq`.

## Privacidad
No se registran datos personales en logs, solo identificadores; el contenido va al canal, no al log (ver [13-operations/observability.md](../../../../../13-operations/observability.md)).
