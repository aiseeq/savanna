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
	worldWidth := float32(50)  // 50 тайлов
	worldHeight := float32(38) // 38 тайлов
	world := core.NewWorld(worldWidth, worldHeight, 12345)
	movementSystem := simulation.NewMovementSystem(worldWidth, worldHeight)

	// Создаем простую сущность в центре мира (позиция в пикселях)
	centerX := constants.TilesToPixels(worldWidth / 2)  // 25 тайлов * 32 = 800 пикселей
	centerY := constants.TilesToPixels(worldHeight / 2) // 19 тайлов * 32 = 608 пикселей
	entity := world.CreateEntity()
	// ЭЛЕГАНТНАЯ МАТЕМАТИКА: создание позиции и скорости
	world.AddPosition(entity, core.NewPosition(centerX, centerY))
	world.AddVelocity(entity, core.NewVelocity(1.0, 0.0)) // 1.0 тайл/сек по X

	// Добавляем Size компонент с разумным радиусом (1 тайл)
	world.AddSize(entity, core.Size{Radius: 1.0, AttackRange: 0.0})

	// Получаем начальную позицию
	initialPos, _ := world.GetPosition(entity)

	// Симулируем движение на 1 секунду (60 тиков по 1/60 сек)
	deltaTime := float32(1.0 / 60.0)
	for i := 0; i < 60; i++ {
		movementSystem.Update(world, deltaTime)
	}

	// Получаем финальную позицию
	finalPos, _ := world.GetPosition(entity)

	// ЭЛЕГАНТНАЯ МАТЕМАТИКА: вычисляем расстояние
	distancePixels := finalPos.X - initialPos.X
	distanceTraveled := distancePixels / 32.0 // Конвертируем пиксели в тайлы

	// Ожидаемое расстояние: 1 тайл
	expectedDistance := float32(1.0) // 1 тайл

	// Проверяем с небольшой погрешностью (0.01 тайла)
	tolerance := float32(0.01)
	if abs(distanceTraveled-expectedDistance) > tolerance {
		t.Errorf(
			"Неверная скорость движения: сущность со скоростью 1.0 тайл/сек прошла %.2f тайлов за секунду, ожидалось %.2f",
			distanceTraveled,
			expectedDistance,
		)
	}

	// Дополнительная проверка: Y координата не должна измениться
	yDistanceTraveled := abs(finalPos.Y - initialPos.Y)
	if yDistanceTraveled > tolerance {
		t.Errorf(
			"Y координата изменилась на %.2f пикселей, ожидалось 0 (сущность двигалась только по X)",
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

	// ЭЛЕГАНТНАЯ МАТЕМАТИКА: устанавливаем скорость и состояние поедания
	world.SetVelocity(entity, core.NewVelocity(2.0, 1.0))
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

	// ЭЛЕГАНТНАЯ МАТЕМАТИКА: устанавливаем скорость направленную к границе
	world.SetVelocity(entity, core.NewVelocity(-5.0, -5.0))

	// Симулируем движение
	deltaTime := float32(1.0 / 60.0)
	for i := 0; i < 60; i++ {
		movementSystem.Update(world, deltaTime)
	}

	// Получаем финальную позицию
	finalPos, _ := world.GetPosition(entity)

	// Получаем радиус животного
	size, _ := world.GetSize(entity)
	radiusInTiles := size.Radius

	// ИСПРАВЛЕНИЕ: Конвертируем размеры мира в пиксели для корректной проверки
	worldWidthPixels := constants.TilesToPixels(worldWidth)
	worldHeightPixels := constants.TilesToPixels(worldHeight)
	radiusPixels := constants.TilesToPixels(radiusInTiles)
	marginPixels := float32(3.2) // 0.1 тайла * 32 пикселя

	minX := marginPixels + radiusPixels
	minY := marginPixels + radiusPixels
	maxX := worldWidthPixels - marginPixels - radiusPixels
	maxY := worldHeightPixels - marginPixels - radiusPixels

	// Проверяем что животное не вышло за границы
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
