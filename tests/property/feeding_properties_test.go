package property

import (
	"math/rand"
	"testing"
	"testing/quick"

	"github.com/aiseeq/savanna/config"
	"github.com/aiseeq/savanna/internal/core"
	"github.com/aiseeq/savanna/internal/generator"
	"github.com/aiseeq/savanna/internal/simulation"
)

// Property-Based Testing для системы питания
// Проверяет инварианты системы на случайных данных

// TestRabbitFeedingInvariant проверяет основной инвариант:
// "Голодный заяц рядом с достаточным количеством травы всегда начинает есть"
func TestRabbitFeedingInvariant(t *testing.T) {
	t.Parallel()

	property := func(hunger, grassAmount float32, posX, posY int) bool {
		// Нормализуем входные данные
		if hunger < 0 || hunger > 100 {
			return true // Skip invalid hunger values
		}
		if grassAmount < 0 || grassAmount > 200 {
			return true // Skip invalid grass amounts
		}
		if posX < 0 || posX > 10 || posY < 0 || posY > 10 {
			return true // Skip invalid positions
		}

		// Создаём тестовую среду
		world := core.NewWorld(640, 640, 12345)

		// Создаём terrain с травой
		cfg := config.LoadDefaultConfig()
		cfg.World.Size = 12
		terrainGen := generator.NewTerrainGenerator(cfg)
		terrain := terrainGen.Generate()

		// Устанавливаем контролируемые условия
		terrain.SetTileType(posX, posY, generator.TileGrass)
		terrain.SetGrassAmount(posX, posY, grassAmount)

		vegetationSystem := simulation.NewVegetationSystem(terrain)
		grassSearchSystem := simulation.NewGrassSearchSystem(vegetationSystem)

		// Создаём зайца с заданным голодом
		rabbit := simulation.CreateAnimal(world, core.TypeRabbit,
			float32(posX*32+16), float32(posY*32+16)) // Центр тайла
		world.SetHunger(rabbit, core.Hunger{Value: hunger})

		// Запускаем систему поиска
		grassSearchSystem.Update(world, 1.0/60.0)

		// Проверяем инвариант
		isHungry := hunger < simulation.RabbitHungerThreshold
		hasEnoughGrass := grassAmount >= simulation.MinGrassAmountToFind
		shouldEat := isHungry && hasEnoughGrass

		isEating := world.HasComponent(rabbit, core.MaskEatingState)

		// ИНВАРИАНТ: shouldEat == isEating
		if shouldEat != isEating {
			t.Logf("FAILED: hunger=%.1f, grass=%.1f, pos=(%d,%d), shouldEat=%v, isEating=%v",
				hunger, grassAmount, posX, posY, shouldEat, isEating)
			return false
		}

		return true
	}

	// Запускаем property-based testing
	config := &quick.Config{
		MaxCount: 1000,                         // 1000 случайных тестов
		Rand:     rand.New(rand.NewSource(42)), // Детерминированная случайность
	}

	err := quick.Check(property, config)
	if err != nil {
		t.Fatalf("Property violation found: %v", err)
	}
}

// TestEnergyConservationProperty проверяет закон сохранения энергии:
// "Общая энергия в экосистеме не должна увеличиваться (может только уменьшаться из-за метаболизма)"
func TestEnergyConservationProperty(t *testing.T) {
	t.Parallel()

	property := func(initialRabbits, initialWolves int, simulationTicks int) bool {
		// Нормализуем параметры
		if initialRabbits < 0 || initialRabbits > 5 {
			return true
		}
		if initialWolves < 0 || initialWolves > 3 {
			return true
		}
		if simulationTicks < 1 || simulationTicks > 100 {
			return true
		}

		// Создаём экосистему
		world := core.NewWorld(640, 640, 12345)

		cfg := config.LoadDefaultConfig()
		cfg.World.Size = 10
		terrainGen := generator.NewTerrainGenerator(cfg)
		terrain := terrainGen.Generate()

		// Заполняем траву (начальная энергия)
		initialGrassEnergy := float32(0)
		for x := 0; x < 10; x++ {
			for y := 0; y < 10; y++ {
				terrain.SetTileType(x, y, generator.TileGrass)
				terrain.SetGrassAmount(x, y, 100.0)
				initialGrassEnergy += 100.0
			}
		}

		vegetationSystem := simulation.NewVegetationSystem(terrain)

		// Создаём животных
		initialAnimalEnergy := float32(0)
		for i := 0; i < initialRabbits; i++ {
			rabbit := simulation.CreateAnimal(world, core.TypeRabbit,
				float32(i*50+100), float32(i*50+100))
			world.SetHunger(rabbit, core.Hunger{Value: 100.0}) // Сытые
			initialAnimalEnergy += 100.0
		}

		for i := 0; i < initialWolves; i++ {
			wolf := simulation.CreateAnimal(world, core.TypeWolf,
				float32(i*80+200), float32(i*80+200))
			world.SetHunger(wolf, core.Hunger{Value: 100.0}) // Сытые
			initialAnimalEnergy += 100.0
		}

		initialTotalEnergy := initialGrassEnergy + initialAnimalEnergy

		// Запускаем симуляцию
		grassSearchSystem := simulation.NewGrassSearchSystem(vegetationSystem)
		hungerSystem := simulation.NewHungerSystem()

		for tick := 0; tick < simulationTicks; tick++ {
			hungerSystem.Update(world, 1.0/60.0)
			grassSearchSystem.Update(world, 1.0/60.0)
		}

		// Считаем финальную энергию
		finalGrassEnergy := float32(0)
		for x := 0; x < 10; x++ {
			for y := 0; y < 10; y++ {
				finalGrassEnergy += terrain.GetGrassAmount(x, y)
			}
		}

		finalAnimalEnergy := float32(0)
		world.ForEachWith(core.MaskHunger, func(entity core.EntityID) {
			if hunger, hasHunger := world.GetHunger(entity); hasHunger {
				finalAnimalEnergy += hunger.Value
			}
		})

		finalTotalEnergy := finalGrassEnergy + finalAnimalEnergy

		// ИНВАРИАНТ: finalTotalEnergy <= initialTotalEnergy (энергия не увеличивается)
		if finalTotalEnergy > initialTotalEnergy+0.1 { // Небольшая погрешность
			t.Logf("ENERGY VIOLATION: initial=%.1f, final=%.1f, increase=%.1f",
				initialTotalEnergy, finalTotalEnergy, finalTotalEnergy-initialTotalEnergy)
			return false
		}

		return true
	}

	config := &quick.Config{
		MaxCount: 500, // Более быстрые симуляции
		Rand:     rand.New(rand.NewSource(42)),
	}

	err := quick.Check(property, config)
	if err != nil {
		t.Fatalf("Energy conservation violation: %v", err)
	}
}

// TestGrassSearchDeterminismProperty проверяет детерминированность:
// "При одинаковых условиях система всегда даёт одинаковый результат"
func TestGrassSearchDeterminismProperty(t *testing.T) {
	t.Parallel()

	property := func(seed int64, hunger float32) bool {
		if hunger < 0 || hunger > 100 {
			return true
		}

		// Функция для запуска одинакового сценария
		runScenario := func() bool {
			world := core.NewWorld(640, 640, seed)

			cfg := config.LoadDefaultConfig()
			cfg.World.Size = 5
			terrainGen := generator.NewTerrainGenerator(cfg)
			terrain := terrainGen.Generate()

			// Фиксированная трава
			terrain.SetTileType(2, 2, generator.TileGrass)
			terrain.SetGrassAmount(2, 2, 50.0)

			vegetationSystem := simulation.NewVegetationSystem(terrain)
			grassSearchSystem := simulation.NewGrassSearchSystem(vegetationSystem)

			rabbit := simulation.CreateAnimal(world, core.TypeRabbit, 80, 80)
			world.SetHunger(rabbit, core.Hunger{Value: hunger})

			grassSearchSystem.Update(world, 1.0/60.0)

			return world.HasComponent(rabbit, core.MaskEatingState)
		}

		// Запускаем сценарий дважды
		result1 := runScenario()
		result2 := runScenario()

		// ИНВАРИАНТ: результаты должны быть одинаковыми
		return result1 == result2
	}

	config := &quick.Config{
		MaxCount: 200,
		Rand:     rand.New(rand.NewSource(42)),
	}

	err := quick.Check(property, config)
	if err != nil {
		t.Fatalf("Determinism violation: %v", err)
	}
}
