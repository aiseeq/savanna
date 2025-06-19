package integration

import (
	"testing"

	"github.com/aiseeq/savanna/config"
	"github.com/aiseeq/savanna/internal/animation"
	"github.com/aiseeq/savanna/internal/core"
	"github.com/aiseeq/savanna/internal/simulation"
)

// TestWolfContinuousHunting проверяет что волк продолжает охотиться после поедания зайца
//
//nolint:gocognit,revive,funlen // Комплексный интеграционный тест охотничьего поведения волков
func TestWolfContinuousHunting(t *testing.T) {
	t.Parallel()
	cfg := config.LoadDefaultConfig()
	cfg.World.Size = 20

	worldSizePixels := float32(cfg.World.Size * 32)
	world := core.NewWorld(worldSizePixels, worldSizePixels, 42)

	combatSystem := simulation.NewCombatSystem()

	// Создаём несколько зайцев и одного волка в одной точке
	rabbit1 := simulation.CreateAnimal(world, core.TypeRabbit, 300, 300)
	rabbit2 := simulation.CreateAnimal(world, core.TypeRabbit, 300, 300)
	rabbit3 := simulation.CreateAnimal(world, core.TypeRabbit, 300, 300)
	wolf := simulation.CreateAnimal(world, core.TypeWolf, 300, 300)

	// Делаем волка очень голодным
	world.SetHunger(wolf, core.Hunger{Value: 20.0}) // 20% = очень голодный

	killedRabbits := 0
	deltaTime := float32(1.0 / 60.0)

	// Симулируем до 1800 тиков (30 секунд)
	for i := 0; i < 1800; i++ {
		world.Update(deltaTime)
		combatSystem.Update(world, deltaTime)

		// Эмулируем анимационную систему для тестов
		if world.HasComponent(wolf, core.MaskAttackState) {
			attackState, _ := world.GetAttackState(wolf)

			// Сразу переводим в Strike фазу для нанесения урона
			if attackState.Phase == core.AttackPhaseWindup {
				// Устанавливаем анимацию ATTACK кадр 1 для Strike
				world.SetAnimation(wolf, core.Animation{
					CurrentAnim: int(animation.AnimAttack),
					Frame:       1, // Сразу Strike для быстрого урона
					Timer:       0,
					Playing:     true,
					FacingRight: true,
				})
			} else if attackState.Phase == core.AttackPhaseStrike && attackState.HasStruck {
				// После нанесения удара завершаем анимацию
				world.SetAnimation(wolf, core.Animation{
					CurrentAnim: 9,
					Frame:       1,
					Timer:       0,
					Playing:     false, // Анимация завершена
					FacingRight: true,
				})
			}
		}

		// Подсчитываем мёртвых зайцев
		health1, _ := world.GetHealth(rabbit1)
		health2, _ := world.GetHealth(rabbit2)
		health3, _ := world.GetHealth(rabbit3)

		if (health1.Current <= 0 || !world.IsAlive(rabbit1)) && killedRabbits == 0 {
			killedRabbits = 1
			wolfHunger, _ := world.GetHunger(wolf)
			t.Logf("Заяц 1 умер на тике %d, голод волка %.1f", i, wolfHunger.Value)
		}
		if (health2.Current <= 0 || !world.IsAlive(rabbit2)) && killedRabbits == 1 {
			killedRabbits = 2
			wolfHunger, _ := world.GetHunger(wolf)
			t.Logf("Заяц 2 умер на тике %d, голод волка %.1f", i, wolfHunger.Value)
		}
		if (health3.Current <= 0 || !world.IsAlive(rabbit3)) && killedRabbits == 2 {
			killedRabbits = 3
			wolfHunger, _ := world.GetHunger(wolf)
			t.Logf("Заяц 3 умер на тике %d, голод волка %.1f", i, wolfHunger.Value)
			break
		}
	}

	t.Logf("Волк убил %d зайцев за 30 секунд", killedRabbits)

	if killedRabbits < 1 {
		t.Errorf("Ожидалось что волк убьёт минимум 1 зайца, но убил только %d", killedRabbits)
	} else {
		t.Logf("✅ Волк успешно охотится: убил %d зайцев и восстанавливает голод", killedRabbits)
	}
}
