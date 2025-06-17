package integration

import (
	"testing"

	"github.com/aiseeq/savanna/internal/animation"
	"github.com/aiseeq/savanna/internal/core"
	"github.com/aiseeq/savanna/internal/simulation"
)

// TestExactDamageTiming точно определяет когда наносится урон
func TestExactDamageTiming(t *testing.T) {
	t.Parallel()
	world := core.NewWorld(96, 96, 42)
	combatSystem := simulation.NewCombatSystem()

	// Создаём животных рядом
	rabbit := simulation.CreateRabbit(world, 40, 48)
	wolf := simulation.CreateWolf(world, 45, 48)

	// Волк голоден
	world.SetHunger(wolf, core.Hunger{Value: 5.0})

	// Сначала создаем AttackState через CombatSystem
	combatSystem.Update(world, 1.0/60.0)

	// Проверяем что AttackState создан
	if !world.HasComponent(wolf, core.MaskAttackState) {
		t.Fatal("AttackState не создан! Убедитесь что волк находится в радиусе атаки")
	}

	initialHealth, _ := world.GetHealth(rabbit)
	t.Logf("=== ТОЧНОЕ ОПРЕДЕЛЕНИЕ МОМЕНТА УРОНА ===")
	t.Logf("Начальное здоровье зайца: %d", initialHealth.Current)

	_ = initialHealth.Current // Используется для логирования

	// Тестируем несколько сценариев
	scenarios := []struct {
		name  string
		frame int
	}{
		{"Кадр 0", 0},
		{"Кадр 1", 1},
	}

	for _, scenario := range scenarios {
		t.Logf("\n--- %s ---", scenario.name)

		// Устанавливаем конкретный кадр
		world.SetAnimation(wolf, core.Animation{
			CurrentAnim: int(animation.AnimAttack),
			Frame:       scenario.frame,
			Timer:       0,
			Playing:     true,
			FacingRight: true,
		})

		// Сбрасываем здоровье
		world.SetHealth(rabbit, core.Health{Current: 50, Max: 50})
		lastHealth := int16(50)

		t.Logf("Волк в ATTACK кадр %d", scenario.frame)

		// Обновляем боевую систему
		combatSystem.Update(world, 1.0/60.0)

		// Дополнительное обновление для executeStrike
		if scenario.frame == 1 {
			combatSystem.Update(world, 1.0/60.0)
		}

		// Проверяем здоровье
		currentHealth, _ := world.GetHealth(rabbit)
		t.Logf("Здоровье: %d -> %d", lastHealth, currentHealth.Current)

		if currentHealth.Current != lastHealth {
			t.Logf("🩸 УРОН НАНЕСЕН на кадре %d!", scenario.frame)
			if scenario.frame != 1 {
				t.Errorf("❌ ОШИБКА: Урон нанесен на кадре %d, а должен на кадре 1!", scenario.frame)
			}
		} else {
			t.Logf("⚪ Урон НЕ нанесен на кадре %d", scenario.frame)
			if scenario.frame == 1 {
				t.Errorf("❌ ОШИБКА: Урон НЕ нанесен на кадре 1, а должен быть нанесен!")
			}
		}

		// Ждем чтобы кулдаун прошел
		for i := 0; i < 60; i++ {
			combatSystem.Update(world, 1.0/60.0)
		}
	}
}
