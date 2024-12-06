package v1

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/kndrad/wcrack/internal/textproc/database"
)

type WordService interface {
	ListWords(ctx context.Context, limit, offset int32) ([]database.ListWordsRow, error)
	CreateWord(ctx context.Context, value string) (database.CreateWordRow, error)
	ListWordBatches(ctx context.Context, limit, offset int32) ([]database.ListWordBatchesRow, error)
}

type wordService struct {
	q      database.Querier
	logger *slog.Logger
}

var _ WordService = (*wordService)(nil)

func NewWordService(q database.Querier, l *slog.Logger) WordService {
	return &wordService{
		q:      q,
		logger: l,
	}
}

func (svc *wordService) ListWords(ctx context.Context, limit, offset int32) ([]database.ListWordsRow, error) {
	rows, err := svc.q.ListWords(ctx, database.ListWordsParams{
		Limit:  limit,
		Offset: offset,
	})
	if err != nil {
		return nil, fmt.Errorf("query all words, err: %w", err)
	}

	return rows, nil
}

func (svc *wordService) CreateWord(ctx context.Context, value string) (database.CreateWordRow, error) {
	if value == "" {
		panic("value cannot be empty")
	}
	row, err := svc.q.CreateWord(ctx, value)
	if err != nil {
		return row, fmt.Errorf("insert word: %w", err)
	}

	return row, nil
}

func (svc *wordService) ListWordBatches(ctx context.Context, limit, offset int32) ([]database.ListWordBatchesRow, error) {
	rows, err := svc.q.ListWordBatches(ctx, database.ListWordBatchesParams{
		Limit:  limit,
		Offset: offset,
	})
	if err != nil {
		return nil, fmt.Errorf("list word batches: %w", err)
	}

	return rows, nil
}
