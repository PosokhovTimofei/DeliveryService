FROM golang:1.23-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN go build -o bin/auth ./auth/cmd/main.go

FROM alpine:3.18

WORKDIR /root
RUN apk add --no-cache ca-certificates

COPY --from=builder /app/bin/auth /root/bin/auth
RUN chmod +x /root/bin/auth

COPY --from=builder /app/auth/configs /root/configs/auth

ENV AUTH_CONFIG=/root/configs/auth/config.yaml

EXPOSE 1703 1704 1705 50052

ENTRYPOINT ["/root/bin/auth"]