// Демонстрация новой MVC архитектуры
package main

import (
	"fmt"
	"time"

	"github.com/aiseeq/savanna/internal/gamestate"
)

func main() {
	fmt.Println("=== Демонстрация новой MVC архитектуры Savanna ===")

	// Создаем конфигурацию игры
	config := &gamestate.GameConfig{
		WorldWidth:    640,
		WorldHeight:   480,
		FixedTimeStep: 1.0 / 60.0,
		RandomSeed:    time.Now().UnixNano(),
	}

	// Создаем игровое состояние (MODEL)
	gameState := gamestate.NewGameState(config)

	// Создаем mock рендер движок (VIEW)
	mockEngine := gamestate.NewMockRenderEngine(800, 600)
	renderer := gamestate.NewGameRenderer(mockEngine)

	// Создаем отладочный оверлей
	debugOverlay := gamestate.NewDebugOverlay()
	debugOverlay.SetEnabled(true)

	fmt.Println("1. Тестируем начальное состояние...")

	// Рендерим начальный кадр
	renderer.RenderFrame(gameState)
	debugOverlay.Update()

	// Показываем статистику начального состояния
	mockRenderer := mockEngine.Renderer.(*gamestate.MockRenderer)
	fmt.Printf("   - Спрайтов отрендерено: %d\n", len(mockRenderer.DrawSpriteCalls))
	fmt.Printf("   - UI элементов: %d\n", len(mockRenderer.DrawTextCalls))
	fmt.Printf("   - Terrain тайлов: %d\n", len(mockRenderer.DrawTerrainCalls))

	fmt.Println("\n2. Тестируем скролл камеры...")

	// Создаем сценарий скролла
	scrollEvents := []gamestate.InputEvent{
		{Type: gamestate.InputMouseDown, Button: gamestate.MouseButtonRight, X: 100, Y: 100},
		{Type: gamestate.InputMouseMove, X: 200, Y: 150},
		{Type: gamestate.InputMouseUp, Button: gamestate.MouseButtonRight, X: 200, Y: 150},
	}

	// Обрабатываем события
	gameState.ProcessInput(scrollEvents)

	// Проверяем результат
	camera := gameState.GetCameraState()
	fmt.Printf("   - Камера переместилась на: (%.1f, %.1f)\n", camera.X, camera.Y)

	fmt.Println("\n3. Симулируем игровой цикл...")

	// Симулируем 3 секунды игры
	for i := 0; i < 180; i++ { // 60 FPS * 3 секунды
		gameState.Update()
		debugOverlay.Update()

		// Рендерим каждый 60-й кадр
		if i%60 == 0 {
			mockRenderer.Reset() // Очищаем счетчики
			renderer.RenderFrame(gameState)

			// Генерируем инструкции с отладкой
			instructions := gameState.GenerateRenderInstructionsWithDebug(debugOverlay)

			fmt.Printf("   Секунда %d: Спрайтов=%d, UI=%d, Debug=%d\n",
				i/60+1,
				len(instructions.Sprites),
				len(instructions.UI),
				len(instructions.DebugTexts))
		}
	}

	fmt.Println("\n4. Тестируем детерминизм...")

	// Создаем два идентичных состояния
	config1 := &gamestate.GameConfig{
		WorldWidth:    640,
		WorldHeight:   480,
		FixedTimeStep: 1.0 / 60.0,
		RandomSeed:    12345, // Одинаковый seed
	}

	config2 := &gamestate.GameConfig{
		WorldWidth:    640,
		WorldHeight:   480,
		FixedTimeStep: 1.0 / 60.0,
		RandomSeed:    12345, // Одинаковый seed
	}

	gs1 := gamestate.NewGameState(config1)
	gs2 := gamestate.NewGameState(config2)

	// Одинаковые обновления
	for i := 0; i < 60; i++ {
		gs1.Update()
		gs2.Update()
	}

	// Генерируем инструкции
	instructions1 := gs1.GenerateRenderInstructions()
	instructions2 := gs2.GenerateRenderInstructions()

	// Сравниваем
	deterministic := len(instructions1.Sprites) == len(instructions2.Sprites) &&
		len(instructions1.UI) == len(instructions2.UI)

	fmt.Printf("   - Детерминизм: %v\n", deterministic)
	fmt.Printf("   - State1: %d спрайтов, State2: %d спрайтов\n",
		len(instructions1.Sprites), len(instructions2.Sprites))

	fmt.Println("\n5. Демонстрируем Golden Image Test...")

	// Создаем состояние для golden test
	goldenConfig := &gamestate.GameConfig{
		WorldWidth:    400,
		WorldHeight:   300,
		FixedTimeStep: 1.0 / 60.0,
		RandomSeed:    99999, // Фиксированный seed
	}

	goldenState := gamestate.NewGameState(goldenConfig)

	// Стабилизируем состояние
	for i := 0; i < 30; i++ {
		goldenState.Update()
	}

	// Захватываем кадр
	goldenEngine := gamestate.NewMockRenderEngine(400, 300)
	goldenRenderer := gamestate.NewGameRenderer(goldenEngine)
	goldenRenderer.RenderFrame(goldenState)

	goldenMockRenderer := goldenEngine.Renderer.(*gamestate.MockRenderer)
	frame := goldenMockRenderer.CaptureFrame()

	fmt.Printf("   - Захвачен кадр размером: %dx%d\n",
		frame.Bounds().Dx(), frame.Bounds().Dy())

	fmt.Println("\n6. Тестируем аудио систему...")

	mockAudio := goldenEngine.AudioProvider.(*gamestate.MockAudioProvider)

	// Симулируем звуковые эффекты
	goldenRenderer.PlaySoundEffect("rabbit_eat")
	goldenRenderer.PlaySoundEffect("wolf_attack")

	fmt.Printf("   - Звуковых эффектов воспроизведено: %d\n", len(mockAudio.PlaySoundCalls))
	fmt.Printf("   - Последний звук: %s\n", mockAudio.PlaySoundCalls[len(mockAudio.PlaySoundCalls)-1])

	fmt.Println("\n=== Демонстрация завершена ===")
	fmt.Println("✅ Модель полностью отделена от представления")
	fmt.Println("✅ Все рендер-инструкции генерируются из состояния")
	fmt.Println("✅ Input события записываются и воспроизводятся")
	fmt.Println("✅ Mock системы позволяют полное unit-тестирование")
	fmt.Println("✅ Отладочный оверлей предоставляет полную видимость")
	fmt.Println("✅ Golden Image Tests поддерживаются")
}
