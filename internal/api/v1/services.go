package v1

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/kndrad/wcrack/internal/textproc/database"
)

type WordService struct {
	q      database.Querier
	logger *slog.Logger
}

func NewWordService(q database.Querier, logger *slog.Logger) *WordService {
	return &WordService{
		q:      q,
		logger: logger,
	}
}

func (svc *WordService) GetAllWords(ctx context.Context, limit, offset int32) ([]database.ListWordsRow, error) {
	rows, err := svc.q.ListWords(ctx, database.ListWordsParams{
		Limit:  limit,
		Offset: offset,
	})
	if err != nil {
		return nil, fmt.Errorf("query all words, err: %w", err)
	}

	return rows, nil
}

func (svc *WordService) InsertWord(ctx context.Context, value string) (database.CreateWordRow, error) {
	if value == "" {
		panic("value cannot be empty")
	}
	row, err := svc.q.CreateWord(ctx, value)
	if err != nil {
		return row, fmt.Errorf("insert word: %w", err)
	}

	return row, nil
}
