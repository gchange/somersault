package direct

import (
	"context"
	"errors"
	"fmt"
	"net"

	"github.com/gchange/somersault/somersault/pipeline"
)

type Config struct {
	Network string `somersault:"network"`
	Address string `somersault:"address"`
	Port    int    `somersault:"port"`
}

type TCP struct {
	*Config
	*pipeline.DefaultPipeline
}

func (c *Config) DeepCopy() pipeline.Config {
	return &Config{
		Network: c.Network,
		Address: c.Address,
		Port:    c.Port,
	}
}

func (c *Config) New(ctx context.Context, input, output pipeline.Pipeline) (pipeline.Pipeline, error) {
	if c.Network == "" || c.Port == 0 {
		fmt.Println(c)
		return nil, errors.New("remote address format error")
	}
	addr := fmt.Sprintf("%s:%d", c.Address, c.Port)
	conn, err := net.Dial(c.Network, addr)
	fmt.Println(conn, err)
	if err != nil {
		return nil, err
	}

	dp, err := pipeline.NewDefaultPipeline(ctx, input, conn)
	if err != nil {
		return nil, err
	}

	t := &TCP{
		c,
		dp,
	}
	go t.Transport()
	return t, nil
}

func init() {
	config := &Config{
		Network: "tcp",
		Address: "0.0.0.0",
		Port:    0,
	}
	pipeline.RegistePipelineCreator("tcp", config)
}
