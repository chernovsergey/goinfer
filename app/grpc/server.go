package serving

import (
	"context"
	"log"
	"math"
	"net"

	"github.com/chapsuk/wait"
	pb "github.com/go-code/goinfer/api"
	"google.golang.org/grpc"
)

//revive:disable:exported

func unpackArray(arr []string, vars ...*string) {
	for i, v := range arr {
		*vars[i] = v
	}
}

func typelook(vars []string, types ...*FeatureName) {
	for i, v := range vars {
		*types[i] = featureNameFromString[FeatureNameString(v)]
	}
}

func Sigmoid(x float64) float64 {
	return 1.0 / (1 + math.Exp(-1*x))
}

func Errch(fn func() error) <-chan error {
	ch := make(chan error)
	wg := wait.Group{}
	wg.Add(func() { ch <- fn() })

	go func() {
		wg.Wait()
		close(ch)
	}()
	return ch
}

func RunListener(port string) *net.Listener {
	lis, err := net.Listen("tcp", port)
	if err != nil {
		log.Panicln(err)
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

func Start(ctx context.Context, addr string, config Yaml) error {
	server := RunServer(config)
	listener := RunListener(addr)

	select {
	case <-ctx.Done():
		server.GracefulStop()
		return ctx.Err()
	case err := <-Errch(func() error { return server.Serve(*listener) }):
		return err
	}

}

func (inf *Inferencer) PredictProba(c context.Context,
	req *pb.Request) (*pb.Response, error) {

	var score float64
	for variable := range inf.variables {
		value, err := variable.makeValue(req, &inf.values)
		if err != nil {
			return &pb.Response{}, err
		}
		coef := inf.coef[variable][value]
		score += coef
	}
	return &pb.Response{Proba: Sigmoid(score), Confidence: 1.0}, nil
}
