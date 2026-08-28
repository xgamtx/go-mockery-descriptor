package embeds

import (
	"context"
	"io"
)

type Base interface {
	Ping(ctx context.Context) error
}

type Middle interface {
	Base

	Mid(rows []string) (map[string]int, error)
}

// Transitive проверяет транзитивные встроенные интерфейсы, встроенный интерфейс из
// другого пакета и пустой интерфейс.
type Transitive interface {
	Middle
	io.ReaderFrom
	any

	Own(x int) (string, error)
}

// Overlapping проверяет пересечение наборов методов встроенных интерфейсов.
type Overlapping interface {
	Base
	Middle
}
