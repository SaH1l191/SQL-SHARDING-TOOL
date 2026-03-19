package repository

import (
	"context"
	"database/sql"
	"time"

	"github.com/google/uuid"
)

// string → “There is always a value (even if empty)”
// *string → “There might be a value… or nothing at all”
type ProjectSchema struct {
	ID          string  `json:"id"`
	ProjectID   string  `json:"project_id"`
	Version     int     `json:"version"`
	State       string  `json:"state"`
	DDL_SQL     string  `json:"ddl_sql"`
	ErrMsg      *string `json:"error_message"`
	CreatedAt   string  `json:"created_at"`
	CommittedAt *string `json:"committed_at"`
	AppliedAt   *string `json:"applied_at"`
}
type ProjectSchemaRepository struct {
	db *sql.DB
}

func NewProjectSchemaRepository(db *sql.DB) *ProjectSchemaRepository {
	return &ProjectSchemaRepository{db: db}
}

func findMaxVer(versions []int) int {
	if len(versions) == 0 {
		return 0
	}
	maxVer := versions[0]
	for _, ver := range versions {
		if ver > maxVer {
			maxVer = ver
		}
	}
	return maxVer
}

// Fetches all schema versions for a project.
func (p *ProjectSchemaRepository) GetProjectSchemaVersions(ctx context.Context, projectID string) ([]int, error) {
	query := `SELECT version FROM project_schemas WHERE project_id = $1 ORDER BY version ASC`
	rows, err := p.db.QueryContext(ctx, query, projectID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var versions []int
	for rows.Next() {
		var version int
		if err := rows.Scan(&version); err != nil {
			return nil, err
		}
		versions = append(versions, version)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return versions, nil
}

// Creates a new schema migration in draft state with next version.
func (p *ProjectSchemaRepository) CreateProjectSchemaDraft(
	ctx context.Context, projectID string, ddlSQL string) (*ProjectSchema, error) {
	versions, err := p.GetProjectSchemaVersions(ctx, projectID)
	if err != nil {
		return nil, err
	}
	nextVersion := findMaxVer(versions) + 1
	schemaID := uuid.New().String()
	now := time.Now()

	query := `INSERT INTO project_schemas (id,project_id,version,state,ddl_sql,created_at) VALUES ($1,$2,$3,$4,$5,$6)`

	_, err = p.db.ExecContext(ctx, query, schemaID, projectID, nextVersion, "draft", ddlSQL, now)
	if err != nil {
		return nil, err
	}
	return &ProjectSchema{
		ID:        schemaID,
		ProjectID: projectID,
		Version:   nextVersion,
		State:     "draft",
		DDL_SQL:   ddlSQL,
		CreatedAt: now.String(),
	}, nil
}

// update existing draft ddl (NO new row, NO version change)
// Updates SQL of an existing draft (no new version created).
//
//	Example: fix typo in draft query before committing.
func (p *ProjectSchemaRepository) ProjectSchemaUpdateDraft(ctx context.Context, schemaID string, ddlSQL string) error {

	query := `
		UPDATE project_schemas
		SET ddl_sql = $1
		WHERE id = $2 AND state = 'draft'
	`

	result, err := p.db.ExecContext(
		ctx,
		query,
		ddlSQL,
		schemaID,
	)
	if err != nil {
		return err
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return sql.ErrNoRows
	}
	return nil
}

// Deletes a draft schema (only if state is draft).
func (p *ProjectSchemaRepository) ProjectSchemaDeleteDraft(ctx context.Context, schemaID string) error {

	query := `
		DELETE FROM project_schemas
		WHERE id = $1 AND state = 'draft'`

	result, err := p.db.ExecContext(ctx, query, schemaID)
	if err != nil {
		return err
	}
	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return sql.ErrNoRows
	}
	return nil
}

//State Transitions
// draft → pending
// pending → applied
// applied → rolled_back

// Moves schema from draft → pending and sets commit time.
func (p *ProjectSchemaRepository) ProjectSchemaCommitDraft(ctx context.Context, schemaID string) error {
	query := `UPDATE project_schemas SET state = 'pending', committed_at = $1 WHERE id = $2`
	result, err := p.db.ExecContext(
		ctx,
		query,
		time.Now(),
		schemaID,
	)
	if err != nil {
		return err
	}
	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return sql.ErrNoRows
	}

	return nil
}

// func to modify pending -> applying
// Marks schema as applying.
//
//	Example: worker picks job → state becomes applying.
func (p *ProjectSchemaRepository) ProjectSchemaSetApplying(ctx context.Context, schemaID string) error {

	query := `
		UPDATE project_schemas
		SET state = 'applying'
		WHERE id = $1
	`

	_, err := p.db.ExecContext(ctx, query, schemaID)
	return err
}

// Updates state + error (like applied/failed).
// Example: shard fails → set state failed + error msg.
func (p *ProjectSchemaRepository) ProjectSchemaUpdateSchemaState(
	ctx context.Context,
	schemaID string,
	state string,
	errorMessage *string,
) error {

	query := `
		UPDATE project_schemas
		SET state = $1, error_message = $2
		WHERE id = $3`

	_, err := p.db.ExecContext(
		ctx,
		query,
		state,
		errorMessage,
		schemaID,
	)
	if err != nil {
		return err
	}
	return nil
}

// Returns current state (draft/applied/failed).
func (p *ProjectSchemaRepository) ProjectSchemaGetState(ctx context.Context, schemaID string) (string, error) {

	query := `
		SELECT state
		FROM project_schemas
		WHERE id = $1`

	row := p.db.QueryRowContext(ctx, query, schemaID)
	var state string
	if err := row.Scan(&state); err != nil {
		return "", err
	}
	return state, nil
}

// Retrieval
// Fetches latest version schema for project.
func (p *ProjectSchemaRepository) ProjectSchemaGetLatest(ctx context.Context, projectID string) (*ProjectSchema, error) {
	versions, err := p.GetProjectSchemaVersions(ctx, projectID)
	if err != nil {
		return nil, err
	}
	maxVer := findMaxVer(versions)
	query := `
		SELECT 
			id, project_id, version, state, ddl_sql,
			error_message, created_at, committed_at, applied_at
		FROM project_schemas
		WHERE project_id = $1 AND version = $2
	`

	row := p.db.QueryRowContext(
		ctx,
		query,
		projectID,
		maxVer,
	)
	var latestSchema ProjectSchema
	err = row.Scan(
		&latestSchema.ID,
		&latestSchema.ProjectID,
		&latestSchema.Version,
		&latestSchema.State,
		&latestSchema.DDL_SQL,
		&latestSchema.ErrMsg,
		&latestSchema.CreatedAt,
		&latestSchema.CommittedAt,
		&latestSchema.AppliedAt,
	)
	if err != nil {
		return nil, err
	}
	return &latestSchema, nil
}

// Returns all non-draft schemas ordered by version.
//
//	Example: shows migration history [v1,v2,v3].
func (p *ProjectSchemaRepository) ProjectSchemaFetchHistory(ctx context.Context, projectID string) ([]ProjectSchema, error) {

	query := `
		SELECT 
			id, project_id, version, state, ddl_sql,
			error_message, created_at, committed_at, applied_at
		FROM project_schemas
		WHERE project_id = $1
		AND state != 'draft'
		ORDER BY version ASC`
	rows, err := p.db.QueryContext(
		ctx,
		query,
		projectID,
	)

	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var history []ProjectSchema

	for rows.Next() {
		var schema ProjectSchema

		err := rows.Scan(
			&schema.ID,
			&schema.ProjectID,
			&schema.Version,
			&schema.State,
			&schema.DDL_SQL,
			&schema.ErrMsg,
			&schema.CreatedAt,
			&schema.CommittedAt,
			&schema.AppliedAt,
		)
		if err != nil {
			return nil, err
		}

		history = append(history, schema)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return history, nil
}

// Returns a specific schema by its ID.
func (p *ProjectSchemaRepository) ProjectSchemaGetBySchemaID(ctx context.Context, schemaID string) (*ProjectSchema, error) {
	query := `
		SELECT 
			id, project_id, version, state, ddl_sql,
			error_message, created_at, committed_at, applied_at
		FROM project_schemas
		WHERE id = $1`

	row := p.db.QueryRowContext(
		ctx,
		query,
		schemaID,
	)
	var schema ProjectSchema
	err := row.Scan(
		&schema.ID,
		&schema.ProjectID,
		&schema.Version,
		&schema.State,
		&schema.DDL_SQL,
		&schema.ErrMsg,
		&schema.CreatedAt,
		&schema.CommittedAt,
		&schema.AppliedAt,
	)
	if err != nil {
		return nil, err
	}
	return &schema, nil
}

// ProjectSchemaGetApplied retrieves the most recently applied schema version for a project
func (p *ProjectSchemaRepository) ProjectSchemaGetApplied(ctx context.Context, projectID string) (*ProjectSchema, error) {

	query := `
	SELECT 
		id, project_id, version, state, ddl_sql,
		error_message, created_at, committed_at, applied_at
	FROM project_schemas
	WHERE project_id = $1 AND state = 'applied'
	ORDER BY version DESC
	LIMIT 1`

	row := p.db.QueryRowContext(
		ctx,
		query,
		projectID,
	)

	var schema ProjectSchema

	err := row.Scan(
		&schema.ID,
		&schema.ProjectID,
		&schema.Version,
		&schema.State,
		&schema.DDL_SQL,
		&schema.ErrMsg,
		&schema.CreatedAt,
		&schema.CommittedAt,
		&schema.AppliedAt,
	)
	if err != nil {
		return nil, err
	}

	return &schema, nil
}

// Gets oldest pending schema (FIFO execution).
func (p *ProjectSchemaRepository) ProjectSchemaGetPending(
	ctx context.Context,
	projectID string,
) (*ProjectSchema, error) {

	query := `
		SELECT 
			id, project_id, version, state, ddl_sql,
			error_message, created_at, committed_at, applied_at
		FROM project_schemas
		WHERE project_id = $1 AND state = 'pending'
		ORDER BY version ASC
		LIMIT 1`

	row := p.db.QueryRowContext(ctx, query, projectID)
	var schema ProjectSchema
	if err := row.Scan(
		&schema.ID,
		&schema.ProjectID,
		&schema.Version,
		&schema.State,
		&schema.DDL_SQL,
		&schema.ErrMsg,
		&schema.CreatedAt,
		&schema.CommittedAt,
		&schema.AppliedAt,
	); err != nil {
		return nil, err
	}
	return &schema, nil
}

// helper to decide correct version fo schema -> can remove
func (p *ProjectSchemaRepository) fetchProjectSchemaVersions(ctx context.Context, projectID string) ([]int, error) {

	query := `
		SELECT version FROM project_schemas
		WHERE project_id = $1 
	`

	rows, err := p.db.QueryContext(
		ctx,
		query,
		projectID,
	)

	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var versions []int
	var version int

	for rows.Next() {
		err := rows.Scan(&version)
		if err != nil {
			return nil, err
		}
		versions = append(versions, version)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return versions, nil

}
