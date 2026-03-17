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

-- p1 has 3 shards , each shard has a unique index
    CONSTRAINT uq_project_shard_index
        UNIQUE (project_id, shard_index),

    CONSTRAINT chk_shard_status
        CHECK (status IN ('active', 'inactive'))
);

-- most queries : select * from shards where project_id = 'p1'
-- lookup time : O(n) -> full table scan
-- with index: O(log n) -> binary search ( sql maintains internally )
CREATE INDEX idx_shards_project_id
    ON shards(project_id);

-- P1 → [pointer_to_S1, pointer_to_S2, pointer_to_S3]
-- P2 → [pointer_to_S4, pointer_to_S5]
-- The index is implemented as a B-tree.
-- Each leaf node contains the project_id value and the pointer(s) to rows.
-- Internal nodes just help quickly navigate the tree.