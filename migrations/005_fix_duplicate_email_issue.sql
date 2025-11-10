-- Fix critical issue: Add UNIQUE constraint on email to prevent duplicate accounts
-- This ensures one email = one user account, regardless of auth provider

-- Step 1: Check for existing duplicate emails before adding constraint
-- If duplicates exist, this will fail and alert you to clean data first
DO $$
DECLARE
    duplicate_count INTEGER;
BEGIN
    SELECT COUNT(*) INTO duplicate_count
    FROM (
        SELECT email, COUNT(*) as cnt
        FROM users
        GROUP BY email
        HAVING COUNT(*) > 1
    ) duplicates;

    IF duplicate_count > 0 THEN
        RAISE EXCEPTION 'Cannot add UNIQUE constraint: % duplicate email(s) found. Clean data first.', duplicate_count;
    END IF;
END $$;

-- Step 2: Add UNIQUE constraint on email
ALTER TABLE users ADD CONSTRAINT users_email_unique UNIQUE (email);

-- Step 3: Create index on email for faster lookups (if not already exists)
-- Note: UNIQUE constraint already creates an index, but we'll be explicit
CREATE INDEX IF NOT EXISTS idx_users_email_unique ON users(email);

-- Step 4: Add comment for documentation
COMMENT ON CONSTRAINT users_email_unique ON users IS
    'Ensures one user account per email address across all authentication providers (Apple, Google, etc.)';
