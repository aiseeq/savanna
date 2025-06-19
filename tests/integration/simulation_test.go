package integration

import (
	"testing"

	"github.com/aiseeq/savanna/config"
	"github.com/aiseeq/savanna/internal/adapters"
	"github.com/aiseeq/savanna/internal/core"
	"github.com/aiseeq/savanna/internal/generator"
	"github.com/aiseeq/savanna/internal/simulation"
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
	systemManager := core.NewSystemManager()

	// Создаём минимальную vegetation систему для тестов
	cfg := config.LoadDefaultConfig()
	cfg.World.Size = int(TestWorldSize / 32)
	terrainGen := generator.NewTerrainGenerator(cfg)
	terrain := terrainGen.Generate()
	vegetationSystem := simulation.NewVegetationSystem(terrain)

	// Добавляем системы
	systemManager.AddSystem(vegetationSystem)
	systemManager.AddSystem(&adapters.BehaviorSystemAdapter{
		System: simulation.NewAnimalBehaviorSystem(vegetationSystem),
	})
	systemManager.AddSystem(&adapters.MovementSystemAdapter{
		System: simulation.NewMovementSystem(TestWorldSize, TestWorldSize),
	})
	systemManager.AddSystem(&adapters.FeedingSystemAdapter{
		System: simulation.NewFeedingSystem(vegetationSystem),
	})

	// Создаем несколько животных
	rabbit1 := simulation.CreateAnimal(world, core.TypeRabbit, 100, 100)
	rabbit2 := simulation.CreateAnimal(world, core.TypeRabbit, 200, 200)
	wolf1 := simulation.CreateAnimal(world, core.TypeWolf, 300, 300)

	// Проверяем что животные созданы
	if world.GetEntityCount() != 3 {
		t.Errorf("Expected 3 entities, got %d", world.GetEntityCount())
	}

	// Запускаем симуляцию на 1 секунду
	deltaTime := float32(1.0 / TestTPS)
	for i := 0; i < TestTPS; i++ {
		world.Update(deltaTime)
		systemManager.Update(world, deltaTime)
	}

	// Проверяем что животные живы и двигаются
	if !world.IsAlive(rabbit1) {
		t.Error("Rabbit1 should be alive after 1 second")
	}

	if !world.IsAlive(rabbit2) {
		t.Error("Rabbit2 should be alive after 1 second")
	}

	if !world.IsAlive(wolf1) {
		t.Error("Wolf1 should be alive after 1 second")
	}

	// Проверяем что позиции изменились (животные двигались)
	pos1, _ := world.GetPosition(rabbit1)
	if pos1.X == 100 && pos1.Y == 100 {
		t.Error("Rabbit1 should have moved from initial position")
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
		vegetationSystem := createTestVegetationSystem()
		systemManager.AddSystem(vegetationSystem)
		systemManager.AddSystem(&adapters.BehaviorSystemAdapter{
			System: simulation.NewAnimalBehaviorSystem(vegetationSystem),
		})
		systemManager.AddSystem(&adapters.MovementSystemAdapter{
			System: simulation.NewMovementSystem(TestWorldSize, TestWorldSize),
		})
		systemManager.AddSystem(&adapters.FeedingSystemAdapter{
			System: simulation.NewFeedingSystem(vegetationSystem),
		})

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
	vegetationSystem := createTestVegetationSystem()
	feedingSystem := simulation.NewFeedingSystem(vegetationSystem)

	// Создаем зайца с низким голодом
	rabbit := simulation.CreateAnimal(world, core.TypeRabbit, 100, 100)

	// Устанавливаем очень низкий голод
	world.SetHunger(rabbit, core.Hunger{Value: 1.0})

	initialHealth, _ := world.GetHealth(rabbit)

	// Запускаем систему голода на несколько секунд
	deltaTime := float32(1.0 / TestTPS)
	for i := 0; i < TestTPS*5; i++ { // 5 секунд
		world.Update(deltaTime)
		feedingSystem.Update(world, deltaTime)
	}

	// Проверяем что голод уменьшился
	currentHunger, _ := world.GetHunger(rabbit)
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
	vegetationSystem := createTestVegetationSystem()
	systemManager.AddSystem(vegetationSystem)
	systemManager.AddSystem(&adapters.BehaviorSystemAdapter{
		System: simulation.NewAnimalBehaviorSystem(vegetationSystem),
	})
	systemManager.AddSystem(&adapters.MovementSystemAdapter{
		System: simulation.NewMovementSystem(TestWorldSize, TestWorldSize),
	})
	systemManager.AddSystem(&adapters.FeedingSystemAdapter{
		System: simulation.NewFeedingSystem(vegetationSystem),
	})

	// Создаем зайца и волка на дистанции видимости
	rabbit := simulation.CreateAnimal(world, core.TypeRabbit, 200, 200)
	wolf := simulation.CreateAnimal(world, core.TypeWolf, 280, 200) // На расстоянии 80 единиц (в пределах видимости волка)

	// Делаем волка голодным
	world.SetHunger(wolf, core.Hunger{Value: 30.0}) // Ниже порога охоты (60%)

	initialRabbitPos, _ := world.GetPosition(rabbit)
	initialWolfPos, _ := world.GetPosition(wolf)

	// Запускаем симуляцию
	deltaTime := float32(1.0 / TestTPS)
	for i := 0; i < TestTPS*2; i++ { // 2 секунды
		world.Update(deltaTime)
		systemManager.Update(world, deltaTime)
	}

	// Проверяем что животные двигались
	finalRabbitPos, _ := world.GetPosition(rabbit)
	finalWolfPos, _ := world.GetPosition(wolf)

	rabbitMoved := finalRabbitPos.X != initialRabbitPos.X || finalRabbitPos.Y != initialRabbitPos.Y
	wolfMoved := finalWolfPos.X != initialWolfPos.X || finalWolfPos.Y != initialWolfPos.Y

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
	vegetationSystem := createTestVegetationSystem()
	feedingSystem := simulation.NewFeedingSystem(vegetationSystem)
	combatSystem := simulation.NewCombatSystem() // Отвечает за удаление мертвых животных

	// Создаем зайца с минимальным здоровьем и голодом
	rabbit := simulation.CreateAnimal(world, core.TypeRabbit, 100, 100)
	world.SetHealth(rabbit, core.Health{Current: 2, Max: 50}) // Минимальное здоровье
	//nolint:gocritic // commentedOutCode: Это описательный комментарий
	world.SetHunger(rabbit, core.Hunger{Value: 0}) // Голод = 0

	if !world.IsAlive(rabbit) {
		t.Fatal("Rabbit should be alive initially")
	}

	// Запускаем систему питания пока заяц не умрет
	deltaTime := float32(1.0 / TestTPS)
	maxIterations := TestTPS * 10 // Максимум 10 секунд

	for i := 0; i < maxIterations && world.IsAlive(rabbit); i++ {
		world.Update(deltaTime)
		feedingSystem.Update(world, deltaTime)
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
	return float32(dx*dx + dy*dy) // Возвращаем квадрат расстояния для скорости
}
