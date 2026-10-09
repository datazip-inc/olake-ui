package utils

import (
	"context"

	"golang.org/x/sync/errgroup"
)

// ForEachConcurrently calls fn for each item, at most limit at a time, with the item's index. The
// first error cancels the context passed to the calls in flight, stops new calls, and is returned;
// otherwise it returns ctx's error, if ctx ended before every item was started.
func ForEachConcurrently[T any](ctx context.Context, items []T, limit int, fn func(ctx context.Context, i int, item T) error) error {
	g, gctx := errgroup.WithContext(ctx)
	g.SetLimit(limit)
	for i, item := range items {
		if gctx.Err() != nil {
			break
		}
		g.Go(func() error { return fn(gctx, i, item) })
	}
	if err := g.Wait(); err != nil {
		return err
	}
	return ctx.Err()
}
