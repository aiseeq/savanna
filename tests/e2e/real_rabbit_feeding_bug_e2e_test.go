package e2e

import (
	"fmt"
	"testing"

	"github.com/aiseeq/savanna/config"
	"github.com/aiseeq/savanna/internal/adapters"
	"github.com/aiseeq/savanna/internal/animation"
	"github.com/aiseeq/savanna/internal/core"
	"github.com/aiseeq/savanna/internal/generator"
	"github.com/aiseeq/savanna/internal/simulation"
)

// TestRealRabbitFeedingBugE2E проверяет РЕАЛЬНУЮ проблему: зайцы едят но сытость не восстанавливается
//
//nolint:gocognit,revive // E2E тест воспроизведения реального бага
func TestRealRabbitFeedingBugE2E(t *testing.T) {
	t.Parallel()

	t.Logf("=== E2E: РЕАЛЬНЫЙ БАГ - Зайцы едят но сытость только падает ===")
	t.Logf("ПРОБЛЕМА: Пользователь видит что зайцы показывают анимацию еды, но цифры голода продолжают падать")
	t.Logf("ОЖИДАНИЕ: Когда заяц ест траву, его голод должен УВЕЛИЧИВАТЬСЯ")

	// Создаём ТОЧНО ТАКУЮ ЖЕ конфигурацию как в реальной GUI игре
	cfg := config.LoadDefaultConfig()

	// Создаём полноценную игру точно как в cmd/game/main.go
	worldWidth := float32(cfg.World.Size * 32)
	worldHeight := float32(cfg.World.Size * 32)
	world := core.NewWorld(worldWidth, worldHeight, 12345)

	// Создаём terrain точно как в игре
	terrainGen := generator.NewTerrainGenerator(cfg)
	terrain := terrainGen.Generate()

	// Создаём GameWorld ТОЧНО как в реальной игре (game_world.go)
	systemManager := core.NewSystemManager()

	// Создаём системы ТОЧНО как в реальной игре (game_world.go)
	vegetationSystem := simulation.NewVegetationSystem(terrain)

	// НОВЫЕ СИСТЕМЫ (следуют принципу SRP):
	hungerSystem := simulation.NewHungerSystem()                           // 1. Только управление голодом
	grassSearchSystem := simulation.NewGrassSearchSystem(vegetationSystem) // 2. Только поиск травы и создание EatingState
	hungerSpeedModifier := simulation.NewHungerSpeedModifierSystem()       // 3. Только влияние голода на скорость
	starvationDamage := simulation.NewStarvationDamageSystem()             // 4. Только урон от голода

	grassEatingSystem := NewDebugGrassEatingSystem(simulation.NewGrassEatingSystem(vegetationSystem))
	animalBehaviorSystem := simulation.NewAnimalBehaviorSystem(vegetationSystem)
	movementSystem := simulation.NewMovementSystem(worldWidth, worldHeight)
	combatSystem := simulation.NewCombatSystem()

	// Добавляем системы в правильном порядке (КРИТИЧЕСКИ ВАЖЕН ДЛЯ ПИТАНИЯ!)
	systemManager.AddSystem(vegetationSystem)              // 1. Рост травы
	systemManager.AddSystem(&adapters.HungerSystemAdapter{ // 2. Управление голодом
		System: hungerSystem,
	})
	systemManager.AddSystem(&adapters.GrassSearchSystemAdapter{ // 3. Создание EatingState
		System: grassSearchSystem,
	})
	systemManager.AddSystem(grassEatingSystem) // 4. Дискретное поедание травы
	// 5. Поведение (проверяет EatingState)
	systemManager.AddSystem(&adapters.BehaviorSystemAdapter{System: animalBehaviorSystem})
	systemManager.AddSystem(&adapters.HungerSpeedModifierSystemAdapter{ // 6. Влияние голода на скорость
		System: hungerSpeedModifier,
	})
	// 7. Движение (сбрасывает скорость едящих)
	systemManager.AddSystem(&adapters.MovementSystemAdapter{System: movementSystem})
	systemManager.AddSystem(combatSystem)                            // 8. Система боя
	systemManager.AddSystem(&adapters.StarvationDamageSystemAdapter{ // 9. Урон от голода
		System: starvationDamage,
	})

	t.Logf("Добавлено систем в systemManager: 9 (новая архитектура с разделёнными системами)")

	// КРИТИЧЕСКИ ВАЖНО: Создаём AnimationManager как в реальной игре
	animationManager := createTestAnimationManager()

	// Загружаем анимации точно как в игре
	if err := animationManager.LoadAnimationsFromConfig(); err != nil {
		t.Errorf("❌ Ошибка загрузки анимаций: %v", err)
		return
	}

	t.Logf("\n=== Анализ реальной игры ===")
	t.Logf("Создали ВСЕ системы точно как в cmd/game/game_world.go")

	// Создаём зайца на позиции с большим количеством травы
	rabbit := simulation.CreateAnimal(world, core.TypeRabbit, 200, 200)

	// Устанавливаем траву под зайцем
	tileX := int(200 / 32)
	tileY := int(200 / 32)
	terrain.SetTileType(tileX, tileY, generator.TileGrass)
	terrain.SetGrassAmount(tileX, tileY, 100.0) // Много травы

	// Делаем зайца голодным чтобы он точно хотел есть
	initialHunger := float32(30.0) // Очень голодный
	world.SetHunger(rabbit, core.Hunger{Value: initialHunger})
	world.SetVelocity(rabbit, core.Velocity{X: 0, Y: 0}) // Стоит на месте

	// Проверяем начальное состояние
	pos, _ := world.GetPosition(rabbit)
	hunger, _ := world.GetHunger(rabbit)

	// КРИТИЧЕСКИЙ ТЕСТ: проверяем есть ли ВООБЩЕ системы которые могут восстанавливать голод
	t.Logf("\n=== Проверяем доступные системы для восстановления голода ===")

	// Попробуем найти системы поедания травы
	grassAmount := vegetationSystem.GetGrassAt(pos.X, pos.Y)

	t.Logf("Начальное состояние:")
	t.Logf("  Позиция зайца: (%.1f, %.1f)", pos.X, pos.Y)
	t.Logf("  Голод зайца: %.1f%%", hunger.Value)
	t.Logf("  Трава в позиции: %.1f единиц", grassAmount)

	if grassAmount < 50.0 {
		t.Errorf("❌ Недостаточно травы для теста: %.1f < 50.0", grassAmount)
		return
	}

	// ГЛАВНЫЙ ТЕСТ: Симулируем то что видит пользователь
	t.Logf("\n=== Симуляция ТОЧНО как видит пользователь ===")

	// ИСПРАВЛЕНИЕ: FeedingSystem в реальной игре может не сработать из-за различных условий
	// Принудительно создаём EatingState чтобы проверить что будет происходить когда заяц ест
	t.Logf("Принудительно создаём EatingState для зайца (имитируем работу FeedingSystem)")
	world.AddEatingState(rabbit, core.EatingState{
		Target:          0,                      // Трава не имеет entity ID
		TargetType:      core.EatingTargetGrass, // Тип: поедание травы
		EatingProgress:  0.0,
		NutritionGained: 0.0,
	})

	deltaTime := float32(1.0 / 60.0) // 60 FPS
	maxTicks := 300                  // 5 секунд симуляции

	t.Logf("Симулируем 5 секунд еды...")

	hungerHistory := []float32{hunger.Value} // История изменения голода
	grassHistory := []float32{grassAmount}   // История потребления травы

	for tick := 0; tick < maxTicks; tick++ {
		// Обновляем ТОЧНО как в реальной игре (game_world.go:49-54)
		world.Update(deltaTime)

		// ИСПРАВЛЕНИЕ: Анимации должны обновляться ПЕРЕД системами
		// чтобы GrassEatingSystem видел актуальные значения таймера
		animationManager.UpdateAnimalAnimations(world, deltaTime)

		systemManager.Update(world, deltaTime)

		// ТЕСТ: проверяем анимацию каждые 3 тика, особенно в момент завершения кадра (15 тиков)
		if tick%3 == 2 {
			currentHunger, _ := world.GetHunger(rabbit)
			currentGrass := vegetationSystem.GetGrassAt(pos.X, pos.Y)

			// Детальная отладка анимации и GrassEatingSystem
			anim, hasAnim := world.GetAnimation(rabbit)
			var animInfo string
			if hasAnim {
				// Получаем анимационную систему для проверки
				animalType, _ := world.GetAnimalType(rabbit)
				animSystem, exists := animationManager.animalSystems[animalType]
				shouldTrigger := anim.Timer >= 0.20 // Ловим момент близкий к завершению кадра

				animInfo = fmt.Sprintf("анимация=%d, таймер=%.3f, готов=%v, система=%v",
					anim.CurrentAnim, anim.Timer, shouldTrigger, exists)

				// ДОПОЛНИТЕЛЬНАЯ ДИАГНОСТИКА: проверяем работу анимационной системы
				if exists && animSystem != nil {
					t.Logf("        [ДИАГНОСТИКА] Анимационная система для %v существует", animalType)

					// Проверяем что анимация AnimEat зарегистрирована
					animData := animSystem.GetAnimation(animation.AnimEat)
					if animData != nil {
						t.Logf("        [ДИАГНОСТИКА] AnimEat: кадры=%d, FPS=%.1f, цикл=%v",
							animData.Frames, animData.FPS, animData.Loop)
						t.Logf("        [ДИАГНОСТИКА] Время кадра: %.3f сек", 1.0/animData.FPS)
					} else {
						t.Logf("        [ПРОБЛЕМА] AnimEat НЕ ЗАРЕГИСТРИРОВАНА!")
					}
				} else {
					t.Logf("        [ПРОБЛЕМА] Анимационная система для %v НЕ НАЙДЕНА!", animalType)
				}
			} else {
				animInfo = "нет анимации"
			}

			var eatingInfo string
			if eatingState, hasEatingState := world.GetEatingState(rabbit); hasEatingState {
				eatingInfo = fmt.Sprintf("цель=%d, прогресс=%.2f, питательность=%.2f",
					eatingState.Target, eatingState.EatingProgress, eatingState.NutritionGained)
			} else {
				eatingInfo = "нет EatingState"
			}

			frame := (tick + 1) / 15
			t.Logf("Кадр %d: голод=%.1f%%, трава=%.1f, %s",
				frame, currentHunger.Value, currentGrass, animInfo)
			t.Logf("        EatingState: %s", eatingInfo)

			// Сохраняем только каждую секунду в историю
			if tick%60 == 59 {
				hungerHistory = append(hungerHistory, currentHunger.Value)
				grassHistory = append(grassHistory, currentGrass)
			}
		}
	}

	// Анализируем результаты
	finalHunger, _ := world.GetHunger(rabbit)
	finalGrass := vegetationSystem.GetGrassAt(pos.X, pos.Y)

	t.Logf("\n=== АНАЛИЗ РЕЗУЛЬТАТОВ ===")
	t.Logf("История голода: %v", hungerHistory)
	t.Logf("История травы: %v", grassHistory)

	// КРИТИЧЕСКАЯ ПРОВЕРКА: Голод должен УВЕЛИЧИВАТЬСЯ при еде
	hungerChange := finalHunger.Value - initialHunger
	grassChange := grassAmount - finalGrass

	t.Logf("Изменение голода: %.1f%% (%.1f%% -> %.1f%%)",
		hungerChange, initialHunger, finalHunger.Value)
	t.Logf("Потребление травы: %.1f единиц (%.1f -> %.1f)",
		grassChange, grassAmount, finalGrass)

	// ГЛАВНАЯ ПРОВЕРКА: Если заяц ест, голод должен восстанавливаться
	if hungerChange <= 0 {
		t.Errorf("❌ БАГ ПОДТВЕРЖДЁН: Заяц ел %.1f единиц травы, но голод НЕ восстановился", grassChange)
		t.Errorf("   Голод изменился на %.1f%% (должен был УВЕЛИЧИТЬСЯ)", hungerChange)
		t.Errorf("   ЭТО ОБЪЯСНЯЕТ ПОЧЕМУ ПОЛЬЗОВАТЕЛЬ ВИДИТ ПАДАЮЩУЮ СЫТОСТЬ")

		// Дополнительный анализ
		if grassChange > 0 {
			t.Errorf("   ПАРАДОКС: Трава потребляется (%.1f единиц) но голод не восстанавливается", grassChange)
			t.Errorf("   ЗНАЧИТ: Системы поедания травы работают, но восстановления голода НЕТ")
		} else {
			t.Errorf("   ПРОБЛЕМА: Трава вообще не потребляется - системы поедания не работают")
		}

		return
	}

	// Если мы дошли сюда, значит голод восстанавливается (тест прошёл)
	t.Logf("✅ Голод восстанавливается корректно: +%.1f%%", hungerChange)
	t.Logf("✅ Потреблено травы: %.1f единиц", grassChange)
}

// TestAnimationManager упрощённая версия AnimationManager для тестов
// Копирует функциональность из cmd/game/animation_manager.go
type TestAnimationManager struct {
	animalSystems map[core.AnimalType]*animation.AnimationSystem
	resolver      *animation.AnimationResolver
}

// createTestAnimationManager создаёт AnimationManager для тестов
func createTestAnimationManager() *TestAnimationManager {
	return &TestAnimationManager{
		animalSystems: make(map[core.AnimalType]*animation.AnimationSystem),
		resolver:      animation.NewAnimationResolver(),
	}
}

// LoadAnimationsFromConfig загружает анимации (копия из cmd/game)
func (tam *TestAnimationManager) LoadAnimationsFromConfig() error {
	// Создаём и регистрируем систему анимаций для зайцев
	rabbitSystem := animation.NewAnimationSystem()
	rabbitAnimations := []struct {
		animType animation.AnimationType
		frames   int
		fps      float32
		loop     bool
	}{
		{animation.AnimIdle, 2, 2.0, true},
		{animation.AnimWalk, 2, 4.0, true},
		{animation.AnimRun, 2, 12.0, true},
		{animation.AnimAttack, 2, 5.0, false},
		{animation.AnimEat, 2, 4.0, true}, // КРИТИЧЕСКИ ВАЖНАЯ АНИМАЦИЯ!
		{animation.AnimDeathDying, 2, 3.0, false},
	}

	for _, config := range rabbitAnimations {
		rabbitSystem.RegisterAnimation(config.animType, config.frames, config.fps, config.loop, nil)
	}
	tam.animalSystems[core.TypeRabbit] = rabbitSystem

	// Создаём и регистрируем систему анимаций для волков
	wolfSystem := animation.NewAnimationSystem()
	wolfAnimations := []struct {
		animType animation.AnimationType
		frames   int
		fps      float32
		loop     bool
	}{
		{animation.AnimIdle, 2, 2.0, true},
		{animation.AnimWalk, 2, 4.0, true},
		{animation.AnimRun, 2, 8.0, true},
		{animation.AnimAttack, 4, 8.0, false},
		{animation.AnimEat, 2, 4.0, true},
		{animation.AnimDeathDying, 2, 3.0, false},
	}

	for _, config := range wolfAnimations {
		wolfSystem.RegisterAnimation(config.animType, config.frames, config.fps, config.loop, nil)
	}
	tam.animalSystems[core.TypeWolf] = wolfSystem

	return nil
}

// UpdateAnimalAnimations обновляет анимации всех животных (копия из cmd/game)
func (tam *TestAnimationManager) UpdateAnimalAnimations(world *core.World, deltaTime float32) {
	// Обходим всех животных с анимациями
	world.ForEachWith(core.MaskAnimalType|core.MaskAnimation, func(entity core.EntityID) {
		animalType, ok := world.GetAnimalType(entity)
		if !ok {
			return
		}

		// Получаем анимационную систему для этого типа животного
		animSystem, exists := tam.animalSystems[animalType]
		if !exists {
			return // Нет системы для этого типа - пропускаем
		}

		// Определяем нужный тип анимации
		expectedAnimType := tam.resolver.ResolveAnimalAnimationType(world, entity, animalType)

		// Обновляем анимацию если нужно
		tam.updateAnimationIfNeeded(world, entity, expectedAnimType)

		// Обновляем направление анимации на основе скорости
		tam.updateAnimationDirection(world, entity)

		// Обрабатываем кадры анимации
		tam.processAnimationUpdate(world, entity, animSystem, deltaTime)
	})
}

// updateAnimationIfNeeded обновляет тип анимации если он изменился
func (tam *TestAnimationManager) updateAnimationIfNeeded(
	world *core.World, entity core.EntityID, newAnimType animation.AnimationType,
) {
	anim, ok := world.GetAnimation(entity)
	if !ok {
		return
	}

	// Проверяем нужно ли менять анимацию
	if anim.CurrentAnim != int(newAnimType) {
		// НЕ прерываем анимацию ATTACK если она играет
		if anim.CurrentAnim == int(animation.AnimAttack) && anim.Playing {
			return
		}

		// Меняем анимацию
		anim.CurrentAnim = int(newAnimType)
		anim.Frame = 0
		anim.Timer = 0
		anim.Playing = true
		world.SetAnimation(entity, anim)
	}
}

// updateAnimationDirection обновляет направление анимации на основе скорости
func (tam *TestAnimationManager) updateAnimationDirection(world *core.World, entity core.EntityID) {
	anim, hasAnim := world.GetAnimation(entity)
	vel, hasVel := world.GetVelocity(entity)

	if !hasAnim || !hasVel {
		return
	}

	// Определяем направление по скорости
	if vel.X > 0.1 {
		anim.FacingRight = true
	} else if vel.X < -0.1 {
		anim.FacingRight = false
	}

	world.SetAnimation(entity, anim)
}

// processAnimationUpdate обрабатывает обновление кадров анимации
func (tam *TestAnimationManager) processAnimationUpdate(
	world *core.World, entity core.EntityID, animSystem *animation.AnimationSystem, deltaTime float32,
) {
	anim, ok := world.GetAnimation(entity)
	if !ok {
		return
	}

	// Создаём компонент для системы анимации
	animComponent := animation.AnimationComponent{
		CurrentAnim: animation.AnimationType(anim.CurrentAnim),
		Frame:       anim.Frame,
		Timer:       anim.Timer,
		Playing:     anim.Playing,
		FacingRight: anim.FacingRight,
	}

	// Обновляем через систему анимации
	animSystem.Update(&animComponent, deltaTime)

	// Сохраняем обновлённое состояние
	anim.Frame = animComponent.Frame
	anim.Timer = animComponent.Timer
	anim.Playing = animComponent.Playing
	anim.FacingRight = animComponent.FacingRight
	world.SetAnimation(entity, anim)
}

// DebugGrassEatingSystem wrapper для отладки GrassEatingSystem
type DebugGrassEatingSystem struct {
	inner core.System
}

func NewDebugGrassEatingSystem(inner core.System) *DebugGrassEatingSystem {
	return &DebugGrassEatingSystem{inner: inner}
}

func (dges *DebugGrassEatingSystem) Update(world *core.World, deltaTime float32) {
	// Детальная отладка каждого зайца с EatingState
	world.ForEachWith(core.MaskEatingState|core.MaskAnimalType, func(entity core.EntityID) {
		animalType, hasType := world.GetAnimalType(entity)
		if !hasType || animalType != core.TypeRabbit {
			return
		}

		eatingState, _ := world.GetEatingState(entity)
		if eatingState.Target != 0 { // Target = 0 означает поедание травы
			return
		}

		// Проверяем анимацию
		anim, hasAnim := world.GetAnimation(entity)
		if !hasAnim {
			fmt.Printf("[DEBUG] Заяц %d: НЕТ АНИМАЦИИ\n", entity)
			return
		}

		frameDuration := float32(1.0 / 4.0) // 0.25 секунды
		frameComplete := anim.Timer >= frameDuration

		// ПРОБЛЕМА: таймер сбрасывается ДО того как мы можем его проверить
		// Попробуем отследить когда таймер близок к завершению кадра
		almostComplete := anim.Timer >= 0.230 // 92% от 0.25

		fmt.Printf("[DEBUG] Заяц %d: anim=%d, timer=%.3f, frame=%d, почти_готов=%v, готов=%v\n",
			entity, anim.CurrentAnim, anim.Timer, anim.Frame, almostComplete, frameComplete)

		if frameComplete {
			fmt.Printf("[DEBUG] *** КАДР ЗАВЕРШЁН! Должно произойти поедание ***\n")
		} else if almostComplete {
			fmt.Printf("[DEBUG] +++ КАДР ПОЧТИ ЗАВЕРШЁН (%.3f/%.3f) +++\n", anim.Timer, frameDuration)
		}
	})

	// Вызываем настоящую систему
	dges.inner.Update(world, deltaTime)
}
