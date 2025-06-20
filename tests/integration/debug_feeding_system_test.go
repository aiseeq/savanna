package integration

import (
	"testing"

	"github.com/aiseeq/savanna/config"
	"github.com/aiseeq/savanna/internal/core"
	"github.com/aiseeq/savanna/internal/generator"
	"github.com/aiseeq/savanna/internal/simulation"
)

// TestFeedingSystemDirect проверяет работу новых систем питания
func TestFeedingSystemDirect(t *testing.T) {
	t.Parallel()

	t.Logf("=== ТЕСТ ПРЯМОЙ РАБОТЫ НОВЫХ СИСТЕМ ПИТАНИЯ ===")

	// Настройка тестовой среды
	world, vegetationSystem, rabbit := setupFeedingTest(t)

	// Проверка начального состояния
	logInitialState(t, world, rabbit, vegetationSystem)

	// Тестирование систем и проверка результата
	testFeedingSystemExecution(t, world, vegetationSystem, rabbit)
}

// setupFeedingTest настраивает тестовую среду для тестирования новых систем питания
func setupFeedingTest(
	t *testing.T,
) (*core.World, *simulation.VegetationSystem, core.EntityID) {
	// Создаём мир
	world := core.NewWorld(1600, 1600, 12345) //nolint:gomnd // Тестовые параметры

	// Создаём terrain с травой
	cfg := config.LoadDefaultConfig()
	cfg.World.Size = 50 //nolint:gomnd // Тестовый размер мира
	terrainGen := generator.NewTerrainGenerator(cfg)
	terrain := terrainGen.Generate()

	// Устанавливаем траву в центр
	centerX, centerY := 25, 25 //nolint:gomnd // Центр тестового мира
	terrain.SetTileType(centerX, centerY, generator.TileGrass)
	terrain.SetGrassAmount(centerX, centerY, 100.0) //nolint:gomnd // Максимум травы
	grassInTerrain := terrain.GetGrassAmount(centerX, centerY)
	t.Logf("Трава в terrain тайле (%d, %d): %.1f", centerX, centerY, grassInTerrain)

	// Создаём системы
	vegetationSystem := simulation.NewVegetationSystem(terrain)

	// Создаём зайца точно в центре тайла с травой
	rabbitX, rabbitY := float32(centerX*32+16), float32(centerY*32+16) //nolint:gomnd // Центр тайла
	rabbit := simulation.CreateAnimal(world, core.TypeRabbit, rabbitX, rabbitY)

	// Делаем зайца голодным
	world.SetHunger(rabbit, core.Hunger{Value: 70.0}) //nolint:gomnd // 70% < 90% = голодный

	return world, vegetationSystem, rabbit
}

// logInitialState выводит информацию о начальном состоянии теста
func logInitialState(
	t *testing.T, world *core.World, rabbit core.EntityID, vegetationSystem *simulation.VegetationSystem,
) {
	pos, _ := world.GetPosition(rabbit)
	hunger, _ := world.GetHunger(rabbit)
	tileX := int(pos.X / 32)   //nolint:gomnd // Размер тайла
	tileY := int(pos.Y / 32)   //nolint:gomnd // Размер тайла
	centerX, centerY := 25, 25 //nolint:gomnd // Центр тестового мира
	grassViaVegetation := vegetationSystem.GetGrassAt(pos.X, pos.Y)

	t.Logf("НАЧАЛЬНОЕ СОСТОЯНИЕ:")
	t.Logf("  Позиция зайца: (%.1f, %.1f)", pos.X, pos.Y)
	t.Logf("  Тайл зайца: (%d, %d), ожидаем (%d, %d)", tileX, tileY, centerX, centerY)
	t.Logf("  Голод зайца: %.1f%% (порог %.1f%%)", hunger.Value, simulation.RabbitHungerThreshold)
	t.Logf("  Трава через VegetationSystem: %.1f", grassViaVegetation)
	t.Logf("  Минимум травы для поедания: %.1f", simulation.MinGrassAmountToFind)

	// Проверяем условия
	isHungry := hunger.Value < simulation.RabbitHungerThreshold
	hasEnoughGrass := grassViaVegetation >= simulation.MinGrassAmountToFind
	hasEatingStateBefore := world.HasComponent(rabbit, core.MaskEatingState)

	t.Logf("ПРОВЕРКА УСЛОВИЙ:")
	t.Logf("  Заяц голоден: %v (%.1f < %.1f)", isHungry, hunger.Value, simulation.RabbitHungerThreshold)
	t.Logf("  Достаточно травы: %v (%.1f >= %.1f)", hasEnoughGrass, grassViaVegetation, simulation.MinGrassAmountToFind)
	t.Logf("  EatingState до FeedingSystem: %v", hasEatingStateBefore)
}

// testFeedingSystemExecution выполняет тест новых систем питания и проверяет результат
func testFeedingSystemExecution(
	t *testing.T, world *core.World, vegetation *simulation.VegetationSystem, rabbit core.EntityID,
) {
	// Создаём новые системы питания
	hungerSystem := simulation.NewHungerSystem()
	grassSearchSystem := simulation.NewGrassSearchSystem(vegetation)
	grassEatingSystem := simulation.NewGrassEatingSystem(vegetation)

	deltaTime := float32(1.0 / 60.0) //nolint:gomnd // 60 FPS

	t.Logf("\n--- ВЫЗОВ НОВЫХ СИСТЕМ ПИТАНИЯ ---")

	// 1. Обновляем голод
	hungerSystem.Update(world, deltaTime)

	// 2. Ищем траву и создаём EatingState
	grassSearchSystem.Update(world, deltaTime)

	// 3. Обрабатываем поедание травы
	grassEatingSystem.Update(world, deltaTime)

	// Проверяем результат
	hasEatingStateAfter := world.HasComponent(rabbit, core.MaskEatingState)
	hungerAfter, _ := world.GetHunger(rabbit)

	t.Logf("РЕЗУЛЬТАТ:")
	t.Logf("  EatingState после новых систем: %v", hasEatingStateAfter)
	t.Logf("  Голод после новых систем: %.1f%%", hungerAfter.Value)

	if hasEatingStateAfter {
		eatingState, _ := world.GetEatingState(rabbit)
		t.Logf("  EatingState детали: Target=%d, Progress=%.2f, Nutrition=%.2f",
			eatingState.Target, eatingState.EatingProgress, eatingState.NutritionGained)
		t.Logf("✅ УСПЕХ: Новые системы питания создали EatingState!")
	} else {
		t.Errorf("❌ ОШИБКА: Новые системы питания НЕ создали EatingState!")
		t.Errorf("   Все условия выполнены, но состояние не создано")
		t.Errorf("   Возможные причины:")
		t.Errorf("   1. Заяц не найден в ForEachWith")
		t.Errorf("   2. Ошибка в проверке условий")
		t.Errorf("   3. Ошибка в AddEatingState")
	}
}
