package main

import (
	"bytes"
	_ "embed"
	"flag"
	"fmt"
	"image/color"
	_ "image/png"
	"log"
	"math"
	"os"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	"github.com/hajimehoshi/ebiten/v2/text/v2"

	"github.com/aiseeq/savanna/internal/animation"
)

// Встроенный шрифт DejaVu Sans Mono
//
//go:embed DejaVuSansMono.ttf
var dejaVuSansMonoTTF []byte

var dejaVuSansMonoFace *text.GoTextFace

// AnimationViewer приложение для просмотра анимаций
type AnimationViewer struct {
	animSystem    *animation.AnimationSystem
	currentAnim   animation.AnimationType
	animComponent *animation.AnimationComponent

	// Управление просмотром
	scale  float64 // Масштаб отображения
	speed  float32 // Множитель скорости анимации
	paused bool    // Пауза

	// Доступные анимации для демонстрации
	availableAnimations []animation.AnimationType
	currentIndex        int

	// Тестовые анимации (заглушки)
	testAnimations map[animation.AnimationType]*ebiten.Image
}

// NewAnimationViewer создает новый просмотрщик анимаций
func NewAnimationViewer(animalType string) *AnimationViewer {
	av := &AnimationViewer{
		animSystem:     animation.NewAnimationSystem(),
		scale:          1.0, // Начальный масштаб 1x
		speed:          1.0, // Нормальная скорость
		paused:         false,
		testAnimations: make(map[animation.AnimationType]*ebiten.Image),
	}

	// Список доступных анимаций для демонстрации
	av.availableAnimations = []animation.AnimationType{
		animation.AnimIdle,
		animation.AnimWalk,
		animation.AnimRun,
		animation.AnimAttack,
		animation.AnimEat,
		animation.AnimDeathDying,
	}

	// Создаем тестовые анимации (заглушки)
	av.createTestAnimations(animalType)

	// Начинаем с первой анимации
	av.setCurrentAnimation(0)

	return av
}

// createTestAnimations создает тестовые анимации-заглушки
func (av *AnimationViewer) createTestAnimations(animalType string) {
	// Создаем все спрайты для волка
	var idleSprite, walkSprite, runSprite, attackSprite, eatSprite, deathSprite *ebiten.Image

	if animalType == "wolf" {
		sprites := av.loadAnimalAnimations("wolf")
		idleSprite, walkSprite, runSprite, attackSprite, eatSprite, deathSprite =
			sprites.Idle, sprites.Walk, sprites.Run, sprites.Attack, sprites.Eat, sprites.Death
	} else if animalType == "hare" || animalType == "rabbit" {
		sprites := av.loadAnimalAnimations("rabbit")
		idleSprite, walkSprite, runSprite, attackSprite, eatSprite, deathSprite =
			sprites.Idle, sprites.Walk, sprites.Run, sprites.Attack, sprites.Eat, sprites.Death
	} else {
		// Тестовые спрайты для других животных
		idleSprite = av.createTestSpriteSheet(color.RGBA{100, 100, 255, 255})
		walkSprite = av.createTestSpriteSheet(color.RGBA{100, 255, 100, 255})
		runSprite = av.createTestSpriteSheet(color.RGBA{255, 255, 100, 255})
		attackSprite = av.createTestSpriteSheet(color.RGBA{255, 100, 100, 255})
		eatSprite = av.createTestSpriteSheet(color.RGBA{255, 150, 100, 255})
		deathSprite = av.createTestSpriteSheet(color.RGBA{200, 50, 50, 255})
	}

	// Регистрируем все анимации (все цикличные по 2 кадра)
	av.animSystem.RegisterAnimation(animation.AnimIdle, 2, 2.0, true, idleSprite)
	av.animSystem.RegisterAnimation(animation.AnimWalk, 2, 4.0, true, walkSprite)
	av.animSystem.RegisterAnimation(animation.AnimRun, 2, 12.0, true, runSprite)
	av.animSystem.RegisterAnimation(animation.AnimAttack, 2, 5.0, true, attackSprite)
	av.animSystem.RegisterAnimation(animation.AnimEat, 2, 4.0, true, eatSprite)
	av.animSystem.RegisterAnimation(animation.AnimDeathDying, 2, 3.0, true, deathSprite)
}

// loadAnimationFrames загружает 2 кадра анимации и объединяет их в спрайт-лист
func (av *AnimationViewer) loadAnimationFrames(animationName string) (*ebiten.Image, error) {
	frameFiles := []string{
		fmt.Sprintf("assets/animations/%s_1.png", animationName),
		fmt.Sprintf("assets/animations/%s_2.png", animationName),
	}

	// Загружаем оба кадра
	frames := make([]*ebiten.Image, 2)
	var frameWidth, frameHeight int

	for i, filepath := range frameFiles {
		if _, err := os.Stat(filepath); os.IsNotExist(err) {
			return nil, fmt.Errorf("файл не найден: %s", filepath)
		}

		img, _, err := ebitenutil.NewImageFromFile(filepath)
		if err != nil {
			return nil, fmt.Errorf("ошибка загрузки %s: %v", filepath, err)
		}

		frames[i] = img
		if i == 0 {
			frameWidth, frameHeight = img.Bounds().Dx(), img.Bounds().Dy()
		}
	}

	// Создаем спрайт-лист 2x1 (два кадра в ряд)
	spriteSheet := ebiten.NewImage(frameWidth*2, frameHeight)

	// Размещаем кадры
	for i, frame := range frames {
		op := &ebiten.DrawImageOptions{}
		op.GeoM.Translate(float64(i*frameWidth), 0)
		spriteSheet.DrawImage(frame, op)
	}

	return spriteSheet, nil
}

// createTestSpriteSheet создает тестовый спрайт-лист заданного цвета (всегда 2 кадра)
func (av *AnimationViewer) createTestSpriteSheet(baseColor color.RGBA) *ebiten.Image {
	frames := 2 // Всегда используем 2 кадра для тестовых спрайтов
	framesPerRow := av.getFramesPerRow(frames)
	rows := (frames + framesPerRow - 1) / framesPerRow

	width := framesPerRow * 32
	height := rows * 32

	img := ebiten.NewImage(width, height)

	// Заполняем каждый кадр разными оттенками базового цвета
	for i := 0; i < frames; i++ {
		row := i / framesPerRow
		col := i % framesPerRow

		x := col * 32
		y := row * 32

		// Изменяем яркость для разных кадров
		brightness := 0.5 + 0.5*float64(i)/float64(frames)
		frameColor := color.RGBA{
			uint8(float64(baseColor.R) * brightness),
			uint8(float64(baseColor.G) * brightness),
			uint8(float64(baseColor.B) * brightness),
			baseColor.A,
		}

		// Рисуем прямоугольник для кадра
		frameImg := ebiten.NewImage(32, 32)
		frameImg.Fill(frameColor)

		// Рисуем простую границу кадра
		borderColor := color.RGBA{255, 255, 255, 150}

		// Верхняя граница
		topBorder := ebiten.NewImage(32, 1)
		topBorder.Fill(borderColor)
		frameImg.DrawImage(topBorder, nil)

		// Нижняя граница
		bottomBorder := ebiten.NewImage(32, 1)
		bottomBorder.Fill(borderColor)
		op3 := &ebiten.DrawImageOptions{}
		op3.GeoM.Translate(0, 31)
		frameImg.DrawImage(bottomBorder, op3)

		// Левая граница
		leftBorder := ebiten.NewImage(1, 32)
		leftBorder.Fill(borderColor)
		frameImg.DrawImage(leftBorder, nil)

		// Правая граница
		rightBorder := ebiten.NewImage(1, 32)
		rightBorder.Fill(borderColor)
		op4 := &ebiten.DrawImageOptions{}
		op4.GeoM.Translate(31, 0)
		frameImg.DrawImage(rightBorder, op4)

		// Рисуем номер кадра
		op := &text.DrawOptions{}
		op.GeoM.Translate(8, 8)
		op.ColorScale.ScaleWithColor(color.RGBA{255, 255, 255, 255})
		text.Draw(frameImg, fmt.Sprintf("%d", i), dejaVuSansMonoFace, op)

		// Добавляем кадр в спрайт-лист
		op2 := &ebiten.DrawImageOptions{}
		op2.GeoM.Translate(float64(x), float64(y))
		img.DrawImage(frameImg, op2)
	}

	return img
}

// AnimalSprites содержит все спрайты для животного (устранение нарушения tooManyResultsChecker)
type AnimalSprites struct {
	Idle, Walk, Run, Attack, Eat, Sleep, Death *ebiten.Image
}

// loadAnimalAnimations загружает все анимации для указанного животного (устраняет дублирование кода)
func (av *AnimationViewer) loadAnimalAnimations(animalType string) AnimalSprites {
	fmt.Printf("Загружаем все анимации %s...\n", animalType)

	// Определяем префикс для файлов анимаций
	prefix := animalType
	if animalType == "hare" || animalType == "rabbit" {
		prefix = "hare"
	}

	// Загружаем каждую анимацию с fallback на тестовые спрайты
	animTypes := []struct {
		name, displayName string
		color             color.RGBA
	}{
		{"idle", "покоя", color.RGBA{100, 100, 255, 255}},
		{"walk", "ходьбы", color.RGBA{100, 255, 100, 255}},
		{"run", "бега", color.RGBA{255, 255, 100, 255}},
		{"attack", "атаки", color.RGBA{255, 100, 100, 255}},
		{"eat", "еды", color.RGBA{255, 150, 100, 255}},
		{"sleep", "сна", color.RGBA{100, 150, 255, 255}},
		{"dead", "смерти", color.RGBA{200, 50, 50, 255}},
	}

	sprites := make([]*ebiten.Image, len(animTypes))

	for i, animType := range animTypes {
		fileName := fmt.Sprintf("%s_%s", prefix, animType.name)
		if sprite, err := av.loadAnimationFrames(fileName); err == nil {
			sprites[i] = sprite
			fmt.Printf("✓ Загружена анимация %s\n", animType.displayName)
		} else {
			fmt.Printf("Не удалось загрузить %s: %v\n", animType.name, err)
			sprites[i] = av.createTestSpriteSheet(animType.color)
		}
	}

	fmt.Printf("Все анимации %s загружены!\n", animalType)
	return AnimalSprites{
		Idle:   sprites[0],
		Walk:   sprites[1],
		Run:    sprites[2],
		Attack: sprites[3],
		Eat:    sprites[4],
		Sleep:  sprites[5],
		Death:  sprites[6],
	}
}

// getFramesPerRow возвращает количество кадров в ряду
func (av *AnimationViewer) getFramesPerRow(totalFrames int) int {
	switch totalFrames {
	case 4:
		return 2 // 2x2
	case 6:
		return 3 // 3x2
	case 8:
		return 4 // 4x2
	case 9:
		return 3 // 3x3
	case 16:
		return 4 // 4x4
	default:
		sqrt := int(math.Sqrt(float64(totalFrames)))
		if sqrt*sqrt == totalFrames {
			return sqrt
		}
		return (totalFrames + sqrt - 1) / sqrt
	}
}

// setCurrentAnimation устанавливает текущую анимацию по индексу
func (av *AnimationViewer) setCurrentAnimation(index int) {
	if index < 0 || index >= len(av.availableAnimations) {
		return
	}

	av.currentIndex = index
	av.currentAnim = av.availableAnimations[index]

	// Создаем новый компонент анимации
	av.animComponent = &animation.AnimationComponent{
		CurrentAnim: av.currentAnim,
		Frame:       0,
		Timer:       0,
		Playing:     true,
		FacingRight: true,
	}
}

// Update обновляет состояние просмотрщика
func (av *AnimationViewer) Update() error {
	// Обработка клавиш
	av.handleInput()

	// Обновляем анимацию если не на паузе
	if !av.paused && av.animComponent != nil {
		deltaTime := (1.0 / 60.0) * av.speed
		av.animSystem.Update(av.animComponent, deltaTime)
	}

	return nil
}

// handleInput обрабатывает пользовательский ввод (рефакторинг: снижена когнитивная сложность)
func (av *AnimationViewer) handleInput() {
	av.handleAnimationControls()
	av.handleSpeedControls()
	av.handlePlaybackControls()
	av.handleScaleControls()
	av.handleExitControls()
}

// handleAnimationControls обрабатывает переключение анимаций (helper-функция)
func (av *AnimationViewer) handleAnimationControls() {
	if inpututil.IsKeyJustPressed(ebiten.KeyArrowLeft) {
		newIndex := av.currentIndex - 1
		if newIndex < 0 {
			newIndex = len(av.availableAnimations) - 1
		}
		av.setCurrentAnimation(newIndex)
	}

	if inpututil.IsKeyJustPressed(ebiten.KeyArrowRight) {
		newIndex := (av.currentIndex + 1) % len(av.availableAnimations)
		av.setCurrentAnimation(newIndex)
	}
}

// handleSpeedControls обрабатывает управление скоростью (helper-функция)
func (av *AnimationViewer) handleSpeedControls() {
	if inpututil.IsKeyJustPressed(ebiten.KeyEqual) || inpututil.IsKeyJustPressed(ebiten.KeyKPAdd) {
		av.speed *= 1.25
		if av.speed > 8.0 {
			av.speed = 8.0
		}
	}

	if inpututil.IsKeyJustPressed(ebiten.KeyMinus) || inpututil.IsKeyJustPressed(ebiten.KeyKPSubtract) {
		av.speed /= 1.25
		if av.speed < 0.125 {
			av.speed = 0.125
		}
	}
}

// handlePlaybackControls обрабатывает управление воспроизведением (helper-функция)
func (av *AnimationViewer) handlePlaybackControls() {
	if inpututil.IsKeyJustPressed(ebiten.KeySpace) {
		av.paused = !av.paused
	}

	if inpututil.IsKeyJustPressed(ebiten.KeyR) {
		if av.animComponent != nil {
			av.animComponent.Frame = 0
			av.animComponent.Timer = 0
			av.animComponent.Playing = true
		}
	}
}

// handleScaleControls обрабатывает масштабирование (helper-функция)
func (av *AnimationViewer) handleScaleControls() {
	_, scrollY := ebiten.Wheel()
	if scrollY != 0 {
		if scrollY > 0 {
			av.scale *= 1.2
			if av.scale > 16.0 {
				av.scale = 16.0
			}
		} else {
			av.scale /= 1.2
			if av.scale < 0.5 {
				av.scale = 0.5
			}
		}
	}
}

// handleExitControls обрабатывает выход из приложения (helper-функция)
func (av *AnimationViewer) handleExitControls() {
	if inpututil.IsKeyJustPressed(ebiten.KeyEscape) {
		return
	}
}

// Draw отрисовывает просмотрщик анимаций
func (av *AnimationViewer) Draw(screen *ebiten.Image) {
	screen.Fill(color.RGBA{40, 40, 40, 255})

	if av.animComponent == nil {
		return
	}

	// Получаем текущий кадр анимации
	frameImg := av.animSystem.GetFrameImage(av.animComponent)
	if frameImg != nil {
		// Рисуем анимацию в центре экрана
		screenWidth, screenHeight := screen.Bounds().Dx(), screen.Bounds().Dy()

		// Получаем реальный размер кадра
		frameWidth, frameHeight := frameImg.Bounds().Dx(), frameImg.Bounds().Dy()

		op := &ebiten.DrawImageOptions{}
		op.GeoM.Scale(av.scale, av.scale)
		op.GeoM.Translate(
			float64(screenWidth)/2-float64(frameWidth)*av.scale/2,
			float64(screenHeight)/2-float64(frameHeight)*av.scale/2,
		)
		screen.DrawImage(frameImg, op)
	}

	// Рисуем UI
	av.drawUI(screen)
}

// drawUI отрисовывает пользовательский интерфейс
func (av *AnimationViewer) drawUI(screen *ebiten.Image) {
	if av.animComponent == nil {
		return
	}

	animData := av.animSystem.GetAnimation(av.currentAnim)
	if animData == nil {
		return
	}

	// Информация об анимации
	lines := []string{
		fmt.Sprintf("Анимация: %s (%d/%d)", av.currentAnim.String(), av.currentIndex+1, len(av.availableAnimations)),
		fmt.Sprintf("Кадр: %d/%d", av.animComponent.Frame+1, animData.Frames),
		fmt.Sprintf("FPS: %.1f", animData.FPS),
		fmt.Sprintf("Зацикленная: %v", animData.Loop),
		fmt.Sprintf("Скорость: x%.2f", av.speed),
		fmt.Sprintf("Масштаб: x%.1f", av.scale),
		"",
		"Управление:",
		"← → : переключение анимаций",
		"+ - : скорость анимации",
		"Колесо: масштаб",
		"Пробел: пауза",
		"R: перезапуск",
		"ESC: выход",
	}

	if av.paused {
		lines = append([]string{"ПАУЗА", ""}, lines...)
	}

	// Отрисовываем текст
	for i, line := range lines {
		op := &text.DrawOptions{}
		op.GeoM.Translate(10, float64(10+i*16))
		op.ColorScale.ScaleWithColor(color.RGBA{255, 255, 255, 255})
		text.Draw(screen, line, dejaVuSansMonoFace, op)
	}
}

// Layout устанавливает размеры экрана
func (av *AnimationViewer) Layout(outsideWidth, outsideHeight int) (screenWidth, screenHeight int) {
	return 800, 600
}

func main() {
	// Парсим аргументы
	animalType := flag.String("show", "wolf", "Тип животного для показа анимаций (wolf, rabbit)")
	flag.Parse()

	// Инициализируем шрифт
	fontSource, err := text.NewGoTextFaceSource(bytes.NewReader(dejaVuSansMonoTTF))
	if err != nil {
		log.Fatal(err)
	}
	dejaVuSansMonoFace = &text.GoTextFace{
		Source: fontSource,
		Size:   12,
	}

	fmt.Printf("Запуск просмотрщика анимаций для: %s\n", *animalType)

	// Создаем просмотрщик
	viewer := NewAnimationViewer(*animalType)

	// Настройки окна
	ebiten.SetWindowSize(800, 600)
	ebiten.SetWindowTitle(fmt.Sprintf("Animation Viewer - %s", *animalType))
	ebiten.SetWindowResizingMode(ebiten.WindowResizingModeEnabled)

	// Запускаем приложение
	if err := ebiten.RunGame(viewer); err != nil {
		log.Fatal(err)
	}
}
