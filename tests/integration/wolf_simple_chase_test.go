package integration

import (
	"math"
	"testing"

	"github.com/aiseeq/savanna/config"
	"github.com/aiseeq/savanna/internal/core"
	"github.com/aiseeq/savanna/internal/generator"
	"github.com/aiseeq/savanna/internal/simulation"
)

// TestWolfSimpleChase проверяет что волк правильно преследует неподвижного зайца
func TestWolfSimpleChase(t *testing.T) {
	t.Parallel()
	// Создаем минимальную симуляцию
	cfg := &config.Config{
		World: config.WorldConfig{
			Size: 10,
			Seed: 42,
		},
	}

	terrainGen := generator.NewTerrainGenerator(cfg)
	terrain := terrainGen.Generate()

	worldSizePixels := float32(320)
	world := core.NewWorld(worldSizePixels, worldSizePixels, 42)

	// Создаем системы
	vegetationSystem := simulation.NewVegetationSystem(terrain)
	animalBehaviorSystem := simulation.NewAnimalBehaviorSystem(vegetationSystem)
	movementSystem := simulation.NewMovementSystem(worldSizePixels, worldSizePixels)

	// Создаем животных
	rabbit := simulation.CreateAnimal(world, core.TypeRabbit, 160, 160) // Центр
	wolf := simulation.CreateAnimal(world, core.TypeWolf, 140, 160)     // Слева от зайца

	// Волка делаем голодным
	world.SetSatiation(wolf, core.Satiation{Value: 20.0})

	// Зайца делаем сытым чтобы он не убегал сильно
	world.SetSatiation(rabbit, core.Satiation{Value: 90.0})

	t.Logf("=== Простой тест преследования ===")
	t.Logf("Начальные позиции: волк (140, 160), заяц (160, 160)")

	deltaTime := float32(1.0 / 60.0)

	for i := 0; i < 180; i++ { // 3 секунды
		wolfPos, _ := world.GetPosition(wolf)
		rabbitPos, _ := world.GetPosition(rabbit)
		wolfVel, _ := world.GetVelocity(wolf)

		dx := wolfPos.X - rabbitPos.X
		dy := wolfPos.Y - rabbitPos.Y
		distance := math.Sqrt(float64(dx*dx + dy*dy))

		if i%30 == 0 { // Каждые 0.5 секунды
			t.Logf("%.1fс: волк (%.1f,%.1f) vel(%.1f,%.1f) | заяц (%.1f,%.1f) | дистанция %.1f",
				float32(i)/60.0, wolfPos.X, wolfPos.Y, wolfVel.X, wolfVel.Y, rabbitPos.X, rabbitPos.Y, distance)
		}

		// Проверяем не перепрыгнул ли волк (если заяц почти не двигается)
		rabbitVel, _ := world.GetVelocity(rabbit)
		rabbitSpeed := math.Sqrt(float64(rabbitVel.X*rabbitVel.X + rabbitVel.Y*rabbitVel.Y))

		if rabbitSpeed < 1.0 && wolfPos.X > rabbitPos.X+5 && i > 60 {
			t.Errorf("ОШИБКА: Волк перепрыгнул медленного зайца на тике %d!", i)
			t.Errorf("  Волк: (%.1f, %.1f), Заяц: (%.1f, %.1f), Скорость зайца: %.1f",
				wolfPos.X, wolfPos.Y, rabbitPos.X, rabbitPos.Y, rabbitSpeed)
			break
		}

		if !world.IsAlive(rabbit) {
			t.Logf("Заяц умер на тике %d", i)
			break
		}

		// Обновляем симуляцию
		animalBehaviorSystem.Update(world, deltaTime)
		movementSystem.Update(world, deltaTime)
		world.Update(deltaTime)
	}
}
