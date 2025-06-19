package integration

import (
	"testing"

	"github.com/aiseeq/savanna/internal/animation"
	"github.com/aiseeq/savanna/internal/core"
	"github.com/aiseeq/savanna/internal/simulation"
)

// TestDisappearingRabbit диагностика исчезающего зайца
//
//nolint:gocognit,revive,funlen // Комплексный диагностический тест исчезновения зайца
func TestDisappearingRabbit(t *testing.T) {
	t.Parallel()
	world := core.NewWorld(1600, 1600, 42)
	combatSystem := simulation.NewCombatSystem()

	rabbit := simulation.CreateAnimal(world, core.TypeRabbit, 800, 800)
	wolf := simulation.CreateAnimal(world, core.TypeWolf, 810, 800)
	world.SetHunger(wolf, core.Hunger{Value: 10.0})

	// Проверяем начальное здоровье
	initialHealth, _ := world.GetHealth(rabbit)
	t.Logf("Начальное здоровье зайца: %d", initialHealth.Current)

	// Симулируем ОДНУ атаку вручную
	anim := core.Animation{
		CurrentAnim: int(animation.AnimAttack),
		Frame:       1, // 2-й кадр
		Timer:       0,
		Playing:     true,
		FacingRight: true,
	}
	world.SetAnimation(wolf, anim)

	t.Logf("Волк в анимации ATTACK кадр 2")

	// ШАГ 1: Одна атака
	combatSystem.Update(world, 1.0/60.0)

	health1, _ := world.GetHealth(rabbit)
	t.Logf("После 1-й атаки: здоровье %d", health1.Current)

	if health1.Current == 0 {
		t.Logf("❌ ПРОБЛЕМА: Заяц умер от 1 удара!")

		// Проверяем что произошло
		if world.HasComponent(rabbit, core.MaskCorpse) {
			t.Logf("✅ Труп создан")
		} else {
			t.Logf("❌ Труп НЕ создан!")
		}

		if world.IsAlive(rabbit) {
			t.Logf("✅ Entity еще существует")
		} else {
			t.Logf("❌ Entity УЖЕ УНИЧТОЖЕН!")
		}

		return
	}

	// ШАГ 2: Вторая атака (ждем кулдаун)
	for i := 0; i < 60; i++ { // 1 секунда кулдауна
		combatSystem.Update(world, 1.0/60.0)
	}

	// Снова атакуем
	anim.Frame = 1
	world.SetAnimation(wolf, anim)
	combatSystem.Update(world, 1.0/60.0)

	health2, _ := world.GetHealth(rabbit)
	t.Logf("После 2-й атаки: здоровье %d", health2.Current)

	if health2.Current == 0 {
		t.Logf("✅ Заяц умер от 2 ударов (правильно)")

		// Проверяем что произошло
		if world.HasComponent(rabbit, core.MaskCorpse) {
			corpse, _ := world.GetCorpse(rabbit)
			t.Logf("✅ Труп создан (питательность %.1f)", corpse.NutritionalValue)
		} else {
			t.Logf("❌ Труп НЕ создан!")
		}

		if world.IsAlive(rabbit) {
			t.Logf("✅ Entity еще существует")
		} else {
			t.Logf("❌ Entity УЖЕ УНИЧТОЖЕН!")
		}

		// ШАГ 3: Проверяем начало поедания
		combatSystem.Update(world, 1.0/60.0)

		if world.HasComponent(wolf, core.MaskEatingState) {
			t.Logf("✅ Поедание началось")
		} else {
			t.Logf("❌ Поедание НЕ началось!")

			// Проверяем условия
			wolfPos, _ := world.GetPosition(wolf)
			rabbitPos, _ := world.GetPosition(rabbit)
			distance := (wolfPos.X-rabbitPos.X)*(wolfPos.X-rabbitPos.X) + (wolfPos.Y-rabbitPos.Y)*(wolfPos.Y-rabbitPos.Y)
			wolfHunger, _ := world.GetHunger(wolf)

			t.Logf("Дистанция: %.1f, голод волка: %.1f%%, есть труп: %t",
				distance, wolfHunger.Value, world.HasComponent(rabbit, core.MaskCorpse))
		}
	} else {
		t.Logf("Заяц выжил после 2 ударов (здоровье %d)", health2.Current)
	}
}
