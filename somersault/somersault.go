package somersault

import (
	"context"
	"fmt"
	"log"
	"net"
	"reflect"

	"github.com/gchange/somersault/somersault/pipeline"
)

type Config struct {
	Config []map[string]interface{} `json:"config"`
}

type Somerasult struct {
	*Config
	logger *log.Logger
}

func (c *Config) New(logger *log.Logger) (*Somerasult, error) {
	s := Somerasult{
		c,
		logger,
	}
	s.logger.Println(c.Config, c)
	for _, config := range c.Config {
		err := s.init(config)
		if err != nil {
			s.Close()
			return nil, err
		}
	}
	return &s, nil
}

func (s *Somerasult) parseBaseConfig(m map[string]interface{}) (string, string, int) {
	protocol, ok := getStringFromMap(m, "network")
	if !ok {
		protocol = "tcp"
	}

	address, ok := getStringFromMap(m, "address")
	if !ok {
		address = ""
	}

	port, ok := getIntFromMap(m, "port")
	if !ok {
		port = 0
	}
	return protocol, address, port
}

func (s *Somerasult) parseChainConfig(m map[string]interface{}) []pipeline.Config {
	ps, ok := m["pipeline"]
	if !ok {
		return nil
	}
	pcs, ok := ps.([]interface{})
	if !ok {
		return nil
	}
	if len(pcs) == 0 {
		return nil
	}

	chain := make([]pipeline.Config, len(pcs))
	for i, p := range pcs {
		s.logger.Println(reflect.TypeOf(p))
		p, ok := p.(map[string]interface{})
		if !ok {
			return nil
		}
		protocol, _ := getStringFromMap(p, "protocol")
		config, ok := p["config"]
		s.logger.Println(config, ok)
		if !ok {
			return nil
		}
		cs, ok := config.(map[string]interface{})
		s.logger.Println(cs, ok)
		if !ok {
			return nil
		}
		c, err := pipeline.GetPipelineCreator(protocol, cs)
		s.logger.Println(c, err)
		if err != nil {
			return nil
		}
		chain[i] = c
	}
	return chain
}

func (s *Somerasult) init(config map[string]interface{}) error {
	network, address, port := s.parseBaseConfig(config)
	s.logger.Println(network, address, port)
	if network == "" || address == "" || port == 0 {
		return nil
	}

	chain := s.parseChainConfig(config)
	s.logger.Println(chain, config)
	if chain == nil {
		return nil
	}

	addr := fmt.Sprintf("%s:%d", address, port)
	s.logger.Printf("create server listen %s\n", addr)
	listener, err := net.Listen(network, addr)
	if err != nil {
		s.logger.Println(err)
		return err
	}

	var input pipeline.Pipeline
	ctx := context.Background()
	go func() {
		defer s.logger.Printf("close server on %s\n", addr)
		for {
			conn, err := listener.Accept()
			s.logger.Println(conn, err, chain)
			if err != nil {
				continue
			}
			input = conn
			for _, c := range chain {
				s.logger.Println(c)
				input, err = c.New(ctx, input, nil)
				s.logger.Println(c, conn, input, err)
				if err != nil {
					conn.Close()
					continue
				}
			}
		}
	}()
	return nil
}

func (s *Somerasult) Close() {
}
