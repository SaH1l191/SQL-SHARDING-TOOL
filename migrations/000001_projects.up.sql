CREATE TABLE projects (
    id UUID PRIMARY KEY,
    name TEXT UNIQUE NOT NULL,
    description TEXT,
    shard_count INT DEFAULT 0,

    status TEXT NOT NULL DEFAULT 'inactive',

    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW(),

    CONSTRAINT chk_project_status
        CHECK (status IN ('active','inactive'))
);