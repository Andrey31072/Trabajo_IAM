# Product Backlog — SENA Gestión de Horarios

> Estado: 🟡 En progreso | Última actualización: 2026-08-01
> Autor: Jesús Ariel González Bonilla (PM + Arquitecto) | Equipo: Producto

> Fuente: [vision.md](./vision.md) · [roadmap.md](./roadmap.md) · [04-requirements/user-stories.md](../04-requirements/user-stories.md)

## Propósito

Backlog priorizado de alto nivel, organizado en **épicas (EPC-NN)** que agrupan features. Cada épica se ancla a una fase del [roadmap.md](./roadmap.md), a los requisitos funcionales de [functional.md](../04-requirements/functional.md) (RF-*) y a las historias de [user-stories.md](../04-requirements/user-stories.md) (HU-*). La priorización usa **MoSCoW** (Must / Should / Could / Won't-now).

## Épicas

| Épica | Nombre | Fase | Servicio(s) | Prioridad |
|-------|--------|------|-------------|-----------|
| **EPC-01** | Identidad y control de acceso | 0 | `iam-service` | Must |
| **EPC-02** | Plataforma de eventos y auditoría | 0 | broker, `audit-service` | Must |
| **EPC-03** | Datos maestros institucionales | 1 | `reference-data-service` | Must |
| **EPC-04** | Gestión académica (programas y fichas) | 1 | `academic-management-service` | Must |
| **EPC-05** | Gestión de ambientes y disponibilidad | 1 | `training-environment-service` | Must |
| **EPC-06** | Gestión de actores (instructores/aprendices) | 1 | `actors-service` | Must |
| **EPC-07** | Motor de horarios: construcción y conflictos | 2 | `scheduling-service` | Must |
| **EPC-08** | Publicación y consulta de horarios | 2 | `scheduling-service` | Must |
| **EPC-09** | Seguimiento, KPIs y alertas | 3 | `monitoring-service` | Should |
| **EPC-10** | Documentos (PDFs) | 3 | `document-service` | Should |
| **EPC-11** | Planes de mejoramiento | 3 | `monitoring`/`actors` | Should |
| **EPC-12** | Extensiones (SOFIA Plus, IA, offline, 2FA) | 4 | varios | Won't-now |

---

## Backlog priorizado por épica

### EPC-01 — Identidad y control de acceso · Must · Fase 0

| Feature | RF | HU | MoSCoW |
|---------|----|----|--------|
| Login con JWT + refresh token | RF-IAM-01 | HU-16 | Must |
| RBAC por feature + scope de centro | RF-IAM-03 | HU-17 | Must |
| Gestión de usuarios y roles; contraseña temporal | RF-IAM-02, RF-IAM-04 | HU-18 | Must |
| Bloqueo por intentos fallidos | RF-IAM-05 | — | Must |
| Doble factor (2FA) | — | — | Won't-now (V2) |

### EPC-02 — Plataforma de eventos y auditoría · Must · Fase 0

| Feature | RF | MoSCoW |
|---------|----|--------|
| Broker de eventos con DLQ y reintentos | RF-AUD-02 | Must |
| Log de auditoría append-only (todos los topics) | RF-AUD-01 | Must |
| Estándar transversal de estados + auditoría (ADR-004) | RF-AUD-03 | Must |

### EPC-03 — Datos maestros institucionales · Must · Fase 1

| Feature | RF | HU | MoSCoW |
|---------|----|----|--------|
| Jerarquía institucional (macrorregión → centro) | RF-REF-01 | HU-19 | Must |
| Catálogos del sistema (solo lectura para usuario final) | RF-REF-02 | HU-19 | Must |
| Parámetros del sistema configurables | RF-REF-03 | — | Should |

### EPC-04 — Gestión académica · Must · Fase 1

| Feature | RF | HU | MoSCoW |
|---------|----|----|--------|
| Programas, competencias y RAPs | RF-ACAD-01 | HU-01 | Must |
| Registro de fichas de caracterización | RF-ACAD-02 | HU-01 | Must |
| Ciclo de estados de ficha con transiciones gobernadas | RF-ACAD-03 | HU-02 | Must |

### EPC-05 — Ambientes y disponibilidad · Must · Fase 1

| Feature | RF | HU | MoSCoW |
|---------|----|----|--------|
| CRUD de ambientes, tipos y capacidad | RF-ENV-01 | HU-03 | Must |
| Reglas de disponibilidad recurrentes (hasta 24) | RF-ENV-02 | HU-03 | Must |
| Mantenimientos que bloquean asignación | RF-ENV-03 | HU-04 | Must |
| Consulta de ambientes disponibles (< 300 ms) | RF-ENV-04 | HU-05 | Must |

### EPC-06 — Actores · Must · Fase 1

| Feature | RF | HU | MoSCoW |
|---------|----|----|--------|
| Instructores: competencias y tipo de vinculación | RF-ACT-01 | HU-13 | Must |
| Disponibilidad y excepciones del instructor | RF-ACT-02 | HU-14 | Must |
| Consulta de instructores disponibles (< 300 ms) | RF-ACT-03 | HU-06 | Must |
| Aprendices: matrícula y estado | RF-ACT-04 | HU-11 | Must |
| Empresas / etapa productiva | RF-ACT-05 | — | Won't-now (Fase 4) |

### EPC-07 — Motor de horarios: construcción y conflictos · Must · Fase 2 · **CORE**

| Feature | RF | HU | MoSCoW |
|---------|----|----|--------|
| Crear borrador para ficha en `EXECUTION` | RF-SCH-01 | HU-07 | Must |
| Agregar/editar/cancelar sesiones de clase | RF-SCH-02 | HU-08 | Must |
| Detección de conflictos (5 tipos) inmediata y completa | RF-SCH-03 | HU-09 | Must |
| Franjas horarias y días laborales SENA (L–S) | RF-SCH-04 | HU-08 | Must |
| Read models de instructores/ambientes por eventos | RF-SCH-07 | — | Must |
| Sugerencia automática de asignación (IA) | — | — | Won't-now (Fase 4) |

### EPC-08 — Publicación y consulta · Must · Fase 2 · **CORE**

| Feature | RF | HU | MoSCoW |
|---------|----|----|--------|
| Publicación inmutable con versionado | RF-SCH-05 | HU-10 | Must |
| Consulta de horario por instructor (semana) | RF-SCH-06 | HU-15 | Must |
| Consulta de horario por aprendiz (ficha) | RF-SCH-06 | HU-12 | Must |

### EPC-09 — Seguimiento, KPIs y alertas · Should · Fase 3 · **CORE**

| Feature | RF | HU | MoSCoW |
|---------|----|----|--------|
| Inicio de seguimiento al publicar horario | RF-MON-01 | — | Should |
| KPIs por ficha (asistencia, avance) | RF-MON-02 | HU-20 | Should |
| Alertas por umbral y su resolución trazable | RF-MON-03 | HU-21 | Should |
| Cálculo de riesgo de deserción | RF-MON-04 | HU-21 | Should |

### EPC-10 — Documentos · Should · Fase 3

| Feature | RF | HU | MoSCoW |
|---------|----|----|--------|
| Generación de PDF de horario | RF-DOC-01 | HU-22 | Should |
| Constancias y versionado de documentos | RF-DOC-02 | — | Could |

### EPC-11 — Planes de mejoramiento · Should · Fase 3

| Feature | RF | HU | MoSCoW |
|---------|----|----|--------|
| Plan de mejoramiento para aprendiz | RF-MON-05 | HU-23 | Should |
| Plan de mejoramiento para instructor | RF-MON-05 | — | Could |

### EPC-12 — Extensiones · Won't-now · Fase 4

Integración SOFIA Plus, scheduling asistido por IA, resolución automática de conflictos, notificaciones push, exportación a calendarios, 2FA, offline/PWA, app móvil nativa. Ver [scope.md](../01-context/scope.md) (fuera de MVP).

---

## Resumen por release

| Release | Épicas | Resultado |
|---------|--------|-----------|
| **R0 — Fundaciones** | EPC-01, EPC-02 | Identidad, eventos y auditoría operativos |
| **R1 — Datos maestros** | EPC-03..06 | Insumos del motor cargados; disponibilidad consultable |
| **R2 — MVP horarios** (hito) | EPC-07, EPC-08 | Construir → validar → publicar → consultar horarios |
| **R3 — Valor pedagógico** | EPC-09..11 | Seguimiento, KPIs, alertas, documentos y planes |
| **R4 — Extensiones** | EPC-12 | Post-MVP |

## Mapa de dependencias (resumen)

```
EPC-01/02 ──▶ EPC-03/04/05/06 ──▶ EPC-07 ──▶ EPC-08 ──▶ EPC-09/10/11 ──▶ EPC-12
(identidad)   (datos maestros)     (motor)   (publicar)  (seguimiento)    (extensiones)
```

- EPC-07 depende de disponibilidad de instructores/ambientes (EPC-05/06) y de fichas en `EXECUTION` (EPC-04).
- EPC-09/10/11 dependen del evento `scheduling.schedule.published` (EPC-08).

## Puntos abiertos

- Priorización interna de Fase 3 (documentos vs. planes) por confirmar con Producto.
- **PA-03**: si SOFIA Plus (EPC-12) se adelanta al MVP, EPC-04 debe incorporar el ACL correspondiente (ver [domain-map.md](../02-domain/domain-map.md)).

## Referencias

- [vision.md](./vision.md) · [roadmap.md](./roadmap.md)
- [04-requirements/functional.md](../04-requirements/functional.md) · [04-requirements/user-stories.md](../04-requirements/user-stories.md) · [04-requirements/traceability-matrix.md](../04-requirements/traceability-matrix.md)
