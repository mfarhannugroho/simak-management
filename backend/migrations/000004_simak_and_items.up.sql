-- Migration: 000004_simak_and_items.up.sql
-- Description: Core tables for SIMAK (Surat Ijin Membawa Alat Kerja), items, and attachments per PRD §4 & §11

CREATE TABLE IF NOT EXISTS simak (
    id                 UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    worker_id          UUID NOT NULL REFERENCES workers(id) ON DELETE CASCADE,
    simak_number       VARCHAR(60) UNIQUE,
    purpose            TEXT NOT NULL,
    direction          VARCHAR(30) NOT NULL DEFAULT 'MASUK',
    destination_area   VARCHAR(150) NOT NULL,
    start_date         DATE NOT NULL,
    end_date           DATE NOT NULL,
    status             VARCHAR(40) NOT NULL DEFAULT 'DRAFT',
    notes              TEXT,
    revision_notes     TEXT,
    rejection_reason   TEXT,
    created_at         TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at         TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_simak_worker_id ON simak(worker_id);
CREATE INDEX IF NOT EXISTS idx_simak_status ON simak(status);
CREATE INDEX IF NOT EXISTS idx_simak_number ON simak(simak_number);
CREATE INDEX IF NOT EXISTS idx_simak_dates ON simak(start_date, end_date);
CREATE INDEX IF NOT EXISTS idx_simak_created_at ON simak(created_at DESC);

CREATE TABLE IF NOT EXISTS simak_items (
    id            UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    simak_id      UUID NOT NULL REFERENCES simak(id) ON DELETE CASCADE,
    category_id   UUID REFERENCES item_categories(id) ON DELETE SET NULL,
    item_name     VARCHAR(150) NOT NULL,
    brand         VARCHAR(100),
    model         VARCHAR(100),
    serial_number VARCHAR(100),
    size_spec     VARCHAR(100),
    quantity      INT NOT NULL DEFAULT 1 CHECK (quantity > 0),
    unit          VARCHAR(30) NOT NULL DEFAULT 'Unit',
    description   TEXT,
    photo_url     TEXT,
    created_at    TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at    TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_simak_items_simak_id ON simak_items(simak_id);
CREATE INDEX IF NOT EXISTS idx_simak_items_category_id ON simak_items(category_id);

CREATE TABLE IF NOT EXISTS simak_attachments (
    id              UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    simak_id        UUID NOT NULL REFERENCES simak(id) ON DELETE CASCADE,
    file_name       VARCHAR(255) NOT NULL,
    file_url        TEXT NOT NULL,
    file_type       VARCHAR(100),
    file_size       BIGINT,
    attachment_type VARCHAR(50) NOT NULL DEFAULT 'SURAT_TUGAS',
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_simak_attachments_simak_id ON simak_attachments(simak_id);

-- ─────────────────────────────────────────────────────────────────────────────
-- Demo Seeding for Ahmad Wahyudi (REK-FAB-104)
-- Worker ID: '77777777-7777-7777-7777-777777777772'
-- ─────────────────────────────────────────────────────────────────────────────

-- Sample 1: DRAFT SIMAK
INSERT INTO simak (
    id, worker_id, simak_number, purpose, direction, destination_area,
    start_date, end_date, status, notes, created_at, updated_at
) VALUES (
    '99999999-9999-9999-9999-999999999901',
    '77777777-7777-7777-7777-777777777772',
    'DFT-2026-09-001',
    'Pengujian kelistrikan interior gerbong K3 di jalur final inspection',
    'MASUK_KELUAR',
    'Workshop Finishing & Final Inspection',
    CURRENT_DATE,
    CURRENT_DATE + INTERVAL '5 days',
    'DRAFT',
    'Membawa toolkit teknisi Rekaindo dan laptop diagnostik',
    NOW() - INTERVAL '2 hours',
    NOW() - INTERVAL '2 hours'
) ON CONFLICT (id) DO NOTHING;

-- Items for Sample 1
INSERT INTO simak_items (
    id, simak_id, category_id, item_name, brand, model, serial_number, size_spec, quantity, unit, description
) VALUES
(
    uuid_generate_v4(),
    '99999999-9999-9999-9999-999999999901',
    '88888888-8888-8888-8888-888888888801',
    'Laptop Dell Latitude 5420 Rugged',
    'Dell',
    'Latitude 5420',
    'DL-LAT-882910',
    '14 inch',
    1,
    'Unit',
    'Laptop diagnostik sistem pengereman & kelistrikan'
),
(
    uuid_generate_v4(),
    '99999999-9999-9999-9999-999999999901',
    '88888888-8888-8888-8888-888888888803',
    'Digital Multimeter Fluke 179',
    'Fluke',
    '179 True-RMS',
    'FK-179-99201',
    'CAT IV 600V',
    1,
    'Unit',
    'Termasuk probe kabel merah-hitam'
) ON CONFLICT DO NOTHING;

-- Sample 2: SUBMITTED SIMAK (Menunggu Approval)
INSERT INTO simak (
    id, worker_id, simak_number, purpose, direction, destination_area,
    start_date, end_date, status, notes, created_at, updated_at
) VALUES (
    '99999999-9999-9999-9999-999999999902',
    '77777777-7777-7777-7777-777777777772',
    'REQ-2026-09-002',
    'Instalasi suspensi sekunder dan perakitan rangka bogie pesanan KRL',
    'MASUK',
    'Workshop Perakitan Bogie PT INKA',
    CURRENT_DATE,
    CURRENT_DATE + INTERVAL '14 days',
    'SUBMITTED',
    'Peralatan montir mekanik bersertifikasi PT Rekaindo',
    NOW() - INTERVAL '1 day',
    NOW() - INTERVAL '1 day'
) ON CONFLICT (id) DO NOTHING;

-- Items for Sample 2
INSERT INTO simak_items (
    id, simak_id, category_id, item_name, brand, model, serial_number, size_spec, quantity, unit, description
) VALUES
(
    uuid_generate_v4(),
    '99999999-9999-9999-9999-999999999902',
    '88888888-8888-8888-8888-888888888804',
    'Cordless Impact Drill 18V',
    'Bosch',
    'GSB 18V-50 Professional',
    'BSH-ID-54311',
    '13mm Chuck',
    1,
    'Set',
    'Lengkap dengan 2 baterai Li-Ion dan charger'
),
(
    uuid_generate_v4(),
    '99999999-9999-9999-9999-999999999902',
    '88888888-8888-8888-8888-888888888804',
    'Torque Wrench Kunci Momen 1/2 inch',
    'Toptul',
    'ANAM1621',
    'TP-TW-8712',
    '40-210 Nm',
    1,
    'Pcs',
    'Kalibrasi berlaku s.d Desember 2026'
) ON CONFLICT DO NOTHING;

-- Sample Attachment for Sample 2
INSERT INTO simak_attachments (
    id, simak_id, file_name, file_url, file_type, file_size, attachment_type
) VALUES (
    uuid_generate_v4(),
    '99999999-9999-9999-9999-999999999902',
    'Surat_Tugas_Rekaindo_Bogie_2026.pdf',
    '/uploads/documents/surat_tugas_rekaindo_demo.pdf',
    'application/pdf',
    245760,
    'SURAT_TUGAS'
) ON CONFLICT DO NOTHING;
