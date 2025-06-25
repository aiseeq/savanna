package gamestate

import (
	"fmt"
	"time"

	"github.com/aiseeq/savanna/internal/core"
)

// DebugOverlay предоставляет отладочную информацию
type DebugOverlay struct {
	enabled       bool
	lastFrameTime time.Time
	frameCount    int
	fps           float64

	// Отладочные флаги
	showEntityCounts bool
	showCameraInfo   bool
	showPerformance  bool
	showSystemInfo   bool
}

// NewDebugOverlay создает новый отладочный оверлей
func NewDebugOverlay() *DebugOverlay {
	return &DebugOverlay{
		enabled:          false,
		lastFrameTime:    time.Now(),
		showEntityCounts: true,
		showCameraInfo:   true,
		showPerformance:  true,
		showSystemInfo:   true,
	}
}

// SetEnabled включает/выключает отладочный оверлей
func (d *DebugOverlay) SetEnabled(enabled bool) {
	d.enabled = enabled
}

// IsEnabled возвращает статус включенности
func (d *DebugOverlay) IsEnabled() bool {
	return d.enabled
}

// Update обновляет отладочную информацию
func (d *DebugOverlay) Update() {
	if !d.enabled {
		return
	}

	// Обновляем FPS
	now := time.Now()
	deltaTime := now.Sub(d.lastFrameTime).Seconds()
	d.lastFrameTime = now
	d.frameCount++

	if deltaTime > 0 {
		// Сглаженный FPS
		d.fps = d.fps*0.9 + (1.0/deltaTime)*0.1
	}
}

// GenerateDebugInstructions создает отладочные инструкции рендеринга
func (d *DebugOverlay) GenerateDebugInstructions(gs *GameState) []DebugTextInstruction {
	if !d.enabled {
		return nil
	}

	var instructions []DebugTextInstruction
	y := 100.0 // Начинаем ниже основного UI

	if d.showPerformance {
		instructions = append(instructions, DebugTextInstruction{
			Text: fmt.Sprintf("FPS: %.1f", d.fps),
			X:    10,
			Y:    y,
		})
		y += 20

		instructions = append(instructions, DebugTextInstruction{
			Text: fmt.Sprintf("Frame: %d", d.frameCount),
			X:    10,
			Y:    y,
		})
		y += 20
	}

	if d.showCameraInfo {
		camera := gs.GetCameraState()
		instructions = append(instructions, DebugTextInstruction{
			Text: fmt.Sprintf("Camera: (%.1f, %.1f)", camera.X, camera.Y),
			X:    10,
			Y:    y,
		})
		y += 20

		if camera.IsScrolling {
			instructions = append(instructions, DebugTextInstruction{
				Text: "SCROLLING",
				X:    10,
				Y:    y,
			})
			y += 20
		}
	}

	if d.showEntityCounts {
		world := gs.GetWorld()

		rabbitCount := d.countEntitiesByType(world, core.TypeRabbit)
		wolfCount := d.countEntitiesByType(world, core.TypeWolf)
		corpseCount := d.countCorpses(world)

		instructions = append(instructions, DebugTextInstruction{
			Text: fmt.Sprintf("Rabbits: %d (Alive)", rabbitCount),
			X:    10,
			Y:    y,
		})
		y += 20

		instructions = append(instructions, DebugTextInstruction{
			Text: fmt.Sprintf("Wolves: %d (Alive)", wolfCount),
			X:    10,
			Y:    y,
		})
		y += 20

		instructions = append(instructions, DebugTextInstruction{
			Text: fmt.Sprintf("Corpses: %d", corpseCount),
			X:    10,
			Y:    y,
		})
		y += 20
	}

	if d.showSystemInfo {
		// Информация о первом зайце (для отладки)
		firstRabbit := d.getFirstEntityOfType(gs.GetWorld(), core.TypeRabbit)
		if firstRabbit != 0 {
			instructions, y = d.addEntityDebugInfo(instructions, gs.GetWorld(), firstRabbit, "First Rabbit", y)
		}

		// Информация о первом волке
		firstWolf := d.getFirstEntityOfType(gs.GetWorld(), core.TypeWolf)
		if firstWolf != 0 {
			instructions, _ = d.addEntityDebugInfo(instructions, gs.GetWorld(), firstWolf, "First Wolf", y)
		}
	}

	return instructions
}

// addEntityDebugInfo добавляет отладочную информацию о сущности
func (d *DebugOverlay) addEntityDebugInfo(
	instructions []DebugTextInstruction,
	world *core.World,
	entity core.EntityID,
	label string,
	startY float64,
) ([]DebugTextInstruction, float64) {
	y := startY

	// Позиция
	if pos, ok := world.GetPosition(entity); ok {
		instructions = append(instructions, DebugTextInstruction{
			Text: fmt.Sprintf("%s Pos: (%.1f, %.1f)", label, pos.X, pos.Y),
			X:    10,
			Y:    y,
		})
		y += 15
	}

	// Здоровье
	if health, ok := world.GetHealth(entity); ok {
		instructions = append(instructions, DebugTextInstruction{
			Text: fmt.Sprintf("%s Health: %d/%d", label, health.Current, health.Max),
			X:    10,
			Y:    y,
		})
		y += 15
	}

	// Сытость
	if satiation, ok := world.GetSatiation(entity); ok {
		instructions = append(instructions, DebugTextInstruction{
			Text: fmt.Sprintf("%s Satiation: %.1f%%", label, satiation.Value),
			X:    10,
			Y:    y,
		})
		y += 15
	}

	// Состояние анимации
	if anim, ok := world.GetAnimation(entity); ok {
		animName := d.getAnimationName(anim.CurrentAnim)
		instructions = append(instructions, DebugTextInstruction{
			Text: fmt.Sprintf("%s Anim: %s (F%d)", label, animName, anim.Frame),
			X:    10,
			Y:    y,
		})
		y += 15
	}

	// Скорость
	if vel, ok := world.GetVelocity(entity); ok {
		speed := fmt.Sprintf("%.1f", vel.X*vel.X+vel.Y*vel.Y) // Приблизительная скорость
		instructions = append(instructions, DebugTextInstruction{
			Text: fmt.Sprintf("%s Speed: %s", label, speed),
			X:    10,
			Y:    y,
		})
		y += 15
	}

	// Состояния
	states := d.getEntityStates(world, entity)
	if len(states) > 0 {
		instructions = append(instructions, DebugTextInstruction{
			Text: fmt.Sprintf("%s States: %s", label, states),
			X:    10,
			Y:    y,
		})
		y += 15
	}

	return instructions, y + 10 // Дополнительный отступ между сущностями
}

// countEntitiesByType подсчитывает сущности по типу
func (d *DebugOverlay) countEntitiesByType(world *core.World, animalType core.AnimalType) int {
	count := 0
	world.ForEachWith(core.MaskAnimalType, func(entity core.EntityID) {
		if entityType, ok := world.GetAnimalType(entity); ok && entityType == animalType {
			// Проверяем, что не труп
			if !world.HasComponent(entity, core.MaskCorpse) {
				count++
			}
		}
	})
	return count
}

// countCorpses подсчитывает трупы
func (d *DebugOverlay) countCorpses(world *core.World) int {
	count := 0
	world.ForEachWith(core.MaskCorpse, func(entity core.EntityID) {
		count++
	})
	return count
}

// getFirstEntityOfType возвращает первую сущность заданного типа
func (d *DebugOverlay) getFirstEntityOfType(world *core.World, animalType core.AnimalType) core.EntityID {
	var firstEntity core.EntityID = 0
	world.ForEachWith(core.MaskAnimalType, func(entity core.EntityID) {
		if firstEntity == 0 {
			if entityType, ok := world.GetAnimalType(entity); ok && entityType == animalType {
				// Проверяем, что не труп
				if !world.HasComponent(entity, core.MaskCorpse) {
					firstEntity = entity
				}
			}
		}
	})
	return firstEntity
}

// getAnimationName возвращает название анимации по ID
func (d *DebugOverlay) getAnimationName(animID int) string {
	switch animID {
	case 0:
		return "Idle"
	case 1:
		return "Walk"
	case 2:
		return "Run"
	case 3:
		return "Sleep"
	case 4:
		return "Attack"
	case 5:
		return "Eat"
	case 6:
		return "Dead"
	case 7:
		return "DeathDying"
	default:
		return fmt.Sprintf("Unknown(%d)", animID)
	}
}

// getEntityStates возвращает строку с состояниями сущности
func (d *DebugOverlay) getEntityStates(world *core.World, entity core.EntityID) string {
	var states []string

	if world.HasComponent(entity, core.MaskEatingState) {
		states = append(states, "Eating")
	}

	if world.HasComponent(entity, core.MaskAttackState) {
		states = append(states, "Attacking")
	}

	if world.HasComponent(entity, core.MaskCorpse) {
		states = append(states, "Corpse")
	}

	if world.HasComponent(entity, core.MaskDamageFlash) {
		states = append(states, "DamageFlash")
	}

	if len(states) == 0 {
		return "None"
	}

	result := ""
	for i, state := range states {
		if i > 0 {
			result += ", "
		}
		result += state
	}

	return result
}

// ToggleEntityCounts переключает отображение счетчиков сущностей
func (d *DebugOverlay) ToggleEntityCounts() {
	d.showEntityCounts = !d.showEntityCounts
}

// ToggleCameraInfo переключает отображение информации о камере
func (d *DebugOverlay) ToggleCameraInfo() {
	d.showCameraInfo = !d.showCameraInfo
}

// TogglePerformance переключает отображение информации о производительности
func (d *DebugOverlay) TogglePerformance() {
	d.showPerformance = !d.showPerformance
}

// ToggleSystemInfo переключает отображение системной информации
func (d *DebugOverlay) ToggleSystemInfo() {
	d.showSystemInfo = !d.showSystemInfo
}
