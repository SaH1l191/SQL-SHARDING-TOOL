
SQL is:
string-based
ambiguous
order-dependent

Example:
ALTER TABLE users ADD email TEXT;
This doesn’t explicitly tell you:
current schema state
conflicts
dependencies
 So you convert it into something deterministic: AST + LogicalSchema

2. You need a “source of truth”
Your metadata tables (columns, fk_edges) become:
 The canonical schema state (not the database itself)
Why?

Because real DB:
is hard to inspect efficiently
varies across shards
may drift

3. Sharding REQUIREMENT
Sharding systems (like what Vitess or CockroachDB do) need:
full schema graph
FK relationships
column-level understanding
 Raw SQL execution cannot give this safely.
 HOW Your Pipeline Works (Deep Explanation)
 STEP 1 — SQL → AST
parseDDL(sql)
What’s really happening:

You use a PostgreSQL parser (like pg_query)

It builds a tree structure

Why AST?
Because:

Trees are structured, SQL strings are not.

Example:

CREATE TABLE users (id INT PRIMARY KEY);

Becomes:

CreateStmt
 ├── table: users
 └── column: id (PK)

 Now your system can programmatically reason about SQL.

 STEP 2 — AST → LogicalSchema (delta)

You extract only the change.

Why delta?

Because:

Every SQL statement is a transformation, not a full schema.

Example:

ALTER TABLE users ADD email TEXT;

Delta:

users:
  + email

 This keeps operations:

small

composable

trackable

STEP 3 — Metadata → LogicalSchema (base)

You reconstruct current schema from your DB metadata.
Why not query Postgres directly every time?
Because:
slow
inconsistent across shards
harder to control


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
Why this is critical:

Because now  can:
 Validate BEFORE execution
duplicate column?
missing parent table for FK?
type mismatch?
 Detect conflicts
ADD COLUMN id INT

→ already exists 
 This is exactly what advanced systems do internally:
Vitess schema tracker
CockroachDB planner

 STEP 5 — Flatten

You convert:

Graph → Tables

Why?

Because relational DBs store:

rows, not trees

Logical:
users:
  id
  email
orders.user_id → users.id
Flattened:
columns table
fk_edges table

 This makes it:

queryable

indexable

scalable

 STEP 6 — Replace Metadata
DELETE + INSERT
Why replace instead of update?

Because:

Schema is state, not events

Benefits:

atomic

consistent snapshot

no partial updates

 What This Enables (REAL POWER)
1.  Pre-flight validation

Before running SQL:

check FK integrity

ensure table exists

prevent bad migrations

2.  Schema evolution engine

You can:

diff schemas

generate migrations

rollback safely

3.  Sharding intelligence

Because you know:

orders.user_id → users.id

You can infer:

co-location

shard keys

4.  Query routing (future)

Incoming query:

SELECT * FROM orders WHERE user_id = 10;

Your system knows:

shard key = user_id

route to correct shard

5.  Dependency graph

You’ve built:

users → orders → payments

Used for:

execution order

cascading operations

safe deletes

The BIG Insight

What you built is:

A compiler pipeline for SQL schema

Just like a programming language:

Code → AST → IR → Optimized → Executed

Your version:

SQL → AST → LogicalSchema → Metadata DB
 Why this is NECESSARY (not optional)

If you want:

sharding

multi-tenant systems

safe migrations

distributed SQL

