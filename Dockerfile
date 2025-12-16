FROM golang:1.22 AS builder
WORKDIR /app

COPY go.mod ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o /server ./cmd/server

FROM gcr.io/distroless/base-debian12
COPY --from=builder /server /server
EXPOSE 50051
ENTRYPOINT ["/server"]

