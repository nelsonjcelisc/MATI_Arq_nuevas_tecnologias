FROM golang:1.23-alpine AS builder

RUN apk add --no-cache protobuf protobuf-dev git

RUN go install google.golang.org/protobuf/cmd/protoc-gen-go@v1.34.2
RUN go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@v1.5.1

WORKDIR /app

COPY proto/ ./proto/
COPY api-handler/go.mod ./
COPY api-handler/main.go ./

RUN mkdir -p proto && \
    protoc --go_out=. --go-grpc_out=. proto/p2p.proto

RUN go mod edit -require google.golang.org/grpc@v1.70.0
RUN go mod edit -require google.golang.org/protobuf@v1.34.2
RUN go mod tidy

RUN go build -o api-handler-go main.go

FROM alpine:latest
WORKDIR /root/
COPY --from=builder /app/api-handler-go .
EXPOSE 8000
CMD ["./api-handler-go"]
