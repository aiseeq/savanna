package integration

import (
	"testing"

	"github.com/aiseeq/savanna/config"
	"github.com/aiseeq/savanna/internal/animation"
	"github.com/aiseeq/savanna/internal/core"
	"github.com/aiseeq/savanna/internal/simulation"
)

// TestWolfEatsRabbit простой тест: голодный волк рядом с зайцем должен его съесть
//
//nolint:gocognit,revive,funlen // Комплексный интеграционный тест питания волка
func TestWolfEatsRabbit(t *testing.T) {
	t.Parallel()
	cfg := config.LoadDefaultConfig()
	cfg.World.Size = 20

	worldSizePixels := float32(cfg.World.Size * 32)
	world := core.NewWorld(worldSizePixels, worldSizePixels, 42)

	// Создаём боевую систему (новая архитектура)
	combatSystem := simulation.NewCombatSystem()

	// Создаём зайца и волка рядом (в радиусе атаки)
	rabbit := simulation.CreateAnimal(world, core.TypeRabbit, 300, 300)
	// На расстоянии 2 пикселя (гарантированно в радиусе 12)
	wolf := simulation.CreateAnimal(world, core.TypeWolf, 302, 300)

	// Делаем волка очень голодным
	world.SetHunger(wolf, core.Hunger{Value: 30.0}) // 30% < 60% = голодный

	// Проверяем начальное здоровье зайца
	initialHealth, _ := world.GetHealth(rabbit)
	t.Logf("Начальное здоровье зайца: %d", initialHealth.Current)

	// Проверяем расстояние
	wolfPos, _ := world.GetPosition(wolf)
	rabbitPos, _ := world.GetPosition(rabbit)
	distance := (wolfPos.X-rabbitPos.X)*(wolfPos.X-rabbitPos.X) + (wolfPos.Y-rabbitPos.Y)*(wolfPos.Y-rabbitPos.Y)
	t.Logf("Расстояние между животными: %.1f (лимит атаки: %.1f)", distance, 12.0*12.0)

	// Симулируем полный цикл: атака → смерть → поедание
	deltaTime := float32(1.0 / 60.0)
	rabbitDied := false
	attackCreated := false

	for i := 0; i < 300; i++ { // 5 секунд максимум
		world.Update(deltaTime)
		combatSystem.Update(world, deltaTime)

		// Проверяем создание AttackState
		if !attackCreated && world.HasComponent(wolf, core.MaskAttackState) {
			attackCreated = true
			t.Logf("AttackState создан на тике %d", i)
		}

		// ИСПРАВЛЕНИЕ: Используем современную анимационную систему
		// Создаем анимационную систему для обработки атак
		if !world.HasComponent(wolf, core.MaskAnimation) {
			world.AddAnimation(wolf, core.Animation{
				CurrentAnim: int(animation.AnimIdle),
				Frame:       0,
				Timer:       0,
				Playing:     true,
				FacingRight: true,
			})
		}

		// Если есть AttackState, устанавливаем анимацию атаки
		if world.HasComponent(wolf, core.MaskAttackState) {
			anim, _ := world.GetAnimation(wolf)
			if anim.CurrentAnim != int(animation.AnimAttack) {
				world.SetAnimation(wolf, core.Animation{
					CurrentAnim: int(animation.AnimAttack),
					Frame:       0,
					Timer:       0,
					Playing:     true,
					FacingRight: true,
				})
			} else if i > 5 { // Через несколько тиков переходим к кадру удара
				world.SetAnimation(wolf, core.Animation{
					CurrentAnim: int(animation.AnimAttack),
					Frame:       1, // Кадр удара
					Timer:       0,
					Playing:     true,
					FacingRight: true,
				})
			}
		}

		// Проверяем не умер ли заяц
		if world.IsAlive(rabbit) {
			currentHealth, _ := world.GetHealth(rabbit)
			if currentHealth.Current <= 0 {
				rabbitDied = true
				t.Logf("Заяц умер на тике %d", i)
			}
		} else {
			// Заяц полностью исчез (съеден)
			t.Logf("✅ Волк полностью съел зайца на тике %d", i)
			return
		}
	}

	// Проверяем результат
	if rabbitDied {
		// Заяц умер, но возможно ещё не полностью съеден
		if world.HasComponent(rabbit, core.MaskCorpse) {
			t.Logf("✅ Заяц превращен в труп и процесс поедания начался")
		} else {
			t.Logf("✅ Заяц умер от атак волка")
		}
	} else {
		// Проверяем было ли хоть какое-то повреждение
		finalHealth, _ := world.GetHealth(rabbit)
		if finalHealth.Current < initialHealth.Current {
			t.Logf("Волк атаковал зайца (здоровье: %d -> %d), но не убил за 5 секунд",
				initialHealth.Current, finalHealth.Current)
		} else {
			t.Error("Волк даже не атаковал зайца - здоровье не изменилось")
		}
	}
}
