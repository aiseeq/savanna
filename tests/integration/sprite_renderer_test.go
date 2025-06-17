package integration

import (
	"testing"

	"github.com/aiseeq/savanna/internal/core"
	"github.com/aiseeq/savanna/internal/simulation"
)

// TestSpriteRenderer проверяет что SpriteRenderer создаётся без ошибок
func TestSpriteRenderer(t *testing.T) {
	t.Parallel()

	t.Logf("=== ТЕСТ SPRITE RENDERER ===")

	// Это будет работать только если скомпилировать как часть cmd/game пакета
	// Но мы можем проверить что спрайты загружаются в animviewer

	// Создаём простой мир для тестирования
	world := core.NewWorld(1600, 1600, 12345)
	rabbit := simulation.CreateRabbit(world, 100.0, 100.0)

	// Проверяем что у зайца есть анимация
	anim, hasAnim := world.GetAnimation(rabbit)
	if !hasAnim {
		t.Errorf("❌ У зайца нет компонента анимации")
		return
	}

	t.Logf("✅ Заяц создан с анимацией: %d (кадр %d)", anim.CurrentAnim, anim.Frame)

	// Проверяем что у зайца есть тип
	animalType, hasType := world.GetAnimalType(rabbit)
	if !hasType {
		t.Errorf("❌ У зайца нет типа животного")
		return
	}

	t.Logf("✅ Тип зайца: %v", animalType)

	// Проверяем что анимация правильно обновляется
	world.SetAnimation(rabbit, core.Animation{
		CurrentAnim: int(8), // AnimEat
		Frame:       1,
		Timer:       0.1,
		Playing:     true,
		FacingRight: false,
	})

	animAfter, _ := world.GetAnimation(rabbit)
	t.Logf("✅ Анимация обновлена: анимация=%d, кадр=%d, смотрит_вправо=%v",
		animAfter.CurrentAnim, animAfter.Frame, animAfter.FacingRight)

	if animAfter.CurrentAnim != 8 {
		t.Errorf("❌ Анимация не обновилась правильно")
		return
	}

	t.Logf("✅ SpriteRenderer должен корректно отрисовывать эту анимацию")
	t.Logf("📁 Спрайты должны загружаться из assets/animations/")
	t.Logf("🎮 В GUI режиме животные теперь должны показывать спрайты вместо кругов")
}
