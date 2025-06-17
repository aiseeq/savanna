package integration

import "github.com/aiseeq/savanna/internal/core"

// Константы для тестов
const (
	WolfHungerThresholdTest = 60.0 // Порог голода волка для атаки в тестах
	WolfVisionRangeTest     = 15.0 // Дальность видения волка в тестах
)

// isWolfAttacking общая вспомогательная функция для всех тестов
func isWolfAttacking(world *core.World, wolf core.EntityID) bool {
	hunger, hasHunger := world.GetHunger(wolf)
	if !hasHunger || hunger.Value > WolfHungerThresholdTest {
		return false
	}

	pos, hasPos := world.GetPosition(wolf)
	if !hasPos {
		return false
	}

	nearestRabbit, foundRabbit := world.FindNearestByType(pos.X, pos.Y, WolfVisionRangeTest, core.TypeRabbit)
	if !foundRabbit {
		return false
	}

	rabbitPos, hasRabbitPos := world.GetPosition(nearestRabbit)
	if !hasRabbitPos {
		return false
	}

	distance := (pos.X-rabbitPos.X)*(pos.X-rabbitPos.X) + (pos.Y-rabbitPos.Y)*(pos.Y-rabbitPos.Y)
	return distance <= 12.0*12.0
}
