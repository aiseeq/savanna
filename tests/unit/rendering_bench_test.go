package unit

import (
	"image/color"
	"testing"
)

// BenchmarkDiamondRendering измеряет производительность изометрических вычислений
func BenchmarkDiamondRendering(b *testing.B) {
	// Тестируем изометрические преобразования координат без GUI зависимостей
	cameraX, cameraY := float32(100), float32(100)
	zoom := float32(1.0)

	// Цвет тайла
	tileColor := color.RGBA{R: 50, G: 150, B: 50, A: 255}

	b.ResetTimer()

	// Симулируем изометрические преобразования большого количества тайлов
	for i := 0; i < b.N; i++ {
		// Рендерим сетку 20x20 тайлов (400 тайлов)
		for y := 0; y < 20; y++ {
			for x := 0; x < 20; x++ {
				// Изометрическое преобразование world->screen
				baseScreenX, baseScreenY := worldToScreen(float32(x), float32(y))
				screenX := baseScreenX*zoom - cameraX
				screenY := baseScreenY*zoom - cameraY

				// Тестируем вычисления рендеринга без GUI
				renderTestTile(screenX, screenY, tileColor, zoom)
			}
		}
	}
}

// Используем функции worldToScreen и screenToWorld из isometric_test.go

// renderTestTile - тестовая версия вычислений рендеринга тайла (без GUI)
func renderTestTile(x, y float32, col color.RGBA, zoom float32) {
	// Копируем оптимизированную логику изометрических вычислений
	if zoom < 0.5 {
		// При маленьком zoom минимальные вычисления
		return
	}

	if zoom < 1.0 {
		// При среднем zoom простые вычисления
		size := zoom * 8
		_ = size
		return
	}

	// При крупном zoom полные изометрические вычисления
	const TileWidth = 32
	const TileHeight = 16

	halfWidth := float32(TileWidth) * zoom / 2
	halfHeight := float32(TileHeight) * zoom / 2

	steps := int(halfHeight)
	if steps > 16 {
		steps = 16
	}
	if steps < 2 {
		steps = 2
	}

	// Эмулируем изометрические вычисления
	for i := 0; i < steps; i++ {
		t := float32(i) / float32(steps-1)

		// Верхняя часть ромба
		topY := y - halfHeight + t*halfHeight
		topWidth := t * halfWidth * 2

		// Нижняя часть ромба
		bottomY := y + t*halfHeight
		bottomWidth := (1 - t) * halfWidth * 2

		// Используем результаты вычислений
		_ = topY + bottomY + topWidth + bottomWidth
	}
}

// BenchmarkCoordinateTransform измеряет производительность преобразований координат
func BenchmarkCoordinateTransform(b *testing.B) {
	cameraX, cameraY := float32(100), float32(100)
	zoom := float32(1.0)

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		// Тестируем 1000 преобразований координат
		for j := 0; j < 1000; j++ {
			x, y := float32(j%100), float32(j/100)

			// World -> Screen с zoom и камерой
			baseScreenX, baseScreenY := worldToScreen(x, y)
			screenX := baseScreenX*zoom - cameraX
			screenY := baseScreenY*zoom - cameraY

			// Screen -> World (обратное преобразование)
			isoX := (screenX + cameraX) / zoom
			isoY := (screenY + cameraY) / zoom
			_, _ = screenToWorld(isoX, isoY)
		}
	}
}

// BenchmarkFrustumCulling измеряет производительность culling'а видимых тайлов
func BenchmarkFrustumCulling(b *testing.B) {
	cameraX, cameraY := float32(100), float32(100)
	zoom := float32(1.0)

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		// Тестируем функцию определения видимых тайлов
		screenWidth := float32(800)
		screenHeight := float32(600)

		// Углы экрана в мировых координатах (с учетом камеры и zoom)
		_, _ = screenToWorld(cameraX/zoom, cameraY/zoom)
		_, _ = screenToWorld((screenWidth+cameraX)/zoom, cameraY/zoom)
		_, _ = screenToWorld(cameraX/zoom, (screenHeight+cameraY)/zoom)
		_, _ = screenToWorld((screenWidth+cameraX)/zoom, (screenHeight+cameraY)/zoom)

		// Эмулируем расчет границ видимых тайлов
		tileCount := 0
		for tx := 0; tx < 100; tx++ {
			for ty := 0; ty < 100; ty++ {
				baseX, baseY := worldToScreen(float32(tx), float32(ty))
				sx := baseX*zoom - cameraX
				sy := baseY*zoom - cameraY
				if sx >= 0 && sx <= screenWidth && sy >= 0 && sy <= screenHeight {
					tileCount++
				}
			}
		}
		_ = tileCount
	}
}
