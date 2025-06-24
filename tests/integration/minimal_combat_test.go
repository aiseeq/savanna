package integration

import (
	"testing"

	"github.com/aiseeq/savanna/internal/core"
	"github.com/aiseeq/savanna/internal/simulation"
)

// TestMinimalCombat - Минимальный тест боевой системы без GUI зависимостей
func TestMinimalCombat(t *testing.T) {
	t.Parallel()

	world := core.NewWorld(640, 640, 42)
	combatSystem := simulation.NewCombatSystem()

	t.Logf("=== МИНИМАЛЬНЫЙ ТЕСТ ===")

	// НЕ создаём животных через CreateAnimal, а создаём вручную
	wolf := world.CreateEntity()
	rabbit := world.CreateEntity()

	// Настраиваем волка вручную без Animation компонента
	world.AddPosition(wolf, core.Position{X: 300, Y: 300})
	world.AddPosition(rabbit, core.Position{X: 305, Y: 300})

	world.AddBehavior(wolf, core.Behavior{
		Type:            core.BehaviorPredator,
		VisionRange:     160.0,
		HungerThreshold: 60.0,
	})

	world.AddSize(wolf, core.Size{
		Radius:      24.0,
		AttackRange: 28.8,
	})

	world.AddAnimalConfig(wolf, core.AnimalConfig{
		AttackDamage: 25,
		HitChance:    1.0,
	})

	world.AddHunger(wolf, core.Hunger{Value: 20.0})

	world.AddBehavior(rabbit, core.Behavior{Type: core.BehaviorHerbivore})
	world.AddSize(rabbit, core.Size{Radius: 16.0})
	world.AddHealth(rabbit, core.Health{Current: 50, Max: 50})

	// Создаём AttackState вручную
	world.AddAttackState(wolf, core.AttackState{
		Target:    rabbit,
		Phase:     core.AttackPhaseStrike,
		HasStruck: false,
	})

	t.Logf("Тест без Animation компонента")

	deltaTime := float32(1.0 / 60.0)
	combatSystem.Update(world, deltaTime)

	// Проверяем создание DamageFlash
	if world.HasComponent(rabbit, core.MaskDamageFlash) {
		t.Logf("✅ DamageFlash создан без GUI зависимостей")
	} else {
		t.Error("DamageFlash не создан")
	}
}
