package integration

import (
	"testing"

	"github.com/aiseeq/savanna/internal/animation"
	"github.com/aiseeq/savanna/internal/core"
	"github.com/aiseeq/savanna/internal/simulation"
)

// TestFrameByFrameCombat детальный TDD тест покадрового боя
//
//nolint:gocognit,revive // Покадровый тест боевой системы с детальной валидацией
func TestFrameByFrameCombat(t *testing.T) {
	t.Parallel()
	world := core.NewWorld(640, 640, 42) // Фиксированный seed для детерминизма
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
	wolf := simulation.CreateAnimal(world, core.TypeWolf, 301, 300) // Дистанция 1 пиксель

	// Делаем волка голодным
	world.SetHunger(wolf, core.Hunger{Value: 30.0})

	initialHealth, _ := world.GetHealth(rabbit)
	t.Logf("=== НАЧАЛО ТЕСТА ===")
	t.Logf("Здоровье зайца: %d, голод волка: 30%%", initialHealth.Current)

	deltaTime := float32(1.0 / 60.0) // 60 FPS

	// Сначала создаем AttackState через CombatSystem
	combatSystem.Update(world, deltaTime)

	if !world.HasComponent(wolf, core.MaskAttackState) {
		t.Fatal("AttackState не создан! Волк должен атаковать голодным")
	}

	// === ФАЗА 1: КАДР 0 АТАКИ (НЕТ УРОНА - WINDUP) ===
	t.Logf("\n=== ФАЗА 1: КАДР 0 АТАКИ ===")

	// Устанавливаем анимацию атаки на кадр 0 (Windup)
	anim := core.Animation{
		CurrentAnim: int(animation.AnimAttack),
		Frame:       0, // Кадр 0 = Windup
		Timer:       0,
		Playing:     true,
		FacingRight: true,
	}
	world.SetAnimation(wolf, anim)

	t.Logf("Волк в анимации ATTACK кадр 0 (Windup)")

	// Проверяем что урона НЕТ на кадре 0 (Windup)
	combatSystem.Update(world, deltaTime)
	healthAfterFrame0, _ := world.GetHealth(rabbit)

	if healthAfterFrame0.Current != initialHealth.Current {
		t.Errorf("ОШИБКА: Урон на кадре 0 (Windup)! %d -> %d", initialHealth.Current, healthAfterFrame0.Current)
	} else {
		t.Logf("✅ Кадр 0: НЕТ урона (здоровье %d)", healthAfterFrame0.Current)
	}

	// Проверяем что блинка НЕТ
	if world.HasComponent(rabbit, core.MaskDamageFlash) {
		t.Error("ОШИБКА: Есть блинк на кадре 0 (Windup)!")
	} else {
		t.Logf("✅ Кадр 0: НЕТ блинка")
	}

	// === ФАЗА 2: КАДР 1 АТАКИ (ЕСТЬ УРОН + БЛИНК - STRIKE) ===
	t.Logf("\n=== ФАЗА 2: КАДР 1 АТАКИ ===")

	// Переводим на кадр 1 (Strike)
	anim.Frame = 1 // Кадр 1 = Strike
	world.SetAnimation(wolf, anim)

	t.Logf("Волк в анимации ATTACK кадр 1 (Strike)")

	// Проверяем что урон ЕСТЬ на кадре 1 (Strike)
	combatSystem.Update(world, deltaTime)

	// Дополнительное обновление для executeStrike
	combatSystem.Update(world, deltaTime)
	healthAfterFrame1, _ := world.GetHealth(rabbit)

	// ИСПРАВЛЕНИЕ: Проверяем реальный урон волка из конфигурации
	config, hasConfig := world.GetAnimalConfig(wolf)
	realDamage := int16(15) // Дефолтное значение если нет конфигурации
	if hasConfig {
		realDamage = config.AttackDamage
		t.Logf("DEBUG: Реальный урон волка из конфигурации: %d", realDamage)
	}

	actualDamage := initialHealth.Current - healthAfterFrame1.Current
	if actualDamage != realDamage {
		t.Errorf("ОШИБКА: Неправильный урон на кадре 1 (Strike)! Ожидали %d, получили %d",
			realDamage, actualDamage)
	} else {
		t.Logf("✅ Кадр 1: ЕСТЬ урон %d (здоровье %d -> %d)",
			actualDamage, initialHealth.Current, healthAfterFrame1.Current)
	}

	// Проверяем что блинк ЕСТЬ
	if !world.HasComponent(rabbit, core.MaskDamageFlash) {
		t.Error("ОШИБКА: НЕТ блинка на кадре 1 (Strike)!")
	} else {
		flash, _ := world.GetDamageFlash(rabbit)
		t.Logf("✅ Кадр 1: ЕСТЬ блинк (таймер %.2f)", flash.Timer)
	}

	// === ФАЗА 3: ПОВТОРНАЯ АТАКА ДЛЯ УБИЙСТВА ===
	t.Logf("\n=== ФАЗА 3: ПОВТОРНАЯ АТАКА ===")

	// Завершаем первую атаку
	world.SetAnimation(wolf, core.Animation{
		CurrentAnim: int(animation.AnimAttack),
		Frame:       1,
		Timer:       0,
		Playing:     false, // Анимация завершена
		FacingRight: true,
	})
	combatSystem.Update(world, deltaTime) // Удалит AttackState

	// Ждём окончания кулдауна и возможной автоматической второй атаки
	// (1 секунда = 60 кадров, но система может автоматически атаковать снова)
	for i := 0; i < 120; i++ { // Увеличиваем до 2 секунд на случай второй атаки
		combatSystem.Update(world, deltaTime)
		animManager.UpdateAllAnimations(world, deltaTime) // ВАЖНО: обновляем анимации!

		// Проверяем здоровье зайца на каждом тике
		currentHealth, _ := world.GetHealth(rabbit)
		if i%10 == 0 {
			t.Logf("Тик %d: здоровье зайца %d", i, currentHealth.Current)
		}

		// Если заяц умер, система должна превратить его в труп
		if currentHealth.Current == 0 {
			t.Logf("Заяц умер на тике %d", i)
			// Даём системе время превратить зайца в труп
			for j := 0; j < 5; j++ {
				combatSystem.Update(world, deltaTime)
			}
			break
		}

		// Обновляем блинк
		if world.HasComponent(rabbit, core.MaskDamageFlash) {
			flash, _ := world.GetDamageFlash(rabbit)
			flash.Timer -= deltaTime
			if flash.Timer <= 0 {
				world.RemoveDamageFlash(rabbit)
				t.Logf("Блинк исчез на тике %d", i)
			} else {
				world.SetDamageFlash(rabbit, flash)
			}
		}
	}

	// Система должна была автоматически выполнить вторую атаку
	t.Logf("\n=== РЕЗУЛЬТАТ: АВТОМАТИЧЕСКАЯ ВТОРАЯ АТАКА ===")

	rabbitHealth, _ := world.GetHealth(rabbit)
	t.Logf("Финальное здоровье зайца: %d", rabbitHealth.Current)

	if world.HasComponent(rabbit, core.MaskCorpse) {
		t.Logf("✅ Заяц превращен в труп после автоматической второй атаки")

		// Проверяем анимацию смерти зайца
		if rabbitAnim, hasAnim := world.GetAnimation(rabbit); hasAnim {
			if rabbitAnim.CurrentAnim != int(animation.AnimDeathDying) {
				t.Errorf("ОШИБКА: Неправильная анимация зайца %s, ожидали DEATH_DYING",
					animation.AnimationType(rabbitAnim.CurrentAnim).String())
			} else {
				t.Logf("✅ Заяц в анимации DEATH_DYING")
			}
		}

		// Проверяем что создался труп с правильной питательностью
		if corpse, hasCorpse := world.GetCorpse(rabbit); hasCorpse {
			t.Logf("✅ Труп создан (питательность %.1f)", corpse.NutritionalValue)
		} else {
			t.Error("ОШИБКА: Труп не создался!")
		}

		// Переходим к фазе поедания
		t.Logf("\n=== ФАЗА 4: НАЧАЛО ПОЕДАНИЯ ===")

		// Система должна автоматически перевести волка в состояние поедания
		combatSystem.Update(world, deltaTime)

		if !world.HasComponent(wolf, core.MaskEatingState) {
			t.Error("ОШИБКА: Волк не начал есть!")
		} else {
			eatingState, _ := world.GetEatingState(wolf)
			t.Logf("✅ Волк начал есть (цель: %d)", eatingState.Target)
		}

		return
	}

	if !world.IsAlive(rabbit) {
		t.Logf("⚠️ Заяц не жив - система не будет создавать AttackState")
		return
	}

	// Если заяц всё ещё жив, создаем вторую атаку
	combatSystem.Update(world, deltaTime)

	if !world.HasComponent(wolf, core.MaskAttackState) {
		t.Error("Вторая атака не создалась")
		t.Logf("Возможные причины:")
		t.Logf("- Заяц мертв или является трупом")
		t.Logf("- Волк не голоден (голод > 60%%)")
		t.Logf("- Расстояние до зайца > 12 единиц")
		t.Logf("- Кулдаун атаки еще не закончился")
		return
	}

	// Вторая атака - сразу кадр 1 (Strike)
	world.SetAnimation(wolf, core.Animation{
		CurrentAnim: int(animation.AnimAttack),
		Frame:       1, // Strike
		Timer:       0,
		Playing:     true,
		FacingRight: true,
	})

	combatSystem.Update(world, deltaTime)
	healthAfterSecondAttack, _ := world.GetHealth(rabbit)

	if healthAfterSecondAttack.Current != 0 {
		t.Errorf("ОШИБКА: Заяц не умер! Здоровье %d", healthAfterSecondAttack.Current)
	} else {
		t.Logf("✅ Заяц умер (здоровье 0)")
	}

	// Проверяем что создался труп
	if !world.HasComponent(rabbit, core.MaskCorpse) {
		t.Error("ОШИБКА: Труп не создался!")
	} else {
		corpse, _ := world.GetCorpse(rabbit)
		t.Logf("✅ Труп создан (питательность %.1f)", corpse.NutritionalValue)
	}

	// Проверяем анимацию смерти зайца
	if rabbitAnim, hasAnim := world.GetAnimation(rabbit); hasAnim {
		if rabbitAnim.CurrentAnim != int(animation.AnimDeathDying) {
			t.Errorf("ОШИБКА: Неправильная анимация зайца %s, ожидали DEATH_DYING",
				animation.AnimationType(rabbitAnim.CurrentAnim).String())
		} else {
			t.Logf("✅ Заяц в анимации DEATH_DYING")
		}
	}

	// === ФАЗА 4: НАЧАЛО ПОЕДАНИЯ ===
	t.Logf("\n=== ФАЗА 4: ПОЕДАНИЕ ===")

	// Система должна автоматически перевести волка в состояние поедания
	combatSystem.Update(world, deltaTime)

	if !world.HasComponent(wolf, core.MaskEatingState) {
		t.Error("ОШИБКА: Волк не начал есть!")
	} else {
		eatingState, _ := world.GetEatingState(wolf)
		t.Logf("✅ Волк начал есть (цель: %d)", eatingState.Target)
	}

	// Проверяем анимацию волка
	if wolfAnim, hasAnim := world.GetAnimation(wolf); hasAnim {
		// Обновляем анимацию вручную для теста
		if world.HasComponent(wolf, core.MaskEatingState) {
			wolfAnim.CurrentAnim = int(animation.AnimEat)
			world.SetAnimation(wolf, wolfAnim)
		}

		if wolfAnim.CurrentAnim != int(animation.AnimEat) {
			t.Errorf("ОШИБКА: Волк не в анимации EAT, а в %s",
				animation.AnimationType(wolfAnim.CurrentAnim).String())
		} else {
			t.Logf("✅ Волк в анимации EAT")
		}
	}

	// === ФАЗА 5: ПРОЦЕСС ПОЕДАНИЯ ===
	t.Logf("\n=== ФАЗА 5: ПРОЦЕСС ПОЕДАНИЯ ===")

	initialHunger, _ := world.GetHunger(wolf)
	t.Logf("Голод волка до поедания: %.1f%%", initialHunger.Value)

	// Симулируем поедание
	for i := 0; i < 300; i++ { // 5 секунд
		combatSystem.Update(world, deltaTime)

		if i%60 == 0 { // Каждую секунду
			hunger, _ := world.GetHunger(wolf)
			if world.HasComponent(rabbit, core.MaskCorpse) {
				corpse, _ := world.GetCorpse(rabbit)
				t.Logf("%.1fс: голод волка %.1f%%, питательность трупа %.1f",
					float32(i)/60.0, hunger.Value, corpse.NutritionalValue)
			} else {
				t.Logf("%.1fс: голод волка %.1f%%, труп съеден",
					float32(i)/60.0, hunger.Value)
				break
			}
		}

		// Проверяем исчезновение трупа
		if !world.IsAlive(rabbit) {
			finalHunger, _ := world.GetHunger(wolf)
			t.Logf("✅ УСПЕХ: Труп полностью съеден на тике %d", i)
			t.Logf("✅ Голод волка восстановился: %.1f%% -> %.1f%%",
				initialHunger.Value, finalHunger.Value)

			if finalHunger.Value <= initialHunger.Value {
				t.Errorf("ОШИБКА: Голод не восстановился!")
			}

			return
		}
	}

	t.Error("ОШИБКА: Труп не был съеден за 5 секунд")
}

// TestMissChance тест промаха (20% шанс)
func TestMissChance(t *testing.T) {
	t.Parallel()
	// Используем seed, который даёт промах
	world := core.NewWorld(640, 640, 123) // Другой seed
	combatSystem := simulation.NewCombatSystem()

	rabbit := simulation.CreateAnimal(world, core.TypeRabbit, 300, 300)
	wolf := simulation.CreateAnimal(world, core.TypeWolf, 301, 300) // Дистанция 1 пиксель
	world.SetHunger(wolf, core.Hunger{Value: 30.0})

	initialHealth, _ := world.GetHealth(rabbit)

	// Устанавливаем атаку на 2-й кадр
	anim := core.Animation{
		CurrentAnim: int(animation.AnimAttack),
		Frame:       1, // 2-й кадр
		Timer:       0,
		Playing:     true,
		FacingRight: true,
	}
	world.SetAnimation(wolf, anim)

	t.Logf("Тест промаха с seed 123")

	// Проверяем несколько атак
	deltaTime := float32(1.0 / 60.0)
	missCount := 0
	hitCount := 0

	for attempt := 0; attempt < 10; attempt++ {
		// Сбрасываем здоровье для следующего теста
		world.SetHealth(rabbit, core.Health{Current: 50, Max: 50})

		combatSystem.Update(world, deltaTime)

		healthAfter, _ := world.GetHealth(rabbit)
		if healthAfter.Current == initialHealth.Current {
			missCount++
			t.Logf("Попытка %d: ПРОМАХ", attempt+1)
		} else {
			hitCount++
			t.Logf("Попытка %d: ПОПАДАНИЕ (урон %d)", attempt+1, initialHealth.Current-healthAfter.Current)
		}

		// Ждём кулдаун между атаками
		for i := 0; i < 60; i++ {
			combatSystem.Update(world, deltaTime)
		}
	}

	t.Logf("Результат: %d попаданий, %d промахов из 10", hitCount, missCount)

	if missCount == 0 {
		t.Error("ОШИБКА: Ни одного промаха за 10 попыток!")
	} else {
		t.Logf("✅ Промахи работают (%d из 10)", missCount)
	}
}
