parseDDL(ddlText)                                  │
│    └─ pg_query.Parse() → raw AST                   │
│                                                     │
│  extractSchemaFromAST(ast)                          │
│    └─ walk CREATE TABLE nodes                       │
│    └─ pull out column names, types, nullable, PK    │
│    └─ pull out FOREIGN KEY constraints              │
│    └─ → LogicalSchema{Tables: map[name]*Table}      │
│                                                     │
│  Save to DB (replaces old data atomically):         │
│    column repo → ReplaceExistingColumns()           │
│    fk_edge repo → ReplaceProjectFKEdges()           │
└─────────────────────────────────────────────────────┘
        │
        │  LogicalSchema now in DB as columns + fk_edges
        ▼
┌─────────────────────────────────────────────────────┐
│  next : internal/shardkey/                        │
│  (reads from DB, writes to table_shard_keys)        │
│                                                     │
│  Stage 1 — Elimination                             │
│    Load columns from DB for this project            │
│    Remove candidates that are invalid shard keys:  │
│      - nullable columns (can't be NULL)             │
│      - boolean columns (only 2 values = data skew) │
│      - metadata: created_at, updated_at, deleted_at │
│    → CandidateSet: map[tableName][]ColumnRef        │
│                                                     │
│  Stage 2 — Fanout                                   │
│    Load fk_edges from DB for this project           │
│    For each remaining candidate column:             │
│      count: how many FK constraints point TO it?    │
│      count: how many distinct tables reference it?  │
│    → FanoutStats per column                         │
│                                                     │
│  Stage 3 — Ranking                                  │
│    Score each candidate per table:                  │
│      +10 per incoming FK reference                  │
│      +5  per distinct referencing table             │
│      +20 if column is itself a FK (ownership col)   │
│      +10 if is primary key                          │
│      -15 if type is text/varchar (bad hash spread)  │
│      +1  baseline                                   │
│    Pick highest score per table                     │
│    → []ShardKeyDecision                             │
│                                                     │
│  Save to DB:                                        │
│    shard_keys repo → ReplaceShardKeys()             │
└─────────────────────────────────────────────────────┘
        │
        │  table_shard_keys now knows: for table X,
        │  shard on column Y
        ▼
┌─────────────────────────────────────────────────────┐
│  next : internal/sqlrouter/                       │
│  (runs on every query)                              │
│                                                     │
│  RouteSQL(projectID, sqlText):                      │
│                                                     │
│  1. pg_query.Parse(sqlText) → AST                   │
│                                                     │
│  2. Extract table name from AST                     │
│     (SELECT/INSERT/UPDATE/DELETE → which table?)    │
│                                                     │
│  3. Look up shard key column for that table         │
│     shard_keys repo → GetShardKeyForTable()         │
│     e.g. "users table shards on user_id"            │
│                                                     │
│  4. Extract shard key VALUE from query              │
│     walk WHERE clause: find "user_id = ?"           │
│     walk INSERT VALUES: find position of user_id    │
│     → ExtractedPredicate{column, value}             │
│                                                     │
│  5. Hash the value                                  │
│     FNV-64a hash(value) → uint64                    │
│                                                     │
│  6. Map hash to shard                               │
│     shardIndex = hash % len(activeShards)           │
│     → ShardID                                       │
│                                                     │
│  7. Return RoutingPlan                              │
│     { Mode: Single, Targets: [{ShardID: "abc"}] }   │
│                                                     │
│  Special cases:                                     │
│    No shard key in WHERE → Broadcast (all shards)   │
│    WHERE x IN (1,2,3)   → Multi (several shards)    │
│    WHERE x=1 OR y=2     → Reject (ambiguous)        │
└─────────────────────────────────────────────────────┘
        │
        │  RoutingPlan says: send to shard "abc"
        ▼
┌─────────────────────────────────────────────────────┐
│  next: internal/executor/                        │
│                                                     │
│  Execute(ctx, projectID, sqlText, plan):            │
│                                                     │
│  for each target in plan.Targets:                   │
│    db = connStore.Get(projectID, shardID)           │
│    result = executeOnShard(ctx, db, sqlText)        │
│      → try QueryContext (SELECT → returns rows)     │
│      → fallback ExecContext (INSERT/UPDATE/DELETE)  │
│                                                     │
│  collect []ExecutionResult                          │
│  return to caller (HTTP handler)                    │
└─────────────────────────────────────────────────────┘
```

---

## Build order — exact sequence
```
Fix connectionSetup.go  (mysql → postgres)
      │
      ▼
repository/fk_edges.go        → go build, verify
repository/shard_keys.go      → go build, verify
      │
      ▼
internal/schema/models.go     → structs only, no logic
internal/schema/parser.go     → one function: parseDDL()
internal/schema/extract.go    → walk AST → LogicalSchema
internal/schema/builder.go    → BuildLogicalSchemaFromDDL()
      │ verify: parse "CREATE TABLE users (id UUID PRIMARY KEY)"
      │         → LogicalSchema.Tables["users"] has 1 column
      ▼
internal/shardkey/types.go
internal/shardkey/elimination.go
internal/shardkey/fanout.go
internal/shardkey/ranker.go
internal/shardkey/inference_service.go
      │ verify: users+orders DDL → users.id gets highest score
      ▼
internal/sqlrouter/types.go
internal/sqlrouter/hasher.go
internal/sqlrouter/ring.go
internal/sqlrouter/extractor.go
internal/sqlrouter/planner.go
internal/sqlrouter/router_service.go
      │ verify: "SELECT * FROM users WHERE user_id = 1"
      │         with 3 shards → RoutingModeSingle
      ▼
internal/executor/result.go
internal/executor/execute_shard.go
internal/executor/executor.go
      │
      ▼
Wire into app.go + expose POST /api/v1/query/execute
      │
      ▼
Shard monitor goroutine (last)