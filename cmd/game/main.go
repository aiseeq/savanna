package main

import (
	"flag"
	"fmt"
	"image/color"
	"log"
	"time"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	"github.com/hajimehoshi/ebiten/v2/text/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"

	"github.com/aiseeq/savanna/config"
	"github.com/aiseeq/savanna/internal/core"
	"github.com/aiseeq/savanna/internal/generator"
	"github.com/aiseeq/savanna/internal/simulation"
)

// Game структура для GUI версии симулятора экосистемы саванны
// Рефакторинг: разбита на специализированные менеджеры (соблюдение SRP)
type Game struct {
	// Менеджеры с единственными ответственностями
	gameWorld        *GameWorld        // Управление симуляцией мира
	cameraController *CameraController // Управление камерой
	timeManager      *TimeManager      // Управление временем
	spriteRenderer   *SpriteRenderer   // Отрисовка спрайтов животных
	fontManager      *FontManager      // Управление шрифтами
}

// Update обновляет логику игры (рефакторинг: использует менеджеры)
func (g *Game) Update() error {
	// Проверяем выход
	if inpututil.IsKeyJustPressed(ebiten.KeyEscape) {
		return fmt.Errorf("игра завершена пользователем")
	}

	// Обновляем менеджеры (каждый отвечает за свою область)
	g.cameraController.Update() // Управление камерой
	g.timeManager.Update()      // Управление временем

	// Обновляем симуляцию с учётом времени
	deltaTime := g.timeManager.GetDeltaTime()
	g.gameWorld.Update(deltaTime)

	return nil
}

// Draw отрисовывает кадр
func (g *Game) Draw(screen *ebiten.Image) {
	// Очищаем экран тёмным цветом
	screen.Fill(color.RGBA{20, 30, 20, 255})

	// LoD Compliance: используем инкапсулированные методы
	camera := g.cameraController.GetCamera()

	// Отрисовываем ландшафт (Game больше не знает о внутренних объектах)
	g.gameWorld.DrawTerrain(screen, camera, g)

	// Отрисовываем животных (Game больше не знает о внутренних объектах)
	g.gameWorld.DrawAnimals(screen, camera, g)

	// Отрисовываем UI
	g.drawUI(screen, camera)
}

// Layout устанавливает размеры экрана
func (g *Game) Layout(outsideWidth, outsideHeight int) (screenWidth, screenHeight int) {
	return outsideWidth, outsideHeight
}

// drawTerrain отрисовывает ландшафт
// DrawTerrain реализует TerrainRenderer (LoD compliance)
func (g *Game) DrawTerrain(screen *ebiten.Image, camera Camera, terrain *generator.Terrain) {
	g.drawTerrain(screen, camera, terrain)
}

func (g *Game) drawTerrain(screen *ebiten.Image, camera Camera, terrain *generator.Terrain) {
	if terrain == nil {
		return
	}

	// Определяем видимую область
	bounds := screen.Bounds()
	screenW, screenH := bounds.Dx(), bounds.Dy()
	startX := int((camera.X) / 32)
	startY := int((camera.Y) / 32)
	endX := startX + int(float32(screenW)/(32*camera.Zoom)) + 2
	endY := startY + int(float32(screenH)/(32*camera.Zoom)) + 2

	// Ограничиваем области размером мира
	size := terrain.GetSize()
	if startX < 0 {
		startX = 0
	}
	if startY < 0 {
		startY = 0
	}
	if endX > size {
		endX = size
	}
	if endY > size {
		endY = size
	}

	// Отрисовываем тайлы
	for y := startY; y < endY; y++ {
		for x := startX; x < endX; x++ {
			screenX := float32(x*32)*camera.Zoom - camera.X*camera.Zoom
			screenY := float32(y*32)*camera.Zoom - camera.Y*camera.Zoom

			// Получаем тип тайла и траву
			tileType := terrain.GetTileType(x, y)
			grassAmount := terrain.GetGrassAmount(x, y)

			// Определяем цвет тайла
			var tileColor color.RGBA
			switch tileType {
			case generator.TileGrass:
				// Цвет травы зависит от количества
				green := uint8(50 + grassAmount*2) // 50-250
				tileColor = color.RGBA{20, green, 20, 255}
			case generator.TileWater:
				tileColor = color.RGBA{30, 50, 150, 255}
			case generator.TileBush:
				tileColor = color.RGBA{60, 80, 40, 255}
			case generator.TileWetland:
				tileColor = color.RGBA{40, 120, 60, 255}
			default:
				tileColor = color.RGBA{100, 100, 100, 255}
			}

			// Отрисовываем тайл
			tileSize := 32 * camera.Zoom
			vector.DrawFilledRect(screen, screenX, screenY, tileSize, tileSize, tileColor, false)
		}
	}
}

// drawAnimals отрисовывает всех животных
// DrawAnimals реализует AnimalRenderer (LoD compliance)
func (g *Game) DrawAnimals(screen *ebiten.Image, camera Camera, world *core.World) {
	g.drawAnimals(screen, camera, world)
}

func (g *Game) drawAnimals(screen *ebiten.Image, camera Camera, world *core.World) {
	world.ForEachWith(core.MaskPosition|core.MaskAnimalType, func(entity core.EntityID) {
		pos, ok := world.GetPosition(entity)
		if !ok {
			return
		}

		_, ok = world.GetAnimalType(entity)
		if !ok {
			return
		}

		// Вычисляем позицию на экране
		screenX := pos.X*camera.Zoom - camera.X*camera.Zoom
		screenY := pos.Y*camera.Zoom - camera.Y*camera.Zoom

		// Проверяем видимость
		bounds := screen.Bounds()
		if screenX < -50 || screenY < -50 || screenX > float32(bounds.Dx())+50 || screenY > float32(bounds.Dy())+50 {
			return
		}

		// Отрисовываем животное как спрайт с анимацией
		g.spriteRenderer.DrawAnimal(screen, world, entity, RenderParams{
			ScreenX: screenX,
			ScreenY: screenY,
			Zoom:    camera.Zoom,
		})

		// Получаем размер для полоски здоровья из компонента Size
		radius := g.getAnimalRadius(entity, world) * camera.Zoom

		// Отрисовываем полоску здоровья
		g.drawHealthBar(screen, entity, world, HealthBarParams{
			ScreenX: screenX,
			ScreenY: screenY,
			Radius:  radius,
		})

		// Отрисовываем значение голода над животным
		g.drawHungerText(screen, entity, world, HungerTextParams{
			ScreenX: screenX,
			ScreenY: screenY,
			Radius:  radius,
		})

		// DamageFlash теперь применяется прямо к спрайту в SpriteRenderer
	})
}

// drawUI отрисовывает пользовательский интерфейс
func (g *Game) drawUI(screen *ebiten.Image, camera Camera) {
	stats := g.gameWorld.GetStats()

	// Получаем шрифт для отрисовки
	font := g.fontManager.GetDebugFont()

	// Создаём текстовую информацию
	y := float64(10)
	lineHeight := float64(20)

	// Статистика животных
	rabbitCount := stats["rabbits"].(int)
	wolfCount := stats["wolves"].(int)
	g.drawText(screen, fmt.Sprintf("Rabbits: %d", rabbitCount), 10, y, font)
	y += lineHeight
	g.drawText(screen, fmt.Sprintf("Wolves: %d", wolfCount), 10, y, font)
	y += lineHeight

	// Масштаб и скорость
	g.drawText(screen, fmt.Sprintf("Zoom: %.1fx", camera.Zoom), 10, y, font)
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

	// Размеры полоски здоровья
	barWidth := params.Radius * 2
	barHeight := float32(4)
	barX := params.ScreenX - barWidth/2
	barY := params.ScreenY - params.Radius - barHeight - 2

	// Фон полоски (красный)
	vector.DrawFilledRect(screen, barX, barY, barWidth, barHeight, color.RGBA{200, 50, 50, 255}, false)

	//nolint:gocritic // commentedOutCode: Это описательный комментарий, не код
	// Здоровье (зелёный)
	healthPercent := float32(health.Current) / float32(health.Max)
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

	// Позиция текста (над полоской здоровья)
	textX := float64(params.ScreenX)
	textY := float64(params.ScreenY - params.Radius - 25) // Над полоской здоровья

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

func main() {
	// Парсим аргументы командной строки
	var seedFlag = flag.Int64(
		"seed", 0,
		"Seed для детерминированной симуляции (если не указан, используется текущее время)",
	)
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

	// Создаём конфигурацию и ландшафт
	cfg := config.LoadDefaultConfig()
	cfg.World.Seed = seed
	terrainGen := generator.NewTerrainGenerator(cfg)
	terrain := terrainGen.Generate()

	// Создаём менеджеры (рефакторинг: разделение ответственностей)
	gameWorld := NewGameWorld(1600, 1600, seed, terrain)
	cameraController := NewCameraController()
	timeManager := NewTimeManager()

	// Заполняем мир животными
	gameWorld.PopulateWorld()

	// Создаём рендерер спрайтов
	spriteRenderer := NewSpriteRenderer()

	// Создаём менеджер шрифтов
	fontManager := NewFontManager()
	if err := fontManager.LoadFonts(); err != nil {
		log.Printf("Предупреждение: не удалось загрузить пользовательские шрифты: %v", err)
		log.Printf("Будет использован дефолтный шрифт")
	}

	// Создаём игру с менеджерами
	game := &Game{
		gameWorld:        gameWorld,
		cameraController: cameraController,
		timeManager:      timeManager,
		spriteRenderer:   spriteRenderer,
		fontManager:      fontManager,
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
