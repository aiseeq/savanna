package integration

import (
	"fmt"
	"math"
	"testing"

	"github.com/aiseeq/savanna/tests/common"
	"github.com/stretchr/testify/assert"
)

func TestFullSystemSpeedBehavior(t *testing.T) {
	fmt.Printf("=== ПОЛНЫЙ ТЕСТ СКОРОСТИ С СИСТЕМАМИ ===\n")

	// Создаём мир и системы как в реальной игре
	world, systemManager, entities := common.NewTestWorld().
		WithSize(640).
		AddRabbit(300, 300, 20.0, 50). // Очень голодный заяц (20% сытости = высокий голод)
		Build()

	rabbitID := entities.Rabbits[0]

	// Получаем конфигурацию зайца
	config, _ := world.GetAnimalConfig(rabbitID)
	speed, _ := world.GetSpeed(rabbitID)
	hunger, _ := world.GetSatiation(rabbitID)

	fmt.Printf("=== НАСТРОЙКИ ЗАЙЦА ===\n")
	fmt.Printf("Base speed: %.3f тайлов/сек\n", config.BaseSpeed)
	fmt.Printf("Current speed: %.3f тайлов/сек\n", speed.Current)
	fmt.Printf("Hunger: %.1f%% (threshold: %.1f%%)\n", hunger.Value, config.SatiationThreshold)
	fmt.Printf("Search speed multiplier: %.1f\n", config.SearchSpeed)
	// ТИПОБЕЗОПАСНОСТЬ: конвертируем physics.TilesPerSecond в float32 для операций
	fmt.Printf("Expected search speed: %.3f тайлов/сек\n", config.BaseSpeed*config.SearchSpeed)
	fmt.Printf("Expected search speed in pixels: %.1f пикс/сек\n", config.BaseSpeed*config.SearchSpeed*32)

	initialPos, _ := world.GetPosition(rabbitID)
	fmt.Printf("\nInitial position: (%.1f, %.1f)\n", initialPos.X, initialPos.Y)

	// Запускаем симуляцию с полными системами
	deltaTime := float32(1.0 / 60.0) // 60 FPS
	totalTime := float32(0)

	fmt.Printf("\n=== ДВИЖЕНИЕ ЗАЙЦА ПО СЕКУНДАМ ===\n")

	for tick := 0; tick < 180; tick++ { // 3 секунды
		totalTime += deltaTime

		// Обновляем все системы как в игре
		systemManager.Update(world, deltaTime)

		// Логируем каждую секунду
		if tick%60 == 0 {
			pos, _ := world.GetPosition(rabbitID)
			vel, _ := world.GetVelocity(rabbitID)
			speed, _ := world.GetSpeed(rabbitID)
			hunger, _ := world.GetSatiation(rabbitID)

			// Считаем пройденную дистанцию от начальной точки
			dx := pos.X - initialPos.X
			dy := pos.Y - initialPos.Y
			distanceTraveled := math.Sqrt(float64(dx*dx + dy*dy))

			// Текущая скорость в пикселях/сек
			currentPixelSpeed := math.Sqrt(float64(vel.X*vel.X+vel.Y*vel.Y)) * 32

			fmt.Printf("Секунда %d: pos=(%.1f,%.1f), vel=(%.3f,%.3f), speed=%.3f, hunger=%.1f%%, distance=%.1f px, pixel_speed=%.1f\n",
				tick/60, pos.X, pos.Y, vel.X, vel.Y, speed.Current, hunger.Value, distanceTraveled, currentPixelSpeed)
		}
	}

	// Финальный анализ
	finalPos, _ := world.GetPosition(rabbitID)
	dx := finalPos.X - initialPos.X
	dy := finalPos.Y - initialPos.Y
	totalDistance := math.Sqrt(float64(dx*dx + dy*dy))

	averageSpeed := totalDistance / float64(totalTime)

	fmt.Printf("\n=== ИТОГОВЫЙ АНАЛИЗ ===\n")
	fmt.Printf("Время симуляции: %.1f секунд\n", totalTime)
	fmt.Printf("Начальная позиция: (%.1f, %.1f)\n", initialPos.X, initialPos.Y)
	fmt.Printf("Финальная позиция: (%.1f, %.1f)\n", finalPos.X, finalPos.Y)
	fmt.Printf("Пройденная дистанция: %.1f пикселей\n", totalDistance)
	fmt.Printf("Средняя скорость: %.1f пикселей/сек\n", averageSpeed)
	fmt.Printf("Средняя скорость: %.3f тайлов/сек\n", averageSpeed/32.0)

	// Ожидаемая скорость: голодный заяц ищет еду с множителем SearchSpeed
	expectedTileSpeed := float64(config.BaseSpeed * config.SearchSpeed) // 0.6 * 0.8 = 0.48 тайла/сек
	expectedPixelSpeed := expectedTileSpeed * 32                        // 0.48 * 32 = 15.36 пикс/сек

	fmt.Printf("\n=== СРАВНЕНИЕ С ОЖИДАНИЯМИ ===\n")
	fmt.Printf("Ожидаемая скорость: %.1f пикселей/сек (голодный заяц ищет еду)\n", expectedPixelSpeed)
	fmt.Printf("Фактическая скорость: %.1f пикселей/сек\n", averageSpeed)
	fmt.Printf("Отношение факт/ожидание: %.2fx\n", averageSpeed/expectedPixelSpeed)

	// Проверяем что скорость разумная (не в 10 раз быстрее)
	if averageSpeed > expectedPixelSpeed*5.0 {
		t.Errorf("СЛИШКОМ БЫСТРОЕ ДВИЖЕНИЕ! Ожидали %.1f пикс/сек, получили %.1f пикс/сек (в %.1fx раз быстрее)",
			expectedPixelSpeed, averageSpeed, averageSpeed/expectedPixelSpeed)
	}

	// Проверяем что животное вообще двигается
	assert.Greater(t, averageSpeed, 5.0, "Животное должно двигаться")

	// Проверяем что скорость в пределах разумного (допуск 200%)
	assert.LessOrEqual(t, averageSpeed, expectedPixelSpeed*3.0,
		"Скорость слишком высокая: %.1f > %.1f", averageSpeed, expectedPixelSpeed*3.0)
}

func TestWolfSpeedComparison(t *testing.T) {
	fmt.Printf("\n=== ТЕСТ СКОРОСТИ ВОЛКА ===\n")

	// Создаём голодного волка
	world, systemManager, entities := common.NewTestWorld().
		WithSize(640).
		AddHungryWolf().
		Build()

	wolfID := entities.Wolves[0]

	// Получаем конфигурацию волка
	config, _ := world.GetAnimalConfig(wolfID)
	speed, _ := world.GetSpeed(wolfID)
	hunger, _ := world.GetSatiation(wolfID)

	fmt.Printf("=== НАСТРОЙКИ ВОЛКА ===\n")
	fmt.Printf("Base speed: %.3f тайлов/сек\n", config.BaseSpeed)
	fmt.Printf("Current speed: %.3f тайлов/сек\n", speed.Current)
	fmt.Printf("Hunger: %.1f%% (threshold: %.1f%%)\n", hunger.Value, config.SatiationThreshold)
	fmt.Printf("Wandering speed multiplier: %.1f\n", config.WanderingSpeed)
	// ТИПОБЕЗОПАСНОСТЬ: конвертируем physics.TilesPerSecond в float32 для операций
	fmt.Printf("Expected wandering speed: %.3f тайлов/сек\n", config.BaseSpeed*config.WanderingSpeed)
	fmt.Printf("Expected wandering speed in pixels: %.1f пикс/сек\n", config.BaseSpeed*config.WanderingSpeed*32)

	initialPos, _ := world.GetPosition(wolfID)

	// Запускаем симуляцию
	deltaTime := float32(1.0 / 60.0)

	for tick := 0; tick < 120; tick++ { // 2 секунды
		systemManager.Update(world, deltaTime)
	}

	// Анализ скорости волка
	finalPos, _ := world.GetPosition(wolfID)
	dx := finalPos.X - initialPos.X
	dy := finalPos.Y - initialPos.Y
	totalDistance := math.Sqrt(float64(dx*dx + dy*dy))

	averageSpeed := totalDistance / 2.0 // 2 секунды

	// Ожидаемая скорость волка при блуждании
	expectedTileSpeed := float64(config.BaseSpeed * config.WanderingSpeed) // 1.0 * 0.7 = 0.7 тайла/сек
	expectedPixelSpeed := expectedTileSpeed * 32                           // 0.7 * 32 = 22.4 пикс/сек

	fmt.Printf("Волк - ожидаемая скорость: %.1f пикс/сек\n", expectedPixelSpeed)
	fmt.Printf("Волк - фактическая скорость: %.1f пикс/сек\n", averageSpeed)
	fmt.Printf("Отношение: %.2fx\n", averageSpeed/expectedPixelSpeed)

	// Проверяем что волк не слишком быстрый
	assert.LessOrEqual(t, averageSpeed, expectedPixelSpeed*3.0,
		"Волк движется слишком быстро: %.1f > %.1f", averageSpeed, expectedPixelSpeed*3.0)
}
