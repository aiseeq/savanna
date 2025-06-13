package main

import (
	"flag"
	"fmt"
	"time"

	"github.com/aiseeq/savanna/config"
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

	// Загружаем конфигурацию
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
		return
	}

	fmt.Printf("Параметры: duration=%v, seed=%d, world_size=%d\n",
		*duration, cfg.World.Seed, cfg.World.Size)
	fmt.Printf("Популяции: %d зайцев, %d волков\n",
		cfg.Population.Rabbits, cfg.Population.Wolves)

	// Генерируем мир
	fmt.Println("Генерация ландшафта...")
	terrainGen := generator.NewTerrainGenerator(cfg)
	terrain := terrainGen.Generate()

	terrainStats := terrain.GetStats()
	fmt.Printf("Ландшафт: %d тайлов травы, %d воды, %d кустов, %d влажных\n",
		terrainStats["grass_tiles"], terrainStats["water_tiles"],
		terrainStats["bush_tiles"], terrainStats["wetland_tiles"])

	// Инициализация мира и систем
	worldSizePixels := float32(cfg.World.Size * 32)
	world := core.NewWorld(worldSizePixels, worldSizePixels, cfg.World.Seed)
	systemManager := core.NewSystemManager()

	// Создаём системы с зависимостями
	vegetationSystem := simulation.NewVegetationSystem(terrain)
	animalBehaviorSystem := simulation.NewAnimalBehaviorSystem(vegetationSystem)
	feedingSystem := simulation.NewFeedingSystem(vegetationSystem)

	// Добавляем системы в правильном порядке
	systemManager.AddSystem(vegetationSystem)
	systemManager.AddSystem(animalBehaviorSystem)
	systemManager.AddSystem(simulation.NewMovementSystem(worldSizePixels, worldSizePixels))
	systemManager.AddSystem(feedingSystem)

	// Размещаем животных с помощью генератора
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

	startTime := time.Now()
	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()

	lastStats := world.GetStats()
	events := ""

	fmt.Println("Время | Зайцы | Волки | События")
	fmt.Println("------|-------|-------|--------")
	fmt.Printf("%5.0fs | %5d | %5d | Начало симуляции\n",
		0.0, lastStats[core.TypeRabbit], lastStats[core.TypeWolf])

	deltaTime := float32(1.0 / TPS)
	ticksPerSecond := 0

	// Основной цикл симуляции
	for elapsed := time.Duration(0); elapsed < *duration; {
		select {
		case <-ticker.C:
			elapsed = time.Since(startTime)

			// Получаем статистику и сравниваем с предыдущей
			currentStats := world.GetStats()
			events = getEvents(lastStats, currentStats)

			fmt.Printf("%5.0fs | %5d | %5d | %s\n",
				elapsed.Seconds(),
				currentStats[core.TypeRabbit],
				currentStats[core.TypeWolf],
				events)

			if *verbose {
				fmt.Printf("        TPS: %d\n", ticksPerSecond)
			}

			lastStats = currentStats
			ticksPerSecond = 0

		default:
			// Быстрая симуляция
			world.Update(deltaTime)
			systemManager.Update(world, deltaTime)
			ticksPerSecond++

			// Небольшая пауза чтобы не перегружать CPU
			time.Sleep(time.Microsecond * 100)
		}
	}

	fmt.Printf("\nСимуляция завершена за %v\n", time.Since(startTime))

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
