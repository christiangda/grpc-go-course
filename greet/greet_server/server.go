package main

import (
	"context"
	"fmt"
	"io"
	"log"
	"net"
	"strconv"
	"time"

	"github.com/christiangda/grpc-go-course/greet/greetpb"
	"google.golang.org/grpc"
)

type server struct{}

func (s *server) Greet(ctx context.Context, req *greetpb.GreetRequest) (*greetpb.GreetResponse, error) {
	log.Printf("Greet function was invoked with %v", req)
	firstName := req.GetGreeting().GetFirstName()
	lastName := req.GetGreeting().GetLastName()
	result := "Hello " + firstName + " " + lastName
	res := &greetpb.GreetResponse{
		Result: result,
	}
	return res, nil
}

func (s *server) GreetManyTimes(req *greetpb.GreetManyTimesRequest, stream greetpb.GreetService_GreetManyTimesServer) error {
	log.Printf("GreetManyTimes function was invoked with %v", req)
	firstName := req.GetGreeting().GetFirstName()
	lastName := req.GetGreeting().GetLastName()
	for i := 0; i < 10; i++ {
		result := "Hello " + firstName + " " + lastName + " number " + strconv.Itoa(i)
		res := &greetpb.GreetManyTimesResponse{
			Result: result,
		}
		stream.Send(res)
		time.Sleep(1000 * time.Millisecond)
	}
	return nil
}

func (s *server) LongGreet(stream greetpb.GreetService_LongGreetServer) error {
	log.Printf("LongGreet function was invoked with a streaming request")
	result := ""
	for {
		req, err := stream.Recv()
		if err == io.EOF {
			// We have finished reading the client stream
			return stream.SendAndClose(&greetpb.LongGreetResponse{
				Result: result,
			})
		}
		if err != nil {
			log.Fatalf("Error while reading client stream: %v", err)
		}

		firstName := req.GetGreeting().GetFirstName()
		lastName := req.GetGreeting().GetLastName()

		result += "Hello " + firstName + " " + lastName + "! "
	}

	return nil
}

func (s *server) GreetEveryone(stream greetpb.GreetService_GreetEveryoneServer) error {
	log.Printf("GreetEveryone function was invoked a streaming request\n")

	for {
		req, err := stream.Recv()
		if err == io.EOF {
			return nil
		}
		if err != nil {
			log.Fatalf("Error while reading client stream: %v", err)
			return err
		}

		firstName := req.GetGreeting().GetFirstName()
		lastName := req.GetGreeting().GetLastName()

		result := "Hello " + firstName + " " + lastName + "! "

		sendErr := stream.Send(&greetpb.GreetEveryoneResponse{
			Result: result,
		})
		if sendErr != nil {
			log.Fatalf("Error while sending data to client stream: %v", sendErr)
			return sendErr
		}
	}
	return nil
}

func main() {
	fmt.Println("Hello World")

	lis, err := net.Listen("tcp", "0.0.0.0:50051")
	if err != nil {
		log.Fatalf("Failed to listen: %v", err)
	}

	s := grpc.NewServer()
	greetpb.RegisterGreetServiceServer(s, &server{})

	if err := s.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
