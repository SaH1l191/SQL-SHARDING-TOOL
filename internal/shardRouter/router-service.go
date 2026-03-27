package shardrouter

import (
	"context"
	"fmt"
	"sort"
	"sqlsharder/internal/repository"
	"sqlsharder/pkg/logger"

	pg_query "github.com/pganalyze/pg_query_go/v5"
)

type RouterService struct {
	shardKeysRepo *repository.ShardKeysRepository
	shardRepo     *repository.ShardRepository
	cfg           RouterConfig
}

func NewRouterService(
	shardKeysRepo *repository.ShardKeysRepository,
	shardRepo *repository.ShardRepository,
	cfg RouterConfig,
) *RouterService {
	fmt.Println("[NewRouterService] creating new RouterService with cfg:", cfg)
	return &RouterService{
		shardKeysRepo: shardKeysRepo,
		shardRepo:     shardRepo,
		cfg:           cfg,
	}
}

func (s *RouterService) RouteSQL(
	ctx context.Context,
	projectID string,
	sql string,
) (*RoutingPlan, error) {
	logger.Logger.Info("router entry reached")
	fmt.Println("[RouteSQL] Router entry reached")
	fmt.Println("[RouteSQL] projectID:", projectID)
	fmt.Println("[RouteSQL] sql:", sql)

	parseResult, err := pg_query.Parse(sql)
	if err != nil {
		fmt.Println("[RouteSQL] SQL parse error:", err)
		return nil, fmt.Errorf("sql parse error: %w", err)
	}

	fmt.Println("[RouteSQL] parseResult:", parseResult)

	if len(parseResult.Stmts) != 1 {
		fmt.Println("[RouteSQL] multiple statements detected, not supported")
		return nil, fmt.Errorf("only single-statement queries supported")
	}

	rawStmt := parseResult.Stmts[0]
	fmt.Println("[RouteSQL] rawStmt:", rawStmt)

	tableName, node, err := extractTableAndNode(rawStmt)
	if err != nil {
		fmt.Println("[RouteSQL] extractTableAndNode error:", err)
		return nil, err
	}

	fmt.Println("[RouteSQL] tableName:", tableName)
	fmt.Println("[RouteSQL] node:", node)

	shardKeys, err := s.shardKeysRepo.FetchShardKeysByProjectID(ctx, projectID)
	if err != nil {
		fmt.Println("[RouteSQL] FetchShardKeysByProjectID error:", err)
		return nil, err
	}

	fmt.Println("[RouteSQL] shardKeys:", shardKeys)

	shardKeyColumn := ""
	for _, k := range shardKeys {
		if k.TableName == tableName {
			shardKeyColumn = k.ShardKeyColumn
			break
		}
	}

	fmt.Println("[RouteSQL] shardKeyColumn:", shardKeyColumn)

	if shardKeyColumn == "" {
		fmt.Println("[RouteSQL] no shard key defined for table", tableName)
		return &RoutingPlan{
			Mode: RoutingModeRejected,
			Reason: fmt.Sprintf(
				"no shard key defined for table %s",
				tableName,
			),
			RejectError: &RoutingError{
				Code:    ErrNoShardKey,
				Message: "shard key not found",
			},
		}, nil
	}

	shards, err := s.shardRepo.ShardList(ctx, projectID)
	if err != nil {
		fmt.Println("[RouteSQL] ShardList error:", err)
		return nil, err
	}

	fmt.Println("[RouteSQL] shards fetched:", shards)

	activeShards := make([]repository.Shard, 0)
	for _, sh := range shards {
		if sh.Status == "active" {
			activeShards = append(activeShards, sh)
		}
	}

	fmt.Println("[RouteSQL] activeShards:", activeShards)

	if len(activeShards) == 0 {
		fmt.Println("[RouteSQL] no active shards for project")
		return nil, fmt.Errorf("no active shards for project")
	}

	sort.Slice(activeShards, func(i, j int) bool {
		return activeShards[i].ShardIndex < activeShards[j].ShardIndex
	})

	fmt.Println("[RouteSQL] activeShards sorted:", activeShards)

	shardIDs := make([]ShardID, 0, len(activeShards))
	for _, sh := range activeShards {
		shardIDs = append(shardIDs, ShardID(sh.ID))
	}

	fmt.Println("[RouteSQL] shardIDs:", shardIDs)

	shardTargets := make([]ShardTarget, 0, len(activeShards))
	for _, sh := range activeShards {
		shardTargets = append(shardTargets, ShardTarget{
			ShardID: ShardID(sh.ID),
		})
	}

	fmt.Println("[RouteSQL] shardIDs:", shardIDs)

	ring, err := NewRing(shardTargets)
	if err != nil {
		fmt.Println("[RouteSQL] NewRing error:", err)
		return nil, fmt.Errorf("failed to create ring: %w", err)
	}
	hasher := NewHasher()

	fmt.Println("[RouteSQL] ring created:", ring)
	fmt.Println("[RouteSQL] hasher created:", hasher)

	planner := NewPlanner(s.cfg, hasher, ring)
	fmt.Println("[RouteSQL] planner created:", planner)

	plan := planner.Plan(*node, tableName, shardKeyColumn)
	fmt.Println("[RouteSQL] routing plan:", plan)

	return plan, nil
}

func extractTableAndNode(
	stmt *pg_query.RawStmt,
) (string, *pg_query.Node, error) {

	fmt.Println("[extractTableAndNode] stmt:", stmt)

	node := stmt.Stmt

	switch n := node.Node.(type) {

	case *pg_query.Node_SelectStmt:
		fmt.Println("[extractTableAndNode] Node_SelectStmt detected")
		from := n.SelectStmt.FromClause
		fmt.Println("[extractTableAndNode] FromClause:", from)
		if len(from) != 1 {
			fmt.Println("[extractTableAndNode] joins not supported")
			return "", nil, fmt.Errorf("joins not supported in v1")
		}
		rv := from[0].Node.(*pg_query.Node_RangeVar)
		fmt.Println("[extractTableAndNode] tableName extracted:", rv.RangeVar.Relname)
		return rv.RangeVar.Relname, node, nil

	case *pg_query.Node_InsertStmt:
		fmt.Println("[extractTableAndNode] Node_InsertStmt detected")
		return n.InsertStmt.Relation.Relname, node, nil

	case *pg_query.Node_UpdateStmt:
		fmt.Println("[extractTableAndNode] Node_UpdateStmt detected")
		return n.UpdateStmt.Relation.Relname, node, nil

	case *pg_query.Node_DeleteStmt:
		fmt.Println("[extractTableAndNode] Node_DeleteStmt detected")
		return n.DeleteStmt.Relation.Relname, node, nil

	default:
		fmt.Println("[extractTableAndNode] unsupported statement type")
		return "", nil, fmt.Errorf("unsupported statement type")
	}
}
