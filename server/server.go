package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
)

type Somersault struct {
	Code int `json:"code"`
	Message string `json:"message"`
	Data map[string]interface{} `json:"data"`
}

type Config struct {
	Address        string `json:"address"`
	Port           int    `json:"port"`
	MaxIdleSession int    `json:"idle_session"`
}

type Server struct {
	*Config
}

func (c *Config) New() (*Server, error) {
	s := &Server{
		c,
	}
	return s, nil
}

func (c *Config) getMaxIdleSession() int {
	if c.MaxIdleSession > 0 {
		return c.MaxIdleSession
	}
	return 100
}

func (s *Server) createSession() (int, error) {
	session := Session{}
	err := session.ListenAndServer()
	if err != nil {
		return 0, err
	}
	return session.Port, nil
}

func (s *Server) somersault(w http.ResponseWriter, req *http.Request) {
	sa := Somersault{}
	defer func() {
		data, err := json.Marshal(sa)
		if err != nil {
			log.Println(err)
			return
		}
		fmt.Println(sa)
		w.Write(data)
	}()

	port, err := s.createSession()
	if err != nil {
		sa.Code = 1
		sa.Message = err.Error()
		return
	}
	sa.Code = 0
	sa.Data = map[string]interface{}{
		"port": port,
	}
}

func (s *Server) ListenAndServer() error {
	http.HandleFunc("/somersault", s.somersault)
	addr := fmt.Sprintf("%s:%d", s.Address, s.Port)
	err := http.ListenAndServe(addr, nil)
	return err
}
