CREATE TABLE shard_connections (
    shard_id UUID PRIMARY KEY,

    host TEXT NOT NULL,
    port INTEGER NOT NULL,
    database_name TEXT NOT NULL,
    username TEXT NOT NULL,
    password TEXT NOT NULL,

    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),

    CONSTRAINT fk_shard_connections_shard
        FOREIGN KEY (shard_id)
        REFERENCES shards(id)
        ON DELETE CASCADE
);