package integration

import (
	"testing"

	"github.com/aiseeq/savanna/internal/core"
	"github.com/aiseeq/savanna/internal/simulation"
)

// TestVisionRangeIsReasonable проверяет что радиус зрения животных разумный
func TestVisionRangeIsReasonable(t *testing.T) {
	// Создаём мир 50x38 тайлов как в реальной игре
	world := core.NewWorld(50.0, 38.0, 12345)

	// Создаём зайца и волка в центре карты
	rabbit := simulation.CreateAnimal(world, core.TypeRabbit, 25.0, 19.0)
	wolf := simulation.CreateAnimal(world, core.TypeWolf, 25.0, 19.0)

	// Проверяем радиус зрения зайца
	rabbitBehavior, hasRabbitBehavior := world.GetBehavior(rabbit)
	if !hasRabbitBehavior {
		t.Fatal("У зайца нет компонента поведения")
	}

	// Проверяем радиус зрения волка
	wolfBehavior, hasWolfBehavior := world.GetBehavior(wolf)
	if !hasWolfBehavior {
		t.Fatal("У волка нет компонента поведения")
	}

	t.Logf("=== АНАЛИЗ РАДИУСА ЗРЕНИЯ ===")
	t.Logf("Размер мира: 50x38 тайлов")
	t.Logf("Радиус зрения зайца: %.1f тайлов", rabbitBehavior.VisionRange)
	t.Logf("Радиус зрения волка: %.1f тайлов", wolfBehavior.VisionRange)

	// Проверяем что радиус зрения разумный (не покрывает всю карту)
	maxMapDimension := float32(50.0) // max(50, 38)

	if rabbitBehavior.VisionRange > maxMapDimension/2 {
		t.Errorf("❌ Радиус зрения зайца слишком большой: %.1f > %.1f (половина карты)",
			rabbitBehavior.VisionRange, maxMapDimension/2)
	}

	if wolfBehavior.VisionRange > maxMapDimension/2 {
		t.Errorf("❌ Радиус зрения волка слишком большой: %.1f > %.1f (половина карты)",
			wolfBehavior.VisionRange, maxMapDimension/2)
	}

	// Проверяем ожидаемые значения из game_balance.go
	expectedRabbitVisionRange := float32(0.5 * 6.0) // RabbitBaseRadius * RabbitVisionMultiplier
	expectedWolfVisionRange := float32(0.75 * 6.7)  // WolfBaseRadius * WolfVisionMultiplier

	if rabbitBehavior.VisionRange != expectedRabbitVisionRange {
		t.Errorf("❌ Неправильный радиус зрения зайца: ожидался %.1f, получен %.1f",
			expectedRabbitVisionRange, rabbitBehavior.VisionRange)
	}

	if wolfBehavior.VisionRange != expectedWolfVisionRange {
		t.Errorf("❌ Неправильный радиус зрения волка: ожидался %.1f, получен %.1f",
			expectedWolfVisionRange, wolfBehavior.VisionRange)
	}

	t.Logf("✅ Радиусы зрения корректны и не покрывают всю карту")
}
