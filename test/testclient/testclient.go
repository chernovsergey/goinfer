package main

import (
	"context"
	"fmt"
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

	t := time.Now()
	success := 0
	errors := make(map[string]bool)
	for i := 0; i < 10000; i++ {
		_, err := client.PredictProba(ctx, &pb.Request{
			BannerId:  4054199,
			Geo:       "us",
			ZoneId:    1093182,
			Browser:   8,
			OsVersion: "mac10.12",
		})
		if err == nil {
			success++
		} else {
			errors[fmt.Sprintf("Failed to get response %v", err)] = true
		}
	}
	log.Println("Finished in ", time.Since(t), "Success:", success)
	log.Println("Errors ", errors)
}
