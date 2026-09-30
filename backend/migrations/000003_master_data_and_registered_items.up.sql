-- Migration: 000003_master_data_and_registered_items.up.sql
-- Creates worker_registered_items table and seeds standard item categories and default items.

CREATE TABLE IF NOT EXISTS worker_registered_items (
    id            UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    worker_id     UUID NOT NULL REFERENCES workers(id) ON DELETE CASCADE,
    category_id   UUID REFERENCES item_categories(id) ON DELETE SET NULL,
    item_name     VARCHAR(150) NOT NULL,
    brand         VARCHAR(100),
    model         VARCHAR(100),
    serial_number VARCHAR(100),
    quantity      INT NOT NULL DEFAULT 1,
    unit          VARCHAR(30) NOT NULL DEFAULT 'Unit',
    description   TEXT,
    photo_url     TEXT,
    created_at    TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at    TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_worker_items_worker ON worker_registered_items(worker_id);
CREATE INDEX IF NOT EXISTS idx_worker_items_category ON worker_registered_items(category_id);

-- Seed Default Item Categories
INSERT INTO item_categories (id, name, description, is_active)
VALUES
    ('88888888-8888-8888-8888-888888888801', 'Laptop & Komputer Portabel', 'Laptop kedinasan, charger adaptor, dan docking station', true),
    ('88888888-8888-8888-8888-888888888802', 'Aksesori Komputer & IT', 'Mouse, keyboard eksternal, hard disk eksternal, flashdisk, dan kabel data', true),
    ('88888888-8888-8888-8888-888888888803', 'Alat Uji & Instrumen Listrik', 'Digital multimeter, clamp meter, insulation tester, dan oscilloscope portabel', true),
    ('88888888-8888-8888-8888-888888888804', 'Perkakas Mekanik & Hand Tools', 'Toolkit set perakitan, obeng, kunci pas/ring, bor listrik portabel, tang krimping', true),
    ('88888888-8888-8888-8888-888888888805', 'Kamera & Alat Dokumentasi', 'Kamera digital DSLR/Mirrorless, tripod, dan perangkat inspeksi visual boroskop', true),
    ('88888888-8888-8888-8888-888888888806', 'Perangkat Keselamatan & APD Khusus', 'Gas detector portabel, full body harness inspeksi tinggi, lampu helm tambang', true),
    ('88888888-8888-8888-8888-888888888807', 'Lain-lain / Perlengkapan Khusus', 'Perlengkapan kerja khusus di luar kategori standar di atas', true)
ON CONFLICT (name) DO NOTHING;

-- Seed Sample Registered Items for Applicant (Ahmad Wahyudi - Teknisi Vendor)
INSERT INTO worker_registered_items (id, worker_id, category_id, item_name, brand, model, serial_number, quantity, unit, description)
VALUES
    ('99999999-9999-9999-9999-999999999901', '77777777-7777-7777-7777-777777777772', '88888888-8888-8888-8888-888888888801', 'Laptop Dell Latitude 5420', 'Dell', 'Latitude 5420 Rugged', 'DL-LAT-882910', 1, 'Unit', 'Laptop operasional pemetaan wiring diagram'),
    ('99999999-9999-9999-9999-999999999902', '77777777-7777-7777-7777-777777777772', '88888888-8888-8888-8888-888888888802', 'Charger Adaptor Laptop Dell 65W', 'Dell', 'HA65NM130', 'CN-06TFFF-001', 1, 'Unit', 'Adaptor charger kabel Type-C original'),
    ('99999999-9999-9999-9999-999999999903', '77777777-7777-7777-7777-777777777772', '88888888-8888-8888-8888-888888888803', 'Digital Multimeter Fluke 179', 'Fluke', 'Fluke 179 True-RMS', 'FK-179-99201', 1, 'Unit', 'Alat ukur tegangan dan resistansi bogie'),
    ('99999999-9999-9999-9999-999999999904', '77777777-7777-7777-7777-777777777772', '88888888-8888-8888-8888-888888888804', 'Cordless Impact Drill 18V', 'Bosch', 'GSB 18V-50 Professional', 'BSH-ID-54311', 1, 'Set', 'Bor baterai + 2 baterai cadangan + charger box')
ON CONFLICT (id) DO NOTHING;
