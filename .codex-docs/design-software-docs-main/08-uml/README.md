# UML

> Estado: 🟡 En progreso | Última actualización: 2026-08-01
> Autor: Jesús Ariel González Bonilla (PM + Arquitecto) | Equipo: Arquitectura

Repositorio de diagramas UML y arquitectura visual del proyecto **Horarios SENA**. Todo diagrama debe tener **fuente editable** (texto versionable) y **exportación revisable** (SVG).

> ⚠️ El sistema está en fase de datos: la mayoría de los diagramas de comportamiento (secuencia, componentes de la capa app) están **previstos** pero aún no dibujados, porque describen componentes que no se han construido. Los diagramas ER por servicio pueden derivarse ya del esquema real de cada repo `*-db`.

## Convenciones

- **Fuentes** en `diagrams/source/` con extensión `.wsd` o `.puml` (PlantUML). Diagramas simples pueden expresarse como bloques ```mermaid``` embebidos en Markdown (como en [dependencies.md](../15-project-control/dependencies.md)).
- **Exportaciones** en `diagrams/exports/` en formato `.svg` (preferido) o `.png`, generadas desde la fuente. Los SVG compartidos se guardan en `assets/` a nivel raíz cuando se referencian desde varios documentos.
- **Nombre de archivo**: `<dominio>-<tipo>.<ext>` (ej. `scheduling-sequence.wsd`, `iam-component.wsd`).
- Todo diagrama debe registrarse en [diagram-index.md](./diagram-index.md).
- La fuente es la **única verdad**: nunca editar un SVG a mano; regenerarlo desde la fuente.

## Tipos de diagrama

| Tipo | Archivo fuente | Herramienta |
|------|---------------|-------------|
| Contexto (C4 nivel 1) | `*-context.wsd` | PlantUML |
| Componentes (C4 nivel 3) | `*-component.wsd` | PlantUML |
| Casos de uso | `*-use-case.wsd` | PlantUML |
| Clases | `*-class.wsd` | PlantUML |
| Secuencia | `*-sequence.wsd` | PlantUML |
| Actividad | `*-activity.wsd` | PlantUML |
| Estado | `*-state.wsd` | PlantUML |
| Despliegue | `*-deployment.wsd` | PlantUML |
| Entidad-Relación (por servicio) | `*-er.wsd` | PlantUML / Mermaid |

## Índice de diagramas previstos

Estado: 🔴 No creado · 🟡 En curso · 🟢 Listo.

| Diagrama | Tipo | Alcance | Fuente prevista | Estado |
|----------|------|---------|-----------------|--------|
| Contexto del sistema | Contexto | Horarios SENA y actores externos (SENA, IdP futuro) | `system-context.wsd` | 🔴 |
| Panorama de microservicios | Componentes | Los 9 servicios y sus dependencias (deriva de [dependency-map.md](../09-microservices/dependency-map.md)) | `platform-component.wsd` | 🔴 |
| Autenticación (login + JWT) | Secuencia | Flujo de [authentication.md](../07-api/authentication.md) | `auth-login-sequence.wsd` | 🔴 |
| Creación de horario | Secuencia | scheduling ↔ actors / training-environment / academic-management | `scheduling-create-sequence.wsd` | 🔴 |
| Propagación de eventos | Secuencia | Publicación async hacia audit / monitoring (ver [ADR-001](../05-architecture/decisions/records/ADR-001-message-broker.md)) | `events-async-sequence.wsd` | 🔴 |
| Estados parametrizables | Estado | Ciclo de vida de estados según [ADR-004](../05-architecture/decisions/records/ADR-004-status-parametrization-and-audit-standard.md) | `status-lifecycle-state.wsd` | 🔴 |
| ER — iam-service | Entidad-Relación | Esquema `identity`/`rbac`/`session` | `iam-er.wsd` | 🔴 |
| ER — reference-data-service | Entidad-Relación | Catálogos y estructura institucional | `reference-data-er.wsd` | 🔴 |
| ER — academic-management-service | Entidad-Relación | Programas y fichas | `academic-management-er.wsd` | 🔴 |
| ER — training-environment-service | Entidad-Relación | Ambientes, inventario, disponibilidad | `training-environment-er.wsd` | 🔴 |
| ER — scheduling-service | Entidad-Relación | Horarios, sesiones, read models | `scheduling-er.wsd` | 🔴 |
| ER — actors-service | Entidad-Relación | Instructores, aprendices, empresas | `actors-er.wsd` | 🔴 |
| ER — document-service | Entidad-Relación | Documentos y `storage_key` (ver [ADR-003](../05-architecture/decisions/records/ADR-003-object-storage.md)) | `document-er.wsd` | 🔴 |
| ER — monitoring-service | Entidad-Relación | KPIs, alertas, seguimiento | `monitoring-er.wsd` | 🔴 |
| ER — audit-service | Entidad-Relación | Log append-only | `audit-er.wsd` | 🔴 |

> Prioridad de dibujo: primero los **ER por servicio** (derivables del esquema actual) y el **contexto/componentes**; los diagramas de secuencia se completan a medida que se construye cada componente de aplicación.

## Archivos

| Archivo | Descripción | Estado |
|---------|-------------|--------|
| [diagram-index.md](./diagram-index.md) | Índice registrado de fuentes y exportaciones | 🟡 |
| [diagrams/source/](./diagrams/source/) | Fuentes editables de diagramas | 🔴 |
| [diagrams/exports/](./diagrams/exports/) | Exportaciones SVG o PNG | 🔴 |
