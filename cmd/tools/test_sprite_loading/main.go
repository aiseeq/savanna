package main

import (
	"fmt"
	"os"
	"path/filepath"
)

func main() {
	fmt.Println("🎮 Тестирование загрузки спрайтов")
	fmt.Println("=================================")

	// Проверяем структуру директорий
	assetsDir := "assets/animations"

	// Проверяем что директория существует
	if _, err := os.Stat(assetsDir); os.IsNotExist(err) {
		fmt.Printf("❌ Директория %s не существует!\n", assetsDir)
		return
	}

	fmt.Printf("✅ Директория %s найдена\n", assetsDir)

	// Анимации которые должны загружаться
	expectedSprites := []struct {
		prefix   string
		animName string
		frames   int
	}{
		{"hare", "idle", 2},
		{"hare", "walk", 2},
		{"hare", "run", 2},
		{"hare", "attack", 2},
		{"hare", "eat", 2},
		{"hare", "dead", 2},
		{"wolf", "idle", 2},
		{"wolf", "walk", 2},
		{"wolf", "run", 2},
		{"wolf", "attack", 2},
		{"wolf", "eat", 2},
		{"wolf", "dead", 2},
	}

	missingFiles := 0
	totalFiles := 0

	fmt.Println("\n🔍 Проверка файлов спрайтов:")

	for _, sprite := range expectedSprites {
		for frame := 1; frame <= sprite.frames; frame++ {
			filename := fmt.Sprintf("%s_%s_%d.png", sprite.prefix, sprite.animName, frame)
			filePath := filepath.Join(assetsDir, filename)
			totalFiles++

			if _, err := os.Stat(filePath); os.IsNotExist(err) {
				fmt.Printf("❌ ОТСУТСТВУЕТ: %s\n", filename)
				missingFiles++
			} else {
				fmt.Printf("✅ OK: %s\n", filename)
			}
		}
	}

	fmt.Printf("\n📊 Результат: %d/%d файлов найдено\n", totalFiles-missingFiles, totalFiles)

	if missingFiles > 0 {
		fmt.Printf("⚠️  %d файлов отсутствует - спрайты будут заменены на fallback\n", missingFiles)
	} else {
		fmt.Println("🎉 Все файлы спрайтов найдены!")
	}

	// Дополнительная проверка - какие файлы реально есть в директории
	fmt.Println("\n📁 Файлы в assets/animations:")
	files, err := os.ReadDir(assetsDir)
	if err != nil {
		fmt.Printf("Ошибка чтения директории: %v\n", err)
		return
	}

	for _, file := range files {
		if !file.IsDir() && filepath.Ext(file.Name()) == ".png" {
			fmt.Printf("  📄 %s\n", file.Name())
		}
	}
}
