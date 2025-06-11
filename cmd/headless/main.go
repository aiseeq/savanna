package main

import (
	"flag"
	"fmt"
	"time"
)

// Параметры командной строки
var (
	duration = flag.Duration("duration", 60*time.Second, "Длительность симуляции")
	seed     = flag.Int64("seed", 42, "Seed для детерминированной симуляции")
	rabbits  = flag.Int("rabbits", 20, "Количество зайцев")
	wolves   = flag.Int("wolves", 3, "Количество волков")
	verbose  = flag.Bool("verbose", false, "Подробный вывод")
)

func main() {
	flag.Parse()

	fmt.Printf("Запуск headless симуляции экосистемы саванны\n")
	fmt.Printf("Параметры: duration=%v, seed=%d, rabbits=%d, wolves=%d\n",
		*duration, *seed, *rabbits, *wolves)

	// TODO: Инициализация мира
	// world := core.NewWorld(*seed)
	// world.SpawnRabbits(*rabbits)
	// world.SpawnWolves(*wolves)

	startTime := time.Now()
	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()

	fmt.Println("Время | Зайцы | Волки | События")
	fmt.Println("------|-------|-------|--------")

	// Основной цикл симуляции
	for elapsed := time.Duration(0); elapsed < *duration; {
		select {
		case <-ticker.C:
			elapsed = time.Since(startTime)

			// TODO: Обновление симуляции
			// world.Update(1.0 / 60.0) // 60 TPS

			// Вывод статистики каждую секунду
			rabbitsAlive := *rabbits // TODO: world.CountRabbits()
			wolvesAlive := *wolves   // TODO: world.CountWolves()

			fmt.Printf("%5.0fs | %5d | %5d | В разработке...\n",
				elapsed.Seconds(), rabbitsAlive, wolvesAlive)

		default:
			// TODO: Быстрая симуляция без ожидания
			// world.Update(1.0 / 60.0)
			time.Sleep(time.Millisecond) // Временная заглушка
		}
	}

	fmt.Printf("\nСимуляция завершена за %v\n", time.Since(startTime))

	// TODO: Финальная статистика
	// fmt.Printf("Финальное состояние: %d зайцев, %d волков\n",
	//	world.CountRabbits(), world.CountWolves())
}
