-- Users table (for authentication)
DROP TABLE IF EXISTS users CASCADE;
CREATE TABLE IF NOT EXISTS users (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    username TEXT UNIQUE NOT NULL,
    password_hash TEXT NOT NULL,
    avatar_data TEXT, -- Base64 encoded avatar image
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- Peers table (for discovery)
DROP TABLE IF EXISTS peers CASCADE;
CREATE TABLE IF NOT EXISTS peers (
    user_id UUID PRIMARY KEY,
    ip_address TEXT NOT NULL,
    port INTEGER NOT NULL,
    status TEXT DEFAULT 'offline',
    status_message TEXT DEFAULT '',
    server_name TEXT DEFAULT '',
    server_icon TEXT DEFAULT '',
    server_description TEXT DEFAULT '',
    last_seen TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY(user_id) REFERENCES users(id)
);

-- Force drop to ensure new schema is applied (Dev only)
DROP TABLE IF EXISTS channels CASCADE;

CREATE TABLE channels (
    id SERIAL PRIMARY KEY,
    name TEXT UNIQUE NOT NULL,
    owner_id UUID NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY(owner_id) REFERENCES users(id)
);

CREATE TABLE IF NOT EXISTS channel_members (
    channel_id INTEGER NOT NULL,
    user_id UUID NOT NULL,
    last_seen TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (channel_id, user_id),
    FOREIGN KEY(channel_id) REFERENCES channels(id),
    FOREIGN KEY(user_id) REFERENCES users(id)
);
