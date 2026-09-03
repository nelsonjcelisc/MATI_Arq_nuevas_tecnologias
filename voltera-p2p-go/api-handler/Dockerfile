FROM golang:1.23-alpine AS builder

WORKDIR /app

COPY go.work ./
COPY proto/ ./proto/
COPY api-handler/ ./api-handler/
COPY matching-engine/ ./matching-engine/
COPY notificador/ ./notificador/
COPY contract-validator/ ./contract-validator/

WORKDIR /app/api-handler
RUN go build -o api-handler-go main.go

FROM alpine:latest
WORKDIR /root/
COPY --from=builder /app/api-handler/api-handler-go .
EXPOSE 8000
CMD ["./api-handler-go"]
