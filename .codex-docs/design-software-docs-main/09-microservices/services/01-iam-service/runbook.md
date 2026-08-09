# Runbook — iam-service

> Estado: 🟢 Estable | Última actualización: 2026-06-20
> Puerto: 8001 | SLO: 99.9 % uptime | Criticidad: P0 (servicio de entrada — ningún otro servicio autentica sin él)

---

## Healthcheck

| Endpoint | Método | Respuesta esperada | Cuándo falla |
|----------|--------|--------------------|--------------|
| `GET /health` | GET | `200 { "status": "ok" }` | Proceso caído o sin red |
| `GET /health/ready` | GET | `200 { "status": "ready", "db": "ok", "redis": "ok" }` | BD o Redis no conectan |

**SLO**: `GET /health` debe responder en < 200 ms (p99). `GET /health/ready` en < 500 ms (p99).

`/health` responde aunque BD y Redis estén caídos — indica que el proceso vive.
`/health/ready` devuelve `503` si la BD no conecta. Si solo Redis falla, devuelve `200` con `"redis": "degraded"` porque el servicio sigue autenticando (sin revocación de tokens).

---

## Variables de entorno requeridas

| Variable | Tipo | Ejemplo | Descripción |
|----------|------|---------|-------------|
| `JWT_SECRET` | string | `s3cr3t-256-bits` | Clave HMAC-SHA256 para firma de tokens. Mínimo 32 caracteres. Rotar sin downtime según procedimiento de rotación. |
| `JWT_EXPIRY_MINUTES` | int | `15` | Vida útil del access token. Recomendado: 15 min para producción. |
| `DB_URL` | string | `postgresql://iam:pass@db:5432/iam_db` | Connection string a `iam_db` en PostgreSQL. Debe incluir pool params en producción (`?pool_max=10`). |
| `REDIS_URL` | string | `redis://redis:6379/0` | URL de Redis usado como lista de revocación de refresh tokens. Si no está disponible el servicio opera en modo degradado. |
| `BCRYPT_COST` | int | `12` | Factor de costo para hashing de contraseñas. Valor mínimo: 12. Aumentar solo en hardware con mayor capacidad de cómputo; impacta latencia de login. |
| `MAX_LOGIN_ATTEMPTS` | int | `5` | Intentos fallidos antes del bloqueo corto (15 min). A los 10 intentos el bloqueo es 24 h. Ver reglas en `data-model.md`. |

> Variables adicionales recomendadas (no obligatorias al arrancar): `LOG_LEVEL`, `PORT` (default `8001`), `SEED_ADMIN_EMAIL`, `SEED_ADMIN_PASSWORD`.

---

## Alertas críticas

| Alerta | Condición de disparo | Severidad | Impacto si no se atiende | Acción inmediata |
|--------|---------------------|-----------|--------------------------|------------------|
| **BD no responde** | Timeout > 5 s al conectar a `iam_db` o `GET /health/ready` devuelve `503` | P0 | Ningún usuario puede autenticarse. Todo el sistema queda bloqueado. | 1. Verificar estado del pod PostgreSQL. 2. Revisar logs de BD. 3. Si la BD está healthy, verificar `DB_URL` y red interna. 4. Si persiste > 5 min, escalar a infra. |
| **Redis no responde** | Timeout > 2 s al conectar a Redis o `GET /health/ready` devuelve `"redis": "degraded"` | P1 | El servicio sigue autenticando. Los refresh tokens emitidos durante la caída **no podrán revocarse** hasta que Redis vuelva. Alertar. | 1. Verificar estado del pod Redis. 2. Revisar `REDIS_URL`. 3. Monitorear ventana de tokens no-revocables. 4. Si Redis no vuelve en 30 min, evaluar invalidación forzada de todas las sesiones activas en BD (`is_revoked = true` masivo). |
| **Error rate > 5 % en 5 min** | Más del 5 % de requests a `/auth/*` retornan 5xx en ventana deslizante de 5 min | P1 | Degradación del flujo de autenticación para usuarios activos. | 1. Revisar logs de errores (`grep "ERROR\|FATAL"` en stdout del contenedor). 2. Verificar conectividad a BD y Redis. 3. Si el error es de pool agotado, escalar réplicas o reiniciar. |
| **JWT_SECRET necesita rotación** | El secreto lleva > 90 días sin rotar, o existe sospecha de compromiso | P0 | Si el secreto está comprometido, un atacante puede emitir tokens válidos arbitrarios para cualquier rol. | Ver procedimiento completo en **Rotación de JWT_SECRET** más abajo. No reiniciar el servicio sin seguir ese procedimiento; un reinicio simple invalida todas las sesiones activas. |
| **Account lockout masivo** | > 20 cuentas bloqueadas (`locked_until IS NOT NULL`) en un intervalo de 10 min | P2 | Posible ataque de fuerza bruta distribuido o credential stuffing. Usuarios legítimos pueden quedar bloqueados. | 1. Revisar `audit_login` por IP de origen. 2. Evaluar bloqueo de IP a nivel de red (firewall / ingress). 3. Notificar al equipo de seguridad. 4. Si las cuentas afectadas son de roles críticos, desbloquear manualmente (ver procedimiento). |

---

## Procedimientos comunes

### Reinicio del servicio

```bash
# Docker Compose
docker compose restart iam-api

# Verificar que levantó correctamente
docker compose logs --tail=50 iam-api

# Kubernetes — rollout restart (sin downtime si hay >= 2 réplicas)
kubectl rollout restart deployment/iam-api -n <namespace>

# Verificar estado del rollout
kubectl rollout status deployment/iam-api -n <namespace>

# Si el rollout se cuelga, inspeccionar el pod nuevo
kubectl describe pod -l app=iam-api -n <namespace> | tail -30
```

> Nota: un reinicio simple **no invalida** los JWT access tokens existentes (su vida útil es de `JWT_EXPIRY_MINUTES`). Los refresh tokens siguen válidos en BD. Solo se pierden las entradas en Redis, lo que impide revocación de tokens emitidos antes del reinicio hasta que Redis vuelva a tener los datos sincronizados.

---

### Revisión de logs de autenticación fallida

```bash
# Docker Compose — últimas 200 líneas filtrando errores de auth
docker compose logs --tail=200 iam-api | grep -E "INVALID_PASSWORD|USER_NOT_FOUND|ACCOUNT_LOCKED|TOKEN_EXPIRED"

# Kubernetes — logs del pod en ejecución
kubectl logs -l app=iam-api -n <namespace> --tail=200 | grep -E "INVALID_PASSWORD|USER_NOT_FOUND|ACCOUNT_LOCKED|TOKEN_EXPIRED"

# Ver intentos fallidos por IP (últimos 30 min) directamente en BD
psql "$DB_URL" -c "
  SELECT
    ip_address,
    COUNT(*)                   AS intentos,
    COUNT(DISTINCT email_attempted) AS emails_distintos,
    MAX(attempted_at)          AS ultimo_intento
  FROM audit_login
  WHERE
    outcome != 'SUCCESS'
    AND attempted_at > NOW() - INTERVAL '30 minutes'
  GROUP BY ip_address
  ORDER BY intentos DESC
  LIMIT 20;
"

# Ver intentos fallidos sobre una cuenta específica
psql "$DB_URL" -c "
  SELECT outcome, ip_address, user_agent, attempted_at
  FROM audit_login
  WHERE email_attempted = 'usuario@ejemplo.com'
  ORDER BY attempted_at DESC
  LIMIT 50;
"
```

---

### Desbloquear cuenta de usuario manualmente

Usar solo cuando el bloqueo es por falsos positivos confirmados, o luego de validar la identidad del usuario por canal alternativo.

```sql
-- 1. Verificar estado actual de la cuenta
SELECT id, email, failed_attempts, locked_until, is_active
FROM "user"
WHERE email = 'usuario@ejemplo.com';

-- 2. Desbloquear: resetear contador y quitar bloqueo temporal
UPDATE "user"
SET
  failed_attempts = 0,
  locked_until    = NULL,
  updated_at      = NOW()
WHERE email = 'usuario@ejemplo.com'
  AND is_active = true;

-- 3. Confirmar el cambio
SELECT id, email, failed_attempts, locked_until
FROM "user"
WHERE email = 'usuario@ejemplo.com';
```

> Si `is_active = false`, la cuenta está deshabilitada (no es un bloqueo por intentos). Para reactivarla se necesita autorización del rol `SYSTEM_ADMIN` o `CENTER_DIRECTOR` con scope sobre el centro del usuario.

---

### Revocar todas las sesiones de un usuario

Usar cuando se sospecha que las credenciales de un usuario fueron comprometidas, o cuando se desactiva una cuenta y se quiere terminar sesiones activas inmediatamente.

```sql
-- 1. Revocar todos los refresh tokens activos en BD
UPDATE refresh_token
SET
  is_revoked = true,
  revoked_at = NOW()
WHERE
  user_id = (SELECT id FROM "user" WHERE email = 'usuario@ejemplo.com')
  AND is_revoked = false
  AND expires_at > NOW();

-- 2. Verificar cuántos tokens fueron revocados
SELECT COUNT(*) AS tokens_revocados
FROM refresh_token
WHERE
  user_id = (SELECT id FROM "user" WHERE email = 'usuario@ejemplo.com')
  AND is_revoked = true
  AND revoked_at > NOW() - INTERVAL '1 minute';
```

```bash
# 3. (Opcional) Si Redis está disponible, forzar la entrada en la lista de revocación
#    para que el access token actual también quede bloqueado antes de que expire.
#    Reemplazar <user_id> y <jti> por los valores reales del JWT.
redis-cli -u "$REDIS_URL" SET "revoked:<jti>" "1" EX 900
```

> Los access tokens vigentes siguen siendo válidos hasta su expiración natural (`JWT_EXPIRY_MINUTES`) a menos que se use la lista de revocación en Redis. Con `JWT_EXPIRY_MINUTES=15`, el riesgo máximo de exposición después de revocar en BD es de 15 minutos.

---

## Escenarios de falla y recuperación

### BD no conecta

**Síntomas**: `GET /health/ready` devuelve `503`. Todos los logins retornan `503`. Los logs muestran `connection refused` o `timeout` a PostgreSQL.

**Pasos**:

1. Verificar que el pod/contenedor de PostgreSQL está corriendo:
   ```bash
   # Docker Compose
   docker compose ps db
   docker compose logs --tail=50 db

   # Kubernetes
   kubectl get pods -l app=postgres -n <namespace>
   kubectl describe pod <postgres-pod> -n <namespace>
   ```

2. Probar conectividad directa desde el pod de IAM:
   ```bash
   # Kubernetes
   kubectl exec -it <iam-pod> -n <namespace> -- \
     pg_isready -h <db-host> -p 5432 -U iam
   ```

3. Si PostgreSQL está healthy pero IAM no conecta, revisar `DB_URL` y secretos del entorno:
   ```bash
   kubectl exec -it <iam-pod> -n <namespace> -- env | grep DB_URL
   ```

4. Si la BD está caída: seguir el runbook de PostgreSQL del equipo de infra. IAM no requiere acciones adicionales — al volver la BD, el pool de conexiones se recupera automáticamente.

5. Si la BD vuelve pero IAM no reconecta en 60 s: reiniciar el pod de IAM (el pool no siempre se recupera solo con conexiones rotas a nivel TCP).

6. Verificar `GET /health/ready` después de la recuperación.

**Tiempo objetivo de recuperación (RTO)**: < 15 min si la causa es reinicio de BD. > 15 min escalar a infra.

---

### Redis no conecta

**Síntomas**: `GET /health/ready` devuelve `200` con `"redis": "degraded"`. Alerta P1 disparada. Login y refresh siguen funcionando.

**Comportamiento en modo degradado**:
- Login (`POST /auth/login`): funciona normalmente.
- Refresh (`POST /auth/refresh`): funciona. El nuevo refresh token se persiste en BD.
- Logout (`POST /auth/logout`): el refresh token se marca `is_revoked = true` en BD, pero el access token activo **no se agrega a la lista de revocación de Redis**. Un access token no-expirado seguirá siendo aceptado por los servicios downstream hasta que expire.
- `POST /auth/revoke`: misma limitación.

**Pasos**:

1. Verificar estado de Redis:
   ```bash
   docker compose ps redis
   redis-cli -u "$REDIS_URL" PING
   ```

2. Si Redis está caído sin causa obvia, reiniciarlo:
   ```bash
   docker compose restart redis
   # o en Kubernetes:
   kubectl rollout restart deployment/redis -n <namespace>
   ```

3. Monitorear el tiempo sin revocación. Si supera 30 min y hay sesiones de usuarios de alto privilegio (`SYSTEM_ADMIN`, `CENTER_DIRECTOR`) activas durante la ventana, considerar invalidación forzada masiva de refresh tokens en BD para obligar re-login:
   ```sql
   UPDATE refresh_token
   SET is_revoked = true, revoked_at = NOW()
   WHERE is_revoked = false AND expires_at > NOW();
   ```

4. Al volver Redis, IAM retoma la escritura a la lista de revocación automáticamente. No requiere reinicio.

---

### JWT_SECRET comprometido — Procedimiento de rotación

> Este procedimiento invalida **todas las sesiones activas** de todos los usuarios. Coordinarlo con el equipo antes de ejecutar.

**Por qué no basta con reiniciar**: cambiar `JWT_SECRET` y reiniciar el servicio hace que todos los tokens firmados con el secreto anterior sean inmediatamente rechazados. Pero durante la ventana de rotación, los servicios downstream que cachean la clave pública (si se usa RS256) o que no recargan el secreto seguirán aceptando tokens viejos. Este procedimiento minimiza esa ventana.

**Pasos**:

```bash
# 1. Generar un nuevo secreto seguro (mínimo 32 bytes de entropía)
openssl rand -hex 32
# Guardar el valor — lo necesitarás en los pasos siguientes

# 2. Revocar todos los refresh tokens activos en BD
#    Esto fuerza re-login a todos los usuarios cuando el access token expire.
psql "$DB_URL" -c "
  UPDATE refresh_token
  SET is_revoked = true, revoked_at = NOW()
  WHERE is_revoked = false AND expires_at > NOW();
"

# 3. Actualizar el secreto en el gestor de secretos (Vault / K8s Secret / .env)
#    Ejemplo con kubectl:
kubectl create secret generic iam-secrets \
  --from-literal=JWT_SECRET='<nuevo-secreto>' \
  --namespace <namespace> \
  --dry-run=client -o yaml | kubectl apply -f -

# 4. Reiniciar el servicio para que tome el nuevo secreto
kubectl rollout restart deployment/iam-api -n <namespace>
kubectl rollout status deployment/iam-api -n <namespace>

# 5. Verificar que el servicio arrancó y emite tokens con el nuevo secreto
curl -s http://localhost:8001/health/ready | jq .

# 6. Verificar que un login produce un token válido
curl -s -X POST http://localhost:8001/auth/login \
  -H 'Content-Type: application/json' \
  -d '{"email":"admin@sena.edu.co","password":"<pass>"}' | jq .access_token

# 7. Registrar el incidente: fecha de rotación, motivo, quién ejecutó
```

**Después de la rotación**: todos los usuarios verán que su sesión expiró y deberán hacer login nuevamente. Los access tokens con el secreto anterior quedarán huérfanos y serán rechazados automáticamente (firma inválida). La ventana de tokens "zombies" aceptados por servicios downstream es de máximo `JWT_EXPIRY_MINUTES` minutos si esos servicios no validan contra IAM (lo cual es el diseño intencionado — los servicios validan la firma localmente).

---

## Escalamiento

| Condición | Paso siguiente | Contacto |
|-----------|---------------|----------|
| BD no vuelve en 15 min después de reinicio | Escalar a equipo de infra con logs de PostgreSQL adjuntos | @infra-oncall |
| Redis no vuelve en 30 min | Evaluar impacto de sesiones sin revocación; escalar a infra | @infra-oncall |
| Error rate > 5 % sostenido > 10 min sin causa clara | Escalar a tech lead con payload de logs y métricas de los últimos 15 min | @tech-lead |
| JWT_SECRET comprometido (sospecha o confirmado) | Activar protocolo de incidente de seguridad P0 **antes** de cualquier acción técnica | @security-lead, @tech-lead |
| Account lockout masivo (> 20 cuentas en 10 min) | Notificar a seguridad; no desbloquear cuentas hasta que seguridad confirme que no es ataque activo | @security-lead |
| Servicio no levanta después de 2 reinicios | Revisar variables de entorno, secretos y logs de arranque; escalar a tech lead si no se identifica causa | @tech-lead |
