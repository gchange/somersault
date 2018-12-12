package main

import (
	"bufio"
	"fmt"
	"github.com/astaxie/beego/logs"
	"io"
	"log"
	"net"
	"net/http"
)

type Session struct {
	Port     int
	Listener net.Listener
}

func (s *Session) handleConnection(conn net.Conn) {
	defer conn.Close()

	for {
		data := make([]byte, 1024)
		n, err := bufio.NewReader(conn).Read(data)
		if err != nil {
			logs.Debug(err)
			break
		}

		uri := string(data[0:n])
		fmt.Println(uri)
		resp, err := http.Get(uri)
		if err != nil {
			conn.Write([]byte(err.Error()))
			break
		}
		io.Copy(conn, resp.Body)
		resp.Body.Close()
	}
}

func (s *Session) ListenAndServer() error {
	listener, err := net.Listen("tcp", ":0")
	if err != nil {
		return err
	}

	s.Port = listener.Addr().(*net.TCPAddr).Port
	s.Listener = listener
	go func() {
		for {
			conn, err := listener.Accept()
			fmt.Println(conn, err)
			if err != nil {
				log.Println(err)
				continue
			}
			go s.handleConnection(conn)
		}
	}()
	return nil
}

func (s *Session) Close() error {
	return nil
}
