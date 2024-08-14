package app

import (
	"bufio"
	"context"
	"log/slog"
	"net"

	"github.com/Garik-/gosocks/internal/socks"
)

func (s *Server) handleConnection(ctx context.Context, conn net.Conn) {
	defer conn.Close()

	reader := bufio.NewReader(conn)

	err := socks.Handle(ctx, reader, conn)
	if err != nil {
		slog.Error("socket handle",
			slog.String("err", err.Error()),
		)
	}
}
