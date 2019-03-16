package main

import (
	"context"
	"fmt"
	"io"
	"log"
	"time"

	"google.golang.org/grpc/codes"

	"github.com/christiangda/grpc-go-course/calculator/calculatorpb"
	"google.golang.org/grpc"
	"google.golang.org/grpc/status"
)

func main() {
	fmt.Println("Calculator Client")
	cc, err := grpc.Dial("localhost:50051", grpc.WithInsecure())
	if err != nil {
		log.Fatalf("Could not connect to: %v", err)
	}
	defer cc.Close()

	c := calculatorpb.NewCalculatorServiceClient(cc)
	//fmt.Println("Created client: %f", c)
	//doUnary(c)

	//doServerStreaming(c)

	//doClientStreaming(c)

	//doBiDiStreaming(c)

	doErrorUnary(c)
}

func doUnary(c calculatorpb.CalculatorServiceClient) {
	fmt.Println("Starting to do a Sum unary RPC...")
	req := &calculatorpb.SumRequest{
		FirstNumber:  3,
		SecondNumber: 4,
	}
	res, err := c.Sum(context.Background(), req)
	if err != nil {
		log.Fatalf("error while calling Sum RPC: %v", err)
	}
	log.Printf("Response from Sum: %v", res.SumResult)
}

func doServerStreaming(c calculatorpb.CalculatorServiceClient) {
	fmt.Println("Starting to do a PrimeDecomposition Server Streaming RPC...")
	req := &calculatorpb.PrimeNumberDecompositionRequest{
		Number: 12390392840,
	}
	stream, err := c.PrimeNumberDecomposition(context.Background(), req)
	if err != nil {
		log.Fatalf("error while calling PrimeDecomposition RPC: %v", err)
	}
	for {
		res, err := stream.Recv()
		if err == io.EOF {
			break
		}
		if err != nil {
			log.Fatalf("Someting is happened: %v", err)
		}
		log.Printf("Response from PrimeDecomposition: %v", res.GetPrimeFactor())
	}
}

func doClientStreaming(c calculatorpb.CalculatorServiceClient) {
	fmt.Println("Starting to do a ComputeAverage Client Streaming RPC...")

	stream, err := c.ComputeAverage(context.Background())
	if err != nil {
		log.Fatalf("Error while opening stream: %v", err)
	}

	numbers := []int32{3, 5, 9, 54, 33}

	for _, number := range numbers {
		fmt.Printf("Sending number: %v\n", number)
		stream.Send(&calculatorpb.ComputeAverageRequest{
			Number: number,
		})
		time.Sleep(1000 * time.Millisecond)
	}

	resp, err := stream.CloseAndRecv()
	if err != nil {
		log.Fatalf("Error while receiving response: %v", err)
	}
	fmt.Printf("The Average is: %v\n", resp.GetAverage())
}

func doBiDiStreaming(c calculatorpb.CalculatorServiceClient) {
	fmt.Println("Starting to do a FindMaximum BiDi Streaming RPC...")

	stream, err := c.FindMaximum(context.Background())
	if err != nil {
		log.Fatalf("Error while opening stream and call FindMaximum: %v", err)
	}

	waitc := make(chan struct{})

	// send go routine
	go func() {
		numbers := []int32{4, 1, 34, 5, 68, 44, 45, 70, 23}
		for _, number := range numbers {
			fmt.Printf("Sending number: %v\n", number)
			stream.Send(&calculatorpb.FindMaximumRequest{
				Number: number,
			})
			time.Sleep(1000 * time.Millisecond)
		}
		stream.CloseSend()
	}()

	// receive go routine
	go func() {
		for {
			res, err := stream.Recv()
			if err == io.EOF {
				break
			}
			if err != nil {
				log.Fatalf("Error while reading sever stream: %v\n", err)
				break
			}
			maximum := res.GetMaximum()
			fmt.Printf("Received new maximum of...: %v\n", maximum)
		}
		close(waitc)
	}()
	<-waitc
}

func doErrorUnary(c calculatorpb.CalculatorServiceClient) {
	fmt.Println("Starting to do a Sum unary RPC...")

	// correct call
	doErrorCall(c, 10)
	// error call
	doErrorCall(c, -10)
}

func doErrorCall(c calculatorpb.CalculatorServiceClient, n int32) {
	fmt.Println("Starting to do a Sum unary RPC...")

	// correct call
	res, err := c.SquareRoot(context.Background(), &calculatorpb.SquareRootRequest{Number: n})
	if err != nil {
		respErr, ok := status.FromError(err)
		if ok {
			fmt.Printf("Error message from server: %v\n", respErr.Message())
			fmt.Println(respErr.Code())
			if respErr.Code() == codes.InvalidArgument {
				fmt.Println("You probably sent a negative number!")
				return
			}
		} else {
			log.Fatalf("Big error calling SquareRoot: %v\n", err)
			return
		}
	}
	fmt.Printf("Result of square root of %v: %v\n", n, res.GetNumberRoot())
	// error call
}
