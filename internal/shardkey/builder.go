package shardkey

import (
	"sqlsharder/internal/repository"
	"sqlsharder/internal/schema"
)

type RankedCandidate struct {
	Column  ColumnReference
	Score   int
	Reasons []string
}

//shardKeyInferenceResult : projectId,for each table :shardkey decision choices
// [projectId, decisions:[{tableName,columnName,score,reasons},{...}..]]

func BuildShardKeyPlan(logicalSchema *schema.LogicalSchema) ShardKeyInferenceResult {

	result := ShardKeyInferenceResult{
		ProjectId: logicalSchema.ProjectId,
	}

	candidateTableColumns := ExtractCandidates(logicalSchema) //Key = table name , Value = slice of candidate columns for that table
	fanout := ComputeFanout(logicalSchema, candidateTableColumns)

	// tableName : [[tname,colname]]
	for tableName, candidateCols := range candidateTableColumns {
		rank := RankTableCandidates(tableName, candidateCols, fanout, logicalSchema) //brings col,score,reasons
		decision := selectBestCandidate(tableName, rank)                             // brings table,column,score,reasons & appends to decsisons for shardkeyInferenceResult
		if decision == nil {
			continue
		}
		result.Decisions = append(result.Decisions, *decision)
	}
	return result
}

func ConvertToShardKeyRecords(shardKeyInferenceResult []ShardKeyDecision) []repository.ShardKeyRecord {
	//input : sharedkeyinferenceResult { projectId, decision: [tableName, column, score, reasons] }
	// output : [TableName ,ShardKeyColumn,IsManual]

	records := make([]repository.ShardKeyRecord, 0, len(shardKeyInferenceResult))
	for _, decision := range shardKeyInferenceResult {
		records = append(records, repository.ShardKeyRecord{
			TableName:      decision.Table,
			ShardKeyColumn: decision.Column.ColumnName,
			IsManual:       false,
		})
	}
	return records
}

func selectBestCandidate(table string, ranked []RankedCandidate) *ShardKeyDecision {
	if len(ranked) == 0 {
		return nil
	}
	best := ranked[0]
	return &ShardKeyDecision{
		Table:   table,
		Column:  best.Column,
		Score:   best.Score,
		Reasons: best.Reasons,
	}
}
