BIN_DIR      := bin
GO           := go

.PHONY: client gateway producer calculate consumer test

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

calculate:
	@echo "🚀 Запуск calculator..."
	@$(GO) build -o $(BIN_DIR)/calculate ./calculator/cmd/
	@$(BIN_DIR)/calculate

db:
	@echo "🚀 Запуск database..."
	@$(GO) build -o $(BIN_DIR)/db ./database/cmd/
	@$(BIN_DIR)/db


test:
	@echo "🧪 Запуск тестов..."
	@$(GO) test -v -race ./...