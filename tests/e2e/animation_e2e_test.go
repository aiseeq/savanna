package e2e

import (
	"testing"

	"github.com/hajimehoshi/ebiten/v2"

	"github.com/aiseeq/savanna/internal/adapters"
	"github.com/aiseeq/savanna/internal/animation"
	"github.com/aiseeq/savanna/internal/core"
	"github.com/aiseeq/savanna/internal/simulation"
)

// TestAnimationE2E минимальный E2E тест анимаций без генерации ландшафта
func TestAnimationE2E(t *testing.T) {
	t.Parallel()
	t.Logf("=== E2E ТЕСТ: АНИМАЦИИ КАК В РЕАЛЬНОЙ ИГРЕ ===")

	// Создаём мир без генерации ландшафта
	world := core.NewWorld(320, 320, 42) // 10x10 тайлов по 32 пикселя

	// Создаём ТОЧНО такие же системы как в main.go
	systemManager := core.NewSystemManager()
	combatSystem := simulation.NewCombatSystem()
	animalBehaviorSystem := simulation.NewAnimalBehaviorSystem(nil) // без растительности
	movementSystem := simulation.NewMovementSystem(320, 320)

	// Добавляем системы в том же порядке что в main.go
	systemManager.AddSystem(&adapters.BehaviorSystemAdapter{System: animalBehaviorSystem})
	systemManager.AddSystem(&adapters.MovementSystemAdapter{System: movementSystem})
	systemManager.AddSystem(combatSystem)

	// КРИТИЧЕСКИ ВАЖНО: создаём анимационные системы как в GUI
	wolfAnimationSystem := animation.NewAnimationSystem()
	rabbitAnimationSystem := animation.NewAnimationSystem()

	// Загружаем анимации ТОЧНО как в main.go
	loadAnimationsForE2E(wolfAnimationSystem, rabbitAnimationSystem)

	// Создаём off-screen буфер для "отрисовки"
	offscreenImage := ebiten.NewImage(320, 320)

	// Создаём животных рядом друг с другом
	rabbit := simulation.CreateRabbit(world, 160, 160) // Центр
	wolf := simulation.CreateWolf(world, 164, 160)     // Рядом с зайцем (4 пикселя)

	// Делаем волка очень голодным
	world.SetHunger(wolf, core.Hunger{Value: 5.0})

	t.Logf("Начальное состояние:")
	rabbitHealth, _ := world.GetHealth(rabbit)
	wolfHunger, _ := world.GetHunger(wolf)
	t.Logf("  Заяц: здоровье %d, позиция (160,160)", rabbitHealth.Current)
	t.Logf("  Волк: голод %.1f%%, позиция (164,160)", wolfHunger.Value)

	// Функция обновления анимаций КАК В РЕАЛЬНОЙ ИГРЕ (main.go)
	updateAnimalAnimations := func() {
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
				newAnimType = getWolfAnimationTypeForE2E(world, entity)
				animSystem = wolfAnimationSystem
			case core.TypeRabbit:
				newAnimType = getRabbitAnimationTypeForE2E(world, entity)
				animSystem = rabbitAnimationSystem
			default:
				return
			}

			// КРИТИЧЕСКИ ВАЖНО: НЕ прерываем анимацию ATTACK пока она играет!
			oldAnimType := animation.AnimationType(anim.CurrentAnim)
			if anim.CurrentAnim != int(newAnimType) {
				if anim.CurrentAnim == int(animation.AnimAttack) && anim.Playing {
					t.Logf("    [ANIM] Entity %d: НЕ сбрасываем ATTACK анимацию (кадр %d)", entity, anim.Frame)
				} else {
					t.Logf("    [ANIM] Entity %d (%s): %s -> %s", entity, animalType.String(), oldAnimType.String(), newAnimType.String())
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
			oldPlaying := animComponent.Playing

			animSystem.Update(&animComponent, 1.0/60.0)

			// Логируем изменения кадров
			if oldFrame != animComponent.Frame || oldPlaying != animComponent.Playing {
				t.Logf("    [FRAME] Entity %d (%s): кадр %d->%d, играет %t->%t",
					entity, animalType.String(), oldFrame, animComponent.Frame, oldPlaying, animComponent.Playing)
			}

			// Сохраняем состояние
			anim.Frame = animComponent.Frame
			anim.Timer = animComponent.Timer
			anim.Playing = animComponent.Playing
			world.SetAnimation(entity, anim)
		})
	}

	// Функция "отрисовки" в off-screen буфер (имитация GUI)
	renderFrame := func() {
		offscreenImage.Clear()

		// Имитируем получение кадров анимации как в GUI
		world.ForEachWith(core.MaskAnimalType|core.MaskAnimation, func(entity core.EntityID) {
			animalType, _ := world.GetAnimalType(entity)
			anim, hasAnim := world.GetAnimation(entity)
			if !hasAnim {
				return
			}

			animComponent := animation.AnimationComponent{
				CurrentAnim: animation.AnimationType(anim.CurrentAnim),
				Frame:       anim.Frame,
				Timer:       anim.Timer,
				Playing:     anim.Playing,
				FacingRight: anim.FacingRight,
			}

			// Получаем кадр как в GUI
			var frameImg *ebiten.Image
			switch animalType {
			case core.TypeWolf:
				frameImg = wolfAnimationSystem.GetFrameImage(&animComponent)
			case core.TypeRabbit:
				frameImg = rabbitAnimationSystem.GetFrameImage(&animComponent)
			}

			// frameImg использовался бы для отрисовки
			_ = frameImg
		})
	}

	// Отслеживание событий
	lastRabbitHealth := rabbitHealth.Current
	attackFramesSeen := make(map[int]bool)
	damageEvents := 0

	// Основной игровой цикл E2E (имитация Update() из main.go)
	deltaTime := float32(1.0 / 60.0)

	for tick := 0; tick < 300; tick++ { // 5 секунд максимум
		// Обновляем мир ТОЧНО как в GUI
		world.Update(deltaTime)
		systemManager.Update(world, deltaTime)

		// Обновляем анимации ТОЧНО как в GUI
		updateAnimalAnimations()

		// "Отрисовываем" кадр ТОЧНО как в GUI
		renderFrame()

		// ДЕТАЛЬНОЕ ЛОГИРОВАНИЕ

		// Отслеживаем анимации волка
		if wolfAnim, hasAnim := world.GetAnimation(wolf); hasAnim {
			if wolfAnim.CurrentAnim == int(animation.AnimAttack) {
				attackFramesSeen[wolfAnim.Frame] = true

				if tick%5 == 0 { // Каждые 5 тиков
					t.Logf("[TICK %3d] 🐺 ВОЛК АТАКУЕТ: кадр %d, играет: %t, таймер: %.3f",
						tick, wolfAnim.Frame, wolfAnim.Playing, wolfAnim.Timer)
				}
			}
		}

		// Отслеживаем урон
		currentRabbitHealth, _ := world.GetHealth(rabbit)
		if currentRabbitHealth.Current != lastRabbitHealth {
			damageEvents++
			t.Logf("[TICK %3d] 🩸 УРОН #%d: %d -> %d", tick, damageEvents, lastRabbitHealth, currentRabbitHealth.Current)

			// Проверяем DamageFlash
			if world.HasComponent(rabbit, core.MaskDamageFlash) {
				flash, _ := world.GetDamageFlash(rabbit)
				t.Logf("[TICK %3d]   ✨ DamageFlash: %.3f сек", tick, flash.Timer)
			} else {
				t.Logf("[TICK %3d]   ❌ DamageFlash НЕ активен!", tick)
			}

			lastRabbitHealth = currentRabbitHealth.Current
		}

		// Если заяц умер, проверяем труп
		if currentRabbitHealth.Current == 0 {
			if world.HasComponent(rabbit, core.MaskCorpse) {
				corpse, _ := world.GetCorpse(rabbit)
				t.Logf("[TICK %3d] ⚰️ ЗАЯЦ СТАЛ ТРУПОМ: питательность %.1f", tick, corpse.NutritionalValue)

				if world.HasComponent(wolf, core.MaskEatingState) {
					t.Logf("[TICK %3d] 🍽️ ВОЛК НАЧАЛ ЕСТЬ", tick)
				}
				break
			}
		}

		// Если заяц исчез (съеден)
		if !world.IsAlive(rabbit) {
			t.Logf("[TICK %3d] 🎉 ЗАЯЦ ПОЛНОСТЬЮ ИСЧЕЗ (съеден)", tick)
			break
		}
	}

	// АНАЛИЗ E2E РЕЗУЛЬТАТОВ
	t.Logf("\n=== АНАЛИЗ E2E РЕЗУЛЬТАТОВ ===")

	// Проверяем анимации
	frame0Seen := attackFramesSeen[0]
	frame1Seen := attackFramesSeen[1]

	t.Logf("Кадры анимации ATTACK:")
	if frame0Seen {
		t.Logf("  ✅ Кадр 0 (замах): ПОКАЗАН")
	} else {
		t.Errorf("  ❌ Кадр 0 (замах): НЕ ПОКАЗАН")
	}

	if frame1Seen {
		t.Logf("  ✅ Кадр 1 (удар): ПОКАЗАН")
	} else {
		t.Errorf("  ❌ Кадр 1 (удар): НЕ ПОКАЗАН")
	}

	t.Logf("События урона: %d", damageEvents)
	if damageEvents == 0 {
		t.Errorf("  ❌ Урон НЕ был нанесен!")
	} else {
		t.Logf("  ✅ Урон был нанесен %d раз(а)", damageEvents)
	}

	// Финальные проверки
	finalRabbitHealth, _ := world.GetHealth(rabbit)
	finalWolfHunger, _ := world.GetHunger(wolf)

	t.Logf("Финальное состояние:")
	t.Logf("  Заяц: здоровье %d", finalRabbitHealth.Current)
	t.Logf("  Волк: голод %.1f%%", finalWolfHunger.Value)

	// КРИТИЧЕСКИЕ E2E ПРОВЕРКИ
	if !frame0Seen || !frame1Seen {
		t.Error("❌ E2E КРИТИЧЕСКАЯ ОШИБКА: Анимация атаки неполная - НЕ 2 кадра!")
	}

	if damageEvents == 0 {
		t.Error("❌ E2E КРИТИЧЕСКАЯ ОШИБКА: Урон не был нанесен!")
	}

	if finalRabbitHealth.Current > 0 && !world.HasComponent(rabbit, core.MaskCorpse) {
		t.Error("❌ E2E КРИТИЧЕСКАЯ ОШИБКА: Заяц должен быть мертв или стать трупом!")
	}

	t.Logf("\n🎯 E2E тест проверил РЕАЛЬНОЕ поведение анимаций как в GUI игре")
}

// Вспомогательные функции

func getWolfAnimationTypeForE2E(world *core.World, entity core.EntityID) animation.AnimationType {
	if world.HasComponent(entity, core.MaskEatingState) {
		return animation.AnimEat
	}

	if isWolfAttackingForE2E(world, entity) {
		return animation.AnimAttack
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

func getRabbitAnimationTypeForE2E(world *core.World, entity core.EntityID) animation.AnimationType {
	if world.HasComponent(entity, core.MaskCorpse) {
		return animation.AnimDeathDying
	}

	velocity, hasVel := world.GetVelocity(entity)
	if !hasVel {
		return animation.AnimIdle
	}

	speed := velocity.X*velocity.X + velocity.Y*velocity.Y
	if speed < 0.1 {
		return animation.AnimIdle
	} else if speed < 300.0 {
		return animation.AnimWalk
	} else {
		return animation.AnimRun
	}
}

func isWolfAttackingForE2E(world *core.World, wolf core.EntityID) bool {
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

func loadAnimationsForE2E(wolfAnimSystem, rabbitAnimSystem *animation.AnimationSystem) {
	// Создаём пустые спрайтшиты (содержимое не важно для E2E)
	emptySheet := ebiten.NewImage(128, 64)

	// Волк - ТОЧНО как в main.go
	wolfAnimSystem.RegisterAnimation(animation.AnimIdle, 2, 1.0, true, emptySheet)
	wolfAnimSystem.RegisterAnimation(animation.AnimWalk, 4, 8.0, true, emptySheet)
	wolfAnimSystem.RegisterAnimation(animation.AnimRun, 4, 12.0, true, emptySheet)
	wolfAnimSystem.RegisterAnimation(animation.AnimAttack, 2, 6.0, false, emptySheet) // НЕ зацикленная!
	wolfAnimSystem.RegisterAnimation(animation.AnimEat, 2, 2.0, true, emptySheet)

	// Заяц - ТОЧНО как в main.go
	rabbitAnimSystem.RegisterAnimation(animation.AnimIdle, 2, 1.0, true, emptySheet)
	rabbitAnimSystem.RegisterAnimation(animation.AnimWalk, 4, 8.0, true, emptySheet)
	rabbitAnimSystem.RegisterAnimation(animation.AnimRun, 4, 12.0, true, emptySheet)
	rabbitAnimSystem.RegisterAnimation(animation.AnimDeathDying, 1, 1.0, false, emptySheet)
}
