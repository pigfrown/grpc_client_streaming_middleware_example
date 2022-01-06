package main

import (
	"flag"
	"fmt"
	"io"
	"log"
	"net"
	"reflect"

	this "grpc_stream_middleware"

	"google.golang.org/grpc"
)

var (
	port = flag.Int("port", 50051, "The server port")
)

type OurStream struct {
	grpc.ServerStream
}

// This function is weird:
// It gets called when Server calls "Recv".
// However the "m" that is passed is empty.. it gets populated during the call to
// ServerStream.RecvMsg(m).
// After this point, we can modify the message and it'll be picked up by the endpoint
func (s *OurStream) RecvMsg(m interface{}) error {
	fmt.Println("IN RECVMSG!!")

	fmt.Println(reflect.TypeOf(m))
	req, ok := m.(*this.HelloWorldRequest)
	if !ok {
		fmt.Println("RECV MSG NOT OK")
		return nil
	}

	if err := s.ServerStream.RecvMsg(req); err != nil {
		fmt.Println("ERROR WITH RECVMSG")
		return err
	}
	fmt.Println("USER SENT MSG : ", req.GetMessage())
	req.Message = "THIS IS NOT THE MESSAGE THE USER SENT"
	return nil
}

// Small example of modifying the returned message
func (s *OurStream) SendMsg(m interface{}) error {
	req, ok := m.(*this.HelloWorldResponse)
	if !ok {
		return nil
	}
	req.Message = "changing message returned"
	return s.ServerStream.SendMsg(req)
}

//func getMiddleware() grpc.StreamClientInterceptor {
//	return func(ctx context.Context,
//		desc *grpc.StreamDesc,
//		cc *grpc.ClientConn,
//		method string,
//		streamer grpc.Streamer,
//		opts ...grpc.CallOption,
//	) (grpc.ClientStream, error) {
//		clientStream, err := streamer(ctx, desc, cc, method, opts...)
//		return &OurStream{clientStream}, err
//	}
//
//}

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
		fmt.Println("IN ENDPOINT ABOUT TO CALL RECV")
		data, err := in.Recv()
		if err == io.EOF {
			return in.SendAndClose(&this.HelloWorldResponse{Message: "bye"})
		}

		if err != nil {
			return err
		}

		fmt.Printf("IN ENDPOINT WITH DATA %+v\n", data)
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
