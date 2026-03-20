package schema

import pg_query "github.com/pganalyze/pg_query_go/v5"

func parseDDLStatement(sql string) (*pg_query.ParseResult, error) {
	astTree, err := pg_query.Parse(sql)
	if err != nil {
		return nil, err
	}
	return astTree, nil
}
