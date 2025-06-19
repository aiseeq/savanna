package integration

import (
	"testing"

	"github.com/aiseeq/savanna/internal/animation"
	"github.com/aiseeq/savanna/internal/core"
	"github.com/aiseeq/savanna/internal/simulation"
)

// TestWolfAnimationSimple простой тест проблемы сохранения анимации
//
//nolint:gocognit,revive,funlen // Комплексный тест системы анимации волков
func TestWolfAnimationSimple(t *testing.T) {
	t.Parallel()
	world := core.NewWorld(640, 640, 42)

	// Создаём волка
	wolf := simulation.CreateAnimal(world, core.TypeWolf, 300, 300)

	// Добавляем анимацию волку
	animComp := core.Animation{
		CurrentAnim: int(animation.AnimIdle),
		Frame:       0,
		Timer:       0,
		Playing:     true,
		FacingRight: true,
	}
	world.SetAnimation(wolf, animComp)

	t.Logf("=== Тест базового сохранения анимации ===")

	// Читаем то что только что записали
	if anim, hasAnim := world.GetAnimation(wolf); hasAnim {
		t.Logf("Исходное состояние: анимация %s, кадр %d", animation.AnimationType(anim.CurrentAnim).String(), anim.Frame)

		// Меняем анимацию на Attack
		anim.CurrentAnim = int(animation.AnimAttack)
		anim.Frame = 0
		anim.Timer = 0
		world.SetAnimation(wolf, anim)
		t.Logf("Устанавливаем: анимация Attack, кадр 0")

		// Сразу читаем обратно
		if newAnim, hasNewAnim := world.GetAnimation(wolf); hasNewAnim {
			t.Logf("Прочитали обратно: анимация %s, кадр %d",
				animation.AnimationType(newAnim.CurrentAnim).String(), newAnim.Frame)

			if newAnim.CurrentAnim != int(animation.AnimAttack) {
				t.Errorf("ОШИБКА: Установили Attack (%d), но получили %s (%d)",
					int(animation.AnimAttack), animation.AnimationType(newAnim.CurrentAnim).String(), newAnim.CurrentAnim)
			}
		} else {
			t.Errorf("ОШИБКА: Не удалось прочитать анимацию после установки")
		}
	} else {
		t.Errorf("ОШИБКА: Не удалось прочитать исходную анимацию")
	}

	t.Logf("=== Тест множественных изменений ===")

	// Проверяем что происходит при множественных изменениях
	for i := 0; i < 5; i++ {
		if anim, hasAnim := world.GetAnimation(wolf); hasAnim {
			oldType := animation.AnimationType(anim.CurrentAnim)

			// Переключаем между Idle и Attack
			var newType animation.AnimationType
			if anim.CurrentAnim == int(animation.AnimIdle) {
				newType = animation.AnimAttack
			} else {
				newType = animation.AnimIdle
			}

			anim.CurrentAnim = int(newType)
			anim.Frame = 0
			anim.Timer = 0
			world.SetAnimation(wolf, anim)

			// Проверяем что сохранилось
			if checkAnim, hasCheckAnim := world.GetAnimation(wolf); hasCheckAnim {
				actualType := animation.AnimationType(checkAnim.CurrentAnim)
				t.Logf("  Итерация %d: %s -> %s (ожидали %s)", i, oldType.String(), actualType.String(), newType.String())

				if actualType != newType {
					t.Errorf("ОШИБКА: Ожидали %s, получили %s", newType.String(), actualType.String())
				}
			} else {
				t.Errorf("ОШИБКА: Не удалось прочитать анимацию в итерации %d", i)
			}
		} else {
			t.Errorf("ОШИБКА: Не удалось прочитать анимацию в итерации %d", i)
		}
	}
}
