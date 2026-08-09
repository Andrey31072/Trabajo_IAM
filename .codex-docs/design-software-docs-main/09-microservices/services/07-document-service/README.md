# document-service — Gestión documental

> Estado: 🟡 En progreso | Última actualización: 2026-08-01
> Autor: Jesús Ariel González Bonilla (PM + Arquitecto) | Equipo: Arquitectura

> **Diseño previsto — no implementado.** La capa de aplicación aún no está construida
> ni se ha elegido lenguaje/framework. Este documento describe el diseño a nivel de
> contratos (REST/JSON, eventos) y modelo de datos; es válido para **cualquier backend**.

## Responsabilidad

Gestiona el ciclo de vida de documentos: generación asíncrona, versionado, almacenamiento
de binarios y exportación mediante URL firmada. Recibe solicitudes de generación de otros
servicios (horarios en PDF, comprobantes de matrícula, constancias, reportes) y es
**agnóstico al dominio de negocio**. Los binarios **nunca** viven en la base de datos: solo
se almacenan metadatos y la `storage_key` que referencia el archivo en object storage.

## Módulo de origen

Transversal — sin módulo de dominio directo. Cualquier servicio puede solicitar la
generación o consulta de documentos.

## Modelo de datos vigente

Entidades propias (naming en inglés, HALT-DB-NAMING). Ver detalle en [data-model.md](./data-model.md).

| Entidad | Descripción |
|---------|-------------|
| `document_template` | Plantilla reutilizable para generar documentos (cuerpo HTML, `output_type` PDF/EXCEL/WORD, versión). |
| `document` | Documento generado o cargado; metadatos + `storage_key`. Máquina de estados de negocio. |
| `document_version` | Historial inmutable de versiones de un documento. |

**Invariante clave:** los archivos binarios nunca se almacenan en `document_db`; solo
`storage_key` referencia el objeto en S3/MinIO (decisión 01 de [decisions.md](./decisions.md), [ADR-003](../../../05-architecture/decisions/records/ADR-003-object-storage.md)).

**Estados de `document`:** `GENERATING` → `AVAILABLE` → `ARCHIVED` / `EXPIRED`; ruta de
error `GENERATING` → `GENERATION_FAILED`. (Enum técnico cerrado `VARCHAR + CHECK`, no tipo
`ENUM` nativo; ver [data-model.md](./data-model.md).)

## Componentes desplegables

| Componente | Sufijo | Tipo | Descripción |
|------------|--------|------|-------------|
| [document-api](./components/document-api/README.md) | `-api` | REST síncrona | Consulta/versionado de documentos y solicitud de generación asíncrona; entrega URL firmada. |
| [template-api](./components/template-api/README.md) | `-api` | REST síncrona | CRUD de plantillas (`document_template`). |
| [pdf-renderer-worker](./components/pdf-renderer-worker/README.md) | `-worker` | Consumidor de cola | Renderiza el documento a partir de plantilla + datos y sube el binario a object storage. |
| [document-lifecycle-worker](./components/document-lifecycle-worker/README.md) | `-worker` | Consumidor de eventos / retención | Orquesta transiciones de estado, publica eventos de salida y aplica la política de retención. |

## Eventos

Detalle y payloads en [events.md](./events.md); catálogo global en [event-catalog.md](../../event-catalog.md).

**Publicados** (topic `document-events`):

| Evento | Cuándo |
|--------|--------|
| `document.document.generated` | El documento se generó y su binario está disponible. |
| `document.version.created` | Se creó una nueva versión de un documento existente. |

**Consumidos:**

| Evento | Origen | Efecto |
|--------|--------|--------|
| `scheduling.schedule.published` | `scheduling-service` | Genera el PDF de horario por entidad afectada. |
| `academic.ficha.opened` | `academic-service` | Genera el comprobante de matrícula del estudiante. |

**Cola interna:** `document-generation-queue` (solicitudes de render encoladas para el `pdf-renderer-worker`).

## Base de datos y almacenamiento

- **BD lógica:** `document_db` — solo metadatos de documentos.
- **Motor:** PostgreSQL 16.
- **Binarios:** object storage compatible S3 — MinIO en DEV/QA, S3 en PROD ([ADR-003](../../../05-architecture/decisions/records/ADR-003-object-storage.md)). Estrategia en [storage-adapters.md](./storage-adapters.md).
- **Broker de eventos:** AMQP — RabbitMQ ([ADR-001](../../../05-architecture/decisions/records/ADR-001-message-broker.md)).
- Detalle transversal en [storage-and-documents.md](../../storage-and-documents.md).

## Dependencias

| Servicio | Tipo | Motivo |
|----------|------|--------|
| `iam-service` | auth | Validación de JWT y scopes/roles. |
| `scheduling-service` | evento | Origen de `scheduling.schedule.published`. |
| `academic-service` | evento | Origen de `academic.ficha.opened`. |
| Object storage (S3/MinIO) | infraestructura | Persistencia de binarios ([ADR-003](../../../05-architecture/decisions/records/ADR-003-object-storage.md)). |
| Broker AMQP (RabbitMQ) | infraestructura | Transporte de eventos y cola de render ([ADR-001](../../../05-architecture/decisions/records/ADR-001-message-broker.md)). |
| `audit-service` | consumidor | Consume los eventos publicados por este servicio. |

## Tecnologías

| Capa | Tecnología |
|------|-----------|
| Runtime | Agnóstico — cualquier backend (a definir) |
| Framework | A definir |
| Base de datos | PostgreSQL 16 |
| Broker | AMQP — RabbitMQ ([ADR-001](../../../05-architecture/decisions/records/ADR-001-message-broker.md)) |
| Object storage | S3-compatible / MinIO ([ADR-003](../../../05-architecture/decisions/records/ADR-003-object-storage.md)) |

## Links

- Repo: (pendiente)
- Data model: [data-model.md](./data-model.md)
- Eventos: [events.md](./events.md)
- Runbook: [runbook.md](./runbook.md)
- Decisiones internas: [decisions.md](./decisions.md)
- Estrategia de almacenamiento: [storage-adapters.md](./storage-adapters.md)
