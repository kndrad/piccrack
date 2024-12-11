package v1

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/kndrad/wcrack/internal/database"
)

type Service interface {
	ListWords(ctx context.Context, limit, offset int32) ([]database.ListWordsRow, error)
	CreateWord(ctx context.Context, value string) (database.CreateWordRow, error)
	ListWordBatches(ctx context.Context, limit, offset int32) ([]database.ListWordBatchesRow, error)
	CreateWordsBatch(ctx context.Context, name string, values []string) (database.CreateWordsBatchRow, error)
	ListWordsByBatchName(ctx context.Context, name string) ([]database.ListWordsByBatchNameRow, error)
	CreateSentencesBatch(ctx context.Context, name string, values []string) (database.CreateSentencesBatchRow, error)
}

type service struct {
	q      database.Querier
	logger *slog.Logger
}

var _ Service = (*service)(nil)

func NewService(q database.Querier, l *slog.Logger) Service {
	return &service{
		q:      q,
		logger: l,
	}
}

func (svc *service) ListWords(ctx context.Context, limit, offset int32) ([]database.ListWordsRow, error) {
	rows, err := svc.q.ListWords(ctx, database.ListWordsParams{
		Limit:  limit,
		Offset: offset,
	})
	if err != nil {
		return nil, fmt.Errorf("query all words, err: %w", err)
	}

	return rows, nil
}

func (svc *service) CreateWord(ctx context.Context, value string) (database.CreateWordRow, error) {
	if value == "" {
		panic("value cannot be empty")
	}
	row, err := svc.q.CreateWord(ctx, value)
	if err != nil {
		return row, fmt.Errorf("insert word: %w", err)
	}

	return row, nil
}

func (svc *service) ListWordBatches(ctx context.Context, limit, offset int32) ([]database.ListWordBatchesRow, error) {
	rows, err := svc.q.ListWordBatches(ctx, database.ListWordBatchesParams{
		Limit:  limit,
		Offset: offset,
	})
	if err != nil {
		return nil, fmt.Errorf("list word batches: %w", err)
	}

	return rows, nil
}

func (svc *service) CreateWordsBatch(ctx context.Context, name string, values []string) (database.CreateWordsBatchRow, error) {
	row, err := svc.q.CreateWordsBatch(ctx, database.CreateWordsBatchParams{
		Name:    name,
		Column2: values,
	})
	if err != nil {
		return row, fmt.Errorf("create word batch: %w", err)
	}

	return row, nil
}

func (svc *service) ListWordsByBatchName(ctx context.Context, name string) ([]database.ListWordsByBatchNameRow, error) {
	rows, err := svc.q.ListWordsByBatchName(ctx, name)
	if err != nil {
		return rows, fmt.Errorf("create word batch: %w", err)
	}

	return rows, nil
}

func (svc *service) CreateSentencesBatch(ctx context.Context, name string, values []string) (database.CreateSentencesBatchRow, error) {
	row, err := svc.q.CreateSentencesBatch(ctx, database.CreateSentencesBatchParams{
		Name:    name,
		Column2: values,
	})
	if err != nil {
		return row, fmt.Errorf("create word batch: %w", err)
	}

	return row, nil
}
