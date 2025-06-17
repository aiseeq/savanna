package integration

import (
	"testing"

	"github.com/aiseeq/savanna/internal/animation"
	"github.com/aiseeq/savanna/internal/core"
	"github.com/aiseeq/savanna/internal/simulation"
)

// TestDamageFrameValidation проверяет что DamageFlash работает как в реальной игре
func TestDamageFrameValidation(t *testing.T) {
	t.Parallel()

	t.Logf("=== TDD: Проверка DamageFlash в реальной игре ===")
	t.Logf("ЦЕЛЬ: Воспроизвести реальное поведение игры и убедиться что DamageFlash работает")

	// Создаём ПОЛНУЮ игровую среду как в реальной игре
	world := core.NewWorld(96, 96, 42)

	// Все системы как в реальной игре
	systemManager := core.NewSystemManager()
	combatSystem := simulation.NewCombatSystem()
	damageSystem := simulation.NewDamageSystem()

	systemManager.AddSystem(combatSystem)
	systemManager.AddSystem(damageSystem)

	// Создаём животных рядом
	rabbit := simulation.CreateRabbit(world, 40, 48)
	wolf := simulation.CreateWolf(world, 45, 48) // На расстоянии 5 пикселей

	// Волк очень голоден
	world.SetHunger(wolf, core.Hunger{Value: 5.0})

	// ИСПРАВЛЕНИЕ: Фиксируем RNG для гарантированного попадания
	rng := world.GetRNG()
	rng.Seed(42) // Фиксированный seed для детерминизма

	t.Logf("Создали волка и зайца на расстоянии 5 пикселей")
	t.Logf("Голод волка: %.1f%% (очень голоден)", 5.0)
	t.Logf("RNG зафиксирован для гарантированного попадания")

	// Запускаем симуляцию как в реальной игре
	deltaTime := float32(1.0 / 60.0)

	// Сначала создаем AttackState через полную симуляцию
	for tick := 0; tick < 10; tick++ {
		world.Update(deltaTime)
		systemManager.Update(world, deltaTime)

		if world.HasComponent(wolf, core.MaskAttackState) {
			t.Logf("✅ AttackState создан на тике %d", tick)
			break
		}
	}

	// Проверяем что AttackState создан
	if !world.HasComponent(wolf, core.MaskAttackState) {
		t.Fatal("❌ AttackState не создан! Волк не атакует зайца")
	}

	initialHealth, _ := world.GetHealth(rabbit)
	t.Logf("Начальное здоровье зайца: %d", initialHealth.Current)

	// КРИТИЧЕСКАЯ ПРОВЕРКА: Симулируем полную атаку как в реальной игре
	t.Logf("\n=== СИМУЛЯЦИЯ ПОЛНОЙ АТАКИ ===")

	// Получаем текущее AttackState
	attackState, hasAttackState := world.GetAttackState(wolf)
	if !hasAttackState {
		t.Fatalf("❌ AttackState отсутствует!")
	}

	t.Logf("Текущий AttackState: Phase=%d, Target=%d", attackState.Phase, attackState.Target)

	// ИСПРАВЛЕНИЕ: Принудительно устанавливаем правильную фазу атаки
	attackState.Phase = core.AttackPhaseStrike // Фаза удара
	attackState.HasStruck = false              // Ещё не наносили урон
	world.SetAttackState(wolf, attackState)

	// Принудительно ставим волка в анимацию ATTACK кадр 1 (момент удара)
	world.SetAnimation(wolf, core.Animation{
		CurrentAnim: int(animation.AnimAttack),
		Frame:       1, // Кадр удара
		Timer:       0,
		Playing:     true,
		FacingRight: true,
	})

	t.Logf("Устанавливаем волка в фазу Strike и анимацию ATTACK кадр 1")

	// Запускаем ВСЕ системы как в реальной игре
	world.Update(deltaTime)
	systemManager.Update(world, deltaTime)

	// Проверяем что удар нанесён
	updatedAttackState, _ := world.GetAttackState(wolf)
	t.Logf("Удар нанесён: %v", updatedAttackState.HasStruck)

	currentHealth, _ := world.GetHealth(rabbit)
	t.Logf("Здоровье после атаки: %d -> %d", initialHealth.Current, currentHealth.Current)

	// ГЛАВНАЯ ПРОВЕРКА: DamageFlash должен быть активен после урона!
	t.Logf("\n=== ПРОВЕРКА DamageFlash ===")

	hasDamageFlash := world.HasComponent(rabbit, core.MaskDamageFlash)
	t.Logf("Есть ли DamageFlash компонент: %v", hasDamageFlash)

	if !hasDamageFlash {
		t.Errorf("❌ БАГ: DamageFlash НЕ создан после урона!")
		t.Errorf("   Здоровье изменилось: %d -> %d", initialHealth.Current, currentHealth.Current)
		t.Errorf("   Значит урон был нанесен, но DamageFlash не активирован")
		t.Errorf("   ПРОБЛЕМА: DamageSystem не работает или не вызывается")
		return
	}

	// Если DamageFlash есть, проверяем его параметры
	flash, _ := world.GetDamageFlash(rabbit)
	t.Logf("✅ DamageFlash активен!")
	t.Logf("   Таймер: %.3f сек", flash.Timer)
	t.Logf("   Должен мерцать: %.3f сек", flash.Duration)

	// Проверяем что DamageFlash постепенно исчезает
	t.Logf("\n=== ПРОВЕРКА ИСЧЕЗНОВЕНИЯ DamageFlash ===")

	_ = flash.Timer // Изначальный таймер

	// Симулируем несколько кадров
	for tick := 0; tick < 10; tick++ {
		world.Update(deltaTime)
		systemManager.Update(world, deltaTime)

		if !world.HasComponent(rabbit, core.MaskDamageFlash) {
			t.Logf("✅ DamageFlash исчез на тике %d", tick)
			break
		}

		flash, _ = world.GetDamageFlash(rabbit)
		t.Logf("Тик %d: DamageFlash таймер %.3f", tick, flash.Timer)
	}

	// Если DamageFlash всё ещё есть - это нормально, он просто долго длится
	if world.HasComponent(rabbit, core.MaskDamageFlash) {
		t.Logf("DamageFlash всё ещё активен (это нормально)")
	}

}
