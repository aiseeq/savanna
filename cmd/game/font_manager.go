package main

import (
	"bytes"
	"os"
	"path/filepath"

	"github.com/aiseeq/savanna/internal/simulation"
	"github.com/hajimehoshi/ebiten/v2/text/v2"
)

// FontManager управляет загрузкой и использованием шрифтов
type FontManager struct {
	debugFont *text.GoTextFace // Шрифт для дебаг информации
}

// NewFontManager создаёт новый менеджер шрифтов
func NewFontManager() *FontManager {
	return &FontManager{}
}

// LoadFonts загружает все необходимые шрифты
//
//nolint:unparam // Возвращает error для consistency с интерфейсом, хотя не критично
func (fm *FontManager) LoadFonts() error {
	// Загружаем пользовательский шрифт DejaVuSansMono для дебаг информации
	fontPath := filepath.Join("assets", "fonts", "DejaVuSansMono.ttf")

	// Читаем файл шрифта
	fontData, err := os.ReadFile(fontPath)
	if err != nil {
		// Если не удалось загрузить файл, используем дефолтный шрифт
		fm.debugFont = nil
		return nil // Не критическая ошибка
	}

	// Создаём source из TTF данных
	source, err := text.NewGoTextFaceSource(bytes.NewReader(fontData))
	if err != nil {
		fm.debugFont = nil
		return nil // Не критическая ошибка
	}

	// Создаём text.GoTextFace
	fm.debugFont = &text.GoTextFace{
		Source: source,
		Size:   simulation.DefaultFontSize,
	}

	return nil
}

// GetDebugFont возвращает шрифт для дебаг информации
func (fm *FontManager) GetDebugFont() *text.GoTextFace {
	return fm.debugFont
}

// HasCustomFont проверяет загружен ли пользовательский шрифт
func (fm *FontManager) HasCustomFont() bool {
	return fm.debugFont != nil
}

// DrawDebugText рендерит дебаг текст с правильным шрифтом
func (fm *FontManager) DrawDebugText(screen interface{}, textStr string, x, y int) {
	// Эта функция будет использоваться в drawUI для отрисовки текста
	// Реализация будет добавлена в следующем коммите
}
