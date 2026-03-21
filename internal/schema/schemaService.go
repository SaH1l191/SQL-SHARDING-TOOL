package schema

import (
	"context"
	"fmt"
	"sqlsharder/internal/repository"
	"sqlsharder/internal/shardkey"
	"sqlsharder/pkg/logger"
)

type SchemaService struct {
	columnRepo       *repository.ColumnRepository
	fkRepo           *repository.FKEdgesRepository
	inferenceService *shardkey.InferenceService
}

func NewSchemaService(colRepo *repository.ColumnRepository, fkRepo *repository.FKEdgesRepository, inferenceService *shardkey.InferenceService) *SchemaService {
	return &SchemaService{
		columnRepo:       colRepo,
		fkRepo:           fkRepo,
		inferenceService: inferenceService,
	}
}

// flow :=
//  1. -> convert input ddl
//  2. -> logical schema
//  3. -> convert to col/rows/table
//  4. -> save to DB (metadata/base schema to compare with )

func (s *SchemaService) ApplyDDL(ctx context.Context, projectId string, ddl string) error {
	// Step 1 — parse the new DDL into flat slices
	newCols, newFKs, err := ParseDDLToMetadata(projectId, ddl)
	if err != nil {
		logger.Logger.Error("SchemaService.ApplyDDL: parse failed", "error", err)
		return fmt.Errorf("ApplyDDL parse: %w", err)
	}

	// Step 2 — load existing metadata from DB
	existingCols, err := s.columnRepo.GetColumnsByProjectId(ctx, projectId)
	if err != nil {
		logger.Logger.Error("SchemaService.ApplyDDL: fetch existing columns failed", "error", err)
		return fmt.Errorf("ApplyDDL fetch columns: %w", err)
	}

	existingFKs, err := s.fkRepo.GetEdgesByProjectID(ctx, projectId)
	if err != nil {
		logger.Logger.Error("SchemaService.ApplyDDL: fetch existing fk_edges failed", "error", err)
		return fmt.Errorf("ApplyDDL fetch fk_edges: %w", err)
	}

	// Step 3 — merge: new columns overwrite matching existing ones,
	//           existing columns not mentioned in new DDL are preserved
	mergedCols := mergeColumns(existingCols, newCols)
	mergedFKs := mergeFKEdges(existingFKs, newFKs)

	logger.Logger.Info("SchemaService.ApplyDDL: merged metadata",
		"existing_cols", len(existingCols),
		"new_cols", len(newCols),
		"merged_cols", len(mergedCols),
		"existing_fks", len(existingFKs),
		"new_fks", len(newFKs),
		"merged_fks", len(mergedFKs),
	)

	// Step 4 — atomically replace DB records with merged result
	if err := s.columnRepo.ReplaceExistingColumns(ctx, projectId, mergedCols); err != nil {
		logger.Logger.Error("SchemaService.ApplyDDL: column replace failed", "error", err)
		return fmt.Errorf("ApplyDDL replace columns: %w", err)
	}

	if err := s.fkRepo.ReplaceFKEdgesForProject(ctx, projectId, mergedFKs); err != nil {
		logger.Logger.Error("SchemaService.ApplyDDL: fk_edges replace failed", "error", err)
		return fmt.Errorf("ApplyDDL replace fk_edges: %w", err)
	}
	
	//shard-key-election process
	if err := s.inferenceService.RunForProject(ctx,projectId); err !=nil {
		logger.Logger.Error("SchemaService.ApplyDDL: inference failed", "error", err)
		return fmt.Errorf("ApplyDDL inference: %w", err)
	}
	logger.Logger.Info("SchemaService.ApplyDDL: inference completed", "project_id", projectId)

	return nil
}

// colKey uniquely identifies a column within a project.
type colKey struct {
	tableName  string
	columnName string
}

// mergeColumns merges existing and new columns.
// New columns overwrite existing ones with the same (table, column) key.
// Existing columns not present in the new DDL are preserved.
func mergeColumns(existing []repository.Column, incoming []repository.Column) []repository.Column {
	// seed the map with existing columns
	merged := make(map[colKey]repository.Column, len(existing))
	for _, col := range existing {
		merged[colKey{col.TableName, col.ColumnName}] = col
	}

	// overwrite / add with incoming columns
	// CREATE TABLE re-submission → overwrites; ALTER ADD COLUMN → adds new key
	for _, col := range incoming {
		merged[colKey{col.TableName, col.ColumnName}] = col
	}

	result := make([]repository.Column, 0, len(merged))
	for _, col := range merged {
		result = append(result, col)
	}
	return result
}

// fkKey uniquely identifies a foreign key relationship.
type fkKey struct {
	childTable   string
	childColumn  string
	parentTable  string
	parentColumn string
}

// mergeFKEdges merges existing and new FK edges.
// New edges overwrite existing ones with the same key.
// Existing edges not present in the new DDL are preserved.
func mergeFKEdges(existing, incoming []repository.FkEdges) []repository.FkEdges {
	merged := make(map[fkKey]repository.FkEdges, len(existing))
	for _, fk := range existing {
		merged[fkKey{fk.ChildTable, fk.ChildColumn, fk.ParentTable, fk.ParentColumn}] = fk
	}

	for _, fk := range incoming {
		merged[fkKey{fk.ChildTable, fk.ChildColumn, fk.ParentTable, fk.ParentColumn}] = fk
	}

	result := make([]repository.FkEdges, 0, len(merged))
	for _, fk := range merged {
		result = append(result, fk)
	}
	return result
}

// why map instead of nesting iteratively on Table-> Col -> relation -> contrainsts
// tc -> tc.columns -> tc.columns.relations -> tc.columns.relations.contraints
// map approach -> simpler and more efficient for deduplication and updates

// check if it exists in existing, add or overwrite.
//  That's O(n²).
// The map approach is O(n) — each column is touched exactly once in
// each loop. For large schemas with hundreds of columns this matters,
// but more importantly the map approach is also simpler to read:
//  build the map from existing, overwrite with incoming, flatten back
// to slice. Three steps, no nested logic.
