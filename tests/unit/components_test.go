package unit

import (
	"testing"

	"github.com/aiseeq/savanna/internal/core"
)

// TestComponentPosition тестирует работу с компонентом Position
func TestComponentPosition(t *testing.T) {
	t.Parallel()

	world := core.NewWorld(100, 100, 42) //nolint:gomnd // Тестовые параметры мира
	entity := world.CreateEntity()

	// Проверяем что компонента изначально нет
	if world.HasComponent(entity, core.MaskPosition) {
		t.Error("Entity should not have Position initially")
	}

	_, ok := world.GetPosition(entity)
	if ok {
		t.Error("GetPosition should return false for non-existent component")
	}

	// Добавляем компонент
	pos := core.Position{X: 10, Y: 20} //nolint:gomnd // Тестовые координаты
	if !world.AddPosition(entity, pos) {
		t.Error("AddPosition should succeed")
	}

	if !world.HasComponent(entity, core.MaskPosition) {
		t.Error("Entity should have Position after adding")
	}

	// Получаем компонент
	retrievedPos, ok := world.GetPosition(entity)
	if !ok {
		t.Error("GetPosition should return true for existing component")
	}

	if retrievedPos.X != pos.X || retrievedPos.Y != pos.Y {
		t.Errorf("Expected position %+v, got %+v", pos, retrievedPos)
	}

	// Изменяем позицию
	newPos := core.Position{X: 30, Y: 40} //nolint:gomnd // Тестовые координаты
	if !world.SetPosition(entity, newPos) {
		t.Error("SetPosition should succeed")
	}

	retrievedPos, _ = world.GetPosition(entity)
	if retrievedPos.X != newPos.X || retrievedPos.Y != newPos.Y {
		t.Errorf("Expected position %+v, got %+v", newPos, retrievedPos)
	}

	// Удаляем компонент
	if !world.RemovePosition(entity) {
		t.Error("RemovePosition should succeed")
	}

	if world.HasComponent(entity, core.MaskPosition) {
		t.Error("Entity should not have Position after removal")
	}
}

// TestComponentMultiple тестирует работу с несколькими компонентами
func TestComponentMultiple(t *testing.T) {
	t.Parallel()

	world := core.NewWorld(100, 100, 42) //nolint:gomnd // Тестовые параметры мира
	entity := world.CreateEntity()

	// Добавляем несколько компонентов
	world.AddPosition(entity, core.Position{X: 5, Y: 5})         //nolint:gomnd // Тестовые значения
	world.AddVelocity(entity, core.Velocity{X: 1, Y: 1})         //nolint:gomnd // Тестовые значения
	world.AddHealth(entity, core.Health{Current: 100, Max: 100}) //nolint:gomnd // Тестовые значения

	// Проверяем наличие компонентов
	if !world.HasComponents(entity, core.MaskPosition|core.MaskVelocity|core.MaskHealth) {
		t.Error("Entity should have all three components")
	}

	if world.HasComponents(entity, core.MaskPosition|core.MaskHunger) {
		t.Error("Entity should not have Hunger component")
	}

	// Проверяем что можем получить все компоненты
	pos, posOk := world.GetPosition(entity)
	vel, velOk := world.GetVelocity(entity)
	health, healthOk := world.GetHealth(entity)

	if !posOk || !velOk || !healthOk {
		t.Error("Should be able to get all added components")
	}

	expectedX := float32(5)
	expectedVelX := float32(1)
	expectedHealth := int16(100)

	if pos.X != expectedX || vel.X != expectedVelX || health.Current != expectedHealth {
		t.Error("Component values should match what was set")
	}
}

// TestComponentAnimalType тестирует работу с компонентом AnimalType
func TestComponentAnimalType(t *testing.T) {
	t.Parallel()

	world := core.NewWorld(100, 100, 42) //nolint:gomnd // Тестовые параметры мира
	entity := world.CreateEntity()
	world.AddAnimalType(entity, core.TypeRabbit)

	animalType, ok := world.GetAnimalType(entity)
	if !ok {
		t.Error("Should be able to get AnimalType")
	}

	if animalType != core.TypeRabbit {
		t.Errorf("Expected TypeRabbit, got %v", animalType)
	}

	// Проверяем строковое представление
	if animalType.String() != "Rabbit" {
		t.Errorf("Expected 'Rabbit', got '%s'", animalType.String())
	}
}
