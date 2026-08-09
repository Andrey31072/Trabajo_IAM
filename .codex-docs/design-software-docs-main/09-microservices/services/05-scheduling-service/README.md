# scheduling-service — Motor de horarios

> Estado: 🟡 En progreso | Última actualización: 2026-08-01
> Autor: Jesús Ariel González Bonilla (PM + Arquitecto) | Equipo: Arquitectura

> **Diseño previsto — no implementado.** No se ha construido la capa de aplicación ni se ha elegido lenguaje/framework de backend. Los contratos aquí descritos son a nivel de protocolo (REST/JSON y eventos), agnósticos de la tecnología de implementación. La capa vigente y estable es el modelo de datos.

## Overview

`scheduling-service` crea, valida y publica los horarios académicos del SENA. Asigna instructor + ambiente + franja horaria a cada sesión de clase de una ficha para un período, detectando conflictos antes de publicar. Es el core del proyecto: el resto de servicios lo alimentan (fichas, competencias, instructores, ambientes) o lo consumen (seguimiento, auditoría, documentos).

El agregado raíz es `schedule`, con máquina de estados `DRAFT → UNDER_REVIEW → PUBLISHED → ARCHIVED`. Una vez publicado, el horario es inmutable (invariante garantizada a nivel de BD); todo cambio genera un nuevo `schedule` en `DRAFT`.

## Bounded context

La capa vigente y autoritativa es el modelo de datos. Entidades propias: `schedule`, `time_slot`, `class_session`, `scheduling_conflict`.

Ver [data-model.md](./data-model.md) (🟢 Estable) para el detalle de campos, invariantes, índices y las *exclusion constraints* de doble-asignación.

## Componentes previstos

| Componente | Sufijo | Responsabilidad | Contrato |
|------------|--------|-----------------|----------|
| `schedules-api` | `-api` | REST API: CRUD de horarios y sesiones, catálogo de franjas, consulta y resolución de conflictos, publicación (relay del Outbox) | [components/schedules-api/contract.md](./components/schedules-api/contract.md) |
| `scheduling-engine-workflow` | `-workflow` | Orquestación (saga) de la generación de un horario: valida ficha, resuelve disponibilidades contra read models, crea sesiones y compensa en caso de fallo | [components/scheduling-engine-workflow/contract.md](./components/scheduling-engine-workflow/contract.md) |
| `conflict-validator-worker` | `-worker` | Consume eventos de ambientes, mantiene los read models locales y detecta conflictos de forma asíncrona (idempotente por `event_id`, con DLQ) | [components/conflict-validator-worker/contract.md](./components/conflict-validator-worker/contract.md) |

## Eventos publicados

Topic (exchange): `scheduling-events`.

| Evento | Origen | Consumidores |
|--------|--------|-------------|
| `scheduling.class_session.created` | `scheduling-engine-workflow` | `monitoring-service`, `audit-service`, `document-service` |
| `scheduling.class_session.cancelled` | `schedules-api` / `scheduling-engine-workflow` | `monitoring-service`, `audit-service`, `document-service`, `actors-service` |
| `scheduling.conflict.detected` | `conflict-validator-worker` | `monitoring-service`, `audit-service` |
| `scheduling.schedule.published` | `schedules-api` (Outbox) | `monitoring-service`, `audit-service`, `document-service`, `actors-service` |

El evento crítico `scheduling.schedule.published` se publica mediante **Outbox pattern** para garantizar entrega at-least-once. Detalle de payloads y política de reintentos en [events.md](./events.md).

## Dependencias

| Servicio | Tipo | Motivo |
|----------|------|--------|
| `iam-service` | auth | Verificación del JWT y del par feature+scope (no cuenta para el límite de dependencias) |
| `academic-management-service` | sync | Validar ficha y competencias al crear el borrador y en la verificación previa a publicar |
| `actors-service` | evento | Poblar el read model local de instructores y competencias (no síncrono) |
| `training-environment-service` | evento | Poblar el read model local de ambientes y disponibilidad (no síncrono) |

> Este servicio requería originalmente 3 dependencias síncronas (academic, actors, environment) y superaba el límite de 2. Se resuelve con **read models locales poblados por eventos** para instructores/competencias y ambientes/disponibilidad, dejando **una sola** dependencia síncrona (academic-management). Ver [ADR-002](../../../05-architecture/decisions/records/ADR-002-scheduling-read-models.md). El transporte de esos eventos es el broker AMQP de [ADR-001](../../../05-architecture/decisions/records/ADR-001-message-broker.md).

## Base de datos

- Nombre lógico: `scheduling_db`
- Motor: PostgreSQL 16
- Nota: requiere consistencia fuerte para la asignación de franjas; incluye la tabla `outbox` y *exclusion constraints* (`EXCLUDE USING gist`) como red de seguridad contra el doble-booking.

## Links

- Modelo de datos (capa vigente): [data-model.md](./data-model.md)
- Eventos: [events.md](./events.md)
- Runbook: [runbook.md](./runbook.md)
- Decisiones internas: [decisions.md](./decisions.md)
- Catálogo de eventos global: [../../event-catalog.md](../../event-catalog.md)
- Observabilidad y operación: [13-operations](../../../13-operations/observability.md)
