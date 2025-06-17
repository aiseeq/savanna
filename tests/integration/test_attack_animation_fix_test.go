package integration

import (
	"testing"

	"github.com/aiseeq/savanna/internal/animation"
	"github.com/aiseeq/savanna/internal/core"
	"github.com/aiseeq/savanna/internal/simulation"
)

// TestAttackAnimationFix проверяет что анимация атаки проигрывается до конца
func TestAttackAnimationFix(t *testing.T) {
	t.Parallel()
	world := core.NewWorld(1600, 1600, 42)
	combatSystem := simulation.NewCombatSystem()
	animSystem := animation.NewAnimationSystem()

	// Регистрируем анимации
	animSystem.RegisterAnimation(animation.AnimAttack, 2, 6.0, false, nil) // не зацикленная!
	animSystem.RegisterAnimation(animation.AnimIdle, 2, 1.0, true, nil)

	rabbit := simulation.CreateRabbit(world, 800, 800)
	wolf := simulation.CreateWolf(world, 810, 800)
	world.SetHunger(wolf, core.Hunger{Value: 10.0})

	initialHealth, _ := world.GetHealth(rabbit)
	t.Logf("=== ТЕСТ ИСПРАВЛЕНИЯ АНИМАЦИИ АТАКИ ===")
	t.Logf("Начальное здоровье зайца: %d", initialHealth.Current)

	deltaTime := float32(1.0 / 60.0)

	// Имитируем updateAnimalAnimations как в игре
	updateWolfAnimationFixed := func() {
		if anim, hasAnim := world.GetAnimation(wolf); hasAnim {
			// Определяем нужную анимацию
			var newAnimType animation.AnimationType
			if isWolfAttacking(world, wolf) {
				newAnimType = animation.AnimAttack
			} else {
				newAnimType = animation.AnimIdle
			}

			// ИСПРАВЛЕННАЯ ЛОГИКА: НЕ прерываем анимацию ATTACK
			if anim.CurrentAnim != int(newAnimType) {
				if anim.CurrentAnim == int(animation.AnimAttack) && anim.Playing {
					// НЕ меняем анимацию атаки пока она играет!
					t.Logf("  Анимация атаки играет - НЕ сбрасываем (кадр %d)", anim.Frame)
				} else {
					anim.CurrentAnim = int(newAnimType)
					anim.Frame = 0
					anim.Timer = 0
					anim.Playing = true
					world.SetAnimation(wolf, anim)
					t.Logf("  Переключение на %s", newAnimType.String())
				}
			}

			// Обновляем анимацию
			animComponent := animation.AnimationComponent{
				CurrentAnim: animation.AnimationType(anim.CurrentAnim),
				Frame:       anim.Frame,
				Timer:       anim.Timer,
				Playing:     anim.Playing,
				FacingRight: anim.FacingRight,
			}

			animSystem.Update(&animComponent, deltaTime)

			// Сохраняем состояние
			anim.Frame = animComponent.Frame
			anim.Timer = animComponent.Timer
			anim.Playing = animComponent.Playing
			world.SetAnimation(wolf, anim)
		}
	}

	attackFramesSeen := make(map[int]bool)
	damageDealt := false

	// Симулируем 3 секунды
	for i := 0; i < 180; i++ {
		world.Update(deltaTime)
		updateWolfAnimationFixed()
		combatSystem.Update(world, deltaTime)

		// Отслеживаем кадры анимации атаки
		if anim, hasAnim := world.GetAnimation(wolf); hasAnim {
			if anim.CurrentAnim == int(animation.AnimAttack) {
				attackFramesSeen[anim.Frame] = true

				if i%10 == 0 { // логируем каждые 10 тиков
					t.Logf("Тик %3d: ATTACK кадр %d, играет: %t", i, anim.Frame, anim.Playing)
				}
			}
		}

		// Отслеживаем урон
		currentHealth, _ := world.GetHealth(rabbit)
		if !damageDealt && currentHealth.Current < initialHealth.Current {
			damageDealt = true
			anim, _ := world.GetAnimation(wolf)
			t.Logf("🩸 УРОН на тике %d! Кадр анимации: %d", i, anim.Frame)
		}

		// Если заяц умер, прекращаем
		if currentHealth.Current == 0 {
			t.Logf("Заяц умер на тике %d", i)
			break
		}
	}

	// Проверяем результаты
	t.Logf("\n=== РЕЗУЛЬТАТЫ ===")
	t.Logf("Кадры анимации атаки, которые были показаны:")
	for frame := 0; frame <= 1; frame++ {
		if attackFramesSeen[frame] {
			t.Logf("  ✅ Кадр %d: ПОКАЗАН", frame)
		} else {
			t.Logf("  ❌ Кадр %d: НЕ ПОКАЗАН", frame)
		}
	}

	if !attackFramesSeen[1] {
		t.Error("❌ КРИТИЧЕСКАЯ ОШИБКА: 2-й кадр анимации атаки НЕ ПОКАЗАН!")
	} else {
		t.Logf("✅ Анимация атаки проигрывается полностью")
	}

	if !damageDealt {
		t.Error("❌ Урон не был нанесен!")
	} else {
		t.Logf("✅ Урон был нанесен")
	}
}
