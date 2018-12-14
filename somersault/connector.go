package somersault

import (
	"errors"
	"net"
	"sync"
)

type Connector interface {
	Connect(network, address string) (net.Conn, error)
	Close()
}

var (
	connectorMap = make(map[string]func(interface{}) (Connector, error), 0)
	connectorLock = sync.RWMutex{}
)

func RegisteConnector(protocol string, connector Connector) error {
	connectorLock.Lock()
	defer connectorLock.Unlock()
	if _, ok := connectorMap[protocol]; ok {
		return errors.New("duplicate connector")
	}
	connectorMap[protocol] = connector
	return nil
}

func GetConnector(protocol string, config interface{}) (Connector, error) {
	connectorLock.RLock()
	defer connectorLock.RUnlock()
	if f, ok := connectorMap[protocol]; ok {
		connector, err := f(config)
		if err != nil {
			return nil, err
		}
		return connector, nil
	}
	return nil, errors.New("protocol connector not found")
}
