package connections

import (
	"fmt"
	"sqlsharder/internal/repository"
)

func buildDsn(s repository.ShardConnection) string {
	return fmt.Sprintf(
		"postgres://%s:%s@%s:%d/%s?sslmode=disable&statement_timeout=20000",
		s.Username,
		s.Password,
		s.Host,
		s.Port,
		s.DatabaseName,
	)
}