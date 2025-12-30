-- Drop indexes
DROP INDEX IF EXISTS idx_refresh_tokens_application_id;
DROP INDEX IF EXISTS idx_users_application_id;
DROP INDEX IF EXISTS idx_applications_domain;
DROP INDEX IF EXISTS idx_applications_api_key_prefix;

-- Drop foreign key columns
ALTER TABLE refresh_tokens DROP COLUMN IF EXISTS application_id;
ALTER TABLE users DROP COLUMN IF EXISTS application_id;

-- Drop tables
DROP TABLE IF EXISTS token_scopes;
DROP TABLE IF EXISTS application_scopes;
DROP TABLE IF EXISTS scopes;
DROP TABLE IF EXISTS applications;

