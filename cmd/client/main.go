package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"time"

	this "grpc_stream_middleware"

	"google.golang.org/grpc"
)

const (
	defaultName = "world"
)

var (
	addr = flag.String("addr", "localhost:50051", "the address to connect to")
	name = flag.String("name", defaultName, "Name to greet")
)

func main() {
	flag.Parse()
	// Set up a connection to the server.
	conn, err := grpc.Dial(*addr, grpc.WithInsecure())
	if err != nil {
		log.Fatalf("did not connect: %v", err)
	}
	defer conn.Close()
	c := this.NewTestServiceClient(conn)

	// Contact the server and print out its response.
	ctx := context.Background()
	//defer cancel()
	stream, err := c.HelloWorld(ctx)
	defer func() {
		stream.CloseSend()
	}()

	if err != nil {
		log.Fatalf("could not greet: %v", err)
	}
	for i := 0; i < 2; i++ {

		fmt.Println("SENDING ", i)
		err = stream.Send(&this.HelloWorldRequest{Message: "hihihih"})
		fmt.Println(err)
		if err != nil {
			log.Fatalf("stream err : %v", err)
		}

		time.Sleep(1 * time.Second)
	}

	res, err := stream.CloseAndRecv()
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println(res)

}
