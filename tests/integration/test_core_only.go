package integration

import (
	"testing"

	"github.com/aiseeq/savanna/internal/core"
)

// TestCoreOnly - Тест только с core пакетом
func TestCoreOnly(t *testing.T) {
	t.Parallel()

	world := core.NewWorld(640, 640, 42)
	entity := world.CreateEntity()

	world.AddPosition(entity, core.Position{X: 100, Y: 100})

	t.Logf("✅ Core пакет работает без GUI зависимостей")
}
