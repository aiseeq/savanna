package integration

import (
	"math"
	"testing"

	"github.com/aiseeq/savanna/internal/adapters"
	"github.com/aiseeq/savanna/internal/constants"
	"github.com/aiseeq/savanna/internal/core"
	"github.com/aiseeq/savanna/internal/physics"
	"github.com/aiseeq/savanna/internal/simulation"
)

// TestDebugCollisionDetection отладка системы коллизий шаг за шагом
func TestDebugCollisionDetection(t *testing.T) {
	t.Parallel()
	world := core.NewWorld(320, 320, 42)
	movementSystem := simulation.NewMovementSystem(320, 320)

	// Создаем зайцев которые пересекаются
	rabbit1 := simulation.CreateAnimal(world, core.TypeRabbit, 150, 160)
	rabbit2 := simulation.CreateAnimal(world, core.TypeRabbit, 158, 160) // 8 пикселей между центрами

	t.Logf("=== ОТЛАДКА СИСТЕМЫ КОЛЛИЗИЙ ===")

	// Проверяем компоненты
	pos1, _ := world.GetPosition(rabbit1)
	pos2, _ := world.GetPosition(rabbit2)
	size1, _ := world.GetSize(rabbit1)
	size2, _ := world.GetSize(rabbit2)

	t.Logf("Позиции: заяц1=(%.1f,%.1f) заяц2=(%.1f,%.1f)", pos1.X, pos1.Y, pos2.X, pos2.Y)
	t.Logf("Радиусы: заяц1=%.1f заяц2=%.1f", size1.Radius, size2.Radius)

	// Тестируем поиск кандидатов
	searchRadiusPixels := size1.Radius * 1.1 // SearchRadiusMultiplier
	// Конвертируем радиус поиска в тайлы
	searchRadiusTiles := constants.SizeRadiusToTiles(searchRadiusPixels)
	posInTiles := physics.Vec2{
		X: constants.PixelsToTiles(pos1.X),
		Y: constants.PixelsToTiles(pos1.Y),
	}

	t.Logf("Поиск коллизий:")
	t.Logf("  Позиция в тайлах: (%.2f, %.2f)", posInTiles.X, posInTiles.Y)
	t.Logf("  Радиус поиска в пикселях: %.2f", searchRadiusPixels)
	t.Logf("  Радиус поиска в тайлах: %.2f", searchRadiusTiles)

	// Проверяем QueryInRadius
	candidates := world.QueryInRadius(posInTiles.X, posInTiles.Y, searchRadiusTiles)
	t.Logf("  Найдено кандидатов: %d", len(candidates))
	for i, candidate := range candidates {
		t.Logf("    Кандидат %d: entity %d", i, candidate)
	}

	// Должен найти как минимум себя и соседа
	if len(candidates) < 2 {
		t.Errorf("❌ QueryInRadius не находит соседей!")
		t.Errorf("   Ожидалось: минимум 2 (заяц1=%d, заяц2=%d)", rabbit1, rabbit2)
		t.Errorf("   Получено: %d кандидатов", len(candidates))
	}

	// Проверяем физику коллизий
	circle1 := physics.Circle{
		Center: physics.Vec2{
			X: constants.PixelsToTiles(pos1.X),
			Y: constants.PixelsToTiles(pos1.Y),
		},
		Radius: constants.SizeRadiusToTiles(size1.Radius),
	}
	circle2 := physics.Circle{
		Center: physics.Vec2{
			X: constants.PixelsToTiles(pos2.X),
			Y: constants.PixelsToTiles(pos2.Y),
		},
		Radius: constants.SizeRadiusToTiles(size2.Radius),
	}

	t.Logf("Круги коллизий в тайлах:")
	t.Logf("  Круг1: центр=(%.2f,%.2f) радиус=%.2f", circle1.Center.X, circle1.Center.Y, circle1.Radius)
	t.Logf("  Круг2: центр=(%.2f,%.2f) радиус=%.2f", circle2.Center.X, circle2.Center.Y, circle2.Radius)

	collision := physics.CircleCircleCollisionWithDetails(circle1, circle2)
	t.Logf("Результат проверки коллизии:")
	t.Logf("  Пересекаются: %v", collision.Colliding)
	t.Logf("  Проникновение: %.2f тайла", collision.Penetration)
	t.Logf("  Нормаль: (%.2f, %.2f)", collision.Normal.X, collision.Normal.Y)

	if !collision.Colliding {
		t.Errorf("❌ Физика коллизий не обнаруживает пересечение!")
		distance := math.Sqrt(float64((circle1.Center.X-circle2.Center.X)*(circle1.Center.X-circle2.Center.X) +
			(circle1.Center.Y-circle2.Center.Y)*(circle1.Center.Y-circle2.Center.Y)))
		minDistance := circle1.Radius + circle2.Radius
		t.Errorf("   Дистанция между центрами: %.2f тайла", distance)
		t.Errorf("   Минимальная дистанция: %.2f тайла", minDistance)
	}

	// Проверяем всю систему движения
	systemManager := core.NewSystemManager()
	systemManager.AddSystem(&adapters.MovementSystemAdapter{System: movementSystem})

	// Останавливаем зайцев
	world.SetVelocity(rabbit1, core.Velocity{X: 0, Y: 0})
	world.SetVelocity(rabbit2, core.Velocity{X: 0, Y: 0})

	t.Logf("\nЗапускаем систему движения...")

	// Один тик
	systemManager.Update(world, float32(1.0/60.0))

	// Проверяем результат
	newPos1, _ := world.GetPosition(rabbit1)
	newPos2, _ := world.GetPosition(rabbit2)

	t.Logf("Позиции после системы движения:")
	t.Logf("  Заяц1: (%.1f,%.1f) -> (%.1f,%.1f)", pos1.X, pos1.Y, newPos1.X, newPos1.Y)
	t.Logf("  Заяц2: (%.1f,%.1f) -> (%.1f,%.1f)", pos2.X, pos2.Y, newPos2.X, newPos2.Y)

	moved1 := math.Abs(float64(newPos1.X-pos1.X)) > 0.1 || math.Abs(float64(newPos1.Y-pos1.Y)) > 0.1
	moved2 := math.Abs(float64(newPos2.X-pos2.X)) > 0.1 || math.Abs(float64(newPos2.Y-pos2.Y)) > 0.1

	if !moved1 && !moved2 {
		t.Errorf("❌ Животные не двигаются после системы коллизий!")
	} else {
		t.Logf("✅ Система коллизий работает - животные сдвинулись")
	}
}
