# Revisión de código

> Estado: 🟡 En progreso | Última actualización: 2026-08-01
> Autor: Jesús Ariel González Bonilla (PM + Arquitecto) | Equipo: Calidad

Criterios y flujo para revisar los cambios del sistema **Horarios SENA**. En el estado actual, la mayoría de los PRs modifican la **capa de datos** (repos `*-db`: changelogs Liquibase, DDL, seeds) y documentación; cuando exista la capa de aplicación (Go) se aplicará además el bloque de revisión de código de servicio. La revisión es obligatoria: ningún cambio entra a una rama padre sin al menos una aprobación (ver [Definition of Done](../00-governance/definition-of-done.md)).

## Flujo de revisión

Alineado con las [convenciones de Git](../00-governance/git-conventions.md) (`develop → qa → staging → main`):

1. El autor abre PR desde su rama hija (`hu-NN-dev`, `feat/*`, `fix/*`, etc.) hacia la rama padre correspondiente.
2. El PR cumple la [Definition of Ready](../00-governance/definition-of-ready.md) antes de solicitar revisión.
3. Se asignan revisores según **CODEOWNERS** (ver abajo). Mínimo **1 aprobación**; para cambios en contratos compartidos (API, modelo de datos, eventos, convenciones de Git) se recomiendan **2**.
4. El revisor aplica el checklist correspondiente y deja comentarios accionables.
5. El autor resuelve los comentarios; los hilos se cierran antes del merge.
6. Merge solo con checks verdes y aprobación. El autor actualiza el `CHANGELOG` si el cambio toca gobernanza, estructura de carpetas o un contrato compartido.

## Checklist base (todo PR)

Deriva de la Definition of Done y de la Definition of Ready:

- [ ] El commit sigue **Conventional Commits**: `<type>(NN-section): descripción en inglés` (tipos permitidos según el repo; ver git-conventions).
- [ ] El cambio es pequeño, atómico y trazable a una HU o tarea.
- [ ] Los enlaces relativos funcionan y el archivo está enlazado desde el `README.md` de su sección (para docs).
- [ ] No hay **secretos** ni datos sensibles: sin credenciales, tokens, `.env` con valores reales, `.pem`/`.key`, ni datos personales de aprendices/instructores (ver [security-rules.md](../00-governance/security-rules.md)).
- [ ] `CHANGELOG` actualizado si aplica (gobernanza, estructura, contrato compartido).
- [ ] Los checks automáticos del PR pasan.

## Checklist de cambios de datos (repos `*-db` / DDL / Liquibase)

Aplica a la capa que existe hoy. Complementa a [modeling-conventions.md](../06-data/modeling-conventions.md) y [migration-strategy.md](../06-data/migration-strategy.md):

- [ ] Cada **changeset** declara `id` + `author` únicos y su **`rollback`** espejo en `05_rollbacks/` con la misma ruta relativa.
- [ ] Se respeta el **orden de aplicación**: tablas sin FK en `03_tables`, FKs vía `ALTER TABLE` en `04_alter`, índices (incluido uno por FK) en `10_indexes`.
- [ ] Cada FK declara acción `ON UPDATE`/`ON DELETE` acorde a la convención (catálogo/padre `RESTRICT`, composición `CASCADE`, opcional `SET NULL`).
- [ ] **Nomenclatura de constraints**: `pk_`, `uq_`, `fk_`, `ck_`, `ix_`.
- [ ] El servicio escribe **solo en su schema** (nunca en `public`); tracking de Liquibase aislado por módulo.
- [ ] **Seeds idempotentes** en `02_dml/` (`ON CONFLICT DO NOTHING`/`MERGE`); datos de prueba aislados con `context`/`labels` para que no lleguen a `staging`/`main`.
- [ ] Lógica almacenada (vistas/funciones/procedimientos/triggers) marcada con **`runOnChange: true`**.
- [ ] **Estados**: los estados de negocio usan el catálogo parametrizable (`status_category`/`status`) según ADR-004; los enums técnicos cerrados permanecen como `VARCHAR + CHECK` (no `ENUM` nativo).
- [ ] Auditoría según estándar: tablas transaccionales con `created_*`/`updated_*`/`deleted_*`, `is_active`, `row_version`; append-only solo con su timestamp de inserción.
- [ ] Migración probada en local: `update` + `status` verdes y `rollback` reversible (ver [testing-strategy.md](./testing-strategy.md)).

## Checklist de código de aplicación (Go — futuro)

Se activará cuando exista la capa API/worker/workflow. Puntos previstos:

- [ ] Respeta las fronteras de servicio y la propiedad de datos ([service-boundary-rules.md](../09-microservices/service-boundary-rules.md)); no accede a la BD de otro servicio.
- [ ] Manejo de errores explícito; sin panics no controlados en rutas de request.
- [ ] Contratos de API y eventos coherentes con [07-api](../07-api/) y [event-catalog.md](../09-microservices/event-catalog.md).
- [ ] Consumo de eventos **idempotente** (tolera at-least-once).
- [ ] Pruebas unitarias/contrato acompañan al cambio y pasan.

> El detalle de este bloque (estilo, linters, umbral de cobertura) es un **punto abierto** hasta que inicie la construcción de la capa de aplicación.

## CODEOWNERS

La asignación automática de revisores se rige por el archivo `CODEOWNERS` del repositorio.

- Cambios de **gobernanza** (`00-governance/`), **datos** (`06-data/`) y **arquitectura** (`05-architecture/`, ADRs) → equipos de Arquitectura/Datos.
- Cada repo de servicio `*-db` → owner del servicio correspondiente (ver [service-catalog.md](../09-microservices/service-catalog.md)).

> **Punto abierto:** el archivo `CODEOWNERS` definitivo (handles reales por servicio/carpeta) está por completar; hoy la asignación se coordina de forma manual con Arquitectura.

## Tono de la revisión

- Comentarios accionables y específicos; distinguir bloqueante de sugerencia.
- Preferir preguntas a imperativos cuando haya duda de intención.
- Revisar la **intención del cambio** (¿resuelve la HU?), no solo la sintaxis.
