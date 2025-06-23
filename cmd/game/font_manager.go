package main

import (
	"bytes"
	"os"
	"path/filepath"

	"github.com/aiseeq/savanna/internal/constants"
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
// Возвращает error только для критических ошибок, для некритичных использует fallback
func (fm *FontManager) LoadFonts() error {
	// Загружаем пользовательский шрифт DejaVuSansMono для дебаг информации
	fontPath := filepath.Join("assets", "fonts", "DejaVuSansMono.ttf")

	// Читаем файл шрифта
	fontData, err := os.ReadFile(fontPath)
	if err != nil {
		// FALLBACK: Файл шрифта не найден - используем дефолтный шрифт
		fm.debugFont = nil
		return nil // Не критическая ошибка - есть fallback
	}

	// Создаём source из TTF данных
	source, err := text.NewGoTextFaceSource(bytes.NewReader(fontData))
	if err != nil {
		// FALLBACK: Файл повреждён - используем дефолтный шрифт
		fm.debugFont = nil
		return nil // Не критическая ошибка - есть fallback
	}

	// Создаём text.GoTextFace
	fm.debugFont = &text.GoTextFace{
		Source: source,
		Size:   constants.DefaultFontSize,
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

// REMOVED: DrawDebugText - функция была неиспользуемой
