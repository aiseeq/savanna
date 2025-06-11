# Makefile для проекта Savanna - симулятор экосистемы саванны

.PHONY: build run run-headless test test-unit test-perf bench fmt lint generate profile simulate balance clean help

# Основные команды
build: ## Собрать обе версии
	@echo "Сборка GUI версии..."
	go build -o bin/savanna-game ./cmd/game
	@echo "Сборка headless версии..."
	go build -o bin/savanna-headless ./cmd/headless
	@echo "Сборка завершена"

run: build ## Запустить GUI версию
	@echo "Запуск GUI версии (с оптимизациями для WSL)..."
	MIT_SHM=0 LIBGL_ALWAYS_SOFTWARE=1 ./bin/savanna-game

run-headless: build ## Запустить headless версию
	./bin/savanna-headless

test: ## Все тесты
	go test ./...

test-unit: ## Только unit тесты
	go test ./tests/unit/...

test-perf: ## Тесты производительности
	go test ./tests/performance/... -v

bench: ## Бенчмарки
	go test -bench=. ./...

# Разработка
fmt: ## Форматирование кода
	go fmt ./...

lint: ## Линтер
	@if command -v golangci-lint >/dev/null 2>&1; then \
		golangci-lint run; \
	else \
		echo "golangci-lint не установлен, используем go vet"; \
		go vet ./...; \
	fi

generate: ## Генерация кода
	go generate ./...

profile: build ## Запуск с профилированием
	./bin/savanna-headless -duration=30s -cpuprofile=cpu.prof
	@echo "Для просмотра профиля: go tool pprof cpu.prof"

# Симуляция
simulate: build ## Запуск headless симуляции
	./bin/savanna-headless -duration=60s

balance: build ## Тесты баланса
	@echo "Запуск тестов баланса экосистемы..."
	go test ./scripts/... -v

# Утилиты
clean: ## Очистить сборочные файлы
	rm -rf bin/
	rm -f *.prof

help: ## Показать эту справку
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-15s\033[0m %s\n", $$1, $$2}'

# По умолчанию показать справку
.DEFAULT_GOAL := help