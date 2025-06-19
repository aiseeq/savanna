package integration

import (
	"testing"

	"github.com/aiseeq/savanna/internal/animation"
	"github.com/aiseeq/savanna/internal/core"
	"github.com/aiseeq/savanna/internal/simulation"
)

// TestStateDrivenCombat проверяет новую архитектуру где каждый кадр соответствует состоянию
//
//nolint:revive // function-length: Комплексный тест state-driven боевой системы
func TestStateDrivenCombat(t *testing.T) {
	t.Parallel()
	world := core.NewWorld(96, 96, 42)
	combatSystem := simulation.NewCombatSystem()

	// Создаём животных рядом (в радиусе атаки 12 пикселей)
	rabbit := simulation.CreateAnimal(world, core.TypeRabbit, 48, 48)
	wolf := simulation.CreateAnimal(world, core.TypeWolf, 50, 48) // Расстояние = 2 пикселя

	// Волк голоден
	world.SetHunger(wolf, core.Hunger{Value: 5.0})

	t.Logf("=== ТЕСТ АРХИТЕКТУРЫ: СОСТОЯНИЕ → АНИМАЦИЯ ===")

	deltaTime := float32(1.0 / 60.0)

	// ФАЗА 1: Волк БЕЗ AttackState - должен найти цель и начать атаку
	t.Logf("\n--- ФАЗА 1: ПОИСК ЦЕЛИ ---")

	// Проверяем что волк пока не атакует
	if world.HasComponent(wolf, core.MaskAttackState) {
		t.Error("❌ Волк не должен иметь AttackState в начале")
	}

	// Обновляем боевую систему - должна создать AttackState
	combatSystem.Update(world, deltaTime)

	// Проверяем что AttackState создан
	if !world.HasComponent(wolf, core.MaskAttackState) {
		t.Fatal("❌ КРИТИЧЕСКАЯ ОШИБКА: AttackState не создан!")
	}

	attackState, hasAttack := world.GetAttackState(wolf)
	if !hasAttack {
		t.Fatal("❌ КРИТИЧЕСКАЯ ОШИБКА: AttackState не получен!")
	}

	t.Logf("✅ AttackState создан:")
	t.Logf("  Цель: %d", attackState.Target)
	t.Logf("  Фаза: %s", attackState.Phase.String())
	t.Logf("  Таймер: %.3f сек", attackState.TotalTimer)

	if attackState.Target != rabbit {
		t.Errorf("❌ Неверная цель: ожидался %d, получен %d", rabbit, attackState.Target)
	}

	if attackState.Phase != core.AttackPhaseWindup {
		t.Errorf("❌ Неверная фаза: ожидался %s, получен %s",
			core.AttackPhaseWindup.String(), attackState.Phase.String())
	}

	// ФАЗА 2: Настроим анимацию ATTACK кадр 0 - состояние Windup
	t.Logf("\n--- ФАЗА 2: КАДР 0 (WINDUP) ---")

	world.SetAnimation(wolf, core.Animation{
		CurrentAnim: int(animation.AnimAttack),
		Frame:       0,
		Timer:       0,
		Playing:     true,
		FacingRight: true,
	})

	// Обновляем боевую систему
	combatSystem.Update(world, deltaTime)

	// Проверяем что фаза все еще Windup
	attackState, _ = world.GetAttackState(wolf)
	if attackState.Phase != core.AttackPhaseWindup {
		t.Errorf("❌ На кадре 0 должна быть фаза Windup, получена %s", attackState.Phase.String())
	}

	t.Logf("✅ Кадр 0: Фаза %s", attackState.Phase.String())

	// Проверяем что урон НЕ нанесен
	initialHealth, _ := world.GetHealth(rabbit)
	if initialHealth.Current != 50 {
		t.Errorf("❌ На кадре 0 урон НЕ должен быть нанесен! Здоровье: %d", initialHealth.Current)
	}

	// ФАЗА 3: Переключаем на кадр 1 - состояние Strike
	t.Logf("\n--- ФАЗА 3: КАДР 1 (STRIKE) ---")

	world.SetAnimation(wolf, core.Animation{
		CurrentAnim: int(animation.AnimAttack),
		Frame:       1,
		Timer:       0,
		Playing:     true,
		FacingRight: true,
	})

	// Проверяем анимацию перед обновлением
	anim, _ := world.GetAnimation(wolf)
	t.Logf("  Перед обновлением: CurrentAnim=%d, Frame=%d, Playing=%t", anim.CurrentAnim, anim.Frame, anim.Playing)

	// Обновляем боевую систему
	combatSystem.Update(world, deltaTime)

	// Проверяем что фаза переключилась на Strike
	attackState, _ = world.GetAttackState(wolf)
	if attackState.Phase != core.AttackPhaseStrike {
		t.Errorf("❌ На кадре 1 должна быть фаза Strike, получена %s", attackState.Phase.String())
	}

	t.Logf("✅ Кадр 1: Фаза %s", attackState.Phase.String())
	t.Logf("  HasStruck (после перехода): %t", attackState.HasStruck)

	// Обновляем еще раз чтобы executeStrike сработал
	combatSystem.Update(world, deltaTime)

	attackState, _ = world.GetAttackState(wolf)
	t.Logf("  HasStruck (после удара): %t", attackState.HasStruck)

	// Проверяем что урон НАНЕСЕН
	currentHealth, _ := world.GetHealth(rabbit)
	if currentHealth.Current >= initialHealth.Current {
		// Дополнительная отладка
		wolfPos, _ := world.GetPosition(wolf)
		rabbitPos, _ := world.GetPosition(rabbit)
		distance := (wolfPos.X-rabbitPos.X)*(wolfPos.X-rabbitPos.X) + (wolfPos.Y-rabbitPos.Y)*(wolfPos.Y-rabbitPos.Y)
		t.Logf("  Позиция волка: (%.1f, %.1f)", wolfPos.X, wolfPos.Y)
		t.Logf("  Позиция зайца: (%.1f, %.1f)", rabbitPos.X, rabbitPos.Y)
		t.Logf("  Расстояние: %.1f (лимит: %.1f)", distance, 12.0*12.0)

		t.Errorf("❌ На кадре 1 урон ДОЛЖЕН быть нанесен! Здоровье: %d -> %d",
			initialHealth.Current, currentHealth.Current)
	} else {
		t.Logf("✅ Урон нанесен: %d -> %d", initialHealth.Current, currentHealth.Current)
	}

	// Проверяем что HasStruck = true
	if !attackState.HasStruck {
		t.Error("❌ HasStruck должно быть true после нанесения урона")
	}

	// ФАЗА 4: Завершение анимации - AttackState должен удалиться
	t.Logf("\n--- ФАЗА 4: ЗАВЕРШЕНИЕ АТАКИ ---")

	world.SetAnimation(wolf, core.Animation{
		CurrentAnim: int(animation.AnimAttack),
		Frame:       1,
		Timer:       0,
		Playing:     false, // Анимация завершена
		FacingRight: true,
	})

	// Обновляем боевую систему
	combatSystem.Update(world, deltaTime)

	// Проверяем что AttackState удален
	if world.HasComponent(wolf, core.MaskAttackState) {
		t.Error("❌ AttackState должен быть удален после завершения анимации")
	} else {
		t.Logf("✅ AttackState удален после завершения анимации")
	}

	// ИТОГОВАЯ ПРОВЕРКА АРХИТЕКТУРЫ
	t.Logf("\n=== ПРОВЕРКА АРХИТЕКТУРЫ ===")
	t.Logf("✅ Кадр 0 → Состояние Windup → НЕТ урона")
	t.Logf("✅ Кадр 1 → Состояние Strike → ЕСТЬ урон")
	t.Logf("✅ Завершение анимации → Удаление AttackState")
	t.Logf("✅ Архитектура СОСТОЯНИЕ ↔ АНИМАЦИЯ работает!")
}
