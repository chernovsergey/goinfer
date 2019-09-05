package main

import (
	"io/ioutil"
	"log"
	"net"
	"strconv"

	"github.com/go-code/goinfer/app/serving"
	"google.golang.org/grpc"
	"gopkg.in/yaml.v2"
)

const (
	// TODO move to os.Argv
	CONFIGPATH = "./config/prod.yml"
)

func loadConfig(path string) serving.Yaml {

	data, err := ioutil.ReadFile(path)
	if err != nil {
		log.Fatalf("Can't reading config: %v", err)
	}

	m := make(serving.Yaml)
	err = yaml.Unmarshal(data, &m)
	if err != nil {
		log.Fatalf("error: %v", err)
	}

	return m
}

func serve(server *grpc.Server, listener *net.Listener) {
	err := server.Serve(*listener)
	if err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}

func main() {
	config := loadConfig(CONFIGPATH)
	server := serving.RunServer(config)

	port := ":" + strconv.Itoa(config["port"].(int))
	listener := serving.RunListener(port)

	serve(server, listener)
}
