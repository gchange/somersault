package socks5

import "errors"

const (
	socksVersion = 5
	addrTypeIPv4 = 1
	addrTypeDomain = 3
	addrTypeIPv6 = 4
)

var (
	UnsupportedProtocol = errors.New("unsupported protocol")
	DuplicateAuthMethod = errors.New("duplicate auth method")
	UnsupportedAuthMethod = errors.New("unsupported auth method")
	UnsupportedCommand = errors.New("unsupported command")
	UnknownAddrType = errors.New("unknown address type")
)

func isValidVersion(version uint8) bool {
	return version == socksVersion
}
