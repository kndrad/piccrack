package textproc

import (
	"github.com/jackc/pgx/v5/pgtype"
)

type Word struct {
	ID        int64              `json:"id"`
	Value     string             `json:"value"`
	CreatedAt pgtype.Timestamptz `json:"created_at"`
	DeletedAt pgtype.Timestamptz `json:"deleted_at"`
}
