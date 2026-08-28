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

// Transitive covers transitive embedding, an interface embedded from another package
// and an empty interface.
type Transitive interface {
	Middle
	io.ReaderFrom
	any

	Own(x int) (string, error)
}

// Overlapping covers intersecting method sets of embedded interfaces.
type Overlapping interface {
	Base
	Middle
}
