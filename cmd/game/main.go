package main

import (
	"fmt"
	"image/color"
	"log"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
)

// Game структура для GUI версии симулятора экосистемы саванны
type Game struct {
	// world *core.World      // Будет добавлено позже
	// renderer *rendering.Renderer
	// camera *rendering.Camera
}

// Update обновляет логику игры (вызывается 60 раз в секунду)
func (g *Game) Update() error {
	// Выход по ESC
	if inpututil.IsKeyJustPressed(ebiten.KeyEscape) {
		return fmt.Errorf("игра завершена пользователем")
	}

	// TODO: Обновление симуляции
	return nil
}

// Draw отрисовывает кадр
func (g *Game) Draw(screen *ebiten.Image) {
	// Заливаем экран темно-зеленым цветом (саванна)
	screen.Fill(color.RGBA{34, 139, 34, 255}) // Forest Green

	// Отображаем информацию
	msg := `Savanna Ecosystem Simulator
GUI версия в разработке...

Управление:
- ESC: выход
- WASD: движение камеры (позже)
- Колесо мыши: масштаб (позже)

Версия: MVP v0.1
Статус: Этап 0 завершен`

	ebitenutil.DebugPrint(screen, msg)
}

// Layout возвращает размер экрана
func (g *Game) Layout(outsideWidth, outsideHeight int) (screenWidth, screenHeight int) {
	return 1024, 768
}

func main() {
	fmt.Println("Запуск GUI версии симулятора экосистемы саванны...")

	game := &Game{}

	ebiten.SetWindowSize(1024, 768)
	ebiten.SetWindowTitle("Savanna Ecosystem Simulator")
	ebiten.SetWindowResizable(true)

	if err := ebiten.RunGame(game); err != nil {
		log.Fatal(err)
	}
}
