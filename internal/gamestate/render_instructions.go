package gamestate

import (
	"fmt"
	"image/color"

	"github.com/aiseeq/savanna/internal/constants"
	"github.com/aiseeq/savanna/internal/core"
)

// RenderInstruction базовая инструкция для рендеринга
type RenderInstruction interface {
	GetType() RenderInstructionType
}

// RenderInstructionType тип инструкции рендеринга
type RenderInstructionType int

const (
	InstructionSprite RenderInstructionType = iota
	InstructionTerrain
	InstructionUI
	InstructionHealthBar
	InstructionDebugText
)

// SpriteInstruction инструкция для рендеринга спрайта
type SpriteInstruction struct {
	EntityID    core.EntityID
	AnimalType  core.AnimalType
	X, Y        float64
	Frame       int
	AnimType    int
	Tint        color.RGBA
	FacingRight bool
	Scale       float64
}

func (s SpriteInstruction) GetType() RenderInstructionType {
	return InstructionSprite
}

// TerrainInstruction инструкция для рендеринга тайла terrain
type TerrainInstruction struct {
	TileX, TileY int
	TileType     int
	GrassAmount  float32
}

func (t TerrainInstruction) GetType() RenderInstructionType {
	return InstructionTerrain
}

// UIInstruction инструкция для рендеринга UI элемента
type UIInstruction struct {
	Text     string
	X, Y     float64
	Color    color.RGBA
	FontSize int
}

func (u UIInstruction) GetType() RenderInstructionType {
	return InstructionUI
}

// HealthBarInstruction инструкция для рендеринга полоски здоровья
type HealthBarInstruction struct {
	EntityID  core.EntityID
	X, Y      float64
	Width     float64
	Health    float32
	MaxHealth float32
	Visible   bool
}

func (h HealthBarInstruction) GetType() RenderInstructionType {
	return InstructionHealthBar
}

// DebugTextInstruction инструкция для отладочного текста
type DebugTextInstruction struct {
	Text string
	X, Y float64
}

func (d DebugTextInstruction) GetType() RenderInstructionType {
	return InstructionDebugText
}

// RenderInstructionSet набор всех инструкций для одного кадра
type RenderInstructionSet struct {
	Sprites    []SpriteInstruction
	Terrain    []TerrainInstruction
	UI         []UIInstruction
	HealthBars []HealthBarInstruction
	DebugTexts []DebugTextInstruction
	CameraX    float64
	CameraY    float64
}

// GenerateRenderInstructions создает инструкции рендеринга из текущего состояния игры
func (gs *GameState) GenerateRenderInstructions() RenderInstructionSet {
	instructions := RenderInstructionSet{
		CameraX: gs.camera.X,
		CameraY: gs.camera.Y,
	}

	// Генерируем инструкции для спрайтов животных
	gs.generateSpriteInstructions(&instructions)

	// Генерируем инструкции для terrain
	gs.generateTerrainInstructions(&instructions)

	// Генерируем UI инструкции
	gs.generateUIInstructions(&instructions)

	// Генерируем инструкции для полосок здоровья
	gs.generateHealthBarInstructions(&instructions)

	// ИСПРАВЛЕНИЕ: Генерируем инструкции для отображения сытости над животными
	gs.generateSatiationDisplayInstructions(&instructions)

	return instructions
}

// GenerateRenderInstructionsWithDebug создает инструкции рендеринга включая отладочный оверлей
func (gs *GameState) GenerateRenderInstructionsWithDebug(debugOverlay *DebugOverlay) RenderInstructionSet {
	instructions := gs.GenerateRenderInstructions()

	// Добавляем отладочные инструкции
	if debugOverlay != nil && debugOverlay.IsEnabled() {
		debugInstructions := debugOverlay.GenerateDebugInstructions(gs)
		instructions.DebugTexts = append(instructions.DebugTexts, debugInstructions...)
	}

	return instructions
}

// generateSpriteInstructions генерирует инструкции для спрайтов животных
func (gs *GameState) generateSpriteInstructions(instructions *RenderInstructionSet) {
	// Обходим всех животных
	gs.world.ForEachWith(core.MaskAnimalType|core.MaskPosition, func(entity core.EntityID) {
		animalType, hasType := gs.world.GetAnimalType(entity)
		if !hasType {
			return
		}

		pos, hasPos := gs.world.GetPosition(entity)
		if !hasPos {
			return
		}

		// Определяем анимацию и кадр
		animType, frame := gs.getEntityAnimationState(entity)

		// Определяем направление
		facingRight := gs.getEntityDirection(entity)

		// Определяем tint (например, для эффекта урона)
		tint := gs.getEntityTint(entity)

		instruction := SpriteInstruction{
			EntityID:    entity,
			AnimalType:  animalType,
			X:           float64(pos.X),
			Y:           float64(pos.Y),
			Frame:       frame,
			AnimType:    animType,
			Tint:        tint,
			FacingRight: facingRight,
			Scale:       1.0,
		}

		instructions.Sprites = append(instructions.Sprites, instruction)
	})
}

// generateTerrainInstructions генерирует инструкции для terrain (упрощенная версия)
func (gs *GameState) generateTerrainInstructions(instructions *RenderInstructionSet) {
	// Для демонстрации - генерируем простые тайлы
	// В реальности здесь должна быть логика terrain системы
	worldSize := int(gs.config.WorldWidth / 32) // размер в тайлах
	for x := 0; x < worldSize; x++ {
		for y := 0; y < worldSize; y++ {
			instruction := TerrainInstruction{
				TileX:       x,
				TileY:       y,
				TileType:    1, // grass
				GrassAmount: 100.0,
			}
			instructions.Terrain = append(instructions.Terrain, instruction)
		}
	}
}

// generateUIInstructions генерирует UI инструкции
func (gs *GameState) generateUIInstructions(instructions *RenderInstructionSet) {
	// Статистика игры
	entityCount := gs.world.GetEntityCount()
	rabbitCount := gs.countAnimalsByType(core.TypeRabbit)
	wolfCount := gs.countAnimalsByType(core.TypeWolf)

	instructions.UI = append(instructions.UI, UIInstruction{
		Text:     fmt.Sprintf("Entities: %d", entityCount),
		X:        10,
		Y:        10,
		Color:    color.RGBA{255, 255, 255, 255},
		FontSize: 16,
	})

	instructions.UI = append(instructions.UI, UIInstruction{
		Text:     fmt.Sprintf("Rabbits: %d", rabbitCount),
		X:        10,
		Y:        30,
		Color:    color.RGBA{255, 255, 255, 255},
		FontSize: 16,
	})

	instructions.UI = append(instructions.UI, UIInstruction{
		Text:     fmt.Sprintf("Wolves: %d", wolfCount),
		X:        10,
		Y:        50,
		Color:    color.RGBA{255, 255, 255, 255},
		FontSize: 16,
	})

	// Hunger первого зайца (для отладки)
	firstRabbit := gs.getFirstAnimalOfType(core.TypeRabbit)
	if firstRabbit != 0 {
		if hunger, ok := gs.world.GetSatiation(firstRabbit); ok {
			instructions.UI = append(instructions.UI, UIInstruction{
				Text:     fmt.Sprintf("First Rabbit Hunger: %.1f%%", hunger.Value),
				X:        10,
				Y:        70,
				Color:    color.RGBA{255, 255, 255, 255},
				FontSize: 16,
			})
		}
	}
}

// generateHealthBarInstructions генерирует инструкции для полосок здоровья
func (gs *GameState) generateHealthBarInstructions(instructions *RenderInstructionSet) {
	gs.world.ForEachWith(core.MaskAnimalType|core.MaskPosition|core.MaskHealth, func(entity core.EntityID) {
		pos, _ := gs.world.GetPosition(entity)
		health, _ := gs.world.GetHealth(entity)
		animalType, _ := gs.world.GetAnimalType(entity)

		// ИСПРАВЛЕНИЕ: Размеры UI зависят от размера СПРАЙТА, не от физического размера
		var healthBarWidth float64 = 40
		var healthBarOffset float64 = 30

		// Настройка под размер спрайта конкретного животного
		switch animalType {
		case core.TypeRabbit:
			healthBarWidth = 32
			healthBarOffset = 20
		case core.TypeWolf:
			healthBarWidth = 40
			healthBarOffset = 30
		}

		// ИСПРАВЛЕНИЕ: Полоска здоровья видна всегда для мониторинга
		visible := true

		instruction := HealthBarInstruction{
			EntityID:  entity,
			X:         float64(pos.X),
			Y:         float64(pos.Y) - healthBarOffset, // над спрайтом животного
			Width:     healthBarWidth,
			Health:    float32(health.Current),
			MaxHealth: float32(health.Max),
			Visible:   visible,
		}

		instructions.HealthBars = append(instructions.HealthBars, instruction)
	})
}

// Вспомогательные методы

func (gs *GameState) getEntityAnimationState(entity core.EntityID) (animType int, frame int) {
	if anim, ok := gs.world.GetAnimation(entity); ok {
		return anim.CurrentAnim, anim.Frame
	}
	return int(constants.AnimIdle), 0
}

func (gs *GameState) getEntityDirection(entity core.EntityID) bool {
	if anim, ok := gs.world.GetAnimation(entity); ok {
		return anim.FacingRight
	}
	return true
}

func (gs *GameState) getEntityTint(entity core.EntityID) color.RGBA {
	// Проверяем, есть ли эффект урона
	if gs.world.HasComponent(entity, core.MaskDamageFlash) {
		// Красный tint для эффекта урона
		return color.RGBA{255, 150, 150, 255}
	}

	// Проверяем, мертвое ли животное
	if gs.world.HasComponent(entity, core.MaskCorpse) {
		// Серый tint для трупов
		return color.RGBA{128, 128, 128, 255}
	}

	return color.RGBA{255, 255, 255, 255} // Обычный цвет
}

func (gs *GameState) countAnimalsByType(animalType core.AnimalType) int {
	count := 0
	gs.world.ForEachWith(core.MaskAnimalType, func(entity core.EntityID) {
		if entityType, ok := gs.world.GetAnimalType(entity); ok && entityType == animalType {
			count++
		}
	})
	return count
}

func (gs *GameState) getFirstAnimalOfType(animalType core.AnimalType) core.EntityID {
	var firstEntity core.EntityID = 0
	gs.world.ForEachWith(core.MaskAnimalType, func(entity core.EntityID) {
		if firstEntity == 0 {
			if entityType, ok := gs.world.GetAnimalType(entity); ok && entityType == animalType {
				firstEntity = entity
			}
		}
	})
	return firstEntity
}

// generateSatiationDisplayInstructions генерирует инструкции для отображения сытости над животными
func (gs *GameState) generateSatiationDisplayInstructions(instructions *RenderInstructionSet) {
	gs.world.ForEachWith(core.MaskAnimalType|core.MaskPosition|core.MaskSatiation, func(entity core.EntityID) {
		pos, _ := gs.world.GetPosition(entity)
		satiation, _ := gs.world.GetSatiation(entity)
		animalType, _ := gs.world.GetAnimalType(entity)

		// ИСПРАВЛЕНИЕ: Позиция текста зависит от размера СПРАЙТА, не от физического размера
		var satiationOffset float64 = 50
		var textOffsetX float64 = 10

		// Настройка под размер спрайта конкретного животного
		switch animalType {
		case core.TypeRabbit:
			satiationOffset = 35 // Выше хелсбара зайца
			textOffsetX = 8
		case core.TypeWolf:
			satiationOffset = 50 // Выше хелсбара волка
			textOffsetX = 10
		}

		// Форматируем текст сытости
		satiationText := fmt.Sprintf("%.0f%%", satiation.Value)

		// Определяем цвет по уровню сытости
		var textColor color.RGBA
		if satiation.Value < 30.0 {
			textColor = color.RGBA{255, 100, 100, 255} // Красный при низкой сытости
		} else if satiation.Value < 60.0 {
			textColor = color.RGBA{255, 255, 100, 255} // Жёлтый при средней сытости
		} else {
			textColor = color.RGBA{100, 255, 100, 255} // Зелёный при высокой сытости
		}

		// Добавляем UI инструкцию для отображения сытости над животным
		instruction := UIInstruction{
			Text:     satiationText,
			X:        float64(pos.X) - textOffsetX,     // Центровка относительно спрайта
			Y:        float64(pos.Y) - satiationOffset, // Выше хелсбара
			Color:    textColor,
			FontSize: 12, // Чуть меньший шрифт для компактности
		}

		instructions.UI = append(instructions.UI, instruction)
	})
}
