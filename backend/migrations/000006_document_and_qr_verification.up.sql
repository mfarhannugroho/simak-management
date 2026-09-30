-- Migration: 000006_document_and_qr_verification.up.sql
-- Description: SIMAK issuance, QR token, and official document registry per PRD §14, §15, §16, §17

ALTER TABLE simak
    ADD COLUMN IF NOT EXISTS qr_token VARCHAR(100) UNIQUE,
    ADD COLUMN IF NOT EXISTS issued_at TIMESTAMPTZ,
    ADD COLUMN IF NOT EXISTS issued_by UUID REFERENCES users(id) ON DELETE SET NULL;

CREATE TABLE IF NOT EXISTS simak_documents (
    id              UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    simak_id        UUID NOT NULL REFERENCES simak(id) ON DELETE CASCADE,
    document_number VARCHAR(100) NOT NULL,
    file_name       VARCHAR(255) NOT NULL,
    file_url        TEXT,
    doc_type        VARCHAR(50) NOT NULL DEFAULT 'OFFICIAL_PDF',
    qr_code_data    TEXT,
    issued_by       UUID REFERENCES users(id) ON DELETE SET NULL,
    issued_at       TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_simak_documents_simak_id ON simak_documents(simak_id);
CREATE INDEX IF NOT EXISTS idx_simak_qr_token ON simak(qr_token);

-- Update existing approved SIMAK sample (99999999-9999-9999-9999-999999999902) to ISSUED with mock QR token
UPDATE simak
SET 
    simak_number = COALESCE(simak_number, 'SIMAK/INKA/2026/000001'),
    qr_token = 'token-inka-2026-demo-verify-001',
    issued_at = NOW(),
    issued_by = '66666666-6666-6666-6666-666666666664', -- SIMAK Officer
    status = 'ISSUED'
WHERE id = '99999999-9999-9999-9999-999999999902' AND status = 'APPROVED';
