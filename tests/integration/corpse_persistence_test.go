package integration

import (
	"testing"

	"github.com/aiseeq/savanna/internal/core"
	"github.com/aiseeq/savanna/internal/simulation"
)

// TestCorpsePersistence - простой тест: убиваем зайца и проверяем что он остается трупом
//
//nolint:gocognit,revive,funlen // TDD тест для проверки базовой механики трупов
func TestCorpsePersistence(t *testing.T) {
	t.Parallel()

	world := core.NewWorld(640, 640, 42)
	combatSystem := simulation.NewCombatSystem()

	// Создаём зайца и убиваем его вручную
	rabbit := simulation.CreateAnimal(world, core.TypeRabbit, 300, 300)

	t.Logf("=== ПРОСТОЙ ТЕСТ ТРУПОВ ===")
	t.Logf("Заяц создан: entity %d", rabbit)

	// Убиваем зайца напрямую
	world.SetHealth(rabbit, core.Health{Current: 0, Max: 50})

	// Превращаем в труп
	corpseEntity := simulation.CreateCorpseAndGetID(world, rabbit)

	t.Logf("Труп создан: entity %d (тот же? %v)", corpseEntity, corpseEntity == rabbit)

	// Проверяем что труп создался правильно
	if corpseEntity == 0 {
		t.Fatal("Труп не создался")
	}

	if !world.IsAlive(corpseEntity) {
		t.Fatal("Сущность трупа не жива")
	}

	if !world.HasComponent(corpseEntity, core.MaskCorpse) {
		t.Fatal("У сущности нет компонента Corpse")
	}

	corpseData, _ := world.GetCorpse(corpseEntity)
	t.Logf("Начальная питательность: %.1f", corpseData.NutritionalValue)

	// НЕ ДОБАВЛЯЕМ ВОЛКА - просто проверяем что труп остается без поедания
	deltaTime := float32(1.0 / 60.0)

	// Симулируем 600 тиков (10 секунд) БЕЗ поедания
	for i := 0; i < 600; i++ {
		combatSystem.Update(world, deltaTime)

		// Проверяем каждые 60 тиков
		if i%60 == 0 {
			if world.IsAlive(corpseEntity) {
				if world.HasComponent(corpseEntity, core.MaskCorpse) {
					corpse, _ := world.GetCorpse(corpseEntity)
					t.Logf("Секунда %d: ТРУП питательность=%.1f, таймер=%.1f",
						i/60, corpse.NutritionalValue, corpse.DecayTimer)
				} else if world.HasComponent(corpseEntity, core.MaskCarrion) {
					carrion, _ := world.GetCarrion(corpseEntity)
					t.Logf("Секунда %d: ПАДАЛЬ питательность=%.1f, таймер=%.1f",
						i/60, carrion.NutritionalValue, carrion.DecayTimer)
				} else {
					t.Logf("Секунда %d: Сущность жива, но без Corpse/Carrion", i/60)
				}
			} else {
				t.Logf("Секунда %d: Сущность мертва/уничтожена", i/60)
				break
			}
		}
	}

	// Финальная проверка
	if world.IsAlive(corpseEntity) {
		t.Logf("✅ Труп выжил 10 секунд без поедания")
		if world.HasComponent(corpseEntity, core.MaskCorpse) {
			corpse, _ := world.GetCorpse(corpseEntity)
			t.Logf("Финальное состояние: ТРУП питательность=%.1f", corpse.NutritionalValue)
		} else if world.HasComponent(corpseEntity, core.MaskCarrion) {
			carrion, _ := world.GetCarrion(corpseEntity)
			t.Logf("Финальное состояние: ПАДАЛЬ питательность=%.1f", carrion.NutritionalValue)
		}
	} else {
		t.Errorf("БАГ: Труп исчез БЕЗ поедания за 10 секунд!")
	}
}
