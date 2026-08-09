# Convenciones de Git

> Estado: 🟢 Estable | Última actualización: 2026-08-01
> Autor: Jesús Ariel González Bonilla (PM + Arquitecto) | Equipo: Por definir

Este repositorio usa ramas protegidas, Pull Requests y Conventional Commits para mantener trazabilidad documental.

## Ramas protegidas

`develop`, `qa`, `staging` y `main` representan ambientes padre y no se trabajan directamente.

| Rama | Propósito | Regla |
|------|-----------|-------|
| `develop` | Integración de trabajo en desarrollo | Recibe PRs desde ramas hijas |
| `qa` | Validación funcional y técnica | Recibe PRs o cherry-picks aprobados desde `develop` |
| `staging` | Preproducción / validación previa al release | Recibe PRs o cherry-picks aprobados desde `qa` |
| `main` | Producción / documentación estable | Recibe solo PRs desde `release/*` (promovidos desde `staging`) |

## Ramas documentales

| Tipo de rama | Cuándo usarla | Ejemplo | Tipo de commit |
|--------------|---------------|---------|----------------|
| `feat` | Documento nuevo | `feat/doc-api-guidelines` | `docs` |
| `fix` | Corrección de contenido | `fix/doc-scope` | `fix` |
| `chore` | Reorganización o renombrado | `chore/doc-move-adr-003` | `chore` |
| `docs` | Actualización de documento existente | `docs/doc-service-catalog` | `docs` |

El tipo de rama describe intención. El tipo del commit sigue Conventional Commits.

## Ramas por historia de usuario

| Caso | Rama base | Formato | Ejemplo |
|------|-----------|---------|---------|
| Desarrollo de HU | `develop` | `hu-<numero>-dev` | `hu-01-dev` |
| Ajuste o validación QA | `qa` | `hu-<numero>-qa` | `hu-01-qa` |
| Validación en staging | `staging` | `hu-<numero>-staging` | `hu-01-staging` |
| Release de iteración | `main` | `release/<iteracion>` | `release/iteration-01` |

Las ramas `hu-*` son un caso especial para trazabilidad por historia. No siguen el formato `<tipo>/doc-*`.

## Flujo hacia develop

```bash
git checkout develop
git pull origin develop
git checkout -b hu-01-dev

git add <archivos>
git commit -m "docs(04-requirements): add scheduling availability user story"
git push origin hu-01-dev
```

Abrir PR de `hu-01-dev` hacia `develop`.

## Flujo hacia qa

Crear rama hija desde `qa`:

```bash
git checkout qa
git pull origin qa
git checkout -b hu-01-qa
```

Llevar cambios con merge cuando la HU completa pasa igual:

```bash
git merge origin/hu-01-dev
git push origin hu-01-qa
```

O con cherry-pick cuando solo pasan commits específicos:

```bash
git cherry-pick <commit-sha>
git push origin hu-01-qa
```

Abrir PR de `hu-01-qa` hacia `qa`.

## Flujo hacia staging

Una vez validado en `qa`, promover a `staging` (preproducción) con merge o cherry-pick:

```bash
git checkout staging
git pull origin staging
git checkout -b hu-01-staging
git merge origin/hu-01-qa
git push origin hu-01-staging
```

Abrir PR de `hu-01-staging` hacia `staging`.

## Release hacia main

`main` representa documentación estable. Para producción, crear una rama release desde `main`:

```bash
git checkout main
git pull origin main
git checkout -b release/iteration-01
```

La rama release puede acumular varias HUs de una iteración:

```bash
git cherry-pick <commit-hu-01>
git cherry-pick <commit-hu-02>
git cherry-pick <commit-hu-03>
git push origin release/iteration-01
```

Abrir PR de `release/iteration-01` hacia `main`.

## Conventional Commits

Formato obligatorio:

```text
<type>(NN-section): short description in English
```

Tipos permitidos:

| Tipo | Uso |
|------|-----|
| `docs` | Crear o actualizar documentación |
| `fix` | Corregir contenido incorrecto |
| `chore` | Mover, renombrar, reordenar o actualizar metadatos |
| `refactor` | Reestructurar documentación sin cambiar significado |

No usar `feat`, `style`, `test`, `perf`, `build` ni `ci` para commits de este repositorio documental.

Ejemplos:

```bash
docs(04-requirements): add scheduling user stories
docs(09-microservices): register auth service
fix(01-context): clarify project scope
chore(08-uml): export sequence diagrams to SVG
refactor(00-governance): split contribution rules by topic
```

## Reglas de commits

- La descripción del commit va en inglés.
- El contenido de los documentos puede estar en español.
- Los commits deben ser pequeños y trazables.
- Si se documentan varios microservicios, usar un commit por microservicio cuando sea posible.
- No mezclar cambios funcionales de varias secciones sin razón clara.

## Hotfix en main

Cuando se detecta un error crítico en `main` que no puede esperar el flujo normal de release:

| Caso | Rama base | Formato | Ejemplo |
|------|-----------|---------|---------|
| Corrección urgente en documentación estable | `main` | `fix/doc-<descripcion>` | `fix/doc-broken-api-contract` |

Flujo:

```bash
git checkout main
git pull origin main
git checkout -b fix/doc-broken-api-contract

git add <archivos>
git commit -m "fix(07-api): correct broken endpoint reference in contract"
git push origin fix/doc-broken-api-contract
```

Abrir PR directo de `fix/doc-*` hacia `main`. Una vez mergeado, aplicar el mismo fix a `staging`, `qa` y `develop` con cherry-pick:

```bash
git checkout qa
git pull origin qa
git cherry-pick <commit-sha>
git push origin qa
```
