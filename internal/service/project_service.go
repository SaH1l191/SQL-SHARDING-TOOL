package service

import (
	"context"
	"errors"
	"sqlsharder/internal/repository"
)
 
var (
	ErrInvalidInput    = errors.New("invalid input")
	ErrProjectNotFound = errors.New("project not found")
)
 
type CreateProjectRequest struct {
	Name        string `json:"name"`
	Description string `json:"description"`
}

type ProjectService struct {
	repo *repository.ProjectRepository
}

func NewProjectService(repo *repository.ProjectRepository) *ProjectService {
	return &ProjectService{repo: repo}
}

func (s *ProjectService) CreateProject(ctx context.Context, req *CreateProjectRequest) (*repository.Project, error) {
	if req.Name == "" {
		return nil, ErrInvalidInput
	}
	return s.repo.ProjectAdd(ctx, req.Name, req.Description)
}

func (s *ProjectService) GetProjects(ctx context.Context) ([]*repository.Project, error) {
	return s.repo.ProjectList(ctx)
}

func (s *ProjectService) DeleteProject(ctx context.Context, id string) error {
	err := s.repo.ProjectRemove(ctx, id)
	if err != nil {
		return ErrProjectNotFound
	}
	return nil
}

func (s *ProjectService) ActivateProject(ctx context.Context, id string) error {
	return s.repo.ProjectActive(ctx, id)
}

func (s *ProjectService) DeactivateProject(ctx context.Context, id string) error {
	return s.repo.ProjectDeactivate(ctx, id)
}

func (s *ProjectService) GetProjectStatus(ctx context.Context, id string) (string, error) {
	status, err := s.repo.FetchProjectStatus(ctx, id)
	if err != nil {
		return "", ErrProjectNotFound
	}
	return status, nil
}
