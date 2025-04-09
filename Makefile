BIN_DIR      := bin
GO           := go

.PHONY: client gateway producer test

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

test:
	@echo "🧪 Запуск тестов..."
	@$(GO) test -v -race ./...