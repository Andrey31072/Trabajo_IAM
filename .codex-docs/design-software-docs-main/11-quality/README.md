# Calidad

> Estado: 🟡 En progreso | Última actualización: 2026-08-01
> Autor: Jesús Ariel González Bonilla (PM + Arquitecto) | Equipo: Calidad

## Contexto

Esta sección define cómo se asegura la calidad del sistema **Horarios SENA**. El proyecto se construye por capas: hoy solo existe la **capa de datos** de los 9 microservicios (repos `*-db` con Liquibase + PostgreSQL 16), mientras que las capas de aplicación (API / worker / workflow en Go) aún no se han construido. Por eso la estrategia de calidad distingue explícitamente **lo que se prueba hoy** (migraciones, DDL, seeds, restauración de BD) de **lo que se probará cuando exista la capa de aplicación** (unitarias, contrato, integración, e2e).

Las prácticas aquí descritas se apoyan en la gobernanza ya definida: la [Definition of Done](../00-governance/definition-of-done.md), la [Definition of Ready](../00-governance/definition-of-ready.md), las [convenciones de Git](../00-governance/git-conventions.md) (git flow `develop → qa → staging → main` y Conventional Commits) y la [estrategia de migraciones](../06-data/migration-strategy.md).

## Contenido

Define prácticas de pruebas, revisión de código y criterios de calidad.

## Archivos

| Archivo | Descripción | Estado |
|---------|-------------|--------|
| [testing-strategy.md](./testing-strategy.md) | Estrategia de pruebas por nivel y tipo | 🟡 |
| [code-review.md](./code-review.md) | Criterios y flujo para revisiones de código | 🟡 |

## Plantillas

| Plantilla | Descripción |
|-----------|-------------|
| [_template-qa-report.md](./_template-qa-report.md) | Reporte de QA: cobertura, defectos, HUs verificadas y gate de calidad |
| [_template-test-evidence.md](./_template-test-evidence.md) | Evidencia de pruebas: casos TC-NNN, defectos y trazabilidad AC-HU-TC |

## Puntos abiertos

- Selección de librerías de prueba de la capa Go (framework de test, mocks, contenedores efímeros): pendiente hasta que se inicie la construcción de la capa de aplicación.
- Umbral de cobertura exacto y su gate automatizado en CI: pendiente de cerrar junto con [ci-cd.md](../10-devops/ci-cd.md).
