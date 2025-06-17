package main

import (
	"flag"
	"fmt"
	"time"

	"github.com/aiseeq/savanna/config"
	"github.com/aiseeq/savanna/internal/adapters"
	"github.com/aiseeq/savanna/internal/animation"
	"github.com/aiseeq/savanna/internal/core"
	"github.com/aiseeq/savanna/internal/generator"
	"github.com/aiseeq/savanna/internal/simulation"
)

// Параметры командной строки
var (
	duration   = flag.Duration("duration", 30*time.Second, "Длительность симуляции")
	configFile = flag.String("config", "", "Путь к файлу конфигурации")
	seed       = flag.Int64("seed", 0, "Seed для детерминированной симуляции (0 = из конфига)")
	verbose    = flag.Bool("verbose", false, "Подробный вывод")
)

const (
	TPS = 60 // Тиков в секунду
)

func main() {
	flag.Parse()
	fmt.Printf("Запуск headless симуляции экосистемы саванны\n")

	cfg := loadAndValidateConfig()
	if cfg == nil {
		return
	}

	printSimulationParameters(cfg)
	runSimulation(cfg)
}

// loadAndValidateConfig загружает и валидирует конфигурацию
func loadAndValidateConfig() *config.Config {
	var cfg *config.Config
	var err error

	if *configFile != "" {
		cfg, err = config.LoadConfig(*configFile)
		if err != nil {
			fmt.Printf("Ошибка загрузки конфигурации: %v\n", err)
			fmt.Println("Используем конфигурацию по умолчанию")
			cfg = config.LoadDefaultConfig()
		}
	} else {
		cfg = config.LoadDefaultConfig()
	}

	// Переопределяем seed если указан в командной строке
	if *seed != 0 {
		cfg.World.Seed = *seed
	}

	// Валидируем конфигурацию
	if err := cfg.Validate(); err != nil {
		fmt.Printf("Ошибка валидации конфигурации: %v\n", err)
		return nil
	}

	return cfg
}

// printSimulationParameters выводит параметры симуляции
func printSimulationParameters(cfg *config.Config) {
	fmt.Printf("Параметры: duration=%v, seed=%d, world_size=%d\n",
		*duration, cfg.World.Seed, cfg.World.Size)
	fmt.Printf("Популяции: %d зайцев, %d волков\n",
		cfg.Population.Rabbits, cfg.Population.Wolves)
}

// runSimulation запускает основной цикл симуляции
func runSimulation(cfg *config.Config) {
	terrain := generateTerrain(cfg)
	world, systemManager, wolfAnimSystem, rabbitAnimSystem := initializeWorldAndSystems(cfg, terrain)
	populateWorld(cfg, world, terrain)
	runMainLoop(world, systemManager, wolfAnimSystem, rabbitAnimSystem)
}

// generateTerrain генерирует ландшафт мира
func generateTerrain(cfg *config.Config) *generator.Terrain {
	fmt.Println("Генерация ландшафта...")
	terrainGen := generator.NewTerrainGenerator(cfg)
	terrain := terrainGen.Generate()

	terrainStats := terrain.GetStats()
	fmt.Printf("Ландшафт: %d тайлов травы, %d воды, %d кустов, %d влажных\n",
		terrainStats["grass_tiles"], terrainStats["water_tiles"],
		terrainStats["bush_tiles"], terrainStats["wetland_tiles"])

	return terrain
}

// initializeWorldAndSystems инициализирует мир и системы
func initializeWorldAndSystems(cfg *config.Config, terrain *generator.Terrain) (*core.World, *core.SystemManager, *animation.AnimationSystem, *animation.AnimationSystem) {
	worldSizePixels := float32(cfg.World.Size * 32)
	world := core.NewWorld(worldSizePixels, worldSizePixels, cfg.World.Seed)
	systemManager := core.NewSystemManager()

	// КРИТИЧЕСКИ ВАЖНО: создаём анимационные системы для headless режима
	wolfAnimationSystem := animation.NewAnimationSystem()
	rabbitAnimationSystem := animation.NewAnimationSystem()

	// Загружаем анимации через общий загрузчик (устраняет дублирование)
	loader := animation.NewAnimationLoader()
	loader.LoadHeadlessAnimations(wolfAnimationSystem, rabbitAnimationSystem)

	// Создаём системы с зависимостями
	vegetationSystem := simulation.NewVegetationSystem(terrain)
	animalBehaviorSystem := simulation.NewAnimalBehaviorSystem(vegetationSystem)
	feedingSystem := simulation.NewFeedingSystem(vegetationSystem)
	grassEatingSystem := simulation.NewGrassEatingSystem(vegetationSystem)
	combatSystem := simulation.NewCombatSystem()

	// Добавляем системы в правильном порядке (ВАЖНО для корректной работы еды!)
	// Используем адаптеры для систем с ISP интерфейсами
	systemManager.AddSystem(vegetationSystem)
	systemManager.AddSystem(&adapters.FeedingSystemAdapter{System: feedingSystem})                                                   // 1. Создаёт EatingState для едящих животных
	systemManager.AddSystem(grassEatingSystem)                                                                                       // 2. Дискретное поедание травы по кадрам анимации
	systemManager.AddSystem(&adapters.BehaviorSystemAdapter{System: animalBehaviorSystem})                                           // 3. Проверяет EatingState и не мешает еде
	systemManager.AddSystem(&adapters.MovementSystemAdapter{System: simulation.NewMovementSystem(worldSizePixels, worldSizePixels)}) // 4. Сбрасывает скорость едящих животных
	systemManager.AddSystem(combatSystem)                                                                                            // 5. Система боя работает после движения

	return world, systemManager, wolfAnimationSystem, rabbitAnimationSystem
}

// populateWorld размещает животных в мире
func populateWorld(cfg *config.Config, world *core.World, terrain *generator.Terrain) {
	fmt.Println("Размещение животных...")
	popGen := generator.NewPopulationGenerator(cfg, terrain)
	placements := popGen.Generate()

	// Создаём животных на основе сгенерированных позиций
	for _, placement := range placements {
		switch placement.Type {
		case core.TypeRabbit:
			simulation.CreateRabbit(world, placement.X, placement.Y)
		case core.TypeWolf:
			simulation.CreateWolf(world, placement.X, placement.Y)
		}
	}

	// Проверяем корректность размещения
	errors := popGen.ValidatePlacement(placements)
	if len(errors) > 0 {
		fmt.Printf("Предупреждения размещения: %v\n", errors)
	}

	popStats := popGen.GetStats(placements)
	fmt.Printf("Размещено: %d зайцев (%d групп), %d волков\n",
		popStats["rabbits"], popStats["rabbit_groups"], popStats["wolves"])
}

// runMainLoop запускает основной цикл симуляции
// simulationState состояние симуляции (устраняет нарушение SRP)
type simulationState struct {
	startTime      time.Time
	ticker         *time.Ticker
	lastStats      map[core.AnimalType]int
	events         string
	deltaTime      float32
	ticksPerSecond int
}

// initializeSimulation инициализация симуляции
func initializeSimulation(world *core.World) *simulationState {
	state := &simulationState{
		startTime:      time.Now(),
		ticker:         time.NewTicker(time.Second),
		lastStats:      world.GetStats(),
		events:         "",
		deltaTime:      float32(1.0 / TPS),
		ticksPerSecond: 0,
	}

	fmt.Println("Время | Зайцы | Волки | События")
	fmt.Println("------|-------|-------|--------")
	fmt.Printf("%5.0fs | %5d | %5d | Начало симуляции\n",
		0.0, state.lastStats[core.TypeRabbit], state.lastStats[core.TypeWolf])

	return state
}

func runMainLoop(world *core.World, systemManager *core.SystemManager, wolfAnimationSystem, rabbitAnimationSystem *animation.AnimationSystem) {
	state := initializeSimulation(world)
	defer state.ticker.Stop()

	// Создаём менеджер анимаций (устраняет дублирование логики)
	animationManager := animation.NewAnimationManager(wolfAnimationSystem, rabbitAnimationSystem)

	// Обновляем анимации через менеджер (устраняет нарушение SRP)
	updateAnimalAnimations := func() {
		animationManager.UpdateAllAnimations(world, state.deltaTime)
	}

	// Основной цикл симуляции
	for elapsed := time.Duration(0); elapsed < *duration; {
		select {
		case <-state.ticker.C:
			elapsed = time.Since(state.startTime)

			// Получаем статистику и сравниваем с предыдущей
			currentStats := world.GetStats()
			state.events = getEvents(state.lastStats, currentStats)

			fmt.Printf("%5.0fs | %5d | %5d | %s\n",
				elapsed.Seconds(),
				currentStats[core.TypeRabbit],
				currentStats[core.TypeWolf],
				state.events)

			if *verbose {
				fmt.Printf("        TPS: %d\n", state.ticksPerSecond)
			}

			state.lastStats = currentStats
			state.ticksPerSecond = 0

		default:
			// Быстрая симуляция
			world.Update(state.deltaTime)
			systemManager.Update(world, state.deltaTime)

			// КРИТИЧЕСКИ ВАЖНО: обновляем анимации для корректной работы боевой системы
			updateAnimalAnimations()

			state.ticksPerSecond++

			// Небольшая пауза чтобы не перегружать CPU
			time.Sleep(time.Microsecond * 100)
		}
	}

	fmt.Printf("\nСимуляция завершена за %v\n", time.Since(state.startTime))

	// Финальная статистика
	finalStats := world.GetStats()
	fmt.Printf("Финальное состояние: %d зайцев, %d волков\n",
		finalStats[core.TypeRabbit], finalStats[core.TypeWolf])

	if finalStats[core.TypeRabbit] == 0 {
		fmt.Println("⚰️  Все зайцы вымерли")
	}
	if finalStats[core.TypeWolf] == 0 {
		fmt.Println("⚰️  Все волки вымерли")
	}
}

// getEvents анализирует изменения в популяции и возвращает описание событий
func getEvents(oldStats, newStats map[core.AnimalType]int) string {
	rabbitChange := newStats[core.TypeRabbit] - oldStats[core.TypeRabbit]
	wolfChange := newStats[core.TypeWolf] - oldStats[core.TypeWolf]

	events := []string{}

	if rabbitChange < 0 {
		events = append(events, fmt.Sprintf("%d зайца погибло", -rabbitChange))
	}
	if wolfChange < 0 {
		events = append(events, fmt.Sprintf("%d волка погибло", -wolfChange))
	}
	if rabbitChange == 0 && wolfChange == 0 {
		events = append(events, "Стабильно")
	}

	if len(events) == 0 {
		return "Спокойно"
	}

	result := ""
	for i, event := range events {
		if i > 0 {
			result += ", "
		}
		result += event
	}
	return result
}
