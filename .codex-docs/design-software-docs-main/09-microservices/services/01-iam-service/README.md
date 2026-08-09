# iam-service — Autenticación, autorización e identidad

> Estado: 🟡 Revisión | Última actualización: 2026-06-17
> Autor: Jesús Ariel González Bonilla (PM + Arquitecto) | Equipo: Arquitectura

## Responsabilidad

Gestiona identidad de usuarios, emisión de tokens JWT y control de acceso basado en roles y features (RBAC). Es el único servicio que autentica usuarios; todos los demás servicios verifican el JWT localmente sin llamar a IAM en cada request.

## Bounded context

| Entidad | Descripción |
|---------|-------------|
| `User` | Cuenta de acceso con bloqueo por intentos fallidos |
| `Module` | Grupo funcional del sistema (sección de UI) |
| `Feature` | Capacidad específica dentro de un módulo (unidad de permiso) |
| `Role` | Conjunto de features y scopes asignables a usuarios |
| `RoleFeature` | Asignación de feature a rol con scope de acceso |
| `UserRole` | Asignación de rol a usuario (opcionalmente restringida a un centro) |
| `UserScopeOverride` | Override individual de acceso para casos excepcionales |
| `RefreshToken` | Token de renovación de sesión; múltiples por usuario |
| `PasswordResetRequest` | Solicitud de recuperación de contraseña |
| `AuditLogin` | Log inmutable de intentos de autenticación |

## Módulo de origen

M1 — IAM / Seguridad

## Dependencias

| Servicio | Tipo | Motivo |
|----------|------|--------|
| Ninguna | — | Servicio de entrada; no depende de otros servicios del sistema |

## Componentes desplegables

| Componente | Sufijo | Descripción |
|------------|--------|-------------|
| `iam-api` | `-api` | REST API: login, logout, refresh, gestión de usuarios, roles y features |

## Base de datos

- Nombre lógico: `iam_db`
- Motor: PostgreSQL
- Encriptación en reposo: habilitada (credenciales y tokens)

## Diseño de seguridad

Ver documentos especializados:
- [rbac-design.md](./rbac-design.md): roles, features, scopes y matriz de permisos
- [data-model.md](./data-model.md): modelo completo con entidades de sesión y auditoría
- [components/iam-api/contract.md](./components/iam-api/contract.md): API completa

## Links

- Repo: (pendiente)
- Data model: [data-model.md](./data-model.md)
- RBAC design: [rbac-design.md](./rbac-design.md)
- Eventos: [events.md](./events.md)
- Runbook: [runbook.md](./runbook.md)
- Decisiones internas: [decisions.md](./decisions.md)
