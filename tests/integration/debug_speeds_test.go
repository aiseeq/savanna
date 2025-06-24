package integration

import (
	"testing"

	"github.com/aiseeq/savanna/internal/adapters"
	"github.com/aiseeq/savanna/internal/core"
	"github.com/aiseeq/savanna/internal/simulation"
)

// TestDebugSpeedsAndMovement проверяет реальные скорости животных
func TestDebugSpeedsAndMovement(t *testing.T) {
	t.Parallel()

	// Создаём простой мир
	world := core.NewWorld(320, 320, 12345)

	// Создаём зайца
	rabbit := simulation.CreateAnimal(world, core.TypeRabbit, 160, 160)

	// Получаем начальные характеристики
	rabbitSpeed, _ := world.GetSpeed(rabbit)
	rabbitBehavior, _ := world.GetBehavior(rabbit)

	t.Logf("=== ОТЛАДКА СКОРОСТЕЙ И ДВИЖЕНИЯ ===")
	t.Logf("Базовая скорость зайца: %.2f тайлов/сек", rabbitSpeed.Base)
	t.Logf("Радиус видения зайца: %.2f тайлов", rabbitBehavior.VisionRange)
	t.Logf("Множители поведения:")
	t.Logf("  - Поиск: %.2f", rabbitBehavior.SearchSpeed)
	t.Logf("  - Блуждание: %.2f", rabbitBehavior.WanderingSpeed)
	t.Logf("  - Спокойствие: %.2f", rabbitBehavior.ContentSpeed)

	// Симулируем движение (БЕЗ BehaviorSystem чтобы не переопределял скорость)
	systemManager := core.NewSystemManager()
	movementSystem := simulation.NewMovementSystem(320, 320)

	// ТОЛЬКО MovementSystem - не добавляем BehaviorSystem чтобы он не переопределял скорость
	systemManager.AddSystem(&adapters.MovementSystemAdapter{System: movementSystem})

	// Делаем зайца сытым чтобы он не искал еду (95% > 90% threshold)
	world.SetHunger(rabbit, core.Hunger{Value: 95.0})

	// Заставляем зайца двигаться
	world.SetVelocity(rabbit, core.Velocity{X: 2.0, Y: 0.0}) // 2 тайла/сек вправо

	deltaTime := float32(1.0 / 60.0) // 60 FPS

	// Начальная позиция
	initialPos, _ := world.GetPosition(rabbit)
	t.Logf("\nНачальная позиция: (%.1f, %.1f) пикселей", initialPos.X, initialPos.Y)

	// Симулируем 60 тиков (1 секунда)
	for i := 0; i < 60; i++ {
		systemManager.Update(world, deltaTime)
		world.Update(deltaTime)
	}

	// Финальная позиция после 1 секунды
	finalPos, _ := world.GetPosition(rabbit)
	t.Logf("Финальная позиция: (%.1f, %.1f) пикселей", finalPos.X, finalPos.Y)

	// Рассчитываем реальную скорость
	distance := finalPos.X - initialPos.X
	realSpeedPixelsPerSec := distance                    // За 1 секунду
	realSpeedTilesPerSec := realSpeedPixelsPerSec / 32.0 // Конвертируем в тайлы

	t.Logf("Реальная скорость: %.2f пикселей/сек = %.2f тайлов/сек", realSpeedPixelsPerSec, realSpeedTilesPerSec)

	// Проверяем что скорость соответствует установленной (БЕЗ BehaviorSystem)
	// Ожидаемая скорость: 2.0 тайлов/сек (установленная напрямую)
	// Допустимый диапазон: 1.8-2.2 тайлов/сек (учитывая дискретность симуляции)
	if realSpeedTilesPerSec < 1.8 || realSpeedTilesPerSec > 2.2 {
		t.Errorf("❌ Неожиданная скорость: %.2f тайлов/сек (ожидалось 1.8-2.2 без поведенческих множителей)", realSpeedTilesPerSec)
	} else {
		t.Logf("✅ Скорость корректна без BehaviorSystem: %.2f тайлов/сек", realSpeedTilesPerSec)
	}

	// Проверим что происходит с текущей скоростью
	currentVel, _ := world.GetVelocity(rabbit)
	t.Logf("Текущая скорость после симуляции: (%.2f, %.2f) тайлов/сек", currentVel.X, currentVel.Y)
}
