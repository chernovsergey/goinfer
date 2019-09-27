package serving

import (
	"bufio"
	"context"
	"log"
	"net"
	"os"
	"strconv"
	"strings"

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
	features     map[string]bool
	interactions map[string]bool
	coef         map[string]float64
}

func NewInferencer(config Yaml) *Inferencer {
	path := config["model"].(string)
	obj := Inferencer{
		features:     make(map[string]bool),
		interactions: make(map[string]bool),
		coef:         make(map[string]float64),
	}
	obj.loadModel(path)
	return &obj
}

func (inf *Inferencer) loadModel(path string) {
	// Todo
	// - load model
	// - obtain features and interactions
	// - obtain coefficients
	// - save to fast access structure
	file, err := os.Open(path)
	if err != nil {
		log.Fatalf("Cant open model file: %s", path)
	}
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		record := scanner.Text()
		tokens := strings.Split(record, ":")

		// parse featurename=featurevalue format
		pair := strings.Split(tokens[1], "=")
		name, _ := pair[0], pair[1]

		if !strings.Contains(tokens[1], "XX") {
			inf.features[name] = true
		} else {

			inf.interactions[name] = true
		}

		coef, err := strconv.ParseFloat(tokens[2], 64)
		if err != nil {
			log.Fatalf("Cant parse coefficient %s=%s", tokens[1], tokens[2])
		}
		inf.coef[tokens[1]] = coef
	}
	log.Printf("Model have load.\nFeatures: %v\nInteractions: %v\nOveral coefficient count: %d",
		inf.features, inf.interactions, len(inf.coef))
}

func (inf *Inferencer) PredictProba(c context.Context,
	req *pb.Request) (*pb.Response, error) {
	log.Println("Serving request: ", req)
	return &pb.Response{Proba: 0.0, Confidence: 1.0}, nil
}
