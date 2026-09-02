FROM golang:1.23-alpine AS builder

RUN apk add --no-cache protobuf protobuf-dev git

RUN go install google.golang.org/protobuf/cmd/protoc-gen-go@v1.34.2
RUN go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@v1.5.1

WORKDIR /app

COPY proto/ ./proto/
COPY notificador/go.mod ./
COPY notificador/main.go ./

RUN mkdir -p proto && \
    protoc --go_out=. --go-grpc_out=. proto/p2p.proto

RUN go mod edit -require google.golang.org/grpc@v1.70.0
RUN go mod edit -require google.golang.org/protobuf@v1.34.2
RUN go mod edit -require github.com/prometheus/client_golang@v1.20.0
RUN go mod tidy

RUN go build -o notificador-go main.go

FROM alpine:latest
WORKDIR /root/
COPY --from=builder /app/notificador-go .
EXPOSE 50051 2112
CMD ["./notificador-go"]
