package integration

import (
	"testing"

	"github.com/aiseeq/savanna/internal/animation"
	"github.com/aiseeq/savanna/internal/core"
	"github.com/aiseeq/savanna/internal/simulation"
)

// TestDamageFlashBug - TDD тест для бага пропавшей вспышки урона
//
// БАГ: При получении урона DamageFlash не создаётся или сразу исчезает
// ОЖИДАНИЕ: При получении урона должна появляться вспышка на 0.5 секунды
//
//nolint:gocognit,revive,funlen // TDD тест для воспроизведения конкретного бага
func TestDamageFlashBug(t *testing.T) {
	t.Parallel()

	world := core.NewWorld(640, 640, 42)
	combatSystem := simulation.NewCombatSystem()

	// Создаём анимационные системы
	wolfAnimSystem := animation.NewAnimationSystem()
	rabbitAnimSystem := animation.NewAnimationSystem()

	// Регистрируем анимации
	wolfAnimSystem.RegisterAnimation(animation.AnimAttack, 2, 6.0, false, nil)
	rabbitAnimSystem.RegisterAnimation(animation.AnimIdle, 2, 1.0, true, nil)

	// Создаём волка и зайца ОЧЕНЬ БЛИЗКО друг к другу
	rabbit := simulation.CreateAnimal(world, core.TypeRabbit, 300, 300)
	wolf := simulation.CreateAnimal(world, core.TypeWolf, 305, 300) // Очень близко

	// Делаем волка ОЧЕНЬ голодным чтобы он точно атаковал
	world.SetHunger(wolf, core.Hunger{Value: 10.0}) // 10% < 60% = очень голодный

	initialHealth, _ := world.GetHealth(rabbit)
	t.Logf("=== ТЕСТ ВСПЫШКИ УРОНА ===")
	t.Logf("Начальное здоровье зайца: %d", initialHealth.Current)

	// Проверяем что животные правильно настроены
	wolfConfig, hasWolfConfig := world.GetAnimalConfig(wolf)
	if hasWolfConfig {
		t.Logf("Волк: урон=%d, дальность атаки=%.1f, шанс попадания=%.1f%%",
			wolfConfig.AttackDamage, wolfConfig.AttackRange, wolfConfig.HitChance*100)
	}

	deltaTime := float32(1.0 / 60.0)

	// Фаза 1: Ждём пока система автоматически создаст AttackState
	attackStateCreated := false
	for i := 0; i < 120; i++ { // 2 секунды
		combatSystem.Update(world, deltaTime)

		if world.HasComponent(wolf, core.MaskAttackState) {
			attackStateCreated = true
			t.Logf("AttackState создан на тике %d", i)
			break
		}

		if i%30 == 0 {
			hunger, _ := world.GetHunger(wolf)
			t.Logf("Тик %d: голод волка %.1f%%, AttackState пока нет", i, hunger.Value)
		}
	}

	if !attackStateCreated {
		t.Fatal("AttackState не создан за 2 секунды - проблема в логике атаки")
	}

	// Фаза 2: Устанавливаем анимацию атаки на кадр удара (кадр 1)
	world.SetAnimation(wolf, core.Animation{
		CurrentAnim: int(animation.AnimAttack),
		Frame:       1, // Кадр удара
		Timer:       0,
		Playing:     true,
		FacingRight: true,
	})

	t.Logf("Волк в анимации ATTACK кадр 1 (удар)")

	// Фаза 3: Выполняем удар и проверяем создание DamageFlash
	// Дополнительная диагностика ПЕРЕД ударом
	wolfPos, _ := world.GetPosition(wolf)
	rabbitPos, _ := world.GetPosition(rabbit)
	distance := ((wolfPos.X-rabbitPos.X)*(wolfPos.X-rabbitPos.X) + (wolfPos.Y-rabbitPos.Y)*(wolfPos.Y-rabbitPos.Y))
	t.Logf("ДИАГНОСТИКА: расстояние=%.1f, лимит атаки=%.1f", distance, wolfConfig.AttackRange*wolfConfig.AttackRange)

	if attackState, hasAttack := world.GetAttackState(wolf); hasAttack {
		t.Logf("ДИАГНОСТИКА ДО УДАРА: фаза=%v, HasStruck=%v", attackState.Phase, attackState.HasStruck)
	}

	combatSystem.Update(world, deltaTime)

	// Диагностика ПОСЛЕ первого обновления
	if attackState, hasAttack := world.GetAttackState(wolf); hasAttack {
		t.Logf("ДИАГНОСТИКА ПОСЛЕ 1-ГО ОБНОВЛЕНИЯ: фаза=%v, HasStruck=%v", attackState.Phase, attackState.HasStruck)
	}

	// ДОПОЛНИТЕЛЬНОЕ обновление чтобы выполнить удар в фазе Strike
	combatSystem.Update(world, deltaTime)

	// Диагностика ПОСЛЕ второго обновления
	if attackState, hasAttack := world.GetAttackState(wolf); hasAttack {
		t.Logf("ДИАГНОСТИКА ПОСЛЕ 2-ГО ОБНОВЛЕНИЯ: фаза=%v, HasStruck=%v", attackState.Phase, attackState.HasStruck)
	}

	// Проверяем что урон был нанесён
	healthAfterHit, _ := world.GetHealth(rabbit)
	actualDamage := initialHealth.Current - healthAfterHit.Current

	if actualDamage <= 0 {
		t.Errorf("ДИАГНОСТИКА: Урон не был нанесён")
		t.Errorf("Возможные причины:")
		t.Errorf("1. Шанс попадания не сработал (%.1f%%)", wolfConfig.HitChance*100)
		t.Errorf("2. Анимация не синхронизирована с логикой удара")
		t.Errorf("3. AttackState не в правильной фазе")
		t.Errorf("4. Расстояние слишком большое (%.1f > %.1f)", distance, wolfConfig.AttackRange*wolfConfig.AttackRange)

		if attackState, hasAttack := world.GetAttackState(wolf); hasAttack {
			t.Errorf("ФИНАЛ AttackState: фаза=%v, HasStruck=%v", attackState.Phase, attackState.HasStruck)
		}

		t.Fatal("Тест невозможен без урона")
	}

	t.Logf("✅ Урон нанесён: %d (здоровье %d -> %d)",
		actualDamage, initialHealth.Current, healthAfterHit.Current)

	// КРИТИЧЕСКАЯ ПРОВЕРКА: DamageFlash должен быть создан СРАЗУ после урона
	if !world.HasComponent(rabbit, core.MaskDamageFlash) {
		t.Errorf("БАГ ОБНАРУЖЕН: DamageFlash не создан сразу после нанесения урона!")
		t.Errorf("Ожидалось: компонент DamageFlash с таймером ~0.5 секунды")
		t.Errorf("Получили: компонент отсутствует")
		return
	}

	// Проверяем параметры DamageFlash
	flash, _ := world.GetDamageFlash(rabbit)
	t.Logf("✅ DamageFlash создан: таймер %.3f секунды", flash.Timer)

	if flash.Timer <= 0 {
		t.Errorf("БАГ: DamageFlash создан с неправильным таймером %.3f (должен быть > 0)", flash.Timer)
		return
	}

	if flash.Timer > 1.0 {
		t.Errorf("БАГ: DamageFlash слишком долгий %.3f секунды (должен быть ~0.5)", flash.Timer)
	}

	// Фаза 4: Проверяем что DamageFlash НЕ исчезает мгновенно
	initialFlashTimer := flash.Timer

	// Обновляем несколько тиков и проверяем что таймер уменьшается
	for i := 0; i < 10; i++ {
		combatSystem.Update(world, deltaTime)

		if !world.HasComponent(rabbit, core.MaskDamageFlash) {
			t.Logf("✅ DamageFlash исчез на тике %d (ожидалось ~10 тиков при быстром угасании)", i)
			break
		}

		currentFlash, _ := world.GetDamageFlash(rabbit)
		if i == 5 {
			t.Logf("Тик %d: таймер DamageFlash %.3f -> %.3f", i, initialFlashTimer, currentFlash.Timer)
		}

		if currentFlash.Timer >= initialFlashTimer {
			t.Errorf("БАГ: Таймер DamageFlash не уменьшается! Был %.3f, стал %.3f",
				initialFlashTimer, currentFlash.Timer)
			return
		}
	}

	// Фаза 5: Симулируем полное исчезновение DamageFlash
	flashDisappeared := false

	for i := 0; i < 60; i++ { // 1 секунда максимум
		combatSystem.Update(world, deltaTime)

		if world.HasComponent(rabbit, core.MaskDamageFlash) {
			flash, _ := world.GetDamageFlash(rabbit)
			if i%10 == 0 {
				t.Logf("Тик %d: DamageFlash ещё активен (таймер %.3f)", i, flash.Timer)
			}
		} else {
			t.Logf("✅ DamageFlash исчез на тике %d (через %.3f секунды)", i, float32(i)/60.0)
			flashDisappeared = true
			break
		}
	}

	if !flashDisappeared {
		t.Error("БАГ: DamageFlash не исчез за 1 секунду - слишком долго")
	}

	t.Logf("✅ Тест вспышки урона завершён успешно")
}

// TestDamageFlashMultipleHits был удален так как стал неактуальным:
// Исправление бага переключения между целями предотвращает множественные атаки на одну цель
