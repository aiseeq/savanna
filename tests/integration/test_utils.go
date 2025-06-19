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
