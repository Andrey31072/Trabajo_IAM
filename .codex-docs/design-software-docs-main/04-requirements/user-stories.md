# Historias de Usuario

> Estado: 🟡 En progreso | Última actualización: 2026-08-01
> Autor: Jesús Ariel González Bonilla (PM + Arquitecto) | Equipo: Producto

> Fuente: [functional.md](./functional.md) · [02-domain/entities-and-rules.md](../02-domain/entities-and-rules.md) · [03-product/problem-framing.md](../03-product/problem-framing.md)

## Cómo leer este documento

Las historias se numeran `HU-##` y se agrupan por **rol**: Coordinador académico, Instructor, Aprendiz y Administrador/Director. Cada historia sigue el formato *Como… quiero… para…* con **criterios de aceptación** verificables, anclados a las reglas de negocio (RN-*) de [entities-and-rules.md](../02-domain/entities-and-rules.md) y a los requisitos funcionales (RF-*) de [functional.md](./functional.md). La trazabilidad completa está en [traceability-matrix.md](./traceability-matrix.md).

Roles (RBAC): `COORDINATOR`, `INSTRUCTOR`, `LEARNER`, `ADMIN_STAFF`/`CENTER_DIRECTOR`/`SYSTEM_ADMIN`.

---

## Rol: Coordinador académico

El coordinador es el usuario primario del MVP: construye y publica los horarios de su centro.

### HU-01 — Registrar una ficha de caracterización
**Como** coordinador **quiero** registrar una ficha con su número SENA, programa y cupo **para** poder construirle un horario. *(RF-ACAD-02)*

Criterios de aceptación:
- Puedo asociar la ficha a un programa existente y fijar su `max_capacity`.
- El `ficha_number` es único; el sistema rechaza duplicados (RN-ACAD-03).
- El `max_capacity` no puede aumentarse después de creado (RN-ACAD-06).

### HU-02 — Cambiar el estado de una ficha
**Como** coordinador **quiero** transicionar una ficha (p. ej. `INDUCTION → EXECUTION`) **para** habilitar la construcción de su horario. *(RF-ACAD-03)*

Criterios de aceptación:
- Solo puedo ejecutar transiciones autorizadas para mi rol (RN-ACAD-05).
- Al pasar a `EXECUTION` se emite `academic.ficha.opened`.
- No puedo cancelar una ficha si esa transición está reservada al director (RN-ACAD-05).

### HU-05 — Consultar ambientes disponibles
**Como** coordinador **quiero** ver los ambientes libres para una franja y fecha **para** elegir dónde ubicar una sesión. *(RF-ENV-04)*

Criterios de aceptación:
- La consulta responde en < 300 ms (p95) (NFR-PERF-01).
- La lista descuenta ambientes en mantenimiento y ya reservados (RN-ENV-03).
- No aparecen ambientes cuya `capacity` sea menor que el cupo de la ficha (RN-ENV-01).

### HU-06 — Consultar instructores disponibles
**Como** coordinador **quiero** ver los instructores habilitados y libres para una competencia y franja **para** asignarlos a una sesión. *(RF-ACT-03)*

Criterios de aceptación:
- La consulta responde en < 300 ms (p95) (NFR-PERF-02).
- Solo aparecen instructores con `CompetencyAssignment` activa para esa competencia (RN-ACT-02).
- No aparecen instructores con excepción de disponibilidad en esa franja (RN-ACT-02).

### HU-07 — Crear un borrador de horario
**Como** coordinador **quiero** crear un borrador de horario para una ficha en ejecución **para** empezar a agregarle sesiones. *(RF-SCH-01)*

Criterios de aceptación:
- Solo puedo crear el borrador si la ficha está en `EXECUTION` (RN-ACAD-04).
- Puedo tener varios `DRAFT` a la vez, pero un solo `PUBLISHED` por ficha/período (RN-SCH-02).
- El borrador queda ligado a mi centro (scope) (RN-IAM-03).

### HU-08 — Agregar sesiones de clase al borrador
**Como** coordinador **quiero** agregar sesiones (competencia + instructor + ambiente + franja + fecha) **para** ir armando el horario. *(RF-SCH-02, RF-SCH-04)*

Criterios de aceptación:
- La sesión dura entre 1 y 6 h y cae dentro de la disponibilidad del ambiente (RN-SCH-05).
- No puedo crear sesiones en domingo ni en festivos (RN-SCH-06).
- Al guardar, el sistema ejecuta la validación inmediata de conflictos (RN-SCH-04).

### HU-09 — Ver y resolver conflictos antes de publicar
**Como** coordinador **quiero** que el sistema detecte y liste los conflictos del horario **para** corregirlos antes de publicar. *(RF-SCH-03)*

Criterios de aceptación:
- El sistema clasifica cada conflicto (`INSTRUCTOR_DOUBLE_BOOKED`, `ENVIRONMENT_DOUBLE_BOOKED`, `SESSIONS_OVERLAP`, `ENVIRONMENT_MAINTENANCE`, `INSTRUCTOR_UNAVAILABLE`) con una descripción legible (RN-SCH-03).
- `POST /schedules/{id}/validate` valida un horario de 20 sesiones en < 2 s (NFR-PERF-03).
- Mientras existan conflictos activos, el botón/acción de publicar permanece bloqueado (RN-SCH-04).

### HU-10 — Publicar el horario
**Como** coordinador **quiero** publicar el horario sin conflictos con una sola acción **para** oficializarlo y comunicarlo. *(RF-SCH-05)*

Criterios de aceptación:
- Solo puedo publicar si la validación completa no arroja conflictos (RN-SCH-04).
- El horario publicado queda **inmutable**; para cambiarlo debo crear una nueva versión, que archiva la anterior (RN-SCH-01).
- Al publicar se emite `scheduling.schedule.published` (inicia seguimiento, notifica actores, habilita PDF).

### HU-11 — Gestionar aprendices de una ficha
**Como** coordinador **quiero** matricular aprendices y actualizar su estado **para** reflejar quién está activo en la ficha. *(RF-ACT-04)*

Criterios de aceptación:
- Un aprendiz `DROPOUT`/`TRANSFERRED`/`GRADUATED` no aparece en el horario (RN-ACT-03).
- No puedo reducir el cupo por debajo del número de aprendices activos (RN-ACAD-06).

### HU-13 — Administrar competencias del instructor
**Como** coordinador **quiero** registrar qué competencias puede impartir cada instructor y su tipo de vinculación **para** que el motor los ofrezca correctamente. *(RF-ACT-01)*

Criterios de aceptación:
- Puedo asignar/retirar competencias (`CompetencyAssignment`) a un instructor (RN-ACT-02).
- El tipo de vinculación fija el límite de horas semana (RN-ACT-01).

### HU-20 — Consultar KPIs de seguimiento por ficha
**Como** coordinador **quiero** ver los KPIs de asistencia y avance de mis fichas **para** detectar desviaciones a tiempo. *(RF-MON-02)*

Criterios de aceptación:
- Veo asistencia acumulada frente al mínimo reglamentario (80 %) (RN-MON-02).
- Cada medición es un registro histórico (append-only), no se sobrescribe (RN-CROSS-01).

### HU-21 — Atender y resolver alertas
**Como** coordinador **quiero** recibir y resolver alertas de riesgo **para** intervenir sobre fichas y aprendices en riesgo. *(RF-MON-03, RF-MON-04)*

Criterios de aceptación:
- Recibo la alerta según el nivel (`INFO`→`CRITICAL`) y su acción asociada (RN-MON-04).
- Al resolver, queda registrado quién, cuándo y las notas; la alerta no se elimina (`is_resolved = true`) (RN-MON-05).

### HU-22 — Descargar el horario en PDF
**Como** coordinador **quiero** descargar el horario publicado en PDF **para** difundirlo o archivarlo. *(RF-DOC-01)*

Criterios de aceptación:
- El PDF se genera bajo demanda y se descarga con una URL firmada de expiración corta (ADR-003).
- El binario no se guarda en base de datos; solo la `storage_key` (ADR-003).

### HU-23 — Crear un plan de mejoramiento
**Como** coordinador **quiero** crear un plan de mejoramiento para un aprendiz en riesgo **para** dar seguimiento a su recuperación. *(RF-MON-05)*

Criterios de aceptación:
- Para una ficha en estado crítico, debo crear el plan dentro de 5 días hábiles o el sistema emite `IMPROVEMENT_PLAN_OVERDUE` (RN-MON-06).

---

## Rol: Instructor

### HU-14 — Registrar mi disponibilidad y excepciones
**Como** instructor **quiero** declarar mis franjas disponibles y mis excepciones **para** que no me asignen clases cuando no puedo. *(RF-ACT-02)*

Criterios de aceptación:
- Una excepción registrada hace que el motor detecte `INSTRUCTOR_UNAVAILABLE` en esa franja (RN-SCH-03).
- El cambio se refleja en `GET /instructors/available`.

### HU-15 — Consultar mi horario de la semana
**Como** instructor **quiero** consultar mi horario vigente por semana **para** conocer mi carga con anticipación. *(RF-SCH-06)*

Criterios de aceptación:
- Veo solo sesiones de horarios `PUBLISHED` (RN-SCH-01).
- La consulta responde en < 500 ms (p95) (NFR-PERF-04).
- Conozco mi horario con ≥ 48 h de anticipación (meta de producto).

---

## Rol: Aprendiz

### HU-12 — Consultar el horario de mi ficha
**Como** aprendiz **quiero** ver el horario vigente de mi ficha **para** saber cuándo y dónde tengo clase. *(RF-SCH-06)*

Criterios de aceptación:
- Veo el horario `PUBLISHED` de mi ficha, con ambiente e instructor por sesión.
- Si estoy `DROPOUT`/`TRANSFERRED`/`GRADUATED` no aparezco como asistente (RN-ACT-03).
- La vista funciona en móvil (responsive, ≥ 360 px) (NFR-USA-02).

---

## Rol: Administrador / Director de centro

### HU-16 — Iniciar sesión de forma segura
**Como** usuario **quiero** autenticarme con mis credenciales **para** acceder según mi rol. *(RF-IAM-01)*

Criterios de aceptación:
- Recibo `access_token` (15 min) y `refresh_token` (7 días) (RN-IAM-02).
- Tras 5 intentos fallidos la cuenta se bloquea 15 min; tras 10, 24 h (RN-IAM-01).
- En mi primer login debo cambiar la contraseña temporal (RN-IAM-05).

### HU-17 — Controlar el acceso por rol y centro
**Como** administrador **quiero** que cada usuario solo acceda a lo permitido por su feature y scope **para** proteger los datos entre centros. *(RF-IAM-03)*

Criterios de aceptación:
- Un acceso sin token devuelve 401; un acceso a otro centro devuelve 403 `SCOPE_VIOLATION` (NFR-SEC-01/02).
- Un coordinador solo opera dentro de su centro asignado (RN-IAM-03).

### HU-18 — Gestionar usuarios e instructores
**Como** administrador **quiero** crear usuarios y sus perfiles de actor **para** dar acceso al personal del centro. *(RF-IAM-02, RF-IAM-04)*

Criterios de aceptación:
- Al crear un instructor se crea de forma coordinada su usuario y su perfil en `actors-service` (RN-IAM-04).
- La contraseña temporal caduca en 72 h (RN-IAM-05).

### HU-19 — Administrar la jerarquía institucional y catálogos
**Como** administrador **quiero** mantener la jerarquía institucional y los catálogos del sistema **para** que el resto de módulos tenga datos de referencia consistentes. *(RF-REF-01, RF-REF-02)*

Criterios de aceptación:
- El `center_code` es inmutable y un centro pertenece a un único municipio (RN-REF-01/02).
- Solo `ADMIN_STAFF`/`SYSTEM_ADMIN` pueden editar catálogos; para el resto son solo lectura (RN-REF-03).

---

## Puntos abiertos

- **HU-11 / HU-12**: pendiente confirmar si un aprendiz puede pertenecer a más de una ficha activa (PA-04 en [vision.md](../03-product/vision.md)).
- Historias de **etapa productiva** (visitas a empresa, convenios) quedan fuera del MVP (ver [scope.md](../01-context/scope.md)).

## Referencias

- [functional.md](./functional.md) · [non-functional.md](./non-functional.md) · [traceability-matrix.md](./traceability-matrix.md)
- [02-domain/entities-and-rules.md](../02-domain/entities-and-rules.md)
- [03-product/product-backlog.md](../03-product/product-backlog.md)
