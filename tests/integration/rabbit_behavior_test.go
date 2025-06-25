package integration

import (
	"testing"

	"github.com/aiseeq/savanna/internal/core"
	"github.com/aiseeq/savanna/internal/gamestate"
	"github.com/aiseeq/savanna/internal/simulation"
)

// TestRabbitMovementStability тестирует стабильность движения зайцев
func TestRabbitMovementStability(t *testing.T) {
	// Создаем игровое состояние
	config := &gamestate.GameConfig{
		WorldWidth:    800,
		WorldHeight:   600,
		FixedTimeStep: 1.0 / 60.0,
		RandomSeed:    12345,
	}
	gs := gamestate.NewGameState(config)
	world := gs.GetWorld()

	// Создаем зайца в центре мира
	rabbit := simulation.CreateAnimal(world, core.TypeRabbit, 400, 300)

	t.Logf("=== Тест стабильности движения зайца ===")

	// Запоминаем начальную позицию
	initialPos, _ := world.GetPosition(rabbit)
	t.Logf("Начальная позиция зайца: (%.1f, %.1f)", initialPos.X, initialPos.Y)

	// Отслеживаем движение зайца
	var positions []core.Position
	var velocities []core.Velocity

	for frame := 0; frame < 60; frame++ { // 1 секунда
		gs.Update()

		pos, _ := world.GetPosition(rabbit)
		vel, _ := world.GetVelocity(rabbit)

		positions = append(positions, pos)
		velocities = append(velocities, vel)

		// Логгируем каждые 10 кадров
		if frame%10 == 0 {
			hunger, _ := world.GetSatiation(rabbit)
			behavior, _ := world.GetBehavior(rabbit)

			t.Logf("Кадр %d: Заяц pos=(%.1f,%.1f) vel=(%.1f,%.1f) hunger=%.1f type=%d",
				frame, pos.X, pos.Y, vel.X, vel.Y, hunger.Value, behavior.Type)
		}

		// КРИТИЧЕСКИЙ ТЕСТ: Заяц не должен выходить за границы мира
		if pos.X < 0 || pos.X > 800 || pos.Y < 0 || pos.Y > 600 {
			t.Errorf("ПРОБЛЕМА: Заяц вышел за границы мира! Позиция: (%.1f, %.1f)", pos.X, pos.Y)
		}
	}

	// Анализируем стабильность движения
	finalPos := positions[len(positions)-1]

	t.Logf("Финальная позиция зайца: (%.1f, %.1f)", finalPos.X, finalPos.Y)

	// Проверяем хаотичность движения
	chaotic := false
	for i := 1; i < len(velocities); i++ {
		prev := velocities[i-1]
		curr := velocities[i]

		// Если скорость резко изменилась (более чем в 2 раза), это может быть хаотичность
		if prev.X != 0 && curr.X != 0 {
			ratio := curr.X / prev.X
			if ratio < -0.5 || ratio > 2.0 {
				chaotic = true
				t.Logf("Хаотичное изменение скорости X на кадре %d: %.1f -> %.1f (ratio: %.2f)",
					i, prev.X, curr.X, ratio)
				break
			}
		}

		if prev.Y != 0 && curr.Y != 0 {
			ratio := curr.Y / prev.Y
			if ratio < -0.5 || ratio > 2.0 {
				chaotic = true
				t.Logf("Хаотичное изменение скорости Y на кадре %d: %.1f -> %.1f (ratio: %.2f)",
					i, prev.Y, curr.Y, ratio)
				break
			}
		}
	}

	if chaotic {
		t.Error("ПРОБЛЕМА: Обнаружено хаотичное движение зайца!")
	}

	// Проверяем общее смещение
	deltaX := finalPos.X - initialPos.X
	deltaY := finalPos.Y - initialPos.Y
	totalDistance := deltaX*deltaX + deltaY*deltaY

	t.Logf("Общее смещение зайца: (%.1f, %.1f), расстояние: %.1f", deltaX, deltaY, totalDistance)

	// Заяц должен двигаться, но не слишком далеко (не должен телепортироваться)
	if totalDistance > 10000 { // Больше 100 пикселей за секунду - подозрительно
		t.Errorf("ПРОБЛЕМА: Заяц переместился слишком далеко за 1 секунду: %.1f пикселей", totalDistance)
	}
}

// TestRabbitWorldBounds тестирует что зайцы не выходят за границы мира
func TestRabbitWorldBounds(t *testing.T) {
	config := &gamestate.GameConfig{
		WorldWidth:    400, // Маленький мир для быстрого тестирования границ
		WorldHeight:   300,
		FixedTimeStep: 1.0 / 60.0,
		RandomSeed:    12345,
	}
	gs := gamestate.NewGameState(config)
	world := gs.GetWorld()

	// Создаем зайцев возле границ мира
	rabbitNearLeft := simulation.CreateAnimal(world, core.TypeRabbit, 10, 150)
	rabbitNearRight := simulation.CreateAnimal(world, core.TypeRabbit, 390, 150)
	rabbitNearTop := simulation.CreateAnimal(world, core.TypeRabbit, 200, 10)
	rabbitNearBottom := simulation.CreateAnimal(world, core.TypeRabbit, 200, 290)

	rabbits := []core.EntityID{rabbitNearLeft, rabbitNearRight, rabbitNearTop, rabbitNearBottom}
	rabbitNames := []string{"Left", "Right", "Top", "Bottom"}

	t.Logf("=== Тест границ мира для зайцев ===")

	// Симулируем длительное время
	for frame := 0; frame < 300; frame++ { // 5 секунд
		gs.Update()

		// Проверяем каждого зайца каждые 60 кадров
		if frame%60 == 0 {
			for i, rabbit := range rabbits {
				pos, _ := world.GetPosition(rabbit)

				t.Logf("Кадр %d: Заяц %s pos=(%.1f,%.1f)", frame, rabbitNames[i], pos.X, pos.Y)

				// КРИТИЧЕСКИЙ ТЕСТ: Проверяем границы
				if pos.X < 0 || pos.X > config.WorldWidth || pos.Y < 0 || pos.Y > config.WorldHeight {
					t.Errorf("КРИТИЧЕСКАЯ ОШИБКА: Заяц %s вышел за границы мира! Позиция: (%.1f, %.1f), границы: (0,0)-(%.0f,%.0f)",
						rabbitNames[i], pos.X, pos.Y, config.WorldWidth, config.WorldHeight)
				}
			}
		}
	}
}
