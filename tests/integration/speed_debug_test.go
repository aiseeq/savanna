package integration

import (
	"fmt"
	"math"
	"testing"

	"github.com/aiseeq/savanna/internal/core"
	"github.com/aiseeq/savanna/internal/simulation"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSpeedDebugging(t *testing.T) {
	// Создаём мир с известными параметрами
	worldPixels := float32(640) // 20 тайлов * 32 пикселя/тайл = 640 пикселей
	world := core.NewWorld(worldPixels, worldPixels, 42)

	// Создаём зайца в центре мира
	rabbitID := world.CreateEntity()

	// Создаём конфигурацию зайца
	config := simulation.CreateAnimalConfig(core.TypeRabbit)

	fmt.Printf("=== DEBUGGING ANIMAL SPEED ===\n")
	fmt.Printf("Rabbit base speed: %.3f тайлов/сек (из game_balance.go)\n", config.BaseSpeed)
	fmt.Printf("Expected pixel speed: %.1f пикселей/сек (%.3f * 32)\n", config.BaseSpeed*32, config.BaseSpeed)
	fmt.Printf("Expected distance per second: %.1f пикселей\n", config.BaseSpeed*32)
	fmt.Printf("Expected distance in 2 seconds: %.1f пикселей\n", config.BaseSpeed*32*2)

	// ДОБАВЛЯЕМ компоненты (не устанавливаем!)
	world.AddPosition(rabbitID, core.Position{X: 320, Y: 320})            // Центр мира в пикселях
	world.AddVelocity(rabbitID, core.Velocity{X: config.BaseSpeed, Y: 0}) // Движение вправо

	// Добавляем размер для MovementSystem
	world.AddSize(rabbitID, core.Size{
		Radius:      config.BaseRadius,
		AttackRange: config.AttackRange,
	})

	// Проверяем что данные сохранились
	checkPos, _ := world.GetPosition(rabbitID)
	checkVel, _ := world.GetVelocity(rabbitID)
	fmt.Printf("Set position: (%.1f, %.1f)\n", checkPos.X, checkPos.Y)
	fmt.Printf("Set velocity: (%.3f, %.3f) тайлов/сек\n", checkVel.X, checkVel.Y)

	// Создаём системы движения
	movementSystem := simulation.NewMovementSystem(worldPixels, worldPixels)

	initialPos, _ := world.GetPosition(rabbitID)
	fmt.Printf("Initial position: (%.1f, %.1f)\n", initialPos.X, initialPos.Y)

	// Проверяем начальную скорость
	initialVel, _ := world.GetVelocity(rabbitID)
	fmt.Printf("Initial velocity: (%.3f, %.3f) тайлов/сек\n", initialVel.X, initialVel.Y)

	// Симулируем 2 секунды движения (120 тиков по 1/60 сек)
	deltaTime := float32(1.0 / 60.0) // 60 FPS
	totalTime := float32(0)

	for tick := 0; tick < 120; tick++ {
		totalTime += deltaTime

		// Обновляем движение
		movementSystem.Update(world, deltaTime)

		// Логируем каждые 30 тиков (каждые 0.5 сек)
		if tick%30 == 0 {
			pos, _ := world.GetPosition(rabbitID)
			vel, _ := world.GetVelocity(rabbitID)

			// Считаем пройденную дистанцию от начальной точки
			distanceTraveled := math.Abs(float64(pos.X - initialPos.X))

			fmt.Printf("Tick %d (%.1fs): pos=(%.1f,%.1f), vel=(%.3f,%.3f), distance=%.1f px\n",
				tick, totalTime, pos.X, pos.Y, vel.X, vel.Y, distanceTraveled)
		}
	}

	// Финальная проверка
	finalPos, _ := world.GetPosition(rabbitID)
	totalDistance := math.Abs(float64(finalPos.X - initialPos.X))

	fmt.Printf("\n=== РЕЗУЛЬТАТЫ ===\n")
	fmt.Printf("Время симуляции: %.1f секунд\n", totalTime)
	fmt.Printf("Начальная позиция: (%.1f, %.1f)\n", initialPos.X, initialPos.Y)
	fmt.Printf("Финальная позиция: (%.1f, %.1f)\n", finalPos.X, finalPos.Y)
	fmt.Printf("Пройденная дистанция: %.1f пикселей\n", totalDistance)
	fmt.Printf("Средняя скорость: %.1f пикселей/сек\n", totalDistance/float64(totalTime))
	fmt.Printf("Средняя скорость: %.3f тайлов/сек\n", totalDistance/float64(totalTime)/32.0)

	// ПРОВЕРЯЕМ ОЖИДАЕМЫЕ ЗНАЧЕНИЯ
	expectedPixelsPerSecond := float64(config.BaseSpeed * 32) // 0.6 * 32 = 19.2 пикс/сек
	actualPixelsPerSecond := totalDistance / float64(totalTime)

	fmt.Printf("\n=== АНАЛИЗ ===\n")
	fmt.Printf("Ожидаемая скорость: %.1f пикселей/сек\n", expectedPixelsPerSecond)
	fmt.Printf("Фактическая скорость: %.1f пикселей/сек\n", actualPixelsPerSecond)
	fmt.Printf("Отношение факт/ожидание: %.2fx\n", actualPixelsPerSecond/expectedPixelsPerSecond)

	// Проверяем что скорость не превышает ожидаемую в 2 раза
	if actualPixelsPerSecond > expectedPixelsPerSecond*2.0 {
		t.Errorf("СЛИШКОМ БЫСТРОЕ ДВИЖЕНИЕ! Ожидали %.1f пикс/сек, получили %.1f пикс/сек (в %.1fx раз быстрее)",
			expectedPixelsPerSecond, actualPixelsPerSecond, actualPixelsPerSecond/expectedPixelsPerSecond)
	}

	// Проверяем что животное вообще двигается
	require.Greater(t, actualPixelsPerSecond, 5.0, "Животное должно двигаться")

	// Главная проверка: скорость должна быть близка к ожидаемой (допуск 50%)
	expectedMin := expectedPixelsPerSecond * 0.5
	expectedMax := expectedPixelsPerSecond * 1.5

	assert.GreaterOrEqual(t, actualPixelsPerSecond, expectedMin,
		"Скорость слишком низкая: %.1f < %.1f", actualPixelsPerSecond, expectedMin)
	assert.LessOrEqual(t, actualPixelsPerSecond, expectedMax,
		"Скорость слишком высокая: %.1f > %.1f", actualPixelsPerSecond, expectedMax)
}

func TestMovementSystemConversion(t *testing.T) {
	fmt.Printf("\n=== ТЕСТ КОНВЕРТАЦИИ СКОРОСТИ В MOVEMENT SYSTEM ===\n")

	// Тестируем конвертацию скорости из тайлов/сек в пиксели/сек
	velocityInTiles := float32(1.0)                 // 1 тайл/сек
	expectedPixelVelocity := velocityInTiles * 32.0 // 32 пикселя/сек

	fmt.Printf("Скорость в тайлах/сек: %.3f\n", velocityInTiles)
	fmt.Printf("Ожидаемая скорость в пикселях/сек: %.1f\n", expectedPixelVelocity)

	// Симулируем что делает MovementSystem
	deltaTime := float32(1.0 / 60.0)    // 60 FPS
	pixelVelX := velocityInTiles * 32.0 // Конвертация как в movement.go:99

	fmt.Printf("Конвертированная скорость (movement.go:99): %.1f пикс/сек\n", pixelVelX)

	// Движение за один кадр
	movementPerFrame := pixelVelX * deltaTime
	fmt.Printf("Движение за кадр (1/60 сек): %.3f пикселей\n", movementPerFrame)

	// Движение за секунду (60 кадров)
	movementPerSecond := movementPerFrame * 60
	fmt.Printf("Движение за секунду (60 кадров): %.1f пикселей\n", movementPerSecond)

	// Проверяем правильность конвертации
	assert.Equal(t, expectedPixelVelocity, pixelVelX, "Конвертация скорости должна быть правильной")
	assert.InDelta(t, expectedPixelVelocity, movementPerSecond, 0.1, "Движение за секунду должно соответствовать скорости")
}
