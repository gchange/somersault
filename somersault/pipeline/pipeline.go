package pipeline

import (
	"context"
	"errors"
	"fmt"
	"io"
	"reflect"
	"strconv"
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
	New(ctx context.Context, input, output Pipeline) (Pipeline, error)
}

type Pipeline interface {
	Read(buf []byte) (int, error)
	Write(buf []byte) (int, error)
	Close() error
}

type DefaultPipeline struct {
	input Pipeline
	output Pipeline
	ctx context.Context
	cancel context.CancelFunc
}

func NewDefaultPipeline(ctx context.Context, input, output Pipeline) (*DefaultPipeline, error) {
	fmt.Println(ctx, input, output)
	ctx, cancel := context.WithCancel(ctx)
	c := func() {
		if input != nil {
			input.Close()
		}
		if output != nil {
			output.Close()
		}
		cancel()
	}
	dp := &DefaultPipeline{
		input: input,
		output: output,
		ctx: ctx,
		cancel: c,
	}
	return dp, nil
}

func (dp *DefaultPipeline) Read(buf []byte) (int, error) {
	if dp.output == nil {
		return 0, io.EOF
	}
	return dp.output.Read(buf)
}

func (dp *DefaultPipeline) Write(buf []byte) (int, error) {
	if dp.input == nil {
		return 0, io.EOF
	}
	return dp.input.Write(buf)
}

func (dp *DefaultPipeline) Transport() {
	defer dp.Close()
	if dp.input == nil || dp.output == nil {
		return
	}

	wg := sync.WaitGroup{}
	transport := func(reader Pipeline, writer Pipeline) {
		defer writer.Close()
		defer wg.Done()
		for {
			buf := make([]byte, 1024)
			n, err := reader.Read(buf)
			if err != nil {
				fmt.Println(reader, writer, err)
				return
			}
			if n != 0 {
				writer.Write(buf[:n])
			}
		}
	}

	wg.Add(2)
	go transport(dp.input, dp.output)
	go transport(dp.output, dp.input)
	wg.Wait()
	fmt.Println("stop transport")
}

func (dp *DefaultPipeline) Close() error {
	err := &Error{}
	if dp.input != nil {
		if e := dp.input.Close(); e != nil {
			err.Append(e)
		}
	}
	if dp.output != nil {
		if e := dp.output.Close(); e != nil {
			err.Append(e)
		}
	}
	if err.IsNil() {
		return nil
	}
	return err
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

func parseInt64(val reflect.Value) (int64, error) {
	switch val.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return val.Int(), nil
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return int64(val.Uint()), nil
	case reflect.Float32, reflect.Float64:
		return int64(val.Float()), nil
	case reflect.String:
		return strconv.ParseInt(val.String(), 10, 64)
	case reflect.Bool:
		if val.Bool() {
			return 1, nil
		} else {
			return 0, nil
		}
	default:
		return 0, errors.New("syntax error")
	}
}

func parseUint64(val reflect.Value) (uint64, error) {
	switch val.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return uint64(val.Int()), nil
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return val.Uint(), nil
	case reflect.Float32, reflect.Float64:
		return uint64(val.Float()), nil
	case reflect.String:
		return strconv.ParseUint(val.String(), 10, 64)
	case reflect.Bool:
		if val.Bool() {
			return 1, nil
		} else {
			return 0, nil
		}
	default:
		return 0, errors.New("syntax error")
	}
}

func parseFloat64(val reflect.Value) (float64, error) {
	switch val.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return float64(val.Int()), nil
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return float64(val.Uint()), nil
	case reflect.Float32, reflect.Float64:
		return val.Float(), nil
	case reflect.String:
		return strconv.ParseFloat(val.String(), 64)
	case reflect.Bool:
		if val.Bool() {
			return 1, nil
		} else {
			return 0, nil
		}
	default:
		return 0, errors.New("syntax error")
	}
}

func parseString(val reflect.Value) (string, error) {
	switch val.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return strconv.FormatInt(val.Int(), 10), nil
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return strconv.FormatUint(val.Uint(), 10), nil
	case reflect.Float32, reflect.Float64:
		return strconv.FormatFloat(val.Float(), 'f', 6, 64), nil
	case reflect.String:
		return val.String(), nil
	case reflect.Bool:
		if val.Bool() {
			return "true", nil
		} else {
			return "false", nil
		}
	default:
		return "", errors.New("syntax error")
	}
}

func parseBool(val reflect.Value) (bool, error) {
	switch val.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return val.Int() > 0, nil
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return val.Uint() > 0, nil
	case reflect.Float32, reflect.Float64:
		return val.Float() > 0, nil
	case reflect.String:
		return val.String() == "", nil
	case reflect.Bool:
		return val.Bool(), nil
	default:
		return false, errors.New("syntax error")
	}
}

func GetPipelineCreator(name string, config map[string]interface{}) (Config, error) {
	pipelineLock.RLock()
	defer pipelineLock.RUnlock()
	fmt.Println("pipeconfig", config)
	if c, ok := pipelineCreatorMap[name]; ok {
		nc := c // copy default config
		v := reflect.Indirect(reflect.ValueOf(nc))
		t := v.Type()
		for i := 0; i < v.NumField(); i++ {
			tf := t.Field(i)
			vf := v.Field(i)
			if !vf.CanSet() {
				continue
			}

			key := ""
			if tag := tf.Tag.Get("somersault"); tag != "" {
				key = tag
			} else {
				key = strings.ToLower(t.Name())
			}

			if val, ok := config[key]; ok {
				m := reflect.ValueOf(val)
				if m.Kind() == vf.Kind() {
					vf.Set(m)
					continue
				}

				switch vf.Kind() {
				case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
					mv, err := parseInt64(m)
					if err != nil {
						return nil, err
					}
					vf.SetInt(mv)
				case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
					mv, err := parseUint64(m)
					if err != nil {
						return nil, err
					}
					vf.SetUint(mv)
				case reflect.Float32, reflect.Float64:
					mv, err := parseFloat64(m)
					if err != nil {
						return nil, err
					}
					vf.SetFloat(mv)
				case reflect.String:
					mv, err := parseString(m)
					if err != nil {
						return nil, err
					}
					vf.SetString(mv)
				case reflect.Bool:
					mv, err := parseBool(m)
					if err != nil {
						return nil, err
					}
					vf.SetBool(mv)
				default:
					return nil, errors.New("wrong type config")
				}
			}
		}
		return nc, nil
	}
	return nil, errors.New("parser creator not found")
}
