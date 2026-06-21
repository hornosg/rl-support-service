-- 003_tickets.sql — Dominio de tickets (E03 / PROP-001).
-- Corre como CONTROL PLANE (superusuario) vía postgres-setup. El rol de app no tiene DDL (RULE-09).
--
-- Tenant-scoped por defecto (RULE-04): lleva tenant_id y se aísla por RLS (mismo patrón que 002).
-- La app NO depende del WHERE tenant_id: si una query olvida filtrar, la base la filtra igual.

CREATE TABLE IF NOT EXISTS tickets (
    id                   uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id            uuid NOT NULL,

    -- Canal de origen (G-07) y solicitante / cliente final (G-04, value object con contacto).
    -- PII del cliente final (Ley 25.326): minimizar y soportar borrado a pedido del tenant.
    canal                text NOT NULL,
    solicitante_nombre   text NOT NULL,
    solicitante_telefono text NOT NULL,

    asunto               text NOT NULL,
    prioridad            text NOT NULL DEFAULT 'media',

    -- Máquina de estados LINEAL (slice POC): abierto→asignado→en_curso→resuelto→cerrado.
    estado               text NOT NULL DEFAULT 'abierto',
    asignado_a           uuid,                       -- operador (G-03); NULL hasta asignar

    created_at           timestamptz NOT NULL DEFAULT now(),
    updated_at           timestamptz NOT NULL DEFAULT now(),

    CONSTRAINT tickets_estado_chk    CHECK (estado IN ('abierto','asignado','en_curso','resuelto','cerrado')),
    CONSTRAINT tickets_prioridad_chk CHECK (prioridad IN ('baja','media','alta','urgente')),
    CONSTRAINT tickets_canal_chk     CHECK (canal IN ('whatsapp','web','email'))
);

CREATE INDEX IF NOT EXISTS tickets_tenant_idx        ON tickets (tenant_id);
CREATE INDEX IF NOT EXISTS tickets_tenant_estado_idx ON tickets (tenant_id, estado);
CREATE INDEX IF NOT EXISTS tickets_tenant_asig_idx   ON tickets (tenant_id, asignado_a);

-- ── RLS fail-closed (RULE-10), mismo patrón que 002_rls.sql ──
ALTER TABLE tickets ENABLE ROW LEVEL SECURITY;
ALTER TABLE tickets FORCE  ROW LEVEL SECURITY;

DROP POLICY IF EXISTS tenant_isolation ON tickets;
CREATE POLICY tenant_isolation ON tickets
    USING      (tenant_id = current_setting('app.tenant_id', true)::uuid)
    WITH CHECK (tenant_id = current_setting('app.tenant_id', true)::uuid);

-- Break-glass del owner (system_admin), auditado en logs por la app.
DROP POLICY IF EXISTS break_glass ON tickets;
CREATE POLICY break_glass ON tickets
    USING (current_setting('app.role', true) = 'system_admin');
