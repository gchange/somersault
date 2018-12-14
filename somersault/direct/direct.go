package direct

import (
	"errors"
	"github.com/somersault/somersault/pipe"
	"sync"
)

type Config struct {
	src pipe.Pipe
	dst pipe.Pipe
}

type TCP struct {
	*Config
}

func NewTCP(config interface{}) (pipe.Pipe, error) {
	if config, ok := config.(Config); ok {
		t := TCP{
			&config,
		}
		go t.pipe()
		return &t, nil
	}
	return nil, errors.New("unknown config")
}

func (t *TCP) pipe() {
	var wg sync.WaitGroup
	f := func(src pipe.Pipe, dst pipe.Pipe) {
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
	f(t.src, t.dst)
	f(t.dst, t.src)
	wg.Wait()
}

func (t *TCP) Read(buf []byte) (int, error) {
	return t.src.Read(buf)
}

func (t *TCP) Write(buf []byte) (int, error) {
	return t.src.Write(buf)
}

func (t *TCP) Close() error {
	var srcErr, dstErr error
	if t.src != nil {
		srcErr = t.src.Close()
	}
	if t.dst != nil {
		dstErr = t.dst.Close()
	}
	if srcErr != nil {
		return srcErr
	}
	return dstErr
}

func init() {
	pipe.RegistePipeCreator("tcp", NewTCP)
}
