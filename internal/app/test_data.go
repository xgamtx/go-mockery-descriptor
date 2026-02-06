package app

import "context"

type Some interface {
	GetX(ctx context.Context) string
	Nothing()
	M(m map[string]int) map[string]int
	Slice(rows []string) error
	Anything(v int)
}
