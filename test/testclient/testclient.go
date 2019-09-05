package main

import (
	"context"
	"log"
	"time"

	pb "github.com/go-code/goinfer/api"
	"google.golang.org/grpc"
)

const (
	adress = "localhost:50077"
)

func main() {
	conn, err := grpc.Dial(adress, grpc.WithInsecure())
	if err != nil {
		log.Fatalf("Cannot connect to %s", adress)
	}
	defer conn.Close()

	client := pb.NewInferencerClient(conn)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	for i := 0; i < 10; i++ {
		res, err := client.PredictProba(ctx, &pb.Request{
			BannerId: 1,
			Geo:      5,
			ZoneId:   100500,
			Platform: 4,
		})
		if err != nil {
			log.Fatalf("Failed to get responce %v", err)
		}
		log.Println(res.Proba, res.Confidence)
	}
}
