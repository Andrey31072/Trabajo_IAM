# notification-worker

> Estado: 🟡 En progreso | Última actualización: 2026-08-01
> Autor: Jesús Ariel González Bonilla (PM + Arquitecto) | Equipo: Arquitectura

> **Diseño previsto — no implementado.** Agnóstico de lenguaje.

## Tipo de componente
`-worker` — consumidor de eventos / envío saliente.

## Responsabilidad
Consumir alertas y **enviar notificaciones** salientes (email / in-app), registrando el resultado en `sent_notification` con reintentos.

## Tecnologías
| Capa | Tecnología |
|------|-----------|
| Runtime | Agnóstico — cualquier backend (a definir) |
| Framework | A definir |
| Persistencia | PostgreSQL 16 · schema `monitoring` |
| Transporte | Broker AMQP (RabbitMQ, ADR-001) |
| Canal saliente | Proveedor de email / push (a definir) |

## Variables de entorno (genéricas)
| Variable | Descripción |
|----------|-------------|
| `DB_URL` | Conexión a PostgreSQL |
| `BROKER_URL` | Conexión al broker |
| `NOTIFICATION_PROVIDER_URL` | Endpoint del proveedor de envío |
| `MAX_RETRIES` | Reintentos antes de marcar fallo |

## Idempotencia
`event_id` único + estado en `sent_notification` (`PENDING/SENT/FAILED`) evitan doble envío. Índice parcial `WHERE send_status='PENDING'` para reproceso.

## Contrato
Ver [contract.md](./contract.md)
