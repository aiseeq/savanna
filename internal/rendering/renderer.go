package rendering

import (
	"image/color"
	"math"
	"sort"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"

	"github.com/aiseeq/savanna/internal/core"
	"github.com/aiseeq/savanna/internal/generator"
)

// Константы изометрической проекции
const (
	TileWidth  = 32 // Ширина тайла в пикселях
	TileHeight = 16 // Высота тайла в пикселях (для изометрии обычно половина ширины)
)

// SpriteRenderer интерфейс для отрисовки спрайтов животных
type SpriteRenderer interface {
	DrawAnimalAt(screen *ebiten.Image, world *core.World, entity core.EntityID, screenX, screenY, zoom float32)
}

// IsometricRenderer отвечает за изометрическую отрисовку мира
type IsometricRenderer struct {
	tileOptions    *ebiten.DrawImageOptions // Переиспользуемые опции для оптимизации
	spriteRenderer SpriteRenderer           // Опциональный рендерер спрайтов
	whitePixel     *ebiten.Image            // Белый пиксель для заливки
}

// NewIsometricRenderer создаёт новый изометрический рендерер
func NewIsometricRenderer() *IsometricRenderer {
	// Создаем белый пиксель для заливки один раз
	whitePixel := ebiten.NewImage(1, 1)
	whitePixel.Fill(color.RGBA{255, 255, 255, 255})

	return &IsometricRenderer{
		tileOptions: &ebiten.DrawImageOptions{},
		whitePixel:  whitePixel,
	}
}

// SetSpriteRenderer устанавливает рендерер спрайтов
func (r *IsometricRenderer) SetSpriteRenderer(spriteRenderer SpriteRenderer) {
	r.spriteRenderer = spriteRenderer
}

// WorldToScreen преобразует мировые координаты в экранные (изометрическая проекция)
func (r *IsometricRenderer) WorldToScreen(worldX, worldY float32) (screenX, screenY float32) {
	// Классическая формула изометрической проекции
	screenX = (worldX - worldY) * TileWidth / 2
	screenY = (worldX + worldY) * TileHeight / 2
	return screenX, screenY
}

// ScreenToWorld преобразует экранные координаты в мировые
func (r *IsometricRenderer) ScreenToWorld(screenX, screenY float32) (worldX, worldY float32) {
	// Обратная формула изометрической проекции
	worldX = (screenX/(TileWidth/2) + screenY/(TileHeight/2)) / 2
	worldY = (screenY/(TileHeight/2) - screenX/(TileWidth/2)) / 2
	return worldX, worldY
}

// RenderWorld отрисовывает весь мир в правильном порядке согласно этапу 7
func (r *IsometricRenderer) RenderWorld(screen *ebiten.Image, terrain *generator.Terrain, world *core.World, camera *Camera) {
	// Порядок отрисовки (критически важен для изометрии):
	// 1. Тайлы местности (трава, вода)
	r.renderTerrain(screen, terrain, camera)

	// 2. Кусты и препятствия
	r.renderObstacles(screen, terrain, camera)

	// 3. Животные, отсортированные по Y (дальние сначала)
	r.renderAnimals(screen, world, camera)

	// DEBUG: Рисуем красный крест в ЦЕНТРЕ КАРТЫ (не экрана)
	mapCenterTileX := float32(terrain.Width) / 2.0
	mapCenterTileY := float32(terrain.Height) / 2.0
	centerScreenX, centerScreenY := camera.WorldToScreen(mapCenterTileX, mapCenterTileY)

	crossSize := float32(20)
	red := color.RGBA{255, 0, 0, 255}
	vector.StrokeLine(screen, centerScreenX-crossSize, centerScreenY, centerScreenX+crossSize, centerScreenY, 3, red, false)
	vector.StrokeLine(screen, centerScreenX, centerScreenY-crossSize, centerScreenX, centerScreenY+crossSize, 3, red, false)
}

// renderTerrain отрисовывает тайлы местности
func (r *IsometricRenderer) renderTerrain(screen *ebiten.Image, terrain *generator.Terrain, camera *Camera) {
	// Определяем видимую область для frustum culling
	minX, minY, maxX, maxY := r.getVisibleTiles(screen, camera)

	// Убрано DEBUG - видимая область считается правильно

	// Ограничиваем видимую область размерами terrain
	if minX < 0 {
		minX = 0
	}
	if minY < 0 {
		minY = 0
	}
	if maxX >= terrain.Width {
		maxX = terrain.Width - 1
	}
	if maxY >= terrain.Height {
		maxY = terrain.Height - 1
	}

	for y := minY; y <= maxY; y++ {
		for x := minX; x <= maxX; x++ {
			r.renderTile(screen, terrain, x, y, camera)
		}
	}
}

// renderTile отрисовывает один тайл местности
func (r *IsometricRenderer) renderTile(screen *ebiten.Image, terrain *generator.Terrain, tileX, tileY int, camera *Camera) {
	// Преобразуем координаты тайла в экранные с учётом камеры
	worldX := float32(tileX)
	worldY := float32(tileY)
	screenX, screenY := camera.WorldToScreen(worldX, worldY)

	// Определяем цвет тайла по типу
	var tileColor color.RGBA
	switch terrain.Tiles[tileY][tileX] {
	case generator.TileGrass:
		// Интенсивность зелёного зависит от количества травы
		grassAmount := terrain.Grass[tileY][tileX] / 100.0
		green := uint8(50 + grassAmount*150) // От 50 до 200
		tileColor = color.RGBA{R: 34, G: green, B: 34, A: 255}
	case generator.TileWater:
		tileColor = color.RGBA{R: 64, G: 164, B: 223, A: 255} // Голубая вода
	case generator.TileBush:
		tileColor = color.RGBA{R: 34, G: 139, B: 34, A: 255} // Тёмно-зелёные кусты
	case generator.TileWetland:
		tileColor = color.RGBA{R: 139, G: 69, B: 19, A: 255} // Коричневая влажная земля
	default:
		tileColor = color.RGBA{R: 128, G: 128, B: 128, A: 255} // Серый для неизвестных
	}

	// Рисуем ромб (изометрический тайл) с учётом zoom камеры
	r.drawIsometricTile(screen, screenX, screenY, tileColor, camera)
}

// drawIsometricTile рисует изометрический тайл МАКСИМАЛЬНО ЭФФЕКТИВНО
func (r *IsometricRenderer) drawIsometricTile(screen *ebiten.Image, x, y float32, col color.RGBA, camera *Camera) {
	// Учитываем zoom камеры для размера тайлов
	zoom := camera.GetZoom()

	// СУПЕРOPTIMIZATION: Используем простые фигуры для разных уровней детализации

	if zoom < 0.5 {
		// При маленьком zoom рисуем просто точки - максимальная производительность
		vector.DrawFilledCircle(screen, x, y, 1, col, false)
		return
	}

	if zoom < 1.0 {
		// При среднем zoom рисуем небольшие прямоугольники
		size := zoom * 8 // От 4 до 8 пикселей
		vector.DrawFilledRect(screen, x-size/2, y-size/2, size, size, col, false)
		return
	}

	// При крупном zoom рисуем ромбы как сплошные многоугольники
	halfWidth := float32(TileWidth) * zoom / 2
	halfHeight := float32(TileHeight) * zoom / 2

	centerX, centerY := x, y

	// Рисуем ромб как заполненный многоугольник (4 вершины)
	topX, topY := centerX, centerY-halfHeight
	rightX, rightY := centerX+halfWidth, centerY
	bottomX, bottomY := centerX, centerY+halfHeight
	leftX, leftY := centerX-halfWidth, centerY

	// Создаем путь для заливки ромба
	var vertices []ebiten.Vertex

	// Добавляем 4 вершины ромба
	vertices = append(vertices, ebiten.Vertex{
		DstX: topX, DstY: topY,
		ColorR: float32(col.R) / 255, ColorG: float32(col.G) / 255,
		ColorB: float32(col.B) / 255, ColorA: float32(col.A) / 255,
	})
	vertices = append(vertices, ebiten.Vertex{
		DstX: rightX, DstY: rightY,
		ColorR: float32(col.R) / 255, ColorG: float32(col.G) / 255,
		ColorB: float32(col.B) / 255, ColorA: float32(col.A) / 255,
	})
	vertices = append(vertices, ebiten.Vertex{
		DstX: bottomX, DstY: bottomY,
		ColorR: float32(col.R) / 255, ColorG: float32(col.G) / 255,
		ColorB: float32(col.B) / 255, ColorA: float32(col.A) / 255,
	})
	vertices = append(vertices, ebiten.Vertex{
		DstX: leftX, DstY: leftY,
		ColorR: float32(col.R) / 255, ColorG: float32(col.G) / 255,
		ColorB: float32(col.B) / 255, ColorA: float32(col.A) / 255,
	})

	// Индексы для треугольников (два треугольника составляют ромб)
	indices := []uint16{0, 1, 2, 0, 2, 3}

	// Рисуем заполненный ромб используя переиспользуемый белый пиксель
	screen.DrawTriangles(vertices, indices, r.whitePixel, nil)

	// Границы рисуем только при очень крупном zoom
	if zoom > 1.5 {
		borderColor := color.RGBA{R: col.R - 20, G: col.G - 20, B: col.B - 20, A: 255}

		// Рисуем контур ромба (4 линии)
		topX, topY := centerX, centerY-halfHeight
		rightX, rightY := centerX+halfWidth, centerY
		bottomX, bottomY := centerX, centerY+halfHeight
		leftX, leftY := centerX-halfWidth, centerY

		vector.StrokeLine(screen, topX, topY, rightX, rightY, 1, borderColor, false)
		vector.StrokeLine(screen, rightX, rightY, bottomX, bottomY, 1, borderColor, false)
		vector.StrokeLine(screen, bottomX, bottomY, leftX, leftY, 1, borderColor, false)
		vector.StrokeLine(screen, leftX, leftY, topX, topY, 1, borderColor, false)
	}
}

// renderObstacles отрисовывает кусты и препятствия
func (r *IsometricRenderer) renderObstacles(screen *ebiten.Image, terrain *generator.Terrain, camera *Camera) {
	// Определяем видимую область
	minX, minY, maxX, maxY := r.getVisibleTiles(screen, camera)

	// Ограничиваем видимую область размерами terrain
	if minX < 0 {
		minX = 0
	}
	if minY < 0 {
		minY = 0
	}
	if maxX >= terrain.Width {
		maxX = terrain.Width - 1
	}
	if maxY >= terrain.Height {
		maxY = terrain.Height - 1
	}

	for y := minY; y <= maxY; y++ {
		for x := minX; x <= maxX; x++ {
			if terrain.Tiles[y][x] == generator.TileBush {
				r.renderBush(screen, x, y, camera)
			}
		}
	}
}

// renderBush отрисовывает куст
func (r *IsometricRenderer) renderBush(screen *ebiten.Image, tileX, tileY int, camera *Camera) {
	worldX := float32(tileX)
	worldY := float32(tileY)
	screenX, screenY := camera.WorldToScreen(worldX, worldY)

	// Рисуем куст как тёмно-зелёный круг поверх тайла
	bushColor := color.RGBA{R: 0, G: 100, B: 0, A: 255}
	radius := float32(TileWidth) / 4
	vector.DrawFilledCircle(screen, screenX, screenY-4, radius, bushColor, false)
}

// renderAnimals отрисовывает животных, отсортированных по Y
func (r *IsometricRenderer) renderAnimals(screen *ebiten.Image, world *core.World, camera *Camera) {
	// Собираем всех животных с их Y координатами для сортировки
	type AnimalRenderInfo struct {
		entity core.EntityID
		y      float32
	}

	var animals []AnimalRenderInfo

	// Собираем животных
	world.ForEachWith(core.MaskPosition|core.MaskAnimalType, func(entity core.EntityID) {
		if pos, hasPos := world.GetPosition(entity); hasPos {
			animals = append(animals, AnimalRenderInfo{
				entity: entity,
				y:      pos.Y,
			})
		}
	})

	// ОПТИМИЗАЦИЯ: Используем sort.SliceStable для O(n log n) вместо bubble sort O(n²)
	// sort.SliceStable сохраняет детерминированность при равных значениях Y
	sort.SliceStable(animals, func(i, j int) bool {
		return animals[i].y < animals[j].y // Дальние объекты (меньший Y) рисуются сначала
	})

	// Отрисовываем в отсортированном порядке
	for _, animal := range animals {
		r.renderAnimal(screen, world, animal.entity, camera)
	}
}

// renderAnimal отрисовывает одно животное
func (r *IsometricRenderer) renderAnimal(screen *ebiten.Image, world *core.World, entity core.EntityID, camera *Camera) {
	pos, hasPos := world.GetPosition(entity)
	if !hasPos {
		return
	}

	// Преобразуем в экранные координаты с учётом камеры
	screenX, screenY := camera.WorldToScreen(pos.X, pos.Y)

	// DEBUG: Отладочный вывод удален для предотвращения спама в консоли

	// ИСПРАВЛЕНИЕ: Используем ТОЛЬКО спрайты - никаких кругов!
	if r.spriteRenderer != nil {
		r.spriteRenderer.DrawAnimalAt(screen, world, entity, screenX, screenY, camera.GetZoom())
		return // Возвращаемся сразу - спрайты есть
	}

	// FALLBACK: Если мы здесь, то spriteRenderer == nil - рисуем круги

	// FALLBACK: Рисуем простые круги только если спрайтов НЕТ
	var animalColor color.RGBA
	if animalType, hasType := world.GetAnimalType(entity); hasType {
		switch animalType {
		case core.TypeRabbit:
			animalColor = color.RGBA{R: 139, G: 69, B: 19, A: 255} // Коричневый заяц
		case core.TypeWolf:
			animalColor = color.RGBA{R: 105, G: 105, B: 105, A: 255} // Серый волк
		default:
			animalColor = color.RGBA{R: 255, G: 0, B: 255, A: 255} // Магента для неизвестных
		}
	} else {
		animalColor = color.RGBA{R: 255, G: 255, B: 255, A: 255} // Белый если тип неизвестен
	}

	// Получаем размер животного
	radius := float32(8) // Значение по умолчанию
	if size, hasSize := world.GetSize(entity); hasSize {
		radius = size.Radius
	}

	// Рисуем животное как круг
	vector.DrawFilledCircle(screen, screenX, screenY, radius, animalColor, false)

	// Добавляем чёрную границу
	borderColor := color.RGBA{R: 0, G: 0, B: 0, A: 255}
	vector.StrokeCircle(screen, screenX, screenY, radius, 1, borderColor, false)
}

// getVisibleTiles возвращает диапазон видимых тайлов для frustum culling
func (r *IsometricRenderer) getVisibleTiles(screen *ebiten.Image, camera *Camera) (minX, minY, maxX, maxY int) {
	// Размеры экрана
	screenWidth := float32(screen.Bounds().Dx())
	screenHeight := float32(screen.Bounds().Dy())

	// Углы экрана в мировых координатах
	topLeftX, topLeftY := camera.ScreenToWorld(0, 0)
	topRightX, topRightY := camera.ScreenToWorld(screenWidth, 0)
	bottomLeftX, bottomLeftY := camera.ScreenToWorld(0, screenHeight)
	bottomRightX, bottomRightY := camera.ScreenToWorld(screenWidth, screenHeight)

	// Находим границы видимой области
	minX = int(math.Floor(float64(min(min(topLeftX, topRightX), min(bottomLeftX, bottomRightX)))))
	minY = int(math.Floor(float64(min(min(topLeftY, topRightY), min(bottomLeftY, bottomRightY)))))
	maxX = int(math.Ceil(float64(max(max(topLeftX, topRightX), max(bottomLeftX, bottomRightX)))))
	maxY = int(math.Ceil(float64(max(max(topLeftY, topRightY), max(bottomLeftY, bottomRightY)))))

	return minX, minY, maxX, maxY
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
