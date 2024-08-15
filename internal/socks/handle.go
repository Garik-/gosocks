package socks

import (
	"context"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"net"
	"strconv"
	"time"
)

func Handle(_ context.Context, r io.Reader, w io.Writer, timeout time.Duration) (net.Conn, error) {
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

	c, err := net.DialTimeout("tcp", dialAddress(rep.addressType, rep.address, rep.port), timeout)
	if err != nil {
		rep.replay = REPLY_ERROR_CONNECT
		errWrite := binary.Write(w, binary.BigEndian, rep.Bytes())

		return nil, fmt.Errorf("net.Dial: %w %w", err, errWrite)
	} else {
		rep.replay = replyOk
		rep.addressType = aTypeIPv4
		rep.address, rep.port, err = splitHostPort(c.LocalAddr().String())

		if err != nil {
			rep.replay = REPLY_ERROR
			errWrite := binary.Write(w, binary.BigEndian, rep.Bytes())

			return nil, fmt.Errorf("splitHostPort: %w %w", err, errWrite)
		}

		if len(rep.address) == net.IPv6len {
			rep.addressType = aTypeIPv6
		}

		err = binary.Write(w, binary.BigEndian, rep.Bytes())
		if err != nil {
			return nil, fmt.Errorf("binary.Write: %w", err)
		}
	}

	return c, nil
}

func splitHostPort(address string) ([]byte, uint16, error) {
	ip, port, err := net.SplitHostPort(address)
	if err != nil {
		return nil, 0, err
	}

	i := net.ParseIP(ip).To4()
	p, err := strconv.ParseUint(port, 10, 16)
	if err != nil {
		return nil, 0, err
	}

	return i, uint16(p), nil
}
