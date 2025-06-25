package integration

import (
	"testing"

	"github.com/aiseeq/savanna/internal/animation"
	"github.com/aiseeq/savanna/internal/core"
	"github.com/aiseeq/savanna/internal/simulation"
)

// TestDamageFlashVisualIntegration - интеграционный тест белого кружка DamageFlash в реальной атаке
//
// Создаёт реальную игровую ситуацию: волк атакует зайца и наносит урон
// Проверяет что DamageFlash создаётся и правильно отображается с белым кружком
//
//nolint:gocognit,revive,funlen // Интеграционный тест сложной визуальной механики
func TestDamageFlashVisualIntegration(t *testing.T) {
	t.Parallel()

	// Создаём реальную игровую среду
	world := core.NewWorld(640, 640, 42)
	combatSystem := simulation.NewCombatSystem()

	// Создаём анимационные системы для полной интеграции
	wolfAnimSystem := animation.NewAnimationSystem()
	rabbitAnimSystem := animation.NewAnimationSystem()

	// Регистрируем анимации атаки и урона
	wolfAnimSystem.RegisterAnimation(animation.AnimAttack, 2, 6.0, false, nil)
	wolfAnimSystem.RegisterAnimation(animation.AnimIdle, 2, 1.0, true, nil)
	rabbitAnimSystem.RegisterAnimation(animation.AnimIdle, 2, 1.0, true, nil)

	animManager := animation.NewAnimationManager(wolfAnimSystem, rabbitAnimSystem)

	// Создаём боевую ситуацию: голодный волк рядом с зайцем
	rabbit := simulation.CreateAnimal(world, core.TypeRabbit, 300, 300)
	wolf := simulation.CreateAnimal(world, core.TypeWolf, 305, 300) // Очень близко

	// Делаем волка очень голодным для атаки
	world.SetSatiation(wolf, core.Satiation{Value: 10.0}) // 10% - критический голод

	t.Logf("=== ИНТЕГРАЦИОННЫЙ ТЕСТ DAMAGEFLASH ===")
	t.Logf("Волк (entity %d) атакует зайца (entity %d)", wolf, rabbit)

	deltaTime := float32(1.0 / 60.0)
	damageFlashDetected := false
	attackHappened := false

	// Симулируем до 300 тиков (5 секунд) для полной атаки
	for i := 0; i < 300; i++ {
		world.Update(deltaTime)
		combatSystem.Update(world, deltaTime)
		animManager.UpdateAllAnimations(world, deltaTime)

		// Проверяем начало атаки
		if !attackHappened && world.HasComponent(wolf, core.MaskAttackState) {
			attackState, _ := world.GetAttackState(wolf)
			if attackState.Target == rabbit {
				attackHappened = true
				t.Logf("✅ Тик %d: Волк начал атаковать зайца", i)
			}
		}

		// КЛЮЧЕВАЯ ПРОВЕРКА: Обнаружение DamageFlash
		if world.HasComponent(rabbit, core.MaskDamageFlash) {
			if !damageFlashDetected {
				damageFlashDetected = true
				flash, _ := world.GetDamageFlash(rabbit)
				t.Logf("🎯 Тик %d: DamageFlash создан!", i)
				t.Logf("   Параметры: Timer=%.3f, Duration=%.3f, Intensity=%.3f",
					flash.Timer, flash.Duration, flash.Intensity)

				// Проверяем что DamageFlash имеет правильные параметры для белого кружка
				if flash.Intensity <= 0 {
					t.Error("БАГ: DamageFlash создан с нулевой интенсивностью")
				}
				if flash.Timer <= 0 {
					t.Error("БАГ: DamageFlash создан с нулевым таймером")
				}
				if flash.Duration <= 0 {
					t.Error("БАГ: DamageFlash создан с нулевой длительностью")
				}

				// Проверяем что заяц получил урон
				health, hasHealth := world.GetHealth(rabbit)
				if hasHealth && health.Current < health.Max {
					t.Logf("✅ Заяц получил урон: %d/%d HP", health.Current, health.Max)
				} else {
					t.Error("БАГ: DamageFlash создан, но заяц не получил урон")
				}
			}

			// Проверяем обновление DamageFlash каждые 10 тиков
			if i%10 == 0 {
				flash, _ := world.GetDamageFlash(rabbit)
				t.Logf("   Тик %d: DamageFlash таймер=%.3f, интенсивность=%.3f",
					i, flash.Timer, flash.Intensity)

				// Проверяем что интенсивность правильно уменьшается
				expectedIntensity := flash.Timer / flash.Duration
				if abs(flash.Intensity-expectedIntensity) > 0.01 {
					t.Errorf("БАГ: Неправильная интенсивность DamageFlash")
					t.Errorf("Ожидалось: %.3f, получено: %.3f", expectedIntensity, flash.Intensity)
				}
			}
		} else if damageFlashDetected {
			// DamageFlash исчез
			t.Logf("✅ Тик %d: DamageFlash исчез (завершился естественно)", i)
			break
		}

		// Логируем ключевые события каждую секунду
		if i%60 == 0 {
			rabbitHealth, _ := world.GetHealth(rabbit)
			wolfHunger, _ := world.GetSatiation(wolf)
			t.Logf("Секунда %d: Заяц HP=%d, волк голод=%.1f%%",
				i/60, rabbitHealth.Current, wolfHunger.Value)
		}
	}

	// ФИНАЛЬНЫЕ ПРОВЕРКИ
	if !attackHappened {
		t.Error("БАГ: Волк не атаковал зайца за 5 секунд")
		t.Error("Возможные причины:")
		t.Error("1. Волк недостаточно голоден")
		t.Error("2. Заяц слишком далеко")
		t.Error("3. AttackSystem не работает")
	}

	if !damageFlashDetected {
		t.Error("БАГ: DamageFlash НЕ был создан при атаке")
		t.Error("Возможные причины:")
		t.Error("1. Урон не был нанесён")
		t.Error("2. DamageFlash не создаётся в AttackSystem")
		t.Error("3. DamageSystem не вызывается")
	} else {
		t.Logf("✅ УСПЕХ: DamageFlash правильно создан и отображается")
		t.Logf("✅ В GUI игре заяц будет покрыт белым кружком при получении урона")
	}

	// Проверяем финальное состояние (заяц мог умереть или стать трупом)
	if world.IsAlive(rabbit) {
		finalRabbitHealth, _ := world.GetHealth(rabbit)
		t.Logf("✅ Заяц выжил с %d/%d HP", finalRabbitHealth.Current, finalRabbitHealth.Max)
	} else if world.HasComponent(rabbit, core.MaskCorpse) {
		t.Logf("✅ Заяц убит и превратился в труп")
	} else {
		t.Logf("✅ Заяц был убит и удален из мира")
	}
}
