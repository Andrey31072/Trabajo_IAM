# Checklist de madurez de servicio

> Estado: 🟡 En progreso | Última actualización: 2026-08-01
> Autor: Jesús Ariel González Bonilla (PM + Arquitecto) | Equipo: Arquitectura

Criterios de madurez por nivel. Agnóstico de lenguaje. Hoy los 9 servicios están en el nivel **Datos** (capa de datos lista; capa de aplicación pendiente).

## Niveles

### Nivel 0 — Doc
- [ ] Servicio registrado en [service-catalog.md](./service-catalog.md) con owner y repo.
- [ ] ADR de creación aprobada (ver [microservices-documentation.md](../00-governance/microservices-documentation.md)).

### Nivel 1 — Datos (estado actual de los 9)
- [x] `data-model.md` completo y verificado contra el esquema real.
- [x] Migraciones Liquibase con rollback espejo y schema propio.
- [x] Estados y auditoría según [ADR-004](../05-architecture/decisions/records/ADR-004-status-parametrization-and-audit-standard.md).

### Nivel 2 — Dev (aplicación)
- [ ] Contratos de componente (`README.md` + `contract.md`) definidos.
- [ ] API/worker/workflow implementados (cualquier backend).
- [ ] Pruebas unitarias y de contrato (ver [11-quality/testing-strategy.md](../11-quality/testing-strategy.md)).
- [ ] Publicación/consumo de eventos con idempotencia.

### Nivel 3 — QA
- [ ] Pruebas de integración y e2e en ambiente `qa`.
- [ ] Observabilidad: logs estructurados, métricas RED, trazas ([13-operations/observability.md](../13-operations/observability.md)).
- [ ] Runbook operativo por componente.

### Nivel 4 — Prod
- [ ] Backup/restore probado ([13-operations/backup-and-recovery.md](../13-operations/backup-and-recovery.md)).
- [ ] Secretos fuera de git (Secret Manager).
- [ ] Alertas y SLO definidos; instancia/credencial de BD separada por servicio.

## Estado por servicio (resumen)
Los 9 servicios: **Nivel 1 (Datos) alcanzado**; Nivel 2+ pendiente de construcción de la capa de aplicación.
