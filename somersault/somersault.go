package somersault

import (
	"errors"
	"fmt"
	"net"
	"socks5"
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
	listener net.Listener
}

func (c *Config) New() (*Somerasult, error) {
	s := Somerasult{
		c,
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
	switch s.ServerConfig.Protocol {
	case "socks5":
		addr := fmt.Sprintf(":%d", s.ServerConfig.Port)
		listener, err := net.Listen("tcp", addr)
		if err != nil {
			return err
		}
		s.listener = listener
		go func() {
			for {
				conn, err := s.listener.Accept()
				if err != nil {
					continue
				}
				socks := socks5.Socks5{
					conn
				}
			}
		}()
	default:
		return errors.New("Unsupported protocol ")
	}
}

func (s *Somerasult) createClient() error {
	if s.ClientConfig.Address == "" {
		return nil
	}
}

func (s *Somerasult) Close() {
	if s.listener != nil {
		s.listener.Close()
	}
}
