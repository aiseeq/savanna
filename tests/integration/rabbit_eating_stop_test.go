package integration

import (
	"testing"

	"github.com/aiseeq/savanna/config"
	"github.com/aiseeq/savanna/internal/core"
	"github.com/aiseeq/savanna/internal/generator"
	"github.com/aiseeq/savanna/internal/simulation"
)

// TestRabbitStopsWhenEating проверяет что заяц останавливается когда ест траву
//
//nolint:gocognit,revive,funlen // Комплексный тест поведения зайца при поедании
func TestRabbitStopsWhenEating(t *testing.T) {
	t.Parallel()

	// Создаём мир
	world := core.NewWorld(TestWorldSize, TestWorldSize, 12345)

	// Создаём terrain и vegetation как в других тестах
	cfg := config.LoadDefaultConfig()
	cfg.World.Size = int(TestWorldSize / 32)
	terrainGen := generator.NewTerrainGenerator(cfg)
	terrain := terrainGen.Generate()
	vegetationSystem := simulation.NewVegetationSystem(terrain)

	// ИСПРАВЛЕНИЕ: Создаём новые системы питания (после рефакторинга)
	grassSearchSystem := simulation.NewGrassSearchSystem(vegetationSystem) // Создаёт EatingState
	behaviorSystem := simulation.NewAnimalBehaviorSystem(vegetationSystem)
	movementSystem := simulation.NewMovementSystem(TestWorldSize, TestWorldSize)

	// Найдём место с травой
	var grassX, grassY float32 = 100, 100 // Попробуем разные места
	for x := float32(50); x < 400; x += 50 {
		for y := float32(50); y < 400; y += 50 {
			if vegetationSystem.GetGrassAt(x, y) > 10.0 {
				grassX, grassY = x, y
				t.Logf("Найдена трава в точке (%.0f, %.0f): %.1f единиц", x, y, vegetationSystem.GetGrassAt(x, y))
				break
			}
		}
		if vegetationSystem.GetGrassAt(grassX, grassY) > 10.0 {
			break
		}
	}

	// Создаём голодного зайца там где есть трава
	rabbit := simulation.CreateAnimal(world, core.TypeRabbit, grassX, grassY)

	// Делаем зайца очень голодным (80%) чтобы он обязательно искал еду
	world.SetHunger(rabbit, core.Hunger{Value: 80.0})

	// Принудительно устанавливаем скорость 0 (заяц стоит на месте)
	world.SetVelocity(rabbit, core.Velocity{X: 0, Y: 0})

	deltaTime := float32(1.0 / 60.0) // 60 FPS

	t.Logf("=== TDD Тест: Заяц должен остановиться при еде ===")
	t.Logf("Начальная позиция зайца: (%.0f, %.0f)", grassX, grassY)
	t.Logf("Начальный голод: 80%% (< 90%% порога)")
	t.Logf("Трава на месте: %.1f единиц", vegetationSystem.GetGrassAt(grassX, grassY))

	// Симулируем 120 тиков (2 секунды)
	for i := 0; i < 120; i++ {
		// ИСПРАВЛЕНИЕ: Обновляем в правильном порядке с новыми системами
		grassSearchSystem.Update(world, deltaTime) // Создаёт EatingState если заяц на траве
		behaviorSystem.Update(world, deltaTime)    // Проверяет EatingState и не устанавливает скорость
		movementSystem.Update(world, deltaTime)    // Сбрасывает скорость если есть EatingState

		// Получаем состояние зайца
		pos, _ := world.GetPosition(rabbit)
		vel, _ := world.GetVelocity(rabbit)
		hunger, _ := world.GetHunger(rabbit)
		isEating := world.HasComponent(rabbit, core.MaskEatingState)
		grassAmount := vegetationSystem.GetGrassAt(pos.X, pos.Y)

		// Логируем каждые 30 тиков (каждые 0.5 сек)
		if i%30 == 0 {
			t.Logf("%.1fс: поз(%.1f,%.1f) скор(%.2f,%.2f) голод %.1f%% ест=%v трава=%.1f",
				float32(i)*deltaTime, pos.X, pos.Y, vel.X, vel.Y, hunger.Value, isEating, grassAmount)
		}

		// КРИТИЧЕСКАЯ ПРОВЕРКА: Если заяц ест, он НЕ должен двигаться
		if isEating {
			speed := vel.X*vel.X + vel.Y*vel.Y
			if speed > 0.1 { // Допускаем очень маленькую погрешность
				t.Errorf("❌ ОШИБКА на тике %d: Заяц ест (EatingState=true) но двигается со скоростью %.2f", i, speed)
				t.Errorf("   Позиция: (%.2f, %.2f), скорость: (%.2f, %.2f)", pos.X, pos.Y, vel.X, vel.Y)
				return
			}
		}

		// Если заяц наелся до 100% - тест завершён успешно
		if hunger.Value >= 100.0 {
			t.Logf("✅ Заяц наелся до %.1f%% на тике %d", hunger.Value, i)
			if isEating {
				// Финальная проверка - заяц ест и стоит
				speed := vel.X*vel.X + vel.Y*vel.Y
				if speed > 0.1 {
					t.Errorf("❌ КРИТИЧЕСКАЯ ОШИБКА: Заяц ест но движется в конце теста!")
					return
				}
				t.Logf("✅ Заяц правильно стоит во время еды (скорость %.4f)", speed)
			}
			return
		}
	}

	// Если дошли сюда - заяц не наелся за 2 секунды
	hunger, _ := world.GetHunger(rabbit)
	isEating := world.HasComponent(rabbit, core.MaskEatingState)

	if !isEating {
		t.Errorf("❌ Заяц не ест через 2 секунды (голод %.1f%%), хотя трава есть", hunger.Value)
	} else {
		t.Logf("✅ Заяц ест, но не успел наесться за 2 секунды (голод %.1f%%)", hunger.Value)
	}
}
