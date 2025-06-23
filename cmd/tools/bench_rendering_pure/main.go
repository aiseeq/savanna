package main

import (
	"fmt"
	"image/color"
	"runtime"
	"time"
)

// Бенчмарк алгоритмов рендеринга без GUI

const (
	TileWidth  = 32
	TileHeight = 16
)

// Эмуляция старого неэффективного алгоритма
func oldDiamondAlgorithm(x, y, zoom float32, col color.RGBA) int {
	halfWidth := float32(TileWidth) * zoom / 2
	halfHeight := float32(TileHeight) * zoom / 2

	drawCalls := 0

	// Старый алгоритм с шагом 0.5
	for dy := -halfHeight; dy <= halfHeight; dy += 0.5 {
		_ = y + dy // currentY (не используется в эмуляции)

		var leftEdgeX, rightEdgeX float32
		if dy <= 0 {
			t := (dy + halfHeight) / halfHeight
			leftEdgeX = x + t*(x-halfWidth-x)
			rightEdgeX = x + t*(x+halfWidth-x)
		} else {
			t := dy / halfHeight
			leftEdgeX = (x - halfWidth) + t*(x-(x-halfWidth))
			rightEdgeX = (x + halfWidth) + t*(x-(x+halfWidth))
		}

		if rightEdgeX > leftEdgeX {
			drawCalls++ // Эмуляция StrokeLine
		}
	}

	// Границы (4 линии)
	if zoom > 0.3 {
		drawCalls += 4
	}

	return drawCalls
}

// Новый оптимизированный алгоритм
func newDiamondAlgorithm(x, y, zoom float32, col color.RGBA) int {
	if zoom < 0.5 {
		return 1 // Точка
	}

	if zoom < 1.0 {
		return 1 // Прямоугольник
	}

	_ = float32(TileWidth) * zoom / 2 // halfWidth (не используется в эмуляции)
	halfHeight := float32(TileHeight) * zoom / 2

	steps := int(halfHeight)
	if steps > 12 {
		steps = 12
	}
	if steps < 3 {
		steps = 3
	}

	drawCalls := steps // Линии заливки

	// Границы только при крупном zoom
	if zoom > 1.5 {
		drawCalls += 4
	}

	return drawCalls
}

// Бенчмарк функция
func benchmarkAlgorithm(name string, algorithm func(float32, float32, float32, color.RGBA) int, iterations int) time.Duration {
	col := color.RGBA{R: 50, G: 150, B: 50, A: 255}
	totalDrawCalls := 0

	start := time.Now()

	for i := 0; i < iterations; i++ {
		// Симулируем рендеринг карты 50x50 тайлов
		for y := 0; y < 50; y++ {
			for x := 0; x < 50; x++ {
				zoom := float32(1.0 + float32(i%4)*0.5) // Различные zoom уровни
				totalDrawCalls += algorithm(float32(x*32), float32(y*16), zoom, col)
			}
		}
	}

	elapsed := time.Since(start)

	fmt.Printf("%s: %d итераций, %d draw calls, время: %v\n",
		name, iterations, totalDrawCalls, elapsed)

	return elapsed
}

func main() {
	runtime.GOMAXPROCS(runtime.NumCPU())

	fmt.Println("🚀 Бенчмарк алгоритмов рендеринга ромбов")
	fmt.Println("=====================================")

	iterations := 1000
	fmt.Printf("Тестируем %d итераций рендеринга карты 50x50 тайлов\n\n", iterations)

	// Прогрев CPU
	fmt.Println("🔥 Прогрев CPU...")
	benchmarkAlgorithm("Прогрев", newDiamondAlgorithm, 100)

	fmt.Println("\n📊 Основные тесты:")

	// Тестируем старый алгоритм
	oldTime := benchmarkAlgorithm("Старый алгоритм (64+ линий)", oldDiamondAlgorithm, iterations)

	// Тестируем новый алгоритм
	newTime := benchmarkAlgorithm("Новый алгоритм (LOD + лимиты)", newDiamondAlgorithm, iterations)

	// Анализ результатов
	fmt.Println("\n📈 Анализ производительности:")
	fmt.Printf("Старый алгоритм: %v\n", oldTime)
	fmt.Printf("Новый алгоритм:  %v\n", newTime)

	if newTime < oldTime {
		speedup := float64(oldTime) / float64(newTime)
		fmt.Printf("🚀 Ускорение: %.2fx (на %.1f%% быстрее)\n", speedup, (speedup-1)*100)
	} else {
		slowdown := float64(newTime) / float64(oldTime)
		fmt.Printf("🐌 Замедление: %.2fx (на %.1f%% медленнее)\n", slowdown, (slowdown-1)*100)
	}

	// Детальный анализ draw calls по zoom уровням
	fmt.Println("\n🔍 Анализ draw calls по zoom уровням:")

	col := color.RGBA{R: 50, G: 150, B: 50, A: 255}
	zooms := []float32{0.25, 0.5, 1.0, 1.5, 2.0, 4.0}

	fmt.Printf("%-10s %-15s %-15s %-15s\n", "Zoom", "Старый", "Новый", "Экономия")
	fmt.Println("--------------------------------------------------------")

	for _, zoom := range zooms {
		oldCalls := oldDiamondAlgorithm(100, 100, zoom, col)
		newCalls := newDiamondAlgorithm(100, 100, zoom, col)
		savings := float64(oldCalls-newCalls) / float64(oldCalls) * 100

		fmt.Printf("%-10.2f %-15d %-15d %-14.1f%%\n", zoom, oldCalls, newCalls, savings)
	}

	fmt.Println("\n✅ Бенчмарк завершен!")
}
