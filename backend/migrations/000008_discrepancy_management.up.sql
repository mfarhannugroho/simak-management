-- Migration 000008: Discrepancy Management
CREATE TABLE IF NOT EXISTS discrepancies (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    discrepancy_number VARCHAR(50) NOT NULL UNIQUE,
    simak_id UUID NOT NULL REFERENCES simak(id) ON DELETE CASCADE,
    inspection_id UUID NOT NULL REFERENCES inspections(id) ON DELETE CASCADE,
    inspection_item_id UUID REFERENCES inspection_items(id) ON DELETE SET NULL,
    item_name VARCHAR(255) NOT NULL,
    brand VARCHAR(100),
    serial_number VARCHAR(100),
    discrepancy_type VARCHAR(50) NOT NULL, -- QUANTITY_MISMATCH, UNREGISTERED_ITEM, MISSING_ITEM, ITEM_DESCRIPTION_MISMATCH, OTHER
    expected_qty INT NOT NULL DEFAULT 0,
    actual_qty INT NOT NULL DEFAULT 0,
    unit VARCHAR(50) NOT NULL DEFAULT 'Unit',
    description TEXT,
    evidence_photo_url TEXT,
    reported_by UUID NOT NULL REFERENCES users(id),
    reviewed_by UUID REFERENCES users(id),
    reviewed_at TIMESTAMP WITH TIME ZONE,
    resolution_action VARCHAR(50), -- ALLOW_ENTRY, DENY_ENTRY, CONFISCATE_HOLD_AT_GATE, RETURN_TO_VENDOR, REVISE_PERMIT, WARNING_ISSUED
    resolution_notes TEXT,
    status VARCHAR(50) NOT NULL DEFAULT 'OPEN', -- OPEN, IN_REVIEW, RESOLVED, CLOSED
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_discrepancies_simak_id ON discrepancies(simak_id);
CREATE INDEX IF NOT EXISTS idx_discrepancies_inspection_id ON discrepancies(inspection_id);
CREATE INDEX IF NOT EXISTS idx_discrepancies_status ON discrepancies(status);
CREATE INDEX IF NOT EXISTS idx_discrepancies_type ON discrepancies(discrepancy_type);
CREATE INDEX IF NOT EXISTS idx_discrepancies_created_at ON discrepancies(created_at);
