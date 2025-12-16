package main

import (
	"context"
	"log/slog"
	"net"
	"os"
	"os/signal"
	"syscall"
	"time"

	"google.golang.org/grpc"
	healthpb "google.golang.org/grpc/health/grpc_health_v1"
	"google.golang.org/grpc/reflection"

	vectordbv1 "github.com/faportillo/vectordb-server/gen/proto/vectordb/v1"
	"github.com/faportillo/vectordb-server/internal/service"
)

// Minimal health service implementation that always reports SERVING.
type healthServer struct {
	healthpb.UnimplementedHealthServer
}

func (h *healthServer) Check(ctx context.Context, req *healthpb.HealthCheckRequest) (*healthpb.HealthCheckResponse, error) {
	return &healthpb.HealthCheckResponse{Status: healthpb.HealthCheckResponse_SERVING}, nil
}

// Watch is not implemented for simplicity.
func (h *healthServer) Watch(req *healthpb.HealthCheckRequest, srv healthpb.Health_WatchServer) error {
	return nil
}

func main() {
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}))

	addr := ":50051"
	lis, err := net.Listen("tcp", addr)
	if err != nil {
		logger.Error("failed to listen", "addr", addr, "err", err)
		os.Exit(1)
	}
	logger.Info("starting gRPC server", "addr", addr)

	s := grpc.NewServer()

	// Register services
	vectordbv1.RegisterVectorDBServer(s, service.NewServer())

	// Health service
	healthpb.RegisterHealthServer(s, &healthServer{})

	// Reflection for grpcurl
	reflection.Register(s)

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	go func() {
		if err := s.Serve(lis); err != nil {
			logger.Error("gRPC server exited", "err", err)
		}
	}()

	<-ctx.Done()
	logger.Info("shutting down gRPC server...")

	done := make(chan struct{})
	go func() {
		s.GracefulStop()
		close(done)
	}()

	select {
	case <-done:
		logger.Info("gRPC server stopped gracefully")
	case <-time.After(5 * time.Second):
		logger.Warn("graceful stop timeout, forcing stop")
		s.Stop()
	}
}

