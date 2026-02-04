package app

import (
	"context"

	"github.com/jackc/pgx/v5"
)

type Some interface {
	GetX(ctx context.Context) string
	SetX(tx pgx.Tx, x string) error
	Nothing()
	M(m map[string]pgx.Tx) map[string]pgx.Tx
	Slice(rows []string) error
	Anything(v int)
}
