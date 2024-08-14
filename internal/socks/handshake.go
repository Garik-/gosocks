package socks

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"net"
	"strconv"
)

const (
	version                  byte = 5
	noAuthenticationRequired byte = 0
	noAcceptableMethods      byte = 0xFF

	aTypeIPv4   byte = 1
	aTypeDomain byte = 3
	aTypeIPv6   byte = 4

	cmdConnect byte = 1

	// replyOk                 byte = 0 .
	replyCmdUnsupport  byte = 7
	replyAddrUnsupport byte = 8
	// REPLY_ERROR             byte = 1 .
	// REPLY_HOST_UNACCESSIBLE byte = 4 .
	// REPLY_ERROR_CONNECT     byte = 5 .
)

type s5Request struct {
	Version     byte
	Command     byte
	Reserved    byte
	AddressType byte
}

type s5Replay struct {
	address     []byte
	port        uint16
	version     byte
	replay      byte
	reserved    byte
	addressType byte
}

func (r *s5Replay) Bytes() []byte {
	buf := make([]byte, 6+len(r.address))

	buf[0] = r.version
	buf[1] = r.replay
	buf[2] = r.reserved
	buf[3] = r.addressType

	j := 4

	for i := range r.address {
		buf[j] = r.address[i]
		j++
	}

	buf[j] = byte(r.port >> 8)
	buf[j+1] = byte(r.port & 0xff)

	return buf
}

var (
	errBadRequest             = errors.New("bad request")
	errCommandUnsupported     = errors.New("CMD UNSUPPORT")
	errAddressTypeUnsupported = errors.New("ATYP UNSUPPORT")
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

func readAddress(atyp byte, r io.Reader) ([]byte, error) {
	var n byte

	switch atyp {
	case aTypeIPv4:
		n = 4
	case aTypeIPv6:
		n = 16
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
