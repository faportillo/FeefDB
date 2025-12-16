package service

import (
	"context"

	vectordbv1 "github.com/faportillo/vectordb-server/gen/proto/vectordb/v1"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
)

type Server struct {
	vectordbv1.UnimplementedVectorDBServer
}

func NewServer() *Server {
	return &Server{}
}

func (s *Server) CreateCollection(ctx context.Context, req *vectordbv1.CreateCollectionRequest) (*vectordbv1.CollectionInfo, error) {
	return nil, status.Error(codes.Unimplemented, "not implemented")
}

func (s *Server) GetCollection(ctx context.Context, req *vectordbv1.GetCollectionRequest) (*vectordbv1.CollectionInfo, error) {
	return nil, status.Error(codes.Unimplemented, "not implemented")
}

func (s *Server) UpsertPoints(ctx context.Context, req *vectordbv1.UpsertPointsRequest) (*vectordbv1.UpsertPointsResponse, error) {
	return nil, status.Error(codes.Unimplemented, "not implemented")
}

func (s *Server) DeletePoints(ctx context.Context, req *vectordbv1.DeletePointsRequest) (*emptypb.Empty, error) {
	return nil, status.Error(codes.Unimplemented, "not implemented")
}

func (s *Server) Search(ctx context.Context, req *vectordbv1.SearchRequest) (*vectordbv1.SearchResponse, error) {
	return nil, status.Error(codes.Unimplemented, "not implemented")
}

