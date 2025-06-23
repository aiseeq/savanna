package gamestate

import (
	"image/color"

	"github.com/aiseeq/savanna/internal/core"
)

// RenderInstructions выполняет рендеринг набора инструкций
func RenderInstructions(instructions RenderInstructionSet, renderer Renderer) {
	// Устанавливаем камеру
	renderer.SetCamera(instructions.CameraX, instructions.CameraY)

	// Рендерим terrain первым (задний план)
	for _, terrain := range instructions.Terrain {
		renderer.DrawTerrain(terrain.TileType, terrain.TileX, terrain.TileY, terrain.GrassAmount)
	}

	// Рендерим спрайты (животные)
	for _, sprite := range instructions.Sprites {
		spriteType := getAnimalSpriteType(sprite.AnimalType)
		renderer.DrawSprite(spriteType, sprite.Frame, sprite.X, sprite.Y, sprite.Tint, sprite.FacingRight, sprite.Scale)
	}

	// Рендерим полоски здоровья
	for _, healthBar := range instructions.HealthBars {
		if healthBar.Visible {
			renderer.DrawHealthBar(healthBar.X, healthBar.Y, healthBar.Width, healthBar.Health, healthBar.MaxHealth, healthBar.Visible)
		}
	}

	// Рендерим UI (передний план)
	for _, ui := range instructions.UI {
		renderer.DrawText(ui.Text, ui.X, ui.Y, ui.FontSize, ui.Color)
	}

	// Рендерим отладочный текст
	for _, debug := range instructions.DebugTexts {
		renderer.DrawText(debug.Text, debug.X, debug.Y, 12, color.RGBA{255, 255, 0, 255}) // Желтый для отладки
	}

	// Завершаем кадр
	renderer.Present()
}

// getAnimalSpriteType преобразует тип животного в строку для спрайта
func getAnimalSpriteType(animalType core.AnimalType) string {
	switch animalType {
	case core.TypeRabbit:
		return "rabbit"
	case core.TypeWolf:
		return "wolf"
	default:
		return "unknown"
	}
}

// GameRenderer основной класс для рендеринга игры
type GameRenderer struct {
	engine *RenderEngine
}

// NewGameRenderer создает новый рендерер игры
func NewGameRenderer(engine *RenderEngine) *GameRenderer {
	return &GameRenderer{
		engine: engine,
	}
}

// RenderFrame рендерит один кадр игры
func (gr *GameRenderer) RenderFrame(gameState *GameState) {
	// Генерируем инструкции из состояния игры
	instructions := gameState.GenerateRenderInstructions()

	// Выполняем рендеринг
	RenderInstructions(instructions, gr.engine.Renderer)
}

// ProcessInput обрабатывает ввод
func (gr *GameRenderer) ProcessInput(gameState *GameState) {
	events := gr.engine.InputProvider.PollEvents()
	gameState.ProcessInput(events)
}

// PlaySoundEffect проигрывает звуковой эффект
func (gr *GameRenderer) PlaySoundEffect(soundName string) {
	gr.engine.AudioProvider.PlaySound(soundName)
}
