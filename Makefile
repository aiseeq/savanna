# Makefile для проекта Savanna - симулятор экосистемы саванны

.PHONY: build build-with-lint build-fast run run-animviewer run-visual-test test test-unit test-perf bench fmt lint lint-install lint-fix check generate profile simulate balance clean help

# Основные команды
build: ## Собрать все версии (без линтинга)
	@echo "Сборка основной игры..."
	go build -buildvcs=false -o bin/savanna-game ./cmd/game
	@echo "Сборка просмотрщика анимаций..."
	go build -buildvcs=false -o bin/savanna-animviewer ./cmd/animviewer
	@echo "Сборка отладчика системы питания..."

build-with-lint: lint ## Собрать с проверкой линтера
	@echo "Сборка с линтингом..."
	go build -buildvcs=false -o bin/savanna-game ./cmd/game
	go build -buildvcs=false -o bin/savanna-animviewer ./cmd/animviewer
	@echo "Сборка с линтингом завершена"

build-fast: ## Собрать без линтинга (быстро)
	@echo "Быстрая сборка без проверок..."
	go build -buildvcs=false -o bin/savanna-game ./cmd/game
	go build -buildvcs=false -o bin/savanna-animviewer ./cmd/animviewer
	@echo "Быстрая сборка завершена"

build-windows: ## Собрать для Windows с отключенным DPI awareness
	@echo "Сборка для Windows с отключенным DPI awareness..."
	cd cmd/game && x86_64-w64-mingw32-windres resource.rc -o resource.syso 2>/dev/null || echo "windres не найден, пропускаем manifest"
	GOOS=windows GOARCH=amd64 go build -buildvcs=false -o bin/savanna-game.exe ./cmd/game
	@echo "Windows сборка завершена"

run: build ## Запустить GUI версию
	@echo "Запуск GUI версии (с оптимизациями для WSL и отключением DPI scaling)..."
	DISPLAY=:0 MIT_SHM=0 LIBGL_ALWAYS_SOFTWARE=1 GDK_SCALE=1 GDK_DPI_SCALE=1 QT_AUTO_SCREEN_SCALE_FACTOR=0 QT_SCALE_FACTOR=1 QT_SCREEN_SCALE_FACTORS=1 XCURSOR_SIZE=16 EBITEN_GRAPHICS_LIBRARY=opengl XFORCEDPI=96 ./bin/savanna-game


run-animviewer: build ## Запустить просмотрщик анимаций
	@echo "Запуск просмотрщика анимаций..."
	DISPLAY=:0 MIT_SHM=0 LIBGL_ALWAYS_SOFTWARE=1 GDK_SCALE=1 GDK_DPI_SCALE=1 QT_AUTO_SCREEN_SCALE_FACTOR=0 QT_SCALE_FACTOR=1 QT_SCREEN_SCALE_FACTORS=1 XCURSOR_SIZE=16 EBITEN_GRAPHICS_LIBRARY=opengl XFORCEDPI=96 ./bin/savanna-animviewer --show wolf

run-visual-test: build ## Запустить визуальный тестовый режим
	@echo "Запуск визуального тестового режима..."
	DISPLAY=:99 MIT_SHM=0 LIBGL_ALWAYS_SOFTWARE=1 GDK_SCALE=1 GDK_DPI_SCALE=1 QT_AUTO_SCREEN_SCALE_FACTOR=0 QT_SCALE_FACTOR=1 QT_SCREEN_SCALE_FACTORS=1 XCURSOR_SIZE=16 EBITEN_GRAPHICS_LIBRARY=opengl XFORCEDPI=96 ./bin/savanna-game --visual-test

test: ## Все тесты с виртуальным дисплеем
	@echo "🧪 Запуск всех тестов с виртуальным дисплеем..."
	xvfb-run -a go test ./...

test-advanced: ## Все продвинутые тесты (property, contract, behavioral, chaos)
	@echo "🧪 Запуск продвинутых тестов с виртуальным дисплеем..."
	xvfb-run -a go test ./tests/property/... -v
	xvfb-run -a go test ./tests/contract/... -v
	xvfb-run -a go test ./tests/behavioral/... -v
	xvfb-run -a go test ./tests/chaos/... -v


test-gui: ## E2E тесты с виртуальным дисплеем
	@echo "🖥️ Запуск E2E тестов с виртуальным дисплеем..."
	xvfb-run -a go test ./tests/e2e/... -v -timeout=60s

test-all: test test-advanced test-gui ## Все тесты включая GUI

test-unit: ## Только unit тесты (без виртуального дисплея)
	@echo "🧪 Запуск unit тестов..."
	go test ./tests/unit/... -v

test-property: ## Property-based тесты
	go test ./tests/property/... -v

test-contract: ## Contract тесты
	go test ./tests/contract/... -v

test-behavioral: ## Behavioral тесты (Given-When-Then)
	go test ./tests/behavioral/... -v

test-chaos: ## Chaos Engineering тесты
	go test ./tests/chaos/... -v

test-mutation: ## Mutation testing (проверка качества тестов)
	@echo "🧬 Запуск Mutation Testing..."
	./scripts/mutation_testing.sh

test-perf: ## Тесты производительности
	go test ./tests/performance/... -v

bench: ## Бенчмарки
	go test -bench=. ./...

# Разработка и качество кода
fmt: ## Форматирование кода
	@echo "Форматирование кода..."
	go fmt ./...

lint-install: ## Установить golangci-lint
	@echo "Установка golangci-lint..."
	@if ! command -v golangci-lint >/dev/null 2>&1; then \
		echo "Скачиваем golangci-lint..."; \
		curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $$(go env GOPATH)/bin v1.54.2; \
	else \
		echo "golangci-lint уже установлен"; \
	fi

lint: ## Линтер (автоустановка если нужно)
	@GOPATH=$$(go env GOPATH); \
	LINTER=$$GOPATH/bin/golangci-lint; \
	if [ ! -f "$$LINTER" ]; then \
		echo "golangci-lint не найден, устанавливаем..."; \
		$(MAKE) lint-install; \
	fi; \
	echo "Запуск линтера..."; \
	$$LINTER run

lint-fix: ## Автоисправление проблем линтера
	@GOPATH=$$(go env GOPATH); \
	LINTER=$$GOPATH/bin/golangci-lint; \
	if [ ! -f "$$LINTER" ]; then \
		$(MAKE) lint-install; \
	fi; \
	echo "Автоисправление проблем линтера..."; \
	$$LINTER run --fix

check: fmt lint test ## Полная проверка кода (форматирование + линтер + тесты)

generate: ## Генерация кода
	go generate ./...

profile: build ## Запуск с профилированием
	./bin/savanna-game --pprof
	@echo "Для просмотра профиля: go tool pprof http://localhost:6060/debug/pprof/profile"

# Симуляция
simulate: build ## Запуск симуляции
	./bin/savanna-game

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