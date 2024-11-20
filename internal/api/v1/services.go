package v1

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/kndrad/wordcrack/internal/textproc"
)

type WordService struct {
	q      textproc.Querier
	logger *slog.Logger
}

func NewWordsService(queries textproc.Querier, logger *slog.Logger) *WordService {
	return &WordService{
		q:      queries,
		logger: logger,
	}
}

func (svc *WordService) GetAllWords(ctx context.Context, limit, offset int32) ([]textproc.AllWordsRow, error) {
	rows, err := svc.q.AllWords(ctx, textproc.AllWordsParams{
		Limit:  limit,
		Offset: offset,
	})
	if err != nil {
		return nil, fmt.Errorf("query all words, err: %w", err)
	}

	return rows, nil
}

func (svc *WordService) InsertWord(ctx context.Context, value string) (textproc.InsertWordRow, error) {
	if value == "" {
		panic("value cannot be empty")
	}
	row, err := svc.q.InsertWord(ctx, value)
	if err != nil {
		return row, fmt.Errorf("insert word: %w", err)
	}

	return row, nil
}
