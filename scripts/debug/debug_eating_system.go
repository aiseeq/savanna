package main

import (
	"fmt"
	"time"

	"github.com/aiseeq/savanna/config"
	"github.com/aiseeq/savanna/internal/adapters"
	"github.com/aiseeq/savanna/internal/animation"
	"github.com/aiseeq/savanna/internal/core"
	"github.com/aiseeq/savanna/internal/generator"
	"github.com/aiseeq/savanna/internal/simulation"
	"github.com/hajimehoshi/ebiten/v2"
)

// Константы диагностической системы (устраняет магические числа)
const (
	// Размеры мира и настройки симуляции
	WorldSeed      = 6  // Фиксированный seed для воспроизводимых результатов
	WorldSize      = 50 // Размер мира в тайлах
	TileSizePixels = 32 // Размер тайла в пикселях
	RabbitCount    = 30 // Количество зайцев в тесте
	WolfCount      = 3  // Количество волков в тесте

	// Параметры симуляции
	TicksPerSecond = 60                             // Частота обновления (TPS)
	SecondsToTest  = 12                             // Время теста до смерти волков
	MaxTicks       = SecondsToTest * TicksPerSecond // 720 тиков

	// Параметры анимации
	EmptyImageWidth  = 128 // Ширина пустого изображения для тестирования
	EmptyImageHeight = 64  // Высота пустого изображения для тестирования

	// Пороги скорости для определения типа анимации
	IdleSpeedThreshold  = 0.1   // Порог неподвижности (скорость < 0.1)
	WolfWalkThreshold   = 400.0 // Порог между ходьбой и бегом для волка
	RabbitWalkThreshold = 300.0 // Порог между ходьбой и бегом для зайца

	// Параметры поиска и атаки
	WolfSearchRadius   = 15.0 // Радиус поиска добычи волком
	WolfAttackDistance = 13.0 // Дистанция атаки волка
	WolfHungerTrigger  = 60.0 // Порог голода для начала охоты

	// Интервалы диагностики и производительности
	DiagnosticInterval = 60  // Интервал вывода диагностики (каждую секунду)
	SleepMicroseconds  = 100 // Пауза между тиками в микросекундах
)

func main() {
	fmt.Println("=== ДИАГНОСТИКА СИСТЕМЫ ПОЕДАНИЯ ТРУПОВ ===")

	// Инициализируем конфигурацию и мир
	cfg, world, systemManager := initializeDebugWorld()

	// Создаём системы симуляции
	terrain, wolfAnimationSystem, rabbitAnimationSystem := setupDebugSystems(cfg, systemManager)

	// Размещаем животных
	_, _ = populateDebugWorld(cfg, world, terrain)

	// Запускаем диагностический цикл
	runDiagnosticLoop(world, systemManager, wolfAnimationSystem, rabbitAnimationSystem)
}

// initializeDebugWorld инициализирует конфигурацию и создаёт мир
func initializeDebugWorld() (*config.Config, *core.World, *core.SystemManager) {
	// Создаём конфигурацию
	cfg := config.LoadDefaultConfig()
	cfg.World.Seed = WorldSeed
	cfg.World.Size = WorldSize
	cfg.Population.Rabbits = RabbitCount
	cfg.Population.Wolves = WolfCount

	// Создаём мир
	worldSizePixels := float32(cfg.World.Size * TileSizePixels)
	world := core.NewWorld(worldSizePixels, worldSizePixels, cfg.World.Seed)
	systemManager := core.NewSystemManager()

	return cfg, world, systemManager
}

// setupDebugSystems создаёт и настраивает все системы симуляции
func setupDebugSystems(
	cfg *config.Config,
	systemManager *core.SystemManager,
) (terrain *generator.Terrain, wolfAnimSystem, rabbitAnimSystem *animation.AnimationSystem) {
	// Генерируем ландшафт
	terrainGen := generator.NewTerrainGenerator(cfg)
	terrain = terrainGen.Generate()

	// Создаём анимационные системы
	wolfAnimSystem = animation.NewAnimationSystem()
	rabbitAnimSystem = animation.NewAnimationSystem()

	// Загружаем анимации
	loader := animation.NewAnimationLoader()
	emptyImg := ebiten.NewImage(128, 64)
	loader.LoadAnimations(wolfAnimSystem, rabbitAnimSystem, emptyImg, emptyImg)

	// Создаём системы с зависимостями
	vegetationSystem := simulation.NewVegetationSystem(terrain)
	animalBehaviorSystem := simulation.NewAnimalBehaviorSystem(vegetationSystem)

	// НОВЫЕ СИСТЕМЫ (следуют принципу SRP):
	satiationSystem := simulation.NewSatiationSystem()                     // 1. Только управление голодом
	grassSearchSystem := simulation.NewGrassSearchSystem(vegetationSystem) // 2. Только поиск травы и создание EatingState
	satiationSpeedModifier := simulation.NewSatiationSpeedModifierSystem() // 3. Только влияние голода на скорость
	starvationDamage := simulation.NewStarvationDamageSystem()             // 4. Только урон от голода

	grassEatingSystem := simulation.NewGrassEatingSystem(vegetationSystem)
	combatSystem := simulation.NewCombatSystem()

	// Добавляем системы в правильном порядке
	worldSizePixels := float32(cfg.World.Size * TileSizePixels)
	systemManager.AddSystem(vegetationSystem)              // 1. Рост травы
	systemManager.AddSystem(&adapters.HungerSystemAdapter{ // 2. Управление голодом
		System: satiationSystem,
	})
	systemManager.AddSystem(&adapters.GrassSearchSystemAdapter{ // 3. Создание EatingState
		System: grassSearchSystem,
	})
	systemManager.AddSystem(grassEatingSystem) // 4. Дискретное поедание травы
	// 5. Поведение (проверяет EatingState)
	systemManager.AddSystem(&adapters.BehaviorSystemAdapter{System: animalBehaviorSystem})
	systemManager.AddSystem(&adapters.HungerSpeedModifierSystemAdapter{ // 6. Влияние голода на скорость
		System: satiationSpeedModifier,
	})
	movementSystem := simulation.NewMovementSystem(worldSizePixels, worldSizePixels)
	// 7. Движение (сбрасывает скорость едящих)
	systemManager.AddSystem(&adapters.MovementSystemAdapter{System: movementSystem})
	systemManager.AddSystem(combatSystem)                            // 8. Система боя
	systemManager.AddSystem(&adapters.StarvationDamageSystemAdapter{ // 9. Урон от голода
		System: starvationDamage,
	})

	return
}

// populateDebugWorld размещает животных в мире
func populateDebugWorld(
	cfg *config.Config,
	world *core.World,
	terrain *generator.Terrain,
) (wolves, rabbits []core.EntityID) {
	// Размещаем животных
	popGen := generator.NewPopulationGenerator(cfg, terrain)
	placements := popGen.Generate()

	for _, placement := range placements {
		switch placement.Type {
		case core.TypeRabbit:
			rabbit := simulation.CreateAnimal(world, core.TypeRabbit, placement.X, placement.Y)
			rabbits = append(rabbits, rabbit)
		case core.TypeWolf:
			wolf := simulation.CreateAnimal(world, core.TypeWolf, placement.X, placement.Y)
			wolves = append(wolves, wolf)
		}
	}

	fmt.Printf("Размещено: %d зайцев, %d волков\n", len(rabbits), len(wolves))
	return
}

// runDiagnosticLoop запускает основной диагностический цикл
func runDiagnosticLoop(
	world *core.World,
	systemManager *core.SystemManager,
	wolfAnimationSystem, rabbitAnimationSystem *animation.AnimationSystem,
) {
	// Создаём менеджер анимаций (устраняет дублирование логики)
	animationManager := animation.NewAnimationManager(wolfAnimationSystem, rabbitAnimationSystem)

	// Функция обновления анимаций через менеджер (устраняет дублирование)
	updateAnimalAnimations := func() {
		animationManager.UpdateAllAnimations(world, 1.0/TicksPerSecond)
	}

	// Диагностический цикл
	deltaTime := float32(1.0 / TicksPerSecond)
	maxTicks := MaxTicks

	for tick := 0; tick < maxTicks; tick++ {
		// Обновляем мир
		world.Update(deltaTime)
		systemManager.Update(world, deltaTime)

		// КРИТИЧЕСКИ ВАЖНО: обновляем анимации для корректной работы боевой системы
		updateAnimalAnimations()

		// Каждую секунду выводим диагностику
		if tick%DiagnosticInterval == 0 {
			printDiagnostics(world, tick)
		}

		time.Sleep(time.Microsecond * SleepMicroseconds)
	}

	fmt.Println("\n=== ДИАГНОСТИКА ЗАВЕРШЕНА ===")
}

// AnimalStats содержит статистику животных для диагностики
type AnimalStats struct {
	CurrentWolves   int
	CurrentRabbits  int
	AttackingWolves int
	EatingWolves    int
	Corpses         int
}

// printDiagnostics выводит диагностическую информацию о состоянии мира
func printDiagnostics(world *core.World, tick int) {
	second := tick / DiagnosticInterval
	stats := countAnimals(world)

	fmt.Printf("[%2ds] Живые зайцы: %2d, Трупы: %2d, Волки: %d (атакуют: %d, едят: %d)\n",
		second, stats.CurrentRabbits, stats.Corpses, stats.CurrentWolves, stats.AttackingWolves, stats.EatingWolves)

	// Детальная информация о волках
	if stats.CurrentWolves > 0 {
		printWolfDetails(world)
	}

	// Остановка если все волки мертвы
	if stats.CurrentWolves == 0 {
		fmt.Printf("❌ ВСЕ ВОЛКИ МЕРТВЫ на %d секунде!\n", second)
	}
}

// countAnimals подсчитывает животных и их состояния
func countAnimals(world *core.World) AnimalStats {
	var stats AnimalStats

	world.ForEachWith(core.MaskAnimalType, func(entity core.EntityID) {
		animalType, ok := world.GetAnimalType(entity)
		if !ok {
			return
		}

		if animalType == core.TypeWolf {
			stats.CurrentWolves++
			countWolfStates(world, entity, &stats)
		} else if animalType == core.TypeRabbit {
			countRabbitStates(world, entity, &stats)
		}
	})

	return stats
}

// countWolfStates подсчитывает состояния волков
func countWolfStates(world *core.World, entity core.EntityID, stats *AnimalStats) {
	if world.HasComponent(entity, core.MaskAttackState) {
		stats.AttackingWolves++
	}
	if world.HasComponent(entity, core.MaskEatingState) {
		stats.EatingWolves++
	}
}

// countRabbitStates подсчитывает состояния зайцев
func countRabbitStates(world *core.World, entity core.EntityID, stats *AnimalStats) {
	if world.HasComponent(entity, core.MaskCorpse) {
		stats.Corpses++
	} else {
		stats.CurrentRabbits++
	}
}

// printWolfDetails выводит детальную информацию о волках
func printWolfDetails(world *core.World) {
	fmt.Printf("      Детали волков:\n")
	wolfIndex := 0
	world.ForEachWith(core.MaskAnimalType|core.MaskSatiation, func(entity core.EntityID) {
		animalType, ok := world.GetAnimalType(entity)
		if !ok || animalType != core.TypeWolf {
			return
		}

		hunger, _ := world.GetSatiation(entity)
		health, _ := world.GetHealth(entity)

		status := "Живой"
		if health.Current <= 0 {
			status = "МЁРТВ"
		}

		// Детальная диагностика состояний
		attackState := world.HasComponent(entity, core.MaskAttackState)
		eatingState := world.HasComponent(entity, core.MaskEatingState)

		var eatingTarget core.EntityID = 0
		if eatingState {
			if eating, hasEating := world.GetEatingState(entity); hasEating {
				eatingTarget = eating.Target
			}
		}

		fmt.Printf("        Волк #%d: %s, голод %.1f%%, здоровье %d\n",
			wolfIndex+1, status, hunger.Value, health.Current)
		fmt.Printf("                  AttackState: %t, EatingState: %t (цель: %d)\n",
			attackState, eatingState, eatingTarget)
		wolfIndex++
	})
}
