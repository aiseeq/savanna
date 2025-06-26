package simulation

import (
	"github.com/aiseeq/savanna/internal/constants"
	"github.com/aiseeq/savanna/internal/core"
)

// GrassEatingSystem отвечает за дискретное поедание травы травоядными по завершении анимации
// Работает аналогично EatingSystem для волков - питательность даётся только по завершении кадра
//
// ВАЖНАЯ ЛОГИКА TargetType в EatingState:
// - TargetType = EatingTargetGrass: поедание травы (обрабатывает GrassEatingSystem)
// - TargetType = EatingTargetAnimal: поедание животного (игнорируется, обрабатывает EatingSystem)
//
// Эта система работает ТОЛЬКО с EatingState где TargetType = EatingTargetGrass
type GrassEatingSystem struct {
	vegetation     core.VegetationProvider // Интерфейс для работы с растительностью (соблюдение DIP)
	previousFrames map[core.EntityID]int   // Память предыдущих кадров для обнаружения смены
}

// NewGrassEatingSystem создаёт новую систему поедания травы
func NewGrassEatingSystem(vegetation core.VegetationProvider) *GrassEatingSystem {
	return &GrassEatingSystem{
		vegetation:     vegetation,
		previousFrames: make(map[core.EntityID]int),
	}
}

// Update обновляет систему поедания травы
func (ges *GrassEatingSystem) Update(world *core.World, deltaTime float32) {
	if ges.vegetation == nil {
		return
	}

	// Ищем травоядных которые едят траву (устраняет нарушение OCP - было core.TypeRabbit)
	grassEatingMask := core.MaskEatingState | core.MaskPosition | core.MaskSatiation | core.MaskBehavior
	world.ForEachWith(grassEatingMask, func(entity core.EntityID) {
		// Проверяем что это травоядное через поведение, а НЕ через захардкоженный тип
		behavior, hasBehavior := world.GetBehavior(entity)
		if !hasBehavior || behavior.Type != core.BehaviorHerbivore {
			return
		}

		eatingState, hasEating := world.GetEatingState(entity)
		if !hasEating || eatingState.TargetType != core.EatingTargetGrass { // Работаем только с поеданием травы
			return
		}

		pos, hasPos := world.GetPosition(entity)
		if !hasPos {
			return
		}

		// Проверяем что рядом есть трава (ТИПОБЕЗОПАСНО)
		grassAmount := ges.vegetation.GetGrassAt(pos.X, pos.Y)
		if grassAmount < MinGrassAmountToFind {
			// Нет травы - убираем состояние поедания
			world.RemoveEatingState(entity)
			// Очищаем память кадров
			delete(ges.previousFrames, entity)
			return
		}

		// Проверяем завершился ли кадр анимации поедания
		frameComplete := ges.isEatingAnimationFrameComplete(world, entity)
		if frameComplete {
			// Кадр завершён - даём питательность и съедаем траву
			ges.processGrassEatingTick(world, entity, eatingState, pos)
		}
	})
}

// isEatingAnimationFrameComplete проверяет произошла ли смена кадра анимации поедания
func (ges *GrassEatingSystem) isEatingAnimationFrameComplete(world *core.World, entity core.EntityID) bool {
	anim, hasAnim := world.GetAnimation(entity)
	if !hasAnim {
		return false
	}

	// Проверяем что животное в анимации поедания
	if anim.CurrentAnim != int(constants.AnimEat) {
		return false
	}

	// Проверяем что анимация играет
	if !anim.Playing {
		return false
	}

	// ИСПРАВЛЕНИЕ: Используем отдельную память для предыдущих кадров
	prevFrame, exists := ges.previousFrames[entity]
	currentFrame := anim.Frame

	// Если нет записи - это первый раз, инициализируем
	if !exists {
		ges.previousFrames[entity] = currentFrame
		return false // Первый раз - не считается сменой
	}

	// ИСПРАВЛЕНИЕ: Питательность даётся ТОЛЬКО при переходе на кадр 1!
	// Как у волков - урон только на кадре 1 атаки
	frameChangedTo1 := (prevFrame == 0 && currentFrame == 1)

	// Обновляем память предыдущего кадра
	ges.previousFrames[entity] = currentFrame

	return frameChangedTo1
}

// processGrassEatingTick обрабатывает один "укус" травы
func (ges *GrassEatingSystem) processGrassEatingTick(
	world *core.World, entity core.EntityID, eatingState core.EatingState, pos core.Position,
) {
	// Количество травы съедаемое за один кадр анимации (как у волка - дискретно)
	grassPerTick := float32(GrassPerEatingTick) // 1.0 единица травы за кадр анимации

	// Съедаем траву (ТИПОБЕЗОПАСНО)
	consumedGrass := ges.vegetation.ConsumeGrassAt(pos.X, pos.Y, grassPerTick)
	if consumedGrass <= 0 {
		// Нет травы - заканчиваем поедание
		world.RemoveEatingState(entity)
		// Очищаем память кадров
		delete(ges.previousFrames, entity)
		return
	}

	// Обновляем состояние поедания
	eatingState.EatingProgress += consumedGrass / GrassEatingProgressDivisor
	eatingState.NutritionGained += consumedGrass
	world.SetEatingState(entity, eatingState)

	// Восстанавливаем голод пропорционально съеденной траве
	// Используем константу питательности травы
	hungerToRestore := consumedGrass * GrassNutritionValue

	hunger, hasHunger := world.GetSatiation(entity)
	if hasHunger {
		ges.feedAnimal(world, entity, hungerToRestore)

		// Обновляем голод для проверки
		hunger, _ = world.GetSatiation(entity)

		// ИСПРАВЛЕНИЕ: Заяц прекращает есть когда почти полностью сыт (допуск для float32)
		const satietyThreshold = MaxSatiationLimit - constants.SatietyTolerance // Используем константы из game_balance.go
		if hunger.Value >= satietyThreshold {
			// Заяц полностью наелся - заканчиваем поедание
			world.RemoveEatingState(entity)
			// Очищаем память кадров
			delete(ges.previousFrames, entity)
		}
	}
}

// feedAnimal восстанавливает голод животного
func (ges *GrassEatingSystem) feedAnimal(world *core.World, entity core.EntityID, foodValue float32) {
	hunger, hasHunger := world.GetSatiation(entity)
	if !hasHunger {
		return
	}

	hunger.Value += foodValue
	if hunger.Value > MaxSatiationLimit {
		hunger.Value = MaxSatiationLimit
	}

	world.SetSatiation(entity, hunger)
}
