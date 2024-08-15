package socks

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
	"net"
	"strconv"
)

func handshakeMethods(reader io.Reader) ([]byte, error) {
	var h uint16

	err := binary.Read(reader, binary.BigEndian, &h)
	if err != nil {
		return nil, fmt.Errorf("binary.Read: %w", err)
	}

	if byte(h>>8) != version {
		return nil, errBadRequest
	}

	methods := make([]byte, byte(h&0xff))

	err = binary.Read(reader, binary.BigEndian, &methods)
	if err != nil {
		return nil, fmt.Errorf("binary.Read: %w", err)
	}

	return methods, nil
}

func selectMethod(methods []byte, writer io.Writer) error {
	method := noAuthenticationRequired

	if bytes.IndexByte(methods, method) == -1 {
		method = noAcceptableMethods
	}

	err := binary.Write(writer, binary.BigEndian, [2]byte{version, method})
	if err != nil {
		return fmt.Errorf("binary.Write: %w", err)
	}

	return nil
}

func dialAddress(aType byte, addr []byte, port uint16) string {
	var host string

	switch aType {
	case aTypeIPv4, aTypeIPv6:
		host = net.IP(addr).String()
	case aTypeDomain:
		host = string(addr)
	}

	return net.JoinHostPort(host, strconv.FormatUint(uint64(port), 10))
}

func readAddress(addressType byte, r io.Reader) ([]byte, error) {
	var n byte

	switch addressType {
	case aTypeIPv4:
		n = net.IPv4len
	case aTypeIPv6:
		n = net.IPv6len
	case aTypeDomain:
		err := binary.Read(r, binary.BigEndian, &n)
		if err != nil {
			return nil, fmt.Errorf("binary.Read: %w", err)
		}
	}

	addr := make([]byte, n)

	err := binary.Read(r, binary.BigEndian, &addr)
	if err != nil {
		return nil, fmt.Errorf("binary.Read: %w", err)
	}

	return addr, nil
}

func readPort(r io.Reader) (uint16, error) {
	var port uint16

	err := binary.Read(r, binary.BigEndian, &port)
	if err != nil {
		return port, fmt.Errorf("binary.Read: %w", err)
	}

	return port, nil
}

func replay(r io.Reader) (*s5Replay, error) {
	var req s5Request

	err := binary.Read(r, binary.BigEndian, &req)
	if err != nil {
		return nil, fmt.Errorf("binary.Read: %w", err)
	}

	if req.Version != version {
		return nil, errBadRequest
	}

	addr, err := readAddress(req.AddressType, r)
	if err != nil {
		return nil, err
	}

	port, err := readPort(r)
	if err != nil {
		return nil, err
	}

	res := &s5Replay{
		version:     req.Version,
		replay:      0xFF,
		reserved:    req.Reserved,
		addressType: req.AddressType,
		address:     addr,
		port:        port,
	}

	if req.Command != cmdConnect {
		res.replay = replyCmdUnsupport

		return res, errCommandUnsupported
	}

	allowAddressTypes := []byte{aTypeIPv4, aTypeDomain, aTypeIPv6}
	if bytes.IndexByte(allowAddressTypes, req.AddressType) == -1 {
		res.replay = replyAddrUnsupport

		return res, errAddressTypeUnsupported
	}

	return res, nil
}
