# Microservicio IAM - Gestión de Identidad y Seguridad

Este proyecto implementa un microservicio IAM para manejar autenticación, usuarios, roles, permisos, sesiones, auditoría y recuperación de contraseña por correo. Está pensado para integrarse con una plataforma de horarios SENA, pero funciona de manera independiente en local usando Docker.

## Objetivo Del Proyecto

El sistema permite centralizar la identidad de los usuarios y controlar qué puede hacer cada persona dentro de una aplicación. En vez de manejar permisos directamente en cada módulo, este microservicio entrega tokens JWT con roles y features ya calculados.

Funcionalmente cubre:

- Inicio de sesión con correo y contraseña.
- Generación de access token y refresh token.
- Administración de usuarios.
- Asignación y revocación de roles.
- Consulta de módulos y permisos.
- Gestión de sesiones activas.
- Overrides de permisos por usuario.
- Auditoría de intentos de login.
- Recuperación de contraseña mediante correo SMTP.

## Arquitectura General

El proyecto está dividido en tres partes principales:

```text
backend/                       API REST en Go
frontend/                      Interfaz web en HTML, CSS y JavaScript
design-software-iam-db-main/   Migraciones de base de datos con Liquibase
docker-compose.yml             Orquestación local de todos los servicios
```

La aplicación se ejecuta con Docker Compose y levanta:

| Servicio | Función | Puerto |
| --- | --- | --- |
| `postgres` | Base de datos PostgreSQL | `5432` |
| `liquibase-iam` | Aplica migraciones SQL | interno |
| `iam-api` | Backend REST en Go | `3001` |
| `iam-frontend` | Frontend web con Nginx | `8080` |

## Backend

El backend está desarrollado en Go y usa una estructura por capas:

```text
backend/cmd/iam-api/                  Punto de entrada
backend/internal/domain/              Modelos y errores del dominio
backend/internal/application/         Casos de uso y reglas de negocio
backend/internal/infrastructure/      Postgres, JWT y correo SMTP
backend/internal/interfaces/httpapi/  Rutas HTTP y middlewares
```

### Flujo De Login

Cuando un usuario inicia sesión:

1. El frontend envía correo y contraseña al endpoint `/auth/login`.
2. El backend busca el usuario en PostgreSQL.
3. Compara la contraseña usando bcrypt.
4. Consulta roles, features y scopes del usuario.
5. Genera un JWT firmado con RS256.
6. Crea un refresh token para mantener la sesión.
7. Registra el intento en la auditoría de login.

El token generado incluye información como:

- ID del usuario.
- Tipo de actor.
- Roles.
- Features permitidas.
- Centro de formación, si aplica.
- Tiempo de expiración.

## Frontend

El frontend está hecho con JavaScript vanilla y se sirve desde Nginx. No usa frameworks pesados, por eso la lógica principal está en:

```text
frontend/src/app.js
frontend/src/styles.css
frontend/src/index.html
```

Vistas principales:

| Vista | Ruta | Qué muestra |
| --- | --- | --- |
| Login | `#/login` | Acceso con correo y contraseña |
| Recuperar contraseña | `#/forgot-password` | Solicitud del enlace por correo |
| Nueva contraseña | `#/reset-password?token=...` | Cambio de contraseña con token |
| Dashboard | `#/dashboard` | Resumen de sesión, roles y módulos |
| Usuarios | `#/users` | Listado, filtros, creación y edición |
| RBAC | `#/rbac` | Roles, módulos y features |
| Auditoría | `#/audit` | Intentos de login exitosos y fallidos |

El frontend consume la API usando la ruta relativa:

```text
/api/v1
```

Nginx recibe esas peticiones y las redirige al backend `iam-api`.

## Base De Datos

La base de datos se construye con Liquibase desde la carpeta:

```text
design-software-iam-db-main/
```

Schemas principales:

| Schema | Responsabilidad |
| --- | --- |
| `identity` | Usuarios y datos base de identidad |
| `rbac_catalog` | Módulos y features disponibles |
| `rbac` | Roles, permisos y overrides |
| `session` | Refresh tokens y recuperación de contraseña |
| `identity_audit` | Auditoría de intentos de login |

Tablas importantes:

- `identity.user`: usuarios del sistema.
- `rbac.role`: roles disponibles.
- `rbac.user_role`: roles asignados a usuarios.
- `rbac.role_feature`: permisos asociados a cada rol.
- `rbac.user_scope_override`: permisos especiales por usuario.
- `session.refresh_token`: sesiones activas.
- `session.password_reset_request`: solicitudes de recuperación.
- `identity_audit.audit_login`: historial de intentos de login.

## Roles Y Permisos

El sistema usa RBAC, es decir, control de acceso basado en roles.

Ejemplo:

- Un usuario tiene uno o varios roles.
- Cada rol tiene features.
- Cada feature puede tener un scope, por ejemplo global, por centro de formación o solo datos propios.

Además, el sistema permite overrides. Esto sirve para dar o bloquear permisos específicos a un usuario sin cambiar todo el rol.

## Recuperación De Contraseña Por Correo

La recuperación de contraseña funciona así:

1. El usuario entra a `#/forgot-password`.
2. Escribe su correo.
3. El frontend llama a `/auth/password-reset/request`.
4. El backend valida si el correo existe.
5. Si existe, genera un token seguro.
6. En la base de datos se guarda solo el hash del token.
7. Se envía un enlace al correo configurado.
8. El usuario abre el enlace y define una contraseña nueva.
9. El backend actualiza la contraseña y revoca sesiones anteriores.

El enlace enviado tiene esta forma:

```text
http://localhost:8080/#/reset-password?token=TOKEN_GENERADO
```

Importante: por seguridad, si el correo no existe el backend también responde correctamente. Esto evita revelar qué cuentas están registradas.

## Configuración SMTP

Para que los correos reales lleguen, el archivo `.env` debe tener SMTP configurado. Con Gmail se usa una App Password, no la contraseña normal de la cuenta.

Ejemplo:

```env
APP_PUBLIC_URL=http://localhost:8080

SMTP_HOST=smtp.gmail.com
SMTP_PORT=465
SMTP_SECURE=true
SMTP_USER=tu-correo@gmail.com
SMTP_PASS=tu-app-password-sin-espacios
EMAIL_FROM="SENA IAM <tu-correo@gmail.com>"
EMAIL_DELIVERY_REQUIRED=true
```

Para probar recuperación, el correo debe existir como usuario dentro de `identity.user`. Si se escribe un correo que no está registrado, no se enviará correo aunque la pantalla muestre una respuesta exitosa.

## Cómo Ejecutar El Proyecto

Desde la raíz del proyecto:

```bash
docker-compose up --build
```

Luego abrir:

```text
http://localhost:8080
```

Backend directo:

```text
http://localhost:3001/api/v1
```

Health check:

```text
http://localhost:3001/health
```

## Usuarios Demo

Todos los usuarios demo usan la contraseña:

```text
Demo2026!
```

Usuarios disponibles:

```text
admin@sena.edu.co
director@sena.edu.co
coordinador@sena.edu.co
instructor@sena.edu.co
aprendiz@sena.edu.co
administrativo@sena.edu.co
```

Para probar correos reales, se recomienda crear un usuario nuevo desde la pantalla de Usuarios con un correo real.

## Endpoints Principales

Autenticación:

```text
POST /api/v1/auth/login
POST /api/v1/auth/refresh
POST /api/v1/auth/logout
GET  /api/v1/auth/me
```

Recuperación de contraseña:

```text
POST /api/v1/auth/password-reset/request
POST /api/v1/auth/password-reset/confirm
```

Usuarios:

```text
GET  /api/v1/users
POST /api/v1/users
GET  /api/v1/users/{id}
PUT  /api/v1/users/{id}
POST /api/v1/users/{id}/deactivate
```

Roles y permisos:

```text
GET    /api/v1/roles
GET    /api/v1/modules
POST   /api/v1/users/{id}/roles
DELETE /api/v1/users/{id}/roles/{role_name}
```

Sesiones y auditoría:

```text
GET    /api/v1/users/{id}/sessions
DELETE /api/v1/users/{id}/sessions/{session_id}
GET    /api/v1/reports/login-audit
```

## Guion Sugerido Para El Video

1. Mostrar la estructura del proyecto.
   Explicar que hay backend, frontend, base de datos y Docker Compose.

2. Levantar el proyecto.
   Ejecutar `docker-compose up --build` y abrir `http://localhost:8080`.

3. Mostrar el login.
   Iniciar sesión con `admin@sena.edu.co` y `Demo2026!`.

4. Explicar el dashboard.
   Mostrar usuario autenticado, roles, módulos y features.

5. Mostrar gestión de usuarios.
   Crear un usuario, editarlo, asignarle rol y revisar su detalle.

6. Mostrar RBAC.
   Explicar roles, módulos y features. Recalcar que el token contiene permisos calculados.

7. Mostrar sesiones.
   Explicar que el sistema guarda refresh tokens y permite revocar sesiones.

8. Mostrar auditoría.
   Enseñar intentos de login y explicar que se registran éxitos, fallos y bloqueos.

9. Mostrar recuperación de contraseña.
   Crear o usar un usuario con correo real, solicitar recuperación y revisar el correo recibido.

10. Cerrar explicando seguridad.
    Mencionar bcrypt, JWT RS256, refresh tokens, token de recuperación hasheado, SMTP y respuesta segura para no revelar usuarios.

## Puntos Clave Para Explicar

- El frontend no guarda contraseñas, solo envía credenciales al backend.
- Las contraseñas se almacenan hasheadas con bcrypt.
- Los tokens de recuperación no se guardan en texto plano, se guarda su hash.
- El JWT lleva roles y features para que otros servicios puedan validar permisos.
- Los refresh tokens permiten mantener sesión y también pueden revocarse.
- La auditoría permite rastrear intentos de acceso.
- SMTP permite enviar correos reales para recuperación y bienvenida.

## Estado Actual

El proyecto ya tiene funcional:

- Backend Go.
- Frontend web.
- Base de datos PostgreSQL con Liquibase.
- Login y refresh token.
- Administración de usuarios.
- RBAC por roles, features y scopes.
- Auditoría de login.
- Recuperación de contraseña por correo.
- Docker Compose para ejecución local.
