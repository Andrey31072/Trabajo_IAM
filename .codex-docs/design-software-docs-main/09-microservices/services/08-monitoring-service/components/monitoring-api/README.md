# monitoring-api

> Estado: 🟡 En progreso | Última actualización: 2026-08-01
> Autor: Jesús Ariel González Bonilla (PM + Arquitecto) | Equipo: Arquitectura

> **Diseño previsto — no implementado.** Contrato especificado a nivel de protocolo (REST);
> sin código. La capa de aplicación es agnóstica del lenguaje.

## Tipo de componente

`-api` — REST API sincrónica (predominantemente de lectura / consulta).

## Responsabilidad

Expone el estado consolidado del seguimiento: consulta de KPIs (`kpi_tracking`), alertas
(`generated_alert`), seguimiento de fichas (`ficha_tracking`) y planes de mejoramiento
(`improvement_plan`). Los coordinadores y el equipo pedagógico usan esta API para ver el
estado de sus fichas y resolver alertas. No consume eventos: eso lo hacen los workers del
servicio.

## Tecnologías

| Capa | Tecnología |
|------|-----------|
| Runtime | Agnóstico — cualquier backend (a definir) |
| Framework | A definir |
| BD | PostgreSQL 16 — `monitoring_db` |
| Auth | JWT emitido por `iam-service` |

## Variables de entorno requeridas

| Variable | Descripción | Ejemplo |
|----------|-------------|---------|
| `SERVICE_PORT` | Puerto de escucha HTTP | `8080` |
| `DB_URL` | Cadena de conexión a `monitoring_db` | `postgresql://host:5432/monitoring_db` |
| `IAM_JWKS_URL` | Endpoint JWKS de `iam-service` para validar el JWT | `https://iam/.well-known/jwks.json` |

## Contrato

Ver [contract.md](./contract.md)
