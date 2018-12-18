package echo

import (
	"errors"
	"fmt"
	"github.com/somersault/somersault/pipeline"
	"sync"
)

type Config struct {
	size uint
	maxIdle int
}

type Echo struct {
	*Config
	input pipeline.Pipeline
	buf []byte
	cursor uint
	inputCond sync.Cond
	lock sync.Mutex
}

func (c *Config) New() (pipeline.Pipeline, error) {
	if c.size <= 0 {
		c.size = 1024
	}
	buf := make([]byte, c.size)
	return &Echo{
		c,
		nil,
		buf,
		0,
		sync.Cond{},
		sync.Mutex{},
	}, nil
}

func (e *Echo) cap() uint {
	return e.size - e.cursor
}

func (e *Echo) SetInput(input pipeline.Pipeline) error {
	if e.input != nil {
		return errors.New("duplicate input")
	}
	e.input = input
	e.inputCond.Broadcast()
	return nil
}

func (e *Echo) SetOutput(output pipeline.Pipeline) error {
	return nil
}

func (e *Echo) Read(buf []byte) (int, error) {
	return 0, nil
}

func (e *Echo) Write(buf []byte) (int, error) {
	if e.input == nil {
		e.inputCond.Wait()
	}
	cap := e.cap()
	size := uint(len(buf))
	if size <= cap {
		e.buf[e.cursor:e.cursor+size] = buf
	} else {
		e.buf[e.cursor:e.size] = buf[:cap]
		fmt.Printf(string(buf))
		e.cursor = 0
		e.buf[0:size-cap] = buf[cap:]
	}
	return int(size), nil
}

func (e *Echo) Transport() {
	for {

	}
}

func (e *Echo) Close() error {
	if e.input != nil {
		return e.input.Close()
	}
	return nil
}

func init() {
	config := &Config{
		size: 1024,
	}
	pipeline.RegistePipelineCreator("echo", config)
}