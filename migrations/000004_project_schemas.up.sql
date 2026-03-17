CREATE TYPE schema_state AS ENUM (
    'draft',
    'pending',
    'applying',
    'applied',
    'failed'
);

CREATE TABLE project_schemas (
    id UUID PRIMARY KEY,
    project_id UUID NOT NULL,

    version INTEGER NOT NULL,

    state schema_state NOT NULL DEFAULT 'draft',

    ddl_sql TEXT NOT NULL,

    error_message TEXT,

    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    committed_at TIMESTAMPTZ,
    applied_at TIMESTAMPTZ,

    CONSTRAINT fk_project_schemas_project
        FOREIGN KEY (project_id)
        REFERENCES projects(id)
        ON DELETE CASCADE,

    CONSTRAINT uq_project_schema_version
        UNIQUE (project_id, version)
);

CREATE INDEX idx_project_schemas_project
ON project_schemas(project_id);

CREATE INDEX idx_project_schemas_state
ON project_schemas(project_id, state);

-- CREATE INDEX idx_project_schemas_project
-- ON project_schemas(project_id);

-- Speeds up queries that filter by project_id (e.g., “give me all schemas for project X”).
-- Speeds up queries that filter by project and state (e.g., “give me all pending schemas for project X”).