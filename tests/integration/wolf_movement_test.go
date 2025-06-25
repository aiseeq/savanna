package integration

import (
	"math"
	"testing"

	"github.com/aiseeq/savanna/internal/core"
	"github.com/aiseeq/savanna/internal/gamestate"
	"github.com/aiseeq/savanna/internal/simulation"
)

// TestWolfMovement тестирует что волки движутся при охоте
func TestWolfMovement(t *testing.T) {
	// Создаем игровое состояние
	config := &gamestate.GameConfig{
		WorldWidth:    800,
		WorldHeight:   600,
		FixedTimeStep: 1.0 / 60.0,
		RandomSeed:    12345,
	}
	gs := gamestate.NewGameState(config)
	world := gs.GetWorld()

	// Создаем волка и зайца в пределах видимости но достаточно далеко для движения
	wolf := simulation.CreateAnimal(world, core.TypeWolf, 100, 100)
	rabbit := simulation.CreateAnimal(world, core.TypeRabbit, 120, 100) // 20 пикселей = 0.6 тайла

	// Делаем волка очень голодным для активации охоты
	world.SetSatiation(wolf, core.Satiation{Value: 20.0}) // Очень голодный

	t.Logf("=== Тест движения волка ===")

	// Запоминаем начальную позицию
	initialPos, _ := world.GetPosition(wolf)
	t.Logf("Начальная позиция волка: (%.1f, %.1f)", initialPos.X, initialPos.Y)

	// Симулируем больше кадров для заметного движения
	for frame := 0; frame < 60; frame++ {
		gs.Update()

		// Логгируем каждые 10 кадров
		if frame%10 == 0 {
			pos, _ := world.GetPosition(wolf)
			vel, _ := world.GetVelocity(wolf)
			hunger, _ := world.GetSatiation(wolf)
			speed, _ := world.GetSpeed(wolf)
			behavior, _ := world.GetBehavior(wolf)

			// Проверяем есть ли заяц в поле зрения
			rabbitPos, _ := world.GetPosition(rabbit)
			distance := float32(math.Sqrt(float64((pos.X-rabbitPos.X)*(pos.X-rabbitPos.X) + (pos.Y-rabbitPos.Y)*(pos.Y-rabbitPos.Y))))

			t.Logf("Кадр %d: Волк pos=(%.1f,%.1f) vel=(%.3f,%.3f) hunger=%.1f speed=%.3f baseSpeed=%.3f",
				frame, pos.X, pos.Y, vel.X, vel.Y, hunger.Value, speed.Current, speed.Base)
			t.Logf("  Behavior: SearchSpeed=%.3f SatiationThreshold=%.1f VisionRange=%.1f",
				behavior.SearchSpeed, behavior.SatiationThreshold, behavior.VisionRange)
			visionRangePixels := behavior.VisionRange * 32.0 // Конвертируем тайлы в пиксели
			t.Logf("  Заяц pos=(%.1f,%.1f) distance=%.1f visionRange=%.1f пикс (в зрении: %v)",
				rabbitPos.X, rabbitPos.Y, distance, visionRangePixels, distance <= visionRangePixels)
		}
	}

	// Проверяем финальную позицию
	finalPos, _ := world.GetPosition(wolf)
	finalVel, _ := world.GetVelocity(wolf)

	t.Logf("Финальная позиция волка: (%.1f, %.1f)", finalPos.X, finalPos.Y)
	t.Logf("Финальная скорость волка: (%.1f, %.1f)", finalVel.X, finalVel.Y)

	// Основной тест: волк должен двигаться (даже медленно)
	deltaX := finalPos.X - initialPos.X
	deltaY := finalPos.Y - initialPos.Y
	moved := (deltaX*deltaX + deltaY*deltaY) > 1.0 // Переместился больше чем на 1 пиксель

	if !moved {
		t.Errorf("ПРОБЛЕМА: Волк не движется! Начальная позиция: (%.1f, %.1f), финальная: (%.1f, %.1f)",
			initialPos.X, initialPos.Y, finalPos.X, finalPos.Y)
	} else {
		t.Logf("✅ Волк движется! Переместился на (%.1f, %.1f)", deltaX, deltaY)
	}

	// Дополнительная проверка: волк должен иметь ненулевую скорость
	hasVelocity := (finalVel.X*finalVel.X + finalVel.Y*finalVel.Y) > 0.1
	if !hasVelocity {
		t.Error("ПРОБЛЕМА: Волк имеет нулевую скорость!")
	} else {
		t.Logf("✅ Волк имеет скорость!")
	}
}
