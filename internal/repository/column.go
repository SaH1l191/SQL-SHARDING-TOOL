package repository

import (
	"context"
	"database/sql"
	"time"
)

type Column struct {
	ProjectID  string    `json:"project_id"`
	TableName  string    `json:"table_name"`
	ColumnName string    `json:"column_name"`
	DataType   string    `json:"data_type"`
	IsNullable bool      `json:"is_nullable"`
	IsPK       bool      `json:"is_pk"`
	CreatedAt  time.Time `json:"created_at"`
}

type ColumnRepository struct {
	db *sql.DB
}

func NewColumnRepository(db *sql.DB) *ColumnRepository {
	return &ColumnRepository{
		db: db,
	}
}

func (c *ColumnRepository) GetColumnsByProjectId(ctx context.Context, projectId string) ([]Column, error) {
	query := `SELECT project_id,table_name,column_name,data_type,nullable,is_primary_key
	FROM columns WHERE project_id = $1`

	rows, err := c.db.QueryContext(ctx, query, projectId)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var columns []Column
	for rows.Next() {
		var column Column
		err := rows.Scan(&column.ProjectID, &column.TableName, &column.ColumnName, &column.DataType, &column.IsNullable, &column.IsPK)
		if err != nil {
			return nil, err
		}
		columns = append(columns, column)
	}
	return columns, nil
}

func (c *ColumnRepository) AddProjectColumn(
	ctx context.Context, projectId string, tableName string, columnName string, dataType string,
	nullable bool, isPK bool,
) error {
	query := `INSERT INTO columns (project_id, table_name, column_name, data_type, nullable, is_primary_key)
	VALUES ($1, $2, $3, $4, $5, $6)`

	_, err := c.db.ExecContext(ctx, query, projectId, tableName, columnName, dataType, nullable, isPK)
	if err != nil {
		return err
	}
	return nil
}

// LogicalSchema = source of truth
// columns table = cache of truth

// So instead of:

// add column email
// update nullable
// remove something

// So :
// wipe → rebuild
//  This avoids:
// drift
// partial updates
// complex diff logic

// tx.Begin()
// DELETE
// INSERT all
// tx.Commit()

// DELETE old columns
// INSERT new columns (from parsed schema)
func (c *ColumnRepository) ReplaceExistingColumns(
	ctx context.Context, projectId string, cols []Column,
) error {
	tx, err := c.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	// DELETE
	query := "DELETE FROM columns WHERE project_id = $1"
	_, err = tx.ExecContext(ctx, query, projectId)
	if err != nil {
		return err
	}

	// INSERT all
	for _, col := range cols {
		query := `INSERT INTO columns (project_id, table_name, column_name, data_type, nullable, is_primary_key) VALUES ($1, $2, $3, $4, $5, $6)`
		_, err = tx.ExecContext(ctx, query, projectId, col.TableName, col.ColumnName, col.DataType, col.IsNullable, col.IsPK)
		if err != nil {
			return err
		}
	}
	return tx.Commit()
}
