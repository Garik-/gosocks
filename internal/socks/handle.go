package socks

import (
	"context"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net"
	"strconv"
	"time"
)

func Handle(_ context.Context, r io.Reader, w io.Writer) (net.Conn, error) {
	methods, err := handshakeMethods(r)
	if err != nil {
		return nil, fmt.Errorf("handshakeMethods: %w", err)
	}

	err = selectMethod(methods, w)
	if err != nil {
		return nil, fmt.Errorf("selectMethod: %w", err)
	}

	rep, err := replay(r)
	if err != nil {
		if errors.Is(err, errCommandUnsupported) || errors.Is(err, errAddressTypeUnsupported) {
			_ = binary.Write(w, binary.BigEndian, rep.Bytes())
		}

		return nil, fmt.Errorf("request: %w", err)
	}

	slog.Info("try connect...", slog.String("address", dialAddress(rep.addressType, rep.address, rep.port)))

	// TODO: move to config dialTimeout
	c, err := net.DialTimeout("tcp", dialAddress(rep.addressType, rep.address, rep.port), time.Second*30)
	if err != nil {
		rep.replay = REPLY_ERROR_CONNECT
		err = binary.Write(w, binary.BigEndian, rep.Bytes())

		return nil, fmt.Errorf("net.Dial: %w", err)
	} else {
		rep.replay = replyOk
		rep.addressType = aTypeIPv4
		rep.address, rep.port = splitHostPort(c.LocalAddr().String())

		if len(rep.address) == net.IPv6len {
			rep.addressType = aTypeIPv6
		}

		//slog.Info(c.RemoteAddr().String())
		//slog.Info(c.LocalAddr().String())

		err = binary.Write(w, binary.BigEndian, rep.Bytes())
		if err != nil {
			return nil, fmt.Errorf("binary.Write: %w", err)
		}
	}

	return c, nil
}

func splitHostPort(hostport string) ([]byte, uint16) {
	ip, port, _ := net.SplitHostPort(hostport)
	i := net.ParseIP(ip).To4()
	p, _ := strconv.ParseUint(port, 10, 16)

	return i, uint16(p)
}
