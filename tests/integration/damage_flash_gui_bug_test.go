package integration

import (
	"testing"

	"github.com/aiseeq/savanna/internal/core"
	"github.com/aiseeq/savanna/internal/simulation"
)

// TestDamageFlashGUIBug - TDD тест для проверки обновления DamageFlash в GUI режиме
//
// БАГ: DamageFlash создаётся, но не обновляется в GUI режиме (DamageSystem не вызывается?)
// ОЖИДАНИЕ: DamageFlash должен автоматически уменьшать таймер и исчезать
//
//nolint:gocognit,revive,funlen // TDD тест для проверки системы обновления DamageFlash
func TestDamageFlashGUIBug(t *testing.T) {
	t.Parallel()

	world := core.NewWorld(640, 640, 42)
	combatSystem := simulation.NewCombatSystem()

	// Создаём зайца и наносим ему урон вручную (имитируя реальную атаку)
	rabbit := simulation.CreateAnimal(world, core.TypeRabbit, 300, 300)

	// Вручную создаём DamageFlash как делает система боя
	world.AddDamageFlash(rabbit, core.DamageFlash{
		Timer:     0.8, // DamageFlashDuration из combat.go
		Duration:  0.8,
		Intensity: 1.0,
	})

	t.Logf("=== ТЕСТ ОБНОВЛЕНИЯ DAMAGEFLASH ===")
	t.Logf("DamageFlash создан с таймером 0.8 секунды")

	deltaTime := float32(1.0 / 60.0) // 60 FPS

	// Проверяем что DamageFlash действительно создан
	if !world.HasComponent(rabbit, core.MaskDamageFlash) {
		t.Fatal("DamageFlash не был создан - тест невозможен")
	}

	initialFlash, _ := world.GetDamageFlash(rabbit)
	t.Logf("Начальный таймер DamageFlash: %.3f секунды", initialFlash.Timer)

	// КРИТИЧЕСКАЯ ПРОВЕРКА: DamageFlash должен автоматически обновляться
	// Симулируем обновления как в реальной игре
	for i := 0; i < 60; i++ { // 1 секунда симуляции
		// Обновляем все системы как в GUI режиме
		world.Update(deltaTime)
		combatSystem.Update(world, deltaTime)

		// Проверяем состояние DamageFlash каждые 10 тиков
		if i%10 == 0 {
			if world.HasComponent(rabbit, core.MaskDamageFlash) {
				currentFlash, _ := world.GetDamageFlash(rabbit)
				t.Logf("Тик %d: DamageFlash таймер %.3f секунды", i, currentFlash.Timer)

				// БАГ ДЕТЕКЦИЯ: Если таймер НЕ уменьшается
				if i > 10 && currentFlash.Timer >= initialFlash.Timer {
					t.Errorf("БАГ ОБНАРУЖЕН: DamageFlash таймер НЕ уменьшается!")
					t.Errorf("Начальный таймер: %.3f, текущий: %.3f (тик %d)",
						initialFlash.Timer, currentFlash.Timer, i)
					t.Errorf("Возможные причины:")
					t.Errorf("1. DamageSystem не вызывается в CombatSystem")
					t.Errorf("2. DamageSystem.Update() не обновляет таймеры")
					t.Errorf("3. DamageFlash не сохраняется после обновления")
					return
				}
			} else {
				t.Logf("Тик %d: DamageFlash исчез", i)

				// Проверяем что исчезновение произошло в правильное время
				expectedDisappearTime := initialFlash.Timer
				actualDisappearTime := float32(i) / 60.0

				if actualDisappearTime < expectedDisappearTime*0.8 {
					t.Errorf("БАГ: DamageFlash исчез слишком рано (%.3f сек вместо ~%.3f сек)",
						actualDisappearTime, expectedDisappearTime)
				} else {
					t.Logf("✅ DamageFlash исчез в правильное время: %.3f секунды", actualDisappearTime)
				}
				return
			}
		}
	}

	// Если дошли до конца - DamageFlash не исчез за разумное время
	if world.HasComponent(rabbit, core.MaskDamageFlash) {
		finalFlash, _ := world.GetDamageFlash(rabbit)
		t.Errorf("БАГ: DamageFlash не исчез за 1 секунду (финальный таймер: %.3f)", finalFlash.Timer)
	}
}

// TestDamageFlashVisualEffect - тест для визуального эффекта DamageFlash
func TestDamageFlashVisualEffect(t *testing.T) {
	t.Parallel()

	world := core.NewWorld(640, 640, 42)

	// Создаём зайца и добавляем очень яркий DamageFlash для тестирования
	rabbit := simulation.CreateAnimal(world, core.TypeRabbit, 300, 300)

	// Создаём ОЧЕНЬ ЯРКИЙ DamageFlash чтобы его было точно видно
	world.AddDamageFlash(rabbit, core.DamageFlash{
		Timer:     5.0, // Долгий таймер для тестирования
		Duration:  5.0,
		Intensity: 2.0, // Увеличенная интенсивность (белый кружок)
	})

	t.Logf("=== ТЕСТ ВИЗУАЛЬНОГО ЭФФЕКТА ===")
	t.Logf("Создан ЯРКИЙ DamageFlash: таймер=5.0 сек, интенсивность=2.0")
	t.Logf("Если запустить GUI игру сейчас, заяц должен быть покрыт белым кружком")

	// Проверяем что компонент создан с правильными параметрами
	if !world.HasComponent(rabbit, core.MaskDamageFlash) {
		t.Fatal("DamageFlash не создан")
	}

	flash, _ := world.GetDamageFlash(rabbit)
	if flash.Intensity < 1.5 {
		t.Error("Интенсивность DamageFlash слишком низкая для тестирования")
	}

	if flash.Timer < 3.0 {
		t.Error("Таймер DamageFlash слишком короткий для тестирования")
	}

	t.Logf("✅ Тестовый DamageFlash готов для визуальной проверки в GUI")
	t.Logf("Параметры: Timer=%.1f, Duration=%.1f, Intensity=%.1f",
		flash.Timer, flash.Duration, flash.Intensity)
}
