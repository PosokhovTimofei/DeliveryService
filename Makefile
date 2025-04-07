BIN_DIR      := bin
CLIENT_NAME  := delivery-client
GO           := go

build: ## make client
	@echo "🔨 Сборка клиента..."
	@mkdir -p $(BIN_DIR)
	@$(GO) build -o $(BIN_DIR)/$(CLIENT_NAME) ./client/presentation/cli

## make run ARGS="create 5.2 Москва 'Санкт-Петербург' 'ул. Ленина, 1'"
run: build ## Запустить клиент
	@echo "🚀 Запуск клиента..."
	@$(BIN_DIR)/$(CLIENT_NAME) $(ARGS)

test:
	@echo "🧪 Запуск тестов..."
	@$(GO) test -v -race ./...

clean:
	@echo "🧹 Очистка..."
	@rm -rf $(BIN_DIR)