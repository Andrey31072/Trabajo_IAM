# DevOps

> Estado: 🟡 En progreso | Última actualización: 2026-08-01
> Autor: Jesús Ariel González Bonilla (PM + Arquitecto) | Equipo: DevOps

Índice de la sección de DevOps del sistema de Gestión de Horarios SENA. Cubre cómo se levanta el entorno local, cómo se validan y promueven los cambios, y qué ambientes existen.

## Estado real

Hoy DevOps opera sobre la **capa de datos**: una base de datos **PostgreSQL 16** en **Docker** y **runners de Liquibase por módulo** que aplican las migraciones de cada repo `*-db`. No hay pipeline de aplicación (APIs/workers) porque esa capa aún no existe. Las páginas marcan lo implementado frente a lo previsto.

## Archivos

| Archivo | Descripción | Estado |
|---------|-------------|--------|
| [local-setup.md](./local-setup.md) | Levantar Postgres en contenedor y aplicar los changelogs de cada módulo en local | 🟡 |
| [ci-cd.md](./ci-cd.md) | Pipeline previsto: validación de changelogs, PR checks, Conventional Commits y promoción por ramas (implementado vs pendiente) | 🟡 |
| [environments.md](./environments.md) | Los 4 ambientes (`develop`/`qa`/`staging`/`main`), variables por `.env.<ambiente>` y advertencia de secretos | 🟡 |

## Plantillas

| Plantilla | Descripción |
|-----------|-------------|
| [_template-deployment-plan.md](./_template-deployment-plan.md) | Plan de despliegue: componentes, pasos, verificación y rollback |
| [_template-release-checklist.md](./_template-release-checklist.md) | Checklist de release: gates de código, QA, seguridad e infraestructura |
| [_template-rollback-plan.md](./_template-rollback-plan.md) | Plan de rollback: criterios de activación y procedimientos por componente |

## Cómo se relaciona

- La **estrategia de migraciones** (orden, seeds idempotentes, rollbacks, tags) vive en [06-data/migration-strategy.md](../06-data/migration-strategy.md).
- La **topología de despliegue** (qué corre y dónde) está en [05-architecture/deployment.md](../05-architecture/deployment.md).
- El **flujo de ramas y Conventional Commits** que gobierna la promoción está en [00-governance/git-conventions.md](../00-governance/git-conventions.md).
- El **riesgo de secretos versionados** en `.env.*` se detalla en [05-architecture/security-threat-model.md](../05-architecture/security-threat-model.md) §Secretos.
