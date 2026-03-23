package shardrouter

import (
	"encoding/binary"
	"fmt"
	"hash/fnv"
)

// 64-bit hash value
type HashValue uint64

// hash function
type Hasher struct{}

func NewHasher() *Hasher {
	return &Hasher{}
}

//for each byte  : hash = hash XOR byte

func (h *Hasher) Hash(value any) HashValue {
	hasher := fnv.New64a() //FNV-1a hash (fast, deterministic)
	//Consistent byte representation → consistent hash
	switch v := value.(type) {
	case string:
		_, _ = hasher.Write([]byte(v))
	case int:
		binary.Write(hasher, binary.LittleEndian, int64(v))
	case int32:
		binary.Write(hasher, binary.LittleEndian, int64(v))
	case int64:
		binary.Write(hasher, binary.LittleEndian, v)
	case uint:
		binary.Write(hasher, binary.LittleEndian, uint64(v))
	case uint32:
		binary.Write(hasher, binary.LittleEndian, uint64(v))
	case uint64:
		binary.Write(hasher, binary.LittleEndian, v)
	default:
		_, _ = hasher.Write([]byte(fmt.Sprintf("%v", v)))
	}
	return HashValue(hasher.Sum64())
}

// Deterministic
// same input → same hash → same shard
// 2. Uniform distribution
// values spread evenly across shards
// 3. Fast
// runs on every query → must be cheap
// 4. Type-safe
