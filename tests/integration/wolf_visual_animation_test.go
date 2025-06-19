package integration

import (
	"testing"

	"github.com/aiseeq/savanna/internal/animation"
	"github.com/aiseeq/savanna/internal/core"
	"github.com/aiseeq/savanna/internal/simulation"
)

// TestWolfVisualAnimationProblem воспроизводит проблему с 1 кадром у волка
//
//nolint:gocognit,revive,funlen // Диагностический тест проблем анимации волка
func TestWolfVisualAnimationProblem(t *testing.T) {
	t.Parallel()
	world := core.NewWorld(640, 640, 42)
	combatSystem := simulation.NewCombatSystem()

	// Создаём систему анимаций ТОЧНО как в игре
	wolfAnimationSystem := animation.NewAnimationSystem()

	// Регистрируем анимации ТОЧНО как в игре
	wolfAnimationSystem.RegisterAnimation(animation.AnimIdle, 2, 1.0, true, nil)
	wolfAnimationSystem.RegisterAnimation(animation.AnimWalk, 2, 4.0, true, nil)
	wolfAnimationSystem.RegisterAnimation(animation.AnimRun, 2, 8.0, true, nil)
	wolfAnimationSystem.RegisterAnimation(animation.AnimAttack, 2, 6.0, true, nil)
	wolfAnimationSystem.RegisterAnimation(animation.AnimEat, 2, 2.0, true, nil)

	t.Logf("=== Проверка анимационной системы волка ===")

	// Проверяем что все анимации загружены с 2 кадрами
	animations := []animation.AnimationType{
		animation.AnimIdle, animation.AnimWalk, animation.AnimRun,
		animation.AnimAttack, animation.AnimEat,
	}

	for _, animType := range animations {
		animData := wolfAnimationSystem.GetAnimation(animType)
		if animData == nil {
			t.Errorf("Анимация %s не зарегистрирована!", animType.String())
		} else {
			t.Logf("Анимация %s: %d кадров, %.1f FPS", animType.String(), animData.Frames, animData.FPS)
			if animData.Frames != 2 {
				t.Errorf("ПРОБЛЕМА: Анимация %s имеет %d кадров вместо 2!", animType.String(), animData.Frames)
			}
		}
	}

	// Создаём зайца и волка
	rabbit := simulation.CreateAnimal(world, core.TypeRabbit, 300, 300)
	wolf := simulation.CreateAnimal(world, core.TypeWolf, 305, 300) // Очень близко для атаки

	// Добавляем анимацию волку
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

	t.Logf("=== Фаза 1: Волк голодный, должен атаковать ===")

	// Симулируем точно как в updateAnimalAnimations()
	for i := 0; i < 180; i++ { // 3 секунды
		world.Update(deltaTime)
		combatSystem.Update(world, deltaTime)

		// Обновляем анимации ТОЧНО как в игре
		if anim, hasAnim := world.GetAnimation(wolf); hasAnim {
			// Определяем тип анимации как в игре
			animalType, _ := world.GetAnimalType(wolf)
			var newAnimType animation.AnimationType

			if animalType == core.TypeWolf {
				newAnimType = getWolfAnimationType(world, wolf)
			}

			// Если анимация изменилась, сбрасываем её (как в игре)
			if anim.CurrentAnim != int(newAnimType) {
				anim.CurrentAnim = int(newAnimType)
				anim.Frame = 0
				anim.Timer = 0
				anim.Playing = true
				// ВАЖНО: Сохраняем изменения сразу
				world.SetAnimation(wolf, anim)

				t.Logf("  Тик %d: Переключение на %s", i, newAnimType.String())
			}

			// Читаем ОБНОВЛЕННОЕ состояние после возможного изменения
			anim, _ = world.GetAnimation(wolf)

			// Обновляем анимацию
			animComponent := animation.AnimationComponent{
				CurrentAnim: animation.AnimationType(anim.CurrentAnim),
				Frame:       anim.Frame,
				Timer:       anim.Timer,
				Playing:     anim.Playing,
				FacingRight: anim.FacingRight,
			}

			wolfAnimationSystem.Update(&animComponent, deltaTime)

			// Сохраняем состояние
			anim.Frame = animComponent.Frame
			anim.Timer = animComponent.Timer
			anim.Playing = animComponent.Playing
			anim.FacingRight = animComponent.FacingRight
			world.SetAnimation(wolf, anim)
		}

		// Логируем каждые 30 тиков (0.5 сек)
		if i%30 == 0 {
			anim, _ := world.GetAnimation(wolf)
			animData := wolfAnimationSystem.GetAnimation(animation.AnimationType(anim.CurrentAnim))
			maxFrames := 0
			if animData != nil {
				maxFrames = animData.Frames
			}

			rabbitHealth, _ := world.GetHealth(rabbit)
			wolfHunger, _ := world.GetHunger(wolf)

			t.Logf("  %.1fс: %s кадр %d/%d, здоровье зайца %d, голод волка %.0f%%",
				float32(i)/60.0, animation.AnimationType(anim.CurrentAnim).String(),
				anim.Frame+1, maxFrames, rabbitHealth.Current, wolfHunger.Value)
		}

		// Проверяем смерть зайца
		if !world.IsAlive(rabbit) {
			t.Logf("Заяц умер на тике %d", i)

			// Симулируем ещё немного после смерти зайца
			t.Logf("=== Фаза 2: После поедания зайца ===")
			for j := 0; j < 120; j++ { // 2 секунды после смерти
				world.Update(deltaTime)
				combatSystem.Update(world, deltaTime)

				// Обновляем анимации ПРАВИЛЬНО
				if anim, hasAnim := world.GetAnimation(wolf); hasAnim {
					newAnimType := getWolfAnimationType(world, wolf)

					if anim.CurrentAnim != int(newAnimType) {
						t.Logf("  Тик %d: Переключение %s -> %s", j,
							animation.AnimationType(anim.CurrentAnim).String(), newAnimType.String())
						anim.CurrentAnim = int(newAnimType)
						anim.Frame = 0
						anim.Timer = 0
						anim.Playing = true
						// ВАЖНО: Сохраняем изменения сразу
						world.SetAnimation(wolf, anim)
					} else if j < 10 {
						// ОТЛАДКА: Почему происходит постоянное переключение?
						t.Logf("  Тик %d: Анимация НЕ изменилась, остается %s (текущая: %d, новая: %d)",
							j, newAnimType.String(), anim.CurrentAnim, int(newAnimType))
					}

					// Читаем ОБНОВЛЕННОЕ состояние после возможного изменения
					anim, _ = world.GetAnimation(wolf)

					animComponent := animation.AnimationComponent{
						CurrentAnim: animation.AnimationType(anim.CurrentAnim),
						Frame:       anim.Frame,
						Timer:       anim.Timer,
						Playing:     anim.Playing,
						FacingRight: anim.FacingRight,
					}

					wolfAnimationSystem.Update(&animComponent, deltaTime)

					// Сохраняем обновленное состояние
					anim.Frame = animComponent.Frame
					anim.Timer = animComponent.Timer
					anim.Playing = animComponent.Playing
					anim.FacingRight = animComponent.FacingRight
					world.SetAnimation(wolf, anim)
				}

				if j%30 == 0 {
					anim, _ := world.GetAnimation(wolf)
					animData := wolfAnimationSystem.GetAnimation(animation.AnimationType(anim.CurrentAnim))
					maxFrames := 0
					if animData != nil {
						maxFrames = animData.Frames
					}

					wolfHunger, _ := world.GetHunger(wolf)

					t.Logf("  %.1fс: %s кадр %d/%d, голод волка %.0f%%",
						float32(j)/60.0, animation.AnimationType(anim.CurrentAnim).String(),
						anim.Frame+1, maxFrames, wolfHunger.Value)
				}
			}
			break
		}
	}
}

// getWolfAnimationType копия функции из игры для тестирования
func getWolfAnimationType(world *core.World, entity core.EntityID) animation.AnimationType {
	// Проверяем атаку
	if isWolfAttacking(world, entity) {
		return animation.AnimAttack
	}

	// Проверяем сытость
	hunger, hasHunger := world.GetHunger(entity)
	if hasHunger && hunger.Value > 85.0 {
		velocity, hasVel := world.GetVelocity(entity)
		if hasVel {
			speed := velocity.X*velocity.X + velocity.Y*velocity.Y
			if speed < 0.1 {
				return animation.AnimIdle
			}
		}
	}

	velocity, hasVel := world.GetVelocity(entity)
	if !hasVel {
		return animation.AnimIdle
	}

	speed := velocity.X*velocity.X + velocity.Y*velocity.Y

	if speed < 0.1 {
		return animation.AnimIdle
	} else if speed < 400.0 {
		return animation.AnimWalk
	} else {
		return animation.AnimRun
	}
}
