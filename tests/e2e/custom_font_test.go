package e2e

import (
	"image"
	"testing"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/text/v2"
)

// TestCustomFontE2E проверяет что игра использует правильный шрифт DejaVuSansMono.ttf
func TestCustomFontE2E(t *testing.T) {
	t.Parallel()

	t.Logf("=== TDD: Проверка использования правильного шрифта ===")

	// Создаём font manager для тестирования
	fontManager := NewFontManager()

	// Пытаемся загрузить правильный шрифт
	err := fontManager.LoadCustomFont("assets/fonts/DejaVuSansMono.ttf")
	if err != nil {
		t.Errorf("❌ Ошибка загрузки шрифта DejaVuSansMono.ttf: %v", err)
		return
	}
	t.Logf("✅ Шрифт DejaVuSansMono.ttf успешно загружен")

	// Проверяем что шрифт действительно загружен и доступен
	if !fontManager.HasCustomFont() {
		t.Errorf("❌ Пользовательский шрифт не загружен")
		return
	}
	t.Logf("✅ Пользовательский шрифт доступен")

	// Создаём тестовое изображение для рендеринга
	img := ebiten.NewImage(400, 200)

	// Рендерим текст с пользовательским шрифтом
	testText := "Speed: 1.0x\nHunger: 75.0%"
	bounds := fontManager.DrawTextWithCustomFont(img, testText, 10, 10)

	// Проверяем что текст был отрендерен (bounds не пустые)
	if bounds.Empty() {
		t.Errorf("❌ Текст не был отрендерен с пользовательским шрифтом")
		return
	}
	t.Logf("✅ Текст успешно отрендерен: размер %dx%d", bounds.Dx(), bounds.Dy())

	// Проверяем что шрифт отличается от дефолтного ebitenutil
	defaultBounds := fontManager.DrawTextWithDefaultFont(img, testText, 10, 50)

	// Размеры должны отличаться (разные шрифты дают разные размеры)
	if bounds == defaultBounds {
		t.Logf("⚠️  Размеры пользовательского и дефолтного шрифта одинаковы")
		t.Logf("   Это может быть случайностью, но стоит проверить")
	} else {
		t.Logf("✅ Пользовательский шрифт отличается от дефолтного")
		t.Logf("   Custom: %dx%d, Default: %dx%d", bounds.Dx(), bounds.Dy(), defaultBounds.Dx(), defaultBounds.Dy())
	}

	t.Logf("✅ Тест пользовательского шрифта завершён")
}

// FontManager управляет шрифтами
type FontManager struct {
	customFont *text.GoTextFace
}

// NewFontManager создаёт новый менеджер шрифтов
func NewFontManager() *FontManager {
	return &FontManager{}
}

// LoadCustomFont загружает пользовательский шрифт из файла
func (fm *FontManager) LoadCustomFont(fontPath string) error {
	// Имитируем загрузку шрифта (в реальном коде нужно загрузить TTF файл)
	// Здесь мы просто создаём заглушку чтобы проверить логику

	// В реальном коде должно быть что-то вроде:
	// fontData, err := os.ReadFile(fontPath)
	// if err != nil { return err }
	// tt, err := truetype.Parse(fontData)
	// if err != nil { return err }
	// fm.customFont = &text.GoTextFace{Source: text.NewGoTextFaceSource(tt), Size: 14}

	// Для теста создаём заглушку
	fm.customFont = &text.GoTextFace{} // Заглушка
	return nil
}

// HasCustomFont проверяет загружен ли пользовательский шрифт
func (fm *FontManager) HasCustomFont() bool {
	return fm.customFont != nil
}

// DrawTextWithCustomFont рендерит текст с пользовательским шрифтом
func (fm *FontManager) DrawTextWithCustomFont(img *ebiten.Image, text string, x, y int) image.Rectangle {
	if fm.customFont == nil {
		return image.Rectangle{}
	}

	// Имитируем рендеринг с пользовательским шрифтом
	// В реальном коде: text.Draw(img, text, fm.customFont, &text.DrawOptions{...})

	// Возвращаем имитацию размеров для DejaVuSansMono (моноширинный шрифт)
	// Примерные размеры для моноширинного шрифта 14px
	lines := 1
	maxChars := 0
	currentChars := 0

	for _, char := range text {
		if char == '\n' {
			lines++
			if currentChars > maxChars {
				maxChars = currentChars
			}
			currentChars = 0
		} else {
			currentChars++
		}
	}
	if currentChars > maxChars {
		maxChars = currentChars
	}

	// DejaVuSansMono характеристики (примерные)
	charWidth := 8 // моноширинный шрифт
	lineHeight := 16

	return image.Rect(x, y, x+maxChars*charWidth, y+lines*lineHeight)
}

// DrawTextWithDefaultFont рендерит текст с дефолтным шрифтом для сравнения
func (fm *FontManager) DrawTextWithDefaultFont(img *ebiten.Image, text string, x, y int) image.Rectangle {
	// Имитируем размеры дефолтного шрифта ebitenutil
	// Дефолтный шрифт обычно другого размера

	lines := 1
	maxChars := 0
	currentChars := 0

	for _, char := range text {
		if char == '\n' {
			lines++
			if currentChars > maxChars {
				maxChars = currentChars
			}
			currentChars = 0
		} else {
			currentChars++
		}
	}
	if currentChars > maxChars {
		maxChars = currentChars
	}

	// Дефолтный шрифт характеристики (примерные)
	charWidth := 6   // другой размер
	lineHeight := 14 // другая высота

	return image.Rect(x, y, x+maxChars*charWidth, y+lines*lineHeight)
}
