#!/bin/bash

# Mutation Testing Script для проекта Savanna
# Проверяет качество тестов путём внесения багов в код и проверки падают ли тесты

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"
MUTATION_TOOL="go-mutesting"

echo "🧬 Mutation Testing для проекта Savanna"
echo "========================================"

# Проверяем установлен ли go-mutesting
if ! command -v $MUTATION_TOOL &> /dev/null; then
    echo "📦 Устанавливаем go-mutesting..."
    go install github.com/zimmski/go-mutesting/cmd/go-mutesting@latest
    
    if ! command -v $MUTATION_TOOL &> /dev/null; then
        echo "❌ Не удалось установить go-mutesting"
        echo "Попробуйте установить вручную:"
        echo "go install github.com/zimmski/go-mutesting/cmd/go-mutesting@latest"
        exit 1
    fi
fi

cd "$PROJECT_ROOT"

# Функция для запуска mutation testing на конкретном пакете
run_mutation_test() {
    local package=$1
    local description=$2
    
    echo ""
    echo "🔬 Тестируем: $description"
    echo "📁 Пакет: $package"
    echo "----------------------------"
    
    # Создаём временную директорию для результатов
    local results_dir="./mutation_results/$(basename "$package")"
    mkdir -p "$results_dir"
    
    # Запускаем mutation testing
    echo "Запускаем mutation testing..."
    local mutation_output="$results_dir/mutation_output.txt"
    
    if $MUTATION_TOOL --disable=branch --output="$results_dir" "$package" > "$mutation_output" 2>&1; then
        echo "✅ Mutation testing завершён"
    else
        echo "⚠️  Mutation testing завершён с предупреждениями"
    fi
    
    # Анализируем результаты
    local total_mutations=0
    local killed_mutations=0
    
    if [[ -f "$mutation_output" ]]; then
        total_mutations=$(grep -c "PASS\|FAIL" "$mutation_output" || echo "0")
        killed_mutations=$(grep -c "FAIL" "$mutation_output" || echo "0")
    fi
    
    if [[ $total_mutations -gt 0 ]]; then
        local survival_rate=$(( (total_mutations - killed_mutations) * 100 / total_mutations ))
        local kill_rate=$(( killed_mutations * 100 / total_mutations ))
        
        echo "📊 Результаты:"
        echo "   Всего мутаций: $total_mutations"
        echo "   Убито тестами: $killed_mutations"
        echo "   Выжило: $((total_mutations - killed_mutations))"
        echo "   Kill Rate: ${kill_rate}% (чем выше, тем лучше)"
        echo "   Survival Rate: ${survival_rate}% (чем ниже, тем лучше)"
        
        # Оценка качества тестов
        if [[ $kill_rate -ge 80 ]]; then
            echo "🏆 Отличное качество тестов!"
        elif [[ $kill_rate -ge 60 ]]; then
            echo "👍 Хорошее качество тестов"
        elif [[ $kill_rate -ge 40 ]]; then
            echo "⚠️  Среднее качество тестов - есть что улучшить"
        else
            echo "❌ Плохое качество тестов - требуется доработка"
        fi
        
        # Если есть выжившие мутации, показываем их
        if [[ $survival_rate -gt 20 ]]; then
            echo ""
            echo "🔍 Выжившие мутации (указывают на пробелы в тестах):"
            grep "PASS" "$mutation_output" | head -5 || true
            if [[ $(grep -c "PASS" "$mutation_output" || echo "0") -gt 5 ]]; then
                echo "   ... и ещё $(($(grep -c "PASS" "$mutation_output") - 5)) мутаций"
            fi
        fi
    else
        echo "❌ Не удалось проанализировать результаты"
    fi
}

# Основные пакеты для тестирования
echo "🎯 Выбор критически важных пакетов для mutation testing..."

# 1. Система питания (критически важна после недавних багов)
run_mutation_test "./internal/simulation" "Симуляция животных (включая питание)"

# 2. Ядро ECS системы
run_mutation_test "./internal/core" "ECS ядро системы"

# 3. Физика и движение
run_mutation_test "./internal/physics" "Физика и пространственные вычисления"

# 4. Генерация мира
run_mutation_test "./internal/generator" "Генерация terrain и мира"

echo ""
echo "📋 Общий отчёт по Mutation Testing"
echo "=================================="

# Подсчитываем общую статистику
total_packages=0
good_packages=0
bad_packages=0

for result_dir in ./mutation_results/*/; do
    if [[ -d "$result_dir" ]]; then
        total_packages=$((total_packages + 1))
        
        output_file="$result_dir/mutation_output.txt"
        if [[ -f "$output_file" ]]; then
            total_mutations=$(grep -c "PASS\|FAIL" "$output_file" || echo "0")
            killed_mutations=$(grep -c "FAIL" "$output_file" || echo "0")
            
            if [[ $total_mutations -gt 0 ]]; then
                kill_rate=$(( killed_mutations * 100 / total_mutations ))
                if [[ $kill_rate -ge 60 ]]; then
                    good_packages=$((good_packages + 1))
                else
                    bad_packages=$((bad_packages + 1))
                fi
            fi
        fi
    fi
done

echo "📊 Итоги:"
echo "   Протестировано пакетов: $total_packages"
echo "   С хорошим качеством тестов: $good_packages"
echo "   Требуют улучшения: $bad_packages"

if [[ $bad_packages -eq 0 ]]; then
    echo "🎉 Все пакеты имеют хорошее качество тестов!"
    exit 0
elif [[ $bad_packages -le 1 ]]; then
    echo "👍 Общее качество тестов хорошее, есть небольшие улучшения"
    exit 0
else
    echo "⚠️  Есть пакеты с плохим качеством тестов - рекомендуется доработка"
    exit 1
fi