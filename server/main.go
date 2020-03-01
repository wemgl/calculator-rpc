package main

import (
	"context"
	"fmt"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/reflection"
	"google.golang.org/grpc/status"
	"io"
	"log"
	"math"
	"net"
	"sum-grpc/calculatorpb"
)

type server struct{}

func (s *server) SquareRoot(
	ctx context.Context,
	req *calculatorpb.SquareRootRequest,
) (*calculatorpb.SquareRootResponse, error) {
	number := req.GetNumber()
	if number < 0 {
		return nil, status.Errorf(
			codes.InvalidArgument,
			"received a negative number: %d",
			number,
		)
	}
	return &calculatorpb.SquareRootResponse{
		NumberRoot: math.Sqrt(float64(number)),
	}, nil
}

func (s *server) FindMaximum(
	stream calculatorpb.CalculatorService_FindMaximumServer,
) error {
	var max int32 = math.MinInt32
	for {
		req, err := stream.Recv()
		if err != nil {
			if err == io.EOF {
				break
			}
			return status.Errorf(
				codes.Internal,
				"failed to receive next request: %v",
				err,
			)
		}
		log.Printf("client sent %d", req.GetNumber())
		if max < req.GetNumber() {
			max = req.GetNumber()
			err := stream.Send(&calculatorpb.FindMaximumResponse{
				Maximum: max,
			})
			if err != nil {
				return status.Errorf(
					codes.Internal,
					"failed to send new maximum: %v",
					max,
				)
			}
		}
	}
	return nil
}

func (s *server) ComputeAverage(
	stream calculatorpb.CalculatorService_ComputeAverageServer,
) error {
	var nums []int32
	for {
		recv, err := stream.Recv()
		if err != nil {
			if err == io.EOF {
				err := stream.SendAndClose(&calculatorpb.ComputeAverageResponse{
					Result: average(nums),
				})
				if err != nil {
					return status.Errorf(
						codes.Internal,
						"failed to return average: %v",
						err,
					)
				}
				return nil
			}
			return status.Errorf(
				codes.Internal,
				"failed to receive request from stream: %v",
				err,
			)
		}
		nums = append(nums, recv.GetNumber())
	}
}

func average(nums []int32) float32 {
	var sum int32
	for _, num := range nums {
		sum = sum + num
	}
	return float32(sum) / float32(len(nums))
}

func (s *server) Sum(ctx context.Context, req *calculatorpb.SumRequest) (*calculatorpb.SumResponse, error) {
	result := req.GetFirstNum() + req.GetSecondNum()
	resp := &calculatorpb.SumResponse{
		Result: result,
	}
	return resp, nil
}

func (s *server) PrimeNumberDecomposition(
	req *calculatorpb.PrimeNumberDecompositionRequest,
	stream calculatorpb.CalculatorService_PrimeNumberDecompositionServer,
) error {
	var divisor int64 = 2
	n := req.GetPrimeNumber()
	for n > 1 {
		if n%divisor == 0 { // if divisor evenly divides into N
			err := stream.Send(&calculatorpb.PrimeNumberDecompositionResponse{
				Result: divisor, // this is a factor
			})
			if err != nil {
				return status.Errorf(
					codes.Internal,
					"failed to send next prime number in decomposition: %v",
					err,
				)
			}
			n = n / divisor // divide N by divisor so that we have the rest of the number left.
		} else {
			divisor = divisor + 1
		}
	}
	return nil
}

func main() {
	lis, err := net.Listen("tcp", "0.0.0.0:50051")
	if err != nil {
		log.Fatalf("failed to listen on 50051")
	}

	serv := grpc.NewServer()
	calculatorpb.RegisterCalculatorServiceServer(serv, &server{})

	reflection.Register(serv)

	if err := serv.Serve(lis); err != nil {
		fmt.Printf("failed to listen server requests: %v", err)
	}
}
