package socks

import (
	"context"
	"fmt"
	"io"
	"net"
	"time"

	"golang.org/x/sync/errgroup"
)

type Proxy struct {
	lconn, rconn net.Conn
}

func NewProxy(lconn, rconn net.Conn) *Proxy {
	return &Proxy{
		lconn: lconn,
		rconn: rconn,
	}
}

func (c *Proxy) Start(ctx context.Context, timeout time.Duration) error {
	eg, egCtx := errgroup.WithContext(ctx)

	eg.Go(func() error {
		return c.pipe(egCtx, c.lconn, c.rconn, timeout)
	})

	eg.Go(func() error {
		return c.pipe(egCtx, c.rconn, c.lconn, timeout)
	})

	return eg.Wait()
}

func (c *Proxy) pipe(ctx context.Context, src, dst net.Conn, timeout time.Duration) error {
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		err := dst.SetDeadline(time.Now().Add(timeout))
		if err != nil {
			return err
		}

		err = src.SetDeadline(time.Now().Add(timeout))
		if err != nil {
			return err
		}

		_, err = io.Copy(dst, src)
		if err != nil {
			return fmt.Errorf("io.Copy: %w", err)
		}
	}
}
