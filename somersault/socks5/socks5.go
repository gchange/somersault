package socks5

import (
	"encoding/binary"
	"errors"
	"fmt"
	"github.com/somersault/somersault"
	"io"
	"log"
	"net"
	"sync"
)

type Config struct {
	Command string
	Dialer somersault.Dialer
}

type Socks5 struct {
	*Config
	Conn net.Conn
	Method uint8
	RemoteConn net.Conn
}

func New(config interface{}) (*Socks5, error) {
	if c, ok := config.(Config); ok {
		return &Socks5{
			&c,
			nil,
			nil,
			nil,
		}, nil
	}
	return nil, errors.New("unknown config")
}

func (s *Socks5) Handshake() error {
	methods := getSupportAuthMethod()
	_, err := s.Conn.Write([]byte{socksVersion, uint8(len(methods))})
	if err != nil {
		return err
	}
	_, err = s.Conn.Write(methods)
	if err != nil {
		return err
	}

	var version uint8
	var method uint8
	err = binary.Read(s.Conn, binary.BigEndian, &version)
	if err != nil {
		return err
	}
	err = binary.Read(s.Conn, binary.BigEndian, &method)
	if err != nil {
		return err
	}

	if !isValidVersion(version) {
		return UnsupportedProtocol
	}

	s.Method = method
	f := getAuthMethod(s.Method)
	if f == nil {
		return UnsupportedAuthMethod
	}
	err = f(s.Conn)
	if err != nil {
		return err
	}
	return nil
}

func (s *Socks5) HandshakeReply() error {
	var version uint8
	var nMethod uint8
	err := binary.Read(s.Conn, binary.BigEndian, &version)
	if err != nil {
		return err
	}
	err = binary.Read(s.Conn, binary.BigEndian, &nMethod)
	if err != nil {
		return err
	}
	if !isValidVersion(version) {
		return UnsupportedProtocol
	}

	methods := make([]byte, nMethod)
	_, err = s.Conn.Read(methods)

	s.Method = 0
	_, err = s.Conn.Write([]byte{version, s.Method})
	if err != nil {
		return err
	}

	f := getAuthReplyMethod(s.Method)
	if f == nil {
		return UnsupportedAuthMethod
	}
	return f(s.Conn)
}

func (s *Socks5) Connect(dstAddress string, dstPort int) error {
	var command uint8 = 0
	var reverse uint8 = 0
	var addressType uint8 = 0
	var port uint16 = 0

	switch s.Command {
	case "connect":
		command = 1
	case "bind":
		return UnsupportedCommand
	case "udp":
		return UnsupportedCommand
	default:
		return UnsupportedCommand
	}

	_, err := s.Conn.Write([]byte{socksVersion, command, reverse, addressType, uint8(len(dstAddress))})
	if err != nil {
		return err
	}
	_, err = s.Conn.Write([]byte(dstAddress))
	if err != nil {
		return err
	}
	err = binary.Write(s.Conn, binary.BigEndian, &port)
	if err != nil {
		return err
	}

	resp := struct {
		Version uint8
		Response uint8
		Reverse uint8
		AddressType uint8
	}{}
	err = binary.Read(s.Conn, binary.BigEndian, &resp)
	if err != nil {
		return err
	}
	if !isValidVersion(resp.Version) {
		return UnsupportedProtocol
	}
	if resp.Response != 0 {
		return errors.New("connect failed")
	}
	var ipLen uint8
	switch resp.AddressType {
	case addrTypeIPv4:
		ipLen = net.IPv4len
	case addrTypeIPv6:
		ipLen = net.IPv6len
	case addrTypeDomain:
		err = binary.Read(s.Conn, binary.BigEndian, &ipLen)
		if err != nil {
			return err
		}
	default:
		return UnsupportedCommand
	}
	ipBuf := make([]byte, ipLen)
	_, err = io.ReadFull(s.Conn, ipBuf)
	if err != nil {
		return err
	}
	var remotePort uint16
	err = binary.Read(s.Conn, binary.BigEndian, &remotePort)
	if err != nil {
		return err
	}
	return nil
}

func (s *Socks5) ConnectReply() error {
	req := struct {
		Version uint8
		Command uint8
		Reverse uint8
		AddressType uint8
	}{}
	err := binary.Read(s.Conn, binary.BigEndian, &req)
	if err != nil {
		return err
	}
	if !isValidVersion(req.Version) {
		 return UnsupportedProtocol
	}

	var addrLen uint8
	switch req.AddressType {
	case addrTypeIPv4:
		addrLen = net.IPv4len
	case addrTypeDomain:
		err = binary.Read(s.Conn, binary.BigEndian, &addrLen)
		if err != nil {
			return err
		}
	case addrTypeIPv6:
		addrLen = net.IPv6len
	default:
		return UnknownAddrType
	}
	remoteAddr := make([]byte, addrLen)
	_, err = io.ReadFull(s.Conn, remoteAddr)
	if err != nil {
		return err
	}
	var remotePort uint16
	err = binary.Read(s.Conn, binary.BigEndian, &remotePort)
	if err != nil {
		return err
	}

	switch req.Command {
	case 1:
		addr := fmt.Sprintf("%s:%d", string(remoteAddr), remotePort)
		log.Printf("remote %s\n", addr)
		s.RemoteConn, err = s.Dialer.Dial("tcp", addr)
		if err != nil {
			return err
		}
	case 2:
		return UnsupportedCommand
	case 3:
		return UnsupportedCommand
	default:
		return UnsupportedCommand
	}

	localAddr := s.RemoteConn.RemoteAddr().(*net.TCPAddr)
	var localIP []byte
	var localAddrType uint8
	if ip := localAddr.IP.To4(); ip != nil {
		localIP = ip
		localAddrType = addrTypeIPv4
	} else if ip := localAddr.IP.To16(); ip != nil {
		localIP = ip
		localAddrType = addrTypeIPv6
	} else {
		localIP = localAddr.IP
		localAddrType = addrTypeDomain
	}
	log.Println(localAddr, len(localAddr.IP))
	_, err = s.Conn.Write([]byte{req.Version, 0, req.Reverse, localAddrType})
	if err != nil {
		return err
	}
	switch localAddrType {
	case addrTypeIPv4, addrTypeIPv6:
	case addrTypeDomain:
		_, err = s.Conn.Write([]byte{uint8(len(localIP))})
		if err != nil {
			return err
		}
	default:
		return UnknownAddrType
	}
	_, err = s.Conn.Write(localIP)
	if err != nil {
		return err
	}
	err = binary.Write(s.Conn, binary.BigEndian, uint16(localAddr.Port))
	if err != nil {
		return err
	}
	return nil
}

func (s *Socks5) Transport() {
	log.Println("start transport", s.Conn.RemoteAddr(), s.RemoteConn.RemoteAddr())
	var wg sync.WaitGroup

	c := func(src, dst net.Conn) {
		defer wg.Done()
		io.Copy(dst, src)
		dst.Close()
	}

	wg.Add(2)
	go c(s.Conn, s.RemoteConn)
	go c(s.RemoteConn, s.Conn)
	wg.Wait()
}

func (s *Socks5) Service() {
	defer s.Close()
	err := s.HandshakeReply()
	if err != nil {
		log.Printf("handshake %v", err)
		return
	}
	err = s.ConnectReply()
	if err != nil {
		log.Printf("connect %v", err)
		return
	}
	s.Transport()
}

func (s *Socks5) Close() {
	if s.Conn != nil {
		s.Conn.Close()
	}
	if s.RemoteConn != nil {
		s.RemoteConn.Close()
	}
	log.Println("close socket")
}

func init() {
	somersault.RegisteConnector("socks5", New)
}
