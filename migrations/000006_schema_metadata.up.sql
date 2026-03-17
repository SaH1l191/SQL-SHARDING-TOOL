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

--     Example entry:
-- project_id	table_name	column_name	data_type	nullable	is_primary_key
-- 1	users	id	UUID	FALSE	TRUE
-- 1	users	name	TEXT	TRUE	FALSE
-- 1	users	email	TEXT	TRUE	FALSE


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


-- order contains user_id, which references users.id
-- A project has shards. A project has schema versions. 
-- Each schema version is applied to each shard, and we track the result in schema_execution_status.
-- A project has shards. A project has schema versions. 
-- Each schema version is applied to each shard, and we track the result in schema_execution_status.
-- orders.user_id is referencing users.id.
-- users.id is being referenced by orders.user_id.


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

-- Example entry:
-- project_id	table_name	shard_key_column	is_manual_override	updated_at
-- 1	users	id	FALSE	2026-03-16 12:00:00
-- 1	orders	user_id	TRUE	2026-03-16 12:05:00