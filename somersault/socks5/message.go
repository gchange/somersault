package socks5

type handshakeRequest struct {
	Version uint8
	NMethods uint8
	Methods []uint8
}

type handshakeResponse struct {
	Version uint8
	Method uint8
}

type connectRequest struct {
	Version uint8
	Command uint8
	Reverse uint8
	AddressType uint8
}

type connectResponse struct {
	Version uint8
	Response uint8
	Reverse uint8
	AddressType uint8
}

type connectAddress struct {
	Length uint8
	Address string
}

