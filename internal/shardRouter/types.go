package shardrouter

// ---------- Routing modes ----------

type RoutingMode int

const (
	RoutingModeInvalid RoutingMode = iota
	RoutingModeSingle
	RoutingModeMulti
	RoutingModeBroadcast
	RoutingModeRejected
)

type ShardID string

// Routing plan

type RoutingPlan struct {
	Mode        RoutingMode
	Targets     []ShardTarget
	Reason      string
	RejectError *RoutingError
}

type ShardTarget struct {
	ShardID ShardID
}

type PredicateType int

const (
	PredicateInvalid PredicateType = iota
	PredicateEquals
	PredicateIn
	PredicateRange
)

type ExtractedPredicate struct {
	Table      string
	Column     string
	Type       PredicateType
	Values     []any
	RangeStart any
	RangeEnd   any
}

type RoutingContext struct {
	RequestID string
}

type ShardKeyMetadata struct {
	Table  string
	Column string
}

 

type RoutingDecision struct {
	Mode     RoutingMode
	ShardIDs []ShardID
}
