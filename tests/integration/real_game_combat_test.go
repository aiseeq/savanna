package integration

import (
	"testing"

	"github.com/aiseeq/savanna/internal/adapters"
	"github.com/aiseeq/savanna/internal/animation"
	"github.com/aiseeq/savanna/internal/core"
	"github.com/aiseeq/savanna/internal/simulation"
)

// TestRealGameCombat тест боя в условиях реальной игры (с полными системами)
//
//nolint:gocognit,revive,funlen // Комплексный тест боевой системы в реальных условиях
func TestRealGameCombat(t *testing.T) {
	t.Parallel()
	world := core.NewWorld(1600, 1600, 42) // Больший мир

	// ИСПРАВЛЕНО: Правильный порядок систем согласно CLAUDE.md
	systemManager := core.NewSystemManager()

	// СНАЧАЛА CombatSystem (создает AttackState)
	combatSystem := simulation.NewCombatSystem()
	systemManager.AddSystem(combatSystem)

	// ПОТОМ BehaviorSystem (видит AttackState и пропускает атакующих)
	animalBehaviorSystem := simulation.NewAnimalBehaviorSystem(nil) // nil vegetation для теста
	systemManager.AddSystem(&adapters.BehaviorSystemAdapter{System: animalBehaviorSystem})

	// ПОСЛЕДНИМ MovementSystem (движение)
	movementSystem := simulation.NewMovementSystem(1600, 1600)
	systemManager.AddSystem(&adapters.MovementSystemAdapter{System: movementSystem})

	// Создаём анимационную систему
	wolfAnimationSystem := animation.NewAnimationSystem()
	wolfAnimationSystem.RegisterAnimation(animation.AnimIdle, 2, 1.0, true, nil)
	wolfAnimationSystem.RegisterAnimation(animation.AnimWalk, 2, 4.0, true, nil)
	wolfAnimationSystem.RegisterAnimation(animation.AnimRun, 2, 8.0, true, nil)
	wolfAnimationSystem.RegisterAnimation(animation.AnimAttack, 2, 6.0, false, nil)
	wolfAnimationSystem.RegisterAnimation(animation.AnimEat, 2, 2.0, true, nil)

	// Создаём животных рядом друг с другом (для тайловой системы)
	rabbit := simulation.CreateAnimal(world, core.TypeRabbit, 800, 800)
	wolf := simulation.CreateAnimal(world, core.TypeWolf, 801, 800) // Дистанция 1 пиксель < атака волка 15 тайлов

	// Делаем волка очень голодным
	world.SetSatiation(wolf, core.Satiation{Value: 10.0}) // 10% - очень голодный

	t.Logf("=== ТЕСТ РЕАЛЬНОЙ ИГРЫ ===")
	t.Logf("Заяц: (800, 800), Волк: (801, 800), расстояние: 1 пиксель")

	initialHealth, _ := world.GetHealth(rabbit)
	initialHunger, _ := world.GetSatiation(wolf)
	t.Logf("Здоровье зайца: %d, голод волка: %.1f%%", initialHealth.Current, initialHunger.Value)

	deltaTime := float32(1.0 / 60.0)
	attackDetected := false
	damageFlashDetected := false
	deathDetected := false
	eatingDetected := false

	// Симулируем игру в течение 10 секунд
	for i := 0; i < 600; i++ {
		world.Update(deltaTime)

		// Обновляем анимации как в реальной игре
		updateWolfAnimation(world, wolf, wolfAnimationSystem, deltaTime)

		// Обновляем все системы
		systemManager.Update(world, deltaTime)

		// Отслеживаем события
		currentHealth, _ := world.GetHealth(rabbit)
		if !attackDetected && currentHealth.Current < initialHealth.Current {
			attackDetected = true
			t.Logf("✅ Тик %d: АТАКА обнаружена! Здоровье %d -> %d",
				i, initialHealth.Current, currentHealth.Current)
		}

		if !damageFlashDetected && world.HasComponent(rabbit, core.MaskDamageFlash) {
			damageFlashDetected = true
			flash, _ := world.GetDamageFlash(rabbit)
			t.Logf("✅ Тик %d: БЛИНК обнаружен! Таймер %.2f", i, flash.Timer)
		}

		if !deathDetected && world.HasComponent(rabbit, core.MaskCorpse) {
			deathDetected = true
			t.Logf("✅ Тик %d: СМЕРТЬ обнаружена! Заяц превратился в труп", i)
		}

		if !eatingDetected && world.HasComponent(wolf, core.MaskEatingState) {
			eatingDetected = true
			t.Logf("✅ Тик %d: ПОЕДАНИЕ началось!", i)
		}

		// Логируем состояние каждые 2 секунды
		if i%120 == 0 {
			health, _ := world.GetHealth(rabbit)
			hunger, _ := world.GetSatiation(wolf)
			wolfPos, _ := world.GetPosition(wolf)
			rabbitPos, _ := world.GetPosition(rabbit)
			anim, _ := world.GetAnimation(wolf)

			distance := (wolfPos.X-rabbitPos.X)*(wolfPos.X-rabbitPos.X) + (wolfPos.Y-rabbitPos.Y)*(wolfPos.Y-rabbitPos.Y)

			status := "жив"
			if world.HasComponent(rabbit, core.MaskCorpse) {
				status = "труп"
			} else if !world.IsAlive(rabbit) {
				status = "съеден"
			}

			t.Logf("%.1fс: заяц %s (HP %d), волк (%.1f,%.1f) голод %.1f%%, анимация %s, дистанция %.1f",
				float32(i)/60.0, status, health.Current, wolfPos.X, wolfPos.Y, hunger.Value,
				animation.AnimationType(anim.CurrentAnim).String(), distance)
		}

		// Если труп съеден - завершаем тест
		if deathDetected && !world.IsAlive(rabbit) {
			finalHunger, _ := world.GetSatiation(wolf)
			t.Logf("🎉 ПОЛНЫЙ ЦИКЛ ЗАВЕРШЁН на тике %d (%.1f сек)", i, float32(i)/60.0)
			t.Logf("Голод волка: %.1f%% -> %.1f%%", initialHunger.Value, finalHunger.Value)
			break
		}
	}

	// Проверяем результаты
	t.Logf("\n=== РЕЗУЛЬТАТЫ ===")
	t.Logf("Атака обнаружена: %t", attackDetected)
	t.Logf("Блинк обнаружен: %t", damageFlashDetected)
	t.Logf("Смерть обнаружена: %t", deathDetected)
	t.Logf("Поедание обнаружено: %t", eatingDetected)

	if !attackDetected {
		t.Error("❌ Атака НЕ произошла в реальной игре!")
	}
	if !damageFlashDetected {
		t.Error("❌ Блинк урона НЕ работает в реальной игре!")
	}
	if !deathDetected {
		t.Error("❌ Смерть НЕ произошла в реальной игре!")
	}
	if !eatingDetected {
		t.Error("❌ Поедание НЕ началось в реальной игре!")
	}
}

// updateWolfAnimation обновляет анимацию волка как в реальной игре
//
//nolint:gocognit,revive // Вспомогательная функция теста для анимации волка
func updateWolfAnimation(
	world *core.World, wolf core.EntityID, animSystem *animation.AnimationSystem, deltaTime float32,
) {
	if anim, hasAnim := world.GetAnimation(wolf); hasAnim {
		// Определяем нужную анимацию как в main.go
		var newAnimType animation.AnimationType
		if world.HasComponent(wolf, core.MaskEatingState) {
			newAnimType = animation.AnimEat
		} else if isWolfAttacking(world, wolf) {
			newAnimType = animation.AnimAttack
		} else {
			// Проверяем движение
			velocity, hasVel := world.GetVelocity(wolf)
			if hasVel {
				speed := velocity.X*velocity.X + velocity.Y*velocity.Y
				if speed < 0.1 {
					newAnimType = animation.AnimIdle
				} else if speed < 400.0 {
					newAnimType = animation.AnimWalk
				} else {
					newAnimType = animation.AnimRun
				}
			} else {
				newAnimType = animation.AnimIdle
			}
		}

		// Если анимация изменилась, сбрасываем её
		if anim.CurrentAnim != int(newAnimType) {
			anim.CurrentAnim = int(newAnimType)
			anim.Frame = 0
			anim.Timer = 0
			anim.Playing = true
			world.SetAnimation(wolf, anim)
		}

		// Обновляем направление взгляда
		if velocity, hasVel := world.GetVelocity(wolf); hasVel {
			if velocity.X > 0.1 {
				anim.FacingRight = true
			} else if velocity.X < -0.1 {
				anim.FacingRight = false
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

		// Сохраняем обновленное состояние
		anim.Frame = animComponent.Frame
		anim.Timer = animComponent.Timer
		anim.Playing = animComponent.Playing
		anim.FacingRight = animComponent.FacingRight
		world.SetAnimation(wolf, anim)
	}
}
