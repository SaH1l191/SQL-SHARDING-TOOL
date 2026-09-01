-- +goose Up
CREATE TABLE shards (
    id UUID PRIMARY KEY,
    project_id UUID NOT NULL,
    shard_index INTEGER NOT NULL,

    status TEXT NOT NULL DEFAULT 'inactive',

    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),

    CONSTRAINT fk_shards_project
        FOREIGN KEY (project_id)
        REFERENCES projects(id)
        ON DELETE CASCADE,

    CONSTRAINT uq_project_shard_index
        UNIQUE (project_id, shard_index),

    CONSTRAINT chk_shard_status
        CHECK (status IN ('active', 'inactive'))
);

CREATE INDEX idx_shards_project_id
    ON shards(project_id);

-- +goose Down
DROP INDEX IF EXISTS idx_shards_project_id;
DROP TABLE IF EXISTS shards;
