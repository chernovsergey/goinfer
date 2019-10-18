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

func sendrequest(req *pb.Request, client pb.InferencerClient) (*pb.Response, error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Millisecond)
	defer cancel()

	resp, err := client.PredictProba(ctx, req)

	return resp, err
}

func main() {
	conn, err := grpc.Dial(adress, grpc.WithInsecure())
	if err != nil {
		log.Fatalf("Cannot connect to %s", adress)
	}
	defer conn.Close()

	client := pb.NewInferencerClient(conn)

	t := time.Now()
	success := 0
	for i := 0; i < 100000; i++ {
		_, err := sendrequest(&pb.Request{
			BannerId:  4054199,
			Geo:       "us",
			ZoneId:    1093182,
			Browser:   8,
			OsVersion: "mac10.12",
		},
			client)
		if err == nil {
			success++
			// fmt.Println(res)
		}
	}

	log.Println("Finished in ", time.Since(t), "Success:", success)
}
