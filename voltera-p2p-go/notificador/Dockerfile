FROM golang:1.23-alpine AS builder

WORKDIR /app

COPY go.work ./
COPY proto/ ./proto/
COPY api-handler/ ./api-handler/
COPY matching-engine/ ./matching-engine/
COPY notificador/ ./notificador/
COPY contract-validator/ ./contract-validator/

WORKDIR /app/notificador
RUN go build -o notificador-go main.go

FROM alpine:latest
WORKDIR /root/
COPY --from=builder /app/notificador/notificador-go .
EXPOSE 50051 2112
CMD ["./notificador-go"]
