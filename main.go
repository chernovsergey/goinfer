package main

import (
	"context"
	"io/ioutil"
	"log"

	gateway "github.com/go-code/goinfer/app/gateway"
	serving "github.com/go-code/goinfer/app/grpc"
	"gopkg.in/yaml.v2"

	_ "net/http/pprof"
)

const (
	CONFIGPATH = "./config/prod.yml"

	// TODO read this params from config
	GRPCP_PORT   = "50077"
	GATEWAY_PORT = "8080"
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

func main() {

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	config := loadConfig(CONFIGPATH)

	addrGRPC := ":" + GRPCP_PORT
	addrGateway := ":" + GATEWAY_PORT

	errGateway := serving.Errch(func() error { return gateway.Start(ctx, addrGateway, addrGRPC) })
	errGRPC := serving.Errch(func() error { return serving.Start(ctx, addrGRPC, config) })

	select {
	case reason := <-errGRPC:
		log.Println("grpc server is down", "reason", reason)
		cancel()
	case reason := <-errGateway:
		log.Println("gateway server is down", "reason", reason)
		cancel()
	case <-ctx.Done():
		log.Println("context is canceled", "reason", ctx.Err())
	}

	// wait grpc and gateway
	<-ctx.Done()
}
