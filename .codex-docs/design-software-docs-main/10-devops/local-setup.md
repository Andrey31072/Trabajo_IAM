# Setup local

> Estado: 🟡 En progreso | Última actualización: 2026-08-01
> Autor: Jesús Ariel González Bonilla (PM + Arquitecto) | Equipo: DevOps

Cómo levantar la base de datos del proyecto y aplicar las migraciones de cada módulo en local. Complementa a [migration-strategy.md](../06-data/migration-strategy.md).

## Prerrequisitos

- Docker + Docker Compose.
- Repos `*-db` clonados junto a la carpeta de infraestructura (ver estructura del monorepo).

## Arquitectura local

Una **única base de datos PostgreSQL 16** en contenedor, y un servicio Liquibase por módulo (perfil `tooling`) que aplica el `changelog-master.yaml` del repo correspondiente. La configuración por ambiente vive en archivos de entorno: `.env.develop`, `.env.qa`, `.env.staging`, `.env.main`.

> **Nota de secretos:** los archivos `.env.*` **no deben contener contraseñas reales versionadas**. Usar `.env.example` como plantilla y mantener los valores reales fuera de git (ver [security-rules.md](../00-governance/security-rules.md)).

## Pasos

1. **Seleccionar ambiente y levantar Postgres.** Docker Compose lee `.env` por defecto, así que hay que pasar el archivo explícitamente:

   ```bash
   docker compose --env-file .env.develop up postgres -d
   ```

2. **Aplicar los changelogs de un módulo** (perfil `tooling`):

   ```bash
   docker compose --env-file .env.develop --profile tooling run --rm liquibase-<modulo> update
   ```

   Módulos disponibles: `liquibase-academic`, `liquibase-actors`, `liquibase-audit`, `liquibase-document`, `liquibase-iam`, `liquibase-monitoring`, `liquibase-reference`, `liquibase-scheduling`, `liquibase-training`.

3. **Verificar** el estado de las migraciones de un módulo:

   ```bash
   docker compose --env-file .env.develop --profile tooling run --rm liquibase-<modulo> status --verbose
   ```

## Rollback local

```bash
docker compose --env-file .env.develop --profile tooling run --rm liquibase-<modulo> rollbackCount 1
```

Para volver a un release etiquetado: `rollback <tag>` (los tags se definen en `04_tcl/`; ver [migration-strategy.md](../06-data/migration-strategy.md)).

## Apagar y limpiar

```bash
docker compose --env-file .env.develop down          # detiene los contenedores
docker compose --env-file .env.develop down -v       # además borra el volumen de datos (reinicio limpio)
```

## Orden recomendado de aplicación

Aplicar primero `reference-data` (catálogos e institucional) e `iam` (identidad), y luego el resto de módulos que los referencian.
