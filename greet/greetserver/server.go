package main

import (
	"context"
	"fmt"
	"grpc_go_boilerplate/greetpb"
	"io"
	"log"
	"math"
	"net"
	"strconv"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/reflection"
	"google.golang.org/grpc/status"
)

type server struct {
	greetpb.UnimplementedGreetServiceServer //  dummyServerService
}

func (*server) GreetWithDeadline(ctx context.Context, req *greetpb.GreetWithDeadlineRequest) (*greetpb.GreetWithDeadlineResponse, error) {
	fmt.Printf("GreetWith Deadline function was invoked with %v \n", req)
	for i := 0; i < 3; i++ {
		if ctx.Err() == context.Canceled {
			fmt.Println("The Client canceled the requests")
			return nil, status.Error(codes.Canceled, "The Client Canceled the requests")
		}
		time.Sleep(1 * time.Second)
	}
	firstName := req.GetGreeting().GetFirstName()

	result := "Hello " + firstName + "\n"
	response := &greetpb.GreetWithDeadlineResponse{
		Result: result,
	}
	return response, nil
}
func (*server) SquareRoot(ctx context.Context, req *greetpb.SquareRootRequest) (*greetpb.SquareRootResponse, error) {
	fmt.Printf("SquareRoot function was invoked\n")
	number := req.GetNumber()
	if number < 0 {
		// throw Error
		return nil, status.Errorf(
			codes.InvalidArgument,
			fmt.Sprintf("Received Negative Number", number),
		)
	}
	return &greetpb.SquareRootResponse{
		NumberRoot: math.Sqrt(float64(number)),
	}, nil
}
func (*server) FindMax(stream greetpb.GreetService_FindMaxServer) error {
	fmt.Printf("FindMax function was invoked\n")
	var max int32 = -1
	for {
		req, err := stream.Recv()
		if err == io.EOF {
			return nil
		}
		if err != nil {
			log.Fatalf("Error while reading client stream: %v", err)
			return nil
		}
		num := req.GetNum()
		if num > max {
			max = num

			sendErr := stream.Send(&greetpb.FindMaxResponse{
				Result: max,
			})

			if sendErr != nil {
				log.Fatalf("Error while sending data to client : %v", sendErr)
				return sendErr
			}
		}
	}
}
func (*server) GreetEveryone(stream greetpb.GreetService_GreetEveryoneServer) error {
	fmt.Printf("GreetEveryone function was invoked\n")

	for {
		req, err := stream.Recv()
		if err == io.EOF {
			return nil
		}
		if err != nil {
			log.Fatalf("Error while reading client stream: %v", err)
			return nil
		}
		firstName := req.GetGreeting().GetFirstName()
		result := "Hello " + firstName + " !"
		sendErr := stream.Send(&greetpb.GreetEveryoneResponse{
			Result: result,
		})
		if sendErr != nil {
			log.Fatalf("Error while sending data to client : %v", sendErr)
			return sendErr
		}
	}
}

func (*server) ComputeAVG(stream greetpb.GreetService_ComputeAVGServer) error {
	fmt.Printf("ComputeAVG function was invoked\n")
	var result float32
	var cnt float32 = 0
	for {
		req, err := stream.Recv()

		if err == io.EOF {
			// finish processing
			return stream.SendAndClose(&greetpb.ComputeAVGResponse{
				Result: float32(result / cnt),
			})

		}
		cnt += 1
		if err != nil {
			log.Fatalf("Error while reading client stream : %v", err)
		}
		num := float32(req.GetNum())
		result += num

	}

}
func (*server) LongGreet(stream greetpb.GreetService_LongGreetServer) error {
	fmt.Printf("LongGreet function was invoked\n")
	result := ""
	for {
		req, err := stream.Recv()

		if err == io.EOF {
			// finish
			return stream.SendAndClose(&greetpb.LongGreetResponse{
				Result: result,
			})
		}

		if err != nil {
			log.Fatalf("Error while reading client stream : %v", err)
		}
		firstName := req.GetGreeting().GetFirstName()
		result += "Hellow " + firstName + " ? "
		fmt.Println("Hello " + firstName + "! ")

	}

}
func (*server) Greet(ctx context.Context, req *greetpb.GreetRequest) (*greetpb.GreetResponse, error) {
	fmt.Printf("Greet function was invoked with %v \n", req)
	firstName := req.GetGreeting().GetFirstName()

	result := "Hello " + firstName + "\n"
	response := &greetpb.GreetResponse{
		Result: result,
	}
	return response, nil
}

func (*server) GreetManyTimes(req *greetpb.GreetManyTimesRequest, stream greetpb.GreetService_GreetManyTimesServer) error {
	fmt.Printf("Greet manytimes server stream function was invoked with %v \n", req)
	firstname := req.GetGreeting().GetFirstName()
	for i := 0; i < 10; i++ {
		result := "hello" + firstname + " number " + strconv.Itoa(i)
		res := &greetpb.GreetManyTimesResponse{
			Result: result,
		}
		stream.Send(res)
		time.Sleep(1 * time.Second)
	}
	return nil
}

func (*server) PrimeDecomposition(req *greetpb.PrimeDecompRequest, stream greetpb.GreetService_PrimeDecompositionServer) error {
	fmt.Printf("Prime Decomposition function was invoked with %v \n", req)

	var k int32 = 2
	N := req.Num

	for N > 1 {
		if N%k == 0 {

			res := &greetpb.PrimeDecompResponse{
				Result: k,
			}
			stream.Send(res)
			time.Sleep(50 * time.Millisecond)

			N = N / k
		} else {
			k++
		}
	}
	return nil
}

func (*server) Sum(ctx context.Context, req *greetpb.NumReq) (*greetpb.NumRes, error) {
	fmt.Printf("Sum function was invoked with %v \n", req)
	res := &greetpb.NumRes{
		Result: req.FirstNum + req.SecondNum,
	}
	return res, nil
}

func main() {
	fmt.Println("hello server init")

	lis, err := net.Listen("tcp", "localhost:50051")
	if err != nil {
		log.Fatalln("Fail to listen: ", err)
	}
	opts := []grpc.ServerOption{}

	tls := true
	if tls {
		// for SSL Encryptions
		certFile := "ssl/server.crt"
		keyFile := "ssl/server.pem"
		creds, sslErr := credentials.NewServerTLSFromFile(certFile, keyFile)
		if sslErr != nil {
			log.Fatalf("Error Loading Certification")
			return
		}
		opts = append(opts, grpc.Creds(creds))

	}

	s := grpc.NewServer(opts...)

	//	s := grpc.NewServer()
	greetpb.RegisterGreetServiceServer(s, &server{})

	// Reflection
	reflection.Register(s)

	if err := s.Serve(lis); err != nil {
		log.Fatalf("Failed to server : %v", err)
	}
}
