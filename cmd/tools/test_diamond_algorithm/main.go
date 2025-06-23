package main

import (
	"fmt"
	"math"
)

// Тестируем алгоритм рендеринга ромба без GUI
func main() {
	fmt.Println("🔷 Тестирование алгоритма рендеринга ромба")
	fmt.Println("=========================================")

	// Параметры ромба
	const TileWidth = 32
	const TileHeight = 16
	zoom := float32(2.0) // 2x zoom для тестирования

	halfWidth := float32(TileWidth) * zoom / 2   // 32
	halfHeight := float32(TileHeight) * zoom / 2 // 16

	centerX, centerY := float32(50), float32(50)

	fmt.Printf("Ромб: центр (%.0f, %.0f), полуширина %.0f, полувысота %.0f\n",
		centerX, centerY, halfWidth, halfHeight)

	// Копируем исправленный алгоритм
	steps := int(halfHeight)
	if steps > 12 {
		steps = 12
	}
	if steps < 3 {
		steps = 3
	}

	fmt.Printf("Количество линий: %d\n\n", steps)

	// Показываем каждую линию
	totalHeight := halfHeight * 2
	for i := 0; i < steps; i++ {
		progress := float32(i) / float32(steps-1)
		currentY := centerY - halfHeight + progress*totalHeight

		var width float32
		if progress <= 0.5 {
			t := progress * 2
			width = t * halfWidth * 2
		} else {
			t := (progress - 0.5) * 2
			width = (1 - t) * halfWidth * 2
		}

		leftX := centerX - width/2
		rightX := centerX + width/2

		fmt.Printf("Линия %2d: Y=%.1f, ширина=%.1f, от X=%.1f до X=%.1f\n",
			i+1, currentY, width, leftX, rightX)
	}

	// Проверяем что нет дублирования в центре
	fmt.Println("\n🔍 Анализ центральной области:")
	centerLines := 0
	for i := 0; i < steps; i++ {
		progress := float32(i) / float32(steps-1)
		currentY := centerY - halfHeight + progress*totalHeight

		if math.Abs(float64(currentY-centerY)) < 1.0 {
			centerLines++
			fmt.Printf("  Линия рядом с центром: Y=%.1f (центр Y=%.0f)\n", currentY, centerY)
		}
	}

	if centerLines > 1 {
		fmt.Printf("⚠️  ПРОБЛЕМА: %d линий рядом с центром!\n", centerLines)
	} else {
		fmt.Printf("✅ ОК: %d линия в центральной области\n", centerLines)
	}

	// Проверяем покрытие углов ромба
	fmt.Println("\n🔷 Проверка углов ромба:")

	topY := centerY - halfHeight
	bottomY := centerY + halfHeight

	fmt.Printf("Верхний угол: Y=%.1f\n", topY)
	fmt.Printf("Нижний угол: Y=%.1f\n", bottomY)

	// Первая линия должна быть близко к верху
	firstProgress := float32(0) / float32(steps-1)
	firstY := centerY - halfHeight + firstProgress*totalHeight
	fmt.Printf("Первая линия: Y=%.1f (отклонение от верха: %.1f)\n", firstY, math.Abs(float64(firstY-topY)))

	// Последняя линия должна быть близко к низу
	lastProgress := float32(steps-1) / float32(steps-1)
	lastY := centerY - halfHeight + lastProgress*totalHeight
	fmt.Printf("Последняя линия: Y=%.1f (отклонение от низа: %.1f)\n", lastY, math.Abs(float64(lastY-bottomY)))

	fmt.Println("\n✅ Алгоритм протестирован!")
}
