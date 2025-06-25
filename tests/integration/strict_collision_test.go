package integration

import (
	"math"
	"testing"

	"github.com/aiseeq/savanna/internal/adapters"
	"github.com/aiseeq/savanna/internal/core"
	"github.com/aiseeq/savanna/internal/simulation"
)

// TestStrictCollisions строго проверяет что животные НЕ ПЕРЕСЕКАЮТСЯ физически
func TestStrictCollisions(t *testing.T) {
	t.Parallel()
	world := core.NewWorld(320, 320, 42)
	movementSystem := simulation.NewMovementSystem(320, 320)

	systemManager := core.NewSystemManager()
	systemManager.AddSystem(&adapters.MovementSystemAdapter{System: movementSystem})

	// Создаем двух зайцев близко друг к другу
	rabbit1 := simulation.CreateAnimal(world, core.TypeRabbit, 150, 160)
	rabbit2 := simulation.CreateAnimal(world, core.TypeRabbit, 170, 160) // 20 пикселей между центрами

	// Получаем их радиусы
	size1, _ := world.GetSize(rabbit1)
	size2, _ := world.GetSize(rabbit2)

	// Радиус зайца в пикселях (от Size компонента)
	radius1 := size1.Radius
	radius2 := size2.Radius
	minAllowedDistance := radius1 + radius2 // Минимальное расстояние между центрами

	t.Logf("=== СТРОГИЙ ТЕСТ КОЛЛИЗИЙ ===")
	t.Logf("Радиус зайца 1: %.1f пикселей", radius1)
	t.Logf("Радиус зайца 2: %.1f пикселей", radius2)
	t.Logf("Минимальное расстояние между центрами: %.1f пикселей", minAllowedDistance)

	// Заставляем их двигаться друг на друга медленно
	world.SetVelocity(rabbit1, core.Velocity{X: 2, Y: 0})  // Медленно вправо
	world.SetVelocity(rabbit2, core.Velocity{X: -2, Y: 0}) // Медленно влево

	deltaTime := float32(1.0 / 60.0)
	violations := 0

	for i := 0; i < 300; i++ { // 5 секунд
		systemManager.Update(world, deltaTime)
		world.Update(deltaTime)

		pos1, _ := world.GetPosition(rabbit1)
		pos2, _ := world.GetPosition(rabbit2)
		vel1, _ := world.GetVelocity(rabbit1)
		vel2, _ := world.GetVelocity(rabbit2)

		distance := math.Sqrt(float64((pos1.X-pos2.X)*(pos1.X-pos2.X) + (pos1.Y-pos2.Y)*(pos1.Y-pos2.Y)))

		if i%60 == 0 { // Каждую секунду
			t.Logf("%.1fс: заяц1 (%.1f,%.1f) заяц2 (%.1f,%.1f) дист=%.1f мин=%.1f",
				float32(i)/60.0, pos1.X, pos1.Y, pos2.X, pos2.Y, distance, minAllowedDistance)
		}

		// КРИТИЧЕСКАЯ ПРОВЕРКА: зайцы НЕ ДОЛЖНЫ пересекаться
		if distance < float64(minAllowedDistance) {
			violations++
			t.Errorf("❌ ПЕРЕСЕЧЕНИЕ на тике %d: дистанция %.2f < мин %.2f",
				i, distance, minAllowedDistance)
			t.Errorf("   Позиции: заяц1 (%.1f,%.1f) заяц2 (%.1f,%.1f)",
				pos1.X, pos1.Y, pos2.X, pos2.Y)
			t.Errorf("   Скорости: заяц1 (%.1f,%.1f) заяц2 (%.1f,%.1f)",
				vel1.X, vel1.Y, vel2.X, vel2.Y)
		}

		// Проверяем что животные не вылетели за границы мира слишком далеко
		if pos1.X < 10 || pos1.X > 310 || pos2.X < 10 || pos2.X > 310 {
			t.Errorf("❌ ЖИВОТНОЕ ВЫЛЕТЕЛО ЗА ГРАНИЦЫ на тике %d", i)
			t.Errorf("   Позиции: заяц1 (%.1f,%.1f) заяц2 (%.1f,%.1f)",
				pos1.X, pos1.Y, pos2.X, pos2.Y)
			break
		}
	}

	// Финальная проверка
	finalPos1, _ := world.GetPosition(rabbit1)
	finalPos2, _ := world.GetPosition(rabbit2)
	finalDistance := math.Sqrt(float64((finalPos1.X-finalPos2.X)*(finalPos1.X-finalPos2.X) + (finalPos1.Y-finalPos2.Y)*(finalPos1.Y-finalPos2.Y)))

	t.Logf("\n=== РЕЗУЛЬТАТЫ ===")
	t.Logf("Финальная дистанция: %.1f (мин %.1f)", finalDistance, minAllowedDistance)
	t.Logf("Нарушений пересечения: %d", violations)

	if violations > 0 {
		t.Errorf("❌ СИСТЕМА КОЛЛИЗИЙ НЕ РАБОТАЕТ: %d нарушений пересечения", violations)
	} else {
		t.Logf("✅ Система коллизий работает корректно")
	}
}

// TestAnimalOverlapDetection проверяет обнаружение пересечений в статическом случае
func TestAnimalOverlapDetection(t *testing.T) {
	t.Parallel()
	world := core.NewWorld(320, 320, 42)

	// Создаем зайцев которые должны пересекаться
	rabbit1 := simulation.CreateAnimal(world, core.TypeRabbit, 150, 160)
	rabbit2 := simulation.CreateAnimal(world, core.TypeRabbit, 158, 160) // Очень близко - 8 пикселей

	size1, _ := world.GetSize(rabbit1)
	size2, _ := world.GetSize(rabbit2)
	radius1 := size1.Radius
	radius2 := size2.Radius
	minDistance := radius1 + radius2

	pos1, _ := world.GetPosition(rabbit1)
	pos2, _ := world.GetPosition(rabbit2)
	distance := math.Sqrt(float64((pos1.X-pos2.X)*(pos1.X-pos2.X) + (pos1.Y-pos2.Y)*(pos1.Y-pos2.Y)))

	t.Logf("=== ТЕСТ ОБНАРУЖЕНИЯ ПЕРЕСЕЧЕНИЙ ===")
	t.Logf("Радиусы: заяц1=%.1f заяц2=%.1f", radius1, radius2)
	t.Logf("Позиции: заяц1=(%.1f,%.1f) заяц2=(%.1f,%.1f)", pos1.X, pos1.Y, pos2.X, pos2.Y)
	t.Logf("Дистанция: %.1f, минимум: %.1f", distance, minDistance)

	if distance < float64(minDistance) {
		t.Logf("✅ Пересечение корректно обнаружено: %.1f < %.1f", distance, minDistance)
	} else {
		t.Errorf("❌ Пересечение НЕ обнаружено: %.1f >= %.1f", distance, minDistance)
	}

	// Теперь запускаем систему движения один раз и проверяем что пересечение исправлено
	movementSystem := simulation.NewMovementSystem(320, 320)
	systemManager := core.NewSystemManager()
	systemManager.AddSystem(&adapters.MovementSystemAdapter{System: movementSystem})

	// Останавливаем зайцев чтобы только коллизии работали
	world.SetVelocity(rabbit1, core.Velocity{X: 0, Y: 0})
	world.SetVelocity(rabbit2, core.Velocity{X: 0, Y: 0})

	// Один тик системы движения
	systemManager.Update(world, float32(1.0/60.0))

	// Проверяем результат
	newPos1, _ := world.GetPosition(rabbit1)
	newPos2, _ := world.GetPosition(rabbit2)
	newDistance := math.Sqrt(float64((newPos1.X-newPos2.X)*(newPos1.X-newPos2.X) + (newPos1.Y-newPos2.Y)*(newPos1.Y-newPos2.Y)))

	t.Logf("\nПосле коррекции:")
	t.Logf("Новые позиции: заяц1=(%.1f,%.1f) заяц2=(%.1f,%.1f)", newPos1.X, newPos1.Y, newPos2.X, newPos2.Y)
	t.Logf("Новая дистанция: %.1f", newDistance)

	if newDistance >= float64(minDistance) {
		t.Logf("✅ Пересечение исправлено: %.1f >= %.1f", newDistance, minDistance)
	} else {
		t.Errorf("❌ Пересечение НЕ ИСПРАВЛЕНО: %.1f < %.1f", newDistance, minDistance)
	}
}
