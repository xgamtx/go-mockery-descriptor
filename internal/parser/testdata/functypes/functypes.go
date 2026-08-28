package functypes

import "context"

type Handler func(ctx context.Context, id string) (string, error)

type Ticker func()
