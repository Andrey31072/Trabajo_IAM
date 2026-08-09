# CI/CD

> Estado: 🟡 En progreso | Última actualización: 2026-08-01
> Autor: Jesús Ariel González Bonilla (PM + Arquitecto) | Equipo: DevOps

Pipeline de integración y entrega para el sistema de Gestión de Horarios SENA. Describe los controles automatizados sobre los cambios y cómo se promueven entre ambientes. Complementa a [00-governance/git-conventions.md](../00-governance/git-conventions.md), a [environments.md](./environments.md) y a [06-data/migration-strategy.md](../06-data/migration-strategy.md).

> **Honestidad sobre el estado:** la disciplina de ramas, Pull Requests y Conventional Commits **ya se practica** (es la forma de trabajo actual, definida en git-conventions). La **automatización** de esos controles como pipeline (checks obligatorios que corren solos en cada PR) está **prevista, no implementada**. Cada control indica su estado.

---

## Alcance

El repositorio actual es principalmente **documental + capa de datos** (repos `*-db` con changelogs Liquibase). Por eso el CI se centra hoy en:

1. Validar que la **documentación** siga las convenciones (git-conventions, documentation-rules).
2. Validar que los **changelogs Liquibase** sean correctos y aplicables antes de promover.

El CI de la **capa de aplicación** (build, tests unitarios, imágenes de contenedor) se definirá cuando esa capa exista.

---

## Controles del pipeline

| Control | Qué verifica | Estado |
|---------|--------------|--------|
| **Conventional Commits** | Formato `<type>(NN-section): description` en inglés; tipos permitidos (`docs`, `fix`, `chore`, `refactor`) | 🟡 Convención vigente; validación automática pendiente |
| **Lint de changelogs** | Estructura de cada changelog, `id` + `author` únicos por changeset, presencia de `rollback` espejo | 🟡 Regla definida ([migration-strategy.md](../06-data/migration-strategy.md)); linter pendiente |
| **Validación Liquibase** | `liquibase validate` + `status` contra una BD limpia; que el `changelog-master.yaml` aplique en orden sin error | 🟡 Se ejecuta manual en local ([local-setup.md](./local-setup.md)); en CI pendiente |
| **PR checks** | Revisión obligatoria antes de merge a rama protegida; sin commits directos a `develop`/`qa`/`staging`/`main` | 🟡 Reglas de rama vigentes; enforcement automático pendiente |
| **Seeds idempotentes** | Que los `02_dml` usen `ON CONFLICT DO NOTHING`/`MERGE` y que los datos de prueba estén aislados por `context`/`labels` | 🟡 Regla definida; verificación pendiente |
| **Secret scanning** | Que no se versionen credenciales reales en `.env.*` ni en el diff | 🔴 **Pendiente — riesgo real** (ver [security-threat-model.md](../05-architecture/security-threat-model.md) §Secretos) |
| **Build / tests de aplicación** | Compilación y pruebas de los servicios | 🔴 No aplica aún (capa app inexistente) |

---

## Flujo previsto por Pull Request

```
Rama hija (hu-NN-dev)  ──PR──▶  develop
        │
        ├─ check: Conventional Commit válido
        ├─ check: liquibase validate + status (BD efímera limpia)
        ├─ check: rollback espejo presente por changeset
        ├─ check: secret scanning (sin credenciales en el diff)
        └─ review humano aprobado
                     │
                     ▼  (todos los checks en verde)
                  merge a develop
```

El mismo conjunto de checks corre en los PRs hacia `qa`, `staging` y `main`, endureciéndose en cada nivel (ver Promoción).

---

## Promoción por ramas/ambientes

La promoción sigue **`develop → qa → staging → main`** ([git-conventions.md](../00-governance/git-conventions.md)):

| PR hacia | Origen | Migraciones | Endurecimiento |
|----------|--------|-------------|----------------|
| `develop` | ramas `hu-*-dev` | forward + rollback local | validación básica |
| `qa` | `hu-*-qa` (merge/cherry-pick desde develop) | **forward-only** | + validación funcional |
| `staging` | `hu-*-staging` (desde qa) | **forward-only** | + validación de preproducción |
| `main` | `release/<iteracion>` (desde staging) | **forward-only** | + tag de release (`04_tcl/`) |

- **Forward-only** en `qa`/`staging`/`main`: el pipeline no revierte migraciones en el flujo normal; el rollback es contingencia documentada.
- Cada release etiqueta una versión (`04_tcl/`) para permitir `rollback <tag>` como plan de contingencia.
- Los **datos de prueba no se promueven**: se aíslan con `context`/`labels` de Liquibase para que no lleguen a `staging`/`main`.

---

## Pendientes de automatización (priorizados)

| # | Acción | Prioridad |
|---|--------|-----------|
| 1 | Secret scanning en cada PR y sacar `.env.*` del control de versiones | **Crítica** |
| 2 | Job de CI que corra `liquibase validate` + `update` + `rollback` contra una BD efímera por PR | Alta |
| 3 | Linter de Conventional Commits como check obligatorio | Alta |
| 4 | Enforcement automático de ramas protegidas (no merge sin checks en verde) | Media |
| 5 | Definir el pipeline de la capa de aplicación cuando exista | Media |

## Referencias

- [00-governance/git-conventions.md](../00-governance/git-conventions.md) — ramas, PRs, Conventional Commits
- [00-governance/documentation-rules.md](../00-governance/documentation-rules.md) — reglas de contenido
- [06-data/migration-strategy.md](../06-data/migration-strategy.md) — changelogs, seeds, rollbacks, tags
- [environments.md](./environments.md) — ambientes y `.env`
- [05-architecture/deployment.md](../05-architecture/deployment.md) — topología de despliegue
