package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
)

func main() {
	flag.Parse()
	fileName := flag.String("config", "config.json", "the json config")
	buf, err := ioutil.ReadFile(*fileName)
	if err != nil {
		panic(err)
	}

	var config Config
	err = json.Unmarshal(buf, &config)
	if err != nil {
		panic(err)
	}

	cli, err := config.New()
	if err != nil {
		panic(err)
	}
	fmt.Println(cli)

	fmt.Println(cli.Send([]byte("http://baidu.com")))
}

