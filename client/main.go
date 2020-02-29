package main

import (
	"context"
	"fmt"
	"google.golang.org/grpc"
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
	// doPrimeNumberDecomposition(client)
	doComputeAverage(client)
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
