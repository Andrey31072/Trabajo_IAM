# pdf-renderer-worker

> Estado: 🟡 En progreso | Última actualización: 2026-08-01
> Autor: Jesús Ariel González Bonilla (PM + Arquitecto) | Equipo: Arquitectura

> **Diseño previsto — no implementado.** Agnóstico de lenguaje.

## Tipo de componente
`-worker` — consumidor de eventos.

## Responsabilidad
Renderizar documentos a PDF a partir de una plantilla y sus datos, almacenar el binario en object storage y emitir el evento de finalización. **Los binarios nunca se guardan en la BD** (solo metadatos/ruta), ver [ADR-003](../../../../../05-architecture/decisions/records/ADR-003-object-storage.md).

## Tecnologías
| Capa | Tecnología |
|------|-----------|
| Runtime | Agnóstico — cualquier backend (a definir) |
| Framework | A definir |
| Persistencia | PostgreSQL 16 · schema `document` (metadatos) |
| Binarios | Object storage S3-compatible (MinIO DEV/QA, S3 PROD) |
| Transporte | Broker AMQP (RabbitMQ, ADR-001) |

## Variables de entorno (genéricas)
| Variable | Descripción |
|----------|-------------|
| `DB_URL` | Conexión a PostgreSQL |
| `BROKER_URL` | Conexión al broker |
| `OBJECT_STORAGE_ENDPOINT` | Endpoint S3-compatible |
| `OBJECT_STORAGE_BUCKET` | Bucket de documentos |

## Idempotencia
Deduplica por `event_id`; si el PDF ya fue generado para esa solicitud, no re-renderiza (verifica metadatos). DLQ ante fallo persistente.

## Contrato
Ver [contract.md](./contract.md) · [storage-adapters.md](../../storage-adapters.md)
