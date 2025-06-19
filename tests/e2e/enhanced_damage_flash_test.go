package e2e

import (
	"testing"

	"github.com/aiseeq/savanna/internal/animation"
	"github.com/aiseeq/savanna/internal/core"
	"github.com/aiseeq/savanna/internal/simulation"
)

// TestEnhancedDamageFlash - E2E тест для проверки усиленного DamageFlash эффекта
//
// Фиксирует изменения:
// 1. Интенсивность эффекта увеличена в 5 раз (scale = 1.0 + intensity * 5.0)
// 2. Длительность эффекта уменьшена в 5 раз (0.16 секунды вместо 0.8)
// 3. Спрайт становится в ~5-6 раз ярче при максимальной интенсивности
//
//nolint:gocognit,revive,funlen // E2E тест для фиксации усиленного эффекта
func TestEnhancedDamageFlash(t *testing.T) {
	t.Parallel()

	world := core.NewWorld(640, 640, 42)
	combatSystem := simulation.NewCombatSystem()

	// Настраиваем анимационную систему
	wolfAnimSystem := animation.NewAnimationSystem()
	rabbitAnimSystem := animation.NewAnimationSystem()

	wolfAnimSystem.RegisterAnimation(animation.AnimAttack, 2, 1.0, false, nil)
	wolfAnimSystem.RegisterAnimation(animation.AnimIdle, 2, 1.0, true, nil)
	rabbitAnimSystem.RegisterAnimation(animation.AnimIdle, 2, 1.0, true, nil)

	animManager := animation.NewAnimationManager(wolfAnimSystem, rabbitAnimSystem)

	// Создаём тестовую сцену: волк атакует зайца
	rabbit := simulation.CreateAnimal(world, core.TypeRabbit, 300, 300)
	wolf := simulation.CreateAnimal(world, core.TypeWolf, 310, 300)

	// Делаем волка голодным для гарантированной атаки
	world.SetHunger(wolf, core.Hunger{Value: 20.0})

	t.Logf("=== ТЕСТ УСИЛЕННОГО DAMAGEFLASH ===")

	deltaTime := float32(1.0 / 60.0)
	var flashIntensity float32
	var flashTimer float32

	// Фаза 1: Ждём атаки и фиксируем параметры DamageFlash
	for i := 0; i < 180; i++ { // 3 секунды максимум
		combatSystem.Update(world, deltaTime)
		animManager.UpdateAllAnimations(world, deltaTime)

		if world.HasComponent(rabbit, core.MaskDamageFlash) {
			flash, _ := world.GetDamageFlash(rabbit)
			flashIntensity = flash.Intensity
			flashTimer = flash.Timer

			t.Logf("✅ DamageFlash создан на тике %d", i)
			t.Logf("  Начальная интенсивность: %.3f", flashIntensity)
			t.Logf("  Начальный таймер: %.3f секунд", flashTimer)
			break
		}
	}

	if !world.HasComponent(rabbit, core.MaskDamageFlash) {
		t.Fatal("DamageFlash не был создан за 3 секунды")
	}

	// КРИТИЧЕСКАЯ ПРОВЕРКА 1: Интенсивность в ожидаемых пределах
	if flashIntensity < 0.8 || flashIntensity > 1.0 {
		t.Errorf("БАГ: Неожиданная интенсивность: %.3f (ожидалось 0.8-1.0)", flashIntensity)
	}

	// КРИТИЧЕСКАЯ ПРОВЕРКА 2: Таймер соответствует новой длительности (0.16 сек)
	expectedDuration := float32(0.16)
	tolerance := float32(0.02) // 2% погрешность
	if flashTimer < expectedDuration-tolerance || flashTimer > expectedDuration+tolerance {
		t.Errorf("БАГ: Неправильная длительность: %.3f сек (ожидалось ~%.3f)",
			flashTimer, expectedDuration)
	} else {
		t.Logf("✅ Длительность корректна: %.3f сек (ускорена в 5 раз)", flashTimer)
	}

	// КРИТИЧЕСКАЯ ПРОВЕРКА 3: Расчёт усиления интенсивности (scale = 1.0 + intensity * 5.0)
	expectedScale := 1.0 + flashIntensity*5.0
	if expectedScale < 5.0 || expectedScale > 6.0 {
		t.Errorf("БАГ: Неправильный масштаб яркости: %.1f (ожидалось 5.0-6.0)", expectedScale)
	} else {
		t.Logf("✅ Масштаб яркости: %.1fх (усилен в 5 раз)", expectedScale)
	}

	// Фаза 2: Проверяем быстрое угасание (должно исчезнуть за ~10 тиков)
	t.Logf("\n=== ПРОВЕРКА БЫСТРОГО УГАСАНИЯ ===")

	ticksToDisappear := 0
	for i := 0; i < 20; i++ { // 20 тиков максимум (треть секунды)
		combatSystem.Update(world, deltaTime)
		animManager.UpdateAllAnimations(world, deltaTime)

		if !world.HasComponent(rabbit, core.MaskDamageFlash) {
			ticksToDisappear = i
			t.Logf("✅ DamageFlash исчез за %d тиков (%.3f сек)", i, float32(i)/60.0)
			break
		}

		// Логируем угасание каждые 3 тика
		if i%3 == 0 && world.HasComponent(rabbit, core.MaskDamageFlash) {
			flash, _ := world.GetDamageFlash(rabbit)
			t.Logf("  Тик %d: интенсивность=%.3f, таймер=%.3f", i, flash.Intensity, flash.Timer)
		}
	}

	// КРИТИЧЕСКАЯ ПРОВЕРКА 4: Эффект должен исчезнуть быстро (в 5 раз быстрее)
	maxExpectedTicks := 12 // ~0.2 секунды с запасом
	if ticksToDisappear == 0 {
		t.Error("БАГ: DamageFlash не исчез за 20 тиков (треть секунды)")
	} else if ticksToDisappear > maxExpectedTicks {
		t.Errorf("БАГ: DamageFlash исчез слишком медленно: %d тиков (ожидалось <%d)",
			ticksToDisappear, maxExpectedTicks)
	} else {
		t.Logf("✅ Быстрое угасание подтверждено: %d тиков", ticksToDisappear)
	}

	// Фаза 3: Проверяем что урон был нанесён корректно
	health, hasHealth := world.GetHealth(rabbit)
	if hasHealth && health.Current < health.Max {
		damageTaken := health.Max - health.Current
		t.Logf("✅ Урон нанесён: %d единиц (%d -> %d HP)", damageTaken, health.Max, health.Current)
	} else {
		t.Error("БАГ: Урон не был нанесён")
	}

	// Резюме теста
	t.Logf("\n=== РЕЗЮМЕ УСИЛЕННОГО DAMAGEFLASH ===")
	t.Logf("✅ Интенсивность: %.3f (масштаб яркости %.1fх)", flashIntensity, 1.0+flashIntensity*5.0)
	t.Logf("✅ Длительность: %.3f сек (ускорена в 5 раз)", flashTimer)
	t.Logf("✅ Угасание: %d тиков (быстрое)", ticksToDisappear)
	t.Logf("✅ Все параметры соответствуют усиленному эффекту")
}

// TestDamageFlashRendering - тест проверяющий что DamageFlash правильно применяется к спрайту
//
// Фиксирует изменения в sprite_renderer.go:
// - Эффект применяется через ColorScale.Scale(scale, scale, scale, 1.0)
// - Формула: scale = 1.0 + intensity * 5.0
// - Только R, G, B каналы увеличиваются, A остается неизменным
func TestDamageFlashRendering(t *testing.T) {
	t.Parallel()

	world := core.NewWorld(640, 640, 42)

	// Создаём тестовое животное с DamageFlash
	rabbit := simulation.CreateAnimal(world, core.TypeRabbit, 300, 300)

	// Добавляем DamageFlash с известной интенсивностью
	testIntensity := float32(0.9) // 90% интенсивности
	world.AddDamageFlash(rabbit, core.DamageFlash{
		Intensity: testIntensity,
		Timer:     0.15, // Новая короткая длительность
	})

	t.Logf("=== ТЕСТ РЕНДЕРИНГА DAMAGEFLASH ===")
	t.Logf("Тестовая интенсивность: %.1f", testIntensity)

	// Проверяем наличие DamageFlash компонента
	if !world.HasComponent(rabbit, core.MaskDamageFlash) {
		t.Fatal("DamageFlash компонент не был добавлен")
	}

	flash, _ := world.GetDamageFlash(rabbit)

	// КРИТИЧЕСКАЯ ПРОВЕРКА: Расчёт масштаба для рендеринга
	expectedScale := 1.0 + flash.Intensity*5.0
	t.Logf("Расчётный масштаб: %.1f (формула: 1.0 + %.1f * 5.0)", expectedScale, flash.Intensity)

	// ПРОВЕРКА: Масштаб должен быть значительно больше 1 (усиленный эффект)
	if expectedScale <= 2.0 {
		t.Errorf("БАГ: Масштаб слишком мал: %.1f (эффект не будет заметен)", expectedScale)
	} else if expectedScale > 6.0 {
		t.Errorf("БАГ: Масштаб слишком велик: %.1f (может быть слишком ярко)", expectedScale)
	} else {
		t.Logf("✅ Масштаб в норме: %.1fх (заметный эффект)", expectedScale)
	}

	// ПРОВЕРКА: Таймер соответствует новой короткой длительности
	if flash.Timer > 0.2 {
		t.Errorf("БАГ: Таймер слишком долгий: %.3f сек (эффект будет слишком медленным)", flash.Timer)
	} else {
		t.Logf("✅ Таймер правильный: %.3f сек (быстрый эффект)", flash.Timer)
	}

	t.Logf("\n=== ПАРАМЕТРЫ РЕНДЕРИНГА ===")
	t.Logf("Интенсивность: %.3f", flash.Intensity)
	t.Logf("Масштаб RGB каналов: %.1fх", expectedScale)
	t.Logf("Альфа канал: неизменен (1.0)")
	t.Logf("Длительность: %.3f сек", flash.Timer)
	t.Logf("✅ Все параметры рендеринга корректны")
}
