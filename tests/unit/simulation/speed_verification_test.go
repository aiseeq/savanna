package simulation

import (
	"testing"

	"github.com/aiseeq/savanna/internal/constants"
	"github.com/aiseeq/savanna/internal/core"
	"github.com/aiseeq/savanna/internal/simulation"
)

// TestMovementSpeedConversion проверяет корректность конвертации скорости из тайлов/сек в пиксели/сек
// Соответствует задаче 1.4 из плана рефакторинга
func TestMovementSpeedConversion(t *testing.T) {
	// Создаем мир и движение системы
	worldWidth := float32(50)
	worldHeight := float32(38)
	world := core.NewWorld(worldWidth, worldHeight, 12345)
	movementSystem := simulation.NewMovementSystem(worldWidth, worldHeight)

	// Создаем животное с базовой скоростью 1.0 тайл/сек
	entity := simulation.CreateAnimal(world, core.TypeRabbit, 25.0, 19.0)

	// Получаем начальную позицию
	initialPos, _ := world.GetPosition(entity)

	// Устанавливаем скорость 1.0 тайл/сек в направлении X
	world.SetVelocity(entity, core.Velocity{X: 1.0, Y: 0.0})

	// Симулируем движение на 1 секунду (60 тиков по 1/60 сек)
	deltaTime := float32(1.0 / 60.0)
	for i := 0; i < 60; i++ {
		movementSystem.Update(world, deltaTime)
	}

	// Получаем финальную позицию
	finalPos, _ := world.GetPosition(entity)

	// Вычисляем расстояние, пройденное за секунду
	distanceTraveled := finalPos.X - initialPos.X

	// Ожидаемое расстояние: 1 тайл = 32 пикселя
	expectedDistance := constants.TilesToPixels(1.0) // 32 пикселя

	// Проверяем с небольшой погрешностью (0.1 пикселя)
	tolerance := float32(0.1)
	if abs(distanceTraveled-expectedDistance) > tolerance {
		t.Errorf(
			"Неверная конвертация скорости: животное со скоростью 1.0 тайл/сек прошло %.2f пикселей за секунду, ожидалось %.2f",
			distanceTraveled,
			expectedDistance,
		)
	}

	// Дополнительная проверка: Y координата не должна измениться
	yDistanceTraveled := abs(finalPos.Y - initialPos.Y)
	if yDistanceTraveled > tolerance {
		t.Errorf(
			"Y координата изменилась на %.2f пикселей, ожидалось 0 (животное двигалось только по X)",
			yDistanceTraveled,
		)
	}
}

// TestMovementWithEatingState проверяет что животные не двигаются во время поедания
func TestMovementWithEatingState(t *testing.T) {
	// Создаем мир и систему движения
	worldWidth := float32(50)
	worldHeight := float32(38)
	world := core.NewWorld(worldWidth, worldHeight, 12345)
	movementSystem := simulation.NewMovementSystem(worldWidth, worldHeight)

	// Создаем животное
	entity := simulation.CreateAnimal(world, core.TypeRabbit, 25.0, 19.0)

	// Получаем начальную позицию
	initialPos, _ := world.GetPosition(entity)

	// Устанавливаем скорость и состояние поедания
	world.SetVelocity(entity, core.Velocity{X: 2.0, Y: 1.0})
	world.AddEatingState(entity, core.EatingState{Target: simulation.GrassEatingTarget})

	// Симулируем движение
	deltaTime := float32(1.0 / 60.0)
	for i := 0; i < 10; i++ {
		movementSystem.Update(world, deltaTime)
	}

	// Получаем финальную позицию
	finalPos, _ := world.GetPosition(entity)

	// Проверяем что позиция не изменилась
	tolerance := float32(0.001)
	if abs(finalPos.X-initialPos.X) > tolerance || abs(finalPos.Y-initialPos.Y) > tolerance {
		t.Errorf(
			"Животное двигалось во время поедания: начальная позиция (%.2f, %.2f), финальная (%.2f, %.2f)",
			initialPos.X, initialPos.Y, finalPos.X, finalPos.Y,
		)
	}

	// Проверяем что скорость была сброшена в ноль
	vel, _ := world.GetVelocity(entity)
	if abs(vel.X) > tolerance || abs(vel.Y) > tolerance {
		t.Errorf(
			"Скорость не была сброшена в ноль во время поедания: (%.2f, %.2f)",
			vel.X, vel.Y,
		)
	}
}

// TestBoundaryConstraints проверяет что животные не выходят за границы мира
func TestBoundaryConstraints(t *testing.T) {
	// Создаем маленький мир для упрощения тестирования
	worldWidth := float32(10)
	worldHeight := float32(10)
	world := core.NewWorld(worldWidth, worldHeight, 12345)
	movementSystem := simulation.NewMovementSystem(worldWidth, worldHeight)

	// Создаем животное рядом с границей
	entity := simulation.CreateAnimal(world, core.TypeRabbit, 1.0, 1.0)

	// Устанавливаем скорость направленную к границе (влево и вверх)
	world.SetVelocity(entity, core.Velocity{X: -5.0, Y: -5.0})

	// Симулируем движение
	deltaTime := float32(1.0 / 60.0)
	for i := 0; i < 60; i++ {
		movementSystem.Update(world, deltaTime)
	}

	// Получаем финальную позицию
	finalPos, _ := world.GetPosition(entity)

	// Получаем радиус животного
	size, _ := world.GetSize(entity)
	radiusInTiles := constants.SizeRadiusToTiles(size.Radius)

	// Проверяем что животное не вышло за границы с учетом радиуса
	margin := float32(0.1) // Минимальный отступ из MovementSystem
	minX := margin + radiusInTiles
	minY := margin + radiusInTiles
	maxX := worldWidth - margin - radiusInTiles
	maxY := worldHeight - margin - radiusInTiles

	if finalPos.X < minX || finalPos.X > maxX || finalPos.Y < minY || finalPos.Y > maxY {
		t.Errorf(
			"Животное вышло за границы мира: позиция (%.2f, %.2f), допустимые границы X:[%.2f, %.2f], Y:[%.2f, %.2f]",
			finalPos.X, finalPos.Y, minX, maxX, minY, maxY,
		)
	}
}

// abs возвращает абсолютное значение float32
func abs(x float32) float32 {
	if x < 0 {
		return -x
	}
	return x
}
