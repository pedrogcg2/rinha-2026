package client

import (
	"context"

	"github.com/pedrogcg2/rinha-2026/pb"
	"google.golang.org/api/transport/grpc"
	"google.golang.org/grpc"
)

func Run(vector [14]float32) float32 {
	dial, err := grpc.Dial("db:50051", grpc.WithInsecure())
	if err != nil {
		return 0
	}
	defer dial.Close()
	getScoreClient := pb.NewGetScoreRequestClient(dial)

	result, err := getScoreClient.GetScore(context.Background(), &pb.VectorRequest{values: vector})

	if err != nil {
		return 0.0
	}
	return result
}
