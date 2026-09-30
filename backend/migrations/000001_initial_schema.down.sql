-- Migration: 000001_initial_schema.down.sql
-- Drops all tables and types created by the up migration.

DROP TABLE IF EXISTS system_configurations;
DROP TABLE IF EXISTS item_categories;
DROP TABLE IF EXISTS workers;
DROP TABLE IF EXISTS user_roles;
DROP TABLE IF EXISTS users;
DROP TABLE IF EXISTS departments;
DROP TABLE IF EXISTS companies;
DROP TABLE IF EXISTS organizations;
DROP TABLE IF EXISTS role_permissions;
DROP TABLE IF EXISTS permissions;
DROP TABLE IF EXISTS roles;

DROP TYPE IF EXISTS discrepancy_status;
DROP TYPE IF EXISTS discrepancy_type;
DROP TYPE IF EXISTS inspection_result;
DROP TYPE IF EXISTS simak_status;
