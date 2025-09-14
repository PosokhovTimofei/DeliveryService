FROM golang:1.23-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN go build -o bin/package ./database/cmd/main.go

FROM alpine:3.18

WORKDIR /root
RUN apk add --no-cache ca-certificates

COPY --from=builder /app/bin/package /root/bin/package
RUN chmod +x /root/bin/package

COPY --from=builder /app/database/configs /root/configs/database

ENV PACKAGE_CONFIG=/root/configs/database/config.yaml

EXPOSE 8333 50054

ENTRYPOINT ["/root/bin/package"]