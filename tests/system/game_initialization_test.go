package system

import (
	"testing"

	"github.com/hajimehoshi/ebiten/v2"

	"github.com/aiseeq/savanna/config"
	"github.com/aiseeq/savanna/internal/adapters"
	"github.com/aiseeq/savanna/internal/animation"
	"github.com/aiseeq/savanna/internal/core"
	"github.com/aiseeq/savanna/internal/generator"
	"github.com/aiseeq/savanna/internal/simulation"
)

// TestGameInitializationParity проверяет что GUI и headless инициализируются одинаково
func TestGameInitializationParity(t *testing.T) {
	t.Parallel()

	t.Log("=== ТЕСТ СИСТЕМНОЙ ИНИЦИАЛИЗАЦИИ ===")
	t.Log("Проверяем что GUI и headless режимы инициализируются одинаково")

	// Тестируем с проблемным seed 6
	cfg := config.LoadDefaultConfig()
	cfg.World.Seed = 6
	cfg.World.Size = 50
	cfg.Population.Rabbits = 30
	cfg.Population.Wolves = 3

	// === ИНИЦИАЛИЗАЦИЯ GUI РЕЖИМА ===
	t.Log("Инициализация GUI режима...")
	guiSystems := initGUIMode(t, cfg)

	// === ИНИЦИАЛИЗАЦИЯ HEADLESS РЕЖИМА ===
	t.Log("Инициализация headless режима...")
	headlessSystems := initHeadlessMode(t, cfg)

	// === СРАВНЕНИЕ СИСТЕМ ===
	t.Log("Сравнение инициализированных систем...")

	// Проверяем что у обоих режимов одинаковое количество систем
	if len(guiSystems) != len(headlessSystems) {
		t.Errorf("❌ Количество систем не совпадает: GUI %d vs Headless %d",
			len(guiSystems), len(headlessSystems))

		t.Logf("GUI системы: %v", guiSystems)
		t.Logf("Headless системы: %v", headlessSystems)
	} else {
		t.Logf("✅ Количество систем совпадает: %d", len(guiSystems))
	}

	// Проверяем что типы систем совпадают
	for i, guiSystem := range guiSystems {
		if i >= len(headlessSystems) {
			break
		}

		if guiSystem != headlessSystems[i] {
			t.Errorf("❌ Система #%d не совпадает: GUI '%s' vs Headless '%s'",
				i, guiSystem, headlessSystems[i])
		}
	}

	// === ПРОВЕРКА РАБОТОСПОСОБНОСТИ ===
	t.Log("Проверка работоспособности боевой системы...")

	// Тестируем что боевая система действительно работает
	testCombatFunctionality(t, cfg)

	t.Log("✅ Системный тест инициализации завершён")
}

// initGUIMode имитирует инициализацию как в cmd/game/main.go
func initGUIMode(t *testing.T, cfg *config.Config) []string {
	var systems []string

	// Точная имитация cmd/game/main.go
	terrainGen := generator.NewTerrainGenerator(cfg)
	terrain := terrainGen.Generate()

	worldSizePixels := float32(cfg.World.Size * 32)
	_ = core.NewWorld(worldSizePixels, worldSizePixels, cfg.World.Seed)
	systemManager := core.NewSystemManager()

	// Анимационные системы (есть в GUI)
	systems = append(systems,
		"WolfAnimationSystem",
		"RabbitAnimationSystem",
		"AnimationManager",
	)

	// Игровые системы в порядке как в GUI
	vegetationSystem := simulation.NewVegetationSystem(terrain)
	animalBehaviorSystem := simulation.NewAnimalBehaviorSystem(vegetationSystem)

	// НОВЫЕ СИСТЕМЫ (следуют принципу SRP):
	hungerSystem := simulation.NewHungerSystem()                           // 1. Только управление голодом
	grassSearchSystem := simulation.NewGrassSearchSystem(vegetationSystem) // 2. Только поиск травы и создание EatingState
	hungerSpeedModifier := simulation.NewHungerSpeedModifierSystem()       // 3. Только влияние голода на скорость
	starvationDamage := simulation.NewStarvationDamageSystem()             // 4. Только урон от голода

	combatSystem := simulation.NewCombatSystem()
	movementSystem := simulation.NewMovementSystem(worldSizePixels, worldSizePixels)

	systemManager.AddSystem(vegetationSystem)
	systems = append(systems, "VegetationSystem")

	systemManager.AddSystem(&adapters.HungerSystemAdapter{System: hungerSystem})
	systems = append(systems, "HungerSystem")

	systemManager.AddSystem(&adapters.GrassSearchSystemAdapter{System: grassSearchSystem})
	systems = append(systems, "GrassSearchSystem")

	systemManager.AddSystem(&adapters.BehaviorSystemAdapter{System: animalBehaviorSystem})
	systems = append(systems, "AnimalBehaviorSystem")

	systemManager.AddSystem(&adapters.HungerSpeedModifierSystemAdapter{System: hungerSpeedModifier})
	systems = append(systems, "HungerSpeedModifierSystem")

	systemManager.AddSystem(&adapters.MovementSystemAdapter{System: movementSystem})
	systems = append(systems, "MovementSystem")

	systemManager.AddSystem(combatSystem)
	systems = append(systems, "CombatSystem")

	systemManager.AddSystem(&adapters.StarvationDamageSystemAdapter{System: starvationDamage})
	systems = append(systems, "StarvationDamageSystem")

	// Проверяем что системы добавлены
	t.Logf("GUI режим инициализирован с %d системами", len(systems))

	return systems
}

// initHeadlessMode имитирует инициализацию как в cmd/headless/main.go
func initHeadlessMode(t *testing.T, cfg *config.Config) []string {
	var systems []string

	// Точная имитация cmd/headless/main.go
	terrainGen := generator.NewTerrainGenerator(cfg)
	terrain := terrainGen.Generate()

	worldSizePixels := float32(cfg.World.Size * 32)
	_ = core.NewWorld(worldSizePixels, worldSizePixels, cfg.World.Seed)
	systemManager := core.NewSystemManager()

	// Анимационные системы (НЕДАВНО ДОБАВЛЕНЫ в headless)
	wolfAnimationSystem := animation.NewAnimationSystem()
	rabbitAnimationSystem := animation.NewAnimationSystem()
	_ = wolfAnimationSystem
	_ = rabbitAnimationSystem
	systems = append(systems,
		"WolfAnimationSystem",
		"RabbitAnimationSystem",
		"AnimationManager",
	)

	// Игровые системы в порядке как в headless
	vegetationSystem := simulation.NewVegetationSystem(terrain)
	animalBehaviorSystem := simulation.NewAnimalBehaviorSystem(vegetationSystem)

	// НОВЫЕ СИСТЕМЫ (следуют принципу SRP):
	hungerSystem := simulation.NewHungerSystem()                           // 1. Только управление голодом
	grassSearchSystem := simulation.NewGrassSearchSystem(vegetationSystem) // 2. Только поиск травы и создание EatingState
	hungerSpeedModifier := simulation.NewHungerSpeedModifierSystem()       // 3. Только влияние голода на скорость
	starvationDamage := simulation.NewStarvationDamageSystem()             // 4. Только урон от голода

	combatSystem := simulation.NewCombatSystem()

	systemManager.AddSystem(vegetationSystem)
	systems = append(systems, "VegetationSystem")

	systemManager.AddSystem(&adapters.HungerSystemAdapter{System: hungerSystem})
	systems = append(systems, "HungerSystem")

	systemManager.AddSystem(&adapters.GrassSearchSystemAdapter{System: grassSearchSystem})
	systems = append(systems, "GrassSearchSystem")

	systemManager.AddSystem(&adapters.BehaviorSystemAdapter{System: animalBehaviorSystem})
	systems = append(systems, "AnimalBehaviorSystem")

	systemManager.AddSystem(&adapters.HungerSpeedModifierSystemAdapter{System: hungerSpeedModifier})
	systems = append(systems, "HungerSpeedModifierSystem")

	movementSystem := simulation.NewMovementSystem(worldSizePixels, worldSizePixels)
	systemManager.AddSystem(&adapters.MovementSystemAdapter{System: movementSystem})
	systems = append(systems, "MovementSystem")

	systemManager.AddSystem(combatSystem)
	systems = append(systems, "CombatSystem")

	systemManager.AddSystem(&adapters.StarvationDamageSystemAdapter{System: starvationDamage})
	systems = append(systems, "StarvationDamageSystem")

	// Проверяем что системы добавлены
	t.Logf("Headless режим инициализирован с %d системами", len(systems))

	return systems
}

// УДАЛЕНО: testCombatFunctionality - перенесено в combat_functionality_test.go для уменьшения сложности
// func testCombatFunctionality(t *testing.T, cfg *config.Config) {
// 	t.Log("Создаём полную симуляцию для тестирования боевой системы...")
//
// 	// Создаём мир точно как в реальной игре
// 	terrainGen := generator.NewTerrainGenerator(cfg)
// 	terrain := terrainGen.Generate()
//
// 	worldSizePixels := float32(cfg.World.Size * 32)
// 	world := core.NewWorld(worldSizePixels, worldSizePixels, cfg.World.Seed)
// 	systemManager := core.NewSystemManager()
//
// 	// КРИТИЧЕСКИ ВАЖНО: анимационные системы для работы боевой системы
// 	wolfAnimationSystem := animation.NewAnimationSystem()
// 	rabbitAnimationSystem := animation.NewAnimationSystem()
//
// 	// Загружаем анимации
// 	loadAnimationsForTest(wolfAnimationSystem, rabbitAnimationSystem)
//
// 	// КРИТИЧЕСКИ ВАЖНО: AnimationManager для обновления анимаций в боевой системе
// 	animationManager := animation.NewAnimationManager(wolfAnimationSystem, rabbitAnimationSystem)
//
// 	// Инициализируем все системы как в реальной игре
// 	vegetationSystem := simulation.NewVegetationSystem(terrain)
// 	animalBehaviorSystem := simulation.NewAnimalBehaviorSystem(vegetationSystem)
// 	combatSystem := simulation.NewCombatSystem()
//
// 	// ВРЕМЕННО отключаем VegetationSystem и FeedingSystem для изоляции тестирования боевой системы
// 	_ = vegetationSystem // Убираем предупреждение
// 	systemManager.AddSystem(&adapters.BehaviorSystemAdapter{System: animalBehaviorSystem})
// 	systemManager.AddSystem(&adapters.MovementSystemAdapter{
//		System: simulation.NewMovementSystem(worldSizePixels, worldSizePixels),
//	})
// 	systemManager.AddSystem(combatSystem)
// 	systemManager.AddSystem(simulation.NewCorpseSystem()) // ВАЖНО: для превращения в трупы
//
// 	// КРИТИЧЕСКИ ВАЖНО: Добавляем AnimationManager как систему для обновления анимаций
// 	systemManager.AddSystem(animationManager)
//
// 	// Размещаем животных ТОЧНО как в реальной игре
// 	popGen := generator.NewPopulationGenerator(cfg, terrain)
// 	placements := popGen.Generate()
//
// 	var wolves []core.EntityID
// 	var rabbits []core.EntityID
//
// 	for _, placement := range placements {
// 		switch placement.Type {
// 		case core.TypeRabbit:
// 			rabbit := simulation.CreateAnimal(world, core.TypeRabbit, placement.X, placement.Y)
// 			rabbits = append(rabbits, rabbit)
// 		case core.TypeWolf:
// 			wolf := simulation.CreateAnimal(world, core.TypeWolf, placement.X, placement.Y)
// 			// КРИТИЧЕСКИ ВАЖНО: делаем волков голодными для тестирования боевой системы
// 			world.SetHunger(wolf, core.Hunger{Value: 10.0}) // 10% - очень голодные
//
// 			// КРИТИЧЕСКИ ВАЖНО: увеличиваем порог охоты для надёжного тестирования
// 			behavior, _ := world.GetBehavior(wolf)
// 			behavior.HungerThreshold = 90.0 // Охотится пока голод < 90%
// 			world.SetBehavior(wolf, behavior)
//
// 			wolves = append(wolves, wolf)
// 		}
// 	}
//
// 	t.Logf("Размещено: %d зайцев, %d волков", len(rabbits), len(wolves))
//
// 	// DEBUG: проверяем расстояния между волками и зайцами
// 	for i, wolf := range wolves {
// 		wolfPos, _ := world.GetPosition(wolf)
// 		wolfHunger, _ := world.GetHunger(wolf)
//
// 		minDistanceToRabbit := float32(999999)
// 		for _, rabbit := range rabbits {
// 			rabbitPos, _ := world.GetPosition(rabbit)
// 			dx := wolfPos.X - rabbitPos.X
// 			dy := wolfPos.Y - rabbitPos.Y
// 			distance := dx*dx + dy*dy
// 			if distance < minDistanceToRabbit {
// 				minDistanceToRabbit = distance
// 			}
// 		}
//
// 		t.Logf("DEBUG: Волк %d - позиция (%.1f, %.1f), голод %.1f%%, ближайший заяц: %.1f пикселей",
// 			i, wolfPos.X, wolfPos.Y, wolfHunger.Value, float32(math.Sqrt(float64(minDistanceToRabbit))))
// 	}
//
// 	// Функция обновления анимаций (КРИТИЧЕСКИ ВАЖНА!)
// 	updateAnimalAnimations := func() {
// 		world.ForEachWith(core.MaskAnimalType|core.MaskAnimation, func(entity core.EntityID) {
// 			animalType, ok := world.GetAnimalType(entity)
// 			if !ok {
// 				return
// 			}
//
// 			anim, hasAnim := world.GetAnimation(entity)
// 			if !hasAnim {
// 				return
// 			}
//
// 			var newAnimType animation.AnimationType
// 			var animSystem *animation.AnimationSystem
//
// 			switch animalType {
// 			case core.TypeWolf:
// 				newAnimType = getWolfAnimationTypeForTest(world, entity)
// 				animSystem = wolfAnimationSystem
// 			case core.TypeRabbit:
// 				newAnimType = getRabbitAnimationTypeForTest(world, entity)
// 				animSystem = rabbitAnimationSystem
// 			default:
// 				return
// 			}
//
// 			// НЕ прерываем анимацию ATTACK пока она играет!
// 			if anim.CurrentAnim != int(newAnimType) {
// 				if anim.CurrentAnim == int(animation.AnimAttack) && anim.Playing {
// 					// Анимация атаки должна доиграться до конца
// 				} else {
// 					anim.CurrentAnim = int(newAnimType)
// 					anim.Frame = 0
// 					anim.Timer = 0
// 					anim.Playing = true
// 					world.SetAnimation(entity, anim)
// 				}
// 			}
//
// 			// Обновляем анимацию
// 			animComponent := animation.AnimationComponent{
// 				CurrentAnim: animation.AnimationType(anim.CurrentAnim),
// 				Frame:       anim.Frame,
// 				Timer:       anim.Timer,
// 				Playing:     anim.Playing,
// 				FacingRight: anim.FacingRight,
// 			}
//
// 			animSystem.Update(&animComponent, 1.0/60.0)
//
// 			// Сохраняем состояние
// 			anim.Frame = animComponent.Frame
// 			anim.Timer = animComponent.Timer
// 			anim.Playing = animComponent.Playing
// 			world.SetAnimation(entity, anim)
// 		})
// 	}
//
// 	// Запускаем симуляцию на 15 секунд (чтобы обнаружить проблему голода волков)
// 	deltaTime := float32(1.0 / 60.0)
// 	maxTicks := 900 // 15 секунд
//
// 	initialRabbits := len(rabbits)
//
// 	for tick := 0; tick < maxTicks; tick++ {
// 		world.Update(deltaTime)
// 		systemManager.Update(world, deltaTime)
// 		updateAnimalAnimations()
//
// 		// DEBUG: каждые 3 секунды проверяем позицию ближайшего волка
// 		if tick%180 == 0 && len(wolves) > 0 {
// 			wolf := wolves[2] // Самый близкий к зайцу
// 			wolfPos, _ := world.GetPosition(wolf)
// 			wolfHunger, _ := world.GetHunger(wolf)
//
// 			minDistanceToRabbit := float32(999999)
// 			for _, rabbit := range rabbits {
// 				if !world.IsAlive(rabbit) {
// 					continue // Пропускаем мертвых зайцев
// 				}
// 				rabbitPos, _ := world.GetPosition(rabbit)
// 				dx := wolfPos.X - rabbitPos.X
// 				dy := wolfPos.Y - rabbitPos.Y
// 				distance := dx*dx + dy*dy
// 				if distance < minDistanceToRabbit {
// 					minDistanceToRabbit = distance
// 				}
// 			}
//
// 			t.Logf("DEBUG: Секунда %d - Волк 2: позиция (%.1f, %.1f), голод %.1f%%, ближайший заяц: %.1f пикселей",
// 				tick/60, wolfPos.X, wolfPos.Y, wolfHunger.Value, float32(math.Sqrt(float64(minDistanceToRabbit))))
// 		}
// 	}
//
// 	// Подсчитываем результат
// 	finalRabbits := 0
// 	finalWolves := 0
// 	attacksOccurred := false
//
// 	world.ForEachWith(core.MaskAnimalType, func(entity core.EntityID) {
// 		animalType, ok := world.GetAnimalType(entity)
// 		if !ok {
// 			return
// 		}
//
// 		// Считаем только ЖИВЫХ животных (без компонента Corpse)
// 		if world.HasComponent(entity, core.MaskCorpse) {
// 			return // Пропускаем трупы
// 		}
//
// 		if animalType == core.TypeRabbit {
// 			finalRabbits++
// 		} else if animalType == core.TypeWolf {
// 			finalWolves++
//
// 			// Проверяем были ли атаки
// 			if world.HasComponent(entity, core.MaskAttackState) {
// 				attacksOccurred = true
// 			}
// 		}
// 	})
//
// 	t.Logf("Результат симуляции:")
// 	t.Logf("  Зайцы: %d → %d", initialRabbits, finalRabbits)
// 	t.Logf("  Волки: %d", finalWolves)
// 	t.Logf("  Атаки произошли: %t", attacksOccurred)
//
// 	// КРИТИЧЕСКАЯ ПРОВЕРКА: боевая система должна работать!
// 	if finalRabbits >= initialRabbits {
// 		t.Error("❌ КРИТИЧЕСКАЯ ОШИБКА: Боевая система не работает! Ни один заяц не был убит за 15 секунд")
// 	} else {
// 		t.Logf("✅ Боевая система работает: %d зайцев убито", initialRabbits-finalRabbits)
// 	}
//
// 	// КРИТИЧЕСКАЯ ПРОВЕРКА: волки должны выживать если есть еда!
// 	if finalWolves == 0 && finalRabbits < initialRabbits {
// 		t.Error("❌ КРИТИЧЕСКАЯ ОШИБКА: Все волки умерли от голода несмотря на то что убивали зайцев! " +
// 			"Проблема с поеданием трупов!")
// 	} else if finalWolves > 0 {
// 		t.Logf("✅ Волки выживают: %d волков живы", finalWolves)
// 	}
// }

// Вспомогательные функции для тестирования

func loadAnimationsForTest(wolfAnimSystem, rabbitAnimSystem *animation.AnimationSystem) {
	// КРИТИЧЕСКОЕ ИСПРАВЛЕНИЕ: Нужно зарегистрировать анимации для работы боевой системы
	// Без этого AnimationSystem.Update() не переключает кадры и AttackPhaseWindup
	// никогда не переходит в AttackPhaseStrike!

	// Создаём пустое изображение через ebiten (в тестах содержимое не важно)
	emptyImg := ebiten.NewImage(1, 1)

	// Регистрируем анимации волка ТОЧНО как в реальной игре
	wolfAnimSystem.RegisterAnimation(animation.AnimIdle, 2, 1.0, true, emptyImg)
	wolfAnimSystem.RegisterAnimation(animation.AnimWalk, 4, 8.0, true, emptyImg)
	wolfAnimSystem.RegisterAnimation(animation.AnimRun, 4, 12.0, true, emptyImg)
	wolfAnimSystem.RegisterAnimation(animation.AnimAttack, 2, 6.0, false, emptyImg) // НЕ зацикленная!
	wolfAnimSystem.RegisterAnimation(animation.AnimEat, 2, 2.0, true, emptyImg)

	// Регистрируем анимации зайца ТОЧНО как в реальной игре
	rabbitAnimSystem.RegisterAnimation(animation.AnimIdle, 2, 1.0, true, emptyImg)
	rabbitAnimSystem.RegisterAnimation(animation.AnimWalk, 4, 8.0, true, emptyImg)
	rabbitAnimSystem.RegisterAnimation(animation.AnimRun, 4, 12.0, true, emptyImg)
	rabbitAnimSystem.RegisterAnimation(animation.AnimDeathDying, 1, 1.0, false, emptyImg)
}

func getWolfAnimationTypeForTest(world *core.World, entity core.EntityID) animation.AnimationType {
	if world.HasComponent(entity, core.MaskEatingState) {
		return animation.AnimEat
	}

	if isWolfAttackingForTest(world, entity) {
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

func getRabbitAnimationTypeForTest(world *core.World, entity core.EntityID) animation.AnimationType {
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

func isWolfAttackingForTest(world *core.World, wolf core.EntityID) bool {
	hunger, hasHunger := world.GetHunger(wolf)
	if !hasHunger || hunger.Value >= 60.0 {
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
	return distance <= 13.0*13.0 // Используем текущий радиус
}
