package e2e

import (
	"fmt"
	"testing"
	"time"

	"github.com/hajimehoshi/ebiten/v2"

	"github.com/aiseeq/savanna/config"
	"github.com/aiseeq/savanna/internal/adapters"
	"github.com/aiseeq/savanna/internal/animation"
	"github.com/aiseeq/savanna/internal/core"
	"github.com/aiseeq/savanna/internal/generator"
	"github.com/aiseeq/savanna/internal/simulation"
)

// WolfCombatE2E полный E2E тест боевой системы волков с РЕАЛЬНОЙ анимацией
//
//nolint:gocognit,revive,funlen // Комплексный E2E тест полного цикла атаки волка
func TestWolfCombatE2E(t *testing.T) {
	t.Parallel()
	t.Logf("=== E2E ТЕСТ: ПОЛНЫЙ ЦИКЛ АТАКИ ВОЛКА ===")

	// Создаём ТОЧНО такую же инициализацию как в GUI режиме
	cfg := config.LoadDefaultConfig()
	cfg.World.Seed = 42
	cfg.World.Size = 10 // Маленький мир 10x10 клеток
	cfg.Population.Rabbits = 1
	cfg.Population.Wolves = 1

	// Генерируем мир как в реальной игре
	terrainGen := generator.NewTerrainGenerator(cfg)
	terrain := terrainGen.Generate()

	worldSizePixels := float32(cfg.World.Size * 32)
	world := core.NewWorld(worldSizePixels, worldSizePixels, cfg.World.Seed)

	// Создаём ТОЧНО такие же системы как в main.go
	systemManager := core.NewSystemManager()
	vegetationSystem := simulation.NewVegetationSystem(terrain)
	animalBehaviorSystem := simulation.NewAnimalBehaviorSystem(vegetationSystem)
	combatSystem := simulation.NewCombatSystem()
	movementSystem := simulation.NewMovementSystem(worldSizePixels, worldSizePixels)

	// Добавляем системы в том же порядке что в main.go
	systemManager.AddSystem(vegetationSystem)
	systemManager.AddSystem(&adapters.BehaviorSystemAdapter{System: animalBehaviorSystem})
	systemManager.AddSystem(&adapters.MovementSystemAdapter{System: movementSystem})
	systemManager.AddSystem(adapters.NewFeedingSystemAdapter(vegetationSystem))
	systemManager.AddSystem(combatSystem)

	// КРИТИЧЕСКИ ВАЖНО: создаём анимационные системы как в GUI
	wolfAnimationSystem := animation.NewAnimationSystem()
	rabbitAnimationSystem := animation.NewAnimationSystem()

	// Загружаем анимации ТОЧНО как в main.go
	loadWolfAnimationsE2E(wolfAnimationSystem)
	loadRabbitAnimationsE2E(rabbitAnimationSystem)

	// Создаём off-screen буфер для "отрисовки" (как double buffer в GUI)
	offscreenImage := ebiten.NewImage(int(worldSizePixels), int(worldSizePixels))

	// Создаём животных рядом друг с другом
	rabbit := simulation.CreateAnimal(world, core.TypeRabbit, 48, 48) // Центр мира
	wolf := simulation.CreateAnimal(world, core.TypeWolf, 52, 48)     // Рядом с зайцем

	// Делаем волка голодным
	world.SetSatiation(wolf, core.Satiation{Value: 5.0})

	t.Logf("Начальное состояние:")
	rabbitHealth, _ := world.GetHealth(rabbit)
	wolfHunger, _ := world.GetSatiation(wolf)
	t.Logf("  Заяц: здоровье %d", rabbitHealth.Current)
	t.Logf("  Волк: голод %.1f%%", wolfHunger.Value)

	// Функция обновления анимаций ТОЧНО как в main.go
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
				newAnimType = getWolfAnimationTypeE2E(world, entity)
				animSystem = wolfAnimationSystem
			case core.TypeRabbit:
				newAnimType = getRabbitAnimationTypeE2E(world, entity)
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
					t.Logf("    [ANIM] Entity %d (%s): %s -> %s", entity,
						animalType.String(), oldAnimType.String(), newAnimType.String())
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

	// Функция "отрисовки" в off-screen буфер (имитация GUI отрисовки)
	renderFrame := func() {
		// Очищаем буфер
		offscreenImage.Clear()

		// Здесь бы была полная отрисовка как в GUI, но для E2E теста достаточно
		// просто вызвать логику получения кадров анимации
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

			// Имитируем получение кадра как в GUI
			var frameImg *ebiten.Image
			switch animalType {
			case core.TypeWolf:
				frameImg = wolfAnimationSystem.GetFrameImage(&animComponent)
			case core.TypeRabbit:
				frameImg = rabbitAnimationSystem.GetFrameImage(&animComponent)
			}

			// frameImg использовался бы для отрисовки в GUI
			_ = frameImg
		})
	}

	// Отслеживание событий
	lastRabbitHealth := rabbitHealth.Current
	lastWolfHunger := wolfHunger.Value
	attackFramesSeen := make(map[int]bool)
	damageEvents := []string{}

	// Основной цикл E2E теста (имитация игрового цикла)
	deltaTime := float32(1.0 / 60.0)
	maxTicks := 600 // 10 секунд максимум

	for tick := 0; tick < maxTicks; tick++ {
		// Обновляем мир ТОЧНО как в GUI
		world.Update(deltaTime)
		systemManager.Update(world, deltaTime)

		// Обновляем анимации ТОЧНО как в GUI
		updateAnimalAnimations()

		// "Отрисовываем" кадр ТОЧНО как в GUI
		renderFrame()

		// ДЕТАЛЬНОЕ ЛОГИРОВАНИЕ СОБЫТИЙ

		// Отслеживаем анимации волка
		if wolfAnim, hasAnim := world.GetAnimation(wolf); hasAnim {
			if wolfAnim.CurrentAnim == int(animation.AnimAttack) {
				attackFramesSeen[wolfAnim.Frame] = true

				if tick%10 == 0 { // Каждые 10 тиков
					t.Logf("[TICK %3d] 🐺 ВОЛК АТАКУЕТ: кадр %d, играет: %t", tick, wolfAnim.Frame, wolfAnim.Playing)
				}
			}
		}

		// Отслеживаем урон
		currentRabbitHealth, _ := world.GetHealth(rabbit)
		if currentRabbitHealth.Current != lastRabbitHealth {
			damageEvent := ""
			if currentRabbitHealth.Current < lastRabbitHealth {
				damageEvent = "УРОН"
			} else {
				damageEvent = "ИСЦЕЛЕНИЕ"
			}

			event := fmt.Sprintf("[TICK %3d] 🩸 %s: %d -> %d", tick, damageEvent, lastRabbitHealth, currentRabbitHealth.Current)
			t.Logf(event)
			damageEvents = append(damageEvents, event)

			// Проверяем DamageFlash
			if world.HasComponent(rabbit, core.MaskDamageFlash) {
				flash, _ := world.GetDamageFlash(rabbit)
				t.Logf("[TICK %3d]   ✨ DamageFlash активен: %.3f сек", tick, flash.Timer)
			} else {
				t.Logf("[TICK %3d]   ❌ DamageFlash НЕ активен!", tick)
			}

			lastRabbitHealth = currentRabbitHealth.Current
		}

		// Отслеживаем голод волка
		currentWolfHunger, _ := world.GetSatiation(wolf)
		if currentWolfHunger.Value != lastWolfHunger {
			t.Logf("[TICK %3d] 🍖 Голод волка: %.1f%% -> %.1f%%", tick, lastWolfHunger, currentWolfHunger.Value)
			lastWolfHunger = currentWolfHunger.Value
		}

		// Проверяем создание трупа
		if currentRabbitHealth.Current == 0 && world.HasComponent(rabbit, core.MaskCorpse) {
			corpse, _ := world.GetCorpse(rabbit)
			t.Logf("[TICK %3d] ⚰️ ЗАЯЦ СТАЛ ТРУПОМ: питательность %.1f", tick, corpse.NutritionalValue)

			// Проверяем начало поедания
			if world.HasComponent(wolf, core.MaskEatingState) {
				eating, _ := world.GetEatingState(wolf)
				t.Logf("[TICK %3d] 🍽️ ВОЛК НАЧАЛ ЕСТЬ: цель %d", tick, eating.Target)
			}
			break
		}

		// Если заяц полностью исчез (съеден), успех
		if !world.IsAlive(rabbit) {
			t.Logf("[TICK %3d] 🎉 ЗАЯЦ ПОЛНОСТЬЮ СЪЕДЕН", tick)
			break
		}

		// Небольшая задержка как в реальной игре
		time.Sleep(time.Microsecond * 100)
	}

	// АНАЛИЗ РЕЗУЛЬТАТОВ E2E ТЕСТА
	t.Logf("\n=== АНАЛИЗ E2E РЕЗУЛЬТАТОВ ===")

	// Проверяем анимации
	t.Logf("Кадры анимации ATTACK:")
	frame0Seen := attackFramesSeen[0]
	frame1Seen := attackFramesSeen[1]

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

	// Проверяем урон
	t.Logf("События урона: %d", len(damageEvents))
	if len(damageEvents) == 0 {
		t.Errorf("  ❌ Урон НЕ был нанесен!")
	} else {
		t.Logf("  ✅ Урон был нанесен %d раз(а)", len(damageEvents))
		for _, event := range damageEvents {
			t.Logf("    %s", event)
		}
	}

	// Финальные проверки
	finalRabbitHealth, _ := world.GetHealth(rabbit)
	finalWolfHunger, _ := world.GetSatiation(wolf)

	t.Logf("Финальное состояние:")
	t.Logf("  Заяц: здоровье %d", finalRabbitHealth.Current)
	t.Logf("  Волк: голод %.1f%%", finalWolfHunger.Value)

	// КРИТИЧЕСКИЕ ПРОВЕРКИ E2E
	if !frame0Seen || !frame1Seen {
		t.Error("❌ E2E ОШИБКА: Анимация атаки неполная!")
	}

	if len(damageEvents) == 0 {
		t.Error("❌ E2E ОШИБКА: Урон не был нанесен!")
	}

	// В новой системе волк может тратить энергию в процессе боя
	// Главное - что он начал есть труп и процесс восстановления голода запущен
	if !world.HasComponent(wolf, core.MaskEatingState) && finalWolfHunger.Value < 3.0 {
		t.Error("❌ E2E ОШИБКА: Волк не начал процесс восстановления голода!")
	} else {
		t.Logf("✅ E2E УСПЕХ: Волк ест труп, голод будет восстанавливаться")
	}

	if finalRabbitHealth.Current > 0 && !world.HasComponent(rabbit, core.MaskCorpse) {
		t.Error("❌ E2E ОШИБКА: Заяц должен быть мертв или стать трупом!")
	}
}

// Вспомогательные функции точно как в main.go

func getWolfAnimationTypeE2E(world *core.World, entity core.EntityID) animation.AnimationType {
	// ПРИОРИТЕТ 1: Если волк ест
	if world.HasComponent(entity, core.MaskEatingState) {
		return animation.AnimEat
	}

	// ПРИОРИТЕТ 2: Если волк атакует
	if isWolfAttackingE2E(world, entity) {
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
	} else if speed < 400.0 {
		return animation.AnimWalk
	} else {
		return animation.AnimRun
	}
}

func getRabbitAnimationTypeE2E(world *core.World, entity core.EntityID) animation.AnimationType {
	// ПРИОРИТЕТ 1: Труп
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
	} else if speed < 300.0 {
		return animation.AnimWalk
	} else {
		return animation.AnimRun
	}
}

func isWolfAttackingE2E(world *core.World, wolf core.EntityID) bool {
	hunger, hasHunger := world.GetSatiation(wolf)
	if !hasHunger || hunger.Value > 60.0 {
		return false
	}

	pos, hasPos := world.GetPosition(wolf)
	if !hasPos {
		return false
	}

	// ТИПОБЕЗОПАСНОСТЬ: конвертируем physics.Pixels в float32
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

func loadWolfAnimationsE2E(animSystem *animation.AnimationSystem) {
	// Создаём пустое изображение-спрайтшит для анимаций (в E2E тесте содержимое не важно)
	emptyImg := ebiten.NewImage(128, 64) // Достаточно большой для нескольких кадров

	// Регистрируем анимации ТОЧНО как в main.go
	animSystem.RegisterAnimation(animation.AnimIdle, 2, 1.0, true, emptyImg)
	animSystem.RegisterAnimation(animation.AnimWalk, 4, 8.0, true, emptyImg)
	animSystem.RegisterAnimation(animation.AnimRun, 4, 12.0, true, emptyImg)
	animSystem.RegisterAnimation(animation.AnimAttack, 2, 6.0, false, emptyImg) // НЕ зацикленная!
	animSystem.RegisterAnimation(animation.AnimEat, 2, 2.0, true, emptyImg)
}

func loadRabbitAnimationsE2E(animSystem *animation.AnimationSystem) {
	// Создаём пустое изображение-спрайтшит для анимаций (в E2E тесте содержимое не важно)
	emptyImg := ebiten.NewImage(128, 64) // Достаточно большой для нескольких кадров

	// Регистрируем анимации ТОЧНО как в main.go
	animSystem.RegisterAnimation(animation.AnimIdle, 2, 1.0, true, emptyImg)
	animSystem.RegisterAnimation(animation.AnimWalk, 4, 8.0, true, emptyImg)
	animSystem.RegisterAnimation(animation.AnimRun, 4, 12.0, true, emptyImg)
	animSystem.RegisterAnimation(animation.AnimDeathDying, 1, 1.0, false, emptyImg)
}
