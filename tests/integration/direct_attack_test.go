package integration

import (
	"testing"

	"github.com/aiseeq/savanna/config"
	"github.com/aiseeq/savanna/internal/animation"
	"github.com/aiseeq/savanna/internal/core"
	"github.com/aiseeq/savanna/internal/simulation"
)

// TestDirectAttack тестирует прямую атаку волка на зайца
func TestDirectAttack(t *testing.T) {
	t.Parallel()
	cfg := config.LoadDefaultConfig()
	cfg.World.Size = 20

	worldSizePixels := float32(cfg.World.Size * 32)
	world := core.NewWorld(worldSizePixels, worldSizePixels, 42)

	// Создаём боевую систему (новая архитектура)
	combatSystem := simulation.NewCombatSystem()

	// Создаём зайца и волка рядом (в радиусе атаки)
	rabbit := simulation.CreateAnimal(world, core.TypeRabbit, 300, 300)
	wolf := simulation.CreateAnimal(world, core.TypeWolf, 301, 300) // Дистанция 1 пиксель для атаки

	// Проверяем начальное здоровье
	initialHealth, _ := world.GetHealth(rabbit)
	t.Logf("Здоровье зайца до атаки: %d", initialHealth.Current)

	// Проверяем типы созданных животных
	wolfType, _ := world.GetAnimalType(wolf)
	rabbitType, _ := world.GetAnimalType(rabbit)

	t.Logf("Тип волка: %d (ожидается %d)", wolfType, core.TypeWolf)
	t.Logf("Тип зайца: %d (ожидается %d)", rabbitType, core.TypeRabbit)

	if wolfType != core.TypeWolf {
		t.Errorf("Волк имеет неправильный тип: %d, ожидается %d", wolfType, core.TypeWolf)
	}
	if rabbitType != core.TypeRabbit {
		t.Errorf("Заяц имеет неправильный тип: %d, ожидается %d", rabbitType, core.TypeRabbit)
	}

	// Делаем волка голодным для активации атаки
	world.SetHunger(wolf, core.Hunger{Value: 30.0})

	deltaTime := float32(1.0 / 60.0)

	// Тик 1: Создание AttackState
	combatSystem.Update(world, deltaTime)

	// Проверяем что AttackState создан
	if !world.HasComponent(wolf, core.MaskAttackState) {
		t.Error("AttackState не создан после первого тика")
		return
	}

	// Устанавливаем анимацию атаки на кадр 0 (Windup)
	world.SetAnimation(wolf, core.Animation{
		CurrentAnim: int(animation.AnimAttack),
		Frame:       0,
		Timer:       0,
		Playing:     true,
		FacingRight: true,
	})

	// Тик 2: Фаза Windup (кадр 0) - урон НЕ должен быть нанесен
	combatSystem.Update(world, deltaTime)

	healthAfterWindup, _ := world.GetHealth(rabbit)
	if healthAfterWindup.Current != initialHealth.Current {
		t.Errorf("Урон нанесен на кадре 0 (Windup), а не должен: %d -> %d",
			initialHealth.Current, healthAfterWindup.Current)
	}

	// Переключаем анимацию на кадр 1 (Strike)
	world.SetAnimation(wolf, core.Animation{
		CurrentAnim: int(animation.AnimAttack),
		Frame:       1,
		Timer:       0,
		Playing:     true,
		FacingRight: true,
	})

	// Тик 3: Фаза Strike (кадр 1) - урон ДОЛЖЕН быть нанесен
	combatSystem.Update(world, deltaTime)

	// Может понадобиться дополнительное обновление для executeStrike
	combatSystem.Update(world, deltaTime)

	finalHealth, _ := world.GetHealth(rabbit)
	if finalHealth.Current >= initialHealth.Current {
		// Дополнительная диагностика
		attackState, hasAttack := world.GetAttackState(wolf)
		if hasAttack {
			t.Logf("AttackState: фаза %s, HasStruck=%t", attackState.Phase.String(), attackState.HasStruck)
		} else {
			t.Logf("AttackState отсутствует")
		}

		t.Errorf("Урон НЕ нанесен на кадре 1 (Strike): здоровье %d -> %d",
			initialHealth.Current, finalHealth.Current)
	} else {
		t.Logf("✅ Волк успешно атаковал зайца (HP: %d -> %d)",
			initialHealth.Current, finalHealth.Current)
	}
}
