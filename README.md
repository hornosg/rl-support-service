# support-service

Bounded context de soporte — dominio de tickets (CRUD + estados), multi-tenant fail-closed

Servicio Go de la plataforma **Devy**, generado con el golden path (`~/Projects/scaffold/new-service.sh`).
Hexagonal, integrado a `lab-network`, fail-closed desde el nacimiento.

## Correr local

```bash
cp .env.example .env      # completar JWT_SECRET
docker compose up -d      # crea DB + rol de app sin DDL, corre migraciones, levanta el servicio
curl localhost:8160/health
```

## Qué trae horneado (Devy)

- **RULE-09 — control/app-plane**: `postgres-setup` (superusuario) crea la DB, el rol `support_service_app`
  **sin DDL** y aplica migraciones. El servicio corre como `support_service_app`: no puede tocar infra.
- **RULE-10 — RLS fail-closed**: `migrations/002_rls.sql` activa Row-Level Security. El middleware
  `database.TenantSession` fija una conexión por request y setea `app.tenant_id` (del header
  `X-Tenant-ID`, validado contra el JWT). Si un handler olvida filtrar, la DB filtra igual.
- **Break-glass**: `system_admin` (header `X-User-Role`) accede cross-tenant — queda **auditado** en logs.
- **TenantValidation** (go-shared) a nivel HTTP + canonical-ish logs (zap) + métricas Prometheus.

## Estructura

```
src/
├── main.go                 # composition root
├── shared/database/        # conexión + TenantSession (RLS)
└── support_service/              # ← crear el dominio acá: application/ domain/ infrastructure/ ports/
migrations/                 # 001_init.sql + 002_rls.sql (corren como control-plane)
```

## Próximos pasos

1. Modelar el dominio en `src/support_service/` (hexagonal).
2. Por cada tabla tenant-scoped nueva, repetir el bloque RLS de `002_rls.sql`.
3. Registrar la ruta en Kong y el scrape en Prometheus (ver salida del generador, o `--wire`).
4. Antes de prod: pasar el **production-readiness scorecard** (roadmap E08).
