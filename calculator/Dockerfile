FROM golang:1.23-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN go build -o bin/calculator ./calculator/cmd/main.go

FROM alpine:3.18

WORKDIR /root
RUN apk add --no-cache ca-certificates

COPY --from=builder /app/bin/calculator /root/bin/calculator
RUN chmod +x /root/bin/calculator

COPY --from=builder /app/calculator/configs /root/configs/calculator

ENV CALCULATOR_CONFIG=/root/configs/calculator/config.yaml

EXPOSE 8121 50051

ENTRYPOINT ["/root/bin/calculator"]