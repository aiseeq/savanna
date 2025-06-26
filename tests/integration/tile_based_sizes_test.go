package integration

import (
	"testing"

	"github.com/aiseeq/savanna/internal/constants"
	"github.com/aiseeq/savanna/internal/core"
	"github.com/aiseeq/savanna/internal/simulation"
)

// TestTileBasedSizes проверяет что все размеры теперь в тайлах
func TestTileBasedSizes(t *testing.T) {
	t.Parallel()

	// Создаём мир
	world := core.NewWorld(50, 38, 12345)

	// Создаём животных
	rabbit := simulation.CreateAnimal(world, core.TypeRabbit, 10, 10)
	wolf := simulation.CreateAnimal(world, core.TypeWolf, 15, 15)

	// Проверяем размеры зайца
	rabbitSize, hasRabbitSize := world.GetSize(rabbit)
	if !hasRabbitSize {
		t.Fatal("Заяц должен иметь компонент Size")
	}

	// Заяц должен иметь радиус 0.5 тайла (конвертируем из пикселей)
	expectedRabbitRadius := float32(0.5)
	actualRabbitRadiusInTiles := rabbitSize.Radius
	if actualRabbitRadiusInTiles != expectedRabbitRadius {
		t.Errorf("Радиус зайца: ожидался %.2f тайла, получен %.2f",
			expectedRabbitRadius, actualRabbitRadiusInTiles)
	}

	// Проверяем размеры волка
	wolfSize, hasWolfSize := world.GetSize(wolf)
	if !hasWolfSize {
		t.Fatal("Волк должен иметь компонент Size")
	}

	// Волк должен иметь радиус 0.75 тайла (конвертируем из пикселей)
	expectedWolfRadius := float32(0.75)
	actualWolfRadiusInTiles := wolfSize.Radius
	if actualWolfRadiusInTiles != expectedWolfRadius {
		t.Errorf("Радиус волка: ожидался %.2f тайла, получен %.2f",
			expectedWolfRadius, actualWolfRadiusInTiles)
	}

	// Проверяем радиус атаки волка (конвертируем из пикселей)
	expectedWolfAttackRange := expectedWolfRadius * 1.2 // WolfAttackRangeMultiplier = 1.2
	actualWolfAttackRangeInTiles := constants.PixelsToTiles(wolfSize.AttackRange)
	if absFloat(actualWolfAttackRangeInTiles-expectedWolfAttackRange) > 0.01 {
		t.Errorf("Радиус атаки волка: ожидался %.2f тайла, получен %.2f",
			expectedWolfAttackRange, actualWolfAttackRangeInTiles)
	}

	// Проверяем скорости
	rabbitSpeed, hasRabbitSpeed := world.GetSpeed(rabbit)
	if !hasRabbitSpeed {
		t.Fatal("Заяц должен иметь компонент Speed")
	}

	expectedRabbitSpeed := float32(0.6) // RabbitBaseSpeed согласно плану стабилизации
	if rabbitSpeed.Base != expectedRabbitSpeed {
		t.Errorf("Скорость зайца: ожидалась %.2f тайла/сек, получена %.2f",
			expectedRabbitSpeed, rabbitSpeed.Base)
	}

	wolfSpeed, hasWolfSpeed := world.GetSpeed(wolf)
	if !hasWolfSpeed {
		t.Fatal("Волк должен иметь компонент Speed")
	}

	expectedWolfSpeed := float32(1.0) // WolfBaseSpeed согласно плану стабилизации
	if wolfSpeed.Base != expectedWolfSpeed {
		t.Errorf("Скорость волка: ожидалась %.2f тайла/сек, получена %.2f",
			expectedWolfSpeed, wolfSpeed.Base)
	}

	// Проверяем поведенческие параметры
	rabbitBehavior, hasRabbitBehavior := world.GetBehavior(rabbit)
	if !hasRabbitBehavior {
		t.Fatal("Заяц должен иметь компонент Behavior")
	}

	// Радиус видения зайца = 0.5 * 6.0 = 3.0 тайла
	expectedRabbitVision := expectedRabbitRadius * 6.0 // RabbitVisionMultiplier
	if rabbitBehavior.VisionRange != expectedRabbitVision {
		t.Errorf("Радиус видения зайца: ожидался %.2f тайла, получен %.2f",
			expectedRabbitVision, rabbitBehavior.VisionRange)
	}

	wolfBehavior, hasWolfBehavior := world.GetBehavior(wolf)
	if !hasWolfBehavior {
		t.Fatal("Волк должен иметь компонент Behavior")
	}

	// Радиус видения волка = 0.75 * 6.7 ≈ 5.0 тайлов
	expectedWolfVision := expectedWolfRadius * 6.7 // WolfVisionMultiplier
	// Допускаем небольшую погрешность округления
	if absFloat(wolfBehavior.VisionRange-expectedWolfVision) > 0.01 {
		t.Errorf("Радиус видения волка: ожидался %.2f тайла, получен %.2f",
			expectedWolfVision, wolfBehavior.VisionRange)
	}

	t.Logf("✅ Все размеры корректно переведены в тайлы:")
	t.Logf("   Заяц: радиус %.2f тайла, видение %.2f тайла, скорость %.2f тайла/сек",
		rabbitSize.Radius, rabbitBehavior.VisionRange, rabbitSpeed.Base)
	t.Logf("   Волк: радиус %.2f тайла, видение %.2f тайла, скорость %.2f тайла/сек",
		wolfSize.Radius, wolfBehavior.VisionRange, wolfSpeed.Base)
	t.Logf("   Атака волка: %.2f тайла", constants.PixelsToTiles(wolfSize.AttackRange))
}

// absFloat возвращает абсолютное значение float32
func absFloat(x float32) float32 {
	if x < 0 {
		return -x
	}
	return x
}
