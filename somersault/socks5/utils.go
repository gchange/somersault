package socks5

import "errors"

const (
	socksVersion = 5
)

var (
	UnsupportedProtocol = errors.New("unsupported protocol")
	DuplicateAuthMethod = errors.New("duplicate auth method")
	UnsupportedAuthMethod = errors.New("unsupported auth method")
	UnsupportedCommand = errors.New("unsupported command")
)

func isValidVersion(version uint8) bool {
	return version == socksVersion
}
