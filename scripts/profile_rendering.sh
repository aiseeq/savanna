#!/bin/bash

# Скрипт для профилирования производительности рендеринга

echo "🔍 Анализ производительности рендеринга Savanna"
echo "=============================================="

# Проверяем что игра скомпилирована
if [ ! -f "bin/savanna-game" ]; then
    echo "❌ Игра не скомпилирована. Запустите: make build"
    exit 1
fi

# Создаем директорию для профилей
mkdir -p profiles

echo "🚀 Запуск игры с профилированием..."
echo "   Игра будет запущена на 30 секунд для сбора данных"

# Запускаем игру в фоне с профилированием
./bin/savanna-game -pprof &
GAME_PID=$!

# Ждем немного чтобы игра запустилась
sleep 3

echo "📊 Сбор CPU профиля (30 секунд)..."
# Собираем CPU профиль
go tool pprof -seconds=30 -output=profiles/cpu.prof http://localhost:6060/debug/pprof/profile &
PPROF_PID=$!

# Ждем завершения профилирования
wait $PPROF_PID

echo "💾 Сбор профиля памяти..."
# Собираем профиль памяти
go tool pprof -output=profiles/memory.prof http://localhost:6060/debug/pprof/heap

# Останавливаем игру
kill $GAME_PID 2>/dev/null
wait $GAME_PID 2>/dev/null

echo "📈 Анализ результатов..."

# Анализируем CPU профиль
echo ""
echo "=== TOP CPU потребители ==="
go tool pprof -top -cum profiles/cpu.prof

echo ""
echo "=== TOP функции рендеринга ==="
go tool pprof -top -cum profiles/cpu.prof | grep -E "(render|draw|Draw|Render|vector|StrokeLine)"

echo ""
echo "=== Профили сохранены в profiles/ ==="
echo "📂 CPU профиль: profiles/cpu.prof"
echo "📂 Память: profiles/memory.prof"

echo ""
echo "🔧 Для интерактивного анализа:"
echo "   go tool pprof profiles/cpu.prof"
echo "   (команды: top, list, web, svg)"

echo ""
echo "📊 Для веб-интерфейса:"
echo "   go tool pprof -http=:8080 profiles/cpu.prof"