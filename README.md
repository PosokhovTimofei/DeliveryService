# DeliveryService

## Gateway
1. Клиент -> POST http://localhost:8228/api/packages -> Gateway

2. Gateway -> POST http://localhost:1234/packages -> Producer Service

3. Producer Service -> Gateway -> Клиент

## Calculator
Route -> http://localhost:8121/calculator

## Dependences
```
Kafka
go get github.com/IBM/sarama

Logrus
go get github.com/sirupsen/logrus

Yaml
go get gopkg.in/yaml.v3

UUID
go get github.com/google/uuid

Mongo
go get go.mongodb.org/mongo-driver

Mux
github.com/gorilla/mux

gRPC
go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest
go get google.golang.org/grpc

PostgreSQL
github.com/jackc/pgx/v4
github.com/jackc/pgx/v4/pgxpool
```