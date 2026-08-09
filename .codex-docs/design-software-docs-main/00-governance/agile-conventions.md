# Convenciones ágiles

> Estado: 🟢 Estable | Última actualización: 2026-08-01
> Autor: Jesús Ariel González Bonilla (PM + Arquitecto) | Equipo: Por definir

Este documento define cómo nombrar, estructurar y referenciar los artefactos ágiles del proyecto. **Este repositorio no persiste las historias de usuario**: el tracker externo (Jira, Trello, GitHub Projects, Azure DevOps u otro) es la fuente de verdad para el seguimiento. Aquí solo viven las convenciones y plantillas.

## Jerarquía de artefactos

```
Épica       EPC-NNN
  └── Feature   FEA-NNN
       └── Historia de usuario   HU-<PROJECT_KEY>-NNN
            └── Criterio de aceptación   AC-NNN
                 └── Caso de prueba   TC-NNN
```

## Naming

### Project Key

Formato: `PRJ-<DOMINIO>-<PRODUCTO>`

- Identifica unívocamente el proyecto en todo el ecosistema.
- Usado como prefijo en los IDs de HU.
- Ejemplo: `PRJ-SENA-HORARIOS`, `PRJ-SENA-NOMINA`

### Épicas

Formato: `EPC-NNN: Título descriptivo`

- `NNN`: número secuencial con tres dígitos (`EPC-001`, `EPC-002`…)
- El título describe el objetivo de negocio, no la solución técnica.
- Numeración continua en todo el proyecto.

| Ejemplo válido | Ejemplo inválido |
|----------------|-----------------|
| `EPC-001: Gestión de disponibilidad de instructores` | `EPIC-001: Módulo horarios` |
| `EPC-002: Notificaciones de cambios de horario` | `EPC-2: Backend notificaciones` |

### Features

Formato: `FEA-NNN: Título`

- Agrupan HUs dentro de una épica.
- `NNN`: número secuencial global del proyecto.

### Historias de usuario

Formato: `HU-<PROJECT_KEY>-NNN: Título`

- `<PROJECT_KEY>`: clave del proyecto (ej: `PRJ-SENA-HORARIOS`)
- `NNN`: número secuencial con tres dígitos, continuo en todo el proyecto.
- El título sigue la forma: `[verbo infinitivo] [objeto] [contexto opcional]`

| Ejemplo válido | Ejemplo inválido |
|----------------|-----------------|
| `HU-PRJ-SENA-HORARIOS-001: Consultar disponibilidad de horario` | `HU-001: Horario` |
| `HU-PRJ-SENA-HORARIOS-042: Exportar reporte de asignaciones` | `HU-42: Feature exportar` |

### Criterios de aceptación

Formato: `AC-NNN` — ligado a su HU.

### Casos de prueba

Formato: `TC-NNN` — ligado a su AC o HU.

### Severidades de defecto

| Severidad | Descripción |
|-----------|-------------|
| `P0` | Bloquea el release — corrección inmediata |
| `P1` | Debe resolverse en el mismo sprint |
| `P2` | Mejora priorizable para el siguiente sprint |
| `P3` | Deuda menor, se agenda sin urgencia |

## Sprint

- Formato sugerido: `Sprint NN` o según la herramienta del equipo.
- Las fechas y capacidad se definen en el tracker externo.
- Cada sprint tiene un `sprint-plan` asociado en `15-project-control/`.

## Estatus de HU

Los siguientes estatus aplican a todo el proyecto y deben reflejarse tanto en el tracker como en la documentación.

| Estatus | Descripción | Quién avanza |
|---------|-------------|--------------|
| `Backlog` | Identificada, sin analizar ni refinar | Producto |
| `Refinamiento` | En análisis activo con el equipo | Producto + Desarrollo |
| `Listo` | Refinada, estimada y aprobada para iniciar | Equipo completo |
| `En desarrollo` | Asignada y en trabajo activo | Desarrollo |
| `En revisión` | Implementación completa, en revisión de par | Revisor |
| `En QA` | En validación funcional y técnica | QA |
| `Hecho` | Validada, aceptada y cerrada | Producto |
| `Cancelada` | Descartada por cambio de alcance o prioridad | Producto |

Una HU no puede pasar a `Listo` sin cumplir el [Definition of Ready](./definition-of-ready.md).
Una HU no puede pasar a `Hecho` sin cumplir el [Definition of Done](./definition-of-done.md).

## Plantilla de HU

Usar la plantilla en [`04-requirements/_template-hu.md`](../04-requirements/_template-hu.md).

## Integración con tracker externo

Este repositorio no duplica el contenido del tracker. En su lugar, cada documento que referencie una HU debe incluir el enlace directo al ítem en la herramienta activa del proyecto.

| Herramienta | Formato de enlace |
|-------------|------------------|
| Jira | `[HU-PRJ-XXX-NNN](https://<dominio>.atlassian.net/browse/<clave>)` |
| Trello | `[HU-PRJ-XXX-NNN](https://trello.com/c/<id>)` |
| GitHub Projects | `[HU-PRJ-XXX-NNN](https://github.com/<org>/<repo>/issues/<numero>)` |
| Azure DevOps | `[HU-PRJ-XXX-NNN](https://dev.azure.com/<org>/<proyecto>/_workitems/edit/<id>)` |
| Otra herramienta | Usar enlace directo al ítem con el formato `[HU-PRJ-XXX-NNN](<URL>)` |

Reemplazar `PRJ-XXX` con el project key real. Mientras no se defina el tracker, usar `[HU-PRJ-XXX-NNN](<pendiente>)` como placeholder.

## Trazabilidad

Cada HU debe poder rastrearse hacia:

| Desde | Hacia |
|-------|-------|
| HU | Épica (`EPC-NNN`) que la contiene |
| HU | Feature (`FEA-NNN`) a la que pertenece |
| HU | Sprint en que se planificó |
| HU | Requisito funcional en `04-requirements/functional.md` |
| HU | Caso de prueba (`TC-NNN`) en `11-quality/` |
| HU | ADR si generó una decisión arquitectónica |

La matriz completa vive en [`04-requirements/traceability-matrix.md`](../04-requirements/traceability-matrix.md).
