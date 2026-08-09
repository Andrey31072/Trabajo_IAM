# Requisitos

> Estado: 🟡 En progreso | Última actualización: 2026-08-01
> Autor: Jesús Ariel González Bonilla (PM + Arquitecto) | Equipo: Producto

## Contenido

Centraliza los requisitos del sistema **SENA Gestión de Horarios**: funcionales, no funcionales, historias de usuario y su trazabilidad. Traduce la visión de producto ([03-product/vision.md](../03-product/vision.md)) y las reglas de dominio ([02-domain/entities-and-rules.md](../02-domain/entities-and-rules.md)) en requisitos verificables que guían la construcción de los 9 servicios.

## Archivos

| Archivo | Descripción | Estado |
|---------|-------------|--------|
| [functional.md](./functional.md) | Requisitos funcionales por módulo (RF-*), derivados de las entidades reales | 🟡 |
| [non-functional.md](./non-functional.md) | Requisitos de calidad, seguridad, rendimiento y operación (NFR-*) | 🟡 |
| [user-stories.md](./user-stories.md) | Historias de usuario por rol (HU-*) con criterios de aceptación | 🟡 |
| [traceability-matrix.md](./traceability-matrix.md) | Relación RF ↔ HU ↔ servicio ↔ RN ↔ ADR | 🟡 |
| [_template-hu.md](./_template-hu.md) | Plantilla para documentar una historia de usuario | — |
| [_template-nfr.md](./_template-nfr.md) | Plantilla para requisitos no funcionales | — |

## Convención de identificadores

| Prefijo | Significado | Ejemplo |
|---------|-------------|---------|
| `RF-<MÓDULO>-##` | Requisito funcional por módulo | `RF-SCH-03` |
| `NFR-<CATEGORÍA>-##` | Requisito no funcional | `NFR-PERF-01` |
| `HU-##` | Historia de usuario | `HU-09` |
| `RN-<DOMINIO>-##` | Regla de negocio (en [02-domain](../02-domain/entities-and-rules.md)) | `RN-SCH-04` |

Módulos de RF: `IAM`, `REF`, `ACAD`, `ENV`, `SCH` (**CORE**), `ACT`, `DOC`, `MON` (**CORE**), `AUD`.

## Flujo de trazabilidad

```
02-domain (RN-*) ──▶ functional (RF-*) ──▶ user-stories (HU-*)
        │                   │                     │
        └───────────────────┴─────────────────────┴──▶ traceability-matrix ──▶ ADRs
```

Ver convenciones de naming, estatus e integración con tracker en [00-governance/agile-conventions.md](../00-governance/agile-conventions.md).
