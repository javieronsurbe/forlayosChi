package main

import (
	"google.golang.org/grpc"
	"log"
	"forlayos/rpc"
	"golang.org/x/net/context"
	"io"
)

const (
	address = "localhost:8080"
)

func main() {
	// Set up a connection to the server.
	conn, err := grpc.Dial(address, grpc.WithInsecure())
	if err != nil {
		log.Fatalf("did not connect: %v", err)
	}
	defer conn.Close()
	log.Println("Connected")
	c := rpc.NewForlayosClient(conn)

	// Contact the server and print out its response.
	stream, err := c.ListForlayos(context.Background(), &rpc.Empty{})
	if err != nil {
		log.Fatalf("could not get forlayos: %v", err)
	}
	for {
		forlayo, err := stream.Recv()
		if err == io.EOF {
			break
		}
		if err != nil {
			log.Fatalf("%v.ListForlayos(_) = _, %v", c, err)
		}
		log.Println(forlayo)
	}

}
