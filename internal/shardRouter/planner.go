package shardrouter

import (
	pg_query "github.com/pganalyze/pg_query_go/v5"
)

 
type Planner struct {
	cfg    RouterConfig
	hasher *Hasher
	ring   *Ring
}

func NewPlanner(
	cfg RouterConfig,
	hasher *Hasher,
	ring *Ring,
) *Planner {
	return &Planner{
		cfg:    cfg,
		hasher: hasher,
		ring:   ring,
	}
}

func (p *Planner) Plan(
	node pg_query.Node,
	table string,
	shardKey string,
) *RoutingPlan {

	pred, err := ExtractShardPredicate(node, table, shardKey)
	if err != nil {
		return &RoutingPlan{
			Mode:        RoutingModeRejected,
			Reason:      err.Message,
			RejectError: err,
		}
	}

	hashes := make([]HashValue, 0, len(pred.Values))
	for _, v := range pred.Values {
		hashes = append(hashes, p.hasher.Hash(v))
	}

	shards := p.ring.LocateShards(hashes)

	if len(shards) == 0 {
		return &RoutingPlan{
			Mode:   RoutingModeRejected,
			Reason: "no shards resolved for shard key",
			RejectError: &RoutingError{
				Code:    ErrInvalid,
				Message: "no shards resolved",
			},
		}
	}
	//Prevents expensive “scatter/gather”
	if len(shards) > 1 && len(shards) > p.cfg.MaxShardFanout {
		return &RoutingPlan{
			Mode:   RoutingModeRejected,
			Reason: "shard fanout exceeded",
			RejectError: &RoutingError{
				Code:    ErrFanoutExceeded,
				Message: "query touches too many shards",
			},
		}
	}

	targets := make([]ShardTarget, 0, len(shards))
	for _, sid := range shards {
		targets = append(targets, sid)
	}

	mode := RoutingModeSingle
	if len(targets) > 1 {
		mode = RoutingModeMulti
	}

	return &RoutingPlan{
		Mode:    mode,
		Targets: targets,
		Reason:  "shard key resolved successfully",
	}
}
//logic : 
// Extract shard key from SQL query (AST-aware).
// Reject if unsupported or missing shard key.
// Hash shard key value(s).
// Locate shards using consistent hash ring.
// Reject if no shards or fanout limit exceeded.
// Return RoutingPlan with shard targets and mode.