FROM golang:1.23-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN go build -o bin/auction ./auction/cmd/main.go

FROM alpine:3.18

WORKDIR /root
RUN apk add --no-cache ca-certificates

COPY --from=builder /app/bin/auction /root/bin/auction
RUN chmod +x /root/bin/auction

COPY --from=builder /app/auction/configs /root/configs/auction

ENV AUCTION_CONFIG=/root/configs/auction/config.yaml

EXPOSE 8777 50055

ENTRYPOINT ["/root/bin/auction"]