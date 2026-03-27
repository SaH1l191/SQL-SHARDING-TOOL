package shardrouter
 
import pg_query "github.com/pganalyze/pg_query_go/v5"

func ExtractShardPredicate(node pg_query.Node, table string, shardKey string) (*ExtractedPredicate, *RoutingError) {

	switch n := node.Node.(type) {

	case *pg_query.Node_InsertStmt:
		return extractFromInsert(n.InsertStmt, table, shardKey)

	case *pg_query.Node_SelectStmt:
		return extractFromWhere(n.SelectStmt.WhereClause, table, shardKey)

	case *pg_query.Node_UpdateStmt:
		return extractFromWhere(n.UpdateStmt.WhereClause, table, shardKey)

	case *pg_query.Node_DeleteStmt:
		return extractFromWhere(n.DeleteStmt.WhereClause, table, shardKey)

	default:
		return nil, &RoutingError{
			Code:    ErrUnsupportedPredicate,
			Message: "unsupported statement type",
		}
	}
}

func extractFromInsert(stmt *pg_query.InsertStmt, table string, shardKey string) (*ExtractedPredicate, *RoutingError) {

	// INSERT ... SELECT is not supported in v1
	if stmt.SelectStmt == nil {
		return nil, &RoutingError{
			Code:    ErrUnsupportedPredicate,
			Message: "invalid insert statement",
		}
	}

	selectNode := stmt.SelectStmt.Node.(*pg_query.Node_SelectStmt)
	selectStmt := selectNode.SelectStmt

	// Only support INSERT ... VALUES
	if selectStmt.ValuesLists == nil {
		return nil, &RoutingError{
			Code:    ErrUnsupportedPredicate,
			Message: "insert-select not supported",
		}
	}

	// Find shard key column index
	colIndex := findShardKeyIndex(stmt.Cols, shardKey)
	if colIndex < 0 {
		return nil, &RoutingError{
			Code:    ErrShardKeyNotInQuery,
			Message: "shard key not present in insert columns",
		}
	}

	var values []any

	for _, row := range selectStmt.ValuesLists {

		list := row.Node.(*pg_query.Node_List)
		items := list.List.Items

		if colIndex >= len(items) {
			return nil, &RoutingError{
				Code:    ErrShardKeyNotInQuery,
				Message: "shard key value missing in values",
			}
		}

		val, ok := extractConst(items[colIndex])
		if !ok {
			return nil, &RoutingError{
				Code:    ErrUnsupportedPredicate,
				Message: "non-constant shard key in insert",
			}
		}

		values = append(values, val)
	}

	return &ExtractedPredicate{
		Table:  table,
		Column: shardKey,
		Type:   predicateTypeForCount(len(values)),
		Values: values,
	}, nil
}

func extractFromWhere(where *pg_query.Node, table string, shardKey string) (*ExtractedPredicate, *RoutingError) {

	if where == nil {
		return nil, &RoutingError{
			Code:    ErrShardKeyNotInQuery,
			Message: "missing where clause",
		}
	}

	values, err := walkWhere(where, shardKey)
	if err != nil {
		return nil, err
	}

	if len(values) == 0 {
		return nil, &RoutingError{
			Code:    ErrShardKeyNotInQuery,
			Message: "shard key not constrained",
		}
	}

	return &ExtractedPredicate{
		Table:  table,
		Column: shardKey,
		Type:   predicateTypeForCount(len(values)),
		Values: values,
	}, nil
}

func walkWhere(node *pg_query.Node, shardKey string) ([]any, *RoutingError) {

	switch n := node.Node.(type) {

	case *pg_query.Node_BoolExpr:
		if n.BoolExpr.Boolop != pg_query.BoolExprType_AND_EXPR {
			return nil, &RoutingError{
				Code:    ErrUnsupportedPredicate,
				Message: "OR predicates not supported",
			}
		}

		var result []any
		for _, arg := range n.BoolExpr.Args {
			vals, err := walkWhere(arg, shardKey)
			if err != nil {
				return nil, err
			}
			result = append(result, vals...)
		}
		return result, nil

	case *pg_query.Node_AExpr:
		return extractFromComparison(n.AExpr, shardKey)

	default:
		return nil, nil
	}
}

func extractFromComparison(expr *pg_query.A_Expr, shardKey string) ([]any, *RoutingError) {

	if expr.Kind != pg_query.A_Expr_Kind_AEXPR_OP {
		return nil, nil
	}

	col, ok := extractColumn(expr.Lexpr)
	if !ok || col != shardKey {
		return nil, nil
	}

	opNode := expr.Name[0].Node.(*pg_query.Node_String_)
	op := opNode.String_.Sval

	if op != "=" {
		return nil, &RoutingError{
			Code:    ErrUnsupportedPredicate,
			Message: "unsupported operator on shard key",
		}
	}

	val, ok := extractConst(expr.Rexpr)
	if !ok {
		return nil, &RoutingError{
			Code:    ErrUnsupportedPredicate,
			Message: "non-constant shard key comparison",
		}
	}

	return []any{val}, nil
}

func extractColumn(node *pg_query.Node) (string, bool) {
	col, ok := node.Node.(*pg_query.Node_ColumnRef)
	if !ok || len(col.ColumnRef.Fields) != 1 {
		return "", false
	}

	field := col.ColumnRef.Fields[0].Node.(*pg_query.Node_String_)
	return field.String_.Sval, true
}

func extractConst(node *pg_query.Node) (any, bool) {
	ac, ok := node.Node.(*pg_query.Node_AConst)
	if !ok {
		return nil, false
	}

	switch v := ac.AConst.Val.(type) {

	case *pg_query.A_Const_Ival:
		return v.Ival, true

	case *pg_query.A_Const_Fval:
		return v.Fval, true

	case *pg_query.A_Const_Boolval:
		return v.Boolval, true

	case *pg_query.A_Const_Sval:
		return v.Sval, true

	case *pg_query.A_Const_Bsval:
		return v.Bsval, true

	default:
		return nil, false
	}
}

func findShardKeyIndex(cols []*pg_query.Node, shardKey string) int {

	for i, c := range cols {
		res := c.Node.(*pg_query.Node_ResTarget)
		if res.ResTarget.Name == shardKey {
			return i
		}
	}
	return -1
}

func predicateTypeForCount(n int) PredicateType {
	if n <= 1 {
		return PredicateEquals
	}
	return PredicateIn
}
