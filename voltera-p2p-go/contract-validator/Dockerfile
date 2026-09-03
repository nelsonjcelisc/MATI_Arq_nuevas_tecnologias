FROM golang:1.23-alpine AS builder

WORKDIR /app

COPY go.work ./
COPY proto/ ./proto/
COPY api-handler/ ./api-handler/
COPY matching-engine/ ./matching-engine/
COPY notificador/ ./notificador/
COPY contract-validator/ ./contract-validator/

WORKDIR /app/contract-validator
RUN go build -o contract-validator-go main.go

FROM alpine:latest
WORKDIR /root/
COPY --from=builder /app/contract-validator/contract-validator-go .
EXPOSE 50052 2113
CMD ["./contract-validator-go"]
