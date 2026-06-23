# support-service

Servicio de **soporte** de Riotless: dueño del ciclo de vida de los **tickets** (alta, asignación
y transición de estado), multi-tenant y *fail-closed*. El canal de entrada (WhatsApp hoy; web/email
a futuro) lo consume por contrato HTTP — no es dueño del dominio.

- **Stack:** Go + Gin · PostgreSQL con Row-Level Security · arquitectura hexagonal + DDD.
- **Puerto:** `8160` · métricas en `/metrics` · salud en `/health`.

## Correr local

Requiere PostgreSQL accesible (la DB `support_service`, su rol de app y las migraciones).

```bash
cp .env.example .env      # completar JWT_SECRET y la password de la DB
docker compose up -d      # provisiona DB + rol de app sin DDL, corre migraciones y levanta el servicio
curl localhost:8160/health
```

## El contrato (resumen)

Todos los endpoints de negocio van bajo `/api/v1` y exigen tenant (`X-Tenant-ID`, validado contra el
JWT). Los errores siguen Problem Details (RFC 7807). Detalle completo en [`api-docs/`](./api-docs/contract.md).

| Verbo | Ruta | Qué |
|-------|------|-----|
| POST | `/api/v1/tickets` | Crear ticket (origen: canal) |
| GET | `/api/v1/tickets` | Listar tickets del tenant (filtros: `estado`, `asignado_a`) |
| GET | `/api/v1/tickets/:id` | Consultar un ticket |
| POST | `/api/v1/tickets/:id/asignar` | Asignar a un operador |
| POST | `/api/v1/tickets/:id/transicionar` | Avanzar estado (`tomar`/`resolver`/`cerrar`) |
| POST | `/api/v1/solicitantes/borrar-pii` | Anonimizar la PII de un solicitante (Ley 25.326) |

**Ciclo de vida del ticket** (lineal): `abierto → asignado → en_curso → resuelto → cerrado`.
Toda transición inválida se rechaza.

## Aislamiento multi-tenant (fail-closed)

El aislamiento entre tenants vive en la base, no solo en el código:

- Cada request fija una conexión y setea `app.tenant_id` (del header `X-Tenant-ID`).
- Las políticas **RLS** de PostgreSQL filtran por ese tenant: si un query olvida filtrar, la base
  filtra igual. El rol de la aplicación no puede saltar RLS ni ejecutar DDL.
- **Break-glass**: solo un `system_admin` accede cross-tenant, y queda auditado en los logs.
- Sin tenant válido, toda operación de negocio se rechaza.

## Estructura

```
src/
├── main.go                      # composition root (router, middlewares, wiring)
├── shared/database/             # conexión + sesión de tenant (RLS)
└── support/
    ├── domain/                  # agregado Ticket, value objects, máquina de estados, eventos, ports
    ├── application/             # use cases + DTOs
    └── infrastructure/          # persistencia (Postgres), HTTP (Gin), Problem Details, eventos
migrations/                      # esquema + RLS (se aplican como control-plane)
test/                            # object mothers, fakes e integración (build tag `integration`)
```

## Tests

```bash
go test ./...                                  # unit (dominio + use cases), sin DB
go test -tags integration ./test/integration/  # aislamiento RLS y ciclo de vida contra una DB real
```
