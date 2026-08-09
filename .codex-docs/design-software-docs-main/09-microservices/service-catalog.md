# Catálogo de servicios

> Última actualización: 2026-08-01

> **Nota:** hoy cada servicio existe como su **repositorio de capa de datos** (`*-db`, Liquibase/PostgreSQL). Las capas de aplicación (API, worker, workflow) aún no están construidas. La columna Repo referencia el repo `*-db` real.

| # | Servicio | Descripción | Módulo(s) | Owner | Repo | Estado doc |
|---|----------|-------------|-----------|-------|------|------------|
| 01 | `iam-service` | Autenticación, autorización e identidad | M1 | Jesús Ariel González Bonilla | `design-software-iam-db` | 🟡 |
| 02 | `reference-data-service` | Jerarquía institucional y catálogos del sistema | M2 + M4 | Jesús Ariel González Bonilla | `design-software-reference-data-db` | 🟡 |
| 03 | `academic-management-service` | Programas de formación y fichas | M5 + M6 | Jesús Ariel González Bonilla | `design-software-academic-management-db` | 🟡 |
| 04 | `training-environment-service` | Ambientes físicos, inventario y disponibilidad | M3 | Jesús Ariel González Bonilla | `design-software-training-environment-db` | 🟡 |
| 05 | `scheduling-service` | Motor de asignación de horarios | M8 | Jesús Ariel González Bonilla | `design-software-scheduling-db` | 🟡 |
| 06 | `actors-service` | Instructores, aprendices y empresas | M7 | Jesús Ariel González Bonilla | `design-software-actors-db` | 🟡 |
| 07 | `document-service` | Gestión documental y generación de PDFs | transversal | Jesús Ariel González Bonilla | `design-software-document-db` | 🟡 |
| 08 | `monitoring-service` | KPIs, alertas y seguimiento integral | M9 | Jesús Ariel González Bonilla | `design-software-monitoring-db` | 🟡 |
| 09 | `audit-service` | Log de auditoría append-only | transversal | Jesús Ariel González Bonilla | `design-software-audit-db` | 🟡 |
