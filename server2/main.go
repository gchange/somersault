package main

import (
	"encoding/json"
	"flag"
	"github.com/somersault/somersault"
	"io/ioutil"
	"log"
	"os"
	"os/signal"
	"syscall"
)

func main() {
	flag.Parse()
	fileName := flag.String("config", "config.json", "the json config")
	buf, err := ioutil.ReadFile(*fileName)
	if err != nil {
		log.Fatal(err)
	}

	var config somersault.Config
	err = json.Unmarshal(buf, &config)
	if err != nil {
		log.Fatal(err)
	}

	logger := log.New(os.Stdout, "somersault", log.Lshortfile|log.LstdFlags)
	srv, err := config.New(*logger)
	if err != nil {
		log.Fatal(err)
	}
	defer srv.Close()

	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM)
	<-sc
}
