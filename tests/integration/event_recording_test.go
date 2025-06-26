package integration

import (
	"testing"

	"github.com/aiseeq/savanna/internal/gamestate"
	testing_mocks "github.com/aiseeq/savanna/tests/mocks"
)

// TestScrollBehaviorRecording тестирует скролл через запись-воспроизведение
func TestScrollBehaviorRecording(t *testing.T) {
	// Сценарий: Пользователь делает скролл правой кнопкой мыши
	scrollScript := []gamestate.InputEvent{
		// Нажимаем правую кнопку мыши
		{
			Type:   gamestate.InputMouseDown,
			Button: gamestate.MouseButtonRight,
			X:      100,
			Y:      100,
		},
		// Двигаем мышь (симулируем drag)
		{
			Type: gamestate.InputMouseMove,
			X:    150, // +50 по X
			Y:    120, // +20 по Y
		},
		{
			Type: gamestate.InputMouseMove,
			X:    200, // еще +50 по X
			Y:    140, // еще +20 по Y
		},
		// Отпускаем кнопку
		{
			Type:   gamestate.InputMouseUp,
			Button: gamestate.MouseButtonRight,
			X:      200,
			Y:      140,
		},
		// Дополнительное движение - НЕ должно влиять на камеру
		{
			Type: gamestate.InputMouseMove,
			X:    300,
			Y:    200,
		},
	}

	// Создаем игровое состояние
	config := &gamestate.GameConfig{
		WorldWidth:    640,
		WorldHeight:   480,
		FixedTimeStep: 1.0 / 60.0,
		RandomSeed:    12345,
	}

	gs := gamestate.NewGameState(config)

	// Начальное состояние камеры
	initialCamera := gs.GetCameraState()
	if initialCamera.X != 0 || initialCamera.Y != 0 {
		t.Errorf("Initial camera should be at (0,0), got (%.1f,%.1f)", initialCamera.X, initialCamera.Y)
	}

	// Воспроизводим скрипт
	mockProvider := testing_mocks.NewMockInputProvider(scrollScript)

	for i := 0; i < len(scrollScript); i++ {
		events := mockProvider.PollEvents()
		if len(events) > 0 {
			gs.ProcessInput(events)
		}
	}

	// Проверяем финальное состояние камеры
	finalCamera := gs.GetCameraState()

	// Ожидаемое смещение: (150-100) + (200-150) = 50 + 50 = 100 по X
	//                     (120-100) + (140-120) = 20 + 20 = 40 по Y
	expectedX := 100.0
	expectedY := 40.0

	if finalCamera.X != expectedX {
		t.Errorf("Expected camera X to be %.1f, got %.1f", expectedX, finalCamera.X)
	}

	if finalCamera.Y != expectedY {
		t.Errorf("Expected camera Y to be %.1f, got %.1f", expectedY, finalCamera.Y)
	}

	// Камера не должна скроллиться после отпускания кнопки
	if finalCamera.IsScrolling {
		t.Error("Camera should not be scrolling after mouse up")
	}
}

// TestScrollBehaviorEdgeCases тестирует граничные случаи скролла
func TestScrollBehaviorEdgeCases(t *testing.T) {
	config := &gamestate.GameConfig{
		WorldWidth:    640,
		WorldHeight:   480,
		FixedTimeStep: 1.0 / 60.0,
		RandomSeed:    12345,
	}

	gs := gamestate.NewGameState(config)

	// Тест 1: Движение мыши без нажатой кнопки НЕ должно скроллить
	t.Run("MoveWithoutButton", func(t *testing.T) {
		events := []gamestate.InputEvent{
			{Type: gamestate.InputMouseMove, X: 100, Y: 100},
		}

		gs.ProcessInput(events)
		camera := gs.GetCameraState()

		if camera.X != 0 || camera.Y != 0 {
			t.Errorf("Camera should remain at (0,0) without button press, got (%.1f,%.1f)", camera.X, camera.Y)
		}
	})

	// Тест 2: Левая кнопка мыши НЕ должна скроллить (только правая)
	t.Run("LeftButtonScroll", func(t *testing.T) {
		events := []gamestate.InputEvent{
			{Type: gamestate.InputMouseDown, Button: gamestate.MouseButtonLeft, X: 100, Y: 100},
			{Type: gamestate.InputMouseMove, X: 200, Y: 200},
			{Type: gamestate.InputMouseUp, Button: gamestate.MouseButtonLeft, X: 200, Y: 200},
		}

		gs.ProcessInput(events)
		camera := gs.GetCameraState()

		if camera.X != 0 || camera.Y != 0 {
			t.Errorf("Left button should not scroll camera, got (%.1f,%.1f)", camera.X, camera.Y)
		}
	})

	// Тест 3: Отпускание неправильной кнопки НЕ должно прекращать скролл
	t.Run("WrongButtonUp", func(t *testing.T) {
		events := []gamestate.InputEvent{
			{Type: gamestate.InputMouseDown, Button: gamestate.MouseButtonRight, X: 100, Y: 100},
			{Type: gamestate.InputMouseUp, Button: gamestate.MouseButtonLeft, X: 100, Y: 100}, // Неправильная кнопка
			{Type: gamestate.InputMouseMove, X: 150, Y: 150},
		}

		gs.ProcessInput(events)
		camera := gs.GetCameraState()

		// Скролл должен продолжаться
		if !camera.IsScrolling {
			t.Error("Camera should still be scrolling after wrong button up")
		}

		if camera.X != 50 || camera.Y != 50 {
			t.Errorf("Camera should have moved to (50,50), got (%.1f,%.1f)", camera.X, camera.Y)
		}
	})
}

// TestComplexScrollScenario тестирует сложный сценарий скролла
func TestComplexScrollScenario(t *testing.T) {
	// Сценарий: Несколько отдельных скроллов
	scenario := []gamestate.InputEvent{
		// Первый скролл
		{Type: gamestate.InputMouseDown, Button: gamestate.MouseButtonRight, X: 100, Y: 100},
		{Type: gamestate.InputMouseMove, X: 120, Y: 110},
		{Type: gamestate.InputMouseUp, Button: gamestate.MouseButtonRight, X: 120, Y: 110},

		// Пауза (движение без кнопки)
		{Type: gamestate.InputMouseMove, X: 200, Y: 200},

		// Второй скролл из нового положения
		{Type: gamestate.InputMouseDown, Button: gamestate.MouseButtonRight, X: 200, Y: 200},
		{Type: gamestate.InputMouseMove, X: 180, Y: 190}, // Скролл в обратную сторону
		{Type: gamestate.InputMouseUp, Button: gamestate.MouseButtonRight, X: 180, Y: 190},
	}

	config := &gamestate.GameConfig{
		WorldWidth:    640,
		WorldHeight:   480,
		FixedTimeStep: 1.0 / 60.0,
		RandomSeed:    12345,
	}

	gs := gamestate.NewGameState(config)
	mockProvider := testing_mocks.NewMockInputProvider(scenario)

	// Воспроизводим сценарий
	for i := 0; i < len(scenario); i++ {
		events := mockProvider.PollEvents()
		if len(events) > 0 {
			gs.ProcessInput(events)
		}
	}

	camera := gs.GetCameraState()

	// Первый скролл: +20 по X, +10 по Y
	// Второй скролл: -20 по X, -10 по Y
	// Итого: 0 по X, 0 по Y
	expectedX := 0.0
	expectedY := 0.0

	if camera.X != expectedX {
		t.Errorf("Expected final camera X to be %.1f, got %.1f", expectedX, camera.X)
	}

	if camera.Y != expectedY {
		t.Errorf("Expected final camera Y to be %.1f, got %.1f", expectedY, camera.Y)
	}

	if camera.IsScrolling {
		t.Error("Camera should not be scrolling at the end")
	}
}

// TestScrollWithGameUpdate тестирует скролл во время обновления игры
func TestScrollWithGameUpdate(t *testing.T) {
	config := &gamestate.GameConfig{
		WorldWidth:    640,
		WorldHeight:   480,
		FixedTimeStep: 1.0 / 60.0,
		RandomSeed:    12345,
	}

	gs := gamestate.NewGameState(config)

	// Начинаем скролл
	startScroll := []gamestate.InputEvent{
		{Type: gamestate.InputMouseDown, Button: gamestate.MouseButtonRight, X: 100, Y: 100},
		{Type: gamestate.InputMouseMove, X: 150, Y: 150},
	}

	gs.ProcessInput(startScroll)

	// Проверяем, что камера в состоянии скролла
	camera := gs.GetCameraState()
	if !camera.IsScrolling {
		t.Error("Camera should be scrolling")
	}

	// Обновляем игру несколько раз (симулируем игровой цикл)
	for i := 0; i < 10; i++ {
		gs.Update()
	}

	// Состояние скролла должно сохраниться
	camera = gs.GetCameraState()
	if !camera.IsScrolling {
		t.Error("Camera should still be scrolling after game updates")
	}

	// Заканчиваем скролл
	endScroll := []gamestate.InputEvent{
		{Type: gamestate.InputMouseUp, Button: gamestate.MouseButtonRight, X: 150, Y: 150},
	}

	gs.ProcessInput(endScroll)

	// Проверяем завершение скролла
	camera = gs.GetCameraState()
	if camera.IsScrolling {
		t.Error("Camera should not be scrolling after mouse up")
	}

	// Финальная позиция должна быть корректной
	if camera.X != 50 || camera.Y != 50 {
		t.Errorf("Expected camera at (50,50), got (%.1f,%.1f)", camera.X, camera.Y)
	}
}
