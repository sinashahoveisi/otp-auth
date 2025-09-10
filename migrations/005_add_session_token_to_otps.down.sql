-- Remove session_token column (unique constraint will be dropped automatically)
ALTER TABLE otps DROP COLUMN session_token;
