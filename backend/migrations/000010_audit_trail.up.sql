-- 000010_audit_trail.up.sql
-- Append-only immutable audit trail log for PT INKA SIMAK system

CREATE TABLE IF NOT EXISTS audit_logs (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id UUID REFERENCES users(id) ON DELETE SET NULL,
    user_email VARCHAR(150),
    user_name VARCHAR(150),
    action VARCHAR(60) NOT NULL,
    entity_type VARCHAR(60) NOT NULL,
    entity_id VARCHAR(100),
    details JSONB DEFAULT '{}'::jsonb,
    ip_address VARCHAR(45),
    user_agent TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_audit_logs_user_id ON audit_logs(user_id);
CREATE INDEX IF NOT EXISTS idx_audit_logs_action ON audit_logs(action);
CREATE INDEX IF NOT EXISTS idx_audit_logs_entity_type ON audit_logs(entity_type);
CREATE INDEX IF NOT EXISTS idx_audit_logs_entity_id ON audit_logs(entity_id);
CREATE INDEX IF NOT EXISTS idx_audit_logs_created_at ON audit_logs(created_at DESC);

-- Seed initial representative audit records from existing activity
INSERT INTO audit_logs (user_id, user_email, user_name, action, entity_type, entity_id, details, ip_address, user_agent, created_at)
VALUES 
(
    '66666666-6666-6666-6666-666666666661',
    'admin@inka.co.id',
    'Bagus Pratama, S.Kom',
    'LOGIN',
    'AUTH',
    '66666666-6666-6666-6666-666666666661',
    '{"message": "Login berhasil menggunakan kredensial email & password", "method": "JWT"}'::jsonb,
    '127.0.0.1',
    'Mozilla/5.0 (Windows NT 10.0; Win64; x64) Chrome/128.0.0.0',
    NOW() - INTERVAL '3 hours'
),
(
    '66666666-6666-6666-6666-666666666662',
    'pekerja@inka.co.id',
    'Ahmad Dahlan',
    'SIMAK_CREATED',
    'SIMAK',
    'SIMAK/INKA/2026/000001',
    '{"simak_number": "SIMAK/INKA/2026/000001", "purpose": "Perbaikan instalasi kelistrikan bogie KRL", "items_count": 2}'::jsonb,
    '10.20.14.52',
    'Mozilla/5.0 (Windows NT 10.0; Win64; x64) Chrome/128.0.0.0',
    NOW() - INTERVAL '2 hours 30 minutes'
),
(
    '66666666-6666-6666-6666-666666666662',
    'pekerja@inka.co.id',
    'Ahmad Dahlan',
    'SIMAK_SUBMITTED',
    'SIMAK',
    'SIMAK/INKA/2026/000001',
    '{"status": "SUBMITTED", "assigned_matrix": "Matriks Standar Workshop Fabrikasi"}'::jsonb,
    '10.20.14.52',
    'Mozilla/5.0 (Windows NT 10.0; Win64; x64) Chrome/128.0.0.0',
    NOW() - INTERVAL '2 hours 25 minutes'
),
(
    '66666666-6666-6666-6666-666666666663',
    'approver@inka.co.id',
    'Ir. Hendro Wijaya, M.T.',
    'SIMAK_APPROVED',
    'SIMAK',
    'SIMAK/INKA/2026/000001',
    '{"step": 1, "comments": "Disetujui. Peralatan sesuai standar K3 keselamatan kerja INKA."}'::jsonb,
    '10.20.10.11',
    'Mozilla/5.0 (Windows NT 10.0; Win64; x64) Chrome/128.0.0.0',
    NOW() - INTERVAL '2 hours'
),
(
    '66666666-6666-6666-6666-666666666664',
    'officer@inka.co.id',
    'Dewi Sartika, S.T.',
    'SIMAK_ISSUED',
    'SIMAK',
    'SIMAK/INKA/2026/000001',
    '{"qr_issued": true, "valid_until": "2026-10-15"}'::jsonb,
    '10.20.12.8',
    'Mozilla/5.0 (Windows NT 10.0; Win64; x64) Chrome/128.0.0.0',
    NOW() - INTERVAL '1 hour 45 minutes'
),
(
    '66666666-6666-6666-6666-666666666665',
    'gate@inka.co.id',
    'Slamet Riyadi',
    'INSPECTION_RECORDED',
    'INSPECTION',
    'SIMAK/INKA/2026/000001',
    '{"gate_name": "Pos Gerbang Utama PT INKA", "inspection_type": "MASUK", "status": "MATCH", "items_checked": 2}'::jsonb,
    '10.20.100.2',
    'Mozilla/5.0 (Android; Mobile) Chrome/128.0.0.0',
    NOW() - INTERVAL '1 hour'
),
(
    '66666666-6666-6666-6666-666666666665',
    'gate@inka.co.id',
    'Slamet Riyadi',
    'DISCREPANCY_CREATED',
    'DISCREPANCY',
    'DISC/INKA/2026/000001',
    '{"type": "QUANTITY_MISMATCH", "item_name": "Toolbox Set Bahco", "expected": 2, "actual": 1}'::jsonb,
    '10.20.100.2',
    'Mozilla/5.0 (Android; Mobile) Chrome/128.0.0.0',
    NOW() - INTERVAL '45 minutes'
),
(
    '66666666-6666-6666-6666-666666666666',
    'supervisor@inka.co.id',
    'Bambang Sugianto, S.T.',
    'DISCREPANCY_RESOLVED',
    'DISCREPANCY',
    'DISC/INKA/2026/000001',
    '{"action": "ALLOW_ENTRY", "notes": "1 toolbox tertinggal di kendaraan vendor dan diverifikasi menyusul."}'::jsonb,
    '10.20.11.45',
    'Mozilla/5.0 (Windows NT 10.0; Win64; x64) Chrome/128.0.0.0',
    NOW() - INTERVAL '20 minutes'
);
