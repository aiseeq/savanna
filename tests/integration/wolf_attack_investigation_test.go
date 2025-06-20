package integration

import (
	"math"
	"testing"

	"github.com/aiseeq/savanna/config"
	"github.com/aiseeq/savanna/internal/adapters"
	"github.com/aiseeq/savanna/internal/core"
	"github.com/aiseeq/savanna/internal/generator"
	"github.com/aiseeq/savanna/internal/simulation"
)

// TestWolfAttackBehavior исследует поведение волка при атаке зайца
//
//nolint:revive // Интеграционный тест может быть длинным
func TestWolfAttackBehavior(t *testing.T) {
	t.Parallel()
	// Создаем минимальную симуляцию
	cfg := &config.Config{
		World: config.WorldConfig{
			Size: 10,
			Seed: 12345, // Фиксированный seed для воспроизводимости
		},
	}

	terrainGen := generator.NewTerrainGenerator(cfg)
	terrain := terrainGen.Generate()

	// Создаем мир и системы
	worldSizePixels := float32(320) // 10 * 32
	world := core.NewWorld(worldSizePixels, worldSizePixels, 12345)
	systemManager := core.NewSystemManager()

	// Создаем только необходимые системы
	vegetationSystem := simulation.NewVegetationSystem(terrain)
	animalBehaviorSystem := simulation.NewAnimalBehaviorSystem(vegetationSystem)
	movementSystem := simulation.NewMovementSystem(worldSizePixels, worldSizePixels)

	systemManager.AddSystem(vegetationSystem)
	systemManager.AddSystem(&adapters.BehaviorSystemAdapter{System: animalBehaviorSystem})
	systemManager.AddSystem(&adapters.MovementSystemAdapter{System: movementSystem})
	systemManager.AddSystem(adapters.NewFeedingSystemAdapter(vegetationSystem))

	// Создаем волка и зайца рядом друг с другом
	rabbitX, rabbitY := float32(160), float32(160) // Центр мира
	wolfX, wolfY := float32(140), float32(160)     // Слева от зайца на расстоянии 20 единиц

	rabbit := simulation.CreateAnimal(world, core.TypeRabbit, rabbitX, rabbitY)
	wolf := simulation.CreateAnimal(world, core.TypeWolf, wolfX, wolfY)

	// Делаем волка голодным для охоты
	world.SetHunger(wolf, core.Hunger{Value: 30.0}) // Меньше 60% - будет охотиться

	t.Logf("=== Исследование поведения волка при атаке ===")
	t.Logf("Начальные позиции: волк (%.1f, %.1f), заяц (%.1f, %.1f)", wolfX, wolfY, rabbitX, rabbitY)

	deltaTime := float32(1.0 / 60.0) // 60 FPS
	tickCount := 0
	maxTicks := 600 // 10 секунд

	// Логируем каждые 6 тиков (10 раз в секунду)
	logInterval := 6

	for tickCount < maxTicks {
		// Проверяем жив ли заяц
		if !world.IsAlive(rabbit) {
			t.Logf("Заяц умер на тике %d", tickCount)
			break
		}

		// Получаем позиции до обновления
		wolfPos, _ := world.GetPosition(wolf)
		rabbitPos, _ := world.GetPosition(rabbit)
		wolfVel, _ := world.GetVelocity(wolf)

		// Вычисляем расстояние
		dx := wolfPos.X - rabbitPos.X
		dy := wolfPos.Y - rabbitPos.Y
		distance := math.Sqrt(float64(dx*dx + dy*dy))

		// Логируем каждые несколько тиков
		if tickCount%logInterval == 0 {
			t.Logf("Тик %3d: волк (%.1f,%.1f) vel(%.1f,%.1f) | заяц (%.1f,%.1f) | дистанция %.1f",
				tickCount, wolfPos.X, wolfPos.Y, wolfVel.X, wolfVel.Y, rabbitPos.X, rabbitPos.Y, distance)
		}

		// Обновляем только поведение волка вручную
		terrain := terrainGen.Generate()
		vegetationSystem := simulation.NewVegetationSystem(terrain)
		animalBehaviorSystem := simulation.NewAnimalBehaviorSystem(vegetationSystem)
		animalBehaviorSystem.Update(world, deltaTime)

		// Обновляем движение для всех
		world.Update(deltaTime)
		systemManager.Update(world, deltaTime)

		tickCount++
	}

	// Финальные позиции
	if world.IsAlive(wolf) {
		wolfPos, _ := world.GetPosition(wolf)
		t.Logf("Финальная позиция волка: (%.1f, %.1f)", wolfPos.X, wolfPos.Y)
	}

	if world.IsAlive(rabbit) {
		rabbitPos, _ := world.GetPosition(rabbit)
		t.Logf("Финальная позиция зайца: (%.1f, %.1f)", rabbitPos.X, rabbitPos.Y)
	}
}

// TestWolfOvershooting проверяет перепрыгивание волка через зайца
func TestWolfOvershooting(t *testing.T) {
	t.Parallel()
	// Создаем простую симуляцию
	worldSizePixels := float32(320)
	world := core.NewWorld(worldSizePixels, worldSizePixels, 54321)
	systemManager := core.NewSystemManager()

	movementSystem := simulation.NewMovementSystem(worldSizePixels, worldSizePixels)

	// Добавляем только движение, поведение будем вызывать вручную для волка
	systemManager.AddSystem(&adapters.MovementSystemAdapter{System: movementSystem})

	// Зайца ставим неподвижно, волка близко
	rabbit := simulation.CreateAnimal(world, core.TypeRabbit, 160, 160)
	wolf := simulation.CreateAnimal(world, core.TypeWolf, 145, 160) // Расстояние 15 единиц

	// Зайца делаем неподвижным
	world.SetVelocity(rabbit, core.Velocity{X: 0, Y: 0})
	world.SetSpeed(rabbit, core.Speed{Base: 0, Current: 0})

	// Волка делаем голодным
	world.SetHunger(wolf, core.Hunger{Value: 20.0})

	t.Logf("=== Тест перепрыгивания волка ===")

	deltaTime := float32(1.0 / 60.0)

	for i := 0; i < 120; i++ { // 2 секунды
		wolfPos, _ := world.GetPosition(wolf)
		rabbitPos, _ := world.GetPosition(rabbit)

		dx := wolfPos.X - rabbitPos.X
		dy := wolfPos.Y - rabbitPos.Y
		distance := math.Sqrt(float64(dx*dx + dy*dy))

		if i%12 == 0 { // Каждые 0.2 секунды
			t.Logf("Сек %.1f: волк (%.1f,%.1f) | заяц (%.1f,%.1f) | дистанция %.1f",
				float32(i)/60.0, wolfPos.X, wolfPos.Y, rabbitPos.X, rabbitPos.Y, distance)
		}

		// Проверяем не перепрыгнул ли волк
		if wolfPos.X > rabbitPos.X && i > 30 { // Если волк прошел за зайца
			t.Logf("ВНИМАНИЕ: Волк перепрыгнул зайца на тике %d!", i)
			t.Logf("  Волк: (%.1f, %.1f), Заяц: (%.1f, %.1f)", wolfPos.X, wolfPos.Y, rabbitPos.X, rabbitPos.Y)
			break
		}

		if !world.IsAlive(rabbit) {
			t.Logf("Заяц умер на тике %d", i)
			break
		}

		world.Update(deltaTime)
		systemManager.Update(world, deltaTime)
	}
}
