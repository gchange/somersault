package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"net/http"
	"reflect"
)

type Somersault struct {
	Code int `json:"code"`
	Message string `json:"message"`
	Data map[string]interface{} `json:"data"`
}

type Config struct {
	Address string `json:"address"`
	Port int `json:"port"`
	Path string `json:path`
}

type Client struct {
	*Config
	conn net.Conn
}

func (config *Config) New() (*Client, error){
	client := &Client{
		config,
		nil,
	}
	return client, nil
}

func (c *Client) CreateSession() error {
	if c.conn != nil {
		return nil
	}

	uri := fmt.Sprintf("http://%s:%d/%s", c.Address, c.Port, c.Path)
	fmt.Println(uri)
	resp, err := http.Get(uri)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	fmt.Printf("%v\n", resp)

	sa := Somersault{}
	decoder := json.NewDecoder(resp.Body)
	err = decoder.Decode(&sa)
	if err != nil {
		return err
	}
	if sa.Code != 0 {
		return errors.New(sa.Message)
	}

	pd, ok := sa.Data["port"]
	fmt.Println(pd, ok, reflect.TypeOf(pd))
	if !ok {
		return errors.New("invalid port")
	}
	port, ok := pd.(float64)
	fmt.Println(port, ok, reflect.TypeOf(port))
	if !ok {
		return errors.New("invalid port")
	}

	addr := fmt.Sprintf("%s:%d", c.Address, int(port))
	conn, err := net.Dial("tcp", addr)
	c.conn = conn
	return err
}

func (c *Client) Send(buf []byte) error {
	if c.conn == nil {
		err := c.CreateSession()
		if err != nil {
			return err
		}
	}

	_, err := c.conn.Write(buf)
	fmt.Println(string(buf), err)
	if err != nil {
		return err
	}
	var b [1024]byte
	for {
		n, err := c.conn.Read(b[0:])
		if err != nil {
			break
		}
		fmt.Println(string(b[0:n]))
	}
	return nil
}
