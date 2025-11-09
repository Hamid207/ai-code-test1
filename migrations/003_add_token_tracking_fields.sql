-- Add token tracking and security fields
ALTER TABLE refresh_tokens
ADD COLUMN IF NOT EXISTS token_id VARCHAR(64),  -- JWT ID (jti) for tracking
ADD COLUMN IF NOT EXISTS ip_address VARCHAR(45),  -- IPv4/IPv6 support
ADD COLUMN IF NOT EXISTS user_agent TEXT,  -- Browser/device info
ADD COLUMN IF NOT EXISTS token_family VARCHAR(64);  -- For token family tracking

-- Create index on token_id for faster lookups
CREATE INDEX IF NOT EXISTS idx_refresh_tokens_token_id ON refresh_tokens(token_id);

-- Create index on token_family for stolen token detection
CREATE INDEX IF NOT EXISTS idx_refresh_tokens_family ON refresh_tokens(token_family);

-- Add comment for documentation
COMMENT ON COLUMN refresh_tokens.token_id IS 'JWT ID (jti) from the token claims - unique identifier';
COMMENT ON COLUMN refresh_tokens.token_family IS 'Family ID for tracking token chains - detects stolen tokens';
COMMENT ON COLUMN refresh_tokens.ip_address IS 'IP address where token was created';
COMMENT ON COLUMN refresh_tokens.user_agent IS 'User agent string for security monitoring';
