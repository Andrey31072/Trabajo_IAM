# alert-worker

> Estado: 🟡 En progreso | Última actualización: 2026-08-01
> Autor: Jesús Ariel González Bonilla (PM + Arquitecto) | Equipo: Arquitectura

> **Diseño previsto — no implementado.** Agnóstico de lenguaje.

## Tipo de componente
`-worker` — consumidor de eventos.

## Responsabilidad
Evaluar KPIs/umbrales ante eventos de dominio y **generar alertas** (`generated_alert`) cuando se cruzan los límites definidos.

## Tecnologías
| Capa | Tecnología |
|------|-----------|
| Runtime | Agnóstico — cualquier backend (a definir) |
| Framework | A definir |
| Persistencia | PostgreSQL 16 · schema `monitoring` |
| Transporte | Broker AMQP (RabbitMQ, ADR-001) |

## Variables de entorno (genéricas)
| Variable | Descripción |
|----------|-------------|
| `DB_URL` | Conexión a PostgreSQL |
| `BROKER_URL` | Conexión al broker |
| `CONSUMER_GROUP` | Grupo del consumidor |

## Idempotencia
Deduplica por `event_id`; no genera alertas duplicadas para el mismo hecho+umbral. DLQ ante fallo persistente.

## Contrato
Ver [contract.md](./contract.md) · [data-model.md](../../data-model.md)
