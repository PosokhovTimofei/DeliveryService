# DeliveryService

## Gateway
1. Клиент -> POST /api/packages -> Gateway

2. Gateway -> POST http://localhost:1234/packages -> Producer Service

3. Producer Service -> Gateway -> Клиент