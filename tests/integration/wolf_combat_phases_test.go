package integration

import (
	"testing"

	"github.com/aiseeq/savanna/internal/animation"
	"github.com/aiseeq/savanna/internal/core"
	"github.com/aiseeq/savanna/internal/simulation"
)

// TestAttackPhaseWindup тестирует фазу замаха атаки (без урона)
func TestAttackPhaseWindup(t *testing.T) {
	t.Parallel()
	world := core.NewWorld(640, 640, 42)
	combatSystem := simulation.NewCombatSystem()

	// Создаём зайца и волка рядом
	rabbit := simulation.CreateRabbit(world, 300, 300)
	wolf := simulation.CreateWolf(world, 310, 300)

	// Делаем волка голодным
	world.SetHunger(wolf, core.Hunger{Value: 30.0})

	initialHealth, _ := world.GetHealth(rabbit)
	deltaTime := float32(1.0 / 60.0)

	// Создаем AttackState
	combatSystem.Update(world, deltaTime)
	if !world.HasComponent(wolf, core.MaskAttackState) {
		t.Fatal("AttackState не создан")
	}

	// Устанавливаем анимацию атаки на кадр 0 (Windup)
	world.SetAnimation(wolf, core.Animation{
		CurrentAnim: int(animation.AnimAttack),
		Frame:       0, // Кадр 0 = Windup
		Timer:       0,
		Playing:     true,
		FacingRight: true,
	})

	// Обновляем боевую систему
	combatSystem.Update(world, deltaTime)

	// Проверяем что урона НЕТ на кадре 0 (Windup)
	healthAfterWindup, _ := world.GetHealth(rabbit)
	if healthAfterWindup.Current != initialHealth.Current {
		t.Errorf("Урон нанесен на кадре 0 (Windup), а не должен: %d -> %d",
			initialHealth.Current, healthAfterWindup.Current)
	}

	// Проверяем что блинка НЕТ
	if world.HasComponent(rabbit, core.MaskDamageFlash) {
		t.Error("Блинк появился на кадре 0 (Windup), а не должен")
	}

	t.Log("✅ Фаза Windup работает корректно - урон не наносится")
}

// TestAttackPhaseStrike тестирует фазу удара (с уроном)
func TestAttackPhaseStrike(t *testing.T) {
	t.Parallel()
	world := core.NewWorld(640, 640, 42)
	combatSystem := simulation.NewCombatSystem()

	// Создаём зайца и волка рядом
	rabbit := simulation.CreateRabbit(world, 300, 300)
	wolf := simulation.CreateWolf(world, 310, 300)

	// Делаем волка голодным
	world.SetHunger(wolf, core.Hunger{Value: 30.0})

	initialHealth, _ := world.GetHealth(rabbit)
	deltaTime := float32(1.0 / 60.0)

	// Создаем AttackState и проходим фазу Windup
	combatSystem.Update(world, deltaTime)
	if !world.HasComponent(wolf, core.MaskAttackState) {
		t.Fatal("AttackState не создан")
	}

	// Сначала устанавливаем кадр 0 (Windup)
	world.SetAnimation(wolf, core.Animation{
		CurrentAnim: int(animation.AnimAttack),
		Frame:       0, // Кадр 0 = Windup
		Timer:       0,
		Playing:     true,
		FacingRight: true,
	})
	combatSystem.Update(world, deltaTime)

	// Теперь переключаем анимацию на кадр 1 (Strike)
	world.SetAnimation(wolf, core.Animation{
		CurrentAnim: int(animation.AnimAttack),
		Frame:       1, // Кадр 1 = Strike
		Timer:       0,
		Playing:     true,
		FacingRight: true,
	})

	// Обновляем боевую систему для нанесения урона
	combatSystem.Update(world, deltaTime)

	// Может потребоваться дополнительное обновление для executeStrike
	combatSystem.Update(world, deltaTime)

	// Отладка состояния атаки
	if attackState, hasAttack := world.GetAttackState(wolf); hasAttack {
		t.Logf("AttackState после Strike: Phase=%d, HasStruck=%t", int(attackState.Phase), attackState.HasStruck)
	} else {
		t.Log("AttackState отсутствует после Strike")
	}

	// Проверяем что урон НАНЕСЕН на кадре 1 (Strike)
	healthAfterStrike, _ := world.GetHealth(rabbit)
	if healthAfterStrike.Current >= initialHealth.Current {
		t.Errorf("Урон НЕ нанесен на кадре 1 (Strike): здоровье %d -> %d",
			initialHealth.Current, healthAfterStrike.Current)
	}

	// Проверяем что блинк ЕСТЬ
	if !world.HasComponent(rabbit, core.MaskDamageFlash) {
		t.Error("Блинк НЕ появился на кадре 1 (Strike), а должен")
	} else {
		flash, _ := world.GetDamageFlash(rabbit)
		if flash.Timer <= 0 {
			t.Error("Таймер блинка должен быть > 0")
		}
	}

	t.Logf("✅ Фаза Strike работает корректно - урон %d HP",
		initialHealth.Current-healthAfterStrike.Current)
}

// TestMultipleAttacksUntilDeath тестирует многократные атаки до смерти зайца
func TestMultipleAttacksUntilDeath(t *testing.T) {
	t.Parallel()
	world := core.NewWorld(640, 640, 42)
	combatSystem := simulation.NewCombatSystem()

	// Создаём зайца и волка рядом БЕЗ анимаций (для быстрого тестирования)
	rabbit := simulation.CreateRabbit(world, 300, 300)
	wolf := simulation.CreateWolf(world, 310, 300)

	// Убираем анимации для таймер-режима
	world.RemoveAnimation(rabbit)
	world.RemoveAnimation(wolf)

	// Делаем волка голодным
	world.SetHunger(wolf, core.Hunger{Value: 30.0})

	initialHealth, _ := world.GetHealth(rabbit)
	deltaTime := float32(1.0 / 60.0)

	// Симулируем до смерти зайца (максимум 300 тиков = 5 секунд)
	for i := 0; i < 300; i++ {
		combatSystem.Update(world, deltaTime)

		health, _ := world.GetHealth(rabbit)
		if health.Current == 0 {
			t.Logf("✅ Заяц умер на тике %d после %d атак",
				i, (initialHealth.Current+24)/25) // 25 урона за атаку

			// Проверяем что труп создан
			if !world.HasComponent(rabbit, core.MaskCorpse) {
				t.Error("Труп не создан после смерти")
			}
			return
		}
	}

	// Если дошли сюда - заяц не умер
	finalHealth, _ := world.GetHealth(rabbit)
	t.Errorf("Заяц не умер за 300 тиков: здоровье %d -> %d",
		initialHealth.Current, finalHealth.Current)
}
