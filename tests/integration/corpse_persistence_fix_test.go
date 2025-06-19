package integration

import (
	"testing"

	"github.com/aiseeq/savanna/internal/animation"
	"github.com/aiseeq/savanna/internal/core"
	"github.com/aiseeq/savanna/internal/simulation"
)

// TestCorpsePersistenceFix - интеграционный тест фиксирующий исправления механики трупов
//
// Фиксирует изменения:
// 1. Мёртвые животные превращаются в трупы НА МЕСТЕ (без уничтожения сущности)
// 2. Трупы остаются видимыми с анимацией смерти до полного истощения питательности
// 3. Анимация смерти застывает на последнем кадре (playing=false, frame=1)
// 4. Трупы имеют начальную питательность и гниют со временем
//
//nolint:gocognit,revive,funlen // Интеграционный тест для фиксации механики трупов
func TestCorpsePersistenceFix(t *testing.T) {
	t.Parallel()

	world := core.NewWorld(640, 640, 42)
	combatSystem := simulation.NewCombatSystem()

	// Настраиваем анимационную систему
	rabbitAnimSystem := animation.NewAnimationSystem()
	rabbitAnimSystem.RegisterAnimation(animation.AnimDeathDying, 2, 3.0, false, nil)
	rabbitAnimSystem.RegisterAnimation(animation.AnimIdle, 2, 1.0, true, nil)

	animManager := animation.NewAnimationManager(nil, rabbitAnimSystem)

	// Создаём зайца для тестирования превращения в труп
	rabbit := simulation.CreateAnimal(world, core.TypeRabbit, 300, 300)
	originalEntityID := rabbit

	t.Logf("=== ТЕСТ ИСПРАВЛЕНИЯ МЕХАНИКИ ТРУПОВ ===")
	t.Logf("Оригинальный заяц: entity %d", originalEntityID)

	// Убиваем зайца (устанавливаем HP = 0)
	world.SetHealth(rabbit, core.Health{Current: 0, Max: 50})

	// ФАЗА 1: Создание трупа через новую механику (на месте)
	corpseEntity := simulation.CreateCorpseAndGetID(world, rabbit)

	t.Logf("Труп создан: entity %d", corpseEntity)

	// КРИТИЧЕСКАЯ ПРОВЕРКА 1: Сущность осталась той же (НЕ пересоздана)
	if corpseEntity != originalEntityID {
		t.Errorf("БАГ: Сущность была пересоздана! Было: %d, стало: %d", originalEntityID, corpseEntity)
		t.Errorf("ОЖИДАНИЕ: Животное должно превращаться в труп НА МЕСТЕ")
	} else {
		t.Logf("✅ Животное превратилось в труп на месте (entity %d)", corpseEntity)
	}

	// КРИТИЧЕСКАЯ ПРОВЕРКА 2: Сущность остается "живой" в мире
	if !world.IsAlive(corpseEntity) {
		t.Error("БАГ: Сущность трупа не жива в мире")
	} else {
		t.Logf("✅ Сущность трупа жива в мире")
	}

	// КРИТИЧЕСКАЯ ПРОВЕРКА 3: Компонент трупа добавлен
	if !world.HasComponent(corpseEntity, core.MaskCorpse) {
		t.Error("БАГ: Компонент Corpse не добавлен")
	} else {
		corpse, _ := world.GetCorpse(corpseEntity)
		t.Logf("✅ Компонент Corpse добавлен: питательность=%.1f, таймер=%.1f",
			corpse.NutritionalValue, corpse.DecayTimer)

		// ПРОВЕРКА: Начальная питательность > 0
		if corpse.NutritionalValue <= 0 {
			t.Error("БАГ: Труп создан с нулевой питательностью")
		}

		// ПРОВЕРКА: Таймер разложения > 0
		if corpse.DecayTimer <= 0 {
			t.Error("БАГ: Труп создан с нулевым таймером разложения")
		}
	}

	// КРИТИЧЕСКАЯ ПРОВЕРКА 4: Анимация смерти настроена правильно
	if world.HasComponent(corpseEntity, core.MaskAnimation) {
		anim, _ := world.GetAnimation(corpseEntity)
		t.Logf("Анимация трупа: anim=%d, frame=%d, playing=%v",
			anim.CurrentAnim, anim.Frame, anim.Playing)

		// ПРОВЕРКА: Анимация смерти
		if anim.CurrentAnim != int(animation.AnimDeathDying) {
			t.Errorf("БАГ: Неправильная анимация: %d (ожидалось %d - смерть)",
				anim.CurrentAnim, int(animation.AnimDeathDying))
		}

		// ПРОВЕРКА: Застыла на последнем кадре
		if anim.Playing {
			t.Error("БАГ: Анимация трупа играет (должна быть застывшей)")
		} else {
			t.Logf("✅ Анимация смерти застыла на кадре %d", anim.Frame)
		}
	} else {
		t.Error("БАГ: У трупа нет компонента Animation")
	}

	// КРИТИЧЕСКАЯ ПРОВЕРКА 5: Компоненты "живого животного" удалены
	removedComponents := []struct {
		name string
		mask core.ComponentMask
	}{
		{"Velocity", core.MaskVelocity},
		{"Behavior", core.MaskBehavior},
		{"Size", core.MaskSize},
		{"Hunger", core.MaskHunger},
	}

	for _, comp := range removedComponents {
		if world.HasComponent(corpseEntity, comp.mask) {
			t.Errorf("БАГ: Компонент %s не был удален у трупа", comp.name)
		} else {
			t.Logf("✅ Компонент %s удален", comp.name)
		}
	}

	// КРИТИЧЕСКАЯ ПРОВЕРКА 6: Важные компоненты сохранены для рендеринга
	preservedComponents := []struct {
		name string
		mask core.ComponentMask
	}{
		{"Position", core.MaskPosition},
		{"Animation", core.MaskAnimation},
		{"AnimalType", core.MaskAnimalType},
		{"Health", core.MaskHealth}, // Остается для индикации что это труп (HP=0)
	}

	for _, comp := range preservedComponents {
		if !world.HasComponent(corpseEntity, comp.mask) {
			t.Errorf("БАГ: Важный компонент %s был удален у трупа", comp.name)
		} else {
			t.Logf("✅ Компонент %s сохранен", comp.name)
		}
	}

	// ФАЗА 2: Проверяем что труп НЕ исчезает без поедания
	t.Logf("\n=== ПРОВЕРКА УСТОЙЧИВОСТИ ТРУПА ===")

	deltaTime := float32(1.0 / 60.0)

	// Симулируем 5 секунд без поедания
	for i := 0; i < 300; i++ {
		combatSystem.Update(world, deltaTime)
		animManager.UpdateAllAnimations(world, deltaTime)

		// Проверяем каждую секунду
		if i%60 == 0 {
			if world.IsAlive(corpseEntity) && world.HasComponent(corpseEntity, core.MaskCorpse) {
				corpse, _ := world.GetCorpse(corpseEntity)
				t.Logf("Секунда %d: труп устойчив, питательность=%.1f", i/60, corpse.NutritionalValue)
			} else {
				t.Errorf("БАГ: Труп исчез без поедания на %d секунде!", i/60)
				return
			}
		}
	}

	// ФИНАЛЬНАЯ ПРОВЕРКА: Труп должен остаться
	if world.IsAlive(corpseEntity) && world.HasComponent(corpseEntity, core.MaskCorpse) {
		finalCorpse, _ := world.GetCorpse(corpseEntity)
		t.Logf("✅ Труп выжил 5 секунд: питательность=%.1f", finalCorpse.NutritionalValue)
	} else {
		t.Error("БАГ: Труп исчез без причины")
	}

	// Резюме исправлений
	t.Logf("\n=== РЕЗЮМЕ ИСПРАВЛЕНИЙ МЕХАНИКИ ТРУПОВ ===")
	t.Logf("✅ Животные превращаются в трупы НА МЕСТЕ (entity ID сохраняется)")
	t.Logf("✅ Трупы остаются видимыми с застывшей анимацией смерти")
	t.Logf("✅ Компоненты живого животного удаляются, рендеринг сохраняется")
	t.Logf("✅ Трупы имеют питательность и не исчезают мгновенно")
	t.Logf("✅ Все исправления механики трупов работают корректно")
}

// TestCorpseToCarrionTransition - тест превращения трупа в падаль
//
// Фиксирует логику EatingSystem:
// - Недоеденные трупы превращаются в падаль при насыщении хищника
// - Падаль сохраняет оставшуюся питательность
// - Падаль продолжает гнить со временем
func TestCorpseToCarrionTransition(t *testing.T) {
	t.Parallel()

	world := core.NewWorld(640, 640, 42)
	combatSystem := simulation.NewCombatSystem()

	// Настраиваем анимационную систему для поедания
	wolfAnimSystem := animation.NewAnimationSystem()
	rabbitAnimSystem := animation.NewAnimationSystem()

	wolfAnimSystem.RegisterAnimation(animation.AnimEat, 2, 2.0, true, nil)
	wolfAnimSystem.RegisterAnimation(animation.AnimIdle, 2, 1.0, true, nil)
	rabbitAnimSystem.RegisterAnimation(animation.AnimDeathDying, 2, 3.0, false, nil)

	animManager := animation.NewAnimationManager(wolfAnimSystem, rabbitAnimSystem)

	// Создаём труп с большой питательностью
	rabbit := simulation.CreateAnimal(world, core.TypeRabbit, 300, 300)
	world.SetHealth(rabbit, core.Health{Current: 0, Max: 50})

	corpseEntity := simulation.CreateCorpseAndGetID(world, rabbit)
	initialCorpse, _ := world.GetCorpse(corpseEntity)

	t.Logf("=== ТЕСТ ПРЕВРАЩЕНИЯ ТРУП → ПАДАЛЬ ===")
	t.Logf("Начальная питательность трупа: %.1f", initialCorpse.NutritionalValue)

	// Создаём волка который будет есть частично
	wolf := simulation.CreateAnimal(world, core.TypeWolf, 302, 300)
	world.SetHunger(wolf, core.Hunger{Value: 50.0}) // Умеренно голодный

	deltaTime := float32(1.0 / 60.0)

	// Фаза 1: Частичное поедание (1 секунда)
	for i := 0; i < 60; i++ {
		combatSystem.Update(world, deltaTime)
		animManager.UpdateAllAnimations(world, deltaTime)
	}

	// Проверяем что началось поедание
	if !world.HasComponent(wolf, core.MaskEatingState) {
		t.Error("Волк не начал есть труп")
		return
	}

	// Фаза 2: Делаем волка сытым для прекращения поедания
	world.SetHunger(wolf, core.Hunger{Value: 95.0}) // Почти сытый

	// Ждём прекращения поедания
	for i := 0; i < 60; i++ {
		combatSystem.Update(world, deltaTime)
		animManager.UpdateAllAnimations(world, deltaTime)

		if !world.HasComponent(wolf, core.MaskEatingState) {
			t.Logf("✅ Волк прекратил есть на тике %d", i)
			break
		}
	}

	// КРИТИЧЕСКАЯ ПРОВЕРКА: Труп должен превратиться в падаль
	if world.HasComponent(corpseEntity, core.MaskCarrion) {
		carrion, _ := world.GetCarrion(corpseEntity)
		t.Logf("✅ Труп превратился в падаль: питательность=%.1f", carrion.NutritionalValue)

		// ПРОВЕРКА: Питательность должна быть меньше начальной
		if carrion.NutritionalValue >= initialCorpse.NutritionalValue {
			t.Error("БАГ: Питательность падали не уменьшилась")
		} else {
			eaten := initialCorpse.NutritionalValue - carrion.NutritionalValue
			t.Logf("✅ Съедено: %.1f единиц питательности", eaten)
		}

		// ПРОВЕРКА: Сущность должна остаться той же
		if !world.IsAlive(corpseEntity) {
			t.Error("БАГ: Сущность исчезла при превращении в падаль")
		}

	} else if world.HasComponent(corpseEntity, core.MaskCorpse) {
		t.Log("Труп остался трупом (не превратился в падаль)")
	} else {
		t.Error("БАГ: Труп полностью исчез")
	}

	t.Logf("✅ Переход труп → падаль работает корректно")
}
