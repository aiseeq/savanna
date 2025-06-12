package main

import (
	"fmt"
	"image/color"
	"log"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	"github.com/hajimehoshi/ebiten/v2/vector"
	
	"github.com/aiseeq/savanna/config"
	"github.com/aiseeq/savanna/internal/core"
	"github.com/aiseeq/savanna/internal/generator"
	"github.com/aiseeq/savanna/internal/simulation"
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

	// Зум колесом мыши
	_, dy := ebiten.Wheel()
	if dy != 0 {
		zoomFactor := float32(1.1)
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
	if startX < 0 { startX = 0 }
	if startY < 0 { startY = 0 }
	if endX >= g.terrain.Size { endX = g.terrain.Size - 1 }
	if endY >= g.terrain.Size { endY = g.terrain.Size - 1 }

	// Отрисовываем тайлы
	for y := startY; y <= endY; y++ {
		for x := startX; x <= endX; x++ {
			screenX := float32(x*32) * g.camera.Zoom - g.camera.X * g.camera.Zoom
			screenY := float32(y*32) * g.camera.Zoom - g.camera.Y * g.camera.Zoom

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

	// Отрисовываем зайцев
	rabbits := g.world.QueryByType(core.TypeRabbit)
	for _, rabbit := range rabbits {
		pos, hasPos := g.world.GetPosition(rabbit)
		if !hasPos {
			continue
		}

		screenX := pos.X * g.camera.Zoom - g.camera.X * g.camera.Zoom
		screenY := pos.Y * g.camera.Zoom - g.camera.Y * g.camera.Zoom
		radius := float32(5) * g.camera.Zoom

		vector.DrawFilledCircle(screen, screenX, screenY, radius, color.RGBA{180, 180, 180, 255}, false)
	}

	// Отрисовываем волков
	wolves := g.world.QueryByType(core.TypeWolf)
	for _, wolf := range wolves {
		pos, hasPos := g.world.GetPosition(wolf)
		if !hasPos {
			continue
		}

		screenX := pos.X * g.camera.Zoom - g.camera.X * g.camera.Zoom
		screenY := pos.Y * g.camera.Zoom - g.camera.Y * g.camera.Zoom
		radius := float32(8) * g.camera.Zoom

		vector.DrawFilledCircle(screen, screenX, screenY, radius, color.RGBA{120, 60, 40, 255}, false)
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

	msg := fmt.Sprintf(`Savanna Ecosystem Simulator
Зайцы: %d  Волки: %d  Время: %s
Камера: %.0f,%.0f  Зум: %.1fx

WASD/стрелки: движение камеры
Колесо мыши: зум
+/- : ускорение/замедление времени
Пробел: пауза
ESC: выход`, rabbits, wolves, timeStatus, g.camera.X, g.camera.Y, g.camera.Zoom)

	ebitenutil.DebugPrint(screen, msg)
}

// Layout возвращает размер экрана
func (g *Game) Layout(outsideWidth, outsideHeight int) (screenWidth, screenHeight int) {
	return 1024, 768
}

func main() {
	fmt.Println("Запуск GUI версии симулятора экосистемы саванны...")

	// Инициализируем симуляцию
	cfg := config.LoadDefaultConfig()
	cfg.Population.Rabbits = 20 // Меньше для демонстрации
	cfg.Population.Wolves = 2

	// Генерируем мир
	terrainGen := generator.NewTerrainGenerator(cfg)
	terrain := terrainGen.Generate()

	// Создаём мир и системы
	worldSizePixels := float32(cfg.World.Size * 32)
	world := core.NewWorld(worldSizePixels, worldSizePixels, cfg.World.Seed)
	systemManager := core.NewSystemManager()

	// Добавляем системы
	systemManager.AddSystem(simulation.NewAnimalBehaviorSystem())
	systemManager.AddSystem(simulation.NewMovementSystem(worldSizePixels, worldSizePixels))
	systemManager.AddSystem(simulation.NewFeedingSystem())

	// Размещаем животных
	popGen := generator.NewPopulationGenerator(cfg, terrain)
	popGen.Generate(world)

	// Создаём игру
	game := &Game{
		world:         world,
		systemManager: systemManager,
		terrain:       terrain,
		camera: Camera{
			X:    worldSizePixels / 4, // Начинаем в центре мира
			Y:    worldSizePixels / 4,
			Zoom: 0.5, // Показываем весь мир
		},
		timeScale: 1.0, // Нормальная скорость времени
	}

	ebiten.SetWindowSize(1024, 768)
	ebiten.SetWindowTitle("Savanna Ecosystem Simulator")
	ebiten.SetWindowResizable(true)

	if err := ebiten.RunGame(game); err != nil {
		log.Fatal(err)
	}
}
