CREATE TABLE IF NOT EXISTS otps (
    id SERIAL PRIMARY KEY,
    phone_number VARCHAR(15) NOT NULL,
    code VARCHAR(10) NOT NULL,
    expires_at TIMESTAMP WITH TIME ZONE NOT NULL,
    is_used BOOLEAN DEFAULT FALSE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    used_at TIMESTAMP WITH TIME ZONE
);

-- Index for fast lookups by phone number and code
CREATE INDEX IF NOT EXISTS idx_otps_phone_number_code ON otps(phone_number, code);

-- Index for cleanup of expired OTPs
CREATE INDEX IF NOT EXISTS idx_otps_expires_at ON otps(expires_at);

-- Index for filtering unused OTPs
CREATE INDEX IF NOT EXISTS idx_otps_is_used ON otps(is_used);

-- Composite index for active OTP lookups
CREATE INDEX IF NOT EXISTS idx_otps_active_lookup ON otps(phone_number, is_used, expires_at);
