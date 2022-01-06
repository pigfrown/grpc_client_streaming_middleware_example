package main

import (
	"flag"
	"fmt"
	"io"
	"log"
	"net"

	this "grpc_stream_middleware"

	"google.golang.org/grpc"
)

var (
	port = flag.Int("port", 50051, "The server port")
)

type OurStream struct {
	grpc.ServerStream
}

func (s *OurStream) RecvMsg(m interface{}) error {
	fmt.Println("IN RECVMSG!!")

	req, ok := m.(*this.HelloWorldRequest)
	if !ok {
		fmt.Println("NOT OK")
		return nil
	}

	// Trying to change Message here but it's not picked up in HelloWorld?
	fmt.Println("AER WE HERE?")
	fmt.Println(req.Message)
	req.Message = "poo"
	fmt.Println(req.Message)
	fmt.Println("YES")
	if err := s.ServerStream.RecvMsg(req); err != nil {
		return err
	}
	return nil
}

func getMiddleware() func(srv interface{}, stream grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
	return func(srv interface{}, stream grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
		fmt.Println("IN MIDDLEWARE")
		wrapper := &OurStream{ServerStream: stream}
		return handler(srv, wrapper)
	}

}

type server struct {
	this.UnimplementedTestServiceServer
}

// SayHello implements helloworld.GreeterServer
func (s *server) HelloWorld(in this.TestService_HelloWorldServer) error {
	for {
		fmt.Println("ABOUT TO CALL RECV")
		data, err := in.Recv()
		if err == io.EOF {
			return in.SendAndClose(&this.HelloWorldResponse{Message: "bye"})
		}

		if err != nil {
			return err
		}

		fmt.Printf("HERE WITH DATA %v\n", data)
	}
}

func main() {
	flag.Parse()
	lis, err := net.Listen("tcp", fmt.Sprintf("localhost:%d", *port))
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	s := grpc.NewServer(
		grpc.StreamInterceptor(getMiddleware()),
	)
	this.RegisterTestServiceServer(s, &server{})
	log.Printf("server listening at %v", lis.Addr())
	if err := s.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
