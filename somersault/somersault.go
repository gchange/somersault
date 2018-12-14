package somersault

import (
	"errors"
	"fmt"
	"image"
	"image/jpeg"
	"log"
	"net"

	"github.com/somersault/somersault/socks5"
)

type ServerConfig struct {
	Protocol string `json:"protocol"`
	Command string `json:"command"`
	Method uint8 `json:"method"`
	Port int `json:"port"`
}

type ClientConfig struct {
	Protocol string `json:"protocol"`
	Command string `json:"command"`
	Method uint8 `json:"method"`
	Address string `json:"address"`
	Port int `json:"port"`
}

type Config struct {
	ServerConfig ServerConfig `json:"server"`
	ClientConfig ClientConfig `json:"client"`
}

type Somerasult struct {
	*Config
	logger log.Logger
	listener net.Listener
}

func (c *Config) New(logger log.Logger) (*Somerasult, error) {
	jpeg.Decode()
	s := Somerasult{
		c,
		logger,
		nil,
	}
	err := s.createServer()
	if err != nil {
		s.Close()
		return nil, err
	}
	err = s.createClient()
	if err != nil {
		s.Close()
		return nil, err
	}
	return &s, nil
}

func (s *Somerasult) createServer() error {
	if s.ServerConfig.Port == 0 {
		return nil
	}

	s.logger.Printf("create server %v\n", s.ServerConfig)
	addr := fmt.Sprintf(":%d", s.ServerConfig.Port)
	listener, err := net.Listen("tcp", addr)
	if err != nil {
		s.logger.Println(err)
		return err
	}
	s.logger.Printf("%v %s %v", listener, addr, listener.Addr())
	s.listener = listener

	switch s.ServerConfig.Protocol {
	case "socks5":
		go func() {
			config := socks5.Config{
				Command: "connect",
				Dialer: net.Dialer{},
			}
			for {
				conn, err := s.listener.Accept()
				if err != nil {
					continue
				}
				s.logger.Printf("connect %v", conn)
				config := socks5.Config{

				}
				socks := socks5.Socks5{
					Conn:conn,
					Command: "connect",
					Method: 0,
					Dialer: net.Dialer{},
				}
				go socks.Service()
			}
		}()
		return nil
	default:
		return errors.New("Unsupported protocol ")
	}
}

func (s *Somerasult) createClient() error {
	if s.ClientConfig.Address == "" {
		return nil
	}

	s.logger.Printf("create client %v\n", s.ClientConfig)

	switch s.ClientConfig.Protocol {
	case "socks5":
		addr := fmt.Sprintf("%s:%d", s.ClientConfig.Address, s.ClientConfig.Port)
		conn, err := net.Dial("tcp", addr)
		if err != nil {
			return err
		}
		socks := socks5.Socks5{
			Conn:conn,
			Command: "connect",
			Method: 0,
			Dialer: ProxyDialer{
				ServerAddress: s.ClientConfig.Address,
			Protocol: s.ClientConfig.Protocol,
		},
		}
		go socks.Service()
	default:
		return errors.New("Unsupported protocol ")
	}

	return nil
}

func (s *Somerasult) Close() {
	if s.listener != nil {
		s.listener.Close()
	}
}
