package app

import (
	"context"
	"io"
)

type Some interface {
	GetX(ctx context.Context) string
	Nothing()
	M(m map[string]int) map[string]int
	Slice(rows []string) error
	Anything(v int)
	Multi() (string, int, error)
}

type Base interface {
	Ping(ctx context.Context) error
}

type Embedded interface {
	Base
	io.ReaderFrom

	Own(x int) (string, error)
}

type Handler func(ctx context.Context, id string) (string, error)

type Ticker func()

type Filter func(rows []string, mode int, ignored bool) error
