# Microservicios

> Estado: 🟡 En progreso | Última actualización: 2026-08-01
> Autor: Jesús Ariel González Bonilla (PM + Arquitecto) | Equipo: Arquitectura

## Contenido

Centraliza el catálogo de microservicios, reglas de frontera, eventos y documentación por servicio.

## Alcance de la documentación por servicio

> **Estado real del sistema (alcance declarado):** hoy cada servicio existe únicamente como su **capa de datos** (repo `*-db`, Liquibase/PostgreSQL). Las **capas de aplicación** — componentes `*-api`, `*-worker`, `*-workflow` y sus contratos — **aún no están construidas como código**.
>
> Por eso, dentro de `services/<servicio>/`:
> - **`data-model.md` está completo (🟢)** y verificado 1:1 contra las tablas reales — es el entregable vigente.
> - Los documentos de componentes de aplicación (`components/*`, `contract.md`, etc.) permanecen en **🔴 por diseño**: son **plantillas a completar cuando se construya el servicio**, no deuda documental oculta.
>
> Ver [service-catalog.md](./service-catalog.md) para el mapeo servicio → repo y [service-readiness-checklist.md](./service-readiness-checklist.md) para los niveles de madurez (doc → dev → qa → prod).

## Archivos raíz

| Archivo | Descripción | Estado |
|---------|-------------|--------|
| [service-catalog.md](./service-catalog.md) | Inventario de servicios, owners, repos y estado documental | 🟡 |
| [communication-patterns.md](./communication-patterns.md) | Patrones síncronos, asíncronos y resiliencia | 🔴 |
| [dependency-map.md](./dependency-map.md) | Mapa de dependencias síncronas y asíncronas entre servicios | 🟡 |
| [data-ownership-matrix.md](./data-ownership-matrix.md) | Qué servicio posee cada entidad de dominio | 🟡 |
| [event-catalog.md](./event-catalog.md) | Catálogo centralizado de eventos por publicador | 🟡 |
| [service-boundary-rules.md](./service-boundary-rules.md) | Reglas de frontera, propiedad de datos y comunicación | 🟡 |
| [service-readiness-checklist.md](./service-readiness-checklist.md) | Criterios de madurez por nivel (doc → dev → qa → prod) | 🟡 |
| [storage-and-documents.md](./storage-and-documents.md) | Estrategia de almacenamiento: BDs, object storage y caché | 🟡 |

## Plantillas

| Carpeta | Usar para |
|---------|-----------|
| [_template/service/](./_template/service/) | Documentar un microservicio completo |
| [_template/component/](./_template/component/) | Documentar un componente desplegable (-api, -worker, -workflow) |

## Servicios

| # | Servicio | Módulo(s) | Componentes | Estado |
|---|----------|-----------|-------------|--------|
| 01 | [iam-service](./services/01-iam-service/) | M1 | `iam-api` | 🔴 |
| 02 | [reference-data-service](./services/02-reference-data-service/) | M2 + M4 | `reference-data-api` | 🔴 |
| 03 | [academic-management-service](./services/03-academic-management-service/) | M5 + M6 | `academic-management-api` | 🔴 |
| 04 | [training-environment-service](./services/04-training-environment-service/) | M3 | `training-environment-api` | 🔴 |
| 05 | [scheduling-service](./services/05-scheduling-service/) | M8 | `schedules-api`, `scheduling-engine-workflow`, `conflict-validator-worker` | 🔴 |
| 06 | [actors-service](./services/06-actors-service/) | M7 | `actors-api` | 🔴 |
| 07 | [document-service](./services/07-document-service/) | transversal | `document-api`, `template-api`, `pdf-renderer-worker`, `document-lifecycle-worker` | 🔴 |
| 08 | [monitoring-service](./services/08-monitoring-service/) | M9 | `monitoring-api`, `alert-worker`, `notification-worker` | 🔴 |
| 09 | [audit-service](./services/09-audit-service/) | transversal | `audit-worker` | 🔴 |
