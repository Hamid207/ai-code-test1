-- Add google_id column to users table
ALTER TABLE users ADD COLUMN IF NOT EXISTS google_id VARCHAR(255) UNIQUE;

-- Create index on google_id for faster lookups (UNIQUE constraint creates an index automatically)
-- This comment is for documentation - the UNIQUE constraint above creates the index

-- Make apple_id nullable since users can now sign in with either Apple or Google
ALTER TABLE users ALTER COLUMN apple_id DROP NOT NULL;

-- Add constraint to ensure at least one of apple_id or google_id is present
ALTER TABLE users ADD CONSTRAINT check_auth_provider
    CHECK (
        (apple_id IS NOT NULL) OR
        (google_id IS NOT NULL)
    );
