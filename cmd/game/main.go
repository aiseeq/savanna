package main

import (
	"bytes"
	_ "embed"
	"flag"
	"fmt"
	"image/color"
	"log"
	"time"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	"github.com/hajimehoshi/ebiten/v2/text/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"

	"github.com/aiseeq/savanna/config"
	"github.com/aiseeq/savanna/internal/core"
	"github.com/aiseeq/savanna/internal/generator"
	"github.com/aiseeq/savanna/internal/simulation"
)

//go:embed DejaVuSansMono.ttf
var dejaVuSansMonoTTF []byte

var (
	dejaVuSansMonoFace *text.GoTextFace
)

// Camera простая камера для просмотра мира
type Camera struct {
	X, Y float32
	Zoom float32
}

// Game структура для GUI версии симулятора экосистемы саванны
type Game struct {
	world         *core.World
	systemManager *core.SystemManager
	terrain       *generator.Terrain
	camera        Camera
	deltaTime     float32
	timeScale     float32 // Масштаб времени (1.0 = нормально, 2.0 = в 2 раза быстрее)

	// Состояние перетаскивания карты
	isDragging bool
	lastMouseX int
	lastMouseY int
}

// Update обновляет логику игры (вызывается 60 раз в секунду)
func (g *Game) Update() error {
	// Выход по ESC
	if inpututil.IsKeyJustPressed(ebiten.KeyEscape) {
		return fmt.Errorf("игра завершена пользователем")
	}

	// Управление камерой (уменьшено в 10 раз)
	moveSpeed := float32(20.0 / g.camera.Zoom) // Скорость зависит от зума
	if ebiten.IsKeyPressed(ebiten.KeyW) || ebiten.IsKeyPressed(ebiten.KeyArrowUp) {
		g.camera.Y -= moveSpeed
	}
	if ebiten.IsKeyPressed(ebiten.KeyS) || ebiten.IsKeyPressed(ebiten.KeyArrowDown) {
		g.camera.Y += moveSpeed
	}
	if ebiten.IsKeyPressed(ebiten.KeyA) || ebiten.IsKeyPressed(ebiten.KeyArrowLeft) {
		g.camera.X -= moveSpeed
	}
	if ebiten.IsKeyPressed(ebiten.KeyD) || ebiten.IsKeyPressed(ebiten.KeyArrowRight) {
		g.camera.X += moveSpeed
	}

	// Зум колесом мыши с фокусом на курсор
	_, dy := ebiten.Wheel()
	if dy != 0 {
		// Получаем позицию курсора
		mouseX, mouseY := ebiten.CursorPosition()

		// Преобразуем экранные координаты курсора в мировые (до зума)
		worldX := g.camera.X + float32(mouseX)/g.camera.Zoom
		worldY := g.camera.Y + float32(mouseY)/g.camera.Zoom

		// Применяем зум
		zoomFactor := float32(1.2)
		if dy > 0 {
			g.camera.Zoom *= zoomFactor
		} else {
			g.camera.Zoom /= zoomFactor
		}

		// Ограничиваем зум
		if g.camera.Zoom < 0.1 {
			g.camera.Zoom = 0.1
		}
		if g.camera.Zoom > 5.0 {
			g.camera.Zoom = 5.0
		}

		// Корректируем позицию камеры чтобы точка под курсором осталась на месте
		newWorldX := g.camera.X + float32(mouseX)/g.camera.Zoom
		newWorldY := g.camera.Y + float32(mouseY)/g.camera.Zoom

		g.camera.X += worldX - newWorldX
		g.camera.Y += worldY - newWorldY
	}

	// Перетаскивание карты правой кнопкой мыши
	mouseX, mouseY := ebiten.CursorPosition()
	if ebiten.IsMouseButtonPressed(ebiten.MouseButtonRight) {
		if !g.isDragging {
			// Начинаем перетаскивание
			g.isDragging = true
			g.lastMouseX = mouseX
			g.lastMouseY = mouseY
		} else {
			// Продолжаем перетаскивание - двигаем камеру
			deltaX := float32(mouseX - g.lastMouseX)
			deltaY := float32(mouseY - g.lastMouseY)

			// Инвертируем движение чтобы карта двигалась под мышкой
			g.camera.X -= deltaX / g.camera.Zoom
			g.camera.Y -= deltaY / g.camera.Zoom

			g.lastMouseX = mouseX
			g.lastMouseY = mouseY
		}
	} else {
		// Прекращаем перетаскивание
		g.isDragging = false
	}

	// Управление временем
	if inpututil.IsKeyJustPressed(ebiten.KeyEqual) || inpututil.IsKeyJustPressed(ebiten.KeyKPAdd) { // + или Num+
		g.timeScale *= 2.0
		if g.timeScale > 16.0 { // Максимум x16
			g.timeScale = 16.0
		}
	}
	if inpututil.IsKeyJustPressed(ebiten.KeyMinus) || inpututil.IsKeyJustPressed(ebiten.KeyKPSubtract) { // - или Num-
		g.timeScale /= 2.0
		if g.timeScale < 0.125 { // Минимум x0.125 (1/8 скорости)
			g.timeScale = 0.125
		}
	}
	if inpututil.IsKeyJustPressed(ebiten.KeySpace) { // Пробел = пауза/продолжить
		if g.timeScale == 0 {
			g.timeScale = 1.0 // Возобновляем с нормальной скоростью
		} else {
			g.timeScale = 0 // Пауза
		}
	}

	// Обновляем симуляцию с учётом масштаба времени
	g.deltaTime = (1.0 / 60.0) * g.timeScale
	if g.timeScale > 0 { // Только если не на паузе
		g.world.Update(g.deltaTime)
		g.systemManager.Update(g.world, g.deltaTime)
	}

	return nil
}

// Draw отрисовывает кадр
func (g *Game) Draw(screen *ebiten.Image) {
	// Очищаем экран тёмным цветом
	screen.Fill(color.RGBA{20, 30, 20, 255})

	// Отрисовываем ландшафт
	g.drawTerrain(screen)

	// Отрисовываем животных
	g.drawAnimals(screen)

	// Отрисовываем UI
	g.drawUI(screen)
}

// drawTerrain отрисовывает ландшафт
func (g *Game) drawTerrain(screen *ebiten.Image) {
	if g.terrain == nil {
		return
	}

	tileSize := float32(32) * g.camera.Zoom

	// Определяем видимую область
	screenW, screenH := screen.Size()
	startX := int((g.camera.X) / 32)
	startY := int((g.camera.Y) / 32)
	endX := startX + int(float32(screenW)/(32*g.camera.Zoom)) + 2
	endY := startY + int(float32(screenH)/(32*g.camera.Zoom)) + 2

	// Ограничиваем области размером мира
	if startX < 0 {
		startX = 0
	}
	if startY < 0 {
		startY = 0
	}
	if endX >= g.terrain.Size {
		endX = g.terrain.Size - 1
	}
	if endY >= g.terrain.Size {
		endY = g.terrain.Size - 1
	}

	// Отрисовываем тайлы
	for y := startY; y <= endY; y++ {
		for x := startX; x <= endX; x++ {
			screenX := float32(x*32)*g.camera.Zoom - g.camera.X*g.camera.Zoom
			screenY := float32(y*32)*g.camera.Zoom - g.camera.Y*g.camera.Zoom

			var tileColor color.RGBA
			switch g.terrain.GetTileType(x, y) {
			case generator.TileGrass:
				grass := g.terrain.GetGrassAmount(x, y)
				green := uint8(50 + grass*2) // 50-250
				tileColor = color.RGBA{20, green, 20, 255}
			case generator.TileWater:
				tileColor = color.RGBA{30, 50, 180, 255} // Синий
			case generator.TileBush:
				tileColor = color.RGBA{80, 50, 30, 255} // Коричневый
			case generator.TileWetland:
				tileColor = color.RGBA{40, 120, 40, 255} // Тёмно-зелёный
			}

			vector.DrawFilledRect(screen, screenX, screenY, tileSize, tileSize, tileColor, false)
		}
	}
}

// drawAnimals отрисовывает животных
func (g *Game) drawAnimals(screen *ebiten.Image) {
	if g.world == nil {
		return
	}

	// Отрисовываем зайцев (обходим баг QueryByType)
	var rabbits []core.EntityID
	g.world.ForEachWith(core.MaskAnimalType, func(entity core.EntityID) {
		animalType, ok := g.world.GetAnimalType(entity)
		if ok && animalType == core.TypeRabbit {
			rabbits = append(rabbits, entity)
		}
	})
	for _, rabbit := range rabbits {
		pos, hasPos := g.world.GetPosition(rabbit)
		if !hasPos {
			continue
		}

		screenX := pos.X*g.camera.Zoom - g.camera.X*g.camera.Zoom
		screenY := pos.Y*g.camera.Zoom - g.camera.Y*g.camera.Zoom
		radius := float32(5) * g.camera.Zoom

		vector.DrawFilledCircle(screen, screenX, screenY, radius, color.RGBA{180, 180, 180, 255}, false)

		// Отображаем сытость зайца
		if hunger, hasHunger := g.world.GetHunger(rabbit); hasHunger && g.camera.Zoom > 0.3 {
			hungerText := fmt.Sprintf("%.0f", hunger.Value)
			// Позиционируем текст точно на границах пикселей и компенсируем масштаб
			baseX := screenX - 10*g.camera.Zoom
			baseY := screenY - radius - 15*g.camera.Zoom
			textX := float64(int(baseX*2) / 2) // Округляем до половинных пикселей для субпиксельной точности
			textY := float64(int(baseY*2) / 2)

			// Используем Inconsolata для цифр
			op := &text.DrawOptions{}
			op.GeoM.Translate(textX, textY)
			op.ColorScale.ScaleWithColor(color.RGBA{255, 255, 255, 255})
			text.Draw(screen, hungerText, dejaVuSansMonoFace, op)
		}
	}

	// Отрисовываем волков (обходим баг QueryByType)
	var wolves []core.EntityID
	g.world.ForEachWith(core.MaskAnimalType, func(entity core.EntityID) {
		animalType, ok := g.world.GetAnimalType(entity)
		if ok && animalType == core.TypeWolf {
			wolves = append(wolves, entity)
		}
	})
	for _, wolf := range wolves {
		pos, hasPos := g.world.GetPosition(wolf)
		if !hasPos {
			continue
		}

		screenX := pos.X*g.camera.Zoom - g.camera.X*g.camera.Zoom
		screenY := pos.Y*g.camera.Zoom - g.camera.Y*g.camera.Zoom
		radius := float32(8) * g.camera.Zoom

		vector.DrawFilledCircle(screen, screenX, screenY, radius, color.RGBA{120, 60, 40, 255}, false)

		// Отображаем сытость волка
		if hunger, hasHunger := g.world.GetHunger(wolf); hasHunger && g.camera.Zoom > 0.3 {
			hungerText := fmt.Sprintf("%.0f", hunger.Value)
			// Позиционируем текст точно на границах пикселей
			baseX := screenX - 15*g.camera.Zoom
			baseY := screenY - radius - 15*g.camera.Zoom
			textX := float64(int(baseX*2) / 2)
			textY := float64(int(baseY*2) / 2)

			// Используем Inconsolata для цифр
			op := &text.DrawOptions{}
			op.GeoM.Translate(textX, textY)
			op.ColorScale.ScaleWithColor(color.RGBA{255, 255, 255, 255})
			text.Draw(screen, hungerText, dejaVuSansMonoFace, op)
		}
	}
}

// drawUI отрисовывает пользовательский интерфейс
func (g *Game) drawUI(screen *ebiten.Image) {
	if g.world == nil {
		return
	}

	stats := g.world.GetStats()
	rabbits := stats[core.TypeRabbit]
	wolves := stats[core.TypeWolf]

	var timeStatus string
	if g.timeScale == 0 {
		timeStatus = "ПАУЗА"
	} else if g.timeScale == 1.0 {
		timeStatus = "x1.0"
	} else {
		timeStatus = fmt.Sprintf("x%.1f", g.timeScale)
	}

	lines := []string{
		"Симулятор экосистемы саванны",
		fmt.Sprintf("Зайцы: %d  Волки: %d  Время: %s", rabbits, wolves, timeStatus),
		fmt.Sprintf("Камера: %.0f,%.0f  Зум: %.1fx", g.camera.X, g.camera.Y, g.camera.Zoom),
		"",
		"Цифры над животными - уровень сытости (0-100):",
		"  Зайцы: красный<30, жёлтый<60, зелёный≥60",
		"  Волки: красный<40, оранжевый<60 (охотится), голубой≥60",
		"",
		"WASD/стрелки: движение камеры",
		"Колесо мыши: зум",
		"+/- : ускорение/замедление времени",
		"Пробел: пауза",
		"ESC: выход",
	}

	// Отрисовываем каждую строку отдельно нашим шрифтом
	for i, line := range lines {
		op := &text.DrawOptions{}
		op.GeoM.Translate(10.0, float64(15+i*15))
		op.ColorScale.ScaleWithColor(color.RGBA{255, 255, 255, 255})
		text.Draw(screen, line, dejaVuSansMonoFace, op)
	}
}

// Layout возвращает размер экрана
func (g *Game) Layout(outsideWidth, outsideHeight int) (screenWidth, screenHeight int) {
	scale := ebiten.DeviceScaleFactor()

	// Компенсируем DPI scaling только если scale > 1.0 (Windows)
	var realWidth, realHeight int
	if scale > 1.0 {
		// Windows: компенсируем масштабирование
		realWidth = int(float64(outsideWidth) * scale)
		realHeight = int(float64(outsideHeight) * scale)
	} else {
		// WSL/Linux: используем исходные размеры
		realWidth = outsideWidth
		realHeight = outsideHeight
	}


	return realWidth, realHeight
}

func main() {
	// Парсим аргументы командной строки
	var seedFlag = flag.Int64("seed", 0, "Seed для детерминированной симуляции (если не указан, используется текущее время)")
	flag.Parse()

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

	// Принудительно отключаем DPI awareness (только на Windows)
	ebiten.SetRunnableOnUnfocused(true)

	// Инициализируем шрифт
	fontSource, err := text.NewGoTextFaceSource(bytes.NewReader(dejaVuSansMonoTTF))
	if err != nil {
		log.Fatal("Failed to load font:", err)
	}
	dejaVuSansMonoFace = &text.GoTextFace{
		Source: fontSource,
		Size:   12,
	}

	// Инициализируем симуляцию
	cfg := config.LoadDefaultConfig()
	cfg.World.Seed = seed       // Устанавливаем seed из флага
	cfg.Population.Rabbits = 20 // Меньше для демонстрации
	cfg.Population.Wolves = 2

	// Генерируем мир
	terrainGen := generator.NewTerrainGenerator(cfg)
	terrain := terrainGen.Generate()

	// Создаём мир и системы
	worldSizePixels := float32(cfg.World.Size * 32)
	world := core.NewWorld(worldSizePixels, worldSizePixels, seed)
	systemManager := core.NewSystemManager()

	// Создаём системы с зависимостями
	vegetationSystem := simulation.NewVegetationSystem(terrain)
	animalBehaviorSystem := simulation.NewAnimalBehaviorSystem(vegetationSystem)
	feedingSystem := simulation.NewFeedingSystem(vegetationSystem)

	// Добавляем системы в правильном порядке
	systemManager.AddSystem(vegetationSystem)
	systemManager.AddSystem(animalBehaviorSystem)
	systemManager.AddSystem(simulation.NewMovementSystem(worldSizePixels, worldSizePixels))
	systemManager.AddSystem(feedingSystem)

	// Размещаем животных
	popGen := generator.NewPopulationGenerator(cfg, terrain)
	placements := popGen.Generate()

	// Создаём животных на основе сгенерированных позиций
	for _, placement := range placements {
		switch placement.Type {
		case core.TypeRabbit:
			simulation.CreateRabbit(world, placement.X, placement.Y)
		case core.TypeWolf:
			simulation.CreateWolf(world, placement.X, placement.Y)
		}
	}

	// Создаём игру
	game := &Game{
		world:         world,
		systemManager: systemManager,
		terrain:       terrain,
		camera: Camera{
			X:    -40, // Центр камеры
			Y:    -40,
			Zoom: 0.5, // Показываем весь мир
		},
		timeScale: 1.0, // Нормальная скорость времени
	}

	ebiten.SetWindowSize(1024, 768)
	ebiten.SetWindowTitle("Savanna Ecosystem Simulator")
	ebiten.SetWindowResizable(true)

	// Отключаем DPI scaling для четкого отображения
	ebiten.SetScreenTransparent(false)
	ebiten.SetVsyncEnabled(true)
	ebiten.SetScreenClearedEveryFrame(true)

	// Устанавливаем пиксельно-точное отображение
	fmt.Printf("Device scale factor: %.2f\n", ebiten.DeviceScaleFactor())

	if err := ebiten.RunGame(game); err != nil {
		log.Fatal(err)
	}
}
