<!-- RESUMEN-EJECUTIVO
agente: Jesús Ariel González Bonilla (PM + Arquitecto)
capacidad: contrato de evento / mensajería async
fase: diseño
estado: accepted
dependencias_entrada: 09-microservices/event-catalog.md, la BD de audit-service
consumidores_siguientes: ninguno (sink terminal; no publica eventos)
tldr: Worker consumidor por wildcard que persiste cada evento del ecosistema como audit_record append-only.
decisiones_clave: Suscripción fan-out a todos los topics (no depende de nombres de evento específicos); append-only sin FKs; idempotencia por event_id UNIQUE (conflicto se ignora, no error).
halts_registrados: ninguno — al consumir por wildcard, el worker no depende de que cada evento esté registrado en event-catalog.md
-->

# Contrato — audit-worker

> Estado: 🟡 En progreso | Última actualización: 2026-08-06
> Autor: Jesús Ariel González Bonilla (PM + Arquitecto) | Equipo: Arquitectura

> **Fuente de verdad:** el inventario de eventos lo gobierna [`event-catalog.md`](../../../../event-catalog.md) y el envelope estándar [`_template/service/events.md`](../../../../_template/service/events.md). Ante discrepancia, esos documentos prevalecen sobre este contrato.

> **Diseño previsto — no implementado.** Contrato de mensajería, agnóstico de lenguaje.

## Eventos consumidos
Fan-out a **todos** los topics del broker (`#` / wildcard) — no hay una tabla cerrada de eventos: el worker no filtra por origen, audita todo evento publicado por cualquier servicio (ver inventario completo en [`event-catalog.md`](../../../../event-catalog.md)).

## Eventos producidos
Ninguno. El `audit-worker` es un **consumidor terminal** (sink): solo consume y persiste, no publica eventos (así lo confirma la nota del [catálogo](../../../../event-catalog.md#audit-service)).

## Envelope de evento consumido (estándar)
```json
{
  "event_id": "uuid-v4",
  "event_type": "<service>.<entity>.<action>",
  "version": "1.0",
  "timestamp": "2026-08-01T14:00:00Z",
  "source_service": "scheduling-service",
  "correlation_id": "uuid-v4",
  "actor_id": "uuid",
  "entity_type": "class_session",
  "entity_id": "uuid",
  "payload": { }
}
```

## Efecto (persistencia)
Inserta un `audit_record` extrayendo `event_type`, `source_service`, `actor_id`, `entity_type`, `entity_id` y `timestamp` para indexar, y conserva el `payload` completo como `JSONB` (schema-on-read).

## Reglas
- **Append-only:** solo `INSERT`. Sin `UPDATE`/`DELETE` ni endpoint de escritura externo.
- **Idempotencia:** `event_id` UNIQUE; conflicto → se ignora (no error).
- **Sin FKs:** el registro es autónomo (sink desacoplado entre servicios).

## Política de reintentos
Fallo de persistencia → reintento con backoff; agotado → **DLQ** `audit.dlq` (no se pierde el evento).
