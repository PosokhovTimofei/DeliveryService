BIN_DIR      := bin
CLIENT_NAME  := delivery-client
GO           := go

.PHONY: build run test clean gateway

build:
	@echo "🔨 Сборка клиента..."
	@mkdir -p $(BIN_DIR)
	@$(GO) build -o $(BIN_DIR)/$(CLIENT_NAME) ./client/presentation/cli

## make run ARGS="create 5.2 Москва 'Санкт-Петербург' 'ул. Ленина, 1'"
run: build
	@echo "🚀 Запуск клиента..."
	@$(BIN_DIR)/$(CLIENT_NAME) $(ARGS)

gateway:
	@echo "🚀 Запуск API Gateway..."
	@$(GO) build -o $(BIN_DIR)/gateway ./gateway/cmd/gateway
	@$(BIN_DIR)/gateway

test:
	@echo "🧪 Запуск тестов..."
	@$(GO) test -v -race ./...

clean:
	@echo "🧹 Очистка..."
	@rm -rf $(BIN_DIR)