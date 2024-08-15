package socks

import (
	"context"
	"io"
)

func StartTunnel(ctx context.Context, src, dst io.ReadWriteCloser) error {
	go (func() {
		<-ctx.Done()
		_ = src.Close()
		_ = dst.Close()
	})()

	go (func() {
		_, _ = io.Copy(src, dst)
	})()

	_, err := io.Copy(dst, src)
	return err
}
