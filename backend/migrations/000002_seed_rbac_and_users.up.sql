-- Migration: 000002_seed_rbac_and_users.up.sql
-- Seeds default organizations, companies, departments, roles, permissions, and test users for 6 roles.

-- 1. Organization & Companies
INSERT INTO organizations (id, name, code, is_active)
VALUES ('11111111-1111-1111-1111-111111111111', 'PT Industri Kereta Api (Persero)', 'INKA', true)
ON CONFLICT (code) DO NOTHING;

INSERT INTO companies (id, organization_id, name, type, address, contact_phone, contact_email, is_active)
VALUES 
    ('22222222-2222-2222-2222-222222222221', '11111111-1111-1111-1111-111111111111', 'PT INKA (Persero) Pusat Madiun', 'INTERNAL', 'Jl. Yos Sudarso No.71, Madiun, Jawa Timur', '0351-452271', 'info@inka.co.id', true),
    ('22222222-2222-2222-2222-222222222222', '11111111-1111-1111-1111-111111111111', 'PT Rekaindo Global Jasa', 'VENDOR', 'Jl. Ringroad Barat, Madiun', '0351-451120', 'vendor@rekaindo.com', true)
ON CONFLICT DO NOTHING;

-- 2. Departments
INSERT INTO departments (id, organization_id, name, code, is_active)
VALUES
    ('33333333-3333-3333-3333-333333333331', '11111111-1111-1111-1111-111111111111', 'Divisi Rekayasa & Fabrikasi Kereta', 'FAB', true),
    ('33333333-3333-3333-3333-333333333332', '11111111-1111-1111-1111-111111111111', 'Divisi Pengamanan & Ketertiban Lingkungan (Security)', 'SEC', true),
    ('33333333-3333-3333-3333-333333333333', '11111111-1111-1111-1111-111111111111', 'Divisi Administrasi & Legalitas Umum', 'ADM', true),
    ('33333333-3333-3333-3333-333333333334', '11111111-1111-1111-1111-111111111111', 'Divisi Pengawasan Mutu & K3LH', 'HSE', true)
ON CONFLICT DO NOTHING;

-- 3. Roles
INSERT INTO roles (id, name, description)
VALUES
    ('44444444-4444-4444-4444-444444444441', 'Admin', 'System Administrator with full master data & RBAC privileges'),
    ('44444444-4444-4444-4444-444444444442', 'Applicant', 'Pekerja/Vendor pemohon izin membawa barang masuk/keluar'),
    ('44444444-4444-4444-4444-444444444443', 'Approver', 'Pejabat/Supervisor berwenang menyetujui, menolak atau merevisi SIMAK'),
    ('44444444-4444-4444-4444-444444444444', 'SIMAK Officer', 'Petugas administrasi penerbitan resmi SIMAK dan dokumen QR'),
    ('44444444-4444-4444-4444-444444444445', 'Gate Officer', 'Petugas satpam pos gerbang pemeriksa barang fisik vs data SIMAK'),
    ('44444444-4444-4444-4444-444444444446', 'Supervisor', 'Pengawas operasional lapangan, discrepancy resolver & laporan')
ON CONFLICT (name) DO NOTHING;

-- 4. Permissions
INSERT INTO permissions (id, name, description)
VALUES
    ('55555555-5555-5555-5555-555555555501', 'simak.create', 'Membuat permohonan SIMAK baru'),
    ('55555555-5555-5555-5555-555555555502', 'simak.view', 'Melihat detail dan list SIMAK'),
    ('55555555-5555-5555-5555-555555555503', 'simak.update', 'Mengubah draft atau revisi SIMAK'),
    ('55555555-5555-5555-5555-555555555504', 'simak.submit', 'Mengajukan permohonan SIMAK ke tahap review'),
    ('55555555-5555-5555-5555-555555555505', 'simak.cancel', 'Membatalkan SIMAK'),
    ('55555555-5555-5555-5555-555555555506', 'simak.review', 'Menelaah kelengkapan permohonan SIMAK'),
    ('55555555-5555-5555-5555-555555555507', 'simak.approve', 'Menyetujui permohonan SIMAK'),
    ('55555555-5555-5555-5555-555555555508', 'simak.reject', 'Menolak permohonan SIMAK'),
    ('55555555-5555-5555-5555-555555555509', 'simak.request_revision', 'Meminta pemohon memperbaiki rincian SIMAK'),
    ('55555555-5555-5555-5555-555555555510', 'simak.issue', 'Menerbitkan nomor resmi dan QR Code SIMAK'),
    ('55555555-5555-5555-5555-555555555511', 'item.create', 'Mendaftarkan item barang bawaan'),
    ('55555555-5555-5555-5555-555555555512', 'item.update', 'Mengedit data item barang bawaan'),
    ('55555555-5555-5555-5555-555555555513', 'item.delete', 'Menghapus item barang bawaan draft'),
    ('55555555-5555-5555-5555-555555555514', 'item.view', 'Melihat detail item barang bawaan'),
    ('55555555-5555-5555-5555-555555555515', 'inspection.view', 'Melihat data pemeriksaan pos gerbang'),
    ('55555555-5555-5555-5555-555555555516', 'inspection.verify', 'Scan QR Code verifikasi pos gerbang'),
    ('55555555-5555-5555-5555-555555555517', 'inspection.perform', 'Mengeksekusi checklist fisik barang dan simpan hasil'),
    ('55555555-5555-5555-5555-555555555518', 'inspection.resolve_discrepancy', 'Menyelesaikan temuan discrepancy dan eskalasi'),
    ('55555555-5555-5555-5555-555555555519', 'report.view', 'Melihat statistik dan laporan analitik'),
    ('55555555-5555-5555-5555-555555555520', 'report.export', 'Mengekspor laporan ke Excel atau PDF'),
    ('55555555-5555-5555-5555-555555555521', 'user.manage', 'Mengelola user dan profile pekerja'),
    ('55555555-5555-5555-5555-555555555522', 'role.manage', 'Mengelola role dan mapping RBAC'),
    ('55555555-5555-5555-5555-555555555523', 'permission.manage', 'Mengatur granular permissions'),
    ('55555555-5555-5555-5555-555555555524', 'system.manage', 'Konfigurasi parameter sistem dan approval matrix'),
    ('55555555-5555-5555-5555-555555555525', 'audit.view', 'Melihat log audit trail kepatuhan')
ON CONFLICT (name) DO NOTHING;

-- 5. Role Permissions Mapping
-- Admin (all)
INSERT INTO role_permissions (role_id, permission_id)
SELECT '44444444-4444-4444-4444-444444444441', id FROM permissions
ON CONFLICT DO NOTHING;

-- Applicant
INSERT INTO role_permissions (role_id, permission_id)
SELECT '44444444-4444-4444-4444-444444444442', id FROM permissions
WHERE name IN ('simak.create', 'simak.view', 'simak.update', 'simak.submit', 'simak.cancel', 'item.create', 'item.update', 'item.delete', 'item.view')
ON CONFLICT DO NOTHING;

-- Approver
INSERT INTO role_permissions (role_id, permission_id)
SELECT '44444444-4444-4444-4444-444444444443', id FROM permissions
WHERE name IN ('simak.view', 'simak.review', 'simak.approve', 'simak.reject', 'simak.request_revision', 'item.view', 'report.view')
ON CONFLICT DO NOTHING;

-- SIMAK Officer
INSERT INTO role_permissions (role_id, permission_id)
SELECT '44444444-4444-4444-4444-444444444444', id FROM permissions
WHERE name IN ('simak.view', 'simak.review', 'simak.issue', 'simak.cancel', 'item.view', 'report.view', 'report.export')
ON CONFLICT DO NOTHING;

-- Gate Officer
INSERT INTO role_permissions (role_id, permission_id)
SELECT '44444444-4444-4444-4444-444444444445', id FROM permissions
WHERE name IN ('inspection.view', 'inspection.verify', 'inspection.perform', 'simak.view', 'item.view')
ON CONFLICT DO NOTHING;

-- Supervisor
INSERT INTO role_permissions (role_id, permission_id)
SELECT '44444444-4444-4444-4444-444444444446', id FROM permissions
WHERE name IN ('simak.view', 'inspection.view', 'inspection.resolve_discrepancy', 'report.view', 'report.export', 'audit.view')
ON CONFLICT DO NOTHING;

-- 6. Users (with precomputed Bcrypt hashes)
INSERT INTO users (id, email, password_hash, is_active)
VALUES
    ('66666666-6666-6666-6666-666666666661', 'admin@inka.co.id', '$2a$10$BD3iH/UrRYunB9oRAmOOjuIp5t6z3yGZJpP.itg./IySuuUaFH1G2', true),
    ('66666666-6666-6666-6666-666666666662', 'pekerja@inka.co.id', '$2a$10$NM9fh0v0U05ejXqFqg2gN.wtY/5ckgoFShVbnBTGeTRbdi7aXxVf.', true),
    ('66666666-6666-6666-6666-666666666663', 'approver@inka.co.id', '$2a$10$x8eZxk3CkwnAnu8.u7NJwOIX/S.kQbGlx.PR6.YsyGki5jBy8P/de', true),
    ('66666666-6666-6666-6666-666666666664', 'officer@inka.co.id', '$2a$10$gCPRzhpsZltkE/eQk7vb2euqWabE8TH.Qb/WRbvVG7/Q8RB/9u8Ge', true),
    ('66666666-6666-6666-6666-666666666665', 'gate@inka.co.id', '$2a$10$toKZfNJmm5IMDLEiu4nY9exVnXm1fnS6COJDHYCdhIguaDP1eOc/u', true),
    ('66666666-6666-6666-6666-666666666666', 'supervisor@inka.co.id', '$2a$10$AkdE5F2fy2oR6n6ipjB2UeGEj0EcolQFzoVRMWIDJfygRHkAEu0Em', true)
ON CONFLICT (email) DO NOTHING;

-- 7. User Roles
INSERT INTO user_roles (user_id, role_id)
VALUES
    ('66666666-6666-6666-6666-666666666661', '44444444-4444-4444-4444-444444444441'), -- Admin
    ('66666666-6666-6666-6666-666666666662', '44444444-4444-4444-4444-444444444442'), -- Applicant
    ('66666666-6666-6666-6666-666666666663', '44444444-4444-4444-4444-444444444443'), -- Approver
    ('66666666-6666-6666-6666-666666666664', '44444444-4444-4444-4444-444444444444'), -- SIMAK Officer
    ('66666666-6666-6666-6666-666666666665', '44444444-4444-4444-4444-444444444445'), -- Gate Officer
    ('66666666-6666-6666-6666-666666666666', '44444444-4444-4444-4444-444444444446')  -- Supervisor
ON CONFLICT DO NOTHING;

-- 8. Workers Profiles
INSERT INTO workers (id, user_id, company_id, department_id, worker_id, full_name, position, phone)
VALUES
    ('77777777-7777-7777-7777-777777777771', '66666666-6666-6666-6666-666666666661', '22222222-2222-2222-2222-222222222221', '33333333-3333-3333-3333-333333333333', 'INKA-ADM-001', 'Bagus Pratama, S.Kom', 'Administrator Sistem SIMAK', '081234567890'),
    ('77777777-7777-7777-7777-777777777772', '66666666-6666-6666-6666-666666666662', '22222222-2222-2222-2222-222222222222', '33333333-3333-3333-3333-333333333331', 'REK-FAB-104', 'Ahmad Wahyudi, A.Md', 'Teknisi Perakitan Bogie (Vendor)', '082198765432'),
    ('77777777-7777-7777-7777-777777777773', '66666666-6666-6666-6666-666666666663', '22222222-2222-2222-2222-222222222221', '33333333-3333-3333-3333-333333333331', 'INKA-MGR-022', 'Ir. Hendro Wijaya, M.T.', 'Manajer Fabrikasi Kereta Penumpang', '081345678901'),
    ('77777777-7777-7777-7777-777777777774', '66666666-6666-6666-6666-666666666664', '22222222-2222-2222-2222-222222222221', '33333333-3333-3333-3333-333333333333', 'INKA-OFC-054', 'Siti Rahmawati, S.AP', 'Petugas Administrasi SIMAK', '081298761234'),
    ('77777777-7777-7777-7777-777777777775', '66666666-6666-6666-6666-666666666665', '22222222-2222-2222-2222-222222222221', '33333333-3333-3333-3333-333333333332', 'INKA-SEC-012', 'Slamet Riyadi', 'Komandan Regu Pos Gerbang 1', '085234567899'),
    ('77777777-7777-7777-7777-777777777776', '66666666-6666-6666-6666-666666666666', '22222222-2222-2222-2222-222222222221', '33333333-3333-3333-3333-333333333334', 'INKA-SPV-008', 'Budi Santoso, S.T.', 'Supervisor Pengawas Mutu & Area', '081233445566')
ON CONFLICT (worker_id) DO NOTHING;
