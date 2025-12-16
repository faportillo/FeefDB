# VectorDB Server (Go, gRPC)

A from-scratch vector database server written in Go, exposing a gRPC API for vector storage and nearest-neighbor search.

This project is intentionally built in phases to deeply understand how vector databases work internally, rather than treating them as a black box:

- Phase 1: In-memory brute-force vector search (correctness first)
- Phase 2: Approximate nearest-neighbor indexing (IVF, HNSW)
- Phase 3: Persistence, metadata filtering, and performance tuning

The server uses gRPC and Protobuf as a stable API contract and is designed around a clean separation of concerns:

- Transport layer: gRPC service implementation
- Core layer: indexing, collections, vector math, and metadata storage

Client libraries (Python async, TypeScript, etc.) are intended to be thin wrappers over this API.

---

## Current Status

- gRPC server with health checks and reflection
- Protobuf definitions with Buf-based code generation
- In-memory collections with vector upsert, search, and delete
- Brute-force index (linear scan) for correctness
- End-to-end tests using a real gRPC client and server
- Approximate indexing (IVF / HNSW) planned
- Persistence and metadata filtering planned

This project is not production-ready. The goal is learning, correctness, and architectural clarity.

## Prerequisites

- Go 1.22+
- buf (brew install bufbuild/buf/buf)
- golangci-lint (brew install golangci-lint)
- grpcurl (brew install grpcurl)

---

## Commands

Generate protobuf code:
make proto

Tidy modules:
make tidy

Build / Test / Lint:
make build
make test
make lint

Run the server on 0.0.0.0:50051:
make run

---

## gRPC API Exploration (grpcurl)

List all exposed services (reflection enabled):
grpcurl -plaintext localhost:50051 list

Health check (overall):
grpcurl -plaintext localhost:50051 grpc.health.v1.Health/Check

Health check (service-specific):
grpcurl -plaintext -d '{"service":"vectordb.v1.VectorDB"}' localhost:50051 grpc.health.v1.Health/Check

## Example Workflow (End-to-End)

Create a collection:
grpcurl -plaintext -d '{
"name":"users",
"dimension":3,
"distance":"dot"
}' localhost:50051 vectordb.v1.VectorDB/CreateCollection

Upsert vectors:
grpcurl -plaintext -d '{
"collection":"users",
"points":[
{"id":"a","vector":{"values":[1,0,0]}},
{"id":"b","vector":{"values":[0,1,0]}},
{"id":"c","vector":{"values":[2,0,0]}}
]
}' localhost:50051 vectordb.v1.VectorDB/UpsertPoints

Search (expect c, then a):
grpcurl -plaintext -d '{
"collection":"users",
"query":{"values":[1,0,0]},
"top_k":2,
"include_vectors":true,
"include_metadata":false
}' localhost:50051 vectordb.v1.VectorDB/Search

Delete a point:
grpcurl -plaintext -d '{
"collection":"users",
"ids":["c"]
}' localhost:50051 vectordb.v1.VectorDB/DeletePoints

## Repository Layout

vectordb-server/
├── proto/ Protobuf source-of-truth (API contract)
├── gen/ Generated Go code (from buf generate)
├── internal/
│ ├── core/ Vector DB engine (indexes, collections, metadata)
│ ├── service/ gRPC service implementation (thin transport layer)
│ └── e2e/ End-to-end gRPC tests (real server + client)
├── cmd/server/ Server entrypoint (main.go)
└── ...

Design notes:

- internal/core is transport-agnostic and fully testable without gRPC
- internal/service handles validation, error mapping, and protobuf translation
- Protobufs are versioned (vectordb.v1) to allow safe API evolution

## Running the Server

cd vectordb-server
make proto
make tidy
make run listens on 0.0.0.0:50051

---

## Testing

Core unit tests (indexing, collections, distance functions):
go test ./internal/core -v

End-to-end gRPC tests (in-process server + real client):
go test ./internal/e2e -v

Everything:
go test ./... -v

End-to-end tests spin up a real gRPC server on an ephemeral port and validate:

- API behavior
- gRPC error codes
- vector search correctness
- protobuf serialization and deserialization

---

## Roadmap (Short Term)

- Metadata filtering (exact-match, then indexed)
- IVF-Flat index implementation
- Benchmarks and profiling
- Persistence (WAL + snapshots)
- Python async client library
