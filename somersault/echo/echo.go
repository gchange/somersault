package echo

import (
	"context"
	"fmt"
	"github.com/somersault/somersault/pipeline"
	"io"
	"os"
)

type Config struct {
}

type echo struct {
	reader chan error
}

type Echo struct {
	*Config
	*pipeline.DefaultPipeline
}

func (c *Config) New(ctx context.Context, input, output pipeline.Pipeline) (pipeline.Pipeline, error) {
 	dp, err := pipeline.NewDefaultPipeline(ctx, input, &echo{make(chan error)})
 	fmt.Println(dp, err)
 	if err != nil {
 		return nil, err
	}
 	e := &Echo{
 		c,
 		dp,
	}
 	go e.Transport()
 	return e, nil
}

func (e *echo) Read(buf []byte) (int, error) {
	err := <- e.reader
	return 0, err
}

func (e *echo) Write(buf []byte) (int, error) {
	return os.Stdout.Write(buf)
}

func (e *echo) Close() error {
	e.reader <- io.EOF
	return nil
}

func init() {
	config := &Config{}
	pipeline.RegistePipelineCreator("echo", config)
}