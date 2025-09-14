FROM golang:1.23-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN go build -o bin/payment ./payment/cmd/main.go

FROM alpine:3.18

WORKDIR /root
RUN apk add --no-cache ca-certificates

COPY --from=builder /app/bin/payment /root/bin/payment
RUN chmod +x /root/bin/payment

COPY --from=builder /app/payment/configs /root/configs/payment

ENV PAYMENT_CONFIG=/root/configs/payment/config.yaml

EXPOSE 8081 50053

ENTRYPOINT ["/root/bin/payment"]