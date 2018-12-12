package socks5

import (
	"encoding/binary"
	"net"
)

type Socks5 struct {
	Command string
	Conn net.Conn
	Method uint8
}

func (s *Socks5) Handshake() error {
	methods := getSupportAuthMethod()
	req := handshakeRequest{
		Version: socksVersion,
		NMethods: uint8(len(methods)),
		Methods: methods,
	}
	err := binary.Write(s.Conn, binary.BigEndian, req)
	if err != nil {
		return err
	}

	resp := handshakeResponse{}
	err = binary.Read(s.Conn, binary.BigEndian, &resp)
	if err != nil {
		return err
	}

	if !isValidVersion(resp.Version) {
		return UnsupportedProtocol
	}

	s.Method = resp.Method
	return s.Auth()
}

func (s *Socks5) HandshakeReply() error {
	req := handshakeRequest{}
	err := binary.Read(s.Conn, binary.BigEndian, &req)
	if err != nil {
		return err
	}

	if !isValidVersion(req.Version) {
		return UnsupportedProtocol
	}

	s.Method = 0
	resp := handshakeResponse{
		Version: req.Version,
		Method: s.Method,
	}
	err = binary.Write(s.Conn, binary.BigEndian, &resp)
	if err != nil {
		return err
	}
	return s.AuthReply()
}

func (s *Socks5) Auth() error {
	f := getAuthMethod(s.Method)
	if f == nil {
		return UnsupportedAuthMethod
	}
	return f(s.Conn)
}

func (s *Socks5) AuthReply() error {
	f := getAuthReplyMethod(s.Method)
	if f == nil {
		return UnsupportedAuthMethod
	}
	return f(s.Conn)
}

func (s *Socks5) Connect() error {
	req := connectRequest{
		Version: socksVersion,
		Reverse: 0,
	}
	address := connectAddress{
		0,
		"",
	}
	var port uint16 = 0
	switch s.Command {
	case "connect":
		req.Command = 1
		req.AddressType = 0
	case "bind":
		return UnsupportedCommand
	case "udp":
		return UnsupportedCommand
	default:
		return UnsupportedCommand
	}

	err := binary.Write(s.Conn, binary.BigEndian, &req)
	if err != nil {
		return err
	}
	err = binary.Write(s.Conn, binary.BigEndian, &address)
	if err != nil {
		return err
	}
	return binary.Write(s.Conn, binary.BigEndian, &port)
}

func (s *Socks5) ConnectReply() error {
	req := connectRequest{}
	err := binary.Read(s.Conn, binary.BigEndian, &req)
	if err != nil {
		return err
	}
	if !isValidVersion(req.Version) {
		 return UnsupportedProtocol
	}

	resp := connectResponse{
		Version: req.Version,
		Response: req.Command,
		Reverse: 0,
	}
	address := connectAddress{
		Length: 0,
		Address: "",
	}
	var port uint16 = 0
	switch req.Command {
	case 1:
	case 2:
		return UnsupportedCommand
	case 3:
		return UnsupportedCommand
	default:
		return UnsupportedCommand
	}
	err := binary.Write(s.Conn, binary.BigEndian, &resp)
	if err != nil {
		return err
	}
	err = binary.Write(s.Conn, binary.BigEndian, &address)
	if err != nil {
		return err
	}
	return binary.Write(s.Conn, binary.BigEndian, &port)
}
