#!/usr/bin/env python3
"""
LLM-Driven Development: Task Executor
Запускает агента-исполнителя и агента-проверяющего для выполнения задач
"""

import os
import sys
import subprocess
import json
import time
import re
from pathlib import Path
from datetime import datetime

def read_task(task_name):
    """Читает задачу из файла"""
    task_file = Path("docs/tasks") / f"{task_name}.md"
    if not task_file.exists():
        raise FileNotFoundError(f"Задача {task_name} не найдена в docs/tasks/")
    
    return task_file.read_text(encoding='utf-8')

def save_state(task_name, iteration, executor_chat, checker_output=None, last_checker_output=None):
    """Сохраняет текущее состояние выполнения задачи"""
    # Убеждаемся что директория llmdd существует
    Path("llmdd").mkdir(exist_ok=True)
    state_file = Path("llmdd") / f"{task_name}_state.json"
    state = {
        "task_name": task_name,
        "iteration": iteration,
        "executor_chat": executor_chat,
        "checker_output": checker_output,
        "last_checker_output": last_checker_output,  # Для восстановления feedback
        "timestamp": datetime.now().isoformat(),
        "status": "in_progress"
    }
    
    with open(state_file, 'w', encoding='utf-8') as f:
        json.dump(state, f, indent=2, ensure_ascii=False)

def load_state(task_name):
    """Загружает состояние выполнения задачи"""
    state_file = Path("llmdd") / f"{task_name}_state.json"
    if not state_file.exists():
        return None
    
    try:
        with open(state_file, 'r', encoding='utf-8') as f:
            return json.load(f)
    except (json.JSONDecodeError, IOError):
        return None

def clear_state(task_name):
    """Очищает состояние задачи (при успешном завершении)"""
    state_file = Path("llmdd") / f"{task_name}_state.json"
    if state_file.exists():
        state_file.unlink()

def mark_task_completed(task_name):
    """Отмечает задачу как завершенную"""
    state_file = Path("llmdd") / f"{task_name}_state.json"
    if state_file.exists():
        with open(state_file, 'r', encoding='utf-8') as f:
            state = json.load(f)
        state["status"] = "completed"
        state["completion_time"] = datetime.now().isoformat()
        with open(state_file, 'w', encoding='utf-8') as f:
            json.dump(state, f, indent=2, ensure_ascii=False)

def run_claude(prompt, continue_chat=None):
    """Запускает Claude CLI с заданным промптом"""
    cmd = ["claude", "--dangerously-skip-permissions", "-p", "--output-format", "json"]
    
    if continue_chat:
        cmd.extend(["--resume", continue_chat])
    
    # Записываем промпт во временный файл
    Path("llmdd").mkdir(exist_ok=True)
    prompt_file = Path("llmdd/temp_prompt.md")
    prompt_file.write_text(prompt, encoding='utf-8')
    
    try:
        # Запускаем Claude с потоковым выводом
        process = subprocess.Popen(
            cmd + [f"@{prompt_file}"],
            stdout=subprocess.PIPE,
            stderr=subprocess.PIPE,
            text=True,
            encoding='utf-8',
            bufsize=1,  # Построчная буферизация
            universal_newlines=True
        )
        
        # Ждем завершения процесса и читаем весь вывод сразу
        stdout, stderr = process.communicate()
        
        if process.returncode != 0:
            print(f"❌ Ошибка запуска Claude (код {process.returncode}): {stderr}")
            return None, None
            
        full_output = stdout
        
        # Парсим JSON ответ для получения session_id и content
        try:
            response_data = json.loads(full_output)
            chat_id = response_data.get('session_id')
            content = response_data.get('result', full_output)  # Используем 'result' вместо 'content'
            
            if chat_id:
                print(f"🔗 Найден Session ID: {chat_id}")
            else:
                print("⚠️ Session ID не найден в JSON ответе")
                
            # Выводим содержимое ответа
            print(content)
            
            return content, chat_id
            
        except json.JSONDecodeError:
            print("⚠️ Ответ не в JSON формате, используем как текст")
            print(full_output)
            return full_output, None
        
    finally:
        # Удаляем временный файл
        if prompt_file.exists():
            prompt_file.unlink()

def format_executor_prompt(task_content):
    """Формирует промпт для агента-исполнителя"""
    return f"""Ты агент-исполнитель в системе LLM-Driven Development. 

ЗАДАЧА ДЛЯ ВЫПОЛНЕНИЯ:
{task_content}

ПРАВИЛА:
1. Выполни задачу полностью и точно
2. НЕ объявляй о готовности до полного выполнения 
3. Если что-то не работает - исправляй до достижения результата
4. В конце работы создай файл llmdd/execution_complete.md с описанием что сделано

Начинай выполнение задачи."""

def format_checker_prompt(task_content):
    """Формирует промпт для агента-проверяющего"""
    return f"""Ты строгий агент-проверяющий в системе LLM-Driven Development.

ЗАДАЧА ДЛЯ ПРОВЕРКИ:
{task_content}

КРИТИЧЕСКИ ВАЖНО:
1. Прочитай и проанализируй КАЖДЫЙ критерий приемки в задаче
2. Проверь ВСЕ файлы результата (скриншоты, отчеты)
3. Сравни фактический результат с КАЖДЫМ требованием
4. НЕ доверяй утверждениям из отчета исполнителя - проверяй САМОСТОЯТЕЛЬНО
5. Если хотя бы ОДИН критерий не выполнен - задача НЕ ВЫПОЛНЕНА

МЕТОДОЛОГИЯ ПРОВЕРКИ:
- Открой и детально изучи все созданные файлы
- Сверь каждый пункт "Критерии приемки" с фактическим результатом
- Игнорируй красивые описания в отчете - смотри только на реальный результат
- При малейших сомнениях - задача НЕ ВЫПОЛНЕНА

ФОРМАТ ОТВЕТА (ОБЯЗАТЕЛЬНО ИСПОЛЬЗУЙ):

ЕСЛИ ВСЕ КРИТЕРИИ ВЫПОЛНЕНЫ:
"ЗАДАЧА ВЫПОЛНЕНА: [детальное подтверждение каждого критерия приемки]"

ЕСЛИ ХОТЯ БЫ ОДИН КРИТЕРИЙ НЕ ВЫПОЛНЕН:
"ЗАДАЧА НЕ ВЫПОЛНЕНА: [точное указание каждого нарушенного критерия]"

ПРИМЕР АНАЛИЗА:
✅ Критерий 1: [проверен, выполнен/не выполнен]
✅/❌ Критерий 2: [проверен, выполнен/не выполнен]
✅/❌ Критерий 3: [проверен, выполнен/не выполнен]

Будь максимально придирчивым и объективным. Начинай проверку."""

def format_feedback_prompt(checker_response):
    """Формирует промпт с обратной связью от проверяющего"""
    return f"""Агент-проверяющий проанализировал результат твоей работы:

{checker_response}

Продолжи работу над задачей с учетом этой обратной связи."""

def main():
    if len(sys.argv) < 2 or len(sys.argv) > 3:
        print("Использование: python task_executor.py <имя_задачи> [--restart]")
        print("Пример: python task_executor.py center_animal_screenshot")
        print("Флаги:")
        print("  --restart    Принудительно начать задачу заново (игнорировать сохраненное состояние)")
        sys.exit(1)
    
    task_name = sys.argv[1]
    force_restart = len(sys.argv) == 3 and sys.argv[2] == "--restart"
    
    try:
        # Читаем задачу
        task_content = read_task(task_name)
        
        print("=" * 70)
        print("🚀 LLM-DRIVEN DEVELOPMENT SYSTEM")
        print("=" * 70)
        print(f"📋 Задача: {task_name}")
        print("🔧 Режим: Агент-исполнитель + Агент-проверяющий")
        print("⚙️  Максимум итераций: 10")
        
        # Проверяем, есть ли сохраненное состояние
        saved_state = load_state(task_name)
        if not force_restart and saved_state and saved_state.get("status") == "in_progress":
            print("🔄 НАЙДЕНО СОХРАНЕННОЕ СОСТОЯНИЕ")
            print(f"📅 Последняя активность: {saved_state['timestamp']}")
            print(f"🔢 Прерванная итерация: {saved_state['iteration'] + 1}")
            print("🚀 Продолжаем выполнение...")
            
            # Если проверяющий уже работал, переходим к следующей итерации
            if saved_state.get('checker_output'):
                start_iteration = saved_state['iteration'] + 1
                checker_output = None  # Сбрасываем для новой итерации
            else:
                start_iteration = saved_state['iteration']
                checker_output = saved_state.get('checker_output')
            
            executor_chat = saved_state['executor_chat']
        else:
            if force_restart:
                print("🔄 ПРИНУДИТЕЛЬНЫЙ ПЕРЕЗАПУСК")
                clear_state(task_name)
            else:
                print("🆕 НОВОЕ ВЫПОЛНЕНИЕ ЗАДАЧИ")
                
            # Убираем предыдущий отчет только при новом запуске
            execution_report = Path("llmdd/execution_complete.md")
            if execution_report.exists():
                execution_report.unlink()
                print("🗑️  Удален предыдущий отчет выполнения")
            
            start_iteration = 0
            executor_chat = None
            checker_output = None
        
        print("=" * 70)
        max_iterations = 10
        
        for iteration in range(start_iteration, max_iterations):
            print(f"\n{'='*60}")
            print(f"📋 ИТЕРАЦИЯ {iteration + 1}/{max_iterations}")
            print(f"{'='*60}")
            
            # Запускаем агента-исполнителя 
            should_run_executor = False
            
            if iteration == 0:
                # Первая итерация - всегда запускаем исполнителя
                should_run_executor = True
                print("\n🔧 АГЕНТ-ИСПОЛНИТЕЛЬ: Начинаю выполнение задачи...")
                print("─" * 50)
                executor_prompt = format_executor_prompt(task_content)
                executor_output, executor_chat = run_claude(executor_prompt)
                
            elif iteration > 0:
                # Последующие итерации - исполнитель исправляет недочеты
                should_run_executor = True
                print("\n🔧 АГЕНТ-ИСПОЛНИТЕЛЬ: Исправляю недочеты...")
                print("─" * 50)
                
                # Используем последний вывод проверяющего для feedback
                if checker_output or (saved_state and saved_state.get('last_checker_output')):
                    feedback_source = checker_output or saved_state.get('last_checker_output')
                    feedback_prompt = format_feedback_prompt(feedback_source)
                    executor_output, _ = run_claude(feedback_prompt, executor_chat)
                else:
                    print("⚠️ Нет обратной связи от проверяющего, начинаю заново...")
                    executor_prompt = format_executor_prompt(task_content)
                    executor_output, executor_chat = run_claude(executor_prompt)
            
            if should_run_executor:
                if not executor_output:
                    print("\n❌ ОШИБКА: Не удалось запустить агента-исполнителя")
                    print("💡 Проверьте подключение к интернету и доступность Claude CLI")
                    break
                
                # Сохраняем состояние после исполнителя
                save_state(task_name, iteration, executor_chat)
                print(f"\n✅ АГЕНТ-ИСПОЛНИТЕЛЬ: Работа завершена")
            
            # Запускаем агента-проверяющего (всегда выполняется)
            print("\n🔍 АГЕНТ-ПРОВЕРЯЮЩИЙ: Начинаю проверку результата...")
            print("─" * 50)
            checker_prompt = format_checker_prompt(task_content)
            checker_output, checker_chat = run_claude(checker_prompt)
            
            if not checker_output:
                print("\n❌ ОШИБКА: Не удалось запустить агента-проверяющего")
                print("💡 Проверьте подключение к интернету и доступность Claude CLI")
                break
            
            # Сохраняем состояние после проверяющего
            save_state(task_name, iteration, executor_chat, checker_output)
            print(f"\n✅ АГЕНТ-ПРОВЕРЯЮЩИЙ: Анализ завершен")
            
            # Анализируем ответ проверяющего
            print(f"\n{'='*60}")
            print("📊 РЕЗУЛЬТАТ ПРОВЕРКИ")
            print(f"{'='*60}")
            
            if "ЗАДАЧА ВЫПОЛНЕНА" in checker_output:
                print("\n🎉 УСПЕХ! Задача выполнена успешно")
                print("─" * 50)
                print("✅ Все требования выполнены! Агент-проверяющий подтвердил корректность результата.")
                print("─" * 50)
                
                # Отмечаем задачу как завершенную и очищаем состояние
                mark_task_completed(task_name)
                clear_state(task_name)
                print("🗑️  Состояние выполнения очищено")
                break
            elif "ЗАДАЧА НЕ ВЫПОЛНЕНА" in checker_output:
                print("\n⚠️  ТРЕБУЕТСЯ ДОРАБОТКА")
                print("─" * 50)
                
                if iteration == max_iterations - 1:
                    print("❌ ДОСТИГНУТО МАКСИМАЛЬНОЕ КОЛИЧЕСТВО ИТЕРАЦИЙ")
                    print("📋 Финальные замечания проверяющего:")
                    print(checker_output)
                    print("─" * 50)
                    print("💡 Рекомендация: уточните требования или разделите задачу")
                    break
                else:
                    print("🔄 Передаю обратную связь агенту-исполнителю для исправления...")
                    # Сохраняем замечания для следующей итерации, но не показываем сейчас
                    save_state(task_name, iteration, executor_chat, checker_output, checker_output)
            else:
                print("\n⚠️  НЕОЖИДАННЫЙ ОТВЕТ ОТ ПРОВЕРЯЮЩЕГО")
                print("─" * 50)
                print(checker_output)
                print("─" * 50)
                break
        
    except FileNotFoundError as e:
        print(f"❌ {e}")
        sys.exit(1)
    except KeyboardInterrupt:
        print("\n⏹️  Выполнение прервано пользователем")
        print("💾 Состояние сохранено, можно продолжить выполнение позже")
        sys.exit(0)
    except Exception as e:
        print(f"❌ Неожиданная ошибка: {e}")
        import traceback
        print(f"📊 Подробности: {traceback.format_exc()}")
        sys.exit(1)

if __name__ == "__main__":
    main()