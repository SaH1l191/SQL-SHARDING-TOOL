package shardkey

import (
	"context"
	"sqlsharder/internal/repository"
	"sqlsharder/pkg/logger"
)

type InferenceService struct {
	colRepo      *repository.ColumnRepository
	fkRepo       *repository.FKEdgesRepository
	shardKeyRepo *repository.ShardKeysRepository
}

func NewInferenceService(colRepo *repository.ColumnRepository, fkRepo *repository.FKEdgesRepository, shardKeyRepo *repository.ShardKeysRepository) *InferenceService {
	return &InferenceService{
		colRepo:      colRepo,
		fkRepo:       fkRepo,
		shardKeyRepo: shardKeyRepo,
	}
}

// takes projectId->finds columns and foreign keys->extracts all cols->finds potential candidate->
// computes fanout->ranks candidates->saves shardkeyForProject
func (s *InferenceService) RunForProject(ctx context.Context, projectId string) error {
	cols, err := s.colRepo.GetColumnsByProjectId(ctx, projectId)
	if err != nil {
		return err
	}
	fks, err := s.fkRepo.GetEdgesByProjectID(ctx, projectId)
	if err != nil {
		return err
	}

	// Stage 1 : potential shard key candidates
	//candidates Type => [tableName]:[[tName,Colname],[tName,Colname]...]
	candidates := ExtractCandidates(cols)

	// Stage 2 : compute fanout for each candidate
	// fk Type => {ProjectId  ,ParentTable,ParentColumn ,ChildTable,ChildColumn }
	//fanout Type => map[{TbName,ColName}]:[IncomingFkCount,ReferencingTableCount]
	fanout := ComputeFanout(fks)

	// Stage 3 : rank candidates
	decisions := RankCandidates(candidates, fanout, cols, fks)

	for _, d := range decisions {
		logger.Logger.Info("shard key decided",
			"table", d.Table,
			"column", d.Column.ColumnName,
			"score", d.Score,
			"reasons", d.Reasons,
		)
	}

	// Save
	records := make([]repository.ShardKeyRecord, 0, len(decisions))
	for _, d := range decisions {
		records = append(records, repository.ShardKeyRecord{
			TableName:      d.Table,
			ShardKeyColumn: d.Column.ColumnName,
			IsManual:       false,
		})
	}
	return s.shardKeyRepo.ReplaceShardKeysForProject(ctx, projectId, records)
}
