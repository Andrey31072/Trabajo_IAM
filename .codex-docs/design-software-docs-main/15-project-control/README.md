# Control de Proyecto

> Estado: 🟡 En progreso | Última actualización: 2026-08-01
> Autor: Jesús Ariel González Bonilla (PM + Arquitecto) | Equipo: Proyecto

## Contenido

Agrupa backlog técnico, riesgos, dependencias y preguntas abiertas del proyecto **Horarios SENA**. Es la vista de gestión: qué está pendiente, qué puede salir mal y de qué depende cada pieza. El detalle técnico vive en las carpetas de arquitectura (`05-architecture`), datos (`06-data`) y microservicios (`09-microservices`).

## Estado del proyecto

Hoy sólo existe la **capa de datos** (repos `*-db` con Liquibase + PostgreSQL 16). La capa de aplicación (API / worker / workflow) está pendiente de construir; ver plan por fases en [technical-backlog.md](./technical-backlog.md).

## Archivos

| Archivo | Descripción | Estado |
|---------|-------------|--------|
| [technical-backlog.md](./technical-backlog.md) | Pendientes técnicos y deuda documentada | 🟡 |
| [risks.md](./risks.md) | Riesgos, impacto, probabilidad y mitigación | 🟡 |
| [dependencies.md](./dependencies.md) | Dependencias internas y externas | 🟡 |
| [open-questions.md](./open-questions.md) | Preguntas pendientes de decisión o aclaración | 🔴 |

## Plantillas

| Plantilla | Descripción |
|-----------|-------------|
| [_template-risk-register.md](./_template-risk-register.md) | Registro de riesgos: activos, bloqueantes, resueltos y aceptados |
| [_template-sprint-plan.md](./_template-sprint-plan.md) | Plan de sprint: HUs comprometidas, trazabilidad QA y Definition of Done |
