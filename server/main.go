package main

import (
	"context"
	"log"
	"net"
	"os"

	pb "github.com/GoingFast/backend2/specs"
	"github.com/golang/protobuf/ptypes/empty"
	"google.golang.org/grpc"
	pbHealth "google.golang.org/grpc/health/grpc_health_v1"
)

type service struct{}

func fallbackEnv(env, fallback string) string {
	e := os.Getenv(env)
	if e == "" {
		return fallback
	}
	return e
}

func (s service) Message(ctx context.Context, _ *empty.Empty) (*pb.MessageResponse, error) {
	return &pb.MessageResponse{
		ServerHostname: fallbackEnv("HOSTNAME", "hostname"),
		Title:          "messauuendusfff",
		Version:        fallbackEnv("VERSION", "version"),
	}, nil
}

func (s service) Check(ctx context.Context, req *pbHealth.HealthCheckRequest) (*pbHealth.HealthCheckResponse, error) {
	return &pbHealth.HealthCheckResponse{Status: 1}, nil
}

func (s service) Watch(req *pbHealth.HealthCheckRequest, srv pbHealth.Health_WatchServer) error {
	return nil
}

func main() {
	ln, err := net.Listen("tcp", ":50051")
	if err != nil {
		log.Fatal(err)
	}
	svc := service{}
	srv := grpc.NewServer()
	pb.RegisterMessageServiceServer(srv, svc)
	pbHealth.RegisterHealthServer(srv, svc)
	srv.Serve(ln)
}
