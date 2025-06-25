package integration

import (
	"math"
	"testing"

	"github.com/aiseeq/savanna/internal/core"
	"github.com/aiseeq/savanna/internal/simulation"
	"github.com/aiseeq/savanna/tests/common"
)

// TestWolfAttackBehavior исследует поведение волка при атаке зайца
//
//nolint:revive // Интеграционный тест может быть длинным
func TestWolfAttackBehavior(t *testing.T) {
	t.Parallel()
	// Создаем мир и системы
	worldSizePixels := float32(320) // 10 * 32
	world := core.NewWorld(worldSizePixels, worldSizePixels, 12345)

	// ИСПРАВЛЕНИЕ: Используем централизованный системный менеджер для правильного порядка систем
	systemManager := common.CreateTestSystemManager(worldSizePixels)

	// Создаем волка и зайца рядом друг с другом
	rabbitX, rabbitY := float32(160), float32(160) // Центр мира
	wolfX, wolfY := float32(140), float32(160)     // Слева от зайца на расстоянии 20 единиц

	rabbit := simulation.CreateAnimal(world, core.TypeRabbit, rabbitX, rabbitY)
	wolf := simulation.CreateAnimal(world, core.TypeWolf, wolfX, wolfY)

	// Делаем волка голодным для охоты
	world.SetSatiation(wolf, core.Satiation{Value: 30.0}) // Меньше 60% - будет охотиться

	t.Logf("=== Исследование поведения волка при атаке ===")
	t.Logf("Начальные позиции: волк (%.1f, %.1f), заяц (%.1f, %.1f)", wolfX, wolfY, rabbitX, rabbitY)

	deltaTime := float32(1.0 / 60.0) // 60 FPS
	tickCount := 0
	maxTicks := 600 // 10 секунд

	// Логируем каждые 6 тиков (10 раз в секунду)
	logInterval := 6

	for tickCount < maxTicks {
		// Проверяем жив ли заяц
		if !world.IsAlive(rabbit) {
			t.Logf("Заяц умер на тике %d", tickCount)
			break
		}

		// Получаем позиции до обновления
		wolfPos, _ := world.GetPosition(wolf)
		rabbitPos, _ := world.GetPosition(rabbit)
		wolfVel, _ := world.GetVelocity(wolf)

		// Вычисляем расстояние
		dx := wolfPos.X - rabbitPos.X
		dy := wolfPos.Y - rabbitPos.Y
		distance := math.Sqrt(float64(dx*dx + dy*dy))

		// Логируем каждые несколько тиков
		if tickCount%logInterval == 0 {
			t.Logf("Тик %3d: волк (%.1f,%.1f) vel(%.1f,%.1f) | заяц (%.1f,%.1f) | дистанция %.1f",
				tickCount, wolfPos.X, wolfPos.Y, wolfVel.X, wolfVel.Y, rabbitPos.X, rabbitPos.Y, distance)
		}

		// DEBUG: Проверяем компоненты волка перед обновлением поведения
		if tickCount == 0 {
			wolfBehavior, hasBehavior := world.GetBehavior(wolf)
			wolfHunger, hasHunger := world.GetSatiation(wolf)
			t.Logf("DEBUG: Волк behavior=%+v (has=%t), hunger=%+v (has=%t)",
				wolfBehavior, hasBehavior, wolfHunger, hasHunger)
		}

		// Обновляем поведение волка через централизованный системный менеджер

		// Обновляем движение для всех
		world.Update(deltaTime)
		systemManager.Update(world, deltaTime)

		tickCount++
	}

	// Финальные позиции и проверки
	if !world.IsAlive(wolf) {
		t.Errorf("❌ Волк не должен умереть в этом тесте")
	} else {
		wolfPos, _ := world.GetPosition(wolf)
		t.Logf("Финальная позиция волка: (%.1f, %.1f)", wolfPos.X, wolfPos.Y)

		// Проверяем что волк сдвинулся в сторону зайца
		if wolfPos.X <= wolfX {
			t.Errorf("❌ Волк не сдвинулся вправо к зайцу: начальная %.1f, финальная %.1f", wolfX, wolfPos.X)
		}
	}

	if !world.IsAlive(rabbit) {
		t.Logf("ℹ️ Заяц был убит волком - это нормальное поведение")
	} else {
		rabbitPos, _ := world.GetPosition(rabbit)
		t.Logf("Финальная позиция зайца: (%.1f, %.1f)", rabbitPos.X, rabbitPos.Y)

		// Если заяц жив, он должен был убежать от волка (сдвинуться вправо)
		if rabbitPos.X <= rabbitX {
			t.Errorf("❌ Заяц не убежал от волка: начальная %.1f, финальная %.1f", rabbitX, rabbitPos.X)
		}
	}

	t.Logf("✅ Тест поведения волка завершен")
}

// TestWolfOvershooting проверяет перепрыгивание волка через зайца
func TestWolfOvershooting(t *testing.T) {
	t.Parallel()
	// Создаем простую симуляцию
	worldSizePixels := float32(320)
	world := core.NewWorld(worldSizePixels, worldSizePixels, 54321)

	// ИСПРАВЛЕНИЕ: Используем централизованный системный менеджер для правильного порядка систем
	systemManager := common.CreateTestSystemManager(worldSizePixels)

	// Зайца ставим неподвижно, волка близко
	rabbit := simulation.CreateAnimal(world, core.TypeRabbit, 160, 160)
	wolf := simulation.CreateAnimal(world, core.TypeWolf, 145, 160) // Расстояние 15 единиц

	// Зайца делаем неподвижным
	world.SetVelocity(rabbit, core.Velocity{X: 0, Y: 0})
	world.SetSpeed(rabbit, core.Speed{Base: 0, Current: 0})

	// Волка делаем голодным
	world.SetSatiation(wolf, core.Satiation{Value: 20.0})

	t.Logf("=== Тест перепрыгивания волка ===")

	deltaTime := float32(1.0 / 60.0)

	for i := 0; i < 120; i++ { // 2 секунды
		wolfPos, _ := world.GetPosition(wolf)
		rabbitPos, _ := world.GetPosition(rabbit)

		dx := wolfPos.X - rabbitPos.X
		dy := wolfPos.Y - rabbitPos.Y
		distance := math.Sqrt(float64(dx*dx + dy*dy))

		if i%12 == 0 { // Каждые 0.2 секунды
			t.Logf("Сек %.1f: волк (%.1f,%.1f) | заяц (%.1f,%.1f) | дистанция %.1f",
				float32(i)/60.0, wolfPos.X, wolfPos.Y, rabbitPos.X, rabbitPos.Y, distance)
		}

		// DEBUG: Проверяем компоненты волка (только для первых 3 тиков)
		if i < 3 {
			wolfBehavior, hasBehavior := world.GetBehavior(wolf)
			wolfHunger, hasHunger := world.GetSatiation(wolf)
			wolfVel, _ := world.GetVelocity(wolf)
			wolfSpeed, hasSpeed := world.GetSpeed(wolf)
			t.Logf("DEBUG Тик %d: Волк behavior=%+v (has=%t), hunger=%+v (has=%t)",
				i, wolfBehavior, hasBehavior, wolfHunger, hasHunger)
			t.Logf("DEBUG Тик %d: Волк velocity=(%.2f,%.2f), speed=%+v (has=%t)",
				i, wolfVel.X, wolfVel.Y, wolfSpeed, hasSpeed)

			// Попытка поиска зайца (в тайлах)
			foundRabbit, found := world.FindNearestByTypeInTiles(wolfPos.X, wolfPos.Y, 5.0, core.TypeRabbit)
			t.Logf("DEBUG Тик %d: FindNearestByType для волка: rabbit=%d, found=%t",
				i, foundRabbit, found)
		}

		// Проверяем не перепрыгнул ли волк
		if wolfPos.X > rabbitPos.X && i > 30 { // Если волк прошел за зайца
			t.Logf("ВНИМАНИЕ: Волк перепрыгнул зайца на тике %d!", i)
			t.Logf("  Волк: (%.1f, %.1f), Заяц: (%.1f, %.1f)", wolfPos.X, wolfPos.Y, rabbitPos.X, rabbitPos.Y)
			break
		}

		if !world.IsAlive(rabbit) {
			t.Logf("Заяц умер на тике %d", i)
			break
		}

		world.Update(deltaTime)
		systemManager.Update(world, deltaTime)
	}

	// Финальные проверки
	if world.IsAlive(wolf) {
		wolfPos, _ := world.GetPosition(wolf)
		t.Logf("Финальная позиция волка: (%.1f, %.1f)", wolfPos.X, wolfPos.Y)

		// Волк должен был приблизиться к зайцу
		if wolfPos.X <= 145 {
			t.Errorf("❌ Волк не приблизился к зайцу: начальная 145, финальная %.1f", wolfPos.X)
		}
	} else {
		t.Errorf("❌ Волк не должен умереть в этом тесте")
	}

	if !world.IsAlive(rabbit) {
		t.Logf("ℹ️ Заяц был убит - нормальное поведение для голодного волка")
	}

	t.Logf("✅ Тест перепрыгивания завершен")
}
