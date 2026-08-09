# Eventos — document-service

> Estado: 🟢 Estable | Ultima actualizacion: 2026-06-20

---

## Eventos publicados

Topic: `document-events`

---

### `document.document.generated`

Se publica cuando un documento ha sido generado exitosamente y su binario esta disponible en el almacenamiento de objetos.

**Payload:**

```json
{
  "document_id": "doc_01HXYZ123456",
  "template_code": "schedule-pdf-v2",
  "domain": "academic",
  "owner_entity_id": "student_00987",
  "owner_service": "academic-service",
  "storage_key": "documents/academic/2026/06/doc_01HXYZ123456.pdf",
  "mime_type": "application/pdf",
  "size_bytes": 204800,
  "generated_by": "pdf-renderer-worker",
  "generated_at": "2026-06-20T14:32:00Z"
}
```

| Campo | Tipo | Descripcion |
|---|---|---|
| `document_id` | `string` | Identificador unico del documento generado |
| `template_code` | `string` | Codigo de la plantilla Handlebars utilizada |
| `domain` | `string` | Dominio de negocio al que pertenece el documento |
| `owner_entity_id` | `string` | ID de la entidad propietaria del documento (estudiante, grupo, etc.) |
| `owner_service` | `string` | Servicio que solicito la generacion |
| `storage_key` | `string` | Ruta del binario en MinIO/S3 |
| `mime_type` | `string` | Tipo MIME del documento generado |
| `size_bytes` | `integer` | Tamano del archivo en bytes |
| `generated_by` | `string` | Componente que ejecuto la generacion |
| `generated_at` | `string (ISO 8601)` | Marca de tiempo de finalizacion de la generacion |

**Consumidores conocidos:** `audit-service`

---

### `document.version.created`

Se publica cuando se crea una nueva version de un documento existente (actualizacion de plantilla, regeneracion por cambio de datos).

**Payload:**

```json
{
  "version_id": "ver_01HABCDEF789",
  "document_id": "doc_01HXYZ123456",
  "version_number": 2,
  "storage_key": "documents/academic/2026/06/doc_01HXYZ123456_v2.pdf",
  "created_by": "pdf-renderer-worker",
  "created_at": "2026-06-20T15:10:00Z"
}
```

| Campo | Tipo | Descripcion |
|---|---|---|
| `version_id` | `string` | Identificador unico de la version |
| `document_id` | `string` | Referencia al documento padre |
| `version_number` | `integer` | Numero de version (incremental, comienza en 1) |
| `storage_key` | `string` | Ruta del binario de esta version en MinIO/S3 |
| `created_by` | `string` | Componente o usuario que origino la nueva version |
| `created_at` | `string (ISO 8601)` | Marca de tiempo de creacion de la version |

**Consumidores conocidos:** `audit-service`

---

## Eventos consumidos

---

### `scheduling.schedule.published`

**Origen:** `scheduling-service`  
**Topic:** `scheduling-events`

Cuando el horario academico de un periodo es publicado y queda en estado definitivo, `document-service` genera automaticamente el PDF de horario para cada grupo o estudiante afectado.

**Flujo:**
1. Se recibe el evento con los datos del horario publicado.
2. Se selecciona la plantilla `schedule-pdf-v{n}` del repositorio de plantillas Handlebars.
3. Se crea un registro de documento con `status=GENERATING` para cada entidad propietaria relevante.
4. El `pdf-renderer-worker` renderiza el PDF combinando la plantilla con los datos del horario.
5. El binario se almacena en MinIO/S3 bajo el prefijo `documents/scheduling/`.
6. El registro del documento pasa a `status=AVAILABLE`.
7. Se publica `document.document.generated`.

**Plantilla utilizada:** `schedule-pdf-v2`

---

### `academic.ficha.opened`

**Origen:** `academic-service`  
**Topic:** `academic-events`

Cuando se abre una ficha de matricula para un estudiante, `document-service` genera el comprobante de matricula (enrollment record) en formato PDF.

**Flujo:**
1. Se recibe el evento con el ID del estudiante y los datos de la ficha.
2. Se selecciona la plantilla `enrollment-record-v{n}` del repositorio de plantillas Handlebars.
3. Se crea un registro de documento con `status=GENERATING` vinculado al `owner_entity_id` del estudiante.
4. El `pdf-renderer-worker` renderiza el PDF con los datos academicos del estudiante.
5. El binario se almacena en MinIO/S3 bajo el prefijo `documents/academic/`.
6. El registro del documento pasa a `status=AVAILABLE`.
7. Se publica `document.document.generated`.

**Plantilla utilizada:** `enrollment-record-v1`

---

## Formato de envelope

Todos los eventos siguen el envelope estandar de la plataforma:

```json
{
  "specversion": "1.0",
  "type": "document.document.generated",
  "source": "document-service",
  "id": "evt_01HXYZ000001",
  "time": "2026-06-20T14:32:00Z",
  "datacontenttype": "application/json",
  "data": {
    "document_id": "doc_01HXYZ123456",
    "template_code": "schedule-pdf-v2",
    "domain": "academic",
    "owner_entity_id": "student_00987",
    "owner_service": "academic-service",
    "storage_key": "documents/academic/2026/06/doc_01HXYZ123456.pdf",
    "mime_type": "application/pdf",
    "size_bytes": 204800,
    "generated_by": "pdf-renderer-worker",
    "generated_at": "2026-06-20T14:32:00Z"
  }
}
```

| Campo | Descripcion |
|---|---|
| `specversion` | Version del estandar CloudEvents (siempre `1.0`) |
| `type` | Nombre completo del evento |
| `source` | Servicio emisor |
| `id` | Identificador unico del evento (ULID) |
| `time` | Timestamp de emision en ISO 8601 UTC |
| `datacontenttype` | Siempre `application/json` |
| `data` | Payload especifico del evento |

---

## Politica de reintentos

La generacion de PDFs puede fallar por errores transitorios en el renderizador, indisponibilidad de MinIO/S3, o datos de plantilla malformados. La politica aplica al `pdf-renderer-worker`.

| Parametro | Valor |
|---|---|
| Maximo de intentos | 3 |
| Delay entre intentos | 60 segundos (fijo) |
| Comportamiento en fallo definitivo | `status=GENERATION_FAILED` + alerta |

**Flujo de reintentos:**

```
Intento 1 → falla → esperar 60s
Intento 2 → falla → esperar 60s
Intento 3 → falla → marcar documento status=GENERATION_FAILED
                   → emitir alerta a canal de operaciones
                   → NO se publica document.document.generated
```

**Errores que activan reintento:**
- Timeout de renderizado (umbral: 30s por pagina, max 300s total)
- Error de conexion a MinIO/S3 al subir el binario
- Error 5xx del motor de renderizado interno

**Errores que NO activan reintento (fallo inmediato):**
- Plantilla no encontrada (`template_code` invalido)
- Datos de entrada con schema invalido
- Error 4xx del motor de renderizado interno

**Acciones post-fallo definitivo:**
- El campo `status` del registro de documento se actualiza a `GENERATION_FAILED`.
- Se registra el motivo del fallo en el campo `failure_reason`.
- Se envia una alerta al canal de monitoreo de operaciones (configurado via variable de entorno `ALERT_WEBHOOK_URL`).
- El evento que origino la generacion queda en estado `dead-letter` para revision manual.

---

## Proceso de generacion asincrona

El `document-service` nunca genera documentos de forma sincrona. Todo flujo de generacion pasa por el `pdf-renderer-worker` via cola interna.

```
Evento entrante (p.ej. scheduling.schedule.published)
        |
        v
[document-lifecycle-worker] recibe el evento del topic externo
        |
        v
[document-api / base de datos] crea registro de documento
        status = GENERATING
        document_id asignado
        template_code resuelto segun tipo de evento
        |
        v
Mensaje encolado en `document-generation-queue`
        { document_id, template_code, data_payload }
        |
        v
[pdf-renderer-worker] toma el mensaje de la cola
        |
        ├── renderiza la plantilla Handlebars con los datos
        ├── genera el PDF
        └── sube el binario a MinIO/S3
        |
        v
[pdf-renderer-worker] actualiza el registro del documento
        status = AVAILABLE
        storage_key = ruta definitiva en MinIO/S3
        size_bytes, mime_type registrados
        |
        v
[document-lifecycle-worker] publica evento de salida
        document.document.generated → topic: document-events
```

**Componentes involucrados:**

| Componente | Puerto / Rol | Responsabilidad en el flujo |
|---|---|---|
| `document-api` | 8007 | Expone REST para consulta de documentos y metadatos |
| `template-api` | 8007 (sub-ruta `/templates`) | Administra plantillas HTML/Handlebars |
| `document-lifecycle-worker` | worker interno | Consume eventos externos; orquesta ciclo de vida del documento |
| `pdf-renderer-worker` | worker interno | Renderiza plantillas y sube binarios a MinIO/S3 |

**Estados del ciclo de vida del documento:**

| Estado | Descripcion |
|---|---|
| `GENERATING` | Registro creado, renderizado en progreso |
| `AVAILABLE` | PDF generado y disponible en almacenamiento |
| `GENERATION_FAILED` | Todos los reintentos agotados; requiere intervencion manual |
| `SUPERSEDED` | Version mas reciente disponible; esta version es historica |
