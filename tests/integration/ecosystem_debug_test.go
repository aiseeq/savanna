package integration

import (
	"testing"

	"github.com/aiseeq/savanna/config"
	"github.com/aiseeq/savanna/internal/core"
	"github.com/aiseeq/savanna/internal/generator"
	"github.com/aiseeq/savanna/internal/simulation"
)

// TestEcosystemSurvival диагностирует проблемы с выживанием животных
func TestEcosystemSurvival(t *testing.T) {
	t.Parallel()
	// Создаём минимальную симуляцию
	cfg := config.LoadDefaultConfig()
	cfg.World.Size = 20
	cfg.Population.Rabbits = 5
	cfg.Population.Wolves = 1

	// Генерируем мир
	terrainGen := generator.NewTerrainGenerator(cfg)
	terrain := terrainGen.Generate()

	// Создаём системы
	worldSizePixels := float32(cfg.World.Size * 32)
	world := core.NewWorld(worldSizePixels, worldSizePixels, cfg.World.Seed)

	vegetationSystem := simulation.NewVegetationSystem(terrain)
	animalBehaviorSystem := simulation.NewAnimalBehaviorSystem(vegetationSystem)
	movementSystem := simulation.NewMovementSystem(worldSizePixels, worldSizePixels)
	feedingSystem := simulation.NewFeedingSystem(vegetationSystem)

	// Размещаем животных
	popGen := generator.NewPopulationGenerator(cfg, terrain)
	placements := popGen.Generate()

	var rabbits []core.EntityID
	var wolves []core.EntityID

	for _, placement := range placements {
		switch placement.Type {
		case core.TypeRabbit:
			rabbit := simulation.CreateRabbit(world, placement.X, placement.Y)
			rabbits = append(rabbits, rabbit)
		case core.TypeWolf:
			wolf := simulation.CreateWolf(world, placement.X, placement.Y)
			wolves = append(wolves, wolf)
		}
	}

	t.Logf("Начальное состояние: %d зайцев, %d волков", len(rabbits), len(wolves))

	// Диагностируем первые несколько кадров
	deltaTime := float32(1.0 / 60.0)

	for frame := 0; frame < 60; frame++ { // 1 секунда
		// Сохраняем состояние до обновления
		aliveRabbits := 0
		aliveWolves := 0

		for _, rabbit := range rabbits {
			if world.IsAlive(rabbit) {
				aliveRabbits++
			}
		}

		for _, wolf := range wolves {
			if world.IsAlive(wolf) {
				aliveWolves++
			}
		}

		// Обновляем по одной системе за раз для диагностики
		world.Update(deltaTime)

		// 1. Система растительности
		vegetationSystem.Update(world, deltaTime)

		// Проверяем выживших после vegetation
		aliveAfterVegetation := 0
		for _, rabbit := range rabbits {
			if world.IsAlive(rabbit) {
				aliveAfterVegetation++
			}
		}

		// 2. Поведение животных
		animalBehaviorSystem.Update(world, deltaTime)

		// Проверяем выживших после behavior
		aliveAfterBehavior := 0
		for _, rabbit := range rabbits {
			if world.IsAlive(rabbit) {
				aliveAfterBehavior++
			}
		}

		// 3. Движение
		movementSystem.Update(world, deltaTime)

		// Проверяем выживших после movement
		aliveAfterMovement := 0
		for _, rabbit := range rabbits {
			if world.IsAlive(rabbit) {
				aliveAfterMovement++
			}
		}

		// 4. Питание (самая подозрительная система)
		feedingSystem.Update(world, deltaTime)

		// Проверяем выживших после feeding
		aliveAfterFeeding := 0
		for _, rabbit := range rabbits {
			if world.IsAlive(rabbit) {
				aliveAfterFeeding++
			}
		}

		// Если кто-то умер, выводим диагностику
		if aliveRabbits != aliveAfterFeeding {
			t.Logf("КАДР %d: Зайцы %d->%d->%d->%d->%d (начало->vegetation->behavior->movement->feeding)",
				frame, aliveRabbits, aliveAfterVegetation, aliveAfterBehavior, aliveAfterMovement, aliveAfterFeeding)

			if aliveAfterVegetation != aliveRabbits {
				t.Logf("  СМЕРТЬ В VEGETATION SYSTEM!")
			}
			if aliveAfterBehavior != aliveAfterVegetation {
				t.Logf("  СМЕРТЬ В BEHAVIOR SYSTEM!")
			}
			if aliveAfterMovement != aliveAfterBehavior {
				t.Logf("  СМЕРТЬ В MOVEMENT SYSTEM!")
			}
			if aliveAfterFeeding != aliveAfterMovement {
				t.Logf("  СМЕРТЬ В FEEDING SYSTEM!")
			}

			// Проверяем позицию волка
			for _, wolf := range wolves {
				if world.IsAlive(wolf) {
					wolfPos, _ := world.GetPosition(wolf)
					wolfHunger, _ := world.GetHunger(wolf)
					t.Logf("  Волк: голод=%.1f, позиция=(%.1f,%.1f)", wolfHunger.Value, wolfPos.X, wolfPos.Y)
				}
			}

			// Анализируем причины смерти
			t.Logf("  Анализ оставшихся зайцев:")
			for i, rabbit := range rabbits {
				if world.IsAlive(rabbit) {
					health, _ := world.GetHealth(rabbit)
					hunger, _ := world.GetHunger(rabbit)
					pos, _ := world.GetPosition(rabbit)
					t.Logf("    Заяц %d: здоровье=%d, голод=%.1f, позиция=(%.1f,%.1f)",
						i, health.Current, hunger.Value, pos.X, pos.Y)
				} else {
					t.Logf("    Заяц %d: МЁРТВ", i)
				}
			}
		}

		// Если все зайцы мертвы, прекращаем
		if aliveAfterFeeding == 0 {
			t.Fatalf("Все зайцы умерли на кадре %d!", frame)
			break
		}
	}

	// Финальная проверка
	finalStats := world.GetStats()
	t.Logf("Финальное состояние: %d зайцев, %d волков",
		finalStats[core.TypeRabbit], finalStats[core.TypeWolf])

	if finalStats[core.TypeRabbit] == 0 {
		t.Error("Все зайцы умерли в течение 1 секунды - это неправильно!")
	}
}

// TestAnimalHealthAndHunger проверяет начальные параметры животных
func TestAnimalHealthAndHunger(t *testing.T) {
	t.Parallel()
	world := core.NewWorld(1000, 1000, 42)

	// Создаём зайца и проверяем его параметры
	rabbit := simulation.CreateRabbit(world, 500, 500)

	health, hasHealth := world.GetHealth(rabbit)
	hunger, hasHunger := world.GetHunger(rabbit)

	if !hasHealth {
		t.Error("Заяц не имеет компонента здоровья")
	}

	if !hasHunger {
		t.Error("Заяц не имеет компонента голода")
	}

	t.Logf("Начальные параметры зайца: здоровье=%d/%d, голод=%.1f",
		health.Current, health.Max, hunger.Value)

	if health.Current <= 0 {
		t.Error("Заяц создается с нулевым здоровьем!")
	}

	if hunger.Value <= 0 {
		t.Error("Заяц создается с нулевым голодом!")
	}

	// Проверяем что заяц жив
	if !world.IsAlive(rabbit) {
		t.Error("Заяц мертв сразу после создания!")
	}
}

// TestFeedingSystemIsolated тестирует систему питания изолированно
func TestFeedingSystemIsolated(t *testing.T) {
	t.Parallel()
	cfg := config.LoadDefaultConfig()
	cfg.World.Size = 20

	terrainGen := generator.NewTerrainGenerator(cfg)
	terrain := terrainGen.Generate()

	world := core.NewWorld(640, 640, 42)
	vegetationSystem := simulation.NewVegetationSystem(terrain)
	feedingSystem := simulation.NewFeedingSystem(vegetationSystem)

	// Создаём зайца с хорошими параметрами
	rabbit := simulation.CreateRabbit(world, 320, 320)

	health, _ := world.GetHealth(rabbit)
	hunger, _ := world.GetHunger(rabbit)

	t.Logf("До feeding: здоровье=%d, голод=%.1f", health.Current, hunger.Value)

	// Обновляем систему питания один раз
	deltaTime := float32(1.0 / 60.0)
	feedingSystem.Update(world, deltaTime)

	if !world.IsAlive(rabbit) {
		t.Error("Заяц умер после одного обновления системы питания!")
		return
	}

	health, _ = world.GetHealth(rabbit)
	hunger, _ = world.GetHunger(rabbit)

	t.Logf("После feeding: здоровье=%d, голод=%.1f", health.Current, hunger.Value)

	// Тестируем много обновлений
	for i := 0; i < 300; i++ { // 5 секунд
		feedingSystem.Update(world, deltaTime)

		if !world.IsAlive(rabbit) {
			health, _ = world.GetHealth(rabbit)
			hunger, _ = world.GetHunger(rabbit)
			t.Logf("Заяц умер на итерации %d: здоровье=%d, голод=%.1f", i, health.Current, hunger.Value)

			if hunger.Value <= 0 && health.Current <= 0 {
				t.Logf("Причина смерти: голодание")
			} else {
				t.Errorf("Заяц умер по неизвестной причине!")
			}
			break
		}
	}
}
