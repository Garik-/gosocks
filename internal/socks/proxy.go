package socks

import (
	"context"
	"io"

	"golang.org/x/sync/errgroup"
)

func StartTunnel(ctx context.Context, src, dst io.ReadWriteCloser) error {
	eg, egCtx := errgroup.WithContext(ctx)

	eg.Go(func() error {
		<-egCtx.Done()
		_ = src.Close()
		_ = dst.Close()

		return nil
	})

	eg.Go(func() error {
		_, err := io.Copy(src, dst)
		return err
	})

	eg.Go(func() error {
		_, err := io.Copy(dst, src)
		return err
	})

	return eg.Wait()
}
