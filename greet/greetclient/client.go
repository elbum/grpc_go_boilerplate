package main

import (
	"context"
	"fmt"
	"grpc_go_boilerplate/greetpb"
	"io"
	"log"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/status"
)

func main() {
	fmt.Println("I'm Client")

	tls := true

	opts := grpc.WithInsecure()

	if tls {
		// SSL Certificate
		certFile := "ssl/ca2.crt"
		creds, sslErr := credentials.NewClientTLSFromFile(certFile, "")
		if sslErr != nil {
			log.Fatalf("Error with CA Trust certification: %v", sslErr)
		}
		opts = grpc.WithTransportCredentials(creds)
	} else {

	}

	conn, err := grpc.Dial("localhost:50051", opts)

	// Insecure
	// conn, err := grpc.Dial("localhost:50051", grpc.WithInsecure())
	if err != nil {
		log.Fatalf("could not connect grpc: %v", err)
	}
	defer conn.Close()

	c := greetpb.NewGreetServiceClient(conn)
	fmt.Printf("Created Client %f", c)
	doUnary(c)

	// go doSum(c)

	// go doPrimeDecomp(c)

	// doServerStreaming(c)

	// doClientStreaming(c)

	// doComputeAVG(c)

	//doBidiStreaming(c)

	// doFindMax(c)

	// doErrorUnary(c)

	// doUnaryWithDeadline(c, 5*time.Second)
	// doUnaryWithDeadline(c, 1*time.Second)
}
func doUnaryWithDeadline(c greetpb.GreetServiceClient, sec time.Duration) {
	fmt.Println("Starting doUnaryDeadline rpc.")

	// timeout context
	ctx, cancel := context.WithTimeout(context.Background(), sec)
	defer cancel()

	req := &greetpb.GreetWithDeadlineRequest{
		Greeting: &greetpb.Greeting{
			FirstName: "Jack",
			LastName:  "LASTMAN",
		},
	}
	res, err := c.GreetWithDeadline(ctx, req)
	if err != nil {

		statusErr, ok := status.FromError(err)
		if ok {
			if statusErr.Code() == codes.DeadlineExceeded {
				fmt.Println("Timeout was hit ! deadline exceeded")
			} else {
				fmt.Println("unexpected error: ", statusErr)
			}
		} else {
			log.Fatalf("error while calling greet WithDeadline RPC: %v", err)
		}
	}
	log.Printf("response from server %v", res.Result)

}
func doErrorUnary(c greetpb.GreetServiceClient) {
	fmt.Println("Starting doErrorUnary rpc.")
	ctx := context.Background()
	req := &greetpb.SquareRootRequest{
		Number: -2,
	}
	res, err := c.SquareRoot(ctx, req)
	if err != nil {
		respErr, ok := status.FromError(err)
		if ok {
			// actual error from gRPC (user Error)
			fmt.Println(respErr.Message())
			fmt.Println(respErr.Code())
			if respErr.Code() == codes.InvalidArgument {
				fmt.Println("USER ERROR No Negative Numbers")
				return
			}
		} else {
			// framework error
			log.Fatalf("Error Calling SquareRoot : %v\n", err)
		}
	}
	fmt.Printf("response from server %v", res.GetNumberRoot())
	// fmt.Printf("response from server %v", res.NumberRoot) // panic 이 난다... Get 쓰자.
}

func doFindMax(c greetpb.GreetServiceClient) {
	fmt.Println("Starting FindMax Steaming rpc.")

	// create stream client
	ctx := context.Background()
	stream, err := c.FindMax(ctx)
	if err != nil {
		log.Fatalf("error creating stream client: %v", err)
		return
	}
	waitc := make(chan struct{})

	// send message to server
	requests := []int32{3, 5, 54, 1, 3, 5, 123, 80, 567}

	go func() {
		// send message to server
		for _, req := range requests {
			fmt.Printf("Sending numbers ... : %v\n", req)
			stream.Send(&greetpb.FindMaxRequest{
				Num: req,
			})
			time.Sleep(time.Second)
		}
		stream.CloseSend()
	}()

	// receive message from server
	go func() {
		// receive message from server

		for {
			res, err := stream.Recv()
			if err == io.EOF {
				break
			}
			if err != nil {
				log.Fatalf("Error while receiving : %v", err)
				break
			}
			fmt.Printf("Received MAX : %v\n", res.GetResult())
		}
		close(waitc)
	}()

	// Channel for Context Waiting
	<-waitc
}

func doBidiStreaming(c greetpb.GreetServiceClient) {
	fmt.Println("Starting Bidi Steaming rpc.")

	// create stream client
	ctx := context.Background()
	stream, err := c.GreetEveryone(ctx)
	if err != nil {
		log.Fatalf("error creating stream client: %v", err)
		return
	}

	requests := []*greetpb.GreetEveryoneRequest{
		&greetpb.GreetEveryoneRequest{
			Greeting: &greetpb.Greeting{
				FirstName: "John",
			},
		},
		&greetpb.GreetEveryoneRequest{
			Greeting: &greetpb.Greeting{
				FirstName: "TOM",
			},
		},
		&greetpb.GreetEveryoneRequest{
			Greeting: &greetpb.Greeting{
				FirstName: "FORD",
			},
		},
		&greetpb.GreetEveryoneRequest{
			Greeting: &greetpb.Greeting{
				FirstName: "Bryan",
			},
		},
		&greetpb.GreetEveryoneRequest{
			Greeting: &greetpb.Greeting{
				FirstName: "LUCY",
			},
		},
	}
	waitc := make(chan struct{})
	go func() {
		// send message to server
		for _, req := range requests {
			fmt.Printf("Sending ... : %v\n", req)
			stream.Send(req)
			time.Sleep(time.Second)
		}
		stream.CloseSend()
	}()

	go func() {
		// receive message from server

		for {
			res, err := stream.Recv()
			if err == io.EOF {
				break
			}
			if err != nil {
				log.Fatalf("Error while receiving : %v", err)
				break
			}
			fmt.Printf("Received : %v\n", res.GetResult())
		}
		close(waitc)
	}()

	<-waitc
}

func doComputeAVG(c greetpb.GreetServiceClient) {
	fmt.Println("Starting doComputeAVG rpc.")
	requests := []*greetpb.LongNumbers{
		&greetpb.LongNumbers{
			Num: 10,
		},
		&greetpb.LongNumbers{
			Num: 20,
		},
		&greetpb.LongNumbers{
			Num: 30,
		},
		&greetpb.LongNumbers{
			Num: 40,
		},
		&greetpb.LongNumbers{
			Num: 50,
		},
		&greetpb.LongNumbers{
			Num: 60,
		},
	}
	ctx := context.Background()
	stream, err := c.ComputeAVG(ctx)
	if err != nil {
		log.Fatalf("error while computeAVG: %v", err)
	}

	// iterate slice and send  each message
	for _, req := range requests {
		log.Printf("Sending Req : %v\n", req)
		stream.Send(req)
		time.Sleep(1000 * time.Millisecond)
	}

	Numbers := []int32{3, 5, 54, 123, 567}
	for _, req := range Numbers {
		log.Printf("Sending Req : %v\n", req)
		stream.Send(&greetpb.LongNumbers{
			Num: req,
		})
		time.Sleep(1000 * time.Millisecond)
	}
	res, err := stream.CloseAndRecv()
	if err != nil {
		log.Fatalf("error while receiving computeAVG: %v", err)
	}
	fmt.Printf("=== computeAVG === : %v", res)
}

func doClientStreaming(c greetpb.GreetServiceClient) {
	fmt.Println("Starting doPrimeDecomp rpc.")
	requests := []*greetpb.LongGreetRequest{
		&greetpb.LongGreetRequest{
			Greeting: &greetpb.Greeting{
				FirstName: "John",
			},
		},
		&greetpb.LongGreetRequest{
			Greeting: &greetpb.Greeting{
				FirstName: "TOM",
			},
		},
		&greetpb.LongGreetRequest{
			Greeting: &greetpb.Greeting{
				FirstName: "FORD",
			},
		},
		&greetpb.LongGreetRequest{
			Greeting: &greetpb.Greeting{
				FirstName: "Bryan",
			},
		},
		&greetpb.LongGreetRequest{
			Greeting: &greetpb.Greeting{
				FirstName: "LUCY",
			},
		},
	}

	ctx := context.Background()
	stream, err := c.LongGreet(ctx)
	if err != nil {
		log.Fatalf("error while greeting: %v", err)
	}

	// iterate slice and send  each message
	for _, req := range requests {
		log.Printf("Sending Req : %v\n", req)
		stream.Send(req)
		time.Sleep(1000 * time.Millisecond)
	}

	res, err := stream.CloseAndRecv()
	if err != nil {
		log.Fatalf("error while receiving long greeting: %v", err)
	}
	fmt.Printf("Long Greet Response : %v", res)

}

func doPrimeDecomp(c greetpb.GreetServiceClient) {
	fmt.Println("Starting doPrimeDecomp rpc.")

	ctx := context.Background()
	req := &greetpb.PrimeDecompRequest{
		Num: 120000000,
	}

	resStream, err := c.PrimeDecomposition(ctx, req)
	if err != nil {
		log.Fatalf("error while calling GreetmanyTimes RPC : %v", err)
	}

	for {
		msg, err := resStream.Recv()
		if err == io.EOF {
			// end of resStream
			break
		}
		if err != nil {
			log.Fatalf("error while getting stream :%v", err)
		}
		log.Printf("Response from PrimeDecomp: %v", msg.GetResult())
	}
}

func doServerStreaming(c greetpb.GreetServiceClient) {
	fmt.Println("Starting doServerStreaming rpc.")

	ctx := context.Background()
	req := &greetpb.GreetManyTimesRequest{
		Greeting: &greetpb.Greeting{
			FirstName: "ELBUM",
			LastName:  "LEEE",
		},
	}
	resStream, err := c.GreetManyTimes(ctx, req)
	if err != nil {
		log.Fatalf("error while calling GreetmanyTimes RPC : %v", err)
	}

	for {
		msg, err := resStream.Recv()
		if err == io.EOF {
			// end of resStream
			break
		}
		if err != nil {
			log.Fatalf("error while getting stream :%v", err)
		}
		log.Printf("Response from greetManyTimes: %v", msg.GetResult())
	}

}

func doSum(c greetpb.GreetServiceClient) {
	fmt.Println("Starting doSum rpc.")
	ctx := context.Background()
	req := &greetpb.NumReq{
		FirstNum:  10,
		SecondNum: 3,
	}
	res, err := c.Sum(ctx, req)
	if err != nil {
		log.Fatalf("error while calling greet RPC:%v", err)
		// fmt.Printf("error while calling greet RPC:%v", err)
	}
	log.Printf("response from server %v", res.Result)
}

func doUnary(c greetpb.GreetServiceClient) {
	fmt.Println("Starting doUnary rpc.")
	ctx := context.Background()
	req := &greetpb.GreetRequest{
		Greeting: &greetpb.Greeting{
			FirstName: "Jack",
			LastName:  "LASTMAN",
		},
	}
	res, err := c.Greet(ctx, req)
	if err != nil {
		log.Fatalf("error while calling greet RPC:%v", err)
		// fmt.Printf("error while calling greet RPC:%v", err)
	}
	log.Printf("response from server %v", res.Result)

}
