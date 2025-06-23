package integration

import (
	"testing"

	"github.com/aiseeq/savanna/config"
	"github.com/aiseeq/savanna/internal/animation"
	"github.com/aiseeq/savanna/internal/core"
	"github.com/aiseeq/savanna/internal/generator"
	"github.com/aiseeq/savanna/internal/simulation"
)

// TestCorpseDecayMechanics - TDD тест для правильной механики разложения трупов
//
// ТЕКУЩИЙ БАГ: Труп исчезает сразу после убийства волком
// ОЖИДАЕМОЕ ПОВЕДЕНИЕ:
// 1. Труп имеет питательность (например, 200 единиц)
// 2. При поедании питательность уменьшается
// 3. Если волк перестаёт есть - труп остаётся с оставшейся питательностью
// 4. Труп гниёт со временем (питательность уменьшается)
// 5. Труп исчезает только когда питательность = 0
//
//nolint:gocognit,revive,funlen // TDD тест для сложной механики трупов
func TestCorpseDecayMechanics(t *testing.T) {
	t.Parallel()

	world := core.NewWorld(640, 640, 42)
	combatSystem := simulation.NewCombatSystem()
	eatingSystem := simulation.NewEatingSystem() // Система поедания трупов

	// Создаём vegetation систему для поведения
	cfg := config.LoadDefaultConfig()
	cfg.World.Size = int(640 / 32)
	terrainGen := generator.NewTerrainGenerator(cfg)
	terrain := terrainGen.Generate()
	vegetationSystem := simulation.NewVegetationSystem(terrain)
	behaviorSystem := simulation.NewAnimalBehaviorSystem(vegetationSystem)

	// ИСПРАВЛЕНИЕ: Добавляем анимационную систему для работы дискретного поедания
	wolfAnimSystem := animation.NewAnimationSystem()
	rabbitAnimSystem := animation.NewAnimationSystem()

	// Регистрируем анимации поедания
	wolfAnimSystem.RegisterAnimation(animation.AnimEat, 2, 2.0, true, nil)
	wolfAnimSystem.RegisterAnimation(animation.AnimIdle, 2, 1.0, true, nil)
	rabbitAnimSystem.RegisterAnimation(animation.AnimDeathDying, 2, 3.0, false, nil)

	animManager := animation.NewAnimationManager(wolfAnimSystem, rabbitAnimSystem)

	// Создаём зайца и волка
	rabbit := simulation.CreateAnimal(world, core.TypeRabbit, 300, 300)
	wolf := simulation.CreateAnimal(world, core.TypeWolf, 301, 300) // Дистанция 1 пиксель

	// Делаем волка голодным
	world.SetHunger(wolf, core.Hunger{Value: 30.0})

	// Убиваем зайца вручную и создаём труп (имитируем убийство волком)
	world.SetHealth(rabbit, core.Health{Current: 0, Max: 50})

	t.Logf("=== ТЕСТ МЕХАНИКИ РАЗЛОЖЕНИЯ ТРУПОВ ===")

	deltaTime := float32(1.0 / 60.0)

	// Фаза 1: Создание трупа (вызываем createCorpse напрямую как делает attack_system)
	corpseEntity := simulation.CreateCorpseAndGetID(world, rabbit)

	if corpseEntity == 0 {
		t.Fatal("Труп не создался после смерти зайца")
	}

	initialCorpse, _ := world.GetCorpse(corpseEntity)
	t.Logf("Труп создан: питательность=%.1f, таймер разложения=%.1f сек",
		initialCorpse.NutritionalValue, initialCorpse.DecayTimer)

	// КРИТИЧЕСКАЯ ПРОВЕРКА: Труп должен иметь питательность > 0
	if initialCorpse.NutritionalValue <= 0 {
		t.Error("БАГ: Труп создан с нулевой питательностью")
	}

	// КРИТИЧЕСКАЯ ПРОВЕРКА: Труп должен иметь таймер разложения > 0
	if initialCorpse.DecayTimer <= 0 {
		t.Error("БАГ: Труп создан с нулевым таймером разложения")
	}

	// Фаза 2: Волк начинает есть труп
	for i := 0; i < 60; i++ { // 1 секунда для начала поедания
		behaviorSystem.Update(world, deltaTime) // КРИТИЧЕСКИ: поиск трупа!
		combatSystem.Update(world, deltaTime)
		eatingSystem.Update(world, deltaTime) // КРИТИЧЕСКИ: система поедания трупов!
		animManager.UpdateAllAnimations(world, deltaTime)
		if world.HasComponent(wolf, core.MaskEatingState) {
			eatingState, _ := world.GetEatingState(wolf)
			if eatingState.Target == corpseEntity {
				t.Logf("✅ Волк начал есть труп (entity %d)", corpseEntity)
				break
			}
		}
	}

	if !world.HasComponent(wolf, core.MaskEatingState) {
		t.Error("Волк не начал есть труп за 1 секунду")
	}

	// Фаза 3: Симулируем частичное поедание (волк ест 3 секунды)
	t.Logf("\n=== ФАЗА ЧАСТИЧНОГО ПОЕДАНИЯ ===")

	for i := 0; i < 180; i++ { // 3 секунды поедания
		behaviorSystem.Update(world, deltaTime) // КРИТИЧЕСКИ: поиск трупа!
		combatSystem.Update(world, deltaTime)
		eatingSystem.Update(world, deltaTime) // КРИТИЧЕСКИ: система поедания трупов!
		animManager.UpdateAllAnimations(world, deltaTime)

		// ДЕТАЛЬНОЕ ЛОГИРОВАНИЕ: Проверяем состояние волка и анимации каждые 10 тиков
		if i%10 == 0 {
			if world.HasComponent(wolf, core.MaskEatingState) {
				eatingState, _ := world.GetEatingState(wolf)
				t.Logf("Тик %d: Волк ест target=%d, progress=%.2f, nutrition=%.1f",
					i, eatingState.Target, eatingState.EatingProgress, eatingState.NutritionGained)

				// Проверяем анимацию волка
				if world.HasComponent(wolf, core.MaskAnimation) {
					anim, _ := world.GetAnimation(wolf)
					t.Logf("  Анимация волка: anim=%d, frame=%d, playing=%v",
						anim.CurrentAnim, anim.Frame, anim.Playing)
				}
			} else {
				t.Logf("Тик %d: Волк НЕ в состоянии поедания!", i)
			}
		}

		// Проверяем состояние трупа каждую секунду
		if i%60 == 0 && world.HasComponent(corpseEntity, core.MaskCorpse) {
			currentCorpse, _ := world.GetCorpse(corpseEntity)
			t.Logf("Секунда %d: питательность=%.1f (было %.1f)",
				i/60, currentCorpse.NutritionalValue, initialCorpse.NutritionalValue)

			// КРИТИЧЕСКАЯ ПРОВЕРКА: Питательность должна уменьшаться при поедании
			if i > 60 && currentCorpse.NutritionalValue >= initialCorpse.NutritionalValue {
				t.Errorf("БАГ: Питательность трупа НЕ уменьшается при поедании!")
				t.Errorf("Начальная: %.1f, текущая: %.1f",
					initialCorpse.NutritionalValue, currentCorpse.NutritionalValue)

				// ДОПОЛНИТЕЛЬНАЯ ДИАГНОСТИКА: Проверяем состояние систем
				if world.HasComponent(wolf, core.MaskEatingState) {
					t.Errorf("Волк В СОСТОЯНИИ поедания - система должна работать")
				} else {
					t.Errorf("Волк НЕ В СОСТОЯНИИ поедания - вот причина проблемы!")
				}
				return
			}
		}
	}

	// Фаза 4: Волк прекращает есть (имитируем насыщение)
	t.Logf("\n=== ФАЗА ПРЕКРАЩЕНИЯ ПОЕДАНИЯ ===")

	// Делаем волка сытым чтобы он перестал есть
	world.SetHunger(wolf, core.Hunger{Value: 90.0}) // Сытый волк

	// Обновляем систему чтобы волк перестал есть
	for i := 0; i < 60; i++ {
		behaviorSystem.Update(world, deltaTime) // КРИТИЧЕСКИ: поиск трупа!
		combatSystem.Update(world, deltaTime)
		eatingSystem.Update(world, deltaTime) // КРИТИЧЕСКИ: система поедания трупов!
		animManager.UpdateAllAnimations(world, deltaTime)
		if !world.HasComponent(wolf, core.MaskEatingState) {
			t.Logf("Волк перестал есть на тике %d", i)
			break
		}
	}

	// КРИТИЧЕСКАЯ ПРОВЕРКА: Труп должен ОСТАТЬСЯ с анимацией смерти
	if !world.IsAlive(corpseEntity) {
		t.Errorf("БАГ: Сущность трупа была уничтожена!")
		return
	}

	if world.HasComponent(corpseEntity, core.MaskCorpse) {
		partiallyEatenCorpse, _ := world.GetCorpse(corpseEntity)
		t.Logf("✅ Труп остался: питательность=%.1f (было %.1f)",
			partiallyEatenCorpse.NutritionalValue, initialCorpse.NutritionalValue)

		// НОВАЯ ПРОВЕРКА: Анимация смерти должна быть застывшей
		if world.HasComponent(corpseEntity, core.MaskAnimation) {
			anim, _ := world.GetAnimation(corpseEntity)
			t.Logf("  Анимация трупа: anim=%d, frame=%d, playing=%v",
				anim.CurrentAnim, anim.Frame, anim.Playing)

			if anim.Playing {
				t.Error("БАГ: Анимация трупа не должна играть - должна застыть")
			}
		}
	} else if world.HasComponent(corpseEntity, core.MaskCarrion) {
		carrion, _ := world.GetCarrion(corpseEntity)
		t.Logf("✅ Труп превратился в падаль: питательность=%.1f (было %.1f)",
			carrion.NutritionalValue, initialCorpse.NutritionalValue)
	} else {
		t.Errorf("БАГ ОБНАРУЖЕН: Труп НЕ имеет компонента Corpse!")
		t.Errorf("ОЖИДАНИЕ: Труп должен остаться с компонентом Corpse")
		return
	}

	// ПРОВЕРКА: Питательность должна быть меньше начальной
	var currentNutritionalValue float32
	if world.HasComponent(corpseEntity, core.MaskCorpse) {
		partiallyEatenCorpse, _ := world.GetCorpse(corpseEntity)
		currentNutritionalValue = partiallyEatenCorpse.NutritionalValue
	} else if world.HasComponent(corpseEntity, core.MaskCarrion) {
		carrion, _ := world.GetCarrion(corpseEntity)
		currentNutritionalValue = carrion.NutritionalValue
	}

	if currentNutritionalValue >= initialCorpse.NutritionalValue {
		t.Error("БАГ: Питательность не уменьшилась после частичного поедания")
	}

	// Фаза 5: Естественное гниение трупа
	t.Logf("\n=== ФАЗА ЕСТЕСТВЕННОГО ГНИЕНИЯ ===")

	// Симулируем долгое время без поедания (труп должен гнить)
	for i := 0; i < 3900; i++ { // 65 секунд гниения (полное разложение гарантировано)
		behaviorSystem.Update(world, deltaTime) // КРИТИЧЕСКИ: поиск трупа!
		combatSystem.Update(world, deltaTime)
		eatingSystem.Update(world, deltaTime) // КРИТИЧЕСКИ: система поедания трупов!
		animManager.UpdateAllAnimations(world, deltaTime)

		// ДЕТАЛЬНОЕ ЛОГИРОВАНИЕ: Отслеживаем превращение труп → падаль
		if i%10 == 0 {
			if world.HasComponent(corpseEntity, core.MaskCorpse) {
				currentCorpse, _ := world.GetCorpse(corpseEntity)
				t.Logf("Тик %d: ТРУП питательность=%.1f, таймер=%.1f",
					i, currentCorpse.NutritionalValue, currentCorpse.DecayTimer)
			} else if world.HasComponent(corpseEntity, core.MaskCarrion) {
				carrion, _ := world.GetCarrion(corpseEntity)
				t.Logf("Тик %d: ПАДАЛЬ питательность=%.1f, таймер=%.1f",
					i, carrion.NutritionalValue, carrion.DecayTimer)
			} else if world.IsAlive(corpseEntity) {
				t.Logf("Тик %d: Сущность жива, но без Corpse/Carrion компонентов", i)
			} else {
				t.Logf("Тик %d: Сущность мертва/уничтожена", i)
			}
		}

		// Проверяем исчезновение трупа/падали
		if !world.HasComponent(corpseEntity, core.MaskCorpse) && !world.HasComponent(corpseEntity, core.MaskCarrion) {
			t.Logf("🔍 ДИАГНОСТИКА: Труп исчез на %d тике (%.1f секунды)", i, float32(i)/60.0)

			// ДИАГНОСТИКА: Проверяем все компоненты сущности трупа
			t.Logf("  Сущность жива: %v", world.IsAlive(corpseEntity))
			t.Logf("  Имеет Position: %v", world.HasComponent(corpseEntity, core.MaskPosition))
			t.Logf("  Имеет Animation: %v", world.HasComponent(corpseEntity, core.MaskAnimation))
			t.Logf("  Имеет Health: %v", world.HasComponent(corpseEntity, core.MaskHealth))
			t.Logf("  Имеет AnimalType: %v", world.HasComponent(corpseEntity, core.MaskAnimalType))
			t.Logf("  Имеет Corpse: %v", world.HasComponent(corpseEntity, core.MaskCorpse))
			t.Logf("  Имеет Carrion: %v", world.HasComponent(corpseEntity, core.MaskCarrion))

			// ПРОВЕРКА: Труп/падаль должен полностью исчезнуть из мира
			if world.IsAlive(corpseEntity) {
				t.Error("БАГ: Сущность остался живой после полного разложения")
			} else {
				t.Logf("✅ Труп/падаль полностью исчез из мира")
			}
			return
		}
	}

	// Если дошли до конца - труп/падаль не разложился
	if world.HasComponent(corpseEntity, core.MaskCorpse) {
		finalCorpse, _ := world.GetCorpse(corpseEntity)
		t.Errorf("БАГ: Труп не разложился за 65 секунд")
		t.Errorf("Финальное состояние: питательность=%.1f, таймер=%.1f",
			finalCorpse.NutritionalValue, finalCorpse.DecayTimer)
	} else if world.HasComponent(corpseEntity, core.MaskCarrion) {
		finalCarrion, _ := world.GetCarrion(corpseEntity)
		t.Errorf("БАГ: Падаль не разложилась за 65 секунд")
		t.Errorf("Финальное состояние: питательность=%.1f, таймер=%.1f",
			finalCarrion.NutritionalValue, finalCarrion.DecayTimer)
	}
}

// TestCorpseNutritionDepletion - тест на полное истощение питательности
func TestCorpseNutritionDepletion(t *testing.T) {
	t.Parallel()

	world := core.NewWorld(640, 640, 42)
	combatSystem := simulation.NewCombatSystem()

	// ИСПРАВЛЕНИЕ: Добавляем анимационную систему как в основном тесте
	wolfAnimSystem := animation.NewAnimationSystem()
	rabbitAnimSystem := animation.NewAnimationSystem()

	// Регистрируем анимации поедания
	wolfAnimSystem.RegisterAnimation(animation.AnimEat, 2, 2.0, true, nil)
	wolfAnimSystem.RegisterAnimation(animation.AnimIdle, 2, 1.0, true, nil)
	rabbitAnimSystem.RegisterAnimation(animation.AnimDeathDying, 2, 3.0, false, nil)

	animManager := animation.NewAnimationManager(wolfAnimSystem, rabbitAnimSystem)

	// Создаём труп вручную с малой питательностью для быстрого теста
	rabbit := simulation.CreateAnimal(world, core.TypeRabbit, 300, 300)
	world.SetHealth(rabbit, core.Health{Current: 0, Max: 50})

	// ИСПРАВЛЕНИЕ: Напрямую создаём труп как в других тестах
	corpseEntity := simulation.CreateCorpseAndGetID(world, rabbit)

	if corpseEntity == 0 || !world.HasComponent(corpseEntity, core.MaskCorpse) {
		t.Fatal("Труп не создался")
	}

	// Устанавливаем очень малую питательность для быстрого теста
	world.SetCorpse(corpseEntity, core.Corpse{
		NutritionalValue: 5.0, // Очень мало питательности
		MaxNutritional:   200.0,
		DecayTimer:       60.0,
	})

	t.Logf("=== ТЕСТ ИСТОЩЕНИЯ ПИТАТЕЛЬНОСТИ ===")
	t.Logf("Установлена питательность: 5.0 единиц")

	// Создаём очень голодного волка для агрессивного поедания (на той же позиции что и труп)
	wolf := simulation.CreateAnimal(world, core.TypeWolf, 300, 300)
	world.SetHunger(wolf, core.Hunger{Value: 5.0}) // Очень голодный

	deltaTime := float32(1.0 / 60.0)

	// Симулируем полное поедание
	for i := 0; i < 300; i++ { // 5 секунд
		combatSystem.Update(world, deltaTime)
		animManager.UpdateAllAnimations(world, deltaTime)

		if !world.HasComponent(corpseEntity, core.MaskCorpse) {
			t.Logf("✅ Труп полностью съеден на тике %d (%.1f сек)", i, float32(i)/60.0)

			// ПРОВЕРКА: Заяц должен исчезнуть из мира
			if world.IsAlive(corpseEntity) {
				t.Error("БАГ: Заяц остался живым после полного поедания")
			}
			return
		}

		// Логируем прогресс каждую секунду
		if i%60 == 0 && world.HasComponent(corpseEntity, core.MaskCorpse) {
			corpse, _ := world.GetCorpse(corpseEntity)
			t.Logf("Секунда %d: питательность=%.1f", i/60, corpse.NutritionalValue)
		}
	}

	t.Error("Труп не был полностью съеден за 5 секунд")
}
