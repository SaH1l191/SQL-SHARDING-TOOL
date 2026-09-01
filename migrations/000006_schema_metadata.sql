-- +goose Up
CREATE TABLE columns (
    project_id UUID NOT NULL,
    table_name TEXT NOT NULL,
    column_name TEXT NOT NULL,

    data_type TEXT NOT NULL,
    nullable BOOLEAN NOT NULL,
    is_primary_key BOOLEAN NOT NULL DEFAULT FALSE,

    PRIMARY KEY (project_id, table_name, column_name)
);

CREATE INDEX idx_columns_project_table
    ON columns(project_id, table_name);

CREATE TABLE fk_edges (
    project_id UUID NOT NULL,

    parent_table TEXT NOT NULL,
    parent_column TEXT NOT NULL,

    child_table TEXT NOT NULL,
    child_column TEXT NOT NULL,

    PRIMARY KEY (
        project_id,
        parent_table,
        parent_column,
        child_table,
        child_column
    )
);

CREATE INDEX idx_fk_edges_project_parent
    ON fk_edges(project_id, parent_table, parent_column);

CREATE INDEX idx_fk_edges_project_child
    ON fk_edges(project_id, child_table, child_column);

CREATE TABLE table_shard_keys (
    project_id UUID NOT NULL,
    table_name TEXT NOT NULL,
    shard_key_column TEXT NOT NULL,
    is_manual_override BOOLEAN NOT NULL DEFAULT FALSE,
    updated_at TIMESTAMP NOT NULL DEFAULT now(),
    PRIMARY KEY (project_id, table_name)
);

CREATE INDEX idx_table_shard_keys_project
    ON table_shard_keys(project_id);

-- +goose Down
DROP INDEX IF EXISTS idx_table_shard_keys_project;
DROP TABLE IF EXISTS table_shard_keys;
DROP INDEX IF EXISTS idx_fk_edges_project_child;
DROP INDEX IF EXISTS idx_fk_edges_project_parent;
DROP TABLE IF EXISTS fk_edges;
DROP INDEX IF EXISTS idx_columns_project_table;
DROP TABLE IF EXISTS columns;
