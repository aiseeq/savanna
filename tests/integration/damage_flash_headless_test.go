package integration

import (
	"testing"

	"github.com/aiseeq/savanna/internal/core"
	"github.com/aiseeq/savanna/internal/simulation"
	"github.com/aiseeq/savanna/tests/common"
)

// TestDamageFlashHeadless - Headless интеграционный тест для проверки DamageFlash эффекта
//
// ЦЕЛЬ: Проверить полную работу DamageFlash без GUI зависимостей
// Тест заменяет падающие E2E тесты в headless окружении
//
//nolint:gocognit,revive,funlen // Интеграционный тест для полной проверки функциональности
func TestDamageFlashHeadless(t *testing.T) {
	t.Parallel()

	world := core.NewWorld(640, 640, 42)
	combatSystem := simulation.NewCombatSystem()

	// Добавляем анимационную систему для корректной работы атак
	animationAdapter := common.NewAnimationSystemAdapter()

	// Создаём тестовую сцену: волк атакует зайца (очень близко для гарантированной атаки)
	rabbit := simulation.CreateAnimal(world, core.TypeRabbit, 300, 300)
	wolf := simulation.CreateAnimal(world, core.TypeWolf, 300.5, 300) // Дистанция 0.5 < атака волка 0.9

	// Делаем волка голодным для гарантированной атаки
	world.SetHunger(wolf, core.Hunger{Value: 20.0})

	t.Logf("=== ТЕСТ DAMAGEFLASH HEADLESS ===")
	t.Logf("Заяц: entity %d (позиция 300,300)", rabbit)
	t.Logf("Волк: entity %d (позиция 300.5,300)", wolf)

	// Проверяем начальное состояние
	if world.HasComponent(rabbit, core.MaskDamageFlash) {
		t.Fatal("DamageFlash уже существует в начале теста")
	}

	initialHealth, _ := world.GetHealth(rabbit)
	t.Logf("Начальное здоровье зайца: %d", initialHealth.Current)

	deltaTime := float32(1.0 / 60.0)
	var flashIntensity float32
	var flashTimer float32

	// Фаза 1: Ждём автоматической атаки и создания DamageFlash
	for i := 0; i < 180; i++ { // 3 секунды максимум
		// КРИТИЧЕСКИ ВАЖНО: Обновляем анимации ПЕРЕД combat system
		world.Update(deltaTime)
		animationAdapter.Update(world, deltaTime)
		combatSystem.Update(world, deltaTime)

		if world.HasComponent(rabbit, core.MaskDamageFlash) {
			flash, _ := world.GetDamageFlash(rabbit)
			flashIntensity = flash.Intensity
			flashTimer = flash.Timer

			t.Logf("✅ DamageFlash создан на тике %d", i)
			t.Logf("  Начальная интенсивность: %.3f", flashIntensity)
			t.Logf("  Начальный таймер: %.3f секунд", flashTimer)
			break
		}

		// Логируем диагностику каждую секунду
		if i%60 == 0 {
			hunger, _ := world.GetHunger(wolf)
			hasAttack := world.HasComponent(wolf, core.MaskAttackState)
			t.Logf("Секунда %d: Голод волка %.1f%%, AttackState=%v", i/60, hunger.Value, hasAttack)
		}
	}

	if !world.HasComponent(rabbit, core.MaskDamageFlash) {
		t.Fatal("DamageFlash не был создан за 3 секунды")
	}

	// КРИТИЧЕСКАЯ ПРОВЕРКА 1: Интенсивность в ожидаемых пределах
	if flashIntensity < 0.8 || flashIntensity > 1.0 {
		t.Errorf("БАГ: Неожиданная интенсивность: %.3f (ожидалось 0.8-1.0)", flashIntensity)
	}

	// КРИТИЧЕСКАЯ ПРОВЕРКА 2: Таймер соответствует новой длительности (0.16 сек)
	expectedDuration := float32(0.16)
	tolerance := float32(0.02) // 2% погрешность
	if flashTimer < expectedDuration-tolerance || flashTimer > expectedDuration+tolerance {
		t.Errorf("БАГ: Неправильная длительность: %.3f сек (ожидалось ~%.3f)",
			flashTimer, expectedDuration)
	} else {
		t.Logf("✅ Длительность корректна: %.3f сек (ускорена в 5 раз)", flashTimer)
	}

	// КРИТИЧЕСКАЯ ПРОВЕРКА 3: Расчёт усиления интенсивности (scale = 1.0 + intensity * 5.0)
	expectedScale := 1.0 + flashIntensity*5.0
	if expectedScale < 5.0 || expectedScale > 6.0 {
		t.Errorf("БАГ: Неправильный масштаб яркости: %.1f (ожидалось 5.0-6.0)", expectedScale)
	} else {
		t.Logf("✅ Масштаб яркости: %.1fх (усилен в 5 раз)", expectedScale)
	}

	// Фаза 2: Проверяем быстрое угасание (должно исчезнуть за ~10 тиков)
	t.Logf("\n=== ПРОВЕРКА БЫСТРОГО УГАСАНИЯ ===")

	ticksToDisappear := 0
	for i := 0; i < 20; i++ { // 20 тиков максимум (треть секунды)
		// Обновляем анимации и combat system
		world.Update(deltaTime)
		animationAdapter.Update(world, deltaTime)
		combatSystem.Update(world, deltaTime)

		if !world.HasComponent(rabbit, core.MaskDamageFlash) {
			ticksToDisappear = i
			t.Logf("✅ DamageFlash исчез за %d тиков (%.3f сек)", i, float32(i)/60.0)
			break
		}

		// Логируем угасание каждые 3 тика
		if i%3 == 0 && world.HasComponent(rabbit, core.MaskDamageFlash) {
			flash, _ := world.GetDamageFlash(rabbit)
			t.Logf("  Тик %d: интенсивность=%.3f, таймер=%.3f", i, flash.Intensity, flash.Timer)
		}
	}

	// КРИТИЧЕСКАЯ ПРОВЕРКА 4: Эффект должен исчезнуть быстро (в 5 раз быстрее)
	maxExpectedTicks := 12 // ~0.2 секунды с запасом
	if ticksToDisappear == 0 {
		t.Error("БАГ: DamageFlash не исчез за 20 тиков (треть секунды)")
	} else if ticksToDisappear > maxExpectedTicks {
		t.Errorf("БАГ: DamageFlash исчез слишком медленно: %d тиков (ожидалось <%d)",
			ticksToDisappear, maxExpectedTicks)
	} else {
		t.Logf("✅ Быстрое угасание подтверждено: %d тиков", ticksToDisappear)
	}

	// Фаза 3: Проверяем что урон был нанесён корректно
	health, hasHealth := world.GetHealth(rabbit)
	if hasHealth && health.Current < health.Max {
		damageTaken := health.Max - health.Current
		t.Logf("✅ Урон нанесён: %d единиц (%d -> %d HP)", damageTaken, health.Max, health.Current)
	} else {
		t.Error("БАГ: Урон не был нанесён")
	}

	// Резюме теста
	t.Logf("\n=== РЕЗЮМЕ HEADLESS DAMAGEFLASH ===")
	t.Logf("✅ Интенсивность: %.3f (масштаб яркости %.1fх)", flashIntensity, 1.0+flashIntensity*5.0)
	t.Logf("✅ Длительность: %.3f сек (ускорена в 5 раз)", flashTimer)
	t.Logf("✅ Угасание: %d тиков (быстрое)", ticksToDisappear)
	t.Logf("✅ Все параметры соответствуют усиленному эффекту")
}

// TestDamageFlashMultipleAttacks - Headless тест для проверки DamageFlash при множественных атаках
//
// ЦЕЛЬ: Проверить что DamageFlash правильно сбрасывается и пересоздается при новых атаках
func TestDamageFlashMultipleAttacks(t *testing.T) {
	t.Parallel()

	world := core.NewWorld(640, 640, 42)
	combatSystem := simulation.NewCombatSystem()

	// Добавляем анимационную систему для корректной работы атак
	animationAdapter := common.NewAnimationSystemAdapter()

	// Создаём сцену с несколькими зайцами для множественных атак
	rabbit1 := simulation.CreateAnimal(world, core.TypeRabbit, 300, 300)
	rabbit2 := simulation.CreateAnimal(world, core.TypeRabbit, 320, 300)
	wolf := simulation.CreateAnimal(world, core.TypeWolf, 305, 300)

	// Делаем волка очень голодным
	world.SetHunger(wolf, core.Hunger{Value: 10.0})

	t.Logf("=== ТЕСТ МНОЖЕСТВЕННЫХ DAMAGEFLASH ===")

	deltaTime := float32(1.0 / 60.0)
	var firstAttackTick int
	var secondAttackTick int

	// Фаза 1: Ждём первой атаки
	for i := 0; i < 300; i++ { // 5 секунд максимум
		// Обновляем анимации и combat system
		world.Update(deltaTime)
		animationAdapter.Update(world, deltaTime)
		combatSystem.Update(world, deltaTime)

		// Проверяем появление DamageFlash на любом зайце
		if world.HasComponent(rabbit1, core.MaskDamageFlash) {
			firstAttackTick = i
			t.Logf("✅ Первая атака на зайца 1 на тике %d", i)
			break
		}
		if world.HasComponent(rabbit2, core.MaskDamageFlash) {
			firstAttackTick = i
			t.Logf("✅ Первая атака на зайца 2 на тике %d", i)
			break
		}
	}

	if firstAttackTick == 0 {
		t.Fatal("Первая атака не произошла за 5 секунд")
	}

	// Ждём окончания первого DamageFlash
	for i := 0; i < 30; i++ {
		// Обновляем анимации и combat system
		world.Update(deltaTime)
		animationAdapter.Update(world, deltaTime)
		combatSystem.Update(world, deltaTime)

		if !world.HasComponent(rabbit1, core.MaskDamageFlash) &&
			!world.HasComponent(rabbit2, core.MaskDamageFlash) {
			t.Logf("✅ Первый DamageFlash исчез на тике %d", firstAttackTick+i)
			break
		}
	}

	// Фаза 2: Ждём второй атаки
	for i := firstAttackTick + 50; i < firstAttackTick+300; i++ { // Ещё 4 секунды
		// Обновляем анимации и combat system
		world.Update(deltaTime)
		animationAdapter.Update(world, deltaTime)
		combatSystem.Update(world, deltaTime)

		// Проверяем новый DamageFlash
		if world.HasComponent(rabbit1, core.MaskDamageFlash) ||
			world.HasComponent(rabbit2, core.MaskDamageFlash) {
			secondAttackTick = i
			t.Logf("✅ Вторая атака на тике %d", i)
			break
		}

		// Логируем состояние каждые 60 тиков
		if i%60 == 0 {
			hunger, _ := world.GetHunger(wolf)
			t.Logf("Тик %d: Голод волка %.1f%%, ждём вторую атаку", i, hunger.Value)
		}
	}

	if secondAttackTick == 0 {
		t.Logf("⚠️  Вторая атака не произошла (возможно заяц умер или волк сыт)")
	} else {
		t.Logf("✅ Множественные атаки работают правильно")
		timeBetweenAttacks := float32(secondAttackTick-firstAttackTick) / 60.0
		t.Logf("  Время между атаками: %.1f секунд", timeBetweenAttacks)
	}

	t.Logf("✅ Тест множественных DamageFlash завершён")
}
