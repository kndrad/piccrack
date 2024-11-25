package textproc

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgtype"
)

const allWords = `-- name: AllWords :many
SELECT id, value, created_at
FROM words
WHERE deleted_at IS NULL
ORDER BY value ASC
LIMIT $1 OFFSET $2
`

type AllWordsParams struct {
	Limit  int32 `json:"limit"`
	Offset int32 `json:"offset"`
}

type AllWordsRow struct {
	ID        int64              `json:"id"`
	Value     string             `json:"value"`
	CreatedAt pgtype.Timestamptz `json:"created_at"`
}

func (q *Queries) AllWords(ctx context.Context, arg AllWordsParams) ([]AllWordsRow, error) {
	rows, err := q.db.Query(ctx, allWords, arg.Limit, arg.Offset)
	if err != nil {
		return nil, fmt.Errorf("query db: %w", err)
	}
	defer rows.Close()
	var items []AllWordsRow
	for rows.Next() {
		var i AllWordsRow
		if err := rows.Scan(&i.ID, &i.Value, &i.CreatedAt); err != nil {
			return nil, fmt.Errorf("scan db: %w", err)
		}
		items = append(items, i)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows: %w", err)
	}

	return items, nil
}

const getWordsFrequencies = `-- name: GetWordsFrequencies :many
SELECT words.value, count(*) AS frequency
FROM words
WHERE deleted_at IS NULL
GROUP BY words.value
ORDER BY frequency ASC
LIMIT $1 OFFSET $2
`

type GetWordsFrequenciesParams struct {
	Limit  int32 `json:"limit"`
	Offset int32 `json:"offset"`
}

type GetWordsFrequenciesRow struct {
	Value     string `json:"value"`
	Frequency int64  `json:"frequency"`
}

func (q *Queries) GetWordsFrequencies(ctx context.Context, arg GetWordsFrequenciesParams) ([]GetWordsFrequenciesRow, error) {
	rows, err := q.db.Query(ctx, getWordsFrequencies, arg.Limit, arg.Offset)
	if err != nil {
		return nil, fmt.Errorf("query db: %w", err)
	}
	defer rows.Close()
	var items []GetWordsFrequenciesRow
	for rows.Next() {
		var i GetWordsFrequenciesRow
		if err := rows.Scan(&i.Value, &i.Frequency); err != nil {
			return nil, fmt.Errorf("scan db: %w", err)
		}
		items = append(items, i)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows: %w", err)
	}

	return items, nil
}

const getWordsRank = `-- name: GetWordsRank :many
SELECT
    words.value,
    ROW_NUMBER() OVER (ORDER BY count(*) DESC) as rank
FROM words
WHERE deleted_at IS NULL
GROUP BY words.value
ORDER BY rank ASC
LIMIT $1 OFFSET $2
`

type GetWordsRankParams struct {
	Limit  int32 `json:"limit"`
	Offset int32 `json:"offset"`
}

type GetWordsRankRow struct {
	Value string `json:"value"`
	Rank  int64  `json:"rank"`
}

func (q *Queries) GetWordsRank(ctx context.Context, arg GetWordsRankParams) ([]GetWordsRankRow, error) {
	rows, err := q.db.Query(ctx, getWordsRank, arg.Limit, arg.Offset)
	if err != nil {
		return nil, fmt.Errorf("query db: %w", err)
	}
	defer rows.Close()
	var items []GetWordsRankRow
	for rows.Next() {
		var i GetWordsRankRow
		if err := rows.Scan(&i.Value, &i.Rank); err != nil {
			return nil, fmt.Errorf("scan db: %w", err)
		}
		items = append(items, i)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows: %w", err)
	}

	return items, nil
}

const insertWord = `-- name: InsertWord :one
INSERT INTO words (value, created_at)
VALUES ($1, CURRENT_TIMESTAMP)
RETURNING id, value, created_at
`

type InsertWordRow struct {
	ID        int64              `json:"id"`
	Value     string             `json:"value"`
	CreatedAt pgtype.Timestamptz `json:"created_at"`
}

func (q *Queries) InsertWord(ctx context.Context, value string) (InsertWordRow, error) {
	row := q.db.QueryRow(ctx, insertWord, value)
	var i InsertWordRow
	err := row.Scan(&i.ID, &i.Value, &i.CreatedAt)

	return i, err
}
