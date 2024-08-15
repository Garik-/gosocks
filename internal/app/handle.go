package app

import (
	"bufio"
	"context"
	"log/slog"
	"net"
	"time"

	"github.com/Garik-/gosocks/internal/socks"
)

func (s *Server) handleConnection(ctx context.Context, srcConn net.Conn) {
	defer srcConn.Close()

	dstConn, err := socks.Handle(ctx, bufio.NewReader(srcConn), srcConn, time.Second*30)
	if err != nil {
		slog.Error("socket handle",
			slog.String("err", err.Error()),
		)
	}

	defer dstConn.Close()

	slog.Info("open connection")

	proxy := socks.NewProxy(srcConn, dstConn)
	err = proxy.Start(ctx, time.Second)
	if err != nil {
		slog.Error("socket handle",
			slog.String("err", err.Error()),
		)
	}

	slog.Info("close connection")
}
