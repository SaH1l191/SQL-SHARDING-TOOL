 # SQL Sharding Tool

## Dependencies

```bash
go get github.com/golang-migrate/migrate/v4 \
    github.com/google/uuid \
    github.com/lib/pq \
    github.com/pganalyze/pg_query_go/v5 \
    github.com/joho/godotenv
```

---

## Graph Architecture

### DDL → AST → Graph

```
DDL (SQL)
  ↓
AST (pg_query)
  ↓
Graph Extraction (nodes + edges)
  ↓
Graph Storage (DB as cache)
  ↓
Graph Analysis (fanout)
  ↓
Graph Scoring (centrality + heuristics)
  ↓
Shard Key Decision
```

---

## Data Model

### Nodes (Vertices)

Each column becomes a node:

```go
repository.Column{
    TableName,
    ColumnName,
}
```

### Edges (Relationships)

Foreign keys become edges:

```go
repository.FkEdges{
    ChildTable,
    ChildColumn,
    ParentTable,
    ParentColumn,
}
```

This represents:
```
(child_table.child_column) ───▶ (parent_table.parent_column)
```
Directed edge from child to parent.

---

## Design Principles

### DB-Friendly Flat Data

- Not an in-memory graph structure
- Flat representation (not adjacency list)
- AST traversal → Graph extraction

### Replace vs Update

Graph must be a **consistent snapshot** — replace instead of update every time. This avoids:
- Orphan edges
- Partial updates
- Schema drift bugs

---

## Graph Algorithms

### Algorithm #1: Fanout — `O(n)`

**File:** `fanout.go`

| Metric | Implementation |
|--------|---------------|
| In-degree | `s.IncomingFkCount++` |
| Unique neighbor count | `childTableSets[parent][fk.ChildTable]` |

### Algorithm #2: Ranking — `O(n³)`

**File:** `ranker.go`

| Factor | Weight |
|--------|--------|
| In-degree centrality | `IncomingFkCount * 10` |
| Breadth of influence | `ReferencingTableCount * 5` |
| Reverse edge signal (FK child) | `+20` if `isFKChild` |
| Root Affinity (2nd-order traversal) | child → parent → parent's importance |
| Primary Key bonus | `+10` |
| Penalty (text data) | `-15` |

**Propagation of centrality** — like PageRank-lite

---

## Graph Construction Pipeline

```
1. AST traversal
        ↓
2. Builds edge list
        ↓
3. Degree Centrality (IncomingFkCount)
        ↓
4. Neighborhood Analysis (ReferencingTableCount)
        ↓
5. Reverse Edge Detection (isFKChild)
        ↓
6. 2-Hop Traversal
   child → parent → fanout(parent)
        ↓
7. Weighted Scoring Model
```

- **Second-order dependency analysis**
- **Heuristic centrality ranking** (similar intuition to PageRank)

---

## Sorting

```go
sort.SliceStable
```

Ensures:
- Deterministic output
- Avoids Go map randomness

---

## Layered Architecture

The system is decoupled into 4 perfect layers:

| Layer | Purpose | Data Flow |
|-------|---------|-----------|
| **1 — Extraction** | DDL → AST | `(columns, edges)` |
| **2 — Storage** | DB as graph cache | `edges → DB` |
| **3 — Computation** | Stateless recomputation | `edges → fanout stats` |
| **4 — Decision** | Shard key selection | `fanout + metadata → shard key` |

### Benefits

- Stateless computation — recompute anytime
- DB as graph cache — no need to rebuild from SQL every time
- `O(n)` algorithms — scales well
- Extensible — easily add:
  - Join cost estimation
  - Query routing
  - Lineage tracking

---

## Consistent Hashing Ring

### FNV-1a Hash

- **Deterministic**: simple XOR + multiply
- **Good distribution**: spreads values well across 64-bit space

### Virtual Nodes

Each real shard is represented by many virtual nodes (150 duplicates for load distribution).

**Example:** 3 shards × 150 = 450 points → keys distributed much more evenly

```
┌─────────────────────────────────────────────────────────┐
│                    HASH RING (0 - 2⁶⁴-1)                │
├─────────────────────────────────────────────────────────┤
│  S0.v0  S2.v1  S1.v2  S0.v3  S2.v4  S1.v0  ... S2.v149 │
│   ●      ●      ●      ●      ●      ●           ●     │
└─────────────────────────────────────────────────────────┘
```

### Data Structures

- `vnodes`: sorted list of all virtual nodes (this IS the ring)
- `shards`: map of `shardID → actual shard`

### Lookup

Binary search for first vnode where `vnode.hash ≥ key.hash`

**Complexity:** `O(N × V log(N × V))` where N = shards, V = 150

### Minimal Data Movement

When adding/removing shards:
- Only nearby keys move
- `~ 1/N` keys affected

---

## Join Strategies

| Strategy | Description |
|----------|-------------|
| **Shuffle** | Cross-shard join requiring data movement |
| **Non-collocated** | Tables on different shards, requires network transfer |

---

## Architecture Diagram

```
┌─────────────────────────────────────────────────────────────────┐
│                     SQL SHARDING TOOL                           │
├─────────────────────────────────────────────────────────────────┤
│  ┌─────────────┐  ┌─────────────┐  ┌─────────────────────────┐  │
│  │   DDL Input │  │   AST Walk  │  │   Graph Extraction      │  │
│  │   (SQL)     │──▶│   (pg_query)│──▶│   Nodes + Edges         │  │
│  └─────────────┘  └─────────────┘  └─────────────────────────┘  │
│                                               │                 │
│                                               ▼                 │
│                                 ┌─────────────────────────┐     │
│                                 │   PostgreSQL (Cache)    │     │
│                                 │   • columns table       │     │
│                                 │   • fk_edges table      │     │
│                                 └─────────────────────────┘     │
│                                               │                 │
│                    ┌──────────────────────────┼──────────┐   │
│                    │                          │          │   │
│                    ▼                          ▼          ▼   │
│           ┌─────────────┐            ┌─────────────┐  ┌────────┴─┐│
│           │  Fanout     │            │  Ranker     │  │  Shard   ││
│           │  O(n)       │            │  O(n³)      │  │  Router  ││
│           └─────────────┘            └─────────────┘  └────┬─────┘│
│                                                             │     │
│                                      ┌──────────────────────┘     │
│                                      ▼                            │
│                         ┌─────────────────────┐                   │
│                         │  Consistent Hashing │                   │
│                         │  • 150 vnodes/shard │                   │
│                         │  • FNV-1a hash      │                   │
│                         └─────────────────────┘                   │
└─────────────────────────────────────────────────────────────────┘
```

---

## Screenshots

### Architecture Overview
![Architecture](image.png)

### Join Strategy
![Join Strategy](image-1.png)