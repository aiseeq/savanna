package simulation

import (
	"testing"

	"github.com/aiseeq/savanna/internal/constants"
	"github.com/aiseeq/savanna/internal/core"
)

func TestSpeedVerification_OneTilePerSecond(t *testing.T) {
	// Создаем мир и систему движения
	world := core.NewWorld(100, 100, 12345)
	movementSystem := NewMovementSystem(100, 100)

	// Создаем животное с базовой скоростью 1.0 тайл/сек
	entity := CreateAnimal(world, core.TypeRabbit, 50, 50)

	// Устанавливаем скорость точно 1.0 тайл/сек по X (Y = 0)
	world.SetVelocity(entity, core.Velocity{X: 1.0, Y: 0.0})

	// Запоминаем начальную позицию
	initialPos, _ := world.GetPosition(entity)
	startX := initialPos.X
	startY := initialPos.Y

	// Симулируем ровно 1 секунду (60 тиков по 1/60 секунды)
	deltaTime := float32(1.0 / 60.0) // 1/60 секунды на тик
	for tick := 0; tick < 60; tick++ {
		movementSystem.Update(world, deltaTime)
	}

	// Проверяем финальную позицию
	finalPos, _ := world.GetPosition(entity)
	finalX := finalPos.X
	finalY := finalPos.Y

	// Вычисляем пройденное расстояние
	distanceX := finalX - startX
	distanceY := finalY - startY

	// КЛЮЧЕВАЯ ПРОВЕРКА: Животное должно пройти ровно 32 пикселя (1 тайл * 32 пикселя/тайл)
	expectedDistancePixels := float32(32.0) // 1 тайл = 32 пикселя
	tolerance := float32(0.1)               // Допуск для погрешностей плавающей точки

	// Проверяем движение по X
	if distanceX < expectedDistancePixels-tolerance || distanceX > expectedDistancePixels+tolerance {
		t.Errorf("Animal with 1.0 tile/sec speed should move exactly %f pixels in 1 second, but moved %f pixels",
			expectedDistancePixels, distanceX)
	}

	// Y не должно измениться (скорость по Y = 0)
	if distanceY < -tolerance || distanceY > tolerance {
		t.Errorf("Animal should not move in Y direction, but moved %f pixels", distanceY)
	}

	t.Logf("SUCCESS: Animal moved %f pixels in X (expected ~%f), %f pixels in Y (expected ~0)",
		distanceX, expectedDistancePixels, distanceY)
}

func TestSpeedVerification_DifferentSpeeds(t *testing.T) {
	testCases := []struct {
		name                 string
		speedTilesPerSec     float32
		expectedPixelsPerSec float32
	}{
		{"Half tile per second", 0.5, 16.0},
		{"One tile per second", 1.0, 32.0},
		{"Two tiles per second", 2.0, 64.0},
		{"Rabbit base speed", RabbitBaseSpeed, RabbitBaseSpeed * constants.TileSizePixels},
		{"Wolf base speed", WolfBaseSpeed, WolfBaseSpeed * constants.TileSizePixels},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Создаем новый мир для каждого теста
			world := core.NewWorld(100, 100, 12345)
			movementSystem := NewMovementSystem(100, 100)

			// Создаем животное
			entity := CreateAnimal(world, core.TypeRabbit, 50, 50)

			// Устанавливаем тестовую скорость
			world.SetVelocity(entity, core.Velocity{X: tc.speedTilesPerSec, Y: 0.0})

			// Запоминаем начальную позицию
			initialPos, _ := world.GetPosition(entity)
			startX := initialPos.X

			// Симулируем 1 секунду
			deltaTime := float32(1.0 / 60.0)
			for tick := 0; tick < 60; tick++ {
				movementSystem.Update(world, deltaTime)
			}

			// Проверяем результат
			finalPos, _ := world.GetPosition(entity)
			distanceX := finalPos.X - startX

			tolerance := float32(0.1)
			if distanceX < tc.expectedPixelsPerSec-tolerance || distanceX > tc.expectedPixelsPerSec+tolerance {
				t.Errorf("Speed %f tiles/sec should result in %f pixels moved, but got %f",
					tc.speedTilesPerSec, tc.expectedPixelsPerSec, distanceX)
			}

			t.Logf("Speed %f tiles/sec correctly moved %f pixels (expected %f)",
				tc.speedTilesPerSec, distanceX, tc.expectedPixelsPerSec)
		})
	}
}

func TestSpeedVerification_TilesToPixelsConversion(t *testing.T) {
	// Тест для проверки функции конвертации используемой в MovementSystem
	testCases := []struct {
		tilesPerSec float32
		expectedPix float32
	}{
		{0.5, 16.0},
		{1.0, 32.0},
		{1.5, 48.0},
		{2.0, 64.0},
		{RabbitBaseSpeed, RabbitBaseSpeed * 32.0},
		{WolfBaseSpeed, WolfBaseSpeed * 32.0},
	}

	for _, tc := range testCases {
		result := constants.TilesToPixels(tc.tilesPerSec)
		if result != tc.expectedPix {
			t.Errorf("TilesToPixels(%f) = %f, expected %f", tc.tilesPerSec, result, tc.expectedPix)
		}
	}
}

func TestSpeedVerification_MovementSystemIntegration(t *testing.T) {
	// Интеграционный тест: проверяем что вся цепочка Animal Factory -> Movement System работает правильно
	world := core.NewWorld(100, 100, 12345)
	movementSystem := NewMovementSystem(100, 100)

	// Создаем зайца через фабрику (он должен иметь правильную скорость из game_balance.go)
	rabbit := CreateAnimal(world, core.TypeRabbit, 50, 50)

	// Проверяем что у зайца правильная базовая скорость
	speed, hasSpeed := world.GetSpeed(rabbit)
	if !hasSpeed {
		t.Fatal("Rabbit should have Speed component")
	}
	if speed.Base != RabbitBaseSpeed {
		t.Errorf("Rabbit base speed should be %f, got %f", RabbitBaseSpeed, speed.Base)
	}

	// Устанавливаем скорость равную базовой скорости
	world.SetVelocity(rabbit, core.Velocity{X: speed.Base, Y: 0.0})

	// Запоминаем начальную позицию
	initialPos, _ := world.GetPosition(rabbit)
	startX := initialPos.X

	// Симулируем 1 секунду
	deltaTime := float32(1.0 / 60.0)
	for tick := 0; tick < 60; tick++ {
		movementSystem.Update(world, deltaTime)
	}

	// Проверяем результат
	finalPos, _ := world.GetPosition(rabbit)
	distanceX := finalPos.X - startX

	// Ожидаемое расстояние = базовая скорость * 32 пикселя/тайл
	expectedDistance := float32(RabbitBaseSpeed * constants.TileSizePixels)
	tolerance := float32(0.1)

	// Проверяем движение
	if distanceX < expectedDistance-tolerance || distanceX > expectedDistance+tolerance {
		t.Errorf("Rabbit with base speed %f tiles/sec should move %f pixels, but moved %f",
			RabbitBaseSpeed, expectedDistance, distanceX)
	}

	t.Logf("SUCCESS: Rabbit with base speed %f tiles/sec moved %f pixels (expected %f)",
		RabbitBaseSpeed, distanceX, expectedDistance)
}
