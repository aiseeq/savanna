package integration

import (
	"testing"

	"github.com/aiseeq/savanna/config"
	"github.com/aiseeq/savanna/internal/adapters"
	"github.com/aiseeq/savanna/internal/animation"
	"github.com/aiseeq/savanna/internal/core"
	"github.com/aiseeq/savanna/internal/generator"
	"github.com/aiseeq/savanna/internal/simulation"
)

// TestEatingFrameValidation проверяет что зайцы получают сытость ТОЛЬКО на кадре 1
//
//nolint:revive // function-length: Критический тест валидации кадров анимации
func TestEatingFrameValidation(t *testing.T) {
	t.Parallel()

	t.Logf("=== TDD: Проверка питания ТОЛЬКО на кадре 1 ===")
	t.Logf("ЦЕЛЬ: Убедиться что зайцы получают сытость только при переходе на кадр 1 анимации поедания")

	// Создаём мир как в реальной игре
	cfg := config.LoadDefaultConfig()
	worldWidth := float32(cfg.World.Size * 32)
	worldHeight := float32(cfg.World.Size * 32)
	world := core.NewWorld(worldWidth, worldHeight, 12345)

	// Создаём terrain
	terrainGen := generator.NewTerrainGenerator(cfg)
	terrain := terrainGen.Generate()

	// Все системы как в реальной игре
	systemManager := core.NewSystemManager()
	vegetationSystem := simulation.NewVegetationSystem(terrain)
	feedingSystem := simulation.NewFeedingSystem(vegetationSystem)
	grassEatingSystem := simulation.NewGrassEatingSystem(vegetationSystem)

	systemManager.AddSystem(vegetationSystem)
	systemManager.AddSystem(&adapters.FeedingSystemAdapter{System: feedingSystem})
	systemManager.AddSystem(grassEatingSystem)

	// Создаём анимационную систему
	_ = animation.NewAnimationSystem()   // Для полноты игровой среды
	_ = animation.NewAnimationResolver() // Для полноты игровой среды

	// Создаём зайца и траву
	rabbit := simulation.CreateAnimal(world, core.TypeRabbit, 200, 200)
	tileX := int(200 / 32)
	tileY := int(200 / 32)
	terrain.SetGrassAmount(tileX, tileY, 100.0)

	// Делаем зайца голодным
	initialHunger := float32(60.0)
	world.SetHunger(rabbit, core.Hunger{Value: initialHunger})
	world.SetVelocity(rabbit, core.Velocity{X: 0, Y: 0})

	// Принудительно создаём EatingState
	world.AddEatingState(rabbit, core.EatingState{
		Target:          0,
		EatingProgress:  0.0,
		NutritionGained: 0.0,
	})

	deltaTime := float32(1.0 / 60.0)

	t.Logf("Начальное состояние:")
	t.Logf("  Голод зайца: %.1f%%", initialHunger)
	t.Logf("  Трава: %.1f единиц", terrain.GetGrassAmount(tileX, tileY))

	// ТЕСТ 1: Кадр 0 - НЕ должно быть питания
	t.Logf("\n=== ТЕСТ 1: Кадр 0 (НЕ должно быть питания) ===")

	// Устанавливаем анимацию поедания кадр 0
	world.SetAnimation(rabbit, core.Animation{
		CurrentAnim: int(animation.AnimEat),
		Frame:       0,
		Timer:       0,
		Playing:     true,
		FacingRight: true,
	})

	hunger1, _ := world.GetHunger(rabbit)
	t.Logf("Голод ДО обновления (кадр 0): %.3f%%", hunger1.Value)

	// Обновляем системы
	world.Update(deltaTime)
	systemManager.Update(world, deltaTime)

	hunger2, _ := world.GetHunger(rabbit)
	t.Logf("Голод ПОСЛЕ обновления (кадр 0): %.3f%%", hunger2.Value)

	hungerChange1 := hunger2.Value - hunger1.Value
	if hungerChange1 > 0 {
		t.Errorf("❌ ОШИБКА: Питание на кадре 0! Изменение: +%.3f%%", hungerChange1)
		t.Errorf("   Зайцы должны получать питание ТОЛЬКО на кадре 1")
	} else {
		t.Logf("✅ Кадр 0: питание НЕ получено (правильно)")
	}

	// ТЕСТ 2: Кадр 1 - ДОЛЖНО быть питание
	t.Logf("\n=== ТЕСТ 2: Кадр 1 (ДОЛЖНО быть питание) ===")

	// Переключаем на кадр 1
	world.SetAnimation(rabbit, core.Animation{
		CurrentAnim: int(animation.AnimEat),
		Frame:       1,
		Timer:       0,
		Playing:     true,
		FacingRight: true,
	})

	hunger3, _ := world.GetHunger(rabbit)
	t.Logf("Голод ДО обновления (кадр 1): %.3f%%", hunger3.Value)

	// Обновляем системы
	world.Update(deltaTime)
	systemManager.Update(world, deltaTime)

	hunger4, _ := world.GetHunger(rabbit)
	t.Logf("Голод ПОСЛЕ обновления (кадр 1): %.3f%%", hunger4.Value)

	hungerChange2 := hunger4.Value - hunger3.Value
	if hungerChange2 <= 0 {
		t.Errorf("❌ БАГ: Питание НЕ получено на кадре 1!")
		t.Errorf("   Ожидалось: +4%% питания")
		t.Errorf("   Получено: %.3f%% изменения", hungerChange2)
		t.Errorf("   ПРОБЛЕМА: GrassEatingSystem не работает на кадре 1")
	} else {
		t.Logf("✅ Кадр 1: питание получено (+%.3f%%)", hungerChange2)
	}

	// ТЕСТ 3: Возврат к кадру 0 - НЕ должно быть питания
	t.Logf("\n=== ТЕСТ 3: Возврат к кадру 0 (НЕ должно быть питания) ===")

	// Переключаем обратно на кадр 0
	world.SetAnimation(rabbit, core.Animation{
		CurrentAnim: int(animation.AnimEat),
		Frame:       0,
		Timer:       0,
		Playing:     true,
		FacingRight: true,
	})

	hunger5, _ := world.GetHunger(rabbit)
	t.Logf("Голод ДО обновления (кадр 0 снова): %.3f%%", hunger5.Value)

	// Обновляем системы
	world.Update(deltaTime)
	systemManager.Update(world, deltaTime)

	hunger6, _ := world.GetHunger(rabbit)
	t.Logf("Голод ПОСЛЕ обновления (кадр 0 снова): %.3f%%", hunger6.Value)

	hungerChange3 := hunger6.Value - hunger5.Value
	if hungerChange3 > 0 {
		t.Errorf("❌ БАГ: Питание при возврате к кадру 0!")
		t.Errorf("   Изменение: +%.3f%%", hungerChange3)
		t.Errorf("   ПРОБЛЕМА: Старая логика 'любая смена кадра' всё ещё работает")
	} else {
		t.Logf("✅ Возврат к кадру 0: питание НЕ получено (правильно)")
	}

	// ИТОГОВАЯ ПРОВЕРКА
	t.Logf("\n=== ИТОГОВАЯ ПРОВЕРКА ===")
	t.Logf("Питание на кадре 0: %.3f%%", hungerChange1)
	t.Logf("Питание на кадре 1: %.3f%% (ожидается > 0)", hungerChange2)
	t.Logf("Питание при возврате к кадру 0: %.3f%%", hungerChange3)

	if hungerChange2 <= 0 {
		t.Errorf("❌ ТЕСТ НЕ ПРОЙДЕН: Питание на кадре 1 не работает")
	} else if hungerChange1 > 0 || hungerChange3 > 0 {
		t.Errorf("❌ ТЕСТ НЕ ПРОЙДЕН: Питание происходит на неправильных кадрах")
	} else {
		t.Logf("✅ ВСЕ ТЕСТЫ ПРОЙДЕНЫ: Питание только на кадре 1")
	}
}
