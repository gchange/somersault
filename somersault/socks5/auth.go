package socks5

import (
	"net"
	"sync"
)

var (
	authMethods = map[uint8]func(conn net.Conn) error {}
	authReplyMethods = map[uint8]func(conn net.Conn) error{}
	authMethodLock = sync.RWMutex{}
)

func registeAuthMethod(method uint8, f func(conn net.Conn) error) error {
	authMethodLock.Lock()
	defer authMethodLock.Unlock()
	if _, ok := authMethods[method]; ok {
		return DuplicateAuthMethod
	}
	authMethods[method] = f
	return nil
}

func getAuthMethod(method uint8) func(conn net.Conn) error {
	authMethodLock.RLock()
	defer authMethodLock.RUnlock()
	if f, ok := authMethods[method]; ok {
		return f
	}
	return nil
}

func getSupportAuthMethod() []uint8 {
	authMethodLock.RLock()
	defer authMethodLock.RUnlock()
	methods := make([]uint8, len(authMethods))
	for method := range authMethods {
		methods = append(methods, method)
	}
	return methods
}

func registeAuthReplyMethod(method uint8, f func(conn net.Conn) error) error {
	authMethodLock.Lock()
	defer authMethodLock.Unlock()
	if _, ok := authReplyMethods[method]; ok {
		return DuplicateAuthMethod
	}
	authReplyMethods[method] = f
	return nil
}

func getAuthReplyMethod(method uint8) func(conn net.Conn) error {
	authMethodLock.RLock()
	defer authMethodLock.RUnlock()
	if f, ok := authReplyMethods[method]; ok {
		return f
	}
	return nil
}

func noAuth(_ net.Conn) error {
	return nil
}

func init() {
	registeAuthMethod(0, noAuth)
	registeAuthReplyMethod(0, noAuth)
}
