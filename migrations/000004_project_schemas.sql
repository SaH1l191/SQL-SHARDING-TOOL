-- +goose Up
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

-- +goose Down
DROP INDEX IF EXISTS idx_project_schemas_state;
DROP INDEX IF EXISTS idx_project_schemas_project;
DROP TABLE IF EXISTS project_schemas;
DROP TYPE IF EXISTS schema_state;
