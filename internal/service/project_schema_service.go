package service

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"sqlsharder/internal/repository"
)

var (
	ErrSchemaNotFound = errors.New("schema not found")
	ErrInvalidState   = errors.New("invalid schema state")
)

type CreateSchemaRequest struct {
	ProjectID string `json:"project_id" binding:"required"`
	DDL_SQL   string `json:"ddl_sql"    binding:"required"`
}

type UpdateSchemaRequest struct {
	SchemaID string `json:"schema_id" binding:"required"`
	DDL_SQL  string `json:"ddl_sql"   binding:"required"`
}

type ProjectSchemaService struct {
	repo *repository.ProjectSchemaRepository
}

func NewProjectSchemaService(repo *repository.ProjectSchemaRepository) *ProjectSchemaService {
	return &ProjectSchemaService{repo: repo}
}

// CreateSchema creates a new draft schema for a project.
func (s *ProjectSchemaService) CreateSchema(ctx context.Context, req *CreateSchemaRequest) (*repository.ProjectSchema, error) {
	if req.ProjectID == "" || req.DDL_SQL == "" {
		return nil, ErrInvalidInput
	}
	schema, err := s.repo.CreateProjectSchemaDraft(ctx, req.ProjectID, req.DDL_SQL)
	if err != nil {
		return nil, fmt.Errorf("ProjectSchemaService.CreateSchema: %w", err)
	}
	return schema, nil
}

// GetSchemasByProject returns all schemas for a project (history + current draft if any).
func (s *ProjectSchemaService) GetSchemasByProject(ctx context.Context, projectID string) ([]repository.ProjectSchema, error) {
	if projectID == "" {
		return nil, ErrInvalidInput
	}

	history, err := s.repo.ProjectSchemaFetchHistory(ctx, projectID)
	if err != nil {
		return nil, fmt.Errorf("ProjectSchemaService.GetSchemasByProject: %w", err)
	}

	// Prepend the current draft if one exists
	latest, err := s.repo.ProjectSchemaGetLatest(ctx, projectID)
	if err == nil && latest.State == "draft" {
		all := make([]repository.ProjectSchema, 0, len(history)+1)
		all = append(all, *latest)
		all = append(all, history...)
		return all, nil
	}

	return history, nil
}

// DeleteSchema deletes a schema. Only drafts can be deleted.
func (s *ProjectSchemaService) DeleteSchema(ctx context.Context, id string) error {
	if id == "" {
		return ErrInvalidInput
	}

	state, err := s.repo.ProjectSchemaGetState(ctx, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return ErrSchemaNotFound
		}
		return fmt.Errorf("ProjectSchemaService.DeleteSchema: %w", err)
	}

	if state != "draft" {
		return ErrInvalidState
	}

	return s.repo.ProjectSchemaDeleteDraft(ctx, id)
}

// CommitSchema moves a draft schema to pending state, ready for execution.
// Named CommitSchema rather than ActivateSchema because "activate" is misleading
// for a schema — it doesn't run DDL, it just marks it ready to be applied.
func (s *ProjectSchemaService) CommitSchema(ctx context.Context, id string) error {
	if id == "" {
		return ErrInvalidInput
	}

	state, err := s.repo.ProjectSchemaGetState(ctx, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return ErrSchemaNotFound
		}
		return fmt.Errorf("ProjectSchemaService.CommitSchema: %w", err)
	}

	if state != "draft" {
		return ErrInvalidState
	}

	return s.repo.ProjectSchemaCommitDraft(ctx, id)
}

// UpdateSchema replaces the DDL of a draft schema. Cannot update committed schemas.
func (s *ProjectSchemaService) UpdateSchema(ctx context.Context, req *UpdateSchemaRequest) (*repository.ProjectSchema, error) {
	if req.SchemaID == "" || req.DDL_SQL == "" {
		return nil, ErrInvalidInput
	}

	state, err := s.repo.ProjectSchemaGetState(ctx, req.SchemaID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrSchemaNotFound
		}
		return nil, fmt.Errorf("ProjectSchemaService.UpdateSchema: %w", err)
	}

	if state != "draft" {
		return nil, ErrInvalidState
	}

	if err := s.repo.ProjectSchemaUpdateDraft(ctx, req.SchemaID, req.DDL_SQL); err != nil {
		return nil, fmt.Errorf("ProjectSchemaService.UpdateSchema: %w", err)
	}

	// FIX: repo method is ProjectSchemaFetchBySchemaID, not ProjectSchemaGetBySchemaID
	return s.repo.ProjectSchemaGetBySchemaID(ctx, req.SchemaID)
}

// GetSchemaByID returns a single schema by ID.
func (s *ProjectSchemaService) GetSchemaByID(ctx context.Context, id string) (*repository.ProjectSchema, error) {
	if id == "" {
		return nil, ErrInvalidInput
	}

	// FIX: repo method is ProjectSchemaFetchBySchemaID, not ProjectSchemaGetBySchemaID
	schema, err := s.repo.ProjectSchemaGetBySchemaID(ctx, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrSchemaNotFound
		}
		return nil, fmt.Errorf("ProjectSchemaService.GetSchemaByID: %w", err)
	}
	return schema, nil
}

// GetSchemaStatus returns the current state string of a schema.
func (s *ProjectSchemaService) GetSchemaStatus(ctx context.Context, id string) (string, error) {
	if id == "" {
		return "", ErrInvalidInput
	}

	status, err := s.repo.ProjectSchemaGetState(ctx, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return "", ErrSchemaNotFound
		}
		return "", fmt.Errorf("ProjectSchemaService.GetSchemaStatus: %w", err)
	}
	return status, nil
}

// GetLatestSchema returns the highest-version schema for a project.
func (s *ProjectSchemaService) GetLatestSchema(ctx context.Context, projectID string) (*repository.ProjectSchema, error) {
	if projectID == "" {
		return nil, ErrInvalidInput
	}

	schema, err := s.repo.ProjectSchemaGetLatest(ctx, projectID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrSchemaNotFound
		}
		return nil, fmt.Errorf("ProjectSchemaService.GetLatestSchema: %w", err)
	}
	return schema, nil
}

// GetAppliedSchema returns the most recently applied schema for a project.
func (s *ProjectSchemaService) GetAppliedSchema(ctx context.Context, projectID string) (*repository.ProjectSchema, error) {
	if projectID == "" {
		return nil, ErrInvalidInput
	}

	schema, err := s.repo.ProjectSchemaGetApplied(ctx, projectID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrSchemaNotFound
		}
		return nil, fmt.Errorf("ProjectSchemaService.GetAppliedSchema: %w", err)
	}
	return schema, nil
}

// GetProjectSchemaVersions returns all version numbers for a project.
func (s *ProjectSchemaService) GetProjectSchemaVersions(ctx context.Context, projectID string) ([]int, error) {
	if projectID == "" {
		return nil, ErrInvalidInput
	}

	versions, err := s.repo.GetProjectSchemaVersions(ctx, projectID)
	if err != nil {
		return nil, fmt.Errorf("ProjectSchemaService.GetProjectSchemaVersions: %w", err)
	}
	return versions, nil
}

// GetPendingSchema returns the oldest pending schema for a project (FIFO).
func (s *ProjectSchemaService) GetPendingSchema(ctx context.Context, projectID string) (*repository.ProjectSchema, error) {
	if projectID == "" {
		return nil, ErrInvalidInput
	}

	schema, err := s.repo.ProjectSchemaGetPending(ctx, projectID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrSchemaNotFound
		}
		return nil, fmt.Errorf("ProjectSchemaService.GetPendingSchema: %w", err)
	}
	return schema, nil
}

// SetApplying marks a schema as being applied.
func (s *ProjectSchemaService) SetApplying(ctx context.Context, schemaID string) error {
	if schemaID == "" {
		return ErrInvalidInput
	}

	err := s.repo.ProjectSchemaSetApplying(ctx, schemaID)
	if err != nil {
		return fmt.Errorf("ProjectSchemaService.SetApplying: %w", err)
	}
	return nil
}

// UpdateSchemaState updates the state and optionally error message of a schema.
func (s *ProjectSchemaService) UpdateSchemaState(ctx context.Context, schemaID string, state string, errorMessage *string) error {
	if schemaID == "" || state == "" {
		return ErrInvalidInput
	}

	err := s.repo.ProjectSchemaUpdateSchemaState(ctx, schemaID, state, errorMessage)
	if err != nil {
		return fmt.Errorf("ProjectSchemaService.UpdateSchemaState: %w", err)
	}
	return nil
}
