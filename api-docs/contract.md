# Contrato — support-service (tickets)

Contrato HTTP que consumen los **canales** (p. ej. `whatsapp-agent`) y el back-office.
support-service es un bounded context cerrado: se accede **solo por este contrato**, nunca
por la base de datos ni compartiendo modelo.

- **Base URL (vía gateway):** `https://<gateway>/support_service` → `/api/v1/...`
- **Formato:** REST puro. Éxito con body directo; error en `application/problem+json` (RFC 7807).
- **Versionado:** por path (`/api/v1`).

## Autenticación y tenancy

Todo endpoint de negocio exige:

| Header | Obligatorio | Descripción |
|--------|-------------|-------------|
| `Authorization: Bearer <jwt>` | sí | JWT emitido por IAM. Sin él → `401`. |
| `X-Tenant-ID: <uuid>` | sí | Tenant del request, validado contra el JWT. Sin tenant válido → `401`. |
| `X-User-Role: system_admin` | no | Break-glass: acceso cross-tenant, auditado. Uso excepcional. |

Sin tenant válido **toda** operación de negocio se rechaza (fail-closed). El aislamiento entre
tenants lo garantiza Row-Level Security en la base.

## Modelo

**Ticket**

| Campo | Tipo | Notas |
|-------|------|-------|
| `id` | uuid | |
| `tenant_id` | uuid | |
| `canal` | string | `whatsapp` \| `web` \| `email` |
| `solicitante_nombre` | string | PII del cliente final |
| `solicitante_telefono` | string | PII del cliente final |
| `asunto` | string | |
| `prioridad` | string | `baja` \| `media` \| `alta` \| `urgente` (default `media`) |
| `estado` | string | `abierto` \| `asignado` \| `en_curso` \| `resuelto` \| `cerrado` |
| `asignado_a` | uuid \| null | operador; `null` hasta asignar |
| `created_at` / `updated_at` | RFC 3339 | |

**Ciclo de vida (lineal):**

```
abierto ──asignar──▶ asignado ──tomar──▶ en_curso ──resolver──▶ resuelto ──cerrar──▶ cerrado
```

Toda transición fuera de este orden → `409 Conflict`. (La reapertura no está en este slice.)

## Endpoints

### POST `/api/v1/tickets` — crear

Origen típico: el canal, cuando detecta un caso a trackear.

```json
{
  "canal": "whatsapp",
  "solicitante_nombre": "Ana Pérez",
  "solicitante_telefono": "+5491100000000",
  "asunto": "No me llega el código de verificación",
  "prioridad": "alta"
}
```

`201 Created` → Ticket (estado `abierto`). `prioridad` es opcional (→ `media`).

### GET `/api/v1/tickets` — listar

Query params opcionales: `estado`, `asignado_a` (uuid), `limit` (default 20, máx 100), `offset`.
`200 OK` → array de Ticket del tenant.

### GET `/api/v1/tickets/{id}` — consultar

`200 OK` → Ticket. `404` si no existe (o no pertenece al tenant).

### POST `/api/v1/tickets/{id}/asignar`

```json
{ "operador_id": "<uuid>" }
```

`200 OK` → Ticket (`abierto` → `asignado`).

### POST `/api/v1/tickets/{id}/transicionar`

```json
{ "accion": "tomar" }   // tomar | resolver | cerrar
```

`200 OK` → Ticket con el nuevo estado. Transición inválida → `409`.

### POST `/api/v1/solicitantes/borrar-pii` — derecho de supresión (Ley 25.326)

Anonimiza la PII (nombre + teléfono) del solicitante en **todos** los tickets del tenant con ese
teléfono. El ticket sobrevive (historial/métricas); la PII se reemplaza por un tombstone.

```json
{ "telefono": "+5491100000000" }
```

`200 OK` → `{ "anonimizados": 3 }`. El teléfono viaja en el body (no en la URL) para no exponerlo en logs.

## Errores (RFC 7807)

`Content-Type: application/problem+json`

```json
{ "type": "about:blank", "title": "transición inválida", "status": 409, "detail": "ticket: transición de estado inválida" }
```

| Status | Cuándo |
|--------|--------|
| `400` | Body o id mal formado |
| `401` | Sin JWT / sin tenant válido |
| `404` | Ticket inexistente o de otro tenant |
| `409` | Transición de estado inválida |
| `422` | Datos de negocio inválidos (canal/prioridad/solicitante) |
| `503` | Base de datos no disponible |

## Eventos

En este slice los eventos de dominio (`ticket.creado`, `ticket.asignado`, `ticket.transicionado`,
`ticket.solicitante_pii_borrada`) se publican **in-process** (se loggean, sin PII). No hay push a
canales todavía: los canales **consultan** por GET. Webhooks/bus quedan para una iteración futura.
