# Eventos — audit-service

> Estado: 🟢 Estable | Última actualización: 2026-06-20

---

## Eventos publicados

`audit-service` no publica eventos. Es el sumidero final del sistema de auditoría. Publicar eventos desde el `audit-service` crearía dependencias circulares y violaría su invariante de solo-escritura.

---

## Eventos consumidos

El `audit-worker` consume todos los topics del sistema mediante suscripción wildcard. La acción en todos los casos es idéntica: persistir el evento como un registro inmutable.

| Topic            | Servicio fuente                  | Acción                                      |
|------------------|----------------------------------|---------------------------------------------|
| `iam-events`     | `iam-service`                    | Persiste cada evento como `audit_record` append-only. |
| `reference-data-events` | `reference-data-service`  | Persiste cada evento como `audit_record` append-only. |
| `academic-events` | `academic-management-service`  | Persiste cada evento como `audit_record` append-only. |
| `environment-events` | `training-environment-service` | Persiste cada evento como `audit_record` append-only. |
| `scheduling-events` | `scheduling-service`          | Persiste cada evento como `audit_record` append-only. |
| `actors-events`  | `actors-service`                 | Persiste cada evento como `audit_record` append-only. |
| `document-events` | `document-service`              | Persiste cada evento como `audit_record` append-only. |
| `monitoring-events` | `monitoring-service`          | Persiste cada evento como `audit_record` append-only. |

> El `audit-worker` no filtra por tipo de evento dentro de cada topic. Todo mensaje recibido se almacena íntegro, incluyendo headers, payload y metadatos del broker.

---

## Suscripción wildcard

El componente `audit-worker` (puerto 8009) se suscribe a todos los topics que coincidan con el patrón:

```
*-events
```

Cada mensaje consumido dispara exactamente un `INSERT INTO audit_record` con el payload completo del evento. No existe lógica de enrutamiento ni transformación: el `audit-worker` es intencionalmente simple para maximizar la fiabilidad del registro.

Flujo de procesamiento por mensaje:

1. El broker entrega un mensaje desde cualquier topic que coincida con `*-events`.
2. El `audit-worker` extrae `event_id`, `topic`, `timestamp`, `payload` y metadatos del broker.
3. Se ejecuta `INSERT INTO audit_record … ON CONFLICT (event_id) DO NOTHING`.
4. El offset del broker se confirma (commit) independientemente de si el INSERT insertó o fue ignorado por conflicto.

---

## Garantías de idempotencia

Si el broker re-entrega un evento (semántica at-least-once), el segundo `INSERT` falla silenciosamente por la restricción `UNIQUE(event_id)`. No se pierden eventos ni se duplican.

La columna `event_id` corresponde al identificador único generado por el servicio publicador en el momento de la emisión. El `audit-service` no genera ni reasigna identificadores propios.

```sql
-- Restricción que garantiza idempotencia
ALTER TABLE audit_record ADD CONSTRAINT uq_audit_record_event_id UNIQUE (event_id);
```

Esta garantía cubre los siguientes escenarios:

| Escenario                                   | Resultado                          |
|---------------------------------------------|------------------------------------|
| Entrega exactamente una vez                 | INSERT exitoso, registro creado.   |
| Re-entrega tras fallo del consumer          | INSERT ignorado, offset confirmado. |
| Re-entrega tras fallo de red                | INSERT ignorado, offset confirmado. |
| Reprocessing manual desde offset pasado     | INSERT ignorado para eventos ya persistidos. |

---

## Retención

- **Mínimo legal**: 7 años desde la fecha de emisión del evento.
- **Particionamiento**: mensual recomendado por columna `emitted_at` (ver `data-model.md`).
- **Archivado a cold storage**: después de 2 años en almacenamiento activo, las particiones se mueven a almacenamiento en frío (e.g., S3 Glacier, Azure Archive).
- **Inmutabilidad**: ninguna fila de `audit_record` puede ser modificada ni eliminada antes de que expire su período de retención obligatorio. Las operaciones `UPDATE` y `DELETE` deben estar bloqueadas a nivel de política de base de datos o rol.
- **Acceso a cold storage**: bajo demanda mediante proceso de restauración documentado en `runbooks/audit-cold-restore.md`.
