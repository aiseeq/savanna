package rendering

import (
	"github.com/hajimehoshi/ebiten/v2"
)

// Camera управляет позицией и масштабом камеры для изометрической проекции
type Camera struct {
	X, Y float32 // Позиция камеры в экранных координатах
	Zoom float32 // Масштаб: 0.5x, 1x, 2x, 4x

	// Границы мира для ограничения движения камеры
	worldWidth  float32
	worldHeight float32

	// Параметры управления
	moveSpeed float32 // Скорость движения камеры (пиксели в секунду)

	// ИСПРАВЛЕНИЕ: Состояние мыши для скроллинга правой кнопкой
	mouseRightPressed bool    // Нажата ли правая кнопка мыши
	lastMouseX        float32 // Последняя позиция мыши X
	lastMouseY        float32 // Последняя позиция мыши Y
}

// NewCamera создаёт новую камеру
func NewCamera(worldWidth, worldHeight float32) *Camera {
	// Центрируем камеру на середине мира в изометрической проекции
	centerX := (worldWidth - worldHeight) * TileWidth / 4
	centerY := (worldWidth + worldHeight) * TileHeight / 4

	return &Camera{
		X:           -centerX, // Отрицательное смещение для центрирования
		Y:           -centerY,
		Zoom:        1.0,
		worldWidth:  worldWidth,
		worldHeight: worldHeight,
		moveSpeed:   200.0, // 200 пикселей в секунду
	}
}

// Update обновляет состояние камеры по пользовательскому вводу
func (c *Camera) Update(deltaTime float32) {
	// Управление движением камеры (WASD)
	moveDistance := c.moveSpeed * deltaTime

	if ebiten.IsKeyPressed(ebiten.KeyW) || ebiten.IsKeyPressed(ebiten.KeyUp) {
		c.Y -= moveDistance
	}
	if ebiten.IsKeyPressed(ebiten.KeyS) || ebiten.IsKeyPressed(ebiten.KeyDown) {
		c.Y += moveDistance
	}
	if ebiten.IsKeyPressed(ebiten.KeyA) || ebiten.IsKeyPressed(ebiten.KeyLeft) {
		c.X -= moveDistance
	}
	if ebiten.IsKeyPressed(ebiten.KeyD) || ebiten.IsKeyPressed(ebiten.KeyRight) {
		c.X += moveDistance
	}

	// ИСПРАВЛЕНИЕ: Управление камерой правой кнопкой мыши (скроллинг)
	if ebiten.IsMouseButtonPressed(ebiten.MouseButtonRight) {
		// Получаем текущую позицию мыши
		mouseX, mouseY := ebiten.CursorPosition()

		// Статическая переменная для хранения предыдущей позиции мыши
		// ЗАМЕТКА: это упрощенная реализация, в продакшене лучше хранить в структуре камеры
		if c.moveSpeed > 0 { // Проверяем что камера инициализирована
			// Статические переменные для отслеживания движения мыши
			// TODO: перенести в структуру Camera для лучшей архитектуры
			if !c.mouseRightPressed {
				c.lastMouseX = float32(mouseX)
				c.lastMouseY = float32(mouseY)
				c.mouseRightPressed = true
			} else {
				// Вычисляем смещение мыши
				deltaX := c.lastMouseX - float32(mouseX)
				deltaY := c.lastMouseY - float32(mouseY)

				// Применяем смещение к камере (инвертируем для естественного движения)
				mouseSensitivity := float32(2.0)
				c.X += deltaX * mouseSensitivity
				c.Y += deltaY * mouseSensitivity

				// Обновляем последнюю позицию мыши
				c.lastMouseX = float32(mouseX)
				c.lastMouseY = float32(mouseY)
			}
		}
	} else {
		c.mouseRightPressed = false
	}

	// Управление масштабом (колесо мыши с привязкой к позиции)
	_, wheelY := ebiten.Wheel()
	if wheelY != 0 {
		// Получаем позицию мыши
		mouseX, mouseY := ebiten.CursorPosition()

		// Преобразуем в мировые координаты ПЕРЕД изменением зума
		worldX, worldY := c.ScreenToWorld(float32(mouseX), float32(mouseY))

		// Сохраняем старый зум
		oldZoom := c.Zoom

		// Изменяем зум
		if wheelY > 0 {
			c.zoomIn()
		} else {
			c.zoomOut()
		}

		// Если зум реально изменился
		if c.Zoom != oldZoom {
			// Преобразуем мировую точку обратно в экранные координаты с НОВЫМ зумом
			newScreenX, newScreenY := c.WorldToScreen(worldX, worldY)

			// Вычисляем разность и корректируем позицию камеры
			deltaX := newScreenX - float32(mouseX)
			deltaY := newScreenY - float32(mouseY)

			c.X += deltaX
			c.Y += deltaY
		}
	}

	// Ограничиваем позицию камеры границами мира
	c.constrainToWorld()
}

// zoomIn увеличивает масштаб
func (c *Camera) zoomIn() {
	switch c.Zoom {
	case 0.5:
		c.Zoom = 1.0
	case 1.0:
		c.Zoom = 2.0
	case 2.0:
		c.Zoom = 4.0
		// При 4x не увеличиваем дальше
	}
}

// zoomOut уменьшает масштаб
func (c *Camera) zoomOut() {
	switch c.Zoom {
	case 4.0:
		c.Zoom = 2.0
	case 2.0:
		c.Zoom = 1.0
	case 1.0:
		c.Zoom = 0.5
		// При 0.5x не уменьшаем дальше
	}
}

// constrainToWorld ограничивает позицию камеры границами мира
func (c *Camera) constrainToWorld() {
	// Приблизительные размеры экрана (используются в будущем)
	_ = float32(1024) // Из main.go
	_ = float32(768)

	// Вычисляем границы для изометрической проекции
	// В изометрии диагональ карты больше её стороны
	worldDiagonal := c.worldWidth + c.worldHeight

	// Ограничиваем X
	minX := -worldDiagonal * TileWidth / 4
	maxX := worldDiagonal * TileWidth / 4
	if c.X < minX {
		c.X = minX
	}
	if c.X > maxX {
		c.X = maxX
	}

	// Ограничиваем Y
	minY := -worldDiagonal * TileHeight / 4
	maxY := worldDiagonal * TileHeight / 4
	if c.Y < minY {
		c.Y = minY
	}
	if c.Y > maxY {
		c.Y = maxY
	}
}

// GetPosition возвращает текущую позицию камеры
func (c *Camera) GetPosition() (x, y float32) {
	return c.X, c.Y
}

// GetZoom возвращает текущий масштаб камеры
func (c *Camera) GetZoom() float32 {
	return c.Zoom
}

// ScreenToWorld преобразует экранные координаты в мировые с учётом камеры
func (c *Camera) ScreenToWorld(screenX, screenY float32) (worldX, worldY float32) {
	// Применяем смещение камеры
	adjustedX := screenX + c.X
	adjustedY := screenY + c.Y

	// Применяем масштаб
	adjustedX /= c.Zoom
	adjustedY /= c.Zoom

	// Преобразуем в мировые координаты (обратная изометрическая проекция)
	worldX = (adjustedX/(TileWidth/2) + adjustedY/(TileHeight/2)) / 2
	worldY = (adjustedY/(TileHeight/2) - adjustedX/(TileWidth/2)) / 2

	return worldX, worldY
}

// WorldToScreen преобразует мировые координаты в экранные с учётом камеры
func (c *Camera) WorldToScreen(worldX, worldY float32) (screenX, screenY float32) {
	// Изометрическая проекция
	screenX = (worldX - worldY) * TileWidth / 2
	screenY = (worldX + worldY) * TileHeight / 2

	// Применяем масштаб
	screenX *= c.Zoom
	screenY *= c.Zoom

	// Применяем смещение камеры
	screenX -= c.X
	screenY -= c.Y

	return screenX, screenY
}

// IsPointVisible проверяет видна ли точка на экране
func (c *Camera) IsPointVisible(screenX, screenY float32, screenWidth, screenHeight int) bool {
	margin := float32(50) // Небольшой отступ для сглаживания

	return screenX >= -margin &&
		screenY >= -margin &&
		screenX <= float32(screenWidth)+margin &&
		screenY <= float32(screenHeight)+margin
}

// SetPosition устанавливает позицию камеры
func (c *Camera) SetPosition(x, y float32) {
	c.X = x
	c.Y = y
}

// SetZoom устанавливает масштаб камеры
func (c *Camera) SetZoom(zoom float32) {
	// Ограничиваем допустимые значения
	if zoom >= 0.25 && zoom <= 4.0 {
		c.Zoom = zoom
	}
}
