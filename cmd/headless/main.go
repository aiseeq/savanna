package main

import (
	"flag"
	"fmt"
	"time"

	"github.com/aiseeq/savanna/internal/core"
	"github.com/aiseeq/savanna/internal/simulation"
)

// Параметры командной строки
var (
	duration = flag.Duration("duration", 30*time.Second, "Длительность симуляции")
	seed     = flag.Int64("seed", 42, "Seed для детерминированной симуляции")
	rabbits  = flag.Int("rabbits", 20, "Количество зайцев")
	wolves   = flag.Int("wolves", 3, "Количество волков")
	verbose  = flag.Bool("verbose", false, "Подробный вывод")
)

const (
	WORLD_SIZE = 50.0 * 32.0 // 50 тайлов по 32 пикселя
	TPS        = 60          // Тиков в секунду
)

func main() {
	flag.Parse()

	fmt.Printf("Запуск headless симуляции экосистемы саванны\n")
	fmt.Printf("Параметры: duration=%v, seed=%d, rabbits=%d, wolves=%d\n",
		*duration, *seed, *rabbits, *wolves)

	// Инициализация мира и систем
	world := core.NewWorld(WORLD_SIZE, WORLD_SIZE, *seed)
	systemManager := core.NewSystemManager()

	// Добавляем системы в правильном порядке
	systemManager.AddSystem(simulation.NewAnimalBehaviorSystem())
	systemManager.AddSystem(simulation.NewMovementSystem(WORLD_SIZE, WORLD_SIZE))
	systemManager.AddSystem(simulation.NewFeedingSystem())

	// Размещаем животных случайно по миру
	rng := world.GetRNG()

	fmt.Printf("Размещение %d зайцев...\n", *rabbits)
	for i := 0; i < *rabbits; i++ {
		x := rng.Float32()*WORLD_SIZE*0.8 + WORLD_SIZE*0.1 // Отступ от краёв
		y := rng.Float32()*WORLD_SIZE*0.8 + WORLD_SIZE*0.1
		simulation.CreateRabbit(world, x, y)
	}

	fmt.Printf("Размещение %d волков...\n", *wolves)
	for i := 0; i < *wolves; i++ {
		x := rng.Float32()*WORLD_SIZE*0.8 + WORLD_SIZE*0.1
		y := rng.Float32()*WORLD_SIZE*0.8 + WORLD_SIZE*0.1
		simulation.CreateWolf(world, x, y)
	}

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
