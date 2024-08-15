package socks

import (
	"context"
	"fmt"
	"golang.org/x/sync/errgroup"
	"io"
)

type Proxy struct {
	lconn, rconn io.ReadWriter
}

func NewProxy(lconn, rconn io.ReadWriter) *Proxy {
	return &Proxy{
		lconn: lconn,
		rconn: rconn,
	}
}

func (c *Proxy) Start(ctx context.Context) error {
	eg, egCtx := errgroup.WithContext(ctx)

	eg.Go(func() error {
		return c.pipe(egCtx, c.lconn, c.rconn)
	})

	eg.Go(func() error {
		return c.pipe(egCtx, c.rconn, c.lconn)
	})

	return eg.Wait()
}

func (c *Proxy) pipe(ctx context.Context, src, dst io.ReadWriter) error {
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
