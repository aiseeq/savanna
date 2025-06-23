package integration

import (
	"testing"

	"github.com/aiseeq/savanna/internal/adapters"
	"github.com/aiseeq/savanna/internal/core"
	"github.com/aiseeq/savanna/internal/simulation"
)

// TestSimpleAnimalInteraction простой тест взаимодействия животных
func TestSimpleAnimalInteraction(t *testing.T) {
	t.Parallel()
	world := core.NewWorld(TestWorldSize, TestWorldSize, 42)
	systemManager := core.NewSystemManager()

	// Создаём vegetation систему для взаимодействия
	vegetationSystem := createTestVegetationSystem()
	systemManager.AddSystem(vegetationSystem)
	systemManager.AddSystem(&adapters.BehaviorSystemAdapter{
		System: simulation.NewAnimalBehaviorSystem(vegetationSystem),
	})
	systemManager.AddSystem(&adapters.MovementSystemAdapter{
		System: simulation.NewMovementSystem(TestWorldSize, TestWorldSize),
	})

	// Создаем зайца и волка очень близко для тайловой системы
	rabbit := simulation.CreateAnimal(world, core.TypeRabbit, 100, 100)
	wolf := simulation.CreateAnimal(world, core.TypeWolf, 101, 100) // Дистанция 1 пиксель

	// Делаем волка голодным чтобы он охотился
	world.SetHunger(wolf, core.Hunger{Value: 30.0}) // 30% < 60% порога

	initialRabbitPos, _ := world.GetPosition(rabbit)
	initialWolfPos, _ := world.GetPosition(wolf)

	// Запускаем симуляцию на 2 секунды
	deltaTime := float32(1.0 / TestTPS)
	for i := 0; i < TestTPS*2; i++ {
		world.Update(deltaTime)
		systemManager.Update(world, deltaTime)
	}

	// Проверяем финальные позиции
	finalRabbitPos, _ := world.GetPosition(rabbit)
	finalWolfPos, _ := world.GetPosition(wolf)

	rabbitMoved := finalRabbitPos.X != initialRabbitPos.X || finalRabbitPos.Y != initialRabbitPos.Y
	wolfMoved := finalWolfPos.X != initialWolfPos.X || finalWolfPos.Y != initialWolfPos.Y

	t.Logf("Начальные позиции: заяц (%.1f,%.1f), волк (%.1f,%.1f)",
		initialRabbitPos.X, initialRabbitPos.Y, initialWolfPos.X, initialWolfPos.Y)
	t.Logf("Финальные позиции: заяц (%.1f,%.1f), волк (%.1f,%.1f)",
		finalRabbitPos.X, finalRabbitPos.Y, finalWolfPos.X, finalWolfPos.Y)
	t.Logf("Движение: заяц %v, волк %v", rabbitMoved, wolfMoved)

	// Проверяем что животные живы
	if !world.IsAlive(rabbit) {
		t.Error("Rabbit should be alive")
	}
	if !world.IsAlive(wolf) {
		t.Error("Wolf should be alive")
	}

	// Проверяем что животные двигались
	if !rabbitMoved {
		t.Error("Rabbit should have moved")
	}
	if !wolfMoved {
		t.Error("Wolf should have moved")
	}

	// Вычисляем расстояние движения зайца
	rabbitMovement := calculateDistance(finalRabbitPos.X-initialRabbitPos.X, finalRabbitPos.Y-initialRabbitPos.Y, 0, 0)

	// Заяц должен двигаться больше чем на 1 пиксель за 2 секунды
	if rabbitMovement < 1.0 {
		t.Error("Rabbit should have moved when wolf is nearby")
	}
}
