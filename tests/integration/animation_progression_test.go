package integration

import (
	"testing"

	"github.com/aiseeq/savanna/internal/animation"
	"github.com/aiseeq/savanna/internal/core"
	"github.com/aiseeq/savanna/internal/simulation"
)

// TestAnimationProgression проверяет что анимация ATTACK проходит от кадра 0 до кадра 1
func TestAnimationProgression(t *testing.T) {
	t.Parallel()
	world := core.NewWorld(96, 96, 42)

	// Создаём анимационную систему КАК В ИГРЕ
	wolfAnimationSystem := animation.NewAnimationSystem()
	wolfAnimationSystem.RegisterAnimation(animation.AnimAttack, 2, 6.0, false, nil) // 2 кадра, НЕ зацикленная

	// Создаём системы
	combatSystem := simulation.NewCombatSystem()
	animalBehaviorSystem := simulation.NewAnimalBehaviorSystem(nil)
	movementSystem := simulation.NewMovementSystem(96, 96)

	// Создаём животных рядом
	rabbit := simulation.CreateRabbit(world, 40, 48)
	wolf := simulation.CreateWolf(world, 45, 48)

	// Волк голоден
	world.SetHunger(wolf, core.Hunger{Value: 5.0})

	t.Logf("=== ТЕСТ ПРОГРЕССИИ АНИМАЦИИ ATTACK ===")

	deltaTime := float32(1.0 / 60.0) // 60 FPS

	// Функция обновления анимации КАК В РЕАЛЬНОЙ ИГРЕ
	updateWolfAnimation := func() animation.AnimationType {
		anim, hasAnim := world.GetAnimation(wolf)
		if !hasAnim {
			return animation.AnimIdle
		}

		// Определяем нужную анимацию
		var newAnimType animation.AnimationType
		if isWolfAttackingSimple(world, wolf) {
			newAnimType = animation.AnimAttack
		} else {
			newAnimType = animation.AnimIdle
		}

		// КРИТИЧЕСКОЕ МЕСТО: НЕ прерываем анимацию ATTACK
		if anim.CurrentAnim != int(newAnimType) {
			if anim.CurrentAnim == int(animation.AnimAttack) && anim.Playing {
				t.Logf("  [SYSTEM] НЕ сбрасываем ATTACK анимацию (кадр %d, играет: %t)", anim.Frame, anim.Playing)
				// НЕ меняем анимацию!
			} else {
				t.Logf("  [SYSTEM] Смена анимации: %s -> %s",
					animation.AnimationType(anim.CurrentAnim).String(), newAnimType.String())
				anim.CurrentAnim = int(newAnimType)
				anim.Frame = 0
				anim.Timer = 0
				anim.Playing = true
				world.SetAnimation(wolf, anim)
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

		oldFrame := animComponent.Frame
		oldPlaying := animComponent.Playing

		wolfAnimationSystem.Update(&animComponent, deltaTime)

		// Логируем изменения
		if oldFrame != animComponent.Frame || oldPlaying != animComponent.Playing {
			t.Logf("  [ANIM] Кадр %d->%d, играет %t->%t",
				oldFrame, animComponent.Frame, oldPlaying, animComponent.Playing)
		}

		// Сохраняем состояние
		anim.Frame = animComponent.Frame
		anim.Timer = animComponent.Timer
		anim.Playing = animComponent.Playing
		world.SetAnimation(wolf, anim)

		return animation.AnimationType(anim.CurrentAnim)
	}

	framesSeenInOrder := []int{}
	lastFrame := -1

	// Симулируем до тех пор пока не увидим полную анимацию или не превысим лимит
	for tick := 0; tick < 300; tick++ {
		world.Update(deltaTime)
		animalBehaviorSystem.Update(world, deltaTime)
		movementSystem.Update(world, deltaTime)
		combatSystem.Update(world, deltaTime)

		currentAnimType := updateWolfAnimation()

		if currentAnimType == animation.AnimAttack {
			anim, _ := world.GetAnimation(wolf)

			// Отслеживаем прогрессию кадров
			if anim.Frame != lastFrame {
				framesSeenInOrder = append(framesSeenInOrder, anim.Frame)
				lastFrame = anim.Frame
				t.Logf("[TICK %3d] ATTACK кадр %d, играет: %t, таймер: %.3f",
					tick, anim.Frame, anim.Playing, anim.Timer)
			}

			// Если анимация завершилась (не играет), останавливаемся
			if !anim.Playing {
				t.Logf("[TICK %3d] Анимация ATTACK завершена", tick)
				break
			}
		}

		// Проверяем ИЗМЕНЕНИЕ урона
		health, _ := world.GetHealth(rabbit)
		if tick == 0 {
			// Запоминаем начальное здоровье
			lastHealth := health.Current
			if lastHealth < 50 {
				t.Logf("[TICK %3d] 🩸 УРОН НАНЕСЕН! Здоровье: %d", tick, health.Current)
			}
		}

		// Если заяц умер, останавливаемся
		if health.Current == 0 {
			t.Logf("[TICK %3d] Заяц умер", tick)
			break
		}
	}

	// АНАЛИЗ РЕЗУЛЬТАТОВ
	t.Logf("\n=== АНАЛИЗ ПРОГРЕССИИ КАДРОВ ===")
	t.Logf("Кадры показанные в порядке: %v", framesSeenInOrder)

	// КРИТИЧЕСКИЕ ПРОВЕРКИ
	frame0Seen := false
	frame1Seen := false

	for _, frame := range framesSeenInOrder {
		if frame == 0 {
			frame0Seen = true
		}
		if frame == 1 {
			frame1Seen = true
		}
	}

	if !frame0Seen {
		t.Error("❌ КРИТИЧЕСКАЯ ОШИБКА: Кадр 0 (первый кадр) НЕ ПОКАЗАН!")
	} else {
		t.Logf("✅ Кадр 0 показан")
	}

	if !frame1Seen {
		t.Error("❌ КРИТИЧЕСКАЯ ОШИБКА: Кадр 1 (второй кадр) НЕ ПОКАЗАН!")
		t.Error("   ЭТО ИМЕННО ТА ПРОБЛЕМА О КОТОРОЙ ГОВОРИТ ПОЛЬЗОВАТЕЛЬ!")
	} else {
		t.Logf("✅ Кадр 1 показан")
	}

	// Проверяем правильную последовательность
	if len(framesSeenInOrder) >= 2 && framesSeenInOrder[0] == 0 && framesSeenInOrder[1] == 1 {
		t.Logf("✅ Правильная последовательность: 0 -> 1")
	} else if frame0Seen && frame1Seen {
		t.Logf("⚠️ Кадры показаны, но возможно в неправильном порядке: %v", framesSeenInOrder)
	}

	// Проверяем что было минимум 2 кадра
	if len(framesSeenInOrder) < 2 {
		t.Error("❌ ПРОБЛЕМА: Показан только 1 кадр вместо 2!")
		t.Error("   Пользователь видит только первый кадр анимации!")
	}
}

// isWolfAttackingSimple простая проверка атаки
func isWolfAttackingSimple(world *core.World, wolf core.EntityID) bool {
	hunger, hasHunger := world.GetHunger(wolf)
	if !hasHunger || hunger.Value > 60.0 {
		return false
	}

	pos, hasPos := world.GetPosition(wolf)
	if !hasPos {
		return false
	}

	nearestRabbit, foundRabbit := world.FindNearestByType(pos.X, pos.Y, 15.0, core.TypeRabbit)
	if !foundRabbit {
		return false
	}

	if world.HasComponent(nearestRabbit, core.MaskCorpse) {
		return false
	}

	rabbitPos, hasRabbitPos := world.GetPosition(nearestRabbit)
	if !hasRabbitPos {
		return false
	}

	distance := (pos.X-rabbitPos.X)*(pos.X-rabbitPos.X) + (pos.Y-rabbitPos.Y)*(pos.Y-rabbitPos.Y)
	return distance <= 12.0*12.0
}
