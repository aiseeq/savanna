package gamestate

import (
	"image"
	"image/color"
)

// Renderer интерфейс для рендеринга
type Renderer interface {
	// Рендеринг спрайтов
	DrawSprite(spriteType string, frame int, x, y float64, tint color.RGBA, facingRight bool, scale float64)

	// Рендеринг terrain
	DrawTerrain(tileType int, x, y int, grassAmount float32)

	// Рендеринг UI
	DrawText(text string, x, y float64, fontSize int, color color.RGBA)

	// Рендеринг полосок здоровья
	DrawHealthBar(x, y, width float64, health, maxHealth float32, visible bool)

	// Управление камерой
	SetCamera(x, y float64)

	// Завершение кадра
	Present()

	// Получение размеров экрана
	GetScreenSize() (int, int)

	// Для Golden Image Tests
	CaptureFrame() *image.RGBA
}

// AudioProvider интерфейс для звуковых эффектов
type AudioProvider interface {
	PlaySound(soundName string)
	SetVolume(volume float64)
}

// RenderEngine объединяет все rendering интерфейсы
type RenderEngine struct {
	Renderer      Renderer
	InputProvider InputProvider
	AudioProvider AudioProvider
}
