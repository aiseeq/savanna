package integration

import (
	"fmt"
	"testing"

	"github.com/aiseeq/savanna/internal/animation"
	"github.com/aiseeq/savanna/internal/core"
	"github.com/aiseeq/savanna/internal/simulation"
)

// TestWolfFullCycle тестирует полный цикл: атака -> смерть -> поедание -> исчезновение трупа
//
//nolint:gocognit,revive,funlen // Комплексный тест полного жизненного цикла волка
func TestWolfFullCycle(t *testing.T) {
	t.Parallel()
	world := core.NewWorld(640, 640, 42)
	combatSystem := simulation.NewCombatSystem()

	// Создаём анимационные системы для разных животных
	wolfAnimSystem := animation.NewAnimationSystem()
	rabbitAnimSystem := animation.NewAnimationSystem()

	// Регистрируем анимации для волков и зайцев
	wolfAnimSystem.RegisterAnimation(animation.AnimAttack, 2, 6.0, false, nil)
	wolfAnimSystem.RegisterAnimation(animation.AnimEat, 2, 2.0, true, nil)
	wolfAnimSystem.RegisterAnimation(animation.AnimIdle, 2, 1.0, true, nil)

	rabbitAnimSystem.RegisterAnimation(animation.AnimIdle, 2, 1.0, true, nil)
	rabbitAnimSystem.RegisterAnimation(animation.AnimDeathDying, 2, 3.0, false, nil)

	animManager := animation.NewAnimationManager(wolfAnimSystem, rabbitAnimSystem)

	// Создаём зайца и волка рядом
	rabbit := simulation.CreateAnimal(world, core.TypeRabbit, 300, 300)
	wolf := simulation.CreateAnimal(world, core.TypeWolf, 301, 300) // Дистанция 1 пиксель для атаки

	// Делаем волка очень голодным
	world.SetHunger(wolf, core.Hunger{Value: 30.0})

	initialHealth, _ := world.GetHealth(rabbit)
	initialHunger, _ := world.GetHunger(wolf)
	t.Logf("Начальное состояние: здоровье зайца %d, голод волка %.1f%%",
		initialHealth.Current, initialHunger.Value)

	deltaTime := float32(1.0 / 60.0)
	phase := "атака"
	attackCount := 0
	lastHealth := initialHealth.Current
	rabbitDied := false
	rabbitDeathTime := 0
	_ = rabbitDeathTime // используется в логах
	eatingStarted := false
	eatingStartTime := 0

	// Симулируем до 1800 тиков (30 секунд)
	for i := 0; i < 1800; i++ {
		world.Update(deltaTime)

		// Обновляем анимацию волка
		if anim, hasAnim := world.GetAnimation(wolf); hasAnim {
			var newAnimType animation.AnimationType
			if world.HasComponent(wolf, core.MaskEatingState) {
				newAnimType = animation.AnimEat
			} else if isWolfAttacking(world, wolf) {
				newAnimType = animation.AnimAttack
			} else {
				newAnimType = animation.AnimIdle
			}

			if anim.CurrentAnim != int(newAnimType) {
				anim.CurrentAnim = int(newAnimType)
				anim.Frame = 0
				anim.Timer = 0
				anim.Playing = true
				world.SetAnimation(wolf, anim)
			}

			animComponent := animation.AnimationComponent{
				CurrentAnim: animation.AnimationType(anim.CurrentAnim),
				Frame:       anim.Frame,
				Timer:       anim.Timer,
				Playing:     anim.Playing,
				FacingRight: anim.FacingRight,
			}

			wolfAnimSystem.Update(&animComponent, deltaTime)

			anim.Frame = animComponent.Frame
			anim.Timer = animComponent.Timer
			anim.Playing = animComponent.Playing
			world.SetAnimation(wolf, anim)
		}

		combatSystem.Update(world, deltaTime)
		animManager.UpdateAllAnimations(world, deltaTime) // ВАЖНО: обновляем анимации!

		// Отслеживаем атаки
		if world.IsAlive(rabbit) {
			currentHealth, _ := world.GetHealth(rabbit)
			if currentHealth.Current < lastHealth {
				attackCount++
				t.Logf("Атака %d на тике %d: здоровье %d -> %d",
					attackCount, i, lastHealth, currentHealth.Current)
				lastHealth = currentHealth.Current
			}
		}

		// Отслеживаем смерть зайца
		if !rabbitDied && world.HasComponent(rabbit, core.MaskCorpse) {
			rabbitDied = true
			_ = i // rabbitDeathTime используется только для логирования
			phase = "смерть"
			wolfHunger, _ := world.GetHunger(wolf)
			t.Logf("🐰💀 Заяц умер на тике %d после %d атак, голод волка %.1f%%",
				i, attackCount, wolfHunger.Value)
		}

		// Отслеживаем начало поедания
		if !eatingStarted && world.HasComponent(wolf, core.MaskEatingState) {
			eatingStarted = true
			eatingStartTime = i
			phase = "поедание"
			corpse, _ := world.GetCorpse(rabbit)
			t.Logf("🐺🍽️ Волк начал есть на тике %d, питательность трупа %.1f",
				i, corpse.NutritionalValue)
		}

		// Отслеживаем завершение поедания (волк больше не ест)
		if rabbitDied && eatingStarted && !world.HasComponent(wolf, core.MaskEatingState) {
			finalHunger, _ := world.GetHunger(wolf)
			t.Logf("✅ Полный цикл завершён на тике %d (%.1f сек): атаки %d, поедание %d тиков, голод %.1f%% -> %.1f%%",
				i, float32(i)/60.0, attackCount, i-eatingStartTime, initialHunger.Value, finalHunger.Value)

			// Проверяем что все этапы прошли
			if attackCount < 1 {
				t.Errorf("Волк не атаковал (атак: %d)", attackCount)
			}
			if !rabbitDied {
				t.Error("Заяц не умер")
			}
			if !eatingStarted {
				t.Error("Волк не начал есть")
			}
			if finalHunger.Value <= initialHunger.Value {
				t.Errorf("Голод волка не восстановился: %.1f%% -> %.1f%%",
					initialHunger.Value, finalHunger.Value)
			}

			t.Logf("🎉 Тест полного цикла ПРОЙДЕН")
			return
		}

		// Логируем прогресс каждые 2 секунды
		if i%120 == 0 {
			hunger, _ := world.GetHunger(wolf)
			anim, _ := world.GetAnimation(wolf)

			var status string
			if world.IsAlive(rabbit) {
				health, _ := world.GetHealth(rabbit)
				status = fmt.Sprintf("здоровье зайца %d", health.Current)
			} else if world.HasComponent(rabbit, core.MaskCorpse) {
				corpse, _ := world.GetCorpse(rabbit)
				status = fmt.Sprintf("труп, питательность %.1f", corpse.NutritionalValue)
			} else {
				status = "труп съеден"
			}

			t.Logf("%.1fс [%s]: %s, атак %d, голод волка %.0f%%, анимация %s",
				float32(i)/60.0, phase, status, attackCount, hunger.Value,
				animation.AnimationType(anim.CurrentAnim).String())
		}
	}

	// Если дошли сюда - что-то пошло не так
	t.Errorf("Полный цикл не завершился за 30 секунд. Фаза: %s, атак: %d, заяц умер: %t, поедание началось: %t",
		phase, attackCount, rabbitDied, eatingStarted)
}
