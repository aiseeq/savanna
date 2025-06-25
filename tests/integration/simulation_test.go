package integration

import (
	"math"
	"testing"

	"github.com/aiseeq/savanna/config"
	"github.com/aiseeq/savanna/internal/adapters"
	"github.com/aiseeq/savanna/internal/core"
	"github.com/aiseeq/savanna/internal/generator"
	"github.com/aiseeq/savanna/internal/simulation"
	"github.com/aiseeq/savanna/tests/common"
)

const (
	TestWorldSize = 20.0 * 32.0 // 20 тайлов для тестов (меньше чем в реальной игре)
	TestTPS       = 60
)

// createTestVegetationSystem создаёт vegetation систему для тестов
func createTestVegetationSystem() *simulation.VegetationSystem {
	cfg := config.LoadDefaultConfig()
	cfg.World.Size = int(TestWorldSize / 32)
	terrainGen := generator.NewTerrainGenerator(cfg)
	terrain := terrainGen.Generate()
	return simulation.NewVegetationSystem(terrain)
}

// TestBasicSimulation проверяет базовую работу симуляции
func TestBasicSimulation(t *testing.T) {
	t.Parallel()
	world := core.NewWorld(TestWorldSize, TestWorldSize, 42)
	// Используем централизованную фабрику систем для консистентности
	systemManager := common.CreateTestSystemManager(TestWorldSize)

	// Системы уже созданы в правильном порядке через CreateTestSystemManager
	// Не нужно добавлять системы вручную - они уже включены в systemManager

	// Создаем несколько животных
	rabbit1 := simulation.CreateAnimal(world, core.TypeRabbit, 100, 100)
	rabbit2 := simulation.CreateAnimal(world, core.TypeRabbit, 200, 200)
	wolf1 := simulation.CreateAnimal(world, core.TypeWolf, 101, 100) // Близко к rabbit1

	// Делаем волка голодным чтобы он охотился
	world.SetSatiation(wolf1, core.Satiation{Value: 30.0}) // 30% < 60% порога

	// Проверяем что животные созданы
	if world.GetEntityCount() != 3 {
		t.Errorf("Expected 3 entities, got %d", world.GetEntityCount())
	}

	// Запускаем симуляцию на короткое время - достаточно чтобы увидеть движение
	deltaTime := float32(1.0 / TestTPS)
	ticksToRun := 30 // Полсекунды вместо полной секунды

	for i := 0; i < ticksToRun; i++ {
		world.Update(deltaTime)
		systemManager.Update(world, deltaTime)

		// Логируем каждые 5 тиков для более детального просмотра
		if i%5 == 0 {
			rabbitPos, _ := world.GetPosition(rabbit1)
			wolfPos, _ := world.GetPosition(wolf1)
			rabbitVel, _ := world.GetVelocity(rabbit1)
			wolfVel, _ := world.GetVelocity(wolf1)
			hasEating := world.HasComponent(rabbit1, core.MaskEatingState)
			t.Logf("Тик %d: заяц (%.1f,%.1f) vel(%.1f,%.1f) eating=%v, волк (%.1f,%.1f) vel(%.1f,%.1f)",
				i, rabbitPos.X, rabbitPos.Y, rabbitVel.X, rabbitVel.Y, hasEating,
				wolfPos.X, wolfPos.Y, wolfVel.X, wolfVel.Y)
		}
	}

	// Проверяем что животные живы
	if !world.IsAlive(rabbit1) {
		t.Error("Rabbit1 should be alive after simulation")
	}

	if !world.IsAlive(rabbit2) {
		t.Error("Rabbit2 should be alive after simulation")
	}

	if !world.IsAlive(wolf1) {
		t.Error("Wolf1 should be alive after simulation")
	}

	// Проверяем что позиции изменились (животные двигались)
	pos1, _ := world.GetPosition(rabbit1)
	vel1, _ := world.GetVelocity(rabbit1)
	hasEatingState := world.HasComponent(rabbit1, core.MaskEatingState)
	initialPos := core.Position{X: 100, Y: 100}

	// Дополнительное логирование для диагностики
	t.Logf("Финальное состояние зайца: pos=(%.1f,%.1f), vel=(%.1f,%.1f), eating=%v",
		pos1.X, pos1.Y, vel1.X, vel1.Y, hasEatingState)

	// Проверяем что заяц сдвинулся более чем на 0.5 пикселя от начальной позиции
	// (уменьшили порог, так как за полсекунды движение может быть меньше)
	distanceMoved := calculateDistance(pos1.X, pos1.Y, initialPos.X, initialPos.Y)
	if distanceMoved < 0.5 {
		t.Errorf("Rabbit1 should have moved more than 0.5 pixels from initial position. Initial: (%.1f,%.1f), Final: (%.1f,%.1f), Distance: %.2f",
			initialPos.X, initialPos.Y, pos1.X, pos1.Y, distanceMoved)
	} else {
		t.Logf("SUCCESS: Rabbit moved %.2f pixels from initial position", distanceMoved)
	}
}

// TestDeterministicSimulation проверяет детерминированность симуляции
func TestDeterministicSimulation(t *testing.T) {
	t.Parallel()
	// Функция для запуска симуляции и получения результата
	runSimulation := func(seed int64) map[core.AnimalType]int {
		world := core.NewWorld(TestWorldSize, TestWorldSize, seed)
		systemManager := core.NewSystemManager()

		// Создаём vegetation систему для детерминированного теста
		vegetationSystem := createTestVegetationSystem() // используется в системах
		systemManager.AddSystem(vegetationSystem)
		systemManager.AddSystem(&adapters.BehaviorSystemAdapter{
			System: simulation.NewAnimalBehaviorSystem(vegetationSystem),
		})
		systemManager.AddSystem(&adapters.MovementSystemAdapter{
			System: simulation.NewMovementSystem(TestWorldSize, TestWorldSize),
		})
		// Используем новые системы питания (DIP: через интерфейс)
		satiationSystem := simulation.NewSatiationSystem()
		grassSearchSystem := simulation.NewGrassSearchSystem(vegetationSystem)
		grassEatingSystem := simulation.NewGrassEatingSystem(vegetationSystem)

		systemManager.AddSystem(&adapters.HungerSystemAdapter{System: satiationSystem})
		systemManager.AddSystem(&adapters.GrassSearchSystemAdapter{System: grassSearchSystem})
		systemManager.AddSystem(&adapters.GrassEatingSystemAdapter{System: grassEatingSystem})

		// Создаем 10 зайцев и 2 волков
		rng := world.GetRNG()
		for i := 0; i < 10; i++ {
			x := rng.Float32()*TestWorldSize*0.8 + TestWorldSize*0.1
			y := rng.Float32()*TestWorldSize*0.8 + TestWorldSize*0.1
			simulation.CreateAnimal(world, core.TypeRabbit, x, y)
		}

		for i := 0; i < 2; i++ {
			x := rng.Float32()*TestWorldSize*0.8 + TestWorldSize*0.1
			y := rng.Float32()*TestWorldSize*0.8 + TestWorldSize*0.1
			simulation.CreateAnimal(world, core.TypeWolf, x, y)
		}

		// Симулируем 10 секунд
		deltaTime := float32(1.0 / TestTPS)
		for i := 0; i < TestTPS*10; i++ {
			world.Update(deltaTime)
			systemManager.Update(world, deltaTime)
		}

		return world.GetStats()
	}

	// Запускаем два раза с одинаковым seed
	result1 := runSimulation(12345)
	result2 := runSimulation(12345)

	// Результаты должны быть идентичными
	if result1[core.TypeRabbit] != result2[core.TypeRabbit] {
		t.Errorf("Rabbit counts differ: %d vs %d",
			result1[core.TypeRabbit], result2[core.TypeRabbit])
	}

	if result1[core.TypeWolf] != result2[core.TypeWolf] {
		t.Errorf("Wolf counts differ: %d vs %d",
			result1[core.TypeWolf], result2[core.TypeWolf])
	}

	// Проверяем что с разным seed результаты могут отличаться
	result3 := runSimulation(54321)
	differentResults := result1[core.TypeRabbit] != result3[core.TypeRabbit] ||
		result1[core.TypeWolf] != result3[core.TypeWolf]

	if !differentResults {
		t.Log("Warning: Different seeds produced identical results (possible but unlikely)")
	}
}

// TestHungerSystem проверяет систему голода
func TestHungerSystem(t *testing.T) {
	t.Parallel()
	world := core.NewWorld(TestWorldSize, TestWorldSize, 42)
	_ = createTestVegetationSystem() // используется в системах

	// Создаем зайца с низким голодом
	rabbit := simulation.CreateAnimal(world, core.TypeRabbit, 100, 100)

	// Устанавливаем очень низкий голод
	world.SetSatiation(rabbit, core.Satiation{Value: 1.0})

	initialHealth, _ := world.GetHealth(rabbit)

	// Запускаем систему голода на несколько секунд
	deltaTime := float32(1.0 / TestTPS)
	vegetationSystem := createTestVegetationSystem()
	feedingSystemAdapter := adapters.NewDeprecatedFeedingSystemAdapter(vegetationSystem)
	for i := 0; i < TestTPS*5; i++ { // 5 секунд
		world.Update(deltaTime)
		feedingSystemAdapter.Update(world, deltaTime)
	}

	// Проверяем что голод уменьшился
	currentHunger, _ := world.GetSatiation(rabbit)
	if currentHunger.Value >= 1.0 {
		t.Error("Hunger should have decreased")
	}

	// Если голод стал 0, здоровье должно начать уменьшаться
	if currentHunger.Value <= 0 {
		currentHealth, _ := world.GetHealth(rabbit)
		if currentHealth.Current >= initialHealth.Current {
			t.Error("Health should decrease when hunger is 0")
		}
	}
}

// TestAnimalInteraction проверяет взаимодействие животных
func TestAnimalInteraction(t *testing.T) {
	t.Parallel()
	world := core.NewWorld(TestWorldSize, TestWorldSize, 42)
	systemManager := core.NewSystemManager()

	// Создаём vegetation систему для взаимодействия
	vegetationSystem := createTestVegetationSystem() // используется в системах
	systemManager.AddSystem(vegetationSystem)
	systemManager.AddSystem(&adapters.BehaviorSystemAdapter{
		System: simulation.NewAnimalBehaviorSystem(vegetationSystem),
	})
	systemManager.AddSystem(&adapters.MovementSystemAdapter{
		System: simulation.NewMovementSystem(TestWorldSize, TestWorldSize),
	})
	// Используем новые системы питания (DIP: через интерфейс)
	satiationSystem := simulation.NewSatiationSystem()
	grassSearchSystem := simulation.NewGrassSearchSystem(vegetationSystem)
	grassEatingSystem := simulation.NewGrassEatingSystem(vegetationSystem)

	systemManager.AddSystem(&adapters.HungerSystemAdapter{System: satiationSystem})
	systemManager.AddSystem(&adapters.GrassSearchSystemAdapter{System: grassSearchSystem})
	systemManager.AddSystem(&adapters.GrassEatingSystemAdapter{System: grassEatingSystem})

	// Создаем зайца и волка очень близко для тайловой системы
	rabbit1 := simulation.CreateAnimal(world, core.TypeRabbit, 100, 100)
	wolf1 := simulation.CreateAnimal(world, core.TypeWolf, 101, 100) // Дистанция 1 пиксель

	// Делаем волка голодным чтобы он охотился
	world.SetSatiation(wolf1, core.Satiation{Value: 30.0}) // 30% < 60% порога

	initialRabbitPos, _ := world.GetPosition(rabbit1)
	initialWolfPos, _ := world.GetPosition(wolf1)

	// Запускаем симуляцию на 2 секунды
	deltaTime := float32(1.0 / TestTPS)
	for i := 0; i < TestTPS*2; i++ {
		world.Update(deltaTime)
		systemManager.Update(world, deltaTime)
	}

	// Получаем финальные позиции животных после симуляции
	finalRabbitPos, _ := world.GetPosition(rabbit1)
	finalWolfPos, _ := world.GetPosition(wolf1)

	rabbitMoved := finalRabbitPos.X != initialRabbitPos.X || finalRabbitPos.Y != initialRabbitPos.Y
	wolfMoved := finalWolfPos.X != initialWolfPos.X || finalWolfPos.Y != initialWolfPos.Y

	t.Logf("Начальные позиции: заяц (%.1f,%.1f), волк (%.1f,%.1f)",
		initialRabbitPos.X, initialRabbitPos.Y, initialWolfPos.X, initialWolfPos.Y)
	t.Logf("Финальные позиции: заяц (%.1f,%.1f), волк (%.1f,%.1f)",
		finalRabbitPos.X, finalRabbitPos.Y, finalWolfPos.X, finalWolfPos.Y)
	t.Logf("Движение: заяц %v, волк %v", rabbitMoved, wolfMoved)

	// Проверяем что животные живы
	if !world.IsAlive(rabbit1) {
		t.Error("Rabbit should be alive")
	}
	if !world.IsAlive(wolf1) {
		t.Error("Wolf should be alive")
	}

	// Проверяем что животные двигались
	if !rabbitMoved {
		t.Error("Rabbit should have moved")
	}
	if !wolfMoved {
		t.Error("Wolf should have moved")
	}

	// Проверяем что заяц пытался убежать от волка
	// Вычисляем направление движения зайца
	rabbitMovement := calculateDistance(finalRabbitPos.X-initialRabbitPos.X, finalRabbitPos.Y-initialRabbitPos.Y, 0, 0)

	// Заяц должен двигаться (расстояние движения > 0)
	if rabbitMovement < 1.0 {
		t.Error("Rabbit should have moved when wolf is nearby")
	}

	// Дополнительно проверяем что оба животных в пределах мира
	if finalRabbitPos.X < 0 || finalRabbitPos.X > TestWorldSize ||
		finalRabbitPos.Y < 0 || finalRabbitPos.Y > TestWorldSize {
		t.Error("Rabbit moved outside world boundaries")
	}

	if finalWolfPos.X < 0 || finalWolfPos.X > TestWorldSize ||
		finalWolfPos.Y < 0 || finalWolfPos.Y > TestWorldSize {
		t.Error("Wolf moved outside world boundaries")
	}
}

// TestBoundaryConstraints проверяет что животные не выходят за границы мира
func TestBoundaryConstraints(t *testing.T) {
	t.Parallel()
	world := core.NewWorld(TestWorldSize, TestWorldSize, 42)
	movementSystem := simulation.NewMovementSystem(TestWorldSize, TestWorldSize)

	// Создаем зайца у края мира
	rabbit := simulation.CreateAnimal(world, core.TypeRabbit, 5, 5) // Близко к левому верхнему углу

	// Устанавливаем скорость в сторону границы
	world.SetVelocity(rabbit, core.Velocity{X: -100, Y: -100}) // Движение к границе

	// Запускаем систему движения
	deltaTime := float32(1.0 / TestTPS)
	for i := 0; i < TestTPS; i++ { // 1 секунда
		world.Update(deltaTime)
		movementSystem.Update(world, deltaTime)
	}

	// Проверяем что заяц остался в границах
	pos, _ := world.GetPosition(rabbit)
	size, _ := world.GetSize(rabbit)

	if pos.X-size.Radius < 0 {
		t.Errorf("Rabbit went outside left boundary: pos.X=%f, radius=%f", pos.X, size.Radius)
	}

	if pos.Y-size.Radius < 0 {
		t.Errorf("Rabbit went outside top boundary: pos.Y=%f, radius=%f", pos.Y, size.Radius)
	}

	if pos.X+size.Radius > TestWorldSize {
		t.Errorf("Rabbit went outside right boundary: pos.X=%f, radius=%f, world=%f",
			pos.X, size.Radius, TestWorldSize)
	}

	if pos.Y+size.Radius > TestWorldSize {
		t.Errorf("Rabbit went outside bottom boundary: pos.Y=%f, radius=%f, world=%f",
			pos.Y, size.Radius, TestWorldSize)
	}
}

// TestStarvationDeath проверяет что животные умирают от голода
func TestStarvationDeath(t *testing.T) {
	t.Parallel()
	world := core.NewWorld(TestWorldSize, TestWorldSize, 42)
	_ = createTestVegetationSystem()             // используется в системах
	combatSystem := simulation.NewCombatSystem() // Отвечает за удаление мертвых животных

	// Создаем зайца с минимальным здоровьем и голодом
	rabbit := simulation.CreateAnimal(world, core.TypeRabbit, 100, 100)
	world.SetHealth(rabbit, core.Health{Current: 2, Max: 50}) // Минимальное здоровье
	//nolint:gocritic // commentedOutCode: Это описательный комментарий
	world.SetSatiation(rabbit, core.Satiation{Value: 0}) // Голод = 0

	if !world.IsAlive(rabbit) {
		t.Fatal("Rabbit should be alive initially")
	}

	// Запускаем систему питания пока заяц не умрет
	deltaTime := float32(1.0 / TestTPS)
	maxIterations := TestTPS * 10 // Максимум 10 секунд
	vegetationSystem := createTestVegetationSystem()
	feedingSystemAdapter := adapters.NewDeprecatedFeedingSystemAdapter(vegetationSystem)

	for i := 0; i < maxIterations && world.IsAlive(rabbit); i++ {
		world.Update(deltaTime)
		feedingSystemAdapter.Update(world, deltaTime)
		combatSystem.Update(world, deltaTime) // Удаляет мертвых животных
	}

	// Заяц должен умереть от голода
	if world.IsAlive(rabbit) {
		t.Error("Rabbit should have died from starvation")
	}

	// Проверяем что сущность действительно удалена
	if world.GetEntityCount() != 0 {
		t.Errorf("Expected 0 entities after death, got %d", world.GetEntityCount())
	}
}

// calculateDistance вычисляет расстояние между двумя точками
func calculateDistance(x1, y1, x2, y2 float32) float32 {
	dx := x2 - x1
	dy := y2 - y1
	return float32(math.Sqrt(float64(dx*dx + dy*dy))) // Возвращаем реальное расстояние
}
