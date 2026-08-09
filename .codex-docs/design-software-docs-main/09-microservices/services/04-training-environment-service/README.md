# training-environment-service

> Estado: 🟡 En progreso | Última actualización: 2026-08-01
> Autor: Jesús Ariel González Bonilla (PM + Arquitecto) | Equipo: Arquitectura

Gestiona los **ambientes físicos de formación**, su inventario, mantenimiento, reglas de disponibilidad y reservas.

> **Estado real:** existe la **capa de datos** ([data-model.md](./data-model.md)). La capa de aplicación es **diseño previsto — no construido**. Documentación agnóstica de lenguaje.

## Módulo
M3 (ambientes).

## Componentes previstos
| Componente | Tipo | Responsabilidad |
|------------|------|-----------------|
| [`training-environment-api`](./components/training-environment-api/) | `-api` | CRUD de ambientes, disponibilidad, reservas, inventario |

## Datos
Schema `training_environment`. Entidades: `environment`, `environment_type`, `availability_rule`, `reservation` (con `EXCLUDE` anti-solape), `inventory_item`, `maintenance`. Ver [data-model.md](./data-model.md).

## Eventos publicados
`environment.availability.changed`, `environment.maintenance.started` — consumidos por `scheduling-service` para re-sincronizar sus read models (ver [ADR-002](../../../05-architecture/decisions/records/ADR-002-scheduling-read-models.md)).

## Dependencias
reference-data (centro/sede) e iam. Referencias cross-servicio por `UUID` sin FK física.
