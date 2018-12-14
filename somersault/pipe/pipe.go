package pipe

import (
	"errors"
	"fmt"
	"sync"
)

var (
	pipeCreatorMap = make(map[string]func(interface{})(Pipe, error))
	pipeLock = sync.RWMutex{}
)

type Pipe interface {
	Read([]byte) (int, error)
	Write([]byte) (int, error)
	Close() error
}

func RegistePipeCreator(name string, creator func(interface{})(Pipe, error)) error {
	pipeLock.Lock()
	defer pipeLock.Unlock()
	if _, ok := pipeCreatorMap[name]; ok {
		return fmt.Errorf("duplicate parser named %s", name)
	}
	pipeCreatorMap[name] = creator
	return nil
}

func GetParser(name string, config interface{}) (Pipe, error) {
	pipeLock.RLock()
	defer pipeLock.RUnlock()
	if creator, ok := pipeCreatorMap[name]; ok {
		return creator(config)
	}
	return nil, errors.New("parser creator not found")
}

