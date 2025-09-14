FROM golang:1.23-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN go build -o bin/telegram ./telegram/cmd/main.go

FROM alpine:3.18

WORKDIR /root
RUN apk add --no-cache ca-certificates

COPY --from=builder /app/bin/telegram /root/bin/telegram
RUN chmod +x /root/bin/telegram

COPY --from=builder /app/telegram/configs /root/configs/telegram

ENV TELEGRAM_CONFIG=/root/configs/telegram/config.yaml

EXPOSE 8555

ENTRYPOINT ["/root/bin/telegram"]