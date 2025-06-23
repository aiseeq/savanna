package animation

import (
	"image"
	"math"

	"github.com/aiseeq/savanna/internal/constants"
	"github.com/hajimehoshi/ebiten/v2"
)

// AnimationType тип анимации - используем из constants
type AnimationType = constants.AnimationType

// Константы анимации для совместимости
const (
	AnimIdle       = constants.AnimIdle
	AnimWalk       = constants.AnimWalk
	AnimRun        = constants.AnimRun
	AnimDeathDying = constants.AnimDeathDying
	AnimDeathDecay = constants.AnimDeathDecay
	AnimEat        = constants.AnimEat
	AnimAttack     = constants.AnimAttack
)

// Константы анимационной системы
const (
	DefaultFrameDuration = 0.25 // Длительность кадра по умолчанию (250мс)

	// Стандартные размеры сетки кадров для различного количества анимаций
	Frames2  = 2  // Простая анимация: idle, walk
	Frames4  = 4  // Направленная анимация: 4 направления
	Frames6  = 6  // Расширенная анимация: idle, walk, run, attack, eat, sleep
	Frames8  = 8  // 8-направленная анимация
	Frames9  = 9  // 3x3 сетка
	Frames16 = 16 // 4x4 сетка
	Frames25 = 25 // 5x5 сетка

	// Оптимальные макеты сеток (количество кадров в ряду)
	Layout2Cols = 2 // 2x1 для двухкадровых анимаций
	Layout3Cols = 3 // 3x2 или 3x3 для многокадровых анимаций
	Layout4Cols = 4 // 4x2 или 4x4 для больших анимаций
	Layout5Cols = 5 // 5x5 для максимальных анимаций
)

// String method is now available from constants.AnimationType

// AnimationData описывает параметры анимации
type AnimationData struct {
	Type        AnimationType
	Frames      int           // Количество кадров
	FPS         float32       // Кадров в секунду
	Loop        bool          // Зациклена ли анимация
	SpriteSheet *ebiten.Image // Спрайт-лист (может быть nil если не загружен)
}

// AnimationComponent компонент анимации для ECS
type AnimationComponent struct {
	CurrentAnim AnimationType
	Frame       int     // Текущий кадр (0-based)
	Timer       float32 // Таймер для смены кадров
	Playing     bool    // Проигрывается ли анимация
	FacingRight bool    // Смотрит ли вправо (для отражения спрайта)
}

// AnimationSystem система анимаций
type AnimationSystem struct {
	animations map[AnimationType]*AnimationData
}

// NewAnimationSystem создает новую систему анимаций
func NewAnimationSystem() *AnimationSystem {
	return &AnimationSystem{
		animations: make(map[AnimationType]*AnimationData),
	}
}

// RegisterAnimation регистрирует анимацию в системе
func (as *AnimationSystem) RegisterAnimation(
	animType AnimationType, frames int, fps float32, loop bool, spriteSheet *ebiten.Image,
) {
	as.animations[animType] = &AnimationData{
		Type:        animType,
		Frames:      frames,
		FPS:         fps,
		Loop:        loop,
		SpriteSheet: spriteSheet,
	}
}

// GetAnimation возвращает данные анимации
func (as *AnimationSystem) GetAnimation(animType AnimationType) *AnimationData {
	return as.animations[animType]
}

// GetAllAnimations возвращает все зарегистрированные анимации
func (as *AnimationSystem) GetAllAnimations() map[AnimationType]*AnimationData {
	return as.animations
}

// Update обновляет анимационный компонент
func (as *AnimationSystem) Update(anim *AnimationComponent, deltaTime float32) {
	if !anim.Playing {
		return
	}

	animData := as.GetAnimation(anim.CurrentAnim)
	if animData == nil {
		return
	}

	// Обновляем таймер
	anim.Timer += deltaTime
	frameTime := 1.0 / animData.FPS

	// Проверяем нужно ли сменить кадр
	if anim.Timer >= frameTime {
		anim.Timer -= frameTime
		anim.Frame++

		// Проверяем конец анимации
		if anim.Frame >= animData.Frames {
			if animData.Loop {
				anim.Frame = 0 // Зацикливаем
			} else {
				anim.Frame = animData.Frames - 1 // Останавливаем на последнем кадре
				anim.Playing = false
			}
		}
	}
}

// GetFrameImage возвращает изображение текущего кадра
func (as *AnimationSystem) GetFrameImage(anim *AnimationComponent) *ebiten.Image {
	animData := as.GetAnimation(anim.CurrentAnim)
	if animData == nil || animData.SpriteSheet == nil {
		return nil
	}

	// Определяем размер кадра исходя из размера спрайт-листа и количества кадров
	framesPerRow := as.getFramesPerRow(animData.Frames)
	rows := (animData.Frames + framesPerRow - 1) / framesPerRow

	spriteWidth, spriteHeight := animData.SpriteSheet.Bounds().Dx(), animData.SpriteSheet.Bounds().Dy()
	frameWidth := spriteWidth / framesPerRow
	frameHeight := spriteHeight / rows

	// Ограничиваем кадр допустимыми значениями
	frame := anim.Frame
	if frame >= animData.Frames {
		frame = animData.Frames - 1
	}
	if frame < 0 {
		frame = 0
	}

	row := frame / framesPerRow
	col := frame % framesPerRow

	x := col * frameWidth
	y := row * frameHeight

	// Создаем изображение кадра
	frameImg := animData.SpriteSheet.SubImage(image.Rect(x, y, x+frameWidth, y+frameHeight)).(*ebiten.Image)

	return frameImg
}

// getFramesPerRow возвращает оптимальное количество кадров в ряду для сетки
func (as *AnimationSystem) getFramesPerRow(totalFrames int) int {
	switch totalFrames {
	case Frames2:
		return Layout2Cols // 2x1 (два кадра в ряд)
	case Frames4:
		return Layout2Cols // 2x2
	case Frames6:
		return Layout3Cols // 3x2
	case Frames8:
		return Layout4Cols // 4x2
	case Frames9:
		return Layout3Cols // 3x3
	case Frames16:
		return Layout4Cols // 4x4
	case Frames25:
		return Layout5Cols // 5x5
	default:
		// Для остальных случаев пытаемся найти квадратную или близкую к квадратной сетку
		sqrt := int(math.Sqrt(float64(totalFrames)))
		if sqrt*sqrt == totalFrames {
			return sqrt
		}
		return (totalFrames + sqrt - 1) / sqrt // Округляем вверх
	}
}
