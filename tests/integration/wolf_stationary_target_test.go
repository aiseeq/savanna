package integration

import (
	"math"
	"testing"

	"github.com/aiseeq/savanna/config"
	"github.com/aiseeq/savanna/internal/core"
	"github.com/aiseeq/savanna/internal/generator"
	"github.com/aiseeq/savanna/internal/simulation"
)

// TestWolfVsStationaryTarget проверяет поведение волка с полностью неподвижной целью
func TestWolfVsStationaryTarget(t *testing.T) {
	t.Parallel()
	cfg := &config.Config{
		World: config.WorldConfig{Size: 10, Seed: 123},
	}

	terrainGen := generator.NewTerrainGenerator(cfg)
	terrain := terrainGen.Generate()

	world := core.NewWorld(320, 320, 123)
	vegetationSystem := simulation.NewVegetationSystem(terrain)
	animalBehaviorSystem := simulation.NewAnimalBehaviorSystem(vegetationSystem)
	movementSystem := simulation.NewMovementSystem(320, 320)

	// Создаем животных дальше от границ (мир 320x320, ставим в центр)
	rabbit := simulation.CreateAnimal(world, core.TypeRabbit, 160, 160)
	wolf := simulation.CreateAnimal(world, core.TypeWolf, 140, 160) // 20 единиц от зайца

	// Делаем волка голодным
	world.SetHunger(wolf, core.Hunger{Value: 20.0})

	t.Logf("=== Тест с неподвижной целью ===")

	deltaTime := float32(1.0 / 60.0)

	for i := 0; i < 120; i++ { // 2 секунды
		// Принудительно фиксируем зайца на месте каждый кадр ПЕРЕД обновлением
		world.SetPosition(rabbit, core.Position{X: 160, Y: 160})
		world.SetVelocity(rabbit, core.Velocity{X: 0, Y: 0})
		world.SetSpeed(rabbit, core.Speed{Base: 0, Current: 0}) // Важно: скорость тоже 0

		// Обновляем системы
		animalBehaviorSystem.Update(world, deltaTime)
		movementSystem.Update(world, deltaTime)
		world.Update(deltaTime)

		// ПОСЛЕ обновления получаем позиции для логирования
		wolfPos, _ := world.GetPosition(wolf)
		rabbitPos, _ := world.GetPosition(rabbit)
		wolfVel, _ := world.GetVelocity(wolf)

		dx := wolfPos.X - rabbitPos.X
		dy := wolfPos.Y - rabbitPos.Y
		distance := math.Sqrt(float64(dx*dx + dy*dy))

		if i%20 == 0 {
			t.Logf("%.1fс: волк (%.1f,%.1f) vel(%.1f,%.1f) | дистанция %.1f",
				float32(i)/60.0, wolfPos.X, wolfPos.Y, wolfVel.X, wolfVel.Y, distance)
		}

		// Проверяем перепрыгивание
		if wolfPos.X > 165 && i > 30 { // Если волк далеко за зайцем
			t.Errorf("ПЕРЕПРЫГИВАНИЕ: Волк (%.1f,%.1f) ушел далеко за неподвижного зайца (160,160) на тике %d",
				wolfPos.X, wolfPos.Y, i)
			break
		}
	}
}
