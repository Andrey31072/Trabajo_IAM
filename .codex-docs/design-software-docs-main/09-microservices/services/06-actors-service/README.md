# actors-service

> Estado: 🟡 En progreso | Última actualización: 2026-08-01
> Autor: Jesús Ariel González Bonilla (PM + Arquitecto) | Equipo: Arquitectura

Gestiona los **actores** de la formación: instructores, aprendices y empresas, con sus contratos, áreas, asignaciones de competencia y etapas productivas.

> **Estado real:** existe la **capa de datos** ([data-model.md](./data-model.md)). La capa de aplicación es **diseño previsto — no construido**. Documentación agnóstica de lenguaje.

## Módulo
M7 (actores).

## Componentes previstos
| Componente | Tipo | Responsabilidad |
|------------|------|-----------------|
| [`actors-api`](./components/actors-api/) | `-api` | CRUD de instructores, aprendices, empresas y asignaciones |

## Datos
Schema `actors_parameterization`. Entidades: `instructor`, `learner`, `company`, `instructor_area`, `instructor_contract`, `learner_ficha_enrollment`, `competency_assignment`, `productive_stage`, y el patrón de estados parametrizables (`status`/`status_category`/`status_transition`, ADR-004). Ver [data-model.md](./data-model.md).

## Eventos publicados
`actors.instructor.availability_changed`, `actors.learner.enrolled` (ver [event-catalog.md](../../event-catalog.md)) — consumidos por `scheduling` y `monitoring`.

## Dependencias
reference-data (centro, catálogos), academic (fichas) e iam. Referencias cross-servicio por `UUID` sin FK física.
