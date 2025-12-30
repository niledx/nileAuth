-- Applications table: represents different systems/applications using the auth service
CREATE TABLE IF NOT EXISTS applications (
  id SERIAL PRIMARY KEY,
  name TEXT NOT NULL,
  domain TEXT NOT NULL,
  api_key_hash TEXT UNIQUE NOT NULL,
  api_key_prefix TEXT NOT NULL, -- First 8 chars for identification
  rate_limit_per_minute INTEGER DEFAULT 100,
  allowed_origins TEXT[], -- CORS origins
  active BOOLEAN DEFAULT true,
  created_at TIMESTAMPTZ DEFAULT now(),
  updated_at TIMESTAMPTZ DEFAULT now()
);

-- Scopes table: define permissions/scopes that can be assigned to tokens
CREATE TABLE IF NOT EXISTS scopes (
  id SERIAL PRIMARY KEY,
  name TEXT UNIQUE NOT NULL,
  description TEXT,
  created_at TIMESTAMPTZ DEFAULT now()
);

-- Application scopes: which scopes each application can request
CREATE TABLE IF NOT EXISTS application_scopes (
  application_id INTEGER NOT NULL REFERENCES applications(id) ON DELETE CASCADE,
  scope_id INTEGER NOT NULL REFERENCES scopes(id) ON DELETE CASCADE,
  PRIMARY KEY (application_id, scope_id)
);

-- Token scopes: track which scopes are assigned to access tokens
CREATE TABLE IF NOT EXISTS token_scopes (
  token_id TEXT NOT NULL, -- JWT ID or refresh token
  scope_id INTEGER NOT NULL REFERENCES scopes(id) ON DELETE CASCADE,
  PRIMARY KEY (token_id, scope_id)
);

-- Add application_id to users for multi-tenant support (optional)
ALTER TABLE users ADD COLUMN IF NOT EXISTS application_id INTEGER REFERENCES applications(id) ON DELETE SET NULL;

-- Add application_id to refresh_tokens
ALTER TABLE refresh_tokens ADD COLUMN IF NOT EXISTS application_id INTEGER REFERENCES applications(id) ON DELETE SET NULL;

-- Add indexes for performance
CREATE INDEX IF NOT EXISTS idx_applications_api_key_prefix ON applications(api_key_prefix);
CREATE INDEX IF NOT EXISTS idx_applications_domain ON applications(domain);
CREATE INDEX IF NOT EXISTS idx_users_application_id ON users(application_id);
CREATE INDEX IF NOT EXISTS idx_refresh_tokens_application_id ON refresh_tokens(application_id);

-- Insert default scopes
INSERT INTO scopes (name, description) VALUES
  ('read:user', 'Read user information'),
  ('write:user', 'Modify user information'),
  ('admin:users', 'Admin access to user management'),
  ('admin:applications', 'Admin access to application management')
ON CONFLICT (name) DO NOTHING;

