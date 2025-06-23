package main

import (
	"fmt"
	"image"
	"image/color"
	"path/filepath"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/hajimehoshi/ebiten/v2/vector"

	"github.com/aiseeq/savanna/internal/animation"
	"github.com/aiseeq/savanna/internal/constants"
	"github.com/aiseeq/savanna/internal/core"
)

// SpriteRenderer отвечает за загрузку и отрисовку спрайтов животных
type SpriteRenderer struct {
	// Загруженные спрайты для каждого типа животного
	animalSprites map[core.AnimalType]AnimalSprites

	// Переиспользуемые объекты для оптимизации производительности
	drawOptions *ebiten.DrawImageOptions // Переиспользуемый объект для отрисовки
}

// AnimalSprites содержит все спрайты для одного типа животного
type AnimalSprites struct {
	// Спрайты по типам анимации (каждый содержит все кадры)
	animations map[animation.AnimationType][]*ebiten.Image
}

// NewSpriteRenderer создаёт новый рендерер спрайтов
func NewSpriteRenderer() *SpriteRenderer {
	sr := &SpriteRenderer{
		animalSprites: make(map[core.AnimalType]AnimalSprites),
		drawOptions:   &ebiten.DrawImageOptions{}, // Инициализируем переиспользуемый объект
	}

	// Загружаем спрайты для всех типов животных
	sr.loadAnimalSprites(core.TypeRabbit, "hare")
	sr.loadAnimalSprites(core.TypeWolf, "wolf")

	return sr
}

// loadAnimalSprites загружает все спрайты для указанного типа животного
func (sr *SpriteRenderer) loadAnimalSprites(animalType core.AnimalType, prefix string) {

	sprites := AnimalSprites{
		animations: make(map[animation.AnimationType][]*ebiten.Image),
	}

	// Определяем какие анимации загружать (соответствуют реальным файлам в assets)
	animationTypes := []struct {
		animType animation.AnimationType
		name     string
		frames   int
	}{
		{animation.AnimIdle, "idle", 2},
		{animation.AnimWalk, "walk", 2},
		{animation.AnimRun, "run", 2},
		{animation.AnimAttack, "attack", 2},
		{animation.AnimEat, "eat", 2},
		{animation.AnimDeathDying, "dead", 2},
	}

	// Загружаем каждую анимацию
	for _, anim := range animationTypes {
		sprites.animations[anim.animType] = sr.loadAnimationFrames(prefix, anim.name, anim.frames)
	}

	sr.animalSprites[animalType] = sprites
}

// loadAnimationFrames загружает кадры одной анимации
func (sr *SpriteRenderer) loadAnimationFrames(prefix, animName string, frameCount int) []*ebiten.Image {
	// ОПТИМИЗАЦИЯ: Предварительно выделяем слайс нужного размера
	frames := make([]*ebiten.Image, 0, frameCount)

	for i := 1; i <= frameCount; i++ {
		filename := fmt.Sprintf("%s_%s_%d.png", prefix, animName, i)
		filePath := filepath.Join("assets", "animations", filename)

		img, _, err := ebitenutil.NewImageFromFile(filePath)
		if err != nil {
			// FALLBACK: Создаём fallback спрайт при ошибке загрузки
			fmt.Printf("⚠️  Спрайт не найден: %s (error: %v)\n", filePath, err)
			img = sr.createFallbackSprite(constants.DefaultSpriteSize, constants.DefaultSpriteSize)
		}

		frames = append(frames, img)
	}

	return frames
}

// createFallbackSprite создаёт простой цветной спрайт как fallback
func (sr *SpriteRenderer) createFallbackSprite(width, height int) *ebiten.Image {
	img := ebiten.NewImage(width, height)
	img.Fill(color.RGBA{255, 0, 255, 255}) // Пурпурный цвет для отладки
	return img
}

// RenderParams параметры отрисовки животного
type RenderParams struct {
	ScreenX, ScreenY, Zoom float32
}

// DrawAnimal отрисовывает животное с правильным спрайтом и анимацией
func (sr *SpriteRenderer) DrawAnimal(
	screen *ebiten.Image,
	world *core.World,
	entity core.EntityID,
	params RenderParams,
) {
	// БЕЗОПАСНОСТЬ: Проверяем входные параметры
	if screen == nil || world == nil {
		return
	}

	// Получаем тип животного
	animalType, hasType := world.GetAnimalType(entity)
	if !hasType {
		return
	}

	// Получаем анимацию
	anim, hasAnim := world.GetAnimation(entity)
	if !hasAnim {
		return
	}

	// Получаем спрайты для этого типа животного
	sprites, hasSprites := sr.animalSprites[animalType]
	if !hasSprites {
		return
	}

	// Получаем кадры для текущей анимации
	animType := animation.AnimationType(anim.CurrentAnim)

	frames, hasFrames := sprites.animations[animType]
	if !hasFrames || len(frames) == 0 {
		return
	}

	// Выбираем правильный кадр
	frameIndex := anim.Frame
	if frameIndex >= len(frames) {
		frameIndex = len(frames) - 1
	}
	if frameIndex < 0 {
		frameIndex = 0
	}

	sprite := frames[frameIndex]

	// ОПТИМИЗАЦИЯ: Используем переиспользуемый объект вместо создания нового
	// БЕЗОПАСНОСТЬ: Проверяем что объект инициализирован
	if sr.drawOptions == nil {
		sr.drawOptions = &ebiten.DrawImageOptions{}
	}

	op := sr.drawOptions
	op.GeoM.Reset()       // Сбрасываем матрицу трансформации
	op.ColorScale.Reset() // Сбрасываем цветовые эффекты

	// Масштабирование (разное для разных животных)
	var spriteScale float64
	if animalType == core.TypeRabbit {
		spriteScale = float64(params.Zoom) * constants.RabbitSpriteScale // Масштаб спрайта зайца
	} else {
		spriteScale = float64(params.Zoom) * constants.WolfSpriteScale // Масштаб спрайта волка
	}
	op.GeoM.Scale(spriteScale, spriteScale)

	// Отражение по горизонтали если животное смотрит влево
	if !anim.FacingRight {
		// Отражаем спрайт
		spriteWidth := float64(sprite.Bounds().Dx())
		op.GeoM.Scale(-1, 1)
		op.GeoM.Translate(spriteWidth*spriteScale, 0)
	}

	// Центрируем спрайт относительно позиции животного
	spriteWidth := float64(sprite.Bounds().Dx()) * spriteScale
	spriteHeight := float64(sprite.Bounds().Dy()) * spriteScale
	op.GeoM.Translate(
		float64(params.ScreenX)-spriteWidth/2,
		float64(params.ScreenY)-spriteHeight/2,
	)

	// ИСПРАВЛЕНИЕ: Применяем DamageFlash эффект к самому спрайту
	sr.applyDamageFlash(world, entity, op)

	// Рисуем спрайт
	screen.DrawImage(sprite, op)
}

// DrawAnimalAt отрисовывает животное с указанными экранными координатами (для IsometricRenderer)
func (sr *SpriteRenderer) DrawAnimalAt(
	screen *ebiten.Image,
	world *core.World,
	entity core.EntityID,
	screenX, screenY, zoom float32,
) {
	// Рисуем спрайт животного
	sr.DrawAnimal(screen, world, entity, RenderParams{
		ScreenX: screenX,
		ScreenY: screenY,
		Zoom:    zoom,
	})

	// ДОБАВЛЕНО: Отрисовываем UI элементы (health bar и hunger text) всегда
	sr.drawUIElements(screen, world, entity, screenX, screenY)
}

// GetSpriteBounds возвращает размеры спрайта для животного (для расчёта коллизий)
func (sr *SpriteRenderer) GetSpriteBounds(animalType core.AnimalType) image.Rectangle {
	sprites, hasSprites := sr.animalSprites[animalType]
	if !hasSprites {
		return image.Rectangle{}
	}

	// БЕЗОПАСНОСТЬ: Берём первый кадр idle анимации для получения размера
	if frames, hasIdle := sprites.animations[animation.AnimIdle]; hasIdle && frames != nil && len(frames) > 0 {
		return frames[0].Bounds()
	}

	return image.Rectangle{}
}

// applyDamageFlash применяет эффект мерцания к спрайту животного
func (sr *SpriteRenderer) applyDamageFlash(world *core.World, entity core.EntityID, op *ebiten.DrawImageOptions) {
	flash, hasFlash := world.GetDamageFlash(entity)
	if !hasFlash {
		return
	}

	// ИСПРАВЛЕНИЕ: Другой подход - увеличиваем яркость всех каналов
	// При интенсивности 1.0 все цвета становятся максимально яркими (белыми)
	// При интенсивности 0.0 цвета остаются нормальными
	intensity := flash.Intensity

	// Увеличиваем масштаб всех цветовых каналов с усилением эффекта
	// Формула: оригинальный цвет * (1 + intensity * multiplier)
	// При intensity=1.0: цвет умножается на 6 (ярко-белый эффект!)
	// При intensity=0.0: цвет остается неизменным
	scale := 1.0 + intensity*constants.DamageFlashIntensityMultiplier

	op.ColorScale.Scale(scale, scale, scale, 1.0) // R, G, B увеличиваются, A остается
}

// drawUIElements отрисовывает UI элементы над животным (health bar и hunger text)
func (sr *SpriteRenderer) drawUIElements(screen *ebiten.Image, world *core.World, entity core.EntityID, screenX, screenY float32) {
	// Рисуем health bar
	sr.drawHealthBar(screen, world, entity, screenX, screenY)

	// Рисуем hunger text
	sr.drawHungerText(screen, world, entity, screenX, screenY)
}

// drawHealthBar отрисовывает полоску здоровья над животным
func (sr *SpriteRenderer) drawHealthBar(screen *ebiten.Image, world *core.World, entity core.EntityID, screenX, screenY float32) {
	health, hasHealth := world.GetHealth(entity)
	if !hasHealth {
		return
	}

	// Размеры полоски здоровья зависят от типа животного
	var barWidth float32 = 48   // Увеличенная ширина для лучшей видимости
	var barHeight float32 = 6   // Увеличенная высота
	var barOffsetY float32 = 35 // Смещение над спрайтом

	if animalType, hasType := world.GetAnimalType(entity); hasType {
		switch animalType {
		case core.TypeRabbit:
			barWidth = 48
			barOffsetY = 35
		case core.TypeWolf:
			barWidth = 60
			barOffsetY = 40
		}
	}

	barX := screenX - barWidth/2
	barY := screenY - barOffsetY

	// Фон полоски (красный)
	vector.DrawFilledRect(screen, barX, barY, barWidth, barHeight, color.RGBA{200, 50, 50, 255}, false)

	// Здоровье (зелёный)
	var healthPercent float32
	if health.Max > 0 {
		healthPercent = float32(health.Current) / float32(health.Max)
	}
	healthWidth := barWidth * healthPercent
	vector.DrawFilledRect(screen, barX, barY, healthWidth, barHeight, color.RGBA{50, 200, 50, 255}, false)
}

// drawHungerText отрисовывает значение голода над животным
func (sr *SpriteRenderer) drawHungerText(screen *ebiten.Image, world *core.World, entity core.EntityID, screenX, screenY float32) {
	hunger, hasHunger := world.GetHunger(entity)
	if !hasHunger {
		return
	}

	// Создаём текст голода
	hungerText := fmt.Sprintf("%.0f%%", hunger.Value)

	// Позиция текста (над полоской здоровья)
	var textOffsetY float32 = 55 // Увеличенное смещение для лучшей видимости

	if animalType, hasType := world.GetAnimalType(entity); hasType {
		switch animalType {
		case core.TypeRabbit:
			textOffsetY = 55
		case core.TypeWolf:
			textOffsetY = 60
		}
	}

	textX := float64(screenX - 20) // Смещаем влево для центровки
	textY := float64(screenY - textOffsetY)

	// Используем простой дефолтный шрифт (цвет всегда белый для простоты)
	// TODO: В будущем можно добавить цветной текст через кастомный шрифт
	ebitenutil.DebugPrintAt(screen, hungerText, int(textX), int(textY))
}
