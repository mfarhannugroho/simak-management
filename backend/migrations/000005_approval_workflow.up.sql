-- Migration: 000005_approval_workflow.up.sql
-- Description: Configurable Approval Matrix and SIMAK Approvals Workflow Engine per PRD §12 & §13

CREATE TABLE IF NOT EXISTS approval_matrices (
    id             UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    name           VARCHAR(150) NOT NULL,
    description    TEXT,
    department_id  UUID REFERENCES departments(id) ON DELETE SET NULL,
    company_id     UUID REFERENCES companies(id) ON DELETE SET NULL,
    is_active      BOOLEAN NOT NULL DEFAULT TRUE,
    created_at     TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at     TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS approval_matrix_steps (
    id                UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    matrix_id         UUID NOT NULL REFERENCES approval_matrices(id) ON DELETE CASCADE,
    step_order        INT NOT NULL,
    step_name         VARCHAR(100) NOT NULL,
    required_role     VARCHAR(50) NOT NULL DEFAULT 'Approver',
    specific_user_id  UUID REFERENCES users(id) ON DELETE SET NULL,
    created_at        TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT uq_matrix_step_order UNIQUE (matrix_id, step_order)
);

CREATE TABLE IF NOT EXISTS simak_approvals (
    id                  UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    simak_id            UUID NOT NULL REFERENCES simak(id) ON DELETE CASCADE,
    step_order          INT NOT NULL,
    step_name           VARCHAR(100) NOT NULL,
    required_role       VARCHAR(50) NOT NULL DEFAULT 'Approver',
    assigned_user_id    UUID REFERENCES users(id) ON DELETE SET NULL,
    status              VARCHAR(30) NOT NULL DEFAULT 'PENDING',
    action_by_user_id   UUID REFERENCES users(id) ON DELETE SET NULL,
    action_at           TIMESTAMPTZ,
    comments            TEXT,
    created_at          TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at          TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_simak_approvals_simak_id ON simak_approvals(simak_id);
CREATE INDEX IF NOT EXISTS idx_simak_approvals_status_role ON simak_approvals(status, required_role);

-- ─────────────────────────────────────────────────────────────────────────────
-- Demo Seeding: Standard PT INKA Approval Matrix & Steps
-- ─────────────────────────────────────────────────────────────────────────────

INSERT INTO approval_matrices (id, name, description, is_active)
VALUES (
    'aaaa1111-1111-1111-1111-111111111111',
    'Matriks Persetujuan Standar PT INKA (Persero)',
    'Alur persetujuan default 2 tingkat: PIC/Manajer Divisi teknis (Level 1) dilanjutkan Supervisor Pengawas Area (Level 2).',
    TRUE
) ON CONFLICT (id) DO NOTHING;

INSERT INTO approval_matrix_steps (id, matrix_id, step_order, step_name, required_role)
VALUES
(
    'bbbb1111-1111-1111-1111-111111111101',
    'aaaa1111-1111-1111-1111-111111111111',
    1,
    'Pemeriksaan Manajer / PIC Divisi Terkait',
    'Approver'
),
(
    'bbbb1111-1111-1111-1111-111111111102',
    'aaaa1111-1111-1111-1111-111111111111',
    2,
    'Verifikasi Supervisor Pengawas Mutu & Area',
    'Supervisor'
) ON CONFLICT (matrix_id, step_order) DO NOTHING;

-- Seed pending approval steps for existing sample submitted SIMAK: '99999999-9999-9999-9999-999999999902' (REQ-2026-09-002)
INSERT INTO simak_approvals (id, simak_id, step_order, step_name, required_role, status)
VALUES
(
    uuid_generate_v4(),
    '99999999-9999-9999-9999-999999999902',
    1,
    'Pemeriksaan Manajer / PIC Divisi Terkait',
    'Approver',
    'PENDING'
),
(
    uuid_generate_v4(),
    '99999999-9999-9999-9999-999999999902',
    2,
    'Verifikasi Supervisor Pengawas Mutu & Area',
    'Supervisor',
    'PENDING'
) ON CONFLICT DO NOTHING;
