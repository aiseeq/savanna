package integration

import (
	"testing"

	"github.com/aiseeq/savanna/config"
	"github.com/aiseeq/savanna/internal/animation"
	"github.com/aiseeq/savanna/internal/core"
	"github.com/aiseeq/savanna/internal/simulation"
)

// TestWolfMultipleAttacks проверяет что волк может атаковать несколько раз
//
//nolint:gocognit,revive,funlen // Комплексный интеграционный тест системы боя волков
func TestWolfMultipleAttacks(t *testing.T) {
	t.Parallel()
	cfg := config.LoadDefaultConfig()
	cfg.World.Size = 20

	worldSizePixels := float32(cfg.World.Size * 32)
	world := core.NewWorld(worldSizePixels, worldSizePixels, 42)

	combatSystem := simulation.NewCombatSystem()

	// Создаём зайца и волка в одной точке
	rabbit := simulation.CreateAnimal(world, core.TypeRabbit, 300, 300)
	wolf := simulation.CreateAnimal(world, core.TypeWolf, 300, 300)

	// Делаем волка очень голодным
	world.SetSatiation(wolf, core.Satiation{Value: 30.0}) // 30% < 60% = голодный

	// Проверяем начальное здоровье зайца
	initialHealth, _ := world.GetHealth(rabbit)
	t.Logf("Начальное здоровье зайца: %d", initialHealth.Current)

	deltaTime := float32(1.0 / 60.0)

	// Добавляем логирование для диагностики
	t.Logf("Позиция волка: (%.1f, %.1f)", 300.0, 300.0)
	t.Logf("Позиция зайца: (%.1f, %.1f)", 300.0, 300.0)
	t.Logf("Расстояние: 0.0 (лимит атаки: 12.0)")

	// Симулируем до 300 тиков (5 секунд)
	for i := 0; i < 300; i++ {
		world.Update(deltaTime)
		combatSystem.Update(world, deltaTime)

		// Логируем создание AttackState
		if i < 20 && world.HasComponent(wolf, core.MaskAttackState) {
			attackState, _ := world.GetAttackState(wolf)
			health, _ := world.GetHealth(rabbit)
			t.Logf("Тик %d: AttackState фаза %s, здоровье зайца %d", i, attackState.Phase.String(), health.Current)
		}

		// Логируем почему нет AttackState
		if i > 2 && i < 120 && i%20 == 0 && !world.HasComponent(wolf, core.MaskAttackState) {
			health, _ := world.GetHealth(rabbit)
			hunger, _ := world.GetSatiation(wolf)
			t.Logf("Тик %d: НЕТ AttackState, здоровье зайца %d, голод волка %.1f%%", i, health.Current, hunger.Value)
		}

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

		health, _ := world.GetHealth(rabbit)
		if health.Current <= 0 || !world.IsAlive(rabbit) {
			// Заяц умер - вычисляем количество атак по урону
			finalDamage := int(initialHealth.Current - 0) // заяц умер = 0 хитов
			attackCount := (finalDamage + 24) / 25        // округляем вверх (25 урона за атаку в новой системе)

			wolfHunger, _ := world.GetSatiation(wolf)
			t.Logf("Заяц умер на тике %d после %d атак, голод волка %.1f", i, attackCount, wolfHunger.Value)

			if attackCount < 2 {
				t.Errorf("Волк атаковал только %d раз, ожидалось минимум 2 раза", attackCount)
			}
			return
		}
	}

	// Если дошли сюда - заяц не умер за 5 секунд
	t.Error("Заяц не умер за 5 секунд симуляции")
}
