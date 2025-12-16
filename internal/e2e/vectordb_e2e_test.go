package e2e

import (
	"context"
	"net"
	"testing"
	"time"

	"github.com/faportillo/vectordb-server/internal/core"
	"github.com/faportillo/vectordb-server/internal/service"
	vectordbv1 "github.com/faportillo/vectordb-server/gen/proto/vectordb/v1"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func startTestServer(t *testing.T) (addr string, stop func()) {
	t.Helper()

	lis, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("listen: %v", err)
	}

	grpcServer := grpc.NewServer()
	store := core.NewStore()
	vectordbv1.RegisterVectorDBServer(grpcServer, service.NewServer(store))

	go func() {
		_ = grpcServer.Serve(lis)
	}()

	return lis.Addr().String(), func() {
		grpcServer.GracefulStop()
		_ = lis.Close()
	}
}

func dialClient(t *testing.T, addr string) vectordbv1.VectorDBClient {
	t.Helper()

	conn, err := grpc.Dial(
		addr,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithBlock(),
		grpc.WithTimeout(2*time.Second),
	)
	if err != nil {
		t.Fatalf("dial: %v", err)
	}
	t.Cleanup(func() { _ = conn.Close() })

	return vectordbv1.NewVectorDBClient(conn)
}

func TestVectorDB_E2E_CreateUpsertSearchDelete(t *testing.T) {
	addr, stop := startTestServer(t)
	t.Cleanup(stop)

	client := dialClient(t, addr)

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	t.Cleanup(cancel)

	// 1) Create collection
	_, err := client.CreateCollection(ctx, &vectordbv1.CreateCollectionRequest{
		Name:      "users",
		Dimension: 3,
		Distance:  "dot",
	})
	if err != nil {
		t.Fatalf("CreateCollection: %v", err)
	}

	// 2) Upsert points
	_, err = client.UpsertPoints(ctx, &vectordbv1.UpsertPointsRequest{
		Collection: "users",
		Points: []*vectordbv1.Point{
			{Id: "a", Vector: &vectordbv1.Vector{Values: []float32{1, 0, 0}}},
			{Id: "b", Vector: &vectordbv1.Vector{Values: []float32{0, 1, 0}}},
			{Id: "c", Vector: &vectordbv1.Vector{Values: []float32{2, 0, 0}}},
		},
	})
	if err != nil {
		t.Fatalf("UpsertPoints: %v", err)
	}

	// 3) Search (query closer to x-axis, c should rank above a)
	resp, err := client.Search(ctx, &vectordbv1.SearchRequest{
		Collection:      "users",
		Query:           &vectordbv1.Vector{Values: []float32{1, 0, 0}},
		TopK:            2,
		IncludeVectors:  true,
		IncludeMetadata: false,
	})
	if err != nil {
		t.Fatalf("Search: %v", err)
	}
	if len(resp.Results) != 2 {
		t.Fatalf("Search results len=%d, want 2", len(resp.Results))
	}
	if resp.Results[0].Id != "c" {
		t.Fatalf("best result=%q, want %q", resp.Results[0].Id, "c")
	}
	if resp.Results[1].Id != "a" {
		t.Fatalf("second result=%q, want %q", resp.Results[1].Id, "a")
	}

	// 4) Delete
	_, err = client.DeletePoints(ctx, &vectordbv1.DeletePointsRequest{
		Collection: "users",
		Ids:        []string{"c"},
	})
	if err != nil {
		t.Fatalf("DeletePoints: %v", err)
	}

	// 5) Search again (c should be gone)
	resp2, err := client.Search(ctx, &vectordbv1.SearchRequest{
		Collection: "users",
		Query:      &vectordbv1.Vector{Values: []float32{1, 0, 0}},
		TopK:       3,
	})
	if err != nil {
		t.Fatalf("Search2: %v", err)
	}
	for _, r := range resp2.Results {
		if r.Id == "c" {
			t.Fatalf("expected c to be deleted, but it appeared in results")
		}
	}
}
