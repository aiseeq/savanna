package integration

import (
	"github.com/aiseeq/savanna/internal/core"
	"github.com/aiseeq/savanna/tests/common"
)

// Deprecated: Используйте common.IsWolfAttacking вместо этой функции
// isWolfAttacking общая вспомогательная функция для всех тестов (оставлена для обратной совместимости)
func isWolfAttacking(world *core.World, wolf core.EntityID) bool {
	return common.IsWolfAttacking(world, wolf)
}

// abs возвращает абсолютное значение float32
func abs(x float32) float32 {
	if x < 0 {
		return -x
	}
	return x
}
