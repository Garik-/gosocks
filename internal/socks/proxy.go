package socks

import (
	"context"
	"fmt"
	"io"

	"golang.org/x/sync/errgroup"
)

func StartTunnel(ctx context.Context, src, dst io.ReadWriter) error {
	eg, egCtx := errgroup.WithContext(ctx)

	eg.Go(func() error {
		return pipe(egCtx, src, dst)
	})

	eg.Go(func() error {
		return pipe(egCtx, dst, src)
	})

	return eg.Wait()
}

func pipe(ctx context.Context, src, dst io.ReadWriter) error {
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		_, err := io.Copy(dst, src)
		if err != nil {
			return fmt.Errorf("io.Copy: %w", err)
		}
	}
}
