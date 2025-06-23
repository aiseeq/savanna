#!/bin/bash
# LLM-Driven Development: Task Runner
# Запускает систему исполнения задач

cd "$(dirname "$0")/.."
python3 llmdd/task_executor.py "$@"