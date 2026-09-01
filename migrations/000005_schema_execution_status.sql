-- +goose Up
CREATE TABLE schema_execution_status (
    id UUID PRIMARY KEY,

    schema_id UUID NOT NULL,
    shard_id UUID NOT NULL,

    state schema_state NOT NULL,

    error_message TEXT,

    executed_at TIMESTAMPTZ,

    CONSTRAINT fk_schema_execution_schema
        FOREIGN KEY (schema_id)
        REFERENCES project_schemas(id)
        ON DELETE CASCADE,

    CONSTRAINT fk_schema_execution_shard
        FOREIGN KEY (shard_id)
        REFERENCES shards(id)
        ON DELETE CASCADE,

    CONSTRAINT uq_schema_shard_execution
        UNIQUE (schema_id, shard_id)
);

CREATE INDEX idx_schema_execution_schema
    ON schema_execution_status(schema_id);

CREATE INDEX idx_schema_execution_shard
    ON schema_execution_status(shard_id);

-- +goose Down
DROP INDEX IF EXISTS idx_schema_execution_shard;
DROP INDEX IF EXISTS idx_schema_execution_schema;
DROP TABLE IF EXISTS schema_execution_status;
