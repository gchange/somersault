package somersault

import (
	"github.com/somersault/somersault/socks5"
	"net"
	"strconv"
	"strings"
)

type Dialer interface {
	Dial(network, address string) (net.Conn, error)
}

type ProxyDialer struct {
	ServerAddress string
	Protocol string
}

func (pd *ProxyDialer) Dial(network, address string) (net.Conn, error) {
	switch pd.Protocol {
	case "socks5":
		conn, err := net.Dial("tcp", pd.ServerAddress)
		if err != nil {
			return nil, err
		}
		socks := socks5.Socks5{
			Conn: conn,
			Command: "connect",
			Method: 0,
			Dialer: net.Dialer{},
		}
		err = socks.Handshake()
		if err != nil {
			return nil, err
		}
		index := strings.LastIndex(address, ":")
		port, err := strconv.Atoi(address[index+1:])
		if err != nil {
			return nil, err
		}
		err = socks.Connect(address[0:index], port)
		if err != nil {
			return nil, err
		}
		return conn, nil
	}
	return nil, socks5.UnsupportedProtocol
}
