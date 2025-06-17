package integration

import (
	"testing"

	"github.com/aiseeq/savanna/internal/animation"
	"github.com/aiseeq/savanna/internal/core"
	"github.com/aiseeq/savanna/internal/simulation"
)

// TestWolfAnimationBehavior тестирует анимации волка до и после поедания зайца
func TestWolfAnimationBehavior(t *testing.T) {
	t.Parallel()
	world := core.NewWorld(640, 640, 42)
	combatSystem := simulation.NewCombatSystem()

	// Создаём GUI систему анимаций для волка
	wolfAnimationSystem := animation.NewAnimationSystem()

	// Регистрируем анимации волка (как в игре)
	wolfAnimationSystem.RegisterAnimation(animation.AnimIdle, 2, 2.0, true, nil)
	wolfAnimationSystem.RegisterAnimation(animation.AnimWalk, 2, 4.0, true, nil)
	wolfAnimationSystem.RegisterAnimation(animation.AnimRun, 2, 12.0, true, nil)
	wolfAnimationSystem.RegisterAnimation(animation.AnimAttack, 2, 5.0, true, nil)
	wolfAnimationSystem.RegisterAnimation(animation.AnimEat, 2, 4.0, true, nil)

	// Создаём зайца и волка в радиусе атаки
	rabbit := simulation.CreateRabbit(world, 300, 300)
	wolf := simulation.CreateWolf(world, 305, 300) // На расстоянии 5 единиц (в радиусе 12)

	// Добавляем анимационный компонент волку (как в GUI игре)
	animComp := core.Animation{
		CurrentAnim: int(animation.AnimIdle),
		Frame:       0,
		Timer:       0,
		Playing:     true,
		FacingRight: true,
	}
	world.SetAnimation(wolf, animComp)

	// Делаем волка голодным
	world.SetHunger(wolf, core.Hunger{Value: 30.0})

	deltaTime := float32(1.0 / 60.0)

	t.Logf("=== Тест анимации волка до и после поедания зайца ===")

	// Фаза 1: до атаки
	t.Logf("Фаза 1: Волк голодный, начинает охоту")

	for i := 0; i < 120; i++ {
		world.Update(deltaTime)
		combatSystem.Update(world, deltaTime)

		// Эмулируем анимационную систему для тестов (как в других тестах)
		if world.HasComponent(wolf, core.MaskAttackState) {
			attackState, _ := world.GetAttackState(wolf)

			// Сразу переводим в Strike фазу для нанесения урона
			if attackState.Phase == core.AttackPhaseWindup {
				// Устанавливаем анимацию ATTACK кадр 1 для Strike
				world.SetAnimation(wolf, core.Animation{
					CurrentAnim: int(animation.AnimAttack),
					Frame:       1, // Сразу Strike для быстрого урона
					Timer:       0,
					Playing:     true,
					FacingRight: true,
				})
			} else if attackState.Phase == core.AttackPhaseStrike && attackState.HasStruck {
				// После нанесения удара завершаем анимацию
				world.SetAnimation(wolf, core.Animation{
					CurrentAnim: int(animation.AnimAttack),
					Frame:       1,
					Timer:       0,
					Playing:     false, // Анимация завершена
					FacingRight: true,
				})
			}
		}

		// Обновляем анимацию
		if anim, hasAnim := world.GetAnimation(wolf); hasAnim {
			animComp := animation.AnimationComponent{
				CurrentAnim: animation.AnimationType(anim.CurrentAnim),
				Frame:       anim.Frame,
				Timer:       anim.Timer,
				Playing:     anim.Playing,
				FacingRight: anim.FacingRight,
			}
			wolfAnimationSystem.Update(&animComp, deltaTime)

			// Сохраняем обратно в мир
			world.SetAnimation(wolf, core.Animation{
				CurrentAnim: int(animComp.CurrentAnim),
				Frame:       animComp.Frame,
				Timer:       animComp.Timer,
				Playing:     animComp.Playing,
				FacingRight: animComp.FacingRight,
			})
		}

		if i%6 == 0 { // каждые 6 тиков (0.1 сек)
			anim, _ := world.GetAnimation(wolf)
			animData := wolfAnimationSystem.GetAnimation(animation.AnimationType(anim.CurrentAnim))
			framesCount := "неизвестно"
			if animData != nil {
				framesCount = string(rune('0' + animData.Frames))
			}

			// Проверяем дистанцию и здоровье
			wolfPos, _ := world.GetPosition(wolf)
			rabbitPos, _ := world.GetPosition(rabbit)
			rabbitHealth, _ := world.GetHealth(rabbit)
			distance := ((wolfPos.X-rabbitPos.X)*(wolfPos.X-rabbitPos.X) + (wolfPos.Y-rabbitPos.Y)*(wolfPos.Y-rabbitPos.Y))

			t.Logf("  Тик %2d: анимация %s, кадр %d/%s, дистанция %.1f, здоровье зайца %d",
				i, animation.AnimationType(anim.CurrentAnim).String(), anim.Frame, framesCount, distance, rabbitHealth.Current)
		}

		// Проверяем жив ли заяц
		rabbitHealth, _ := world.GetHealth(rabbit)
		if rabbitHealth.Current <= 0 || !world.IsAlive(rabbit) {
			t.Logf("Заяц умер на тике %d", i)
			break
		}
	}

	// Фаза 2: после поедания
	rabbitHealth, _ := world.GetHealth(rabbit)
	if rabbitHealth.Current <= 0 || !world.IsAlive(rabbit) {
		t.Logf("Фаза 2: Заяц съеден, проверяем анимации волка")

		for i := 0; i < 60; i++ {
			world.Update(deltaTime)
			combatSystem.Update(world, deltaTime)

			// Обновляем анимацию
			if anim, hasAnim := world.GetAnimation(wolf); hasAnim {
				animComp := animation.AnimationComponent{
					CurrentAnim: animation.AnimationType(anim.CurrentAnim),
					Frame:       anim.Frame,
					Timer:       anim.Timer,
					Playing:     anim.Playing,
					FacingRight: anim.FacingRight,
				}
				wolfAnimationSystem.Update(&animComp, deltaTime)

				// Сохраняем обратно в мир
				world.SetAnimation(wolf, core.Animation{
					CurrentAnim: int(animComp.CurrentAnim),
					Frame:       animComp.Frame,
					Timer:       animComp.Timer,
					Playing:     animComp.Playing,
					FacingRight: animComp.FacingRight,
				})
			}

			if i%6 == 0 { // каждые 6 тиков (0.1 сек)
				anim, _ := world.GetAnimation(wolf)
				animData := wolfAnimationSystem.GetAnimation(animation.AnimationType(anim.CurrentAnim))
				framesCount := "неизвестно"
				if animData != nil {
					framesCount = string(rune('0' + animData.Frames))
				}

				t.Logf("  Тик %2d: анимация %s, кадр %d, кадров в анимации: %s",
					i, animation.AnimationType(anim.CurrentAnim).String(), anim.Frame, framesCount)
			}
		}
	} else {
		t.Errorf("Заяц должен был быть съеден, но остался жив")
	}
}
