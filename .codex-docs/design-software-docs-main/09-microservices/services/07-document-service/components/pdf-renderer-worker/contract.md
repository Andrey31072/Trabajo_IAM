<!-- RESUMEN-EJECUTIVO
agente: Jesús Ariel González Bonilla (PM + Arquitecto)
capacidad: contrato de evento / mensajería async
fase: diseño
estado: accepted
dependencias_entrada: 09-microservices/event-catalog.md, la BD de document-service (document_version)
consumidores_siguientes: document-lifecycle-worker (pendiente reconciliación de eventos producidos, ver halts)
tldr: Worker que renderiza el PDF de una versión de documento, lo sube a object storage y publica el resultado.
decisiones_clave: storage_key determinista; evita re-render si el binario ya existe; DLQ pdf-renderer.dlq.
halts_registrados: document.render.requested, document.render.completed y document.render.failed no están en event-catalog.md (pendiente reconciliación)
-->

# Contrato — pdf-renderer-worker

> Estado: 🟡 En progreso | Última actualización: 2026-08-06
> Autor: Jesús Ariel González Bonilla (PM + Arquitecto) | Equipo: Arquitectura

> **Fuente de verdad:** el inventario de eventos lo gobierna [`event-catalog.md`](../../../../event-catalog.md) y el envelope estándar [`_template/service/events.md`](../../../../_template/service/events.md). Ante discrepancia, esos documentos prevalecen sobre este contrato.

> **Diseño previsto — no implementado.** Contrato de mensajería, agnóstico de lenguaje.

## Eventos consumidos
| Evento | Origen | Acción |
|--------|--------|--------|
| `document.render.requested` | document-api | Renderizar PDF de un `document_version` |

## Eventos producidos
| Evento | Descripción |
|--------|-------------|
| `document.render.completed` | PDF generado y almacenado; incluye la ruta en object storage |
| `document.render.failed` | Falló el render (tras reintentos) |

## Envelope (estándar)
```json
{
  "event_id": "uuid-v4",
  "event_type": "document.render.completed",
  "version": "1.0",
  "timestamp": "2026-08-01T14:00:00Z",
  "source_service": "document-service",
  "correlation_id": "uuid-v4",
  "payload": { "document_id": "uuid", "version_number": 2, "storage_key": "documents/uuid/v2.pdf" }
}
```

## Salida (efectos)
1. Sube el binario al bucket (`OBJECT_STORAGE_BUCKET`) con una `storage_key` determinista.
2. Actualiza los metadatos del `document_version` (ruta, tamaño, hash).
3. Publica `document.render.completed`.

## Idempotencia
`event_id` único; re-render evitado si ya existe el binario.

## Política de reintentos
Fallo persistente → reintento con backoff; agotados los reintentos → **DLQ** `pdf-renderer.dlq` + publicación de `document.render.failed`.
