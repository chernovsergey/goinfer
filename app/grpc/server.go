package serving

import (
	"context"
	"log"
	"net"

	pb "github.com/go-code/goinfer/api"
	grpc_prometheus "github.com/grpc-ecosystem/go-grpc-prometheus"
	"google.golang.org/grpc"
)

// RunListener produces listener of given port
func RunListener(port string) *net.Listener {
	lis, err := net.Listen("tcp", port)
	if err != nil {
		log.Panicln(err)
		log.Fatalf("Cannot listen %s", port)
	}
	log.Printf("Listening port: %s", port)
	return &lis
}

// Start function runs grpc service with exporting prometheus service
func Start(ctx context.Context, addr string, config Yaml) error {

	listener := RunListener(addr)

	server := grpc.NewServer(
		grpc.UnaryInterceptor(grpc_prometheus.UnaryServerInterceptor),
	)
	myservice := NewInferencer(config)
	pb.RegisterInferencerServer(server, myservice)

	grpc_prometheus.EnableHandlingTimeHistogram(
		grpc_prometheus.WithHistogramBuckets([]float64{
			.001, .005, .01, .025, .05, .1,
		}),
	)
	grpc_prometheus.Register(server)

	select {
	case <-ctx.Done():
		server.GracefulStop()
		return ctx.Err()
	case err := <-Errch(func() error { return server.Serve(*listener) }):
		return err
	}
}
