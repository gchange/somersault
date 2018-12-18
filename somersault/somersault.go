package somersault

import (
	"errors"
	"fmt"
	"github.com/somersault/somersault/pipeline"
	"log"
	"net"

	"github.com/somersault/somersault/socks5"
)

type Config struct {
	ServerConfig map[interface{}]interface{} `json:"server"`
	ClientConfig map[interface{}]interface{} `json:"client"`
}

type Somerasult struct {
	*Config
	logger log.Logger
	listener net.Listener
}

func (c *Config) New(logger log.Logger) (*Somerasult, error) {
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
	port, ok := getIntFromMap(s.ServerConfig, "listen")
	if !ok || port <= 0 {
		return nil
	}

	address, ok := getStringFromMap(s.ServerConfig, "address")
	if !ok {
		address = ""
	}

	protocol, ok := getStringFromMap(s.ServerConfig, "protocol")
	if !ok {
		protocol = "tcp"
	}

	ps, ok := s.ServerConfig["pipeline"]
	if !ok {
		return nil
	}
	pcs, ok := ps.([]map[string]interface{})
	if !ok {
		return nil
	}
	if len(pcs) == 0 {
		return errors.New("pipeline not config")
	}

	addr := fmt.Sprintf("%s:%d", address, port)
	s.logger.Printf("create server listen %s\n", addr)
	defer s.logger.Printf("close server on %s\n", addr)
	listener, err := net.Listen(protocol, addr)
	if err != nil {
		s.logger.Println(err)
		return err
	}
	s.listener = listener

	chain := make([]pipeline.Pipeline, len(pcs))
	for i, p := range pcs {
		protocol, _ := getStringFromMap(p, "protocol")
		config, _ := p["config"]
		c, err := pipeline.GetPipelineCreator(protocol, config)
		if err != nil {
			return err
		}
		chain = append(chain, c)
	}

	go func() {
		for {
			conn, err := s.listener.Accept()
			if err != nil {
				continue
			}

		}
	}()

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
