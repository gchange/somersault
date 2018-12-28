package socks5

import (
	"context"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"log"
	"net"

	"github.com/gchange/somersault/somersault/pipeline"
)

type Config struct {
	Command string `somersault:"command"`
	Network string `somersault:"network"`
	Address string `somersault:"address"`
	Port    uint16 `somersault:"port"`
	Reverse uint8
}

type Socks5 struct {
	*Config
	*pipeline.DefaultPipeline
}

func (c *Config) DeepCopy() pipeline.Config {
	return &Config{
		Command: c.Command,
		Network: c.Network,
		Address: c.Address,
		Port:    c.Port,
		Reverse: c.Reverse,
	}
}

func (c *Config) New(ctx context.Context, input, output pipeline.Pipeline) (pipeline.Pipeline, error) {
	if input == nil {
		return nil, errors.New("input not found")
	}
	output, err := c.HandshakeReply(input)
	fmt.Println(output, err)
	if err != nil {
		return nil, err
	}
	dp, err := pipeline.NewDefaultPipeline(ctx, input, output)
	if err != nil {
		return nil, err
	}
	s := &Socks5{
		c,
		dp,
	}
	go s.Transport()
	return s, nil
}

func (c *Config) Connect(command uint8, address string, port uint16) (net.Conn, error) {
	fmt.Println(command, address, port, "server info", c.Address, c.Port)
	if c.Address != "" && c.Port != 0 {
		return c.ConnectToServer(command, address, port)
	}
	network := ""
	switch command {
	case 1:
		network = "tcp"
	case 2:
		return nil, UnsupportedCommand
	case 3:
		network = "udp"
	default:
		return nil, UnsupportedCommand
	}
	addr := fmt.Sprintf("%s:%d", address, port)
	fmt.Println(addr)
	conn, err := net.Dial(network, addr)
	if err != nil {
		return nil, err
	}
	return conn, nil
}

func (c *Config) ConnectToServer(command uint8, address string, port uint16) (net.Conn, error) {
	fmt.Println("connect to server", command, address, port)
	addr := fmt.Sprintf("%s:%d", c.Address, c.Port)
	conn, err := net.Dial(c.Network, addr)
	if err != nil {
		return nil, err
	}
	err = c.Handshake(conn, command, address, port)
	if err != nil {
		return nil, err
	}
	return conn, nil
}

func (c *Config) Handshake(input pipeline.Pipeline, command uint8, address string, port uint16) error {
	methods := getSupportAuthMethod()
	_, err := input.Write([]byte{socksVersion, uint8(len(methods))})
	if err != nil {
		return err
	}
	_, err = input.Write(methods)
	if err != nil {
		return err
	}

	var version uint8
	var method uint8
	err = binary.Read(input, binary.BigEndian, &version)
	if err != nil {
		return err
	}
	err = binary.Read(input, binary.BigEndian, &method)
	if err != nil {
		return err
	}

	if !isValidVersion(version) {
		return UnsupportedProtocol
	}

	f := getAuthMethod(method)
	if f == nil {
		return UnsupportedAuthMethod
	}
	err = f(input)
	if err != nil {
		return err
	}

	_, err = input.Write([]byte{socksVersion, command, c.Reverse})
	if err != nil {
		return err
	}

	addressType := addrTypeDomain
	ip := net.ParseIP(address)
	if ip := ip.To4(); ip != nil {
		addressType := addrTypeIPv4
		_, err = input.Write([]byte{uint8(addressType)})
		if err != nil {
			return err
		}
		_, err = input.Write(ip)
	} else if ip := ip.To16(); ip != nil {
		addressType := addrTypeIPv6
		_, err = input.Write([]byte{uint8(addressType)})
		if err != nil {
			return err
		}
		_, err = input.Write(ip)
	} else {
		_, err = input.Write([]byte{uint8(addressType), uint8(len(address))})
		if err != nil {
			return err
		}
		_, err = input.Write([]byte(address))
	}
	if err != nil {
		return err
	}

	err = binary.Write(input, binary.BigEndian, &port)
	if err != nil {
		return err
	}

	resp := struct {
		Version     uint8
		Response    uint8
		Reverse     uint8
		AddressType uint8
	}{}
	err = binary.Read(input, binary.BigEndian, &resp)
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
		err = binary.Read(input, binary.BigEndian, &ipLen)
		if err != nil {
			return err
		}
	default:
		return UnsupportedCommand
	}
	ipBuf := make([]byte, ipLen)
	_, err = io.ReadFull(input, ipBuf)
	if err != nil {
		return err
	}
	var remotePort uint16
	err = binary.Read(input, binary.BigEndian, &remotePort)
	if err != nil {
		return err
	}
	return nil
}

func (c *Config) HandshakeReply(input pipeline.Pipeline) (pipeline.Pipeline, error) {
	var version uint8
	var nMethod uint8
	err := binary.Read(input, binary.BigEndian, &version)
	if err != nil {
		return nil, err
	}
	err = binary.Read(input, binary.BigEndian, &nMethod)
	if err != nil {
		return nil, err
	}
	if !isValidVersion(version) {
		return nil, UnsupportedProtocol
	}

	methods := make([]byte, nMethod)
	_, err = input.Read(methods)

	fmt.Println(version, methods)
	var method uint8 = 0
	_, err = input.Write([]byte{version, method})
	if err != nil {
		return nil, err
	}

	f := getAuthReplyMethod(method)
	if f == nil {
		return nil, UnsupportedAuthMethod
	}
	err = f(input)
	if err != nil {
		return nil, err
	}

	req := struct {
		Version     uint8
		Command     uint8
		Reverse     uint8
		AddressType uint8
	}{}
	err = binary.Read(input, binary.BigEndian, &req)
	fmt.Println(req, err)
	if err != nil {
		return nil, err
	}
	if !isValidVersion(req.Version) {
		return nil, UnsupportedProtocol
	}

	var remoteIP net.IP
	remoteAddr := ""
	switch req.AddressType {
	case addrTypeIPv4:
		remoteIP = make(net.IP, net.IPv4len)
		_, err = io.ReadFull(input, remoteIP)
		if err != nil {
			return nil, err
		}
		remoteAddr = remoteIP.String()
	case addrTypeDomain:
		var addrLen uint8
		err = binary.Read(input, binary.BigEndian, &addrLen)
		fmt.Println("domain len", addrLen)
		if err != nil {
			return nil, err
		}
		remoteIP = make(net.IP, addrLen)
		_, err = io.ReadFull(input, remoteIP)
		if err != nil {
			return nil, err
		}
		remoteAddr = string(remoteIP)
	case addrTypeIPv6:
		remoteIP = make(net.IP, net.IPv6len)
		_, err = io.ReadFull(input, remoteIP)
		if err != nil {
			return nil, err
		}
		remoteAddr = remoteIP.String()
	default:
		return nil, UnknownAddrType
	}
	fmt.Println(remoteAddr, err)
	if err != nil {
		return nil, err
	}
	var remotePort uint16
	err = binary.Read(input, binary.BigEndian, &remotePort)
	fmt.Println("remote port", remotePort, err)
	if err != nil {
		return nil, err
	}

	conn, err := c.Connect(req.Command, string(remoteAddr), remotePort)
	fmt.Println(conn, err)
	if err != nil {
		return nil, err
	}

	localAddr := conn.RemoteAddr().(*net.TCPAddr)
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
	_, err = input.Write([]byte{req.Version, 0, req.Reverse, localAddrType})
	if err != nil {
		return nil, err
	}
	switch localAddrType {
	case addrTypeIPv4, addrTypeIPv6:
	case addrTypeDomain:
		_, err = input.Write([]byte{uint8(len(localIP))})
		if err != nil {
			return nil, err
		}
	default:
		return nil, UnknownAddrType
	}
	_, err = input.Write(localIP)
	if err != nil {
		return nil, err
	}
	err = binary.Write(input, binary.BigEndian, uint16(localAddr.Port))
	if err != nil {
		return nil, err
	}
	return conn, nil
}

func init() {
	config := &Config{
		Network: "tcp",
		Reverse: 0,
	}
	pipeline.RegistePipelineCreator("socks5", config)
}
