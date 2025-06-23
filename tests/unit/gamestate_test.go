package unit

import (
	"image/color"
	"testing"

	"github.com/aiseeq/savanna/internal/gamestate"
)

// TestGameStateCreation проверяет создание игрового состояния
func TestGameStateCreation(t *testing.T) {
	config := &gamestate.GameConfig{
		WorldWidth:    640,
		WorldHeight:   480,
		FixedTimeStep: 1.0 / 60.0,
		RandomSeed:    12345,
	}

	gs := gamestate.NewGameState(config)

	if gs == nil {
		t.Fatal("GameState should not be nil")
	}

	world := gs.GetWorld()
	if world == nil {
		t.Fatal("World should not be nil")
	}
}

// TestCameraScroll проверяет логику скролла камеры
func TestCameraScroll(t *testing.T) {
	config := &gamestate.GameConfig{
		WorldWidth:    640,
		WorldHeight:   480,
		FixedTimeStep: 1.0 / 60.0,
		RandomSeed:    12345,
	}

	gs := gamestate.NewGameState(config)

	// Начальное состояние камеры
	camera := gs.GetCameraState()
	if camera.X != 0 || camera.Y != 0 {
		t.Errorf("Initial camera position should be (0,0), got (%.1f,%.1f)", camera.X, camera.Y)
	}

	// Имитируем скролл: нажатие правой кнопки мыши
	events := []gamestate.InputEvent{
		{
			Type:   gamestate.InputMouseDown,
			Button: gamestate.MouseButtonRight,
			X:      100,
			Y:      100,
		},
		{
			Type: gamestate.InputMouseMove,
			X:    150, // сдвинули на 50 пикселей
			Y:    100,
		},
		{
			Type:   gamestate.InputMouseUp,
			Button: gamestate.MouseButtonRight,
			X:      150,
			Y:      100,
		},
	}

	gs.ProcessInput(events)

	camera = gs.GetCameraState()
	if camera.X != 50 {
		t.Errorf("Camera X should be 50 after scroll, got %.1f", camera.X)
	}

	// После отпускания кнопки скролл должен прекратиться
	if camera.IsScrolling {
		t.Error("Camera should not be scrolling after mouse up")
	}

	// Дополнительное движение мыши не должно влиять на камеру
	moreEvents := []gamestate.InputEvent{
		{
			Type: gamestate.InputMouseMove,
			X:    200,
			Y:    100,
		},
	}

	gs.ProcessInput(moreEvents)
	camera = gs.GetCameraState()
	if camera.X != 50 {
		t.Errorf("Camera X should remain 50 after move without scroll, got %.1f", camera.X)
	}
}

// TestRenderInstructionGeneration проверяет генерацию инструкций рендеринга
func TestRenderInstructionGeneration(t *testing.T) {
	config := &gamestate.GameConfig{
		WorldWidth:    640,
		WorldHeight:   480,
		FixedTimeStep: 1.0 / 60.0,
		RandomSeed:    12345,
	}

	gs := gamestate.NewGameState(config)

	// Генерируем инструкции рендеринга
	instructions := gs.GenerateRenderInstructions()

	// Проверяем, что инструкции созданы
	if len(instructions.UI) == 0 {
		t.Error("Should have UI instructions")
	}

	// Проверяем, что есть инструкции для terrain
	if len(instructions.Terrain) == 0 {
		t.Error("Should have terrain instructions")
	}

	// Должны быть инструкции для спрайтов (животных)
	if len(instructions.Sprites) == 0 {
		t.Error("Should have sprite instructions for animals")
	}
}

// TestMockRenderer проверяет работу mock рендерера
func TestMockRenderer(t *testing.T) {
	mockRenderer := gamestate.NewMockRenderer(800, 600)

	// Тестируем вызовы рендерера
	mockRenderer.DrawSprite("rabbit", 0, 100, 200, color.RGBA{255, 255, 255, 255}, true, 1.0)
	mockRenderer.DrawText("Test", 10, 10, 16, color.RGBA{255, 255, 255, 255})
	mockRenderer.SetCamera(50, 50)
	mockRenderer.Present()

	// Проверяем записанные вызовы
	if len(mockRenderer.DrawSpriteCalls) != 1 {
		t.Errorf("Expected 1 DrawSprite call, got %d", len(mockRenderer.DrawSpriteCalls))
	}

	if len(mockRenderer.DrawTextCalls) != 1 {
		t.Errorf("Expected 1 DrawText call, got %d", len(mockRenderer.DrawTextCalls))
	}

	if len(mockRenderer.SetCameraCalls) != 1 {
		t.Errorf("Expected 1 SetCamera call, got %d", len(mockRenderer.SetCameraCalls))
	}

	if mockRenderer.PresentCalls != 1 {
		t.Errorf("Expected 1 Present call, got %d", mockRenderer.PresentCalls)
	}

	// Проверяем данные первого вызова DrawSprite
	spriteCall := mockRenderer.DrawSpriteCalls[0]
	if spriteCall.SpriteType != "rabbit" {
		t.Errorf("Expected sprite type 'rabbit', got '%s'", spriteCall.SpriteType)
	}
	if spriteCall.X != 100 || spriteCall.Y != 200 {
		t.Errorf("Expected position (100,200), got (%.1f,%.1f)", spriteCall.X, spriteCall.Y)
	}
}

// TestEventRecording проверяет запись и воспроизведение событий
func TestEventRecording(t *testing.T) {
	// Создаем рекордер
	recorder := gamestate.NewEventRecorder()

	// Записываем события
	events := []gamestate.InputEvent{
		{Type: gamestate.InputMouseDown, Button: gamestate.MouseButtonRight, X: 100, Y: 100},
		{Type: gamestate.InputMouseMove, X: 150, Y: 100},
		{Type: gamestate.InputMouseUp, Button: gamestate.MouseButtonRight, X: 150, Y: 100},
	}

	for _, event := range events {
		recorder.Record(event)
	}

	// Проверяем записанные события
	recordedEvents := recorder.GetEvents()
	if len(recordedEvents) != 3 {
		t.Errorf("Expected 3 recorded events, got %d", len(recordedEvents))
	}

	// Создаем mock провайдер из записанных событий
	mockProvider := gamestate.NewMockInputProvider(recordedEvents)

	// Воспроизводим события
	config := &gamestate.GameConfig{
		WorldWidth:    640,
		WorldHeight:   480,
		FixedTimeStep: 1.0 / 60.0,
		RandomSeed:    12345,
	}

	gs := gamestate.NewGameState(config)

	// Обрабатываем все записанные события
	for i := 0; i < 3; i++ {
		events := mockProvider.PollEvents()
		if len(events) > 0 {
			gs.ProcessInput(events)
		}
	}

	// Проверяем результат
	camera := gs.GetCameraState()
	if camera.X != 50 {
		t.Errorf("Expected camera X to be 50 after replaying events, got %.1f", camera.X)
	}
}

// TestFullRenderingPipeline проверяет полный пайплайн рендеринга
func TestFullRenderingPipeline(t *testing.T) {
	// Создаем игровое состояние
	config := &gamestate.GameConfig{
		WorldWidth:    640,
		WorldHeight:   480,
		FixedTimeStep: 1.0 / 60.0,
		RandomSeed:    12345,
	}

	gs := gamestate.NewGameState(config)

	// Создаем mock рендер движок
	mockEngine := gamestate.NewMockRenderEngine(800, 600)

	// Создаем рендерер
	gameRenderer := gamestate.NewGameRenderer(mockEngine)

	// Рендерим один кадр
	gameRenderer.RenderFrame(gs)

	// Проверяем, что рендерер был вызван
	mockRenderer := mockEngine.Renderer.(*gamestate.MockRenderer)

	if mockRenderer.PresentCalls != 1 {
		t.Errorf("Expected 1 Present call, got %d", mockRenderer.PresentCalls)
	}

	// Должны быть вызовы SetCamera
	if len(mockRenderer.SetCameraCalls) != 1 {
		t.Errorf("Expected 1 SetCamera call, got %d", len(mockRenderer.SetCameraCalls))
	}

	// Должны быть UI элементы
	if len(mockRenderer.DrawTextCalls) == 0 {
		t.Error("Expected UI text calls")
	}

	// Проверяем, что есть вызовы для рендеринга животных
	if len(mockRenderer.DrawSpriteCalls) == 0 {
		t.Error("Expected sprite calls for animals")
	}
}
