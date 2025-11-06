-- ==========================================
-- Migration: 002_seed_policies.sql
-- Purpose: Seed default system-wide auth policies
-- ==========================================

INSERT INTO auth_policies (name, value)
VALUES
  ('registration_mode', '"super_admin_only"'),
  ('allowed_roles_for_registration', '["admin"]'),
  ('allow_password_reset', '"true"'),
  ('require_email_verification', '"false"')
ON CONFLICT (name) DO NOTHING;
