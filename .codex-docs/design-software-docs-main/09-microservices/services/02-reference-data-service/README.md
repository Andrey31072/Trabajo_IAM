# reference-data-service — Jerarquía institucional y catálogos maestros

> Estado: 🟡 En progreso | Última actualización: 2026-08-01
> Autor: Jesús Ariel González Bonilla (PM + Arquitecto) | Equipo: Arquitectura

## Overview

Provee la jerarquía geográfica-institucional del SENA (`macroregion` → `microregion` → `department` → `municipality` → `training_center` → `institutional_unit`) y los catálogos maestros y parámetros del sistema. Es un servicio base: fuente de verdad de solo-lectura para el resto de servicios, y el más consultado del ecosistema. Ningún otro servicio escribe en sus tablas.

> **Estado de construcción:** la capa de datos está definida y estable (ver [data-model.md](./data-model.md)). La capa de aplicación (API) es **diseño previsto — no implementado**: no se ha construido ni se ha elegido lenguaje/framework.

## Responsabilidad

Administrar y exponer datos de referencia estables (jerarquía institucional, catálogos, parámetros de configuración). NO gestiona datos transaccionales de negocio ni estados de agregados de otros dominios.

## Bounded context

Entidades propias (fuente autoritativa: [data-model.md](./data-model.md)). Ningún otro servicio escribe estas tablas.

| Entidad | Descripción |
|---------|-------------|
| `macroregion` | Agrupación geográfica de mayor nivel |
| `microregion` | División intermedia de la macrorregión |
| `department` | Departamento (código DANE) |
| `municipality` | Municipio (código DANE) |
| `training_center` | Centro de formación SENA (`center_code`) |
| `institutional_unit` | Sede o unidad dentro del centro |
| `catalog` | Categoría de catálogo (ej: `MODALITY`, `SHIFT`) |
| `catalog_detail` | Valor concreto de un catálogo (ej: `IN_PERSON`) |
| `parameter` | Configuración clave-valor del sistema (patrón EAV) |

## Módulo de origen

M2 — Estructura Institucional + M4 — Parametrización y Catálogos Base

## Base de datos (capa de datos — vigente)

- Nombre lógico: `ref_db`
- Motor: **PostgreSQL 16**
- Estado del modelo: 🟢 Estable — ver [data-model.md](./data-model.md)
- Nota: alta frecuencia de lectura; caché externa (ej: Redis) es opcional a nivel de componente, no del modelo.

## Componentes previstos

> Diseño previsto — no implementado.

| Componente | Sufijo | Descripción | Estado |
|------------|--------|-------------|--------|
| `reference-data-api` | `-api` | REST API: consulta y administración de jerarquía, catálogos y parámetros | Diseño previsto |

Contrato: [components/reference-data-api/contract.md](./components/reference-data-api/contract.md)

## Eventos publicados

Topic: `reference-data-events` (broker AMQP / RabbitMQ — ver [ADR-001](../../../05-architecture/decisions/records/ADR-001-message-broker.md)). Detalle completo en [events.md](./events.md).

| Evento | Descripción | Consumidores |
|--------|-------------|-------------|
| `reference.catalog.updated` | Un catálogo fue modificado (los consumidores invalidan su caché) | `audit-service`, `scheduling-service`, `actors-service` |
| `reference.training_center.created` | Centro de formación nuevo registrado | `audit-service` |

Este servicio **no consume eventos**: es fuente de escritura manual (operaciones administrativas).

## Dependencias

| Servicio | Tipo | Motivo |
|----------|------|--------|
| `iam-service` | sync (auth) | Validación del JWT en cada request; RBAC (features `REF_*`) en operaciones de escritura |

`reference-data-service` e `iam-service` son los **servicios base**: no dependen de ningún otro servicio de dominio.

## Links

- Repo: (pendiente)
- Data model: [data-model.md](./data-model.md)
- Eventos: [events.md](./events.md)
- Runbook: [runbook.md](./runbook.md)
- Decisiones internas: [decisions.md](./decisions.md)
- ADR-004 (estados/auditoría): [ADR-004](../../../05-architecture/decisions/records/ADR-004-status-parametrization-and-audit-standard.md)
