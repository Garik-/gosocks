package socks

import "errors"

const (
	version                  byte = 5
	noAuthenticationRequired byte = 0
	noAcceptableMethods      byte = 0xFF

	aTypeIPv4   byte = 1
	aTypeDomain byte = 3
	aTypeIPv6   byte = 4

	cmdConnect byte = 1

	replyOk            byte = 0
	replyCmdUnsupport  byte = 7
	replyAddrUnsupport byte = 8
	replyError         byte = 1
	replyErrorConnect  byte = 5
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

func (r *s5Replay) Network() string {
	return "tcp"
}

func (r *s5Replay) String() string {
	return dialAddress(r.addressType, r.address, r.port)
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
