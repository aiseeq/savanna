package main

import (
	"flag"
	"fmt"
	"image"
	"image/color"
	"image/png"
	"log"
	"math"
	"net/http"
	_ "net/http/pprof"
	"os"
	"time"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	"github.com/hajimehoshi/ebiten/v2/text/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"

	"github.com/aiseeq/savanna/config"
	"github.com/aiseeq/savanna/internal/core"
	"github.com/aiseeq/savanna/internal/generator"
	"github.com/aiseeq/savanna/internal/rendering"
	"github.com/aiseeq/savanna/internal/simulation"
)

// Game структура для GUI версии симулятора экосистемы саванны
// Рефакторинг: разбита на специализированные менеджеры (соблюдение SRP)
type Game struct {
	// Менеджеры с единственными ответственностями
	gameWorld      *GameWorld      // Управление симуляцией мира
	timeManager    *TimeManager    // Управление временем
	spriteRenderer *SpriteRenderer // Отрисовка спрайтов животных
	fontManager    *FontManager    // Управление шрифтами

	// Изометрическая система отрисовки
	isometricRenderer *rendering.IsometricRenderer // Изометрическая отрисовка
	camera            *rendering.Camera            // Камера для изометрии
	terrain           *generator.Terrain           // Ландшафт

	// Дебаг режим
	debugMode bool // Включен ли дебаг режим (F3)
}

// Update обновляет логику игры (рефакторинг: использует менеджеры)
func (g *Game) Update() error {
	// Проверяем выход
	if inpututil.IsKeyJustPressed(ebiten.KeyEscape) {
		return fmt.Errorf("игра завершена пользователем")
	}

	// Обновляем менеджеры (каждый отвечает за свою область)
	g.timeManager.Update() // Управление временем

	// Обновляем новую камеру
	cameraUpdateDeltaTime := g.timeManager.GetDeltaTime()
	g.camera.Update(cameraUpdateDeltaTime)

	// Убрано автоматическое рецентрирование - камера должна быть статичной

	// Переключение дебаг режима (F3)
	if inpututil.IsKeyJustPressed(ebiten.KeyF3) {
		g.debugMode = !g.debugMode
	}

	// Скриншот с дебаг-режимом (F2)
	if inpututil.IsKeyJustPressed(ebiten.KeyF2) {
		g.takeDebugScreenshot()
	}

	// Обновляем симуляцию с учётом времени
	deltaTime := g.timeManager.GetDeltaTime()
	g.gameWorld.Update(deltaTime)

	return nil
}

// Draw отрисовывает кадр
func (g *Game) Draw(screen *ebiten.Image) {
	// Очищаем экран тёмным цветом саванны
	screen.Fill(color.RGBA{101, 67, 33, 255}) // Коричневый цвет земли

	// Используем новую изометрическую систему отрисовки
	world := g.gameWorld.GetWorld()
	g.isometricRenderer.RenderWorld(screen, g.terrain, world, g.camera)

	// Дебаг отрисовка
	if g.debugMode {
		g.drawDebugInfo(screen, world)
	}

	// Отрисовываем пользовательский интерфейс
	g.drawUI(screen)

	// FPS счетчик (этап 7)
	g.drawFPS(screen)
}

// Layout устанавливает размеры экрана
func (g *Game) Layout(outsideWidth, outsideHeight int) (screenWidth, screenHeight int) {
	return outsideWidth, outsideHeight
}

// REMOVED: Старые методы отрисовки terrain и animals
// Новая изометрическая система отрисовки используется через isometricRenderer

// drawUI отрисовывает пользовательский интерфейс
func (g *Game) drawUI(screen *ebiten.Image) {
	stats := g.gameWorld.GetStats()

	// Получаем шрифт для отрисовки
	font := g.fontManager.GetDebugFont()

	// Создаём текстовую информацию
	y := float64(10)
	lineHeight := float64(20)

	// ТИПОБЕЗОПАСНОСТЬ: Статистика теперь типизирована
	g.drawText(screen, fmt.Sprintf("Rabbits: %d", stats.Rabbits), 10, y, font)
	y += lineHeight
	g.drawText(screen, fmt.Sprintf("Wolves: %d", stats.Wolves), 10, y, font)
	y += lineHeight

	// Масштаб и скорость
	g.drawText(screen, fmt.Sprintf("Zoom: %.1fx", g.camera.GetZoom()), 10, y, font)
	y += lineHeight

	timeScale := g.timeManager.GetTimeScale()
	isPaused := g.timeManager.IsPaused()
	if isPaused {
		g.drawText(screen, "Speed: PAUSED", 10, y, font)
	} else {
		g.drawText(screen, fmt.Sprintf("Speed: %.1fx", timeScale), 10, y, font)
	}
	y += lineHeight

	// Голод первого зайца для отладки
	world := g.gameWorld.GetWorld()
	var firstRabbit core.EntityID
	found := false
	world.ForEachWith(core.MaskAnimalType, func(entity core.EntityID) {
		if !found {
			if animalType, ok := world.GetAnimalType(entity); ok && animalType == core.TypeRabbit {
				firstRabbit = entity
				found = true
			}
		}
	})

	if found {
		if hunger, ok := world.GetHunger(firstRabbit); ok {
			g.drawText(screen, fmt.Sprintf("Hunger: %.1f%%", hunger.Value), 10, y, font)
		}
	}
}

// REMOVED: legacy UI код был удалён и заменён на единую функцию drawUI

// drawText рендерит текст с использованием пользовательского или дефолтного шрифта
//
//nolint:unparam // x всегда 10 для UI элементов, но оставляем для гибкости
func (g *Game) drawText(screen *ebiten.Image, textStr string, x, y float64, font *text.GoTextFace) {
	if font != nil {
		// Используем пользовательский шрифт
		op := &text.DrawOptions{}
		op.GeoM.Translate(x, y)
		op.ColorScale.ScaleWithColor(color.White)
		text.Draw(screen, textStr, font, op)
	} else {
		// Фолбэк на дефолтный шрифт
		ebitenutil.DebugPrintAt(screen, textStr, int(x), int(y))
	}
}

// Helper-методы

// getAnimalRadius получает радиус животного из компонента Size (устраняет DRY нарушение)
// Ранее размеры дублировались между game_balance.go и GUI кодом
func (g *Game) getAnimalRadius(entity core.EntityID, world *core.World) float32 {
	if size, ok := world.GetSize(entity); ok {
		return size.Radius
	}
	return simulation.DefaultAnimalRadius // Фолбэк из централизованных констант
}

// HealthBarParams параметры отрисовки полоски здоровья
type HealthBarParams struct {
	ScreenX, ScreenY, Radius float32
}

func (g *Game) drawHealthBar(
	screen *ebiten.Image,
	entity core.EntityID,
	world *core.World,
	params HealthBarParams,
) {
	health, hasHealth := world.GetHealth(entity)
	if !hasHealth {
		return
	}

	// ИСПРАВЛЕНИЕ: Размеры полоски здоровья зависят от размера СПРАЙТА, не от физического радиуса
	var barWidth float32 = 32 // Стандартная ширина для зайца
	var barHeight float32 = 4
	var barOffsetY float32 = 25 // Смещение над спрайтом

	// Настройка под тип животного
	if animalType, hasType := world.GetAnimalType(entity); hasType {
		switch animalType {
		case core.TypeRabbit:
			barWidth = 32
			barOffsetY = 25
		case core.TypeWolf:
			barWidth = 40
			barOffsetY = 30
		}
	}

	barX := params.ScreenX - barWidth/2
	barY := params.ScreenY - barOffsetY

	// Фон полоски (красный)
	vector.DrawFilledRect(screen, barX, barY, barWidth, barHeight, color.RGBA{200, 50, 50, 255}, false)

	// БЕЗОПАСНОСТЬ: Здоровье (зелёный) с защитой от деления на ноль
	var healthPercent float32
	if health.Max > 0 {
		healthPercent = float32(health.Current) / float32(health.Max)
	}
	healthWidth := barWidth * healthPercent
	vector.DrawFilledRect(screen, barX, barY, healthWidth, barHeight, color.RGBA{50, 200, 50, 255}, false)
}

// HungerTextParams параметры отрисовки текста голода
type HungerTextParams struct {
	ScreenX, ScreenY, Radius float32
}

// drawHungerText отрисовывает значение голода над животным
func (g *Game) drawHungerText(
	screen *ebiten.Image,
	entity core.EntityID,
	world *core.World,
	params HungerTextParams,
) {
	hunger, hasHunger := world.GetHunger(entity)
	if !hasHunger {
		return
	}

	// Создаём текст голода
	hungerText := fmt.Sprintf("%.0f%%", hunger.Value)

	// ИСПРАВЛЕНИЕ: Позиция текста зависит от размера СПРАЙТА, не от физического радиуса
	var textOffsetY float32 = 40 // Стандартное смещение над спрайтом для зайца

	// Настройка под тип животного
	if animalType, hasType := world.GetAnimalType(entity); hasType {
		switch animalType {
		case core.TypeRabbit:
			textOffsetY = 40
		case core.TypeWolf:
			textOffsetY = 45
		}
	}

	// Позиция текста (над полоской здоровья)
	textX := float64(params.ScreenX)
	textY := float64(params.ScreenY - textOffsetY) // Над полоской здоровья

	// Определяем цвет в зависимости от уровня голода
	var textColor color.Color
	if hunger.Value < 30.0 {
		// Критический голод - красный
		textColor = color.RGBA{255, 50, 50, 255}
	} else if hunger.Value < 60.0 {
		// Средний голод - жёлтый
		textColor = color.RGBA{255, 255, 50, 255}
	} else {
		// Сытость - зелёный
		textColor = color.RGBA{50, 255, 50, 255}
	}

	// Получаем шрифт
	font := g.fontManager.GetDebugFont()

	if font != nil {
		// Используем пользовательский шрифт
		op := &text.DrawOptions{}
		op.GeoM.Translate(textX-20, textY) // Смещаем влево для центровки
		op.ColorScale.ScaleWithColor(textColor)
		text.Draw(screen, hungerText, font, op)
	} else {
		// Фолбэк на дефолтный шрифт (но с ограниченными возможностями цвета)
		// К сожалению, ebitenutil.DebugPrintAt не поддерживает цвета
		ebitenutil.DebugPrintAt(screen, hungerText, int(textX-20), int(textY))
	}
}

// drawDebugInfo отрисовывает дебаг информацию (F3)
func (g *Game) drawDebugInfo(screen *ebiten.Image, world *core.World) {
	// Отрисовываем границы тайлов
	g.drawTileGrid(screen)

	// Отрисовываем ID животных и их состояния
	g.drawAnimalDebugInfo(screen, world)

	// Отрисовываем камеру информацию
	g.drawCameraInfo(screen)
}

// drawTileGrid отрисовывает сетку тайлов
func (g *Game) drawTileGrid(screen *ebiten.Image) {
	gridColor := color.RGBA{R: 100, G: 100, B: 100, A: 128} // Полупрозрачная сетка

	// ОПТИМИЗАЦИЯ: Переиспользуемый буфер для точек ромба (избегаем аллокаций)
	var points [8]float32 // 4 точки × 2 координаты

	// Определяем видимую область (оптимизация производительности)
	bounds := screen.Bounds()
	screenW, screenH := float32(bounds.Dx()), float32(bounds.Dy())

	// Углы экрана в мировых координатах
	topLeftX, topLeftY := g.camera.ScreenToWorld(0, 0)
	topRightX, topRightY := g.camera.ScreenToWorld(screenW, 0)
	bottomLeftX, bottomLeftY := g.camera.ScreenToWorld(0, screenH)
	bottomRightX, bottomRightY := g.camera.ScreenToWorld(screenW, screenH)

	// Находим границы видимой области
	minX := int(math.Floor(float64(min(min(topLeftX, topRightX), min(bottomLeftX, bottomRightX)))))
	minY := int(math.Floor(float64(min(min(topLeftY, topRightY), min(bottomLeftY, bottomRightY)))))
	maxX := int(math.Ceil(float64(max(max(topLeftX, topRightX), max(bottomLeftX, bottomRightX)))))
	maxY := int(math.Ceil(float64(max(max(topLeftY, topRightY), max(bottomLeftY, bottomRightY)))))

	// Ограничиваем видимую область размерами terrain
	if minX < 0 {
		minX = 0
	}
	if minY < 0 {
		minY = 0
	}
	if maxX >= g.terrain.Width {
		maxX = g.terrain.Width - 1
	}
	if maxY >= g.terrain.Height {
		maxY = g.terrain.Height - 1
	}

	// Отрисовываем только видимые тайлы (frustum culling)
	for y := minY; y <= maxY; y++ {
		for x := minX; x <= maxX; x++ {
			// Преобразуем в экранные координаты с учётом камеры
			screenX, screenY := g.camera.WorldToScreen(float32(x), float32(y))

			// Рисуем границы тайла
			tileW := float32(rendering.TileWidth)  // Используем константу из пакета rendering
			tileH := float32(rendering.TileHeight) // Используем константу из пакета rendering

			// ОПТИМИЗАЦИЯ: Переиспользуем буфер вместо создания нового slice
			points[0], points[1] = screenX, screenY-tileH/2 // Верх
			points[2], points[3] = screenX+tileW/2, screenY // Право
			points[4], points[5] = screenX, screenY+tileH/2 // Низ
			points[6], points[7] = screenX-tileW/2, screenY // Лево

			// Рисуем линии ромба
			vector.StrokeLine(screen, points[0], points[1], points[2], points[3], 1, gridColor, false)
			vector.StrokeLine(screen, points[2], points[3], points[4], points[5], 1, gridColor, false)
			vector.StrokeLine(screen, points[4], points[5], points[6], points[7], 1, gridColor, false)
			vector.StrokeLine(screen, points[6], points[7], points[0], points[1], 1, gridColor, false)
		}
	}
}

// drawAnimalDebugInfo отрисовывает дебаг информацию о животных
func (g *Game) drawAnimalDebugInfo(screen *ebiten.Image, world *core.World) {
	font := g.fontManager.GetDebugFont()

	world.ForEachWith(core.MaskPosition|core.MaskAnimalType, func(entity core.EntityID) {
		pos, hasPos := world.GetPosition(entity)
		if !hasPos {
			return
		}

		// Преобразуем в экранные координаты с учётом камеры
		screenX, screenY := g.camera.WorldToScreen(pos.X, pos.Y)

		// Проверяем видимость
		bounds := screen.Bounds()
		if screenX < -50 || screenY < -50 || screenX > float32(bounds.Dx())+50 || screenY > float32(bounds.Dy())+50 {
			return
		}

		// Получаем размер и тип животного
		radius := float32(8)               // Значение по умолчанию
		var visionMultiplier float32 = 5.0 // По умолчанию

		if size, hasSize := world.GetSize(entity); hasSize {
			radius = size.Radius
		}

		// Определяем правильный множитель зрения по типу животного
		if animalType, hasType := world.GetAnimalType(entity); hasType {
			switch animalType {
			case core.TypeRabbit:
				visionMultiplier = 6.0 // RabbitVisionMultiplier из game_balance.go (обновлено)
			case core.TypeWolf:
				visionMultiplier = 6.7 // WolfVisionMultiplier из game_balance.go (обновлено)
			default:
				visionMultiplier = 8.0 // DefaultVisionMultiplier (обновлено)
			}
		}

		// Рисуем физический размер (синий круг)
		physicalColor := color.RGBA{R: 0, G: 150, B: 255, A: 128} // Синий полупрозрачный
		vector.StrokeCircle(screen, screenX, screenY, radius, 1, physicalColor, false)

		// Рисуем радиус обзора (жёлтый круг)
		visionRadius := radius * visionMultiplier
		visionColor := color.RGBA{R: 255, G: 255, B: 0, A: 64} // Желтый полупрозрачный
		vector.StrokeCircle(screen, screenX, screenY, visionRadius, 2, visionColor, false)

		// Отрисовываем ID животного
		idText := fmt.Sprintf("ID:%d", entity)
		textY := float64(screenY - radius - 35)

		if font != nil {
			op := &text.DrawOptions{}
			op.GeoM.Translate(float64(screenX-20), textY)
			op.ColorScale.ScaleWithColor(color.White)
			text.Draw(screen, idText, font, op)
		} else {
			ebitenutil.DebugPrintAt(screen, idText, int(screenX-20), int(textY))
		}

		// ДОБАВЛЕНО: Отрисовываем хелсбар
		g.drawHealthBar(screen, entity, world, HealthBarParams{
			ScreenX: screenX,
			ScreenY: screenY,
			Radius:  radius,
		})

		// ДОБАВЛЕНО: Отрисовываем текст голода
		g.drawHungerText(screen, entity, world, HungerTextParams{
			ScreenX: screenX,
			ScreenY: screenY,
			Radius:  radius,
		})
	})
}

// drawCameraInfo отрисовывает информацию о камере
func (g *Game) drawCameraInfo(screen *ebiten.Image) {
	font := g.fontManager.GetDebugFont()

	infoText := fmt.Sprintf("Camera: X=%.1f Y=%.1f Zoom=%.1fx",
		g.camera.X, g.camera.Y, g.camera.GetZoom())

	if font != nil {
		op := &text.DrawOptions{}
		op.GeoM.Translate(10, 150) // Под основным UI
		op.ColorScale.ScaleWithColor(color.RGBA{R: 255, G: 255, B: 0, A: 255})
		text.Draw(screen, infoText, font, op)
	} else {
		ebitenutil.DebugPrintAt(screen, infoText, 10, 150)
	}
}

// drawFPS отрисовывает FPS счетчик
func (g *Game) drawFPS(screen *ebiten.Image) {
	font := g.fontManager.GetDebugFont()

	// Получаем TPS и рассчитываем FPS
	tps := ebiten.ActualTPS()
	fps := ebiten.ActualFPS()

	fpsText := fmt.Sprintf("FPS: %.1f / TPS: %.1f", fps, tps)

	// Отображаем в правом верхнем углу
	bounds := screen.Bounds()
	x := float64(bounds.Dx() - 150)
	y := float64(20)

	if font != nil {
		op := &text.DrawOptions{}
		op.GeoM.Translate(x, y)
		op.ColorScale.ScaleWithColor(color.White)
		text.Draw(screen, fpsText, font, op)
	} else {
		ebitenutil.DebugPrintAt(screen, fpsText, int(x), int(y))
	}
}

func main() {
	// ПРОФИЛИРОВАНИЕ: Запускаем pprof сервер для анализа производительности
	go func() {
		log.Println("Запуск pprof сервера на http://localhost:6060")
		log.Println("Для профиля CPU: go tool pprof http://localhost:6060/debug/pprof/profile")
		log.Println("Для профиля памяти: go tool pprof http://localhost:6060/debug/pprof/heap")
		if err := http.ListenAndServe("localhost:6060", nil); err != nil {
			log.Printf("Ошибка pprof сервера: %v", err)
		}
	}()

	// Парсим аргументы командной строки
	var seedFlag = flag.Int64(
		"seed", 0,
		"Seed для детерминированной симуляции (если не указан, используется текущее время)",
	)
	var pprofFlag = flag.Bool(
		"pprof", false,
		"Включить профилирование производительности на порту 6060",
	)
	flag.Parse()

	if *pprofFlag {
		log.Println("Профилирование включено. Доступно на http://localhost:6060/debug/pprof/")
	}

	// Устанавливаем seed
	var seed int64
	if *seedFlag != 0 {
		seed = *seedFlag
		fmt.Printf("Используется заданный seed: %d\n", seed)
	} else {
		seed = time.Now().UnixNano()
		fmt.Printf("Используется случайный seed: %d\n", seed)
	}

	fmt.Println("Запуск GUI версии симулятора экосистемы саванны...")

	// Создаём конфигурацию и ландшафт
	cfg := config.LoadDefaultConfig()
	cfg.World.Seed = seed
	terrainGen := generator.NewTerrainGenerator(cfg)
	// Генерируем прямоугольную карту для изометрической проекции (50x38 тайлов)
	terrain := terrainGen.GenerateRectangular(50, 38)

	// ИСПРАВЛЕНИЕ: Размеры мира в тайлах для изометрической проекции
	worldWidthTiles := terrain.Width   // 50 тайлов
	worldHeightTiles := terrain.Height // 38 тайлов
	gameWorld := NewGameWorld(worldWidthTiles, worldHeightTiles, seed, terrain)
	timeManager := NewTimeManager()

	// Заполняем мир животными
	gameWorld.PopulateWorld(cfg)

	// Создаём рендерер спрайтов
	spriteRenderer := NewSpriteRenderer()

	// Создаём менеджер шрифтов
	fontManager := NewFontManager()
	if err := fontManager.LoadFonts(); err != nil {
		log.Printf("Предупреждение: не удалось загрузить пользовательские шрифты: %v", err)
		log.Printf("Будет использован дефолтный шрифт")
	}

	// Создаём новую изометрическую систему отрисовки
	isometricRenderer := rendering.NewIsometricRenderer()
	camera := rendering.NewCamera(float32(terrain.Width), float32(terrain.Height))
	camera.SetZoom(1.0) // Стандартный zoom 1x (как требуется)

	// ИСПРАВЛЕНИЕ: Центрируем камеру правильно на центре карты
	mapCenterTileX := float32(terrain.Width) / 2.0
	mapCenterTileY := float32(terrain.Height) / 2.0

	// Изометрическая проекция центра карты в экранные координаты
	centerScreenX := (mapCenterTileX - mapCenterTileY) * 32 / 2 // TileWidth = 32
	centerScreenY := (mapCenterTileX + mapCenterTileY) * 16 / 2 // TileHeight = 16

	// Экран 1024x768, центр в (512, 384)
	// Камера должна сместиться так, чтобы centerScreenX,centerScreenY стали 512,384
	cameraX := centerScreenX - 512
	cameraY := centerScreenY - 384
	camera.SetPosition(cameraX, cameraY)

	// ИСПРАВЛЕНИЕ: Подключаем спрайтовый рендерер к изометрическому
	isometricRenderer.SetSpriteRenderer(spriteRenderer)

	// Создаём игру с менеджерами
	game := &Game{
		gameWorld:         gameWorld,
		timeManager:       timeManager,
		spriteRenderer:    spriteRenderer,
		fontManager:       fontManager,
		isometricRenderer: isometricRenderer,
		camera:            camera,
		terrain:           terrain,
		debugMode:         false, // По умолчанию выключен
	}

	// Настройки окна
	ebiten.SetWindowSize(1024, 768)
	ebiten.SetWindowTitle("Savanna Ecosystem Simulator")
	ebiten.SetWindowResizingMode(ebiten.WindowResizingModeEnabled)
	ebiten.SetVsyncEnabled(true)
	ebiten.SetScreenClearedEveryFrame(true)

	// Запускаем игру
	if err := ebiten.RunGame(game); err != nil {
		log.Fatal(err)
	}
}

// takeDebugScreenshot создаёт скриншот с включённым дебаг-режимом
func (g *Game) takeDebugScreenshot() {
	// Временно включаем дебаг-режим для скриншота
	originalDebugMode := g.debugMode
	g.debugMode = true

	// Создаем изображение размером с экран
	screen := ebiten.NewImage(1024, 768)

	// Рендерим кадр с дебаг-информацией
	g.Draw(screen)

	// Восстанавливаем исходный дебаг-режим
	g.debugMode = originalDebugMode

	// Генерируем имя файла с временной меткой
	timestamp := time.Now().Format("20060102_150405")
	filename := fmt.Sprintf("tmp/debug_screenshot_%s.png", timestamp)

	// Создаем директорию если её нет
	os.MkdirAll("tmp", 0755)

	// Сохраняем скриншот
	rgba := screen.SubImage(screen.Bounds())
	file, err := os.Create(filename)
	if err != nil {
		fmt.Printf("⚠️  Ошибка создания файла %s: %v\n", filename, err)
		return
	}
	defer file.Close()

	err = png.Encode(file, rgba.(image.Image))
	if err != nil {
		fmt.Printf("⚠️  Ошибка сохранения PNG %s: %v\n", filename, err)
		return
	}

	fmt.Printf("📸 Дебаг-скриншот сохранён: %s\n", filename)
}

// min возвращает минимальное из двух float32
func min(a, b float32) float32 {
	if a < b {
		return a
	}
	return b
}

// max возвращает максимальное из двух float32
func max(a, b float32) float32 {
	if a > b {
		return a
	}
	return b
}
