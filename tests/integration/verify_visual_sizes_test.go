package integration

import (
	"testing"

	"github.com/aiseeq/savanna/internal/constants"
	"github.com/aiseeq/savanna/internal/core"
	"github.com/aiseeq/savanna/internal/simulation"
)

// TestVisualSizeVerification создаёт простую сцену для визуальной проверки размеров
func TestVisualSizeVerification(t *testing.T) {
	t.Parallel()

	// Создаём мир 50x38 тайлов (как в main.go)
	world := core.NewWorld(50, 38, 12345)

	// Размещаем животных в центре карты для удобства
	centerX, centerY := float32(25), float32(19)

	// Создаём зайца и волка рядом
	rabbit := simulation.CreateAnimal(world, core.TypeRabbit, centerX-2, centerY)
	wolf := simulation.CreateAnimal(world, core.TypeWolf, centerX+2, centerY)

	// Получаем их размеры для логирования
	rabbitSize, _ := world.GetSize(rabbit)
	wolfSize, _ := world.GetSize(wolf)
	rabbitBehavior, _ := world.GetBehavior(rabbit)
	wolfBehavior, _ := world.GetBehavior(wolf)

	t.Logf("=== ВИЗУАЛЬНАЯ ПРОВЕРКА РАЗМЕРОВ ===")
	t.Logf("Центр карты: (%.1f, %.1f) тайлов", centerX, centerY)
	t.Logf("")
	// ИСПРАВЛЕНИЕ: Конвертируем размеры из пикселей в тайлы для отображения
	rabbitRadiusTiles := constants.PixelsToTiles(rabbitSize.Radius)
	wolfRadiusTiles := constants.PixelsToTiles(wolfSize.Radius)
	wolfAttackRangeTiles := constants.PixelsToTiles(wolfSize.AttackRange)

	t.Logf("🐰 Заяц (ID:%d) на позиции (%.1f, %.1f):", rabbit, centerX-2, centerY)
	t.Logf("   Физический радиус: %.2f тайла", rabbitRadiusTiles)
	t.Logf("   Радиус видения: %.2f тайла", rabbitBehavior.VisionRange)
	t.Logf("   Дистанция побега: %.2f тайла", rabbitBehavior.FleeThreshold)
	t.Logf("")
	t.Logf("🐺 Волк (ID:%d) на позиции (%.1f, %.1f):", wolf, centerX+2, centerY)
	t.Logf("   Физический радиус: %.2f тайла", wolfRadiusTiles)
	t.Logf("   Радиус видения: %.2f тайла", wolfBehavior.VisionRange)
	t.Logf("   Радиус атаки: %.2f тайла", wolfAttackRangeTiles)
	t.Logf("")
	t.Logf("🔍 Ожидаемые визуальные размеры:")
	t.Logf("   Синий круг зайца: диаметр = 0.5 тайла")
	t.Logf("   Жёлтый круг зайца: диаметр = 6.0 тайлов (видение)")
	t.Logf("   Синий круг волка: диаметр = 1.0 тайл")
	t.Logf("   Жёлтый круг волка: диаметр = 10.0 тайлов (видение)")
	t.Logf("")
	t.Logf("📏 Расстояние между животными: 4.0 тайла")
	t.Logf("   Волк ВИДИТ зайца: %t (4.0 < 5.0)", 4.0 < wolfBehavior.VisionRange)
	t.Logf("   Заяц ВИДИТ волка: %t (4.0 < 3.0)", 4.0 < rabbitBehavior.VisionRange)
	t.Logf("   Волк может АТАКОВАТЬ: %t (4.0 < %.1f)", 4.0 < wolfAttackRangeTiles, wolfAttackRangeTiles)
	t.Logf("")
	// ИСПРАВЛЕНИЕ: Автоматические проверки разумности размеров используют конвертированные значения
	if rabbitRadiusTiles <= 0 || rabbitRadiusTiles > 2.0 {
		t.Errorf("❌ Неразумный радиус зайца: %.2f (должен быть 0-2.0 тайла)", rabbitRadiusTiles)
	}

	if wolfRadiusTiles <= 0 || wolfRadiusTiles > 3.0 {
		t.Errorf("❌ Неразумный радиус волка: %.2f (должен быть 0-3.0 тайла)", wolfRadiusTiles)
	}

	if rabbitBehavior.VisionRange <= 0 || rabbitBehavior.VisionRange > 10.0 {
		t.Errorf("❌ Неразумная дальность видения зайца: %.2f (должна быть 0-10 тайлов)", rabbitBehavior.VisionRange)
	}

	if wolfBehavior.VisionRange <= 0 || wolfBehavior.VisionRange > 15.0 {
		t.Errorf("❌ Неразумная дальность видения волка: %.2f (должна быть 0-15 тайлов)", wolfBehavior.VisionRange)
	}

	if wolfAttackRangeTiles <= 0 || wolfAttackRangeTiles > 5.0 {
		t.Errorf("❌ Неразумная дальность атаки волка: %.2f (должна быть 0-5 тайлов)", wolfAttackRangeTiles)
	}

	t.Logf("✅ Все размеры прошли проверку разумности")
	t.Logf("   Для визуальной проверки: make build && ./bin/savanna-game")
	t.Logf("   Желтые круги должны быть РАЗУМНОГО размера, не гигантские!")
}
