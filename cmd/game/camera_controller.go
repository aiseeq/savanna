package main

import (
	"github.com/aiseeq/savanna/internal/constants"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
)

// Camera простая камера для просмотра мира
type Camera struct {
	X, Y float32
	Zoom float32
}

// CameraController управляет камерой и её движением
// Соблюдает SRP - единственная ответственность: управление камерой
type CameraController struct {
	camera Camera

	// Состояние перетаскивания карты
	isDragging bool
	lastMouseX int
	lastMouseY int
}

// NewCameraController создаёт новый контроллер камеры
func NewCameraController() *CameraController {
	return &CameraController{
		camera: Camera{
			X:    constants.DefaultCameraX, // Начальная позиция
			Y:    constants.DefaultCameraY,
			Zoom: 1.0,
		},
		isDragging: false,
		lastMouseX: 0,
		lastMouseY: 0,
	}
}

// GetCamera возвращает текущее состояние камеры
func (cc *CameraController) GetCamera() Camera {
	return cc.camera
}

// Update обновляет состояние камеры на основе ввода
func (cc *CameraController) Update() {
	cc.handleMovement()
	cc.handleZoom()
	cc.handleDragging()
}

// handleMovement обрабатывает движение камеры клавишами
func (cc *CameraController) handleMovement() {
	moveSpeed := float32(constants.CameraMoveSpeed / cc.camera.Zoom)

	if ebiten.IsKeyPressed(ebiten.KeyW) || ebiten.IsKeyPressed(ebiten.KeyArrowUp) {
		cc.camera.Y -= moveSpeed
	}
	if ebiten.IsKeyPressed(ebiten.KeyS) || ebiten.IsKeyPressed(ebiten.KeyArrowDown) {
		cc.camera.Y += moveSpeed
	}
	if ebiten.IsKeyPressed(ebiten.KeyA) || ebiten.IsKeyPressed(ebiten.KeyArrowLeft) {
		cc.camera.X -= moveSpeed
	}
	if ebiten.IsKeyPressed(ebiten.KeyD) || ebiten.IsKeyPressed(ebiten.KeyArrowRight) {
		cc.camera.X += moveSpeed
	}
}

// handleZoom обрабатывает масштабирование камеры
func (cc *CameraController) handleZoom() {
	_, scrollY := ebiten.Wheel()
	if scrollY != 0 {
		zoomFactor := float32(constants.CameraZoomFactor)
		if scrollY > 0 {
			cc.camera.Zoom *= zoomFactor
		} else {
			cc.camera.Zoom /= zoomFactor
		}

		// Ограничиваем зум
		if cc.camera.Zoom < constants.MinCameraZoom {
			cc.camera.Zoom = constants.MinCameraZoom
		}
		if cc.camera.Zoom > constants.MaxCameraZoom {
			cc.camera.Zoom = constants.MaxCameraZoom
		}
	}
}

// handleDragging обрабатывает перетаскивание карты мышью
func (cc *CameraController) handleDragging() {
	// ИСПРАВЛЕНО: используем правую кнопку мыши для перетаскивания камеры
	// Левая кнопка должна использоваться для выбора и взаимодействия с объектами
	if inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonRight) {
		cc.isDragging = true
		cc.lastMouseX, cc.lastMouseY = ebiten.CursorPosition()
	}

	if inpututil.IsMouseButtonJustReleased(ebiten.MouseButtonRight) {
		cc.isDragging = false
	}

	if cc.isDragging {
		mouseX, mouseY := ebiten.CursorPosition()
		deltaX := float32(mouseX - cc.lastMouseX)
		deltaY := float32(mouseY - cc.lastMouseY)

		// Перемещаем камеру в обратном направлении для интуитивного перетаскивания
		cc.camera.X -= deltaX / cc.camera.Zoom
		cc.camera.Y -= deltaY / cc.camera.Zoom

		cc.lastMouseX = mouseX
		cc.lastMouseY = mouseY
	}
}
