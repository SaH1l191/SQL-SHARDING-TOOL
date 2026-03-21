package repository

import (
	"context"
	"database/sql"
)

// 1 project → N tables → many FK relationships
type FkEdges struct {
	ProjectId    string `json:"project_id"`
	ParentTable  string `json:"parent_table"`
	ParentColumn string `json:"parent_column"`
	ChildTable   string `json:"child_table"`
	ChildColumn  string `json:"child_column"`
}

type FKEdgesRepository struct {
	db *sql.DB
}

func NewFKEdgesRepository(db *sql.DB) *FKEdgesRepository {
	return &FKEdgesRepository{db: db}
}

func (e *FKEdgesRepository) GetEdgesByProjectID(
	ctx context.Context,
	projectID string,
) ([]FkEdges, error) {

	query := `
		SELECT project_id, parent_table, parent_column, child_table, child_column
		FROM fk_edges
		WHERE project_id = $1`

	rows, err := e.db.QueryContext(
		ctx,
		query,
		projectID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var result []FkEdges
	for rows.Next() {
		var temp FkEdges
		err = rows.Scan(
			&temp.ProjectId,
			&temp.ParentTable,
			&temp.ParentColumn,
			&temp.ChildTable,
			&temp.ChildColumn,
		)
		if err != nil {
			return nil, err
		}
		result = append(result, temp)
	}
	return result, rows.Err()
}

func (e *FKEdgesRepository) GetEdgesByChildTable(
	ctx context.Context, projectID string, tableName string,
) ([]FkEdges, error) {

	query := `
		SELECT 
			project_id, parent_table, parent_column, child_table, child_column
		FROM fk_edges
		WHERE project_id = $1 AND child_table = $2`

	rows, err := e.db.QueryContext(
		ctx,
		query,
		projectID,
		tableName,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var result []FkEdges
	for rows.Next() {
		var temp FkEdges
		err = rows.Scan(
			&temp.ProjectId,
			&temp.ParentTable,
			&temp.ParentColumn,
			&temp.ChildTable,
			&temp.ChildColumn,
		)
		if err != nil {
			return nil, err
		}
		result = append(result, temp)
	}
	return result, rows.Err()
}

func (e *FKEdgesRepository) GetEdgesByParentTable(
	ctx context.Context, projectID string, tableName string,
) ([]FkEdges, error) {

	query := `
		SELECT 
		project_id, parent_table, parent_column, child_table, child_column
		FROM fk_edges
		WHERE project_id = $1 AND parent_table = $2`

	rows, err := e.db.QueryContext(
		ctx,
		query,
		projectID,
		tableName,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var result []FkEdges
	for rows.Next() {
		var temp FkEdges
		err = rows.Scan(
			&temp.ProjectId,
			&temp.ParentTable,
			&temp.ParentColumn,
			&temp.ChildTable,
			&temp.ChildColumn,
		)
		if err != nil {
			return nil, err
		}
		result = append(result, temp)
	}
	return result, rows.Err()
}

//process : Deletes old FK relationships and inserts newly computed ones.

//because :LogicalSchema = truth & fk_edges = cached graph
// recompute → replace entire graph
//  avoids inconsistencies.
// failure → rollback → old graph restored

func (e *FKEdgesRepository) AddFKEdge(
	ctx context.Context, projectID string, parentTable string,
	parentColumn string, childTable string, childColumn string,
) error {

	query := `
		INSERT INTO fk_edges
		(project_id, parent_table, parent_column, child_table, child_column)
		VALUES ($1, $2, $3, $4, $5)`

	_, err := e.db.ExecContext(
		ctx,
		query,
		projectID,
		parentTable,
		parentColumn,
		childTable,
		childColumn,
	)
	if err != nil {
		return err
	}
	return nil
}

// ReplaceFKEdgesForProject replaces all FK edges for a project
func (e *FKEdgesRepository) ReplaceFKEdgesForProject(
	ctx context.Context,
	projectID string,
	edges []FkEdges,
) error {

	tx, err := e.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	deleteQuery := `DELETE FROM fk_edges WHERE project_id = $1`
	_, err = tx.ExecContext(ctx, deleteQuery, projectID)
	if err != nil {
		return err
	}

	insertQuery := `
		INSERT INTO fk_edges (project_id, parent_table, parent_column, child_table, child_column)
		VALUES ($1, $2, $3, $4, $5)`

	for _, edge := range edges {
		_, err = tx.ExecContext(
			ctx,
			insertQuery,
			edge.ProjectId,
			edge.ParentTable,
			edge.ParentColumn,
			edge.ChildTable,
			edge.ChildColumn,
		)
		if err != nil {
			return err
		}
	}
	return tx.Commit()
}
