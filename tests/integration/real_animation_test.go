package integration

import (
	"testing"

	"github.com/aiseeq/savanna/internal/animation"
	"github.com/aiseeq/savanna/internal/core"
	"github.com/aiseeq/savanna/internal/simulation"
)

// TestRealAnimationBehavior тест РЕАЛЬНОГО поведения анимации как в игре
func TestRealAnimationBehavior(t *testing.T) {
	t.Parallel()
	// Создаём маленький мир 3x3 клетки (96x96 пикселей)
	world := core.NewWorld(96, 96, 42)

	// Создаём ТОЧНО такие же системы как в main.go
	combatSystem := simulation.NewCombatSystem()
	animalBehaviorSystem := simulation.NewAnimalBehaviorSystem(nil)
	movementSystem := simulation.NewMovementSystem(96, 96)

	// Создаём анимационные системы как в игре
	wolfAnimationSystem := animation.NewAnimationSystem()
	rabbitAnimationSystem := animation.NewAnimationSystem()

	// Регистрируем анимации ТОЧНО как в main.go
	wolfAnimationSystem.RegisterAnimation(animation.AnimIdle, 2, 1.0, true, nil)
	wolfAnimationSystem.RegisterAnimation(animation.AnimWalk, 4, 8.0, true, nil)
	wolfAnimationSystem.RegisterAnimation(animation.AnimRun, 4, 12.0, true, nil)
	wolfAnimationSystem.RegisterAnimation(animation.AnimAttack, 2, 6.0, false, nil) // НЕ зацикленная!
	wolfAnimationSystem.RegisterAnimation(animation.AnimEat, 2, 2.0, true, nil)

	rabbitAnimationSystem.RegisterAnimation(animation.AnimIdle, 2, 1.0, true, nil)
	rabbitAnimationSystem.RegisterAnimation(animation.AnimWalk, 4, 8.0, true, nil)
	rabbitAnimationSystem.RegisterAnimation(animation.AnimRun, 4, 12.0, true, nil)
	rabbitAnimationSystem.RegisterAnimation(animation.AnimDeathDying, 1, 1.0, false, nil)

	// Создаём животных В ЦЕНТРЕ маленькой карты
	rabbit := simulation.CreateRabbit(world, 40, 48) // Центр
	wolf := simulation.CreateWolf(world, 56, 48)     // Рядом с зайцем, на расстоянии 16 пикселей

	// Делаем волка ОЧЕНЬ голодным чтобы он точно атаковал
	world.SetHunger(wolf, core.Hunger{Value: 5.0})

	t.Logf("=== ТЕСТ РЕАЛЬНОГО ПОВЕДЕНИЯ АНИМАЦИИ ===")
	t.Logf("Карта: 96x96, заяц: (40,48), волк: (56,48), расстояние: 16")

	deltaTime := float32(1.0 / 60.0)

	// Функция обновления анимаций КАК В РЕАЛЬНОЙ ИГРЕ
	updateAnimations := func() {
		// Обходим всех животных и обновляем их анимации
		world.ForEachWith(core.MaskAnimalType|core.MaskAnimation, func(entity core.EntityID) {
			animalType, ok := world.GetAnimalType(entity)
			if !ok {
				return
			}

			anim, hasAnim := world.GetAnimation(entity)
			if !hasAnim {
				return
			}

			// Определяем тип анимации и систему на основе типа животного
			var newAnimType animation.AnimationType
			var animSystem *animation.AnimationSystem

			switch animalType {
			case core.TypeWolf:
				newAnimType = getWolfAnimationTypeReal(world, entity)
				animSystem = wolfAnimationSystem
			case core.TypeRabbit:
				newAnimType = getRabbitAnimationTypeReal(world, entity)
				animSystem = rabbitAnimationSystem
			default:
				return
			}

			// КРИТИЧЕСКИ ВАЖНО: НЕ прерываем анимацию ATTACK пока она играет!
			if anim.CurrentAnim != int(newAnimType) {
				if anim.CurrentAnim == int(animation.AnimAttack) && anim.Playing {
					// Анимация атаки должна доиграться до конца
					// НЕ меняем анимацию!
					t.Logf("  [ANIM] Entity %d: НЕ сбрасываем ATTACK анимацию (кадр %d)", entity, anim.Frame)
				} else {
					// Обычная смена анимации
					t.Logf("  [ANIM] Entity %d: %s -> %s", entity, animation.AnimationType(anim.CurrentAnim).String(), newAnimType.String())
					anim.CurrentAnim = int(newAnimType)
					anim.Frame = 0
					anim.Timer = 0
					anim.Playing = true
					world.SetAnimation(entity, anim)
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
			animSystem.Update(&animComponent, deltaTime)

			// Логируем изменения кадров
			if oldFrame != animComponent.Frame {
				t.Logf("  [FRAME] Entity %d (%s): кадр %d -> %d, играет: %t",
					entity, animalType.String(), oldFrame, animComponent.Frame, animComponent.Playing)
			}

			// Сохраняем состояние
			anim.Frame = animComponent.Frame
			anim.Timer = animComponent.Timer
			anim.Playing = animComponent.Playing
			world.SetAnimation(entity, anim)
		})
	}

	// Отслеживание событий
	lastRabbitHealth := int16(50)
	lastWolfHunger := float32(5.0)
	attackFramesSeen := make(map[int]bool)
	damageEvents := 0

	// Симулируем 600 тиков (10 секунд)
	for tick := 0; tick < 600; tick++ {
		// Обновляем мир
		world.Update(deltaTime)

		// Обновляем анимации КАК В ИГРЕ
		updateAnimations()

		// Обновляем системы
		animalBehaviorSystem.Update(world, deltaTime)
		movementSystem.Update(world, deltaTime)
		combatSystem.Update(world, deltaTime)

		// ДЕТАЛЬНОЕ ЛОГИРОВАНИЕ СОБЫТИЙ

		// Отслеживаем анимации волка
		if wolfAnim, hasAnim := world.GetAnimation(wolf); hasAnim {
			if wolfAnim.CurrentAnim == int(animation.AnimAttack) {
				attackFramesSeen[wolfAnim.Frame] = true

				if tick%5 == 0 { // Каждые 5 тиков
					t.Logf("[TICK %3d] ВОЛК АТАКУЕТ: кадр %d, играет: %t", tick, wolfAnim.Frame, wolfAnim.Playing)
				}
			}
		}

		// Отслеживаем урон
		if rabbitHealth, hasHealth := world.GetHealth(rabbit); hasHealth {
			if rabbitHealth.Current != lastRabbitHealth {
				damageEvents++
				t.Logf("[TICK %3d] 🩸 УРОН #%d: %d -> %d", tick, damageEvents, lastRabbitHealth, rabbitHealth.Current)

				// Проверяем DamageFlash
				if world.HasComponent(rabbit, core.MaskDamageFlash) {
					flash, _ := world.GetDamageFlash(rabbit)
					t.Logf("[TICK %3d] ✅ DamageFlash активен: %.3f сек", tick, flash.Timer)
				} else {
					t.Logf("[TICK %3d] ❌ DamageFlash НЕ АКТИВЕН!", tick)
				}

				lastRabbitHealth = rabbitHealth.Current
			}

			// Если заяц умер
			if rabbitHealth.Current == 0 && !world.HasComponent(rabbit, core.MaskCorpse) {
				// Ждём один тик на создание трупа
			} else if rabbitHealth.Current == 0 && world.HasComponent(rabbit, core.MaskCorpse) {
				corpse, _ := world.GetCorpse(rabbit)
				t.Logf("[TICK %3d] 💀 ЗАЯЦ СТАЛ ТРУПОМ: питательность %.1f", tick, corpse.NutritionalValue)

				// Проверяем начало поедания
				if world.HasComponent(wolf, core.MaskEatingState) {
					eating, _ := world.GetEatingState(wolf)
					t.Logf("[TICK %3d] 🍖 ВОЛК НАЧАЛ ЕСТЬ: цель %d", tick, eating.Target)
				}
				break
			}
		}

		// Отслеживаем голод волка
		if wolfHunger, hasHunger := world.GetHunger(wolf); hasHunger {
			if wolfHunger.Value != lastWolfHunger {
				t.Logf("[TICK %3d] 🐺 Голод волка: %.1f%% -> %.1f%%", tick, lastWolfHunger, wolfHunger.Value)
				lastWolfHunger = wolfHunger.Value
			}
		}

		// Логируем позиции каждые 30 тиков
		if tick%30 == 0 {
			wolfPos, _ := world.GetPosition(wolf)
			rabbitPos, _ := world.GetPosition(rabbit)
			distance := (wolfPos.X-rabbitPos.X)*(wolfPos.X-rabbitPos.X) + (wolfPos.Y-rabbitPos.Y)*(wolfPos.Y-rabbitPos.Y)
			t.Logf("[TICK %3d] Позиции: волк(%.1f,%.1f) заяц(%.1f,%.1f) дист=%.1f",
				tick, wolfPos.X, wolfPos.Y, rabbitPos.X, rabbitPos.Y, distance)
		}
	}

	// АНАЛИЗ РЕЗУЛЬТАТОВ
	t.Logf("\n=== АНАЛИЗ РЕЗУЛЬТАТОВ ===")
	t.Logf("Кадры анимации ATTACK которые были показаны:")
	for frame := 0; frame <= 1; frame++ {
		if attackFramesSeen[frame] {
			t.Logf("  ✅ Кадр %d: ПОКАЗАН", frame)
		} else {
			t.Logf("  ❌ Кадр %d: НЕ ПОКАЗАН", frame)
		}
	}

	t.Logf("Всего событий урона: %d", damageEvents)

	// ПРОВЕРКИ
	if !attackFramesSeen[0] {
		t.Error("❌ ОШИБКА: Кадр 0 анимации атаки НЕ ПОКАЗАН!")
	}

	if !attackFramesSeen[1] {
		t.Error("❌ ОШИБКА: Кадр 1 анимации атаки НЕ ПОКАЗАН!")
	}

	if damageEvents == 0 {
		t.Error("❌ ОШИБКА: Урон не был нанесен!")
	}

	if damageEvents > 0 && (!attackFramesSeen[0] || !attackFramesSeen[1]) {
		t.Error("❌ КРИТИЧЕСКАЯ ОШИБКА: Урон есть, но анимация атаки неполная!")
	}
}

// getWolfAnimationTypeReal определяет тип анимации для волка ТОЧНО как в main.go
func getWolfAnimationTypeReal(world *core.World, entity core.EntityID) animation.AnimationType {
	// ПРИОРИТЕТ 1: Если волк ест - показываем анимацию еды
	if world.HasComponent(entity, core.MaskEatingState) {
		return animation.AnimEat
	}

	// ПРИОРИТЕТ 2: Если волк атакует - показываем анимацию атаки
	if isWolfAttackingInTest(world, entity) {
		return animation.AnimAttack
	}

	// ПРИОРИТЕТ 3: Движение
	velocity, hasVel := world.GetVelocity(entity)
	if !hasVel {
		return animation.AnimIdle
	}

	speed := velocity.X*velocity.X + velocity.Y*velocity.Y

	if speed < 0.1 {
		return animation.AnimIdle
	} else if speed < 400.0 { // Примерно скорость ходьбы (20^2)
		return animation.AnimWalk
	} else {
		return animation.AnimRun
	}
}

// getRabbitAnimationTypeReal определяет тип анимации для зайца ТОЧНО как в main.go
func getRabbitAnimationTypeReal(world *core.World, entity core.EntityID) animation.AnimationType {
	// ПРИОРИТЕТ 1: Проверяем, является ли заяц трупом
	if world.HasComponent(entity, core.MaskCorpse) {
		return animation.AnimDeathDying
	}

	// ПРИОРИТЕТ 2: Движение
	velocity, hasVel := world.GetVelocity(entity)
	if !hasVel {
		return animation.AnimIdle
	}

	speed := velocity.X*velocity.X + velocity.Y*velocity.Y

	if speed < 0.1 {
		return animation.AnimIdle
	} else if speed < 300.0 { // Примерно скорость ходьбы зайца
		return animation.AnimWalk
	} else {
		return animation.AnimRun // Быстрое движение
	}
}

// isWolfAttackingInTest проверяет, атакует ли волк ТОЧНО как в main.go
func isWolfAttackingInTest(world *core.World, wolf core.EntityID) bool {
	// Сначала проверяем голод волка - сытый волк не атакует
	hunger, hasHunger := world.GetHunger(wolf)
	if !hasHunger || hunger.Value > 60.0 {
		return false
	}

	pos, hasPos := world.GetPosition(wolf)
	if !hasPos {
		return false
	}

	// Ищем ближайшего зайца в радиусе атаки
	nearestRabbit, foundRabbit := world.FindNearestByType(pos.X, pos.Y, 15.0, core.TypeRabbit)
	if !foundRabbit {
		return false
	}

	// Не атакуем трупы
	if world.HasComponent(nearestRabbit, core.MaskCorpse) {
		return false
	}

	// Проверяем, достаточно ли близко заяц для атаки
	rabbitPos, hasRabbitPos := world.GetPosition(nearestRabbit)
	if !hasRabbitPos {
		return false
	}

	distance := (pos.X-rabbitPos.X)*(pos.X-rabbitPos.X) + (pos.Y-rabbitPos.Y)*(pos.Y-rabbitPos.Y)
	return distance <= 12.0*12.0 // Дистанция атаки = 12 пикселей
}
