package shardrouter

import (
	"errors"
	"fmt"
	"hash/fnv"
	"sort"
)

var ErrNoActiveShards = errors.New("no active shards available")

// ring : {[ShardID,ShardIndex ],[],....}

const virtualNodesPerShard = 100

type virtualNode struct {
	hash    uint64 //hash value
	shardID string //shard id
	idx     int    //ring index
}

type Ring struct {
	vnodes []virtualNode
	shards map[string]ShardTarget // shdId->shdTarget
}

func NewRing(shards []ShardTarget) (*Ring, error) {
	if len(shards) == 0 {
		return nil, ErrNoActiveShards
	}
	r := &Ring{
		vnodes: make([]virtualNode, 0, len(shards)*virtualNodesPerShard),
		shards: make(map[string]ShardTarget),
	}
	for i, s := range shards {
		r.shards[string(s.ShardID)] = s
		for v := 0; v < virtualNodesPerShard; v++ {
			r.vnodes = append(r.vnodes, virtualNode{
				hash:    vNodeHash(string(s.ShardID), v),
				shardID: string(s.ShardID),
				idx:     i,
			})
		}
	}
	sort.Slice(r.vnodes, func(i, j int) bool {
		return r.vnodes[i].hash < r.vnodes[j].hash
	})
	return r, nil
}

func vNodeHash(shardID string, vnode int) uint64 {
	h := fnv.New64a()
	h.Write([]byte(fmt.Sprintf("%s:%d", shardID, vnode)))
	return h.Sum64()
}

func (r *Ring) LocateShard(hash HashValue) ShardTarget {
	h := uint64(hash)
	n := len(r.vnodes)

	// binary search: find first vnode with position >= h
	pos := sort.Search(n, func(i int) bool {
		return r.vnodes[i].hash >= h
	})

	if pos == n {
		pos = 0
	}
	return r.shards[r.vnodes[pos].shardID]
} 

func (r *Ring) LocateShards(hashes []HashValue) []ShardTarget {
	if len(hashes) == 0 {
		return nil
	}

	seen := make(map[string]struct{})
	result := make([]ShardTarget, 0, len(hashes))

	for _, h := range hashes {
		shard := r.LocateShard(h)
		shardIDStr := string(shard.ShardID)
		if _, ok := seen[shardIDStr]; ok {
			continue
		}
		seen[shardIDStr] = struct{}{}
		result = append(result, shard)
	}

	return result
}

func (r *Ring) Shards() []ShardTarget {
	result := make([]ShardTarget, 0, len(r.shards))
	for _, s := range r.shards {
		result = append(result, s)
	}
	return result
}
