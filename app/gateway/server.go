package gateway

//revive:disable:exported

import (
	"context"
	"log"
	"net/http"

	pb "github.com/go-code/goinfer/api"
	"github.com/grpc-ecosystem/grpc-gateway/runtime"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"golang.org/x/sync/errgroup"
	"google.golang.org/grpc"
)

func Start(ctx context.Context, addr, endpoint string) error {
	gatewayMux := runtime.NewServeMux()

	opts := []grpc.DialOption{grpc.WithInsecure()}

	if err := pb.RegisterInferencerHandlerFromEndpoint(ctx, gatewayMux, endpoint, opts); err != nil {
		log.Fatal("failed to listen grpc addr", "value", err)
		return err
	}

	mux := http.DefaultServeMux
	mux.Handle("/", gatewayMux)
	mux.Handle("/metrics", promhttp.Handler())

	srv := &http.Server{
		Handler: mux,
		Addr:    addr,
	}

	log.Println("starting grpc gateway server", "address", addr)
	e, _ := errgroup.WithContext(ctx)
	e.Go(func() error {
		return srv.ListenAndServe()
	})

	e.Go(func() error {
		<-ctx.Done()
		return srv.Shutdown(ctx)
	})

	return e.Wait()
}
