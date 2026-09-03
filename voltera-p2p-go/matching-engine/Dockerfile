FROM golang:1.23-alpine AS builder

WORKDIR /app

COPY go.work ./
COPY proto/ ./proto/
COPY api-handler/ ./api-handler/
COPY matching-engine/ ./matching-engine/
COPY notificador/ ./notificador/
COPY contract-validator/ ./contract-validator/

WORKDIR /app/matching-engine
RUN go build -o matching-engine-go main.go

FROM alpine:latest
WORKDIR /root/
COPY --from=builder /app/matching-engine/matching-engine-go .
EXPOSE 8080
CMD ["./matching-engine-go"]
