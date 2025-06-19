package e2e

import (
	"testing"

	"github.com/aiseeq/savanna/internal/animation"
	"github.com/aiseeq/savanna/internal/core"
	"github.com/aiseeq/savanna/internal/simulation"
)

// TestDamageFlashGUI - E2E тест для проверки DamageFlash в GUI режиме
//
// Проверяет что:
// 1. При нанесении урона создается DamageFlash компонент
// 2. DamageFlash интенсивность уменьшается со временем
// 3. DamageFlash исчезает автоматически
//
//nolint:gocognit,revive,funlen // E2E тест для проверки GUI эффектов
func TestDamageFlashGUI(t *testing.T) {
	t.Parallel()

	world := core.NewWorld(640, 640, 42)
	combatSystem := simulation.NewCombatSystem()

	// Добавляем анимационную систему для работы боевой системы
	wolfAnimSystem := animation.NewAnimationSystem()
	rabbitAnimSystem := animation.NewAnimationSystem()

	// Регистрируем анимации атаки и поедания
	wolfAnimSystem.RegisterAnimation(animation.AnimAttack, 2, 1.0, false, nil)
	wolfAnimSystem.RegisterAnimation(animation.AnimEat, 2, 2.0, true, nil)
	wolfAnimSystem.RegisterAnimation(animation.AnimIdle, 2, 1.0, true, nil)
	rabbitAnimSystem.RegisterAnimation(animation.AnimIdle, 2, 1.0, true, nil)

	animManager := animation.NewAnimationManager(wolfAnimSystem, rabbitAnimSystem)

	// Создаём зайца и волка для реального боя
	rabbit := simulation.CreateAnimal(world, core.TypeRabbit, 300, 300)
	wolf := simulation.CreateAnimal(world, core.TypeWolf, 310, 300) // Рядом для атаки

	// Делаем волка голодным чтобы он атаковал
	world.SetHunger(wolf, core.Hunger{Value: 30.0})

	t.Logf("=== ТЕСТ DAMAGEFLASH В GUI РЕЖИМЕ ===")
	t.Logf("Заяц: entity %d, Волк: entity %d", rabbit, wolf)

	deltaTime := float32(1.0 / 60.0)

	// Фаза 1: Ждём атаки волка (до 3 секунд)
	var attackDetected bool
	for i := 0; i < 180 && !attackDetected; i++ {
		combatSystem.Update(world, deltaTime)
		animManager.UpdateAllAnimations(world, deltaTime)

		// Проверяем появился ли DamageFlash на зайце
		if world.HasComponent(rabbit, core.MaskDamageFlash) {
			flash, _ := world.GetDamageFlash(rabbit)
			t.Logf("✅ DAMAGEFLASH ОБНАРУЖЕН на тике %d! Интенсивность: %.3f", i, flash.Intensity)
			attackDetected = true

			// КРИТИЧЕСКАЯ ПРОВЕРКА: Интенсивность должна быть в разумных пределах
			if flash.Intensity <= 0 || flash.Intensity > 1.0 {
				t.Errorf("БАГ: Неправильная интенсивность DamageFlash: %.3f (ожидалось 0-1)", flash.Intensity)
			}
			break
		}

		// Логируем прогресс каждую секунду
		if i%60 == 0 {
			if world.HasComponent(wolf, core.MaskAnimation) {
				anim, _ := world.GetAnimation(wolf)
				t.Logf("Секунда %d: Анимация волка: anim=%d, frame=%d", i/60, anim.CurrentAnim, anim.Frame)
			}
		}
	}

	if !attackDetected {
		t.Error("БАГ: DamageFlash не появился за 3 секунды")
		return
	}

	// Фаза 2: Отслеживаем угасание DamageFlash
	t.Logf("\n=== ФАЗА УГАСАНИЯ DAMAGEFLASH ===")

	var initialIntensity float32
	if flash, hasFlash := world.GetDamageFlash(rabbit); hasFlash {
		initialIntensity = flash.Intensity
		t.Logf("Начальная интенсивность: %.3f", initialIntensity)
	}

	// Симулируем 2 секунды для полного угасания
	for i := 0; i < 120; i++ {
		combatSystem.Update(world, deltaTime)
		animManager.UpdateAllAnimations(world, deltaTime)

		// Проверяем состояние DamageFlash каждые 10 тиков
		if i%10 == 0 {
			if world.HasComponent(rabbit, core.MaskDamageFlash) {
				flash, _ := world.GetDamageFlash(rabbit)
				t.Logf("Тик %d: DamageFlash интенсивность=%.3f", i, flash.Intensity)

				// ПРОВЕРКА: Интенсивность должна уменьшаться
				if flash.Intensity > initialIntensity {
					t.Errorf("БАГ: Интенсивность увеличилась! Было: %.3f, стало: %.3f",
						initialIntensity, flash.Intensity)
				}
			} else {
				t.Logf("Тик %d: DamageFlash исчез", i)
				break
			}
		}
	}

	// Фаза 3: Проверяем что DamageFlash исчез
	if world.HasComponent(rabbit, core.MaskDamageFlash) {
		flash, _ := world.GetDamageFlash(rabbit)
		t.Logf("⚠️  DamageFlash всё ещё активен: интенсивность=%.3f", flash.Intensity)

		if flash.Intensity > 0.01 { // Небольшая погрешность
			t.Errorf("БАГ: DamageFlash не угас за 2 секунды (интенсивность=%.3f)", flash.Intensity)
		}
	} else {
		t.Logf("✅ DamageFlash правильно исчез")
	}

	// Фаза 4: Проверяем что заяц получил урон
	health, hasHealth := world.GetHealth(rabbit)
	if hasHealth {
		t.Logf("Финальное здоровье зайца: %d/%d", health.Current, health.Max)

		if health.Current >= health.Max {
			t.Error("БАГ: Заяц не получил урон во время атаки с DamageFlash")
		} else {
			t.Logf("✅ Заяц правильно получил урон")
		}
	}

	t.Logf("\n=== РЕЗУЛЬТАТ ТЕСТА ===")
	t.Logf("✅ DamageFlash создаётся при атаке")
	t.Logf("✅ DamageFlash угасает со временем")
	t.Logf("✅ DamageFlash исчезает автоматически")
	t.Logf("✅ Урон наносится корректно")
}
