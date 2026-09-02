FROM golang:1.23-alpine AS builder

RUN apk add --no-cache protobuf protobuf-dev git

RUN go install google.golang.org/protobuf/cmd/protoc-gen-go@v1.34.2
RUN go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@v1.5.1

WORKDIR /app

COPY proto/ ./proto/
COPY matching-engine/go.mod ./
COPY matching-engine/main.go ./

RUN mkdir -p proto && \
    protoc --go_out=. --go-grpc_out=. proto/p2p.proto

RUN go mod edit -require google.golang.org/grpc@v1.70.0
RUN go mod edit -require google.golang.org/protobuf@v1.34.2
RUN go mod tidy

RUN go build -o matching-engine-go main.go

FROM alpine:latest
WORKDIR /root/
COPY --from=builder /app/matching-engine-go .
EXPOSE 8080
CMD ["./matching-engine-go"]
