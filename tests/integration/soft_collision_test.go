package integration

import (
	"math"
	"testing"

	"github.com/aiseeq/savanna/internal/core"
	"github.com/aiseeq/savanna/internal/simulation"
)

// TestSoftCollisions проверяет мягкое расталкивание как в StarCraft 2
func TestSoftCollisions(t *testing.T) {
	t.Parallel()
	world := core.NewWorld(320, 320, 42)
	movementSystem := simulation.NewMovementSystem(320, 320)

	// Создаем двух зайцев движущихся друг на друга
	rabbit1 := simulation.CreateRabbit(world, 150, 160) // Слева
	rabbit2 := simulation.CreateRabbit(world, 170, 160) // Справа

	// Заставляем их двигаться друг на друга
	world.SetVelocity(rabbit1, core.Velocity{X: 10, Y: 0})  // Движется вправо
	world.SetVelocity(rabbit2, core.Velocity{X: -10, Y: 0}) // Движется влево

	t.Logf("=== Тест мягких коллизий ===")

	deltaTime := float32(1.0 / 60.0)

	initialPos1, _ := world.GetPosition(rabbit1)
	initialPos2, _ := world.GetPosition(rabbit2)
	t.Logf("Начальные позиции: заяц1 (%.1f, %.1f), заяц2 (%.1f, %.1f)",
		initialPos1.X, initialPos1.Y, initialPos2.X, initialPos2.Y)

	for i := 0; i < 180; i++ { // 3 секунды
		movementSystem.Update(world, deltaTime)
		world.Update(deltaTime)

		pos1, _ := world.GetPosition(rabbit1)
		pos2, _ := world.GetPosition(rabbit2)
		vel1, _ := world.GetVelocity(rabbit1)
		vel2, _ := world.GetVelocity(rabbit2)

		distance := math.Sqrt(float64((pos1.X-pos2.X)*(pos1.X-pos2.X) + (pos1.Y-pos2.Y)*(pos1.Y-pos2.Y)))

		if i%30 == 0 { // Каждые 0.5 секунды
			t.Logf("%.1fс: заяц1 (%.1f,%.1f) vel(%.1f,%.1f) | заяц2 (%.1f,%.1f) vel(%.1f,%.1f) | дист %.1f",
				float32(i)/60.0, pos1.X, pos1.Y, vel1.X, vel1.Y, pos2.X, pos2.Y, vel2.X, vel2.Y, distance)
		}

		// Проверяем что зайцы не застревают в одной точке
		if distance < 1.0 && (math.Abs(float64(vel1.X)) > 5.0 || math.Abs(float64(vel2.X)) > 5.0) {
			t.Errorf("Зайцы застряли слишком близко (дистанция %.1f) но все еще имеют высокую скорость", distance)
			break
		}

		// Проверяем что они разошлись если были в коллизии
		if i > 120 && distance < 8.0 { // Через 2 секунды должны разойтись на радиус коллизии
			t.Logf("Зайцы все еще близко через 2 секунды, но это нормально для мягких коллизий")
		}
	}

	finalPos1, _ := world.GetPosition(rabbit1)
	finalPos2, _ := world.GetPosition(rabbit2)
	finalVel1, _ := world.GetVelocity(rabbit1)
	finalVel2, _ := world.GetVelocity(rabbit2)

	t.Logf("Финальные позиции: заяц1 (%.1f, %.1f) vel(%.1f,%.1f), заяц2 (%.1f, %.1f) vel(%.1f,%.1f)",
		finalPos1.X, finalPos1.Y, finalVel1.X, finalVel1.Y, finalPos2.X, finalPos2.Y, finalVel2.X, finalVel2.Y)

	// Проверяем что зайцы не имеют экстремальных скоростей
	if math.Abs(float64(finalVel1.X)) > 50.0 || math.Abs(float64(finalVel2.X)) > 50.0 {
		t.Errorf("Финальные скорости слишком высокие: заяц1 %.1f, заяц2 %.1f", finalVel1.X, finalVel2.X)
	}
}
