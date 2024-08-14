package socks

import (
	"context"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
)

func Handle(_ context.Context, r io.Reader, w io.Writer) error {
	methods, err := handshakeMethods(r)
	if err != nil {
		return fmt.Errorf("handshakeMethods: %w", err)
	}

	err = selectMethod(methods, w)
	if err != nil {
		return fmt.Errorf("selectMethod: %w", err)
	}

	res, err := request(r)
	if err != nil {
		if errors.Is(err, errCommandUnsupported) || errors.Is(err, errAddressTypeUnsupported) {
			_ = binary.Write(w, binary.BigEndian, res.bytes())
		}

		return fmt.Errorf("request: %w", err)
	}

	return nil
}
