package integration

import (
	"strings"
	"testing"

	"github.com/aiseeq/savanna/internal/core"
	"github.com/aiseeq/savanna/internal/gamestate"
	"github.com/aiseeq/savanna/internal/simulation"
)

// TestHealthBarsAndHungerDisplayArePresent тестирует что хелсбары и показатели сытости отображаются
func TestHealthBarsAndHungerDisplayArePresent(t *testing.T) {
	// Создаём gamestate для генерации инструкций рендеринга
	config := &gamestate.GameConfig{
		WorldWidth:    50.0,
		WorldHeight:   38.0,
		FixedTimeStep: 1.0 / 60.0,
		RandomSeed:    12345,
	}
	gameState := gamestate.NewGameState(config)

	// Создаём животных
	world := gameState.GetWorld()
	rabbit := simulation.CreateAnimal(world, core.TypeRabbit, 25.0, 19.0)
	wolf := simulation.CreateAnimal(world, core.TypeWolf, 25.0, 19.0)

	// Устанавливаем разные уровни здоровья и голода для тестирования
	world.SetHealth(rabbit, core.Health{Current: 30, Max: 50}) // 60% здоровья
	world.SetHunger(rabbit, core.Hunger{Value: 75.0})          // 75% сытости

	world.SetHealth(wolf, core.Health{Current: 80, Max: 100}) // 80% здоровья
	world.SetHunger(wolf, core.Hunger{Value: 40.0})           // 40% сытости

	// Генерируем инструкции рендеринга
	instructions := gameState.GenerateRenderInstructions()

	t.Logf("=== ТЕСТ ОТОБРАЖЕНИЯ UI ЭЛЕМЕНТОВ ===")
	t.Logf("Спрайты: %d, UI: %d, Полоски здоровья: %d, Отладочный текст: %d",
		len(instructions.Sprites), len(instructions.UI), len(instructions.HealthBars), len(instructions.DebugTexts))

	// Проверяем наличие полосок здоровья
	healthBarFound := len(instructions.HealthBars) > 0
	if healthBarFound {
		t.Logf("✅ Найдено %d полосок здоровья", len(instructions.HealthBars))
		for i, hb := range instructions.HealthBars {
			t.Logf("  Полоска %d: здоровье %.0f/%.0f, видимость: %v", i+1, hb.Health, hb.MaxHealth, hb.Visible)
		}
	}

	// Проверяем наличие отображения голода в UI или отладочном тексте
	hungerDisplayFound := false

	// Проверяем UI инструкции
	for _, ui := range instructions.UI {
		if strings.Contains(ui.Text, "%") || strings.Contains(ui.Text, "голод") || strings.Contains(ui.Text, "сытость") {
			hungerDisplayFound = true
			t.Logf("✅ Найден показатель голода в UI: %s", ui.Text)
		}
	}

	// Проверяем отладочный текст
	for _, debug := range instructions.DebugTexts {
		if strings.Contains(debug.Text, "%") || strings.Contains(debug.Text, "голод") || strings.Contains(debug.Text, "сытость") {
			hungerDisplayFound = true
			t.Logf("✅ Найден показатель голода в отладочном тексте: %s", debug.Text)
		}
	}

	// Проверяем результаты
	if !healthBarFound {
		t.Errorf("❌ Полоски здоровья не найдены в инструкциях рендеринга")
	}

	if !hungerDisplayFound {
		t.Errorf("❌ Показатели сытости не найдены в инструкциях рендеринга")
	}

	if healthBarFound && hungerDisplayFound {
		t.Logf("✅ И полоски здоровья, и показатели сытости корректно генерируются")
	}

	// Дополнительно проверяем что у животных есть нужные компоненты
	rabbitHealth, hasRabbitHealth := world.GetHealth(rabbit)
	rabbitHunger, hasRabbitHunger := world.GetHunger(rabbit)

	if !hasRabbitHealth {
		t.Errorf("❌ У зайца нет компонента здоровья")
	}
	if !hasRabbitHunger {
		t.Errorf("❌ У зайца нет компонента голода")
	}

	if hasRabbitHealth && hasRabbitHunger {
		t.Logf("✅ Заяц имеет здоровье %d/%d и голод %.1f%%", rabbitHealth.Current, rabbitHealth.Max, rabbitHunger.Value)
	}
}
