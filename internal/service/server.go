package service

import (
	"context"
	"errors"
	"fmt"

	"github.com/faportillo/vectordb-server/internal/core"
	vectordbv1 "github.com/faportillo/vectordb-server/gen/proto/vectordb/v1"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
)

type Server struct {
	vectordbv1.UnimplementedVectorDBServer
	store *core.Store
}

func NewServer(store *core.Store) *Server {
	return &Server{store: store}
}

// mapCoreErr converts core-layer errors into appropriate gRPC status codes.
func mapCoreErr(err error) error {
	switch {
	case errors.Is(err, core.ErrCollectionNotFound):
		return status.Error(codes.NotFound, err.Error())
	case errors.Is(err, core.ErrCollectionExists):
		return status.Error(codes.AlreadyExists, err.Error())
	case errors.Is(err, core.ErrInvalidK):
		return status.Error(codes.InvalidArgument, err.Error())
	case errors.Is(err, core.ErrDimensionMismatch):
		return status.Error(codes.InvalidArgument, err.Error())
	default:
		// For now, treat unexpected errors as Internal.
		return status.Error(codes.Internal, fmt.Sprintf("internal error: %v", err))
	}
}

func (s *Server) CreateCollection(ctx context.Context, req *vectordbv1.CreateCollectionRequest) (*vectordbv1.CollectionInfo, error) {
	if req.GetName() == "" {
		return nil, status.Error(codes.InvalidArgument, "name must not be empty")
	}
	if req.GetDimension() == 0 {
		return nil, status.Error(codes.InvalidArgument, "dimension must be > 0")
	}

	metric, err := core.ParseDistanceMetric(req.GetDistance())
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	c, err := s.store.CreateCollection(req.GetName(), int(req.GetDimension()), metric)
	if err != nil {
		return nil, mapCoreErr(err)
	}

	return &vectordbv1.CollectionInfo{
		Name:      c.Name(),
		Dimension: uint32(c.Dimension()),
		Distance:  string(c.Metric()),
		Count:     uint64(c.Size()),
	}, nil
}

func (s *Server) GetCollection(ctx context.Context, req *vectordbv1.GetCollectionRequest) (*vectordbv1.CollectionInfo, error) {
	if req.GetName() == "" {
		return nil, status.Error(codes.InvalidArgument, "name must not be empty")
	}

	c, err := s.store.GetCollection(req.GetName())
	if err != nil {
		return nil, mapCoreErr(err)
	}

	return &vectordbv1.CollectionInfo{
		Name:      c.Name(),
		Dimension: uint32(c.Dimension()),
		Distance:  string(c.Metric()),
		Count:     uint64(c.Size()),
	}, nil
}

func (s *Server) UpsertPoints(ctx context.Context, req *vectordbv1.UpsertPointsRequest) (*vectordbv1.UpsertPointsResponse, error) {
	if req.GetCollection() == "" {
		return nil, status.Error(codes.InvalidArgument, "collection must not be empty")
	}

	c, err := s.store.GetCollection(req.GetCollection())
	if err != nil {
		return nil, mapCoreErr(err)
	}

	points := req.GetPoints()
	if len(points) == 0 {
		return &vectordbv1.UpsertPointsResponse{Upserted: 0}, nil
	}

	var upserted uint64
	for i, p := range points {
		if p.GetId() == "" {
			return nil, status.Errorf(codes.InvalidArgument, "points[%d].id must not be empty", i)
		}
		vec := p.GetVector().GetValues()
		// Metadata is already a *structpb.Struct in the generated type (or nil).
		if err := c.Upsert(p.GetId(), vec, p.GetMetadata()); err != nil {
			return nil, mapCoreErr(err)
		}
		upserted++
	}

	return &vectordbv1.UpsertPointsResponse{Upserted: upserted}, nil
}

func (s *Server) DeletePoints(ctx context.Context, req *vectordbv1.DeletePointsRequest) (*emptypb.Empty, error) {
	if req.GetCollection() == "" {
		return nil, status.Error(codes.InvalidArgument, "collection must not be empty")
	}

	c, err := s.store.GetCollection(req.GetCollection())
	if err != nil {
		return nil, mapCoreErr(err)
	}

	for _, id := range req.GetIds() {
		// Delete is idempotent; ignore empty ids
		_ = c.Delete(id)
	}

	return &emptypb.Empty{}, nil
}

func (s *Server) Search(ctx context.Context, req *vectordbv1.SearchRequest) (*vectordbv1.SearchResponse, error) {
	if req.GetCollection() == "" {
		return nil, status.Error(codes.InvalidArgument, "collection must not be empty")
	}
	if req.GetTopK() == 0 {
		return nil, status.Error(codes.InvalidArgument, "top_k must be > 0")
	}
	if req.GetQuery() == nil {
		return nil, status.Error(codes.InvalidArgument, "query must not be nil")
	}

	c, err := s.store.GetCollection(req.GetCollection())
	if err != nil {
		return nil, mapCoreErr(err)
	}

	query := req.GetQuery().GetValues()

	// NOTE: filter is ignored for now (Phase 1). We'll implement it next.
	results, err := c.Search(query, int(req.GetTopK()), req.GetIncludeVectors(), req.GetIncludeMetadata())
	if err != nil {
		return nil, mapCoreErr(err)
	}

	resp := &vectordbv1.SearchResponse{
		Results: make([]*vectordbv1.ScoredPoint, 0, len(results)),
	}

	for _, r := range results {
		sp := &vectordbv1.ScoredPoint{
			Id:    r.ID,
			Score: r.Score,
		}

		if req.GetIncludeVectors() {
			sp.Vector = &vectordbv1.Vector{Values: r.Vector}
		}
		if req.GetIncludeMetadata() {
			sp.Metadata = r.Metadata
		}

		resp.Results = append(resp.Results, sp)
	}

	return resp, nil
}

