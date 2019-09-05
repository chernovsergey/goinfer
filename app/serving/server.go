package serving

import (
	"context"
	"log"
	"net"

	pb "github.com/go-code/goinfer/api"
	"google.golang.org/grpc"
)

type Yaml map[interface{}]interface{}

func RunListener(port string) *net.Listener {
	lis, err := net.Listen("tcp", port)
	if err != nil {
		log.Fatalf("Cannot listen %s", port)
	}
	log.Printf("Listening port: %s", port)
	return &lis
}

func RunServer(config Yaml) *grpc.Server {
	grpcServer := grpc.NewServer()
	inferenceServer := NewInferencer(config)
	pb.RegisterInferencerServer(grpcServer, inferenceServer)
	return grpcServer
}

type Inferencer struct {
}

func NewInferencer(config Yaml) *Inferencer {
	path := config["model"].(string)
	loadModel(path)
	return &Inferencer{}
}

func loadModel(path string) {
	// Todo
	// - load model
	// - obtain features and interactions
	// - obtain coefficients
	// - save to fast access structure
}

func (inf *Inferencer) PredictProba(c context.Context,
	req *pb.Request) (*pb.Response, error) {
	return &pb.Response{Proba: 0.0, Confidence: 1.0}, nil
}
