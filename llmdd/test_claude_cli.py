#!/usr/bin/env python3
"""
Тестовый скрипт для проверки работы Claude CLI
"""

import subprocess
import sys
import json
from pathlib import Path

def test_claude_cli():
    """Проверяет работу Claude CLI с простой командой"""
    print("🧪 Тестирование Claude CLI...")
    
    try:
        # Создаем простой промпт
        prompt_file = Path("llmdd/test_prompt.md")
        prompt_file.parent.mkdir(exist_ok=True)
        prompt_file.write_text("Просто ответь 'Тест работает' и ничего больше", encoding='utf-8')
        
        # Запускаем Claude CLI с JSON выводом
        cmd = ["claude", "--dangerously-skip-permissions", "-p", "--output-format", "json", f"@{prompt_file}"]
        print(f"📋 Команда: {' '.join(cmd)}")
        
        # Используем communicate() вместо Popen для JSON формата
        result = subprocess.run(
            cmd,
            capture_output=True,
            text=True,
            encoding='utf-8'
        )
        
        stdout = result.stdout
        stderr = result.stderr
        
        print(f"🔢 Exit code: {result.returncode}")
        print(f"📤 STDOUT:")
        print(stdout)
        
        if stderr:
            print(f"⚠️ STDERR:")
            print(stderr)
            
        # Пробуем парсить JSON для извлечения session_id
        try:
            response_data = json.loads(stdout)
            session_id = response_data.get('session_id')
            content = response_data.get('content', stdout)
            print(f"🔗 Session ID: {session_id}")
            print(f"📄 Content: {content}")
        except:
            print("📝 Не удалось парсить как JSON")
        
        # Удаляем тестовый файл
        if prompt_file.exists():
            prompt_file.unlink()
            
        return result.returncode == 0, stdout
        
    except Exception as e:
        print(f"❌ Ошибка: {e}")
        return False, str(e)

if __name__ == "__main__":
    success, output = test_claude_cli()
    if success:
        print("✅ Claude CLI работает")
    else:
        print("❌ Claude CLI не работает")
        sys.exit(1)