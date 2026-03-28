package repository

import (
	"context"
	"database/sql"
	"time"
	"github.com/google/uuid"
)

type Project struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	Status      string    `json:"status"`
	ShardCount  int       `json:"shard_count"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// Project
// id	name	shard_count	status
// p1	ecommerce	4	active

type ProjectRepository struct {
	db *sql.DB
}

func NewProjectRepository(db *sql.DB) *ProjectRepository {
	return &ProjectRepository{db: db}
}

func (r *ProjectRepository) ProjectAdd(ctx context.Context, name string, description string) (*Project, error) {
	newID := uuid.New().String()

	project := &Project{
		ID:          newID,
		Name:        name,
		Description: description,
		Status:      "inactive",
		ShardCount:  0,
		CreatedAt:   time.Now(),
	}
	query := `INSERT INTO projects (id,name,description,status,shard_count,created_at)
			  VALUES ($1, $2, $3, $4, $5, $6)`
	_, err := r.db.ExecContext(ctx, query, project.ID, project.Name, project.Description, project.Status, project.ShardCount, project.CreatedAt)
	if err != nil {
		return nil, err
	}
	return project, nil
}

func (r *ProjectRepository) ProjectList(ctx context.Context) ([]*Project, error) {
	query := `SELECT id, name,description, status, shard_count, created_at, updated_at FROM projects ORDER BY created_at DESC`
	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var projects []*Project
	for rows.Next() {
		project := &Project{}
		err := rows.Scan(&project.ID, &project.Name, &project.Description, &project.Status, &project.ShardCount, &project.CreatedAt, &project.UpdatedAt)
		if err != nil {
			return nil, err
		}
		projects = append(projects, project)
	}
	return projects, nil
}

func (r *ProjectRepository) ProjectRemove(ctx context.Context, id string) error {
	projectID, err := uuid.Parse(id)
	if err != nil {
		return err
	}
	query := `DELETE FROM projects WHERE id = $1`
	result, err := r.db.ExecContext(ctx, query, projectID)
	if err != nil {
		return err
	}
	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}
	// no rows affected means the project doesn't exist
	if rows == 0 {
		return sql.ErrNoRows
	}
	return nil
}

func (r *ProjectRepository) ProjectActive(ctx context.Context, projectID string) error {
	query := `UPDATE projects SET status='active' WHERE id=$1`
	_, err := r.db.ExecContext(ctx, query, projectID)
	if err != nil {
		return err
	}
	return nil
}

func (r *ProjectRepository) ProjectDeactivate(ctx context.Context, projectID string) error {
	query := `UPDATE projects SET status='inactive' WHERE id=$1`
	_, err := r.db.ExecContext(ctx, query, projectID)
	if err != nil {
		return err
	}
	return nil
}

func (r *ProjectRepository) GetProjectById(ctx context.Context, id string) (Project, error) {
	query := `SELECT id, name, description, status, shard_count, created_at, updated_at FROM projects WHERE id=$1`
	var project Project
	err := r.db.QueryRowContext(ctx, query, id).Scan(&project.ID, &project.Name, &project.Description, &project.Status, &project.ShardCount, &project.CreatedAt, &project.UpdatedAt)
	if err != nil {
		return Project{}, err
	}
	return project, nil
}

func (r *ProjectRepository) GetActiveProjectId(ctx context.Context) (string, error) {
	query := `SELECT id from projects WHERE status='active'`
	row := r.db.QueryRowContext(ctx, query)
	var projectId string
	err := row.Scan(&projectId)
	if err != nil {
		return "", err
	}
	// defer row.Close() doesnt return cursor as single row is queried not multiple rows
	return projectId, nil
}

func (r *ProjectRepository) FetchProjectStatus(ctx context.Context, projectID string) (string, error) {
	query := `SELECT status FROM projects WHERE id=$1`
	var status string
	err := r.db.QueryRowContext(ctx, query, projectID).Scan(&status)
	if err != nil {
		return "", err
	}
	return status, nil
}
