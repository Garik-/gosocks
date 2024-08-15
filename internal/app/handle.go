package app

import (
	"bufio"
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"log/slog"
	"net"
	"syscall"
	"time"

	"github.com/Garik-/gosocks/internal/socks"
)

func (s *Server) handleConnection(ctx context.Context, srcConn net.Conn) {
	defer srcConn.Close()

	ID := createClientID()

	dstConn, err := socks.Handshake(ctx, bufio.NewReader(srcConn), srcConn, time.Second*30)
	if err != nil {
		slog.Error("socket handle",
			slog.String("err", err.Error()),
		)

		return
	}

	defer dstConn.Close()

	slog.Debug("open connection",
		slog.String("id", ID),
		slog.String("address", dstConn.RemoteAddr().String()))

	err = socks.StartTunnel(ctx, srcConn, dstConn)
	if err != nil && !errors.Is(err, syscall.ECONNRESET) {
		slog.Error("socket handle",
			slog.String("err", err.Error()),
		)
	}

	slog.Debug("close connection",
		slog.String("id", ID),
	)
}

func createClientID() string {
	b := make([]byte, 6)

	_, err := rand.Read(b)
	if err != nil {
		return ""
	}

	return hex.EncodeToString(b)
}
