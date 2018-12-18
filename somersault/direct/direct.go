package direct

import (
	"errors"
	"github.com/somersault/somersault/pipeline"
	"sync"
)

type Config struct {
}

type TCP struct {
	*Config
	input pipeline.Pipeline
	output pipeline.Pipeline
}

func (c *Config) New() (pipeline.Pipeline, error) {
	t := TCP{
		c,
		nil,
		nil,
	}
	go t.transport()
}

func (t *TCP) transport() {
	var wg sync.WaitGroup
	f := func(src pipeline.Pipeline, dst pipeline.Pipeline) {
		defer dst.Close()
		defer wg.Done()
		for {
			buf := make([]byte, 1024)
			n, err := src.Read(buf)
			if err != nil {
				break
			}
			_, err = dst.Write(buf[0:n])
			if err != nil {
				src.Close()
			}
		}
	}

	wg.Add(2)
	f(t.input, t.output)
	f(t.output, t.output)
	wg.Wait()
}

func (t *TCP) SetInput(input pipeline.Pipeline) error {
	if t.input != nil {
		return errors.New("duplicate input")
	}
	t.input = input
	return nil
}

func (t *TCP) SetOutput(output pipeline.Pipeline) error {
	if t.output != nil {
		return errors.New("duplicate output")
	}
	t.output = output
	return nil
}

func (t *TCP) Read(buf []byte) (int, error) {
	return t.input.Read(buf)
}

func (t *TCP) Write(buf []byte) (int, error) {
	return t.input.Write(buf)
}

func (t *TCP) Close() error {
	var srcErr, dstErr error
	if t.input != nil {
		srcErr = t.input.Close()
	}
	if t.output != nil {
		dstErr = t.output.Close()
	}
	if srcErr != nil {
		return srcErr
	}
	return dstErr
}

func init() {
	config := &Config{}
	pipeline.RegistePipelineCreator("tcp", config)
}
