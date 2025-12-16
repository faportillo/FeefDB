# VectorDB Server (Go, gRPC)

Minimal gRPC server scaffold with Buf-based protobuf generation, health service, and reflection enabled. All VectorDB RPCs return UNIMPLEMENTED.

## Prerequisites
- Go 1.22+
- buf (`brew install bufbuild/buf/buf`)
- golangci-lint (`brew install golangci-lint`)
- grpcurl (`brew install grpcurl`)

## Commands

```bash
# Generate protobuf code
make proto

# Tidy modules
make tidy

# Build / Test / Lint
make build
make test
make lint

# Run the server on 0.0.0.0:50051
make run
```

## grpcurl Examples

```bash
# List services
grpcurl -plaintext localhost:50051 list

# Health check (overall)
grpcurl -plaintext localhost:50051 grpc.health.v1.Health/Check

# Health check (service-specific)
grpcurl -plaintext -d '{"service":"vectordb.v1.VectorDB"}' localhost:50051 grpc.health.v1.Health/Check

# Call GetCollection (will be UNIMPLEMENTED)
grpcurl -plaintext -d '{"name":"test"}' localhost:50051 vectordb.v1.VectorDB/GetCollection
```

## Repo Layout
- `proto/` – protobuf sources
- `gen/` – generated code (created by `make proto`)
- `internal/service/` – service stubs
- `cmd/server/` – server entrypoint

