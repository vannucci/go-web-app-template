-- Create users table if not exists
CREATE TABLE IF NOT EXISTS users (
    id SERIAL PRIMARY KEY,
    email VARCHAR(255) UNIQUE NOT NULL,
    password_hash VARCHAR(255) NOT NULL,
    name VARCHAR(255) NOT NULL,
    role VARCHAR(50) DEFAULT 'super_admin',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Create index on email for faster lookups
CREATE INDEX IF NOT EXISTS idx_users_email ON users(email);

-- Create index on user_tier for filtering
CREATE INDEX IF NOT EXISTS idx_users_tier ON users(user_tier);

-- Insert default admin user (password: Password123!)
-- Note: In production, use proper password hashing!
INSERT INTO users (email, password_hash, name, user_tier) 
VALUES (
    'admin@throtle.io', 
    '$2a$10$placeholder.hash.for.Password123!',  -- Replace with real hash
    'Administrator',
    'admin'
) ON CONFLICT (email) DO NOTHING;