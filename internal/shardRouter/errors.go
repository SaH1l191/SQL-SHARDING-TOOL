package shardrouter
 

type RoutingErrorCode int

const (
	ErrInvalid RoutingErrorCode = iota
	ErrNoShardKey
	ErrShardKeyNotInQuery
	ErrUnsupportedPredicate
	ErrPolicyViolation
	ErrFanoutExceeded
)

type RoutingError struct {
	Code    RoutingErrorCode
	Message string
}

func (e *RoutingError) Error() string {
	return e.Message
}
