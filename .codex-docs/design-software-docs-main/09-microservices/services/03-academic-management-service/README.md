# academic-management-service

> Estado: 🟡 En progreso | Última actualización: 2026-08-01
> Autor: Jesús Ariel González Bonilla (PM + Arquitecto) | Equipo: Arquitectura

Gestiona la oferta de formación (líneas, redes, programas, competencias, resultados de aprendizaje) y las **fichas** (grupos de aprendices matriculados).

> **Estado real:** existe la **capa de datos** ([data-model.md](./data-model.md), verificada). La capa de aplicación (API/workers) es **diseño previsto — no construido**. Documentación agnóstica de lenguaje.

## Módulos
M5 (oferta de formación) + M6 (fichas).

## Componentes previstos
| Componente | Tipo | Responsabilidad |
|------------|------|-----------------|
| [`academic-management-api`](./components/academic-management-api/) | `-api` | CRUD de programas/competencias/fichas y consultas |

## Datos
Schema `academic_management`. Entidades: `training_program`, `competency`, `learning_outcome`, `enrollment_ficha`, jerarquía `tech_line → tech_network → knowledge_network`. Ver [data-model.md](./data-model.md).

## Eventos publicados
`academic.enrollment_ficha.created`, `academic.enrollment_ficha.status_changed`, `academic.training_program.published` (ver [event-catalog.md](../../event-catalog.md)) — consumidos por `scheduling` y `monitoring`.

## Dependencias
- **reference-data** (centro de formación, catálogos) e **iam** (identidad) como base.
- Referencias cross-servicio por `UUID` sin FK física (regla de microservicios, ver [service-boundary-rules.md](../../service-boundary-rules.md)).
