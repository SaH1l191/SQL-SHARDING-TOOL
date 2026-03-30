package schema

import (
	"context"
	"sqlsharder/internal/repository"
)

type SchemaService struct {
	columnRepo *repository.ColumnRepository
	fkRepo     *repository.FKEdgesRepository
}

func NewSchemaService(columnRepo *repository.ColumnRepository, fkRepo *repository.FKEdgesRepository) *SchemaService {
	return &SchemaService{
		columnRepo: columnRepo,
		fkRepo:     fkRepo,
	}
}

func (s *SchemaService) ApplyDDL(ctx context.Context, projectId string, ddl string) error {

	//step1 - incoming ddl to Delta Logical Schema
	//pending
	deltaSchema, err := BuildLogicalSchemaFromDDL(ddl)
	if err != nil {
		return err
	}
	//step2 - bring existing data from db & built logical schema from it
	columns, err := s.columnRepo.GetColumnsByProjectId(ctx, projectId)
	if err != nil {
		return err
	}
	fkEdges, err := s.fkRepo.GetEdgesByProjectID(ctx, projectId)
	if err != nil {
		return err
	}
	baseSchema, err := BuildLogicalSchemaFromDb(projectId, columns, fkEdges)
	if err != nil {
		return err
	}
	mergedSchema, err := MergeLogicalSchema(baseSchema, deltaSchema)
	if err != nil {
		return err
	}
	newColumns, newFKEdges, err := FlattenLogicalSchema(mergedSchema)
	if err != nil {
		return err
	}
	// step  — replace new table data in db atomically thread secure way
	if err := s.columnRepo.ReplaceExistingColumns(ctx, projectId, newColumns); err != nil {
		return err
	}
	if err := s.fkRepo.ReplaceFKEdgesForProject(ctx, projectId, newFKEdges); err != nil {
		return err
	}
	return nil
}
