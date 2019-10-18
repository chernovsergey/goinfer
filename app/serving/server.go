package serving

import (
	"context"
	"log"
	"net"

	pb "github.com/go-code/goinfer/api"
	"google.golang.org/grpc"
)

//revive:disable:exported

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

func (inf *Inferencer) PredictProba(c context.Context,
	req *pb.Request) (*pb.Response, error) {

	var score Coefficient
	for variable := range inf.variables {
		value, err := variable.makeValue(req, &inf.values)
		if err != nil {
			return &pb.Response{}, err
		}
		coef := inf.coef[variable][value]
		score += coef
	}
	return &pb.Response{Proba: Sigmoid(float32(score)), Confidence: 1.0}, nil
}
