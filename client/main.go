package main

import (
	"context"
	"fmt"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"io"
	"log"
	calculatorpb "sum-grpc/calculatorpb"
)

func main() {
	clientConn, err := grpc.Dial("localhost:50051", grpc.WithInsecure())
	if err != nil {
		log.Fatalf("failed to connect to server: %v", err)
	}
	defer func() {
		if err := clientConn.Close(); err != nil {
			fmt.Printf("failed to close server connection properly: %v", err)
		}
	}()
	client := calculatorpb.NewCalculatorServiceClient(clientConn)
	// doSum(client)
	// doPrimeNumberDecomposition(client)
	// doComputeAverage(client)
	// doFindMaximum(client)
	doErrorUnary(client)
}

func doErrorUnary(client calculatorpb.CalculatorServiceClient) {
	var input int32 = 10
	// correct call
	doErrorCall(client, input)
	input = -10
	// error call
	doErrorCall(client, input)
}

func doErrorCall(client calculatorpb.CalculatorServiceClient, input int32) {
	squareRootResponse, err := client.SquareRoot(context.Background(), &calculatorpb.SquareRootRequest{
		Number: input,
	})
	if err != nil {
		if errStat, ok := status.FromError(err); ok {
			// actual error from GRPC (user error)
			log.Println(errStat.Message())
			log.Println(errStat.Code())
			if errStat.Code() == codes.InvalidArgument {
				log.Fatalf("we probably sent a negative number: %v", errStat.Err())
			}
		} else {
			log.Fatalf(
				"unspecified error when calculating square root: %v",
				err,
			)
		}
	}
	log.Printf("sqrt(%d) = %f", input, squareRootResponse.GetNumberRoot())
}

func doFindMaximum(client calculatorpb.CalculatorServiceClient) {
	stream, err := client.FindMaximum(context.Background())
	if err != nil {
		log.Fatalf("couldn't start find maximum process: %v", err)
	}
	nums := []int32{1, 5, 3, 6, 2, 20}
	done := make(chan struct{})
	go sendNums(nums, stream)
	go receiveNums(stream, done)
	<-done
}

func receiveNums(stream calculatorpb.CalculatorService_FindMaximumClient, done chan struct{}) {
	for {
		recv, err := stream.Recv()
		if err != nil {
			if err == io.EOF {
				log.Println("finishing receiving maximum numbers")
				break
			}
			done <- struct{}{}
			log.Fatalf("failed to receive next maximum value: %v", err)
		}
		log.Printf("next maximum value is %d", recv.GetMaximum())
	}
	done <- struct{}{}
}

func sendNums(nums []int32, stream calculatorpb.CalculatorService_FindMaximumClient) {
	for _, num := range nums {
		err := stream.Send(&calculatorpb.FindMaximumRequest{
			Number: num,
		})
		if err != nil {
			log.Fatalf("couldn't send find maximum request %d: %v", num, err)
		}
		log.Printf("sent %d…", num)
	}
	err := stream.CloseSend()
	if err != nil {
		log.Fatalf("failed to close sending client: %v", err)
	}
}

func doComputeAverage(client calculatorpb.CalculatorServiceClient) {
	computeAverageClient, err := client.ComputeAverage(context.Background())
	if err != nil {
		log.Fatalf("couldn't connect to compute average client: %v", err)
	}
	nums := []int32{1, 2, 3, 4}
	for _, num := range nums {
		err := computeAverageClient.Send(&calculatorpb.ComputeAverageRequest{
			Number: num,
		})
		if err != nil {
			log.Fatalf("failed to stream next number %d to client: %v", num, err)
		}
	}
	response, err := computeAverageClient.CloseAndRecv()
	if err != nil {
		log.Fatalf("didn't receive response from server: %v", err)
	}
	log.Printf("the average of %v = %.2f\n", nums, response.GetResult())
}

func doPrimeNumberDecomposition(client calculatorpb.CalculatorServiceClient) {
	var i int64 = 12390392840
	decResp, err := client.PrimeNumberDecomposition(context.Background(), &calculatorpb.PrimeNumberDecompositionRequest{
		PrimeNumber: i,
	})
	if err != nil {
		log.Fatalf("decomposition request failed: %v", err)
	}
	log.Printf("%d decomposes into the following prime numbers:\n", i)
	for {
		result, err := decResp.Recv()
		if err != nil {
			if err == io.EOF {
				break
			}
			log.Fatalf("failed to get all responses: %v", err)
		}
		log.Printf("%d", result.GetResult())
	}
}

func doSum(client calculatorpb.CalculatorServiceClient) {
	var first int32 = 10
	var second int32 = 3
	resp, err := client.Sum(context.Background(), &calculatorpb.SumRequest{
		FirstNum:  first,
		SecondNum: second,
	})
	if err != nil {
		fmt.Printf("calculator response error: %v", err)
		return
	}
	fmt.Printf("%d + %d = %d\n", first, second, resp.GetResult())
}
