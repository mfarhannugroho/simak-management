-- Migration: 000007_gate_inspection.up.sql
-- Description: Gate inspections and inspection items per PRD §18

CREATE TABLE IF NOT EXISTS inspections (
    id                      UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    simak_id                UUID NOT NULL REFERENCES simak(id) ON DELETE CASCADE,
    inspector_id            UUID NOT NULL REFERENCES users(id) ON DELETE RESTRICT,
    gate_name               VARCHAR(100) NOT NULL DEFAULT 'Pos Gerbang Utama PT INKA',
    inspection_type         VARCHAR(30) NOT NULL DEFAULT 'MASUK',
    status                  VARCHAR(30) NOT NULL DEFAULT 'MATCH',
    total_registered_items  INT NOT NULL DEFAULT 0,
    total_actual_items      INT NOT NULL DEFAULT 0,
    notes                   TEXT,
    inspected_at            TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    created_at              TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS inspection_items (
    id             UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    inspection_id  UUID NOT NULL REFERENCES inspections(id) ON DELETE CASCADE,
    simak_item_id  UUID REFERENCES simak_items(id) ON DELETE SET NULL,
    item_name      VARCHAR(150) NOT NULL,
    brand          VARCHAR(100),
    serial_number  VARCHAR(100),
    expected_qty   INT NOT NULL DEFAULT 1,
    actual_qty     INT NOT NULL DEFAULT 0,
    unit           VARCHAR(30) NOT NULL DEFAULT 'Unit',
    status         VARCHAR(30) NOT NULL DEFAULT 'MATCH',
    notes          TEXT,
    photo_url      TEXT,
    created_at     TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_inspections_simak_id ON inspections(simak_id);
CREATE INDEX IF NOT EXISTS idx_inspections_inspected_at ON inspections(inspected_at DESC);
CREATE INDEX IF NOT EXISTS idx_inspections_status ON inspections(status);
CREATE INDEX IF NOT EXISTS idx_inspection_items_inspection_id ON inspection_items(inspection_id);
