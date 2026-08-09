# Estrategia de pruebas

> Estado: 🟡 En progreso | Última actualización: 2026-08-01
> Autor: Jesús Ariel González Bonilla (PM + Arquitecto) | Equipo: Calidad

Estrategia de pruebas para los 9 microservicios del sistema **Horarios SENA** (iam, reference-data, academic-management, training-environment, scheduling, actors, document, monitoring, audit). La construcción es por capas: **hoy solo existe la capa de datos** (`*-db`, Liquibase + PostgreSQL 16 en Docker). Este documento describe la pirámide completa, pero marca en cada nivel **qué se prueba hoy** y **qué se prueba cuando exista la capa de aplicación (Go)**.

## Principios

- **Prueba lo que existe, no lo que no existe.** Mientras la única capa construida sea la de datos, el foco de calidad son las migraciones, el DDL, los seeds y la restauración; no se escriben pruebas de API para servicios que aún no tienen código.
- **Aislamiento por servicio.** Cada `*-db` se prueba de forma independiente, con su propio schema y su propio tracking de Liquibase (ver [migration-strategy.md](../06-data/migration-strategy.md)).
- **Idempotencia y reversibilidad.** Toda migración debe poder aplicarse, revertirse y re-aplicarse en un entorno limpio sin intervención manual.
- **Rápido y determinista.** Las pruebas corren sobre contenedores efímeros de PostgreSQL 16 y no dependen de datos preexistentes ni de orden entre servicios.

## Pirámide de pruebas

De más numerosas/baratas (base) a más costosas (cima). El estado indica la disponibilidad **hoy**.

| Nivel | Qué valida | Estado hoy |
|-------|------------|------------|
| Migraciones y DDL | El esquema se crea, revierte y re-crea correctamente por servicio | ✅ Aplicable hoy (capa de datos) |
| Datos semilla (seeds) | Catálogos idempotentes se cargan sin duplicar ni fallar | ✅ Aplicable hoy |
| Unitarias (Go) | Lógica de negocio pura en handlers/services | ⏳ Cuando exista la capa de aplicación |
| Contrato entre servicios | Eventos y respuestas respetan el contrato publicado | ⏳ Cuando existan API/worker |
| Integración | Servicio + su BD + broker reales | ⏳ Cuando exista la capa de aplicación |
| End-to-end | Flujo de negocio completo entre servicios | ⏳ Fase posterior |

## Nivel 1 — Pruebas de migraciones Liquibase (hoy)

Es el nivel con mayor cobertura efectiva en el estado actual del proyecto. Cada repo `*-db` tiene un `changelog-master.yaml` y se despliega con el perfil `tooling` de Docker Compose (ver [local-setup.md](../10-devops/local-setup.md)).

Casos mínimos por servicio:

1. **Forward limpio.** Sobre una BD vacía, `update` aplica todo el changelog sin error y respetando el orden `00_extensions → 01_schemas → 02_types → 03_tables → 04_alter (FKs) → 05_views…08_triggers → 10_indexes → 02_dml (seeds) → 03_dcl → 04_tcl`.
2. **Rollback.** Cada changeset declara su `rollback` espejo en `05_rollbacks/`; se valida `rollback <tag>` hasta dejar la BD en el estado anterior. Los rollbacks se prueban **en local**; en `qa`/`staging`/`main` las migraciones son forward-only.
3. **Re-aplicación (idempotencia de esquema).** Tras un rollback, un nuevo `update` vuelve a dejar el esquema íntegro.
4. **`runOnChange` en lógica almacenada.** Vistas, funciones, procedimientos y triggers (`05_views`/`06_functions`/`07_procedures`/`08_triggers`) marcados con `runOnChange: true` se re-aplican al cambiar su definición.
5. **Integridad referencial.** Las FKs creadas en `04_alter` existen con la acción `ON UPDATE`/`ON DELETE` correcta y hay un índice por cada FK.

Comandos de referencia (un módulo a la vez; módulos: `liquibase-academic`, `liquibase-actors`, `liquibase-audit`, `liquibase-document`, `liquibase-iam`, `liquibase-monitoring`, `liquibase-reference`, `liquibase-scheduling`, `liquibase-training`):

```bash
# Levantar Postgres del ambiente de pruebas
docker compose --env-file .env.develop up postgres -d

# Aplicar y verificar el estado de un módulo
docker compose --env-file .env.develop --profile tooling run --rm liquibase-<modulo> update
docker compose --env-file .env.develop --profile tooling run --rm liquibase-<modulo> status --verbose

# Validar reversibilidad en local
docker compose --env-file .env.develop --profile tooling run --rm liquibase-<modulo> rollback <tag>
```

> Buena práctica: usar `liquibase validate` para detectar changesets con `id`/`author` duplicados o `rollback` faltante antes de abrir el PR.

## Nivel 2 — Datos de prueba y seeds (hoy)

- Los **seeds de catálogo** viven en `02_dml/` y son **idempotentes** (`INSERT ... ON CONFLICT DO NOTHING` o `MERGE`), de modo que re-ejecutarlos no falla ni duplica. La prueba consiste en aplicar el seed dos veces y verificar conteos estables.
- Los **datos de prueba** (fixtures que no deben llegar a `staging`/`main`) se aíslan con `context`/`labels` de Liquibase, de modo que un `update` en ambientes compartidos no los cargue.
- Para escenarios que necesiten datos ficticios, usar valores seguros (ver [security-rules.md](../00-governance/security-rules.md)); nunca datos personales reales de aprendices o instructores.

## Nivel 3 — Unitarias de la capa de aplicación (futuro)

Cuando se construya la capa Go de cada servicio:

- Cubrir lógica de negocio pura (validaciones, máquinas de estado parametrizables de ADR-004, cálculo de KPIs en `monitoring`) con dependencias externas simuladas.
- El componente `scheduling-engine-workflow` y `conflict-validator-worker` (validación de choques de horario) son candidatos prioritarios por su densidad de reglas.
- Umbral de cobertura objetivo: **punto abierto**, a fijar en el gate de [ci-cd.md](../10-devops/ci-cd.md).

## Nivel 4 — Pruebas de contrato entre servicios (futuro)

La arquitectura es orientada a eventos (broker, ADR-001) con read models (ADR-002). Cuando existan publicadores/consumidores:

- Validar el **envelope de eventos** y el `event_type` (ej. `scheduling.class_session.created`) contra el [event-catalog.md](../09-microservices/event-catalog.md).
- Verificar **idempotencia de consumo**: `audit` y otros consumidores deben tolerar re-entrega at-least-once (la unicidad por `event_id` en `audit_record` es la referencia de diseño).
- Enfoque consumer-driven contract entre cada par publicador/consumidor, para desacoplar despliegues.

## Nivel 5 — Integración y E2E (futuro)

- **Integración:** servicio real + su PostgreSQL + broker reales en contenedores, ejercitando endpoints y flujos de eventos.
- **E2E:** flujo de negocio completo (ej. crear ambiente → programar sesión → validar conflicto → emitir documento), atravesando varios servicios. Se planificará tras estabilizar integración por servicio.

## Trazabilidad y evidencia

- Cada HU probada se registra con la plantilla [_template-test-evidence.md](./_template-test-evidence.md) (casos `TC-NNN`, defectos, trazabilidad `AC → HU → TC`).
- Cada sprint/release consolida resultados con [_template-qa-report.md](./_template-qa-report.md), incluyendo el gate de calidad.
- La calidad forma parte de la [Definition of Done](../00-governance/definition-of-done.md): un cambio no está *done* si sus pruebas aplicables no pasan.

## Puntos abiertos

- Framework de pruebas y utilería de contenedores efímeros para la capa Go: por definir al iniciar la construcción de aplicación.
- Umbral de cobertura y su automatización en CI: por cerrar con [ci-cd.md](../10-devops/ci-cd.md).
- Herramienta de pruebas de contrato: por evaluar (opciones a valorar cuando exista tráfico de eventos real).
