SQL is:

string-based
ambiguous
order-dependent

Example:

ALTER TABLE users ADD email TEXT;

This does not explicitly convey:

current schema state
conflicts
dependencies

So it is converted into something deterministic: AST + LogicalSchema.

2. A “source of truth” is needed

Metadata tables (columns, fk_edges) become the canonical schema state (not the database itself).

Reason:

The real database is hard to inspect efficiently.
It varies across shards.
It may drift.
3. Sharding REQUIREMENT

Sharding systems (like those in Vitess or CockroachDB) require:

full schema graph
foreign key relationships
column-level understanding

Raw SQL execution cannot provide this safely.

How the Pipeline Works (Deep Explanation)
STEP 1 — SQL → AST

parseDDL(sql)

What happens:

A PostgreSQL parser (like pg_query) builds a tree structure.

Why AST?

Trees are structured, whereas SQL strings are not.

Example:

CREATE TABLE users (id INT PRIMARY KEY);

Becomes:

CreateStmt
 ├── table: users
 └── column: id (PK)

Now the system can programmatically reason about SQL.

STEP 2 — AST → LogicalSchema (delta)

Only the change is extracted.

Why delta?

Every SQL statement is a transformation, not a full schema.

Example:

ALTER TABLE users ADD email TEXT;

Delta:

users:
  + email

Benefits:

Keeps operations small
Composable
Trackable
STEP 3 — Metadata → LogicalSchema (base)

The current schema is reconstructed from DB metadata.

Reason for not querying PostgreSQL directly every time:

Slow
Inconsistent across shards
Harder to control
STEP 4 — Merge (THE HEART)

MergeLogicalSchema(base, delta)

Example:

Base:

users:
  id

Delta:

+ email

After merge:

users:
  id
  email

Why critical:

Validates before execution
Detects duplicates, missing parent tables for FK, type mismatches

Example:

ADD COLUMN id INT → already exists

This is exactly what advanced systems do internally:

Vitess schema tracker
CockroachDB planner
STEP 5 — Flatten

The graph is converted to tables.

Reason:

Relational databases store rows, not trees.

Logical:

users:
  id
  email
orders.user_id → users.id

Flattened:

columns table
fk_edges table

Benefits:

Queryable
Indexable
Scalable
STEP 6 — Replace Metadata
DELETE + INSERT

Why replace instead of update:

Schema represents state, not events

Benefits:

Atomic
Consistent snapshot
No partial updates
What This Enables (REAL POWER)
Pre-flight validation
Check FK integrity
Ensure table exists
Prevent bad migrations
Schema evolution engine
Diff schemas
Generate migrations
Rollback safely
Sharding intelligence
Orders.user_id → users.id
Can infer co-location and shard keys
Query routing (future)
Incoming query:
SELECT * FROM orders WHERE user_id = 10;
System knows shard key = user_id
Routes to correct shard
Dependency graph
users → orders → payments
Used for execution order, cascading operations, safe deletes
The BIG Insight

This creates a compiler pipeline for SQL schema, analogous to a programming language:

Code → AST → IR → Optimized → Executed

The pipeline:

SQL → AST → LogicalSchema → Metadata DB
Why This is NECESSARY (not optional)

Required for:

Sharding
Multi-tenant systems
Safe migrations
Distributed SQL