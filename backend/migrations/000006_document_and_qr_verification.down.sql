-- Migration: 000006_document_and_qr_verification.down.sql

DROP TABLE IF EXISTS simak_documents;
ALTER TABLE simak DROP COLUMN IF EXISTS qr_token;
ALTER TABLE simak DROP COLUMN IF EXISTS issued_at;
ALTER TABLE simak DROP COLUMN IF EXISTS issued_by;
