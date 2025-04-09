# DeliveryService

## Gateway
1. Клиент -> POST http://localhost:8228/api/packages -> Gateway

2. Gateway -> POST http://localhost:1234/packages -> Producer Service

3. Producer Service -> Gateway -> Клиент

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
```