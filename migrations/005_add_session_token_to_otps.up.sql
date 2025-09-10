-- Add session_token column to otps table
ALTER TABLE otps ADD COLUMN session_token VARCHAR(255) UNIQUE;

-- Index for session token lookups (unique constraint creates index automatically)
-- CREATE INDEX IF NOT EXISTS idx_otps_session_token ON otps(session_token);
