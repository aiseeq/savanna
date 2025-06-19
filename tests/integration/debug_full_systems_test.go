package integration

import (
	"testing"

	"github.com/aiseeq/savanna/config"
	"github.com/aiseeq/savanna/internal/adapters"
	"github.com/aiseeq/savanna/internal/core"
	"github.com/aiseeq/savanna/internal/generator"
	"github.com/aiseeq/savanna/internal/simulation"
)

// TestFullSystemsOrder проверяет работу всех систем в том же порядке что в большом тесте
//
//nolint:revive // function-length: Полный интеграционный тест всех систем
func TestFullSystemsOrder(t *testing.T) {
	t.Parallel()

	t.Logf("=== ТЕСТ ПОЛНОГО НАБОРА СИСТЕМ ===")

	// Создаём мир точно как в большом тесте
	world := core.NewWorld(1600, 1600, 12345) // 50x50 тайлов как в игре

	// Создаём terrain точно как в GUI
	cfg := config.LoadDefaultConfig()
	cfg.World.Size = 50
	terrainGen := generator.NewTerrainGenerator(cfg)
	terrain := terrainGen.Generate()

	// Принудительно устанавливаем траву в центр
	centerX, centerY := 25, 25
	terrain.SetGrassAmount(centerX, centerY, 100.0)

	vegetationSystem := simulation.NewVegetationSystem(terrain)

	// Создаём ВСЕ системы точно как в большом тесте
	systemManager := core.NewSystemManager()
	animalBehaviorSystem := simulation.NewAnimalBehaviorSystem(vegetationSystem)
	feedingSystem := simulation.NewFeedingSystem(vegetationSystem)
	grassEatingSystem := simulation.NewGrassEatingSystem(vegetationSystem)
	combatSystem := simulation.NewCombatSystem()
	movementSystem := simulation.NewMovementSystem(1600, 1600)

	// ДЕБАГ: проверяем что системы созданы правильно
	t.Logf("СОЗДАННЫЕ СИСТЕМЫ:")
	t.Logf("  feedingSystem: %v", feedingSystem != nil)
	t.Logf("  animalBehaviorSystem: %v", animalBehaviorSystem != nil)
	t.Logf("  grassEatingSystem: %v", grassEatingSystem != nil)

	// Добавляем системы в ТОМ ЖЕ порядке что в большом тесте
	systemManager.AddSystem(vegetationSystem)
	systemManager.AddSystem(&adapters.FeedingSystemAdapter{System: feedingSystem}) // 1. Создаёт EatingState
	// 2. Дискретное поедание травы по кадрам анимации
	systemManager.AddSystem(grassEatingSystem)
	// 3. Проверяет EatingState и не мешает еде
	systemManager.AddSystem(&adapters.BehaviorSystemAdapter{System: animalBehaviorSystem})
	systemManager.AddSystem(&adapters.MovementSystemAdapter{System: movementSystem}) // 4. Сбрасывает скорость едящих
	systemManager.AddSystem(combatSystem)                                            // 5. Система боя

	// Создаём зайца в центре где есть трава
	rabbitX, rabbitY := float32(centerX*32+16), float32(centerY*32+16) // Центр тайла
	rabbit := simulation.CreateAnimal(world, core.TypeRabbit, rabbitX, rabbitY)

	// Делаем зайца голодным чтобы он точно ел
	world.SetHunger(rabbit, core.Hunger{Value: 70.0}) // 70% - точно будет есть (порог 90%)
	world.SetVelocity(rabbit, core.Velocity{X: 0, Y: 0})

	deltaTime := float32(1.0 / 60.0)

	t.Logf("НАЧАЛЬНОЕ СОСТОЯНИЕ:")
	pos, _ := world.GetPosition(rabbit)
	hunger, _ := world.GetHunger(rabbit)
	grassAmount := vegetationSystem.GetGrassAt(pos.X, pos.Y)
	behavior, _ := world.GetBehavior(rabbit)

	t.Logf("  Позиция зайца: (%.1f, %.1f)", pos.X, pos.Y)
	t.Logf("  Голод зайца: %.1f%% (порог: %.1f%%)", hunger.Value, behavior.HungerThreshold)
	t.Logf("  Трава в позиции: %.1f единиц", grassAmount)

	// Проверяем что оба метода поиска травы работают
	minGrassToFind := float32(simulation.MinGrassToFind)
	searchRadius := float32(16.0) // Радиус зайца (было simulation.RabbitBaseRadius)
	grassX, grassY, foundGrass := vegetationSystem.FindNearestGrass(pos.X, pos.Y, searchRadius, minGrassToFind)

	t.Logf("  GetGrassAt: %.1f", grassAmount)
	t.Logf("  FindNearestGrass: найдено=%v", foundGrass)
	if foundGrass {
		t.Logf("    Координаты: (%.1f, %.1f)", grassX, grassY)
	}

	// Запускаем ТОЛЬКО ОДИН тик
	t.Logf("\n--- ОДИН ТИК ВСЕХ СИСТЕМ ---")

	// === ЭТАП 1: Состояние ДО обновления ===
	hasEatingStateBefore := world.HasComponent(rabbit, core.MaskEatingState)
	t.Logf("  EatingState ДО: %v", hasEatingStateBefore)

	// === ЭТАП 2: Обновление мира и систем ===
	world.Update(deltaTime)
	systemManager.Update(world, deltaTime)

	// === ЭТАП 3: Состояние ПОСЛЕ систем ===
	hasEatingStateAfter := world.HasComponent(rabbit, core.MaskEatingState)
	hungerAfter, _ := world.GetHunger(rabbit)
	velAfter, _ := world.GetVelocity(rabbit)
	speedAfter := velAfter.X*velAfter.X + velAfter.Y*velAfter.Y

	t.Logf("РЕЗУЛЬТАТ:")
	t.Logf("  EatingState ПОСЛЕ: %v", hasEatingStateAfter)
	t.Logf("  Голод ПОСЛЕ: %.1f%%", hungerAfter.Value)
	t.Logf("  Скорость ПОСЛЕ: %.2f", speedAfter)

	if hasEatingStateAfter {
		eatingState, _ := world.GetEatingState(rabbit)
		t.Logf("  EatingState детали: Target=%d, Progress=%.2f, Nutrition=%.2f",
			eatingState.Target, eatingState.EatingProgress, eatingState.NutritionGained)
		t.Logf("✅ УСПЕХ: Заяц начал есть!")
	} else {
		t.Errorf("❌ ОШИБКА: Заяц НЕ начал есть за 1 тик")
		t.Errorf("   Проверьте, выполняются ли системы правильно")
	}
}
