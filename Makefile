BIN_DIR      := bin
GO           := go

.PHONY: client gateway producer consumer test

gateway:
	@echo "🚀 Запуск API Gateway..."
	@$(GO) build -o $(BIN_DIR)/gateway ./gateway/cmd/gateway
	@$(BIN_DIR)/gateway

client:
	@echo "🚀 Запуск клиента..."
	@$(GO) build -o $(BIN_DIR)/client ./client/presentation/cli
	@$(BIN_DIR)/client $(ARGS)

producer:
	@echo "🚀 Запуск producer..."
	@$(GO) build -o $(BIN_DIR)/producer ./producer/cmd/
	@$(BIN_DIR)/producer

consumer:
	@echo "🚀 Запуск consumer..."
	@$(GO) build -o $(BIN_DIR)/consumer ./consumer/cmd/
	@$(BIN_DIR)/consumer


test:
	@echo "🧪 Запуск тестов..."
	@$(GO) test -v -race ./...