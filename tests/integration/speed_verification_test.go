package integration

import (
	"testing"

	"github.com/aiseeq/savanna/internal/core"
	"github.com/aiseeq/savanna/internal/simulation"
)

// TestAnimalSpeedsAreReasonable проверяет что скорости животных не слишком высокие
func TestAnimalSpeedsAreReasonable(t *testing.T) {
	world := core.NewWorld(50.0, 38.0, 12345)

	// Создаём зайца и волка
	rabbit := simulation.CreateAnimal(world, core.TypeRabbit, 25.0, 19.0)
	wolf := simulation.CreateAnimal(world, core.TypeWolf, 25.0, 19.0)

	// Проверяем базовые скорости
	rabbitSpeed, hasRabbitSpeed := world.GetSpeed(rabbit)
	if !hasRabbitSpeed {
		t.Fatal("У зайца нет компонента скорости")
	}

	wolfSpeed, hasWolfSpeed := world.GetSpeed(wolf)
	if !hasWolfSpeed {
		t.Fatal("У волка нет компонента скорости")
	}

	// Проверяем поведенческие множители
	rabbitBehavior, hasRabbitBehavior := world.GetBehavior(rabbit)
	if !hasRabbitBehavior {
		t.Fatal("У зайца нет компонента поведения")
	}

	wolfBehavior, hasWolfBehavior := world.GetBehavior(wolf)
	if !hasWolfBehavior {
		t.Fatal("У волка нет компонента поведения")
	}

	t.Logf("=== АНАЛИЗ СКОРОСТЕЙ ЖИВОТНЫХ ===")
	t.Logf("Заяц - базовая скорость: %.1f тайлов/сек", rabbitSpeed.Base)
	t.Logf("Заяц - текущая скорость: %.1f тайлов/сек", rabbitSpeed.Current)
	t.Logf("Заяц - поведенческие множители:")
	t.Logf("  - ContentSpeed: %.1f", rabbitBehavior.ContentSpeed)
	t.Logf("  - WanderingSpeed: %.1f", rabbitBehavior.WanderingSpeed)
	t.Logf("  - SearchSpeed: %.1f", rabbitBehavior.SearchSpeed)

	t.Logf("Волк - базовая скорость: %.1f тайлов/сек", wolfSpeed.Base)
	t.Logf("Волк - текущая скорость: %.1f тайлов/сек", wolfSpeed.Current)
	t.Logf("Волк - поведенческие множители:")
	t.Logf("  - ContentSpeed: %.1f", wolfBehavior.ContentSpeed)
	t.Logf("  - WanderingSpeed: %.1f", wolfBehavior.WanderingSpeed)
	t.Logf("  - SearchSpeed: %.1f", wolfBehavior.SearchSpeed)

	// Расчёт эффективных скоростей в разных состояниях
	rabbitContentSpeed := rabbitSpeed.Base * rabbitBehavior.ContentSpeed
	rabbitWanderingSpeed := rabbitSpeed.Base * rabbitBehavior.WanderingSpeed
	rabbitSearchSpeed := rabbitSpeed.Base * rabbitBehavior.SearchSpeed

	wolfContentSpeed := wolfSpeed.Base * wolfBehavior.ContentSpeed
	wolfWanderingSpeed := wolfSpeed.Base * wolfBehavior.WanderingSpeed
	wolfSearchSpeed := wolfSpeed.Base * wolfBehavior.SearchSpeed

	t.Logf("=== ЭФФЕКТИВНЫЕ СКОРОСТИ ===")
	t.Logf("Заяц - спокойный: %.1f тайлов/сек", rabbitContentSpeed)
	t.Logf("Заяц - блуждание: %.1f тайлов/сек", rabbitWanderingSpeed)
	t.Logf("Заяц - поиск еды: %.1f тайлов/сек", rabbitSearchSpeed)

	t.Logf("Волк - спокойный: %.1f тайлов/сек", wolfContentSpeed)
	t.Logf("Волк - блуждание: %.1f тайлов/сек", wolfWanderingSpeed)
	t.Logf("Волк - охота: %.1f тайлов/сек", wolfSearchSpeed)

	// Проверяем что скорости разумные (не больше 10 тайлов/сек)
	maxReasonableSpeed := float32(10.0)

	if rabbitSpeed.Base > maxReasonableSpeed {
		t.Errorf("❌ Базовая скорость зайца слишком высокая: %.1f > %.1f", rabbitSpeed.Base, maxReasonableSpeed)
	}

	if wolfSpeed.Base > maxReasonableSpeed {
		t.Errorf("❌ Базовая скорость волка слишком высокая: %.1f > %.1f", wolfSpeed.Base, maxReasonableSpeed)
	}

	// Проверяем ожидаемые значения из game_balance.go
	expectedRabbitSpeed := float32(0.6) // RabbitBaseSpeed
	expectedWolfSpeed := float32(1.0)   // WolfBaseSpeed

	if rabbitSpeed.Base != expectedRabbitSpeed {
		t.Errorf("❌ Неправильная базовая скорость зайца: ожидалась %.1f, получена %.1f",
			expectedRabbitSpeed, rabbitSpeed.Base)
	}

	if wolfSpeed.Base != expectedWolfSpeed {
		t.Errorf("❌ Неправильная базовая скорость волка: ожидалась %.1f, получена %.1f",
			expectedWolfSpeed, wolfSpeed.Base)
	}

	t.Logf("✅ Скорости животных корректны и разумные")
}
