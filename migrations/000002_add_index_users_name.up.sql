-- Add index on name column for faster queries
CREATE INDEX IF NOT EXISTS idx_users_name ON users(name);
