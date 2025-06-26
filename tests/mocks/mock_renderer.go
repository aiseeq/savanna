package mocks

import (
	"image"
	"image/color"

	"github.com/aiseeq/savanna/internal/gamestate"
)

// MockRenderer для тестов - записывает все вызовы
type MockRenderer struct {
	DrawSpriteCalls    []MockDrawSpriteCall
	DrawTerrainCalls   []MockDrawTerrainCall
	DrawTextCalls      []MockDrawTextCall
	DrawHealthBarCalls []MockDrawHealthBarCall
	SetCameraCalls     []MockSetCameraCall
	PresentCalls       int
	LastCapturedFrame  *image.RGBA
	ScreenWidth        int
	ScreenHeight       int
}

type MockDrawSpriteCall struct {
	SpriteType  string
	Frame       int
	X, Y        float64
	Tint        color.RGBA
	FacingRight bool
	Scale       float64
}

type MockDrawTerrainCall struct {
	TileType    int
	X, Y        int
	GrassAmount float32
}

type MockDrawTextCall struct {
	Text     string
	X, Y     float64
	FontSize int
	Color    color.RGBA
}

type MockDrawHealthBarCall struct {
	X, Y, Width       float64
	Health, MaxHealth float32
	Visible           bool
}

type MockSetCameraCall struct {
	X, Y float64
}

// NewMockRenderer создает новый mock renderer
func NewMockRenderer(screenWidth, screenHeight int) *MockRenderer {
	return &MockRenderer{
		DrawSpriteCalls:    make([]MockDrawSpriteCall, 0),
		DrawTerrainCalls:   make([]MockDrawTerrainCall, 0),
		DrawTextCalls:      make([]MockDrawTextCall, 0),
		DrawHealthBarCalls: make([]MockDrawHealthBarCall, 0),
		SetCameraCalls:     make([]MockSetCameraCall, 0),
		PresentCalls:       0,
		ScreenWidth:        screenWidth,
		ScreenHeight:       screenHeight,
	}
}

func (m *MockRenderer) DrawSprite(spriteType string, frame int, x, y float64, tint color.RGBA, facingRight bool, scale float64) {
	m.DrawSpriteCalls = append(m.DrawSpriteCalls, MockDrawSpriteCall{
		SpriteType:  spriteType,
		Frame:       frame,
		X:           x,
		Y:           y,
		Tint:        tint,
		FacingRight: facingRight,
		Scale:       scale,
	})
}

func (m *MockRenderer) DrawTerrain(tileType int, x, y int, grassAmount float32) {
	m.DrawTerrainCalls = append(m.DrawTerrainCalls, MockDrawTerrainCall{
		TileType:    tileType,
		X:           x,
		Y:           y,
		GrassAmount: grassAmount,
	})
}

func (m *MockRenderer) DrawText(text string, x, y float64, fontSize int, color color.RGBA) {
	m.DrawTextCalls = append(m.DrawTextCalls, MockDrawTextCall{
		Text:     text,
		X:        x,
		Y:        y,
		FontSize: fontSize,
		Color:    color,
	})
}

func (m *MockRenderer) DrawHealthBar(x, y, width float64, health, maxHealth float32, visible bool) {
	m.DrawHealthBarCalls = append(m.DrawHealthBarCalls, MockDrawHealthBarCall{
		X:         x,
		Y:         y,
		Width:     width,
		Health:    health,
		MaxHealth: maxHealth,
		Visible:   visible,
	})
}

func (m *MockRenderer) SetCamera(x, y float64) {
	m.SetCameraCalls = append(m.SetCameraCalls, MockSetCameraCall{
		X: x,
		Y: y,
	})
}

func (m *MockRenderer) Present() {
	m.PresentCalls++
}

func (m *MockRenderer) GetScreenSize() (int, int) {
	return m.ScreenWidth, m.ScreenHeight
}

func (m *MockRenderer) CaptureFrame() *image.RGBA {
	// Создаем простое тестовое изображение
	img := image.NewRGBA(image.Rect(0, 0, m.ScreenWidth, m.ScreenHeight))
	// Заливаем зеленым (имитация травы)
	for y := 0; y < m.ScreenHeight; y++ {
		for x := 0; x < m.ScreenWidth; x++ {
			img.Set(x, y, color.RGBA{0, 128, 0, 255})
		}
	}
	m.LastCapturedFrame = img
	return img
}

// Reset очищает все записанные вызовы
func (m *MockRenderer) Reset() {
	m.DrawSpriteCalls = m.DrawSpriteCalls[:0]
	m.DrawTerrainCalls = m.DrawTerrainCalls[:0]
	m.DrawTextCalls = m.DrawTextCalls[:0]
	m.DrawHealthBarCalls = m.DrawHealthBarCalls[:0]
	m.SetCameraCalls = m.SetCameraCalls[:0]
	m.PresentCalls = 0
	m.LastCapturedFrame = nil
}

// MockAudioProvider для тестов
type MockAudioProvider struct {
	PlaySoundCalls []string
	Volume         float64
}

func NewMockAudioProvider() *MockAudioProvider {
	return &MockAudioProvider{
		PlaySoundCalls: make([]string, 0),
		Volume:         1.0,
	}
}

func (m *MockAudioProvider) PlaySound(soundName string) {
	m.PlaySoundCalls = append(m.PlaySoundCalls, soundName)
}

func (m *MockAudioProvider) SetVolume(volume float64) {
	m.Volume = volume
}

// NewMockRenderEngine создает полностью mock движок для тестов
func NewMockRenderEngine(screenWidth, screenHeight int) *gamestate.RenderEngine {
	return &gamestate.RenderEngine{
		Renderer:      NewMockRenderer(screenWidth, screenHeight),
		InputProvider: NewMockInputProvider([]gamestate.InputEvent{}),
		AudioProvider: NewMockAudioProvider(),
	}
}
