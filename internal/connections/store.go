package connections

import (
	"database/sql"
	"fmt"
	"sync"
)

// ConnectionStore
//       │
//       ▼
// map[projectID] → map[shardID] → *sql.DB

// sync.RWMutex allows:
// many concurrent readers
// one writer
type ConnectionStore struct {
	mu    sync.RWMutex
	conns map[string]map[string]*sql.DB
}

func NewConnectionStore() *ConnectionStore {
	return &ConnectionStore{
		conns: make(map[string]map[string]*sql.DB),
	}
}

// thread safe concurrent reads
func (c *ConnectionStore) Get(projectID, shardID string) (*sql.DB, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	shards, ok := c.conns[projectID]
	if !ok {
		return nil, fmt.Errorf("project %s not found", projectID)
	}

	db, ok := shards[shardID]
	if !ok {
		return nil, fmt.Errorf("shard %s not found for project %s", shardID, projectID)
	}
	return db, nil
}

// thread safe concurrent writes
func (c *ConnectionStore) Set(projectID, shardID string, db *sql.DB) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.conns[projectID] == nil {
		c.conns[projectID] = make(map[string]*sql.DB)
	}
	oldDbConn := c.conns[projectID][shardID]
	if oldDbConn != nil {
		oldDbConn.Close()
	}
	c.conns[projectID][shardID] = db
}
