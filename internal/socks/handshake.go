package socks

import (
	"bytes"
	"encoding/binary"
	"errors"
	"io"
	"net"
	"strconv"
)

const (
	VER                        byte = 0x05
	NO_AUTHENTICATION_REQUIRED byte = 0x00
	NO_ACCEPTABLE_METHODS      byte = 0xFF

	ATYP_IPV4   byte = 1
	ATYP_DOMAIN byte = 3
	ATYP_IPV6   byte = 4

	CMD_CONNECT byte = 1

	REPLY_OK                byte = 0
	REPLY_CMD_UNSUPPORT     byte = 7
	REPLY_ADD_UNSUPPORT     byte = 8
	REPLY_ERROR             byte = 1
	REPLY_HOST_UNACCESSIBLE byte = 4
	REPLY_ERROR_CONNECT     byte = 5
)

type s5Request struct {
	VER  byte
	CMD  byte
	RSV  byte
	ATYP byte
}

type s5Response struct {
	VER  byte
	REP  byte
	RSV  byte
	ATYP byte
	ADDR []byte
	PORT uint16
}

func (r *s5Response) bytes() []byte {
	buf := make([]byte, 6+len(r.ADDR))

	buf[0] = r.VER
	buf[1] = r.REP
	buf[2] = r.RSV
	buf[3] = r.ATYP

	j := 4

	for i := range r.ADDR {
		buf[j] = r.ADDR[i]
		j++
	}

	buf[j] = byte(r.PORT >> 8)
	buf[j+1] = byte(r.PORT & 0xff)

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
		return nil, err
	}

	if byte(h>>8) != VER {
		return nil, errBadRequest
	}

	methods := make([]byte, byte(h&0xff))

	err = binary.Read(reader, binary.BigEndian, &methods)
	if err != nil {
		return nil, err
	}

	return methods, nil
}

func selectMethod(methods []byte, writer io.Writer) error {
	method := NO_AUTHENTICATION_REQUIRED

	if bytes.IndexByte(methods, method) == -1 {
		method = NO_ACCEPTABLE_METHODS
	}

	return binary.Write(writer, binary.BigEndian, [2]byte{VER, method})
}

func address(atyp byte, addr []byte, port uint16) string {
	var host string

	switch atyp {
	case ATYP_IPV4, ATYP_IPV6:
		host = net.IP(addr).String()
	case ATYP_DOMAIN:
		host = string(addr)
	}

	return net.JoinHostPort(host, strconv.FormatUint(uint64(port), 10))
}

func requestAddress(atyp byte, r io.Reader) ([]byte, error) {
	var n byte

	switch atyp {
	case ATYP_IPV4:
		n = 4
	case ATYP_IPV6:
		n = 16
	case ATYP_DOMAIN:
		err := binary.Read(r, binary.BigEndian, &n)
		if err != nil {
			return nil, err
		}
	}

	addr := make([]byte, n)

	err := binary.Read(r, binary.BigEndian, &addr)
	if err != nil {
		return nil, err
	}

	return addr, nil
}

func request(r io.Reader) (*s5Response, error) {
	var req s5Request

	err := binary.Read(r, binary.BigEndian, &req)
	if err != nil {
		return nil, err
	}

	if req.VER != VER {
		return nil, errBadRequest
	}

	addr, err := requestAddress(req.ATYP, r)
	if err != nil {
		return nil, err
	}

	var port uint16

	err = binary.Read(r, binary.BigEndian, &port)
	if err != nil {
		return nil, err
	}

	res := &s5Response{
		VER:  req.VER,
		REP:  0xFF,
		RSV:  req.RSV,
		ATYP: req.ATYP,
		ADDR: addr,
		PORT: port,
	}

	if req.CMD != CMD_CONNECT {
		res.REP = REPLY_CMD_UNSUPPORT

		return res, errCommandUnsupported
	}

	if bytes.IndexByte([]byte{ATYP_IPV4, ATYP_DOMAIN, ATYP_IPV6}, req.ATYP) == -1 {
		res.REP = REPLY_ADD_UNSUPPORT

		return res, errAddressTypeUnsupported
	}

	return res, nil
}
