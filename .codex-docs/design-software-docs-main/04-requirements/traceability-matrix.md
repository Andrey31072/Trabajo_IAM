# Matriz de Trazabilidad

> Estado: 🟡 En progreso | Última actualización: 2026-08-01
> Autor: Jesús Ariel González Bonilla (PM + Arquitecto) | Equipo: Producto

> Fuente: [functional.md](./functional.md) · [user-stories.md](./user-stories.md) · [02-domain/entities-and-rules.md](../02-domain/entities-and-rules.md)

## Propósito

Relaciona cada **requisito funcional (RF-*)** con la **historia de usuario (HU-*)** que lo motiva, el **servicio/módulo** dueño, la **regla de negocio (RN-*)** que lo gobierna y la **decisión de arquitectura (ADR)** aplicable. Permite verificar cobertura: que ningún RF quede sin historia ni sin dueño, y que cada decisión de arquitectura respalde requisitos reales.

Leyenda de servicios: `iam`, `reference-data`, `academic-management`, `training-environment`, `scheduling` (**CORE**), `actors`, `document`, `monitoring` (**CORE**), `audit`.

---

## Matriz RF ↔ HU ↔ Servicio ↔ RN ↔ ADR

| RF | Historia(s) | Servicio / Módulo | Regla(s) RN | ADR | Épica |
|----|-------------|-------------------|-------------|-----|-------|
| RF-IAM-01 | HU-16 | `iam-service` (M1) | RN-IAM-02 | — | EPC-01 |
| RF-IAM-02 | HU-18 | `iam-service` (M1) | RN-IAM-05 | — | EPC-01 |
| RF-IAM-03 | HU-17 | `iam-service` (M1) | RN-IAM-03 | — | EPC-01 |
| RF-IAM-04 | HU-18 | `iam-service` (M1) | RN-IAM-04 | — | EPC-01 |
| RF-IAM-05 | HU-16 | `iam-service` (M1) | RN-IAM-01 | — | EPC-01 |
| RF-REF-01 | HU-19 | `reference-data-service` (M2) | RN-REF-01, RN-REF-02 | — | EPC-03 |
| RF-REF-02 | HU-19 | `reference-data-service` (M4) | RN-REF-03 | ADR-004 | EPC-03 |
| RF-REF-03 | HU-19 | `reference-data-service` (M4) | RN-REF-04 | — | EPC-03 |
| RF-ACAD-01 | HU-01 | `academic-management-service` (M5) | RN-ACAD-01, RN-ACAD-02 | — | EPC-04 |
| RF-ACAD-02 | HU-01, HU-11 | `academic-management-service` (M6) | RN-ACAD-03, RN-ACAD-06 | — | EPC-04 |
| RF-ACAD-03 | HU-02 | `academic-management-service` (M6) | RN-ACAD-04, RN-ACAD-05 | ADR-004 | EPC-04 |
| RF-ENV-01 | HU-03 | `training-environment-service` (M3) | RN-ENV-01, RN-ENV-04 | — | EPC-05 |
| RF-ENV-02 | HU-03 | `training-environment-service` (M3) | RN-ENV-02 | — | EPC-05 |
| RF-ENV-03 | HU-04 | `training-environment-service` (M3) | RN-ENV-03 | — | EPC-05 |
| RF-ENV-04 | HU-05 | `training-environment-service` (M3) | RN-ENV-01, RN-ENV-03 | ADR-002 | EPC-05 |
| RF-SCH-01 | HU-07 | `scheduling-service` (M8) **CORE** | RN-ACAD-04, RN-SCH-02 | ADR-004 | EPC-07 |
| RF-SCH-02 | HU-08 | `scheduling-service` (M8) **CORE** | RN-SCH-05 | — | EPC-07 |
| RF-SCH-03 | HU-09 | `scheduling-service` (M8) **CORE** | RN-SCH-03, RN-SCH-04 | ADR-001, ADR-002 | EPC-07 |
| RF-SCH-04 | HU-08 | `scheduling-service` (M8) **CORE** | RN-SCH-05, RN-SCH-06 | — | EPC-07 |
| RF-SCH-05 | HU-10 | `scheduling-service` (M8) **CORE** | RN-SCH-01 | ADR-001, ADR-004 | EPC-08 |
| RF-SCH-06 | HU-12, HU-15 | `scheduling-service` (M8) **CORE** | RN-SCH-01, RN-ACT-03 | — | EPC-08 |
| RF-SCH-07 | (soporte de HU-05/06/07) | `scheduling-service` (M8) **CORE** | RN-SCH-03 | ADR-002 | EPC-07 |
| RF-ACT-01 | HU-13 | `actors-service` (M7) | RN-ACT-01, RN-ACT-02 | — | EPC-06 |
| RF-ACT-02 | HU-14 | `actors-service` (M7) | RN-ACT-02 | ADR-002 | EPC-06 |
| RF-ACT-03 | HU-06 | `actors-service` (M7) | RN-ACT-02 | ADR-002 | EPC-06 |
| RF-ACT-04 | HU-11 | `actors-service` (M7) | RN-ACT-03 | — | EPC-06 |
| RF-ACT-05 | (fuera de MVP) | `actors-service` (M7) | RN-ACT-04, RN-ACT-05 | — | EPC-12 |
| RF-DOC-01 | HU-22 | `document-service` | — | ADR-003 | EPC-10 |
| RF-DOC-02 | HU-22 | `document-service` | — | ADR-003 | EPC-10 |
| RF-MON-01 | (disparo por HU-10) | `monitoring-service` (M9) **CORE** | RN-MON-01 | ADR-001 | EPC-09 |
| RF-MON-02 | HU-20 | `monitoring-service` (M9) **CORE** | RN-MON-02, RN-CROSS-01 | — | EPC-09 |
| RF-MON-03 | HU-21 | `monitoring-service` (M9) **CORE** | RN-MON-04, RN-MON-05 | — | EPC-09 |
| RF-MON-04 | HU-21 | `monitoring-service` (M9) **CORE** | RN-MON-03 | — | EPC-09 |
| RF-MON-05 | HU-23 | `monitoring`/`actors` (M9/M7) | RN-MON-06, RN-ACT-06 | — | EPC-11 |
| RF-AUD-01 | (transversal, todas) | `audit-service` | RN-CROSS-01 | ADR-001 | EPC-02 |
| RF-AUD-02 | (transversal) | `audit-service` / broker | RN-CROSS-01 | ADR-001 | EPC-02 |
| RF-AUD-03 | (transversal) | todos | RN-CROSS-01..05 | ADR-004 | EPC-02 |

---

## Cobertura inversa: HU → RF

Verifica que cada historia esté respaldada por al menos un requisito funcional.

| Historia | RF que la implementan | Rol |
|----------|-----------------------|-----|
| HU-01 | RF-ACAD-01, RF-ACAD-02 | Coordinador |
| HU-02 | RF-ACAD-03 | Coordinador |
| HU-03 | RF-ENV-01, RF-ENV-02 | Admin |
| HU-04 | RF-ENV-03 | Admin |
| HU-05 | RF-ENV-04 | Coordinador |
| HU-06 | RF-ACT-03 | Coordinador |
| HU-07 | RF-SCH-01 | Coordinador |
| HU-08 | RF-SCH-02, RF-SCH-04 | Coordinador |
| HU-09 | RF-SCH-03 | Coordinador |
| HU-10 | RF-SCH-05 | Coordinador |
| HU-11 | RF-ACT-04, RF-ACAD-02 | Coordinador |
| HU-12 | RF-SCH-06 | Aprendiz |
| HU-13 | RF-ACT-01 | Coordinador |
| HU-14 | RF-ACT-02 | Instructor |
| HU-15 | RF-SCH-06 | Instructor |
| HU-16 | RF-IAM-01, RF-IAM-05 | Todos |
| HU-17 | RF-IAM-03 | Admin |
| HU-18 | RF-IAM-02, RF-IAM-04 | Admin |
| HU-19 | RF-REF-01, RF-REF-02, RF-REF-03 | Admin |
| HU-20 | RF-MON-02 | Coordinador |
| HU-21 | RF-MON-03, RF-MON-04 | Coordinador |
| HU-22 | RF-DOC-01, RF-DOC-02 | Coordinador |
| HU-23 | RF-MON-05 | Coordinador |

---

## Trazabilidad de decisiones de arquitectura (ADR → RF)

| ADR | Decisión | RF que soporta |
|-----|----------|----------------|
| [ADR-001](../05-architecture/decisions/records/ADR-001-message-broker.md) | Broker RabbitMQ (eventos, DLQ) | RF-SCH-03/05, RF-MON-01, RF-AUD-01/02 |
| [ADR-002](../05-architecture/decisions/records/ADR-002-scheduling-read-models.md) | Read models locales en scheduling | RF-SCH-07, RF-SCH-03, RF-ENV-04, RF-ACT-02/03 |
| [ADR-003](../05-architecture/decisions/records/ADR-003-object-storage.md) | Object storage compatible S3 | RF-DOC-01, RF-DOC-02 |
| [ADR-004](../05-architecture/decisions/records/ADR-004-status-parametrization-and-audit-standard.md) | Estados parametrizables + auditoría | RF-ACAD-03, RF-SCH-01/05, RF-REF-02, RF-AUD-03 |

---

## Cobertura de criterios de éxito del MVP

Enlaza los criterios funcionales (CF-*) de [problem-framing.md](../03-product/problem-framing.md) con los RF que los cumplen.

| Criterio | Descripción | RF |
|----------|-------------|----|
| CF-01 | Crear borrador de horario para ficha activa | RF-SCH-01 |
| CF-02 | Detectar conflictos antes de publicar | RF-SCH-03 |
| CF-03 | Publicar horario sin conflictos con una acción | RF-SCH-05 |
| CF-04 | Instructor consulta su horario por semana | RF-SCH-06 |
| CF-05 | Aprendiz consulta el horario de su ficha | RF-SCH-06 |
| CF-06 | Horario publicado inmutable; cambios crean versión | RF-SCH-05 |

---

## Notas de cobertura y puntos abiertos

- **RF-ACT-05** (etapa productiva) y **EPC-12** están fuera del MVP; se listan para completitud pero no tienen historia priorizada en el alcance actual.
- **RF-MON-01** y **RF-AUD-01/02/03** son requisitos disparados por eventos o transversales; su "historia" es indirecta (se activan desde HU-10 o aplican a todas las operaciones).
- Los puntos abiertos PA-01…PA-04 (ver [vision.md](../03-product/vision.md)) pueden modificar la relación RF ↔ HU si cambian los supuestos (p. ej. integración SOFIA Plus, multi-ficha por aprendiz).

## Referencias

- [functional.md](./functional.md) · [user-stories.md](./user-stories.md) · [non-functional.md](./non-functional.md)
- [02-domain/entities-and-rules.md](../02-domain/entities-and-rules.md)
- [03-product/product-backlog.md](../03-product/product-backlog.md) · [03-product/problem-framing.md](../03-product/problem-framing.md)
- ADRs: [001](../05-architecture/decisions/records/ADR-001-message-broker.md) · [002](../05-architecture/decisions/records/ADR-002-scheduling-read-models.md) · [003](../05-architecture/decisions/records/ADR-003-object-storage.md) · [004](../05-architecture/decisions/records/ADR-004-status-parametrization-and-audit-standard.md)
