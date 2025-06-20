package integration

import (
	"testing"

	"github.com/aiseeq/savanna/config"
	"github.com/aiseeq/savanna/internal/core"
	"github.com/aiseeq/savanna/internal/generator"
	"github.com/aiseeq/savanna/internal/simulation"
)

// abs возвращает абсолютное значение float32
func abs(x float32) float32 {
	if x < 0 {
		return -x
	}
	return x
}

// TestNewSpeedLogic проверяет новую логику влияния сытости и здоровья на скорость
//
//nolint:revive // function-length: Комплексный интеграционный тест скорости
func TestNewSpeedLogic(t *testing.T) {
	t.Parallel()

	t.Logf("=== TDD: Проверка новой логики скорости ===")
	t.Logf("НОВАЯ ЛОГИКА:")
	t.Logf("1. Голодные (< 80%%) бегают с полной скоростью")
	t.Logf("2. Сытые (> 80%%) замедляются: скорость *= (1 + 0.8 - сытость/100)")
	t.Logf("3. Раненые: скорость *= (процент_здоровья / 100)")

	// Создаём мир
	cfg := config.LoadDefaultConfig()
	worldWidth := float32(cfg.World.Size * 32)
	worldHeight := float32(cfg.World.Size * 32)
	world := core.NewWorld(worldWidth, worldHeight, 12345)

	// Создаём системы
	terrainGen := generator.NewTerrainGenerator(cfg)
	terrain := terrainGen.Generate()
	_ = simulation.NewVegetationSystem(terrain)                    // используется в системах
	hungerSpeedSystem := simulation.NewHungerSpeedModifierSystem() // Система изменения скорости
	deltaTime := float32(1.0 / 60.0)                               // Стандартный deltaTime

	// Создаём зайца
	rabbit := simulation.CreateAnimal(world, core.TypeRabbit, 100, 100)

	// Проверяем базовую скорость
	speed, _ := world.GetSpeed(rabbit)
	baseSpeed := speed.Base
	t.Logf("Базовая скорость зайца: %.1f", baseSpeed)

	_ = float32(1.0 / 60.0) // deltaTime - используется в основном цикле

	// ТЕСТ 1: Голодное животное (50%) должно бегать с полной скоростью
	t.Logf("\n=== ТЕСТ 1: Голодное животное (50%%) ===")
	world.SetHunger(rabbit, core.Hunger{Value: 50.0})
	world.SetHealth(rabbit, core.Health{Current: 50, Max: 50}) // Полное здоровье

	hungerSpeedSystem.Update(world, deltaTime)

	speed, _ = world.GetSpeed(rabbit)
	expectedSpeed := baseSpeed * 1.0 // Полная скорость
	t.Logf("Голод: 50%%, Здоровье: 100%% => Скорость: %.1f (ожидалось: %.1f)", speed.Current, expectedSpeed)

	tolerance := float32(0.01) // Допуск для float32
	if abs(speed.Current-expectedSpeed) > tolerance {
		t.Errorf("❌ Неправильная скорость голодного: %.1f != %.1f", speed.Current, expectedSpeed)
	} else {
		t.Logf("✅ Голодные бегают с полной скоростью")
	}

	// ТЕСТ 2: Сытое животное (90%) должно замедляться
	t.Logf("\n=== ТЕСТ 2: Сытое животное (90%%) ===")
	world.SetHunger(rabbit, core.Hunger{Value: 90.0})
	world.SetHealth(rabbit, core.Health{Current: 50, Max: 50}) // Полное здоровье

	hungerSpeedSystem.Update(world, deltaTime)

	speed, _ = world.GetSpeed(rabbit)

	// скорость *= (1 + 0.8 - 90/100) = (1 + 0.8 - 0.9) = 0.9
	expectedSpeed = baseSpeed * 0.9
	t.Logf("Голод: 90%%, Здоровье: 100%% => Скорость: %.1f (ожидалось: %.1f)", speed.Current, expectedSpeed)

	if abs(speed.Current-expectedSpeed) > tolerance {
		t.Errorf("❌ Неправильная скорость сытого: %.1f != %.1f", speed.Current, expectedSpeed)
	} else {
		t.Logf("✅ Сытые замедляются правильно")
	}

	// ТЕСТ 3: Очень сытое животное (95%) ещё медленнее
	t.Logf("\n=== ТЕСТ 3: Очень сытое животное (95%%) ===")
	world.SetHunger(rabbit, core.Hunger{Value: 95.0})
	world.SetHealth(rabbit, core.Health{Current: 50, Max: 50}) // Полное здоровье

	hungerSpeedSystem.Update(world, deltaTime)

	speed, _ = world.GetSpeed(rabbit)
	// скорость *= (1 + 0.8 - 95/100) = (1 + 0.8 - 0.95) = 0.85
	expectedSpeed = baseSpeed * 0.85
	t.Logf("Голод: 95%%, Здоровье: 100%% => Скорость: %.1f (ожидалось: %.1f)", speed.Current, expectedSpeed)

	if abs(speed.Current-expectedSpeed) > tolerance {
		t.Errorf("❌ Неправильная скорость очень сытого: %.1f != %.1f", speed.Current, expectedSpeed)
	} else {
		t.Logf("✅ Очень сытые ещё медленнее")
	}

	// ТЕСТ 4: Раненое голодное животное (50% голод, 25% здоровье)
	t.Logf("\n=== ТЕСТ 4: Раненое голодное животное (50%% голод, 25%% здоровье) ===")
	world.SetHunger(rabbit, core.Hunger{Value: 50.0})
	world.SetHealth(rabbit, core.Health{Current: 25, Max: 100}) // 25% здоровья

	hungerSpeedSystem.Update(world, deltaTime)

	speed, _ = world.GetSpeed(rabbit)
	// Голод 50% < 80% => множитель сытости = 1.0 (полная скорость)
	// Здоровье 25% => множитель здоровья = 0.25
	// Итого: скорость *= 1.0 * 0.25 = 0.25
	expectedSpeed = baseSpeed * 0.25
	t.Logf("Голод: 50%%, Здоровье: 25%% => Скорость: %.1f (ожидалось: %.1f)", speed.Current, expectedSpeed)

	if abs(speed.Current-expectedSpeed) > tolerance {
		t.Errorf("❌ Неправильная скорость раненого голодного: %.1f != %.1f", speed.Current, expectedSpeed)
	} else {
		t.Logf("✅ Раненые голодные правильно замедляются")
	}

	// ТЕСТ 5: Раненое сытое животное (90% голод, 50% здоровье)
	t.Logf("\n=== ТЕСТ 5: Раненое сытое животное (90%% голод, 50%% здоровье) ===")
	world.SetHunger(rabbit, core.Hunger{Value: 90.0})
	world.SetHealth(rabbit, core.Health{Current: 50, Max: 100}) // 50% здоровья

	hungerSpeedSystem.Update(world, deltaTime)

	speed, _ = world.GetSpeed(rabbit)
	// Голод 90% > 80% => множитель сытости = (1 + 0.8 - 0.9) = 0.9
	// Здоровье 50% => множитель здоровья = 0.5
	// Итого: скорость *= 0.9 * 0.5 = 0.45
	expectedSpeed = baseSpeed * 0.45
	t.Logf("Голод: 90%%, Здоровье: 50%% => Скорость: %.1f (ожидалось: %.1f)", speed.Current, expectedSpeed)

	if abs(speed.Current-expectedSpeed) > tolerance {
		t.Errorf("❌ Неправильная скорость раненого сытого: %.1f != %.1f", speed.Current, expectedSpeed)
	} else {
		t.Logf("✅ Раненые сытые получают двойной штраф")
	}

	// ТЕСТ 6: Пограничный случай - ровно 80% голода
	t.Logf("\n=== ТЕСТ 6: Пограничный случай (80%% голод) ===")
	world.SetHunger(rabbit, core.Hunger{Value: 80.0})
	world.SetHealth(rabbit, core.Health{Current: 100, Max: 100}) // Полное здоровье

	hungerSpeedSystem.Update(world, deltaTime)

	speed, _ = world.GetSpeed(rabbit)
	expectedSpeed = baseSpeed * 1.0 // При 80% должна быть полная скорость
	t.Logf("Голод: 80%%, Здоровье: 100%% => Скорость: %.1f (ожидалось: %.1f)", speed.Current, expectedSpeed)

	if abs(speed.Current-expectedSpeed) > tolerance {
		t.Errorf("❌ Неправильная скорость на границе: %.1f != %.1f", speed.Current, expectedSpeed)
	} else {
		t.Logf("✅ На границе 80%% работает правильно")
	}

	t.Logf("\n✅ Все тесты новой логики скорости пройдены!")
}
