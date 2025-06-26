package simulation

import (
	"github.com/aiseeq/savanna/internal/constants"
	"github.com/aiseeq/savanna/internal/core"
	"github.com/aiseeq/savanna/internal/vec2"
)

// EatingSystem отвечает ТОЛЬКО за поедание трупов хищниками (устраняет нарушение SRP)
//
// ВАЖНАЯ ЛОГИКА TargetType в EatingState:
// - TargetType = EatingTargetAnimal: поедание трупа/падали (обрабатывает EatingSystem)
// - TargetType = EatingTargetGrass: поедание травы травоядными (обрабатывает GrassEatingSystem)
//
// Эта система игнорирует EatingState с TargetType = EatingTargetGrass, оставляя их для GrassEatingSystem
type EatingSystem struct {
	previousFrames map[core.EntityID]int // Память предыдущих кадров для дискретного поедания
}

// NewEatingSystem создаёт новую систему поедания
func NewEatingSystem() *EatingSystem {
	return &EatingSystem{
		previousFrames: make(map[core.EntityID]int),
	}
}

// Update обновляет систему поедания (устраняет нарушение OCP)
func (es *EatingSystem) Update(world *core.World, deltaTime float32) {
	// Проходим по всем животным с поведением через компонент Behavior (универсально!)
	world.ForEachWith(core.MaskBehavior|core.MaskPosition, func(animal core.EntityID) {
		behavior, ok := world.GetBehavior(animal)
		if !ok {
			return
		}

		// Проверяем есть ли состояние поедания
		if eatingState, hasEating := world.GetEatingState(animal); hasEating {
			// ВАЖНО: Если TargetType = EatingTargetGrass, это поедание травы - НЕ трогаем (обрабатывает GrassEatingSystem)
			if eatingState.TargetType == core.EatingTargetGrass {
				return
			}
			// Животное уже ест труп/падаль - продолжаем процесс
			es.continueEating(world, animal, eatingState, 0)
		} else if behavior.Type == core.BehaviorPredator {
			// Животное не ест - ищем что поесть
			// Хищники едят трупы
			es.findCorpseToEat(world, animal)
			// УДАЛЕНО: BehaviorScavenger - не используется в игре
		}
	})
}

// findCorpseToEat ищет ближайший труп для поедания
func (es *EatingSystem) findCorpseToEat(world *core.World, predator core.EntityID) {
	// Хищник начинает есть только если голоден (используем AnimalConfig)
	satiation, hasSatiation := world.GetSatiation(predator)
	config, hasConfig := world.GetAnimalConfig(predator)
	if !hasSatiation || !hasConfig || satiation.Value >= config.SatiationThreshold {
		return // Сыт - не нужно есть (сытость выше порога)
	}

	predatorPos, hasPos := world.GetPosition(predator)
	if !hasPos {
		return
	}

	// Ищем ближайший труп
	var closestCorpse core.EntityID
	var closestDistance float32 = 999999.0

	world.ForEachWith(core.MaskCorpse|core.MaskPosition, func(corpse core.EntityID) {
		corpsePos, hasCorpsePos := world.GetPosition(corpse)
		if !hasCorpsePos {
			return
		}

		// ЭЛЕГАНТНАЯ МАТЕМАТИКА: расстояние через векторы
		predatorVec := vec2.Vec2{X: predatorPos.X, Y: predatorPos.Y}
		corpseVec := vec2.Vec2{X: corpsePos.X, Y: corpsePos.Y}
		distanceSquared := predatorVec.DistanceSquared(corpseVec)

		// ИСПРАВЛЕНИЕ: Конвертируем EatingRange из тайлов в пиксели
		eatingRangePixels := EatingRange * float32(constants.TileSizePixels)
		if distanceSquared < closestDistance && distanceSquared <= eatingRangePixels*eatingRangePixels {
			closestDistance = distanceSquared
			closestCorpse = corpse
		}
	})

	// Если нашли труп рядом, начинаем есть
	if closestCorpse != constants.NoTarget {
		world.AddEatingState(predator, core.EatingState{
			Target:          closestCorpse,
			TargetType:      core.EatingTargetAnimal, // Тип: поедание животного
			EatingProgress:  constants.InitialProgress,
			NutritionGained: constants.InitialNutrition,
		})
	}
}

// continueEating продолжает процесс поедания
func (es *EatingSystem) continueEating(
	world *core.World,
	predator core.EntityID,
	eatingState core.EatingState,
	_ float32,
) {
	// ИСПРАВЛЕНИЕ: Хищник ест до полного насыщения (99.9%), а НЕ до SatiationThreshold
	satiation, hasSatiation := world.GetSatiation(predator)
	if !hasSatiation {
		return
	}

	// Проверяем достигнуто ли полное насыщение (как у зайцев)
	const satietyThreshold = MaxSatiationLimit - constants.SatietyTolerance // 99.9%
	if satiation.Value >= satietyThreshold {
		// Хищник полностью наелся - превращаем недоеденный труп в падаль
		es.convertCorpseToCarrion(world, eatingState.Target, predator)
		world.RemoveEatingState(predator)
		return
	}

	// Проверяем что цель всё ещё существует и определяем тип еды
	if !world.IsAlive(eatingState.Target) {
		world.RemoveEatingState(predator)
		return
	}

	// Проверяем тип еды: труп или падаль
	var nutritionalValue float32
	var maxNutritional float32
	var isCorpse bool

	if corpse, hasCorpse := world.GetCorpse(eatingState.Target); hasCorpse {
		// Едим труп
		nutritionalValue = corpse.NutritionalValue
		maxNutritional = corpse.MaxNutritional
		isCorpse = true
	} else if carrion, hasCarrion := world.GetCarrion(eatingState.Target); hasCarrion {
		// Едим падаль
		nutritionalValue = carrion.NutritionalValue
		maxNutritional = carrion.MaxNutritional
		isCorpse = false
	} else {
		// Цель не является ни трупом, ни падалью
		world.RemoveEatingState(predator)
		return
	}

	if nutritionalValue <= 0 {
		world.RemoveEatingState(predator)
		delete(es.previousFrames, predator) // Очищаем память кадров
		return
	}

	// ИСПРАВЛЕНИЕ: Дискретное поедание как у зайцев - только при смене кадра анимации
	if !es.isEatingAnimationFrameComplete(world, predator) {
		return // Кадр анимации ещё не завершился - ждём
	}

	// Кадр завершён - обрабатываем "укус"
	es.processCorpseEatingTick(world, predator, eatingState, CorpseEatingParams{
		NutritionalValue: nutritionalValue,
		MaxNutritional:   maxNutritional,
		IsCorpse:         isCorpse,
	})
}

// isEatingAnimationFrameComplete проверяет произошла ли смена кадра анимации поедания
// Работает аналогично GrassEatingSystem - питательность даётся только при переходе на определённый кадр
func (es *EatingSystem) isEatingAnimationFrameComplete(world *core.World, entity core.EntityID) bool {
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

	// Используем отдельную память для предыдущих кадров
	prevFrame, exists := es.previousFrames[entity]
	currentFrame := anim.Frame

	// Если нет записи - это первый раз, инициализируем
	if !exists {
		es.previousFrames[entity] = currentFrame
		return false // Первый раз - не считается сменой
	}

	// ИСПРАВЛЕНИЕ: Питательность даётся при переходе на кадр 1 (как у зайцев)
	// Пользователь просил "после второго кадра", но думаю он имел в виду кадр 1 (второй кадр в массиве 0,1)
	frameChangedTo1 := (prevFrame == constants.AnimationFrameZero && currentFrame == constants.AnimationFrameOne)

	// Обновляем память предыдущего кадра
	es.previousFrames[entity] = currentFrame

	return frameChangedTo1
}

// CorpseEatingParams параметры поедания трупов
type CorpseEatingParams struct {
	NutritionalValue, MaxNutritional float32
	IsCorpse                         bool
}

// processCorpseEatingTick обрабатывает один "укус" трупа/падали
func (es *EatingSystem) processCorpseEatingTick(
	world *core.World,
	predator core.EntityID,
	eatingState core.EatingState,
	params CorpseEatingParams,
) {
	// Количество питательности съедаемое за один кадр анимации (как у зайцев - дискретно)
	nutritionPerTick := float32(CorpseNutritionPerTick)

	// Съедаем питательность
	nutritionEaten := nutritionPerTick
	if nutritionEaten > params.NutritionalValue {
		nutritionEaten = params.NutritionalValue
	}

	// Обновляем состояние еды
	params.NutritionalValue -= nutritionEaten
	if params.IsCorpse {
		// Обновляем труп
		corpse, _ := world.GetCorpse(eatingState.Target)
		corpse.NutritionalValue = params.NutritionalValue
		world.SetCorpse(eatingState.Target, corpse)
	} else {
		// Обновляем падаль
		carrion, _ := world.GetCarrion(eatingState.Target)
		carrion.NutritionalValue = params.NutritionalValue
		world.SetCarrion(eatingState.Target, carrion)
	}

	// Обновляем состояние поедания
	eatingState.NutritionGained += nutritionEaten
	eatingState.EatingProgress = (params.MaxNutritional - params.NutritionalValue) / params.MaxNutritional
	world.SetEatingState(predator, eatingState)

	// Восстанавливаем сытость животного
	es.feedPredator(world, predator, nutritionEaten*constants.NutritionToHungerRatio)

	// Если еда полностью съедена, убираем её
	if params.NutritionalValue <= 0 {
		world.RemoveEatingState(predator)
		delete(es.previousFrames, predator) // Очищаем память кадров
		world.DestroyEntity(eatingState.Target)
	}
}

// feedPredator восстанавливает сытость хищника
func (es *EatingSystem) feedPredator(world *core.World, predator core.EntityID, foodValue float32) {
	satiation, hasSatiation := world.GetSatiation(predator)
	if !hasSatiation {
		return
	}

	satiation.Value += foodValue
	if satiation.Value > MaxSatiationLimit {
		satiation.Value = MaxSatiationLimit
	}

	world.SetSatiation(predator, satiation)
}

// convertCorpseToCarrion превращает недоеденный труп в падаль
func (es *EatingSystem) convertCorpseToCarrion(world *core.World, corpseEntity, abandonedBy core.EntityID) {
	corpse, hasCorpse := world.GetCorpse(corpseEntity)
	if !hasCorpse || corpse.NutritionalValue <= 0 {
		// Труп полностью съеден или не существует - не превращаем в падаль
		return
	}

	// Создаём падаль на основе трупа
	carrion := core.Carrion{
		NutritionalValue: corpse.NutritionalValue,
		MaxNutritional:   corpse.MaxNutritional,
		DecayTimer:       corpse.DecayTimer,
		AbandonedBy:      abandonedBy,
	}

	// Удаляем компонент трупа и добавляем компонент падали
	world.RemoveCorpse(corpseEntity)
	world.AddCarrion(corpseEntity, carrion)
}
