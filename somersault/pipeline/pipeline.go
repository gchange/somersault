package pipeline

import (
	"context"
	"errors"
	"fmt"
	"net"
	"reflect"
	"strings"
	"sync"
)

var (
	pipelineCreatorMap = make(map[string]Config)
	pipelineLock = sync.RWMutex{}
)

type Error struct {
	errs []error
}

func (e *Error) Error() string {
	if len(e.errs) == 0 {
		return ""
	}
	msgs := make([]string, len(e.errs))
	for i, e := range e.errs {
		msgs[i] = e.Error()
	}
	return strings.Join(msgs, "\t")
}

func (e *Error) IsNil() bool {
	return len(e.errs) == 0
}

func (e *Error) Append(err error) {
	e.errs = append(e.errs, err)
}

type Config interface {
	New() (Pipeline, error)
}

type Pipeline interface {
	SetInput(Pipeline) error
	SetOutput(Pipeline) error
	Read([]byte) (int, error)
	Write([]byte) (int, error)
	Transport()
	Close() error
}

type Transport struct {
	input []Pipeline
	inputWriter chan []byte
	inputLock sync.RWMutex
	output []Pipeline
	outputLock sync.RWMutex
	outputWriter chan []byte
	ctx context.Context
	cancel context.CancelFunc
}

func NewTransport(ctx context.Context) (*Transport, error) {
	ctx, cancel := context.WithCancel(ctx)
	return &Transport{
		input: make([]Pipeline, 0),
		inputWriter:  make(chan []byte),
		inputLock: sync.RWMutex{},
		output: make([]Pipeline, 0),
		outputWriter: make(chan []byte),
		outputLock: sync.RWMutex{},
		ctx: ctx,
		cancel:cancel,
	}, nil
}

func (t *Transport) SetInput(input Pipeline) error {
	t.inputLock.Lock()
	defer t.inputLock.Unlock()
	t.input = append(t.input, input)
	go func() {
		defer t.DeleteInput(input)
		t.transport(input, t.inputWriter)
	}()
	return nil
}

func (t *Transport) DeleteInput(input Pipeline) error {
	t.inputLock.Lock()
	defer t.inputLock.Unlock()
	for i, v := range t.input {
		if v == input {
			t.input = append(t.input[:i], t.input[i+1:]...)
			break
		}
	}
	if len(t.input) == 0 {
		close(t.inputWriter)
	}
}

func (t *Transport) SetOutput(output Pipeline) error {
	t.outputLock.Lock()
	defer t.outputLock.Unlock()
	t.output = append(t.output, output)
	go func() {
		defer t.DeleteOutput(output)
		t.transport(output, t.outputWriter)
	}()
	return nil
}

func (t *Transport) DeleteOutput(output Pipeline) error {
	t.outputLock.Lock()
	defer t.outputLock.Unlock()
	for i, v := range t.output {
		if v == output {
			t.output = append(t.output[:i], t.output[i+1:]...)
			break
		}
	}
	if len(t.output) == 0 {
		close(t.outputWriter)
	}
}

func (t *Transport) CloseOutput() error {
	var err Error
	t.outputLock.Lock()
	defer t.outputLock.Unlock()
	for _, output := range t.output {
		if e := output.Close(); e != nil {
			err.Append(e)
		}
	}
	if err.IsNil() {
		return nil
	}
	return &err
}

func (t *Transport) Read(buf []byte) (int, error) {
	t.inputLock.RLock()
	input := t.input
	t.inputLock.RUnlock()

	if len(input) == 0 {
		t.inputCond.Wait()
		return t.Read(buf)
	}
	net.Conn

	return t.read(input, buf)
}

func (t *Transport) convertToSelectCase(pipelines []Pipeline) []reflect.SelectCase {
	cases := make([]reflect.SelectCase, len(pipelines))
	for i, p := range pipelines {
		cases[i] = reflect.SelectCase{
			Dir: reflect.SelectRecv,
			Chan: reflect.ValueOf(p),
		}
	}
	return cases
}

func (t *Transport) read(input []reflect.SelectCase, buf []byte) (int, error) {
	if len(input) == 0 {
		return 0, nil
	}
	for {
		reflect.Select(input)
	}
}

func (t *Transport) Write(buf []byte) (int, error) {
}

func (t *Transport) write(buf []byte) (int, error) {
}

func (t *Transport) transport(pipeline Pipeline, writer chan []byte) {
	buf := make([]byte, 1024)
	for {
		select {
		case <- t.ctx.Done():
			break
		default:
			n, err := pipeline.Read(buf)
			if err != nil {
				break
			}
			if n == 0 {
				continue
			}
			select {
			case writer <- buf[:n]:
				continue
			case <- t.ctx.Done():
				break
			}
		}
	}
}

func (t *Transport) Transport() {
}

func (t *Transport) Close() error {
}

func RegistePipelineCreator(name string, config Config) error {
	pipelineLock.Lock()
	defer pipelineLock.Unlock()
	if _, ok := pipelineCreatorMap[name]; ok {
		return fmt.Errorf("duplicate parser named %s", name)
	}
	pipelineCreatorMap[name] = config
	return nil
}

func GetPipelineCreator(name string, config map[string]interface{}) (Pipeline, error) {
	pipelineLock.RLock()
	defer pipelineLock.RUnlock()
	if c, ok := pipelineCreatorMap[name]; ok {
		nc := c
		v := reflect.ValueOf(nc).Elem()
		for name, val := range config {
			field := v.FieldByName(name)
			if !field.CanSet() {
				continue
			}
			v := reflect.ValueOf(val)
			if v.Type() != field.Type() {
				return nil, errors.New("wrong type config")
			}
			field.Set(v)
		}
		return nc.New()
	}
	return nil, errors.New("parser creator not found")
}

