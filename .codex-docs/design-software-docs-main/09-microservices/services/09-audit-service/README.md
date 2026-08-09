# audit-service — Auditoría transversal (event store)

> Estado: 🟡 En progreso | Última actualización: 2026-08-01
> Autor: Jesús Ariel González Bonilla (PM + Arquitecto) | Equipo: Arquitectura

> **Diseño previsto — no implementado.** Este documento describe el diseño acordado
> del servicio. La capa de aplicación aún no se construye ni se ha elegido su lenguaje;
> los contratos se especifican a nivel de protocolo (eventos consumidos, esquema de BD).

## Responsabilidad

Es el **event store append-only** del sistema: persiste de forma inmutable cada evento de
negocio que circula por el bus. No ejecuta lógica de negocio ni escritura externa; es un
consumidor pasivo activado exclusivamente por eventos. El log de auditoría **nunca** recibe
UPDATE ni DELETE.

## Bounded context

Entidad propia del servicio. Detalle completo en [data-model.md](./data-model.md).

| Entidad | Descripción |
|---------|-------------|
| `audit_record` | Registro inmutable de un evento consumido: `event_id` (UNIQUE, idempotencia), `event_type`, `source_service`, `actor_id`, `entity_type`, `entity_id`, `payload` (JSONB), `event_occurred_at`, `received_at` |

> **Sin FKs por diseño.** `audit_record` no declara claves foráneas: es un log
> desacoplado que debe poder registrar eventos de cualquier servicio sin depender de la
> existencia de sus entidades. Es una tabla append-only (solo INSERT), por lo que **no**
> lleva `updated_at`, soft delete ni `row_version` (ver ADR-004 y data-model.md).

## Módulo de origen

Transversal — sin módulo de dominio directo (soporta trazabilidad de M1–M10).

## Dependencias

| Servicio | Tipo | Motivo |
|----------|------|--------|
| Todos los servicios | async | Consume sus eventos de dominio por suscripción wildcard (`*-events`) |

> Este servicio **no** tiene dependencias síncronas. No expone API. Se activa únicamente
> por mensajes del bus (fan-out global, ADR-001).

## Componentes desplegables

| Componente | Sufijo | Descripción |
|------------|--------|-------------|
| [`audit-worker`](./components/audit-worker/README.md) | `-worker` | Consumidor universal: se suscribe a todos los topics y escribe `audit_record` (append-only) |

## Base de datos

- Nombre lógico: `audit_db`
- Motor: **PostgreSQL 16**
- Esquema: ver [data-model.md](./data-model.md) — vigente 🟢 Estable
- Regla de integridad: **solo INSERT**. Cualquier UPDATE o DELETE en esta BD es un error de diseño.
- Idempotencia: `UNIQUE(event_id)` — la re-entrega at-least-once del broker se absorbe con `ON CONFLICT (event_id) DO NOTHING`.
- Retención: mínimo 7 años (a confirmar con normativa legal SENA); particionado RANGE mensual por `received_at` y archivado a cold storage tras 2 años.

## Eventos

Ver [events.md](./events.md). El `audit-service` **no publica** eventos (sería una
dependencia circular y violaría su invariante de solo-escritura). Consume todos los topics
`*-events` del ecosistema.

## Links

- Repo: (pendiente)
- Data model: [data-model.md](./data-model.md)
- Eventos: [events.md](./events.md)
- Runbook: [runbook.md](./runbook.md)
- Decisiones internas: [decisions.md](./decisions.md)
- Catálogo global de eventos: [event-catalog.md](../../event-catalog.md)
- ADR-001 (broker de mensajes): [ADR-001](../../../05-architecture/decisions/records/ADR-001-message-broker.md)
- ADR-004 (estados y auditoría): [ADR-004](../../../05-architecture/decisions/records/ADR-004-status-parametrization-and-audit-standard.md)
