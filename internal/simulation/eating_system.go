package simulation

import (
	"github.com/aiseeq/savanna/internal/animation"
	"github.com/aiseeq/savanna/internal/core"
)

// EatingSystem отвечает ТОЛЬКО за поедание трупов хищниками (устраняет нарушение SRP)
//
// ВАЖНАЯ ЛОГИКА Target в EatingState:
// - Target = EntityID животного: поедание трупа/падали (обрабатывает EatingSystem)
// - Target = 0: поедание травы травоядными (обрабатывает GrassEatingSystem)
//
// Эта система игнорирует EatingState с Target = 0, оставляя их для GrassEatingSystem
type EatingSystem struct{
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
			// ВАЖНО: Если Target = 0, это поедание травы - НЕ трогаем (обрабатывает GrassEatingSystem)
			if eatingState.Target == GrassEatingTarget {
				return
			}
			// Животное уже ест труп/падаль - продолжаем процесс
			es.continueEating(world, animal, eatingState, deltaTime)
		} else {
			// Животное не ест - ищем что поесть
			switch behavior.Type {
			case core.BehaviorPredator:
				// Хищники едят трупы
				es.findCorpseToEat(world, animal)
				// УДАЛЕНО: BehaviorScavenger - не используется в игре
			}
		}
	})
}

// findCorpseToEat ищет ближайший труп для поедания
func (es *EatingSystem) findCorpseToEat(world *core.World, predator core.EntityID) {
	// Хищник начинает есть только если голоден (используем AnimalConfig)
	hunger, hasHunger := world.GetHunger(predator)
	config, hasConfig := world.GetAnimalConfig(predator)
	if !hasHunger || !hasConfig || hunger.Value >= config.HungerThreshold {
		return
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

		distance := (predatorPos.X-corpsePos.X)*(predatorPos.X-corpsePos.X) + (predatorPos.Y-corpsePos.Y)*(predatorPos.Y-corpsePos.Y)
		if distance < closestDistance && distance <= EatingRange*EatingRange {
			closestDistance = distance
			closestCorpse = corpse
		}
	})

	// Если нашли труп рядом, начинаем есть
	if closestCorpse != 0 {
		world.AddEatingState(predator, core.EatingState{
			Target:          closestCorpse,
			EatingProgress:  0.0,
			NutritionGained: 0.0,
		})
	}
}

// continueEating продолжает процесс поедания
func (es *EatingSystem) continueEating(world *core.World, predator core.EntityID, eatingState core.EatingState, deltaTime float32) {
	// ИСПРАВЛЕНИЕ: Хищник ест до полного насыщения (99.9%), а НЕ до HungerThreshold
	hunger, hasHunger := world.GetHunger(predator)
	if !hasHunger {
		return
	}
	
	// Проверяем достигнуто ли полное насыщение (как у зайцев)
	const satietyThreshold = MaxHungerLimit - SatietyTolerance // 99.9%
	if hunger.Value >= satietyThreshold {
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
	es.processCorpseEatingTick(world, predator, eatingState, nutritionalValue, maxNutritional, isCorpse)
}

// isEatingAnimationFrameComplete проверяет произошла ли смена кадра анимации поедания
// Работает аналогично GrassEatingSystem - питательность даётся только при переходе на определённый кадр
func (es *EatingSystem) isEatingAnimationFrameComplete(world *core.World, entity core.EntityID) bool {
	anim, hasAnim := world.GetAnimation(entity)
	if !hasAnim {
		return false
	}

	// Проверяем что животное в анимации поедания
	if anim.CurrentAnim != int(animation.AnimEat) {
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
	frameChangedTo1 := (prevFrame == 0 && currentFrame == 1)

	// Обновляем память предыдущего кадра
	es.previousFrames[entity] = currentFrame

	return frameChangedTo1
}

// processCorpseEatingTick обрабатывает один "укус" трупа/падали
func (es *EatingSystem) processCorpseEatingTick(world *core.World, predator core.EntityID, eatingState core.EatingState, nutritionalValue, maxNutritional float32, isCorpse bool) {
	// Количество питательности съедаемое за один кадр анимации (как у зайцев - дискретно)
	nutritionPerTick := float32(CorpseNutritionPerTick)

	// Съедаем питательность
	nutritionEaten := nutritionPerTick
	if nutritionEaten > nutritionalValue {
		nutritionEaten = nutritionalValue
	}

	// Обновляем состояние еды
	nutritionalValue -= nutritionEaten
	if isCorpse {
		// Обновляем труп
		corpse, _ := world.GetCorpse(eatingState.Target)
		corpse.NutritionalValue = nutritionalValue
		world.SetCorpse(eatingState.Target, corpse)
	} else {
		// Обновляем падаль
		carrion, _ := world.GetCarrion(eatingState.Target)
		carrion.NutritionalValue = nutritionalValue
		world.SetCarrion(eatingState.Target, carrion)
	}

	// Обновляем состояние поедания
	eatingState.NutritionGained += nutritionEaten
	eatingState.EatingProgress = (maxNutritional - nutritionalValue) / maxNutritional
	world.SetEatingState(predator, eatingState)

	// Восстанавливаем голод животного
	es.feedPredator(world, predator, nutritionEaten*NutritionToHungerRatio)

	// Если еда полностью съедена, убираем её
	if nutritionalValue <= 0 {
		world.RemoveEatingState(predator)
		delete(es.previousFrames, predator) // Очищаем память кадров
		world.DestroyEntity(eatingState.Target)
	}
}

// feedPredator восстанавливает голод хищника
func (es *EatingSystem) feedPredator(world *core.World, predator core.EntityID, foodValue float32) {
	hunger, hasHunger := world.GetHunger(predator)
	if !hasHunger {
		return
	}

	hunger.Value += foodValue
	if hunger.Value > MaxHungerLimit {
		hunger.Value = MaxHungerLimit
	}

	world.SetHunger(predator, hunger)
}

// convertCorpseToCarrion превращает недоеденный труп в падаль
func (es *EatingSystem) convertCorpseToCarrion(world *core.World, corpseEntity core.EntityID, abandonedBy core.EntityID) {
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

// findCarrionToEat ищет ближайшую падаль для поедания
func (es *EatingSystem) findCarrionToEat(world *core.World, scavenger core.EntityID) {
	// Падальщик начинает есть только если голоден (используем AnimalConfig)
	hunger, hasHunger := world.GetHunger(scavenger)
	config, hasConfig := world.GetAnimalConfig(scavenger)
	if !hasHunger || !hasConfig || hunger.Value >= config.HungerThreshold {
		return
	}

	scavengerPos, hasPos := world.GetPosition(scavenger)
	if !hasPos {
		return
	}

	// Ищем ближайшую падаль
	var closestCarrion core.EntityID
	var closestDistance float32 = 999999.0

	world.ForEachWith(core.MaskCarrion|core.MaskPosition, func(carrion core.EntityID) {
		carrionPos, hasCarrionPos := world.GetPosition(carrion)
		if !hasCarrionPos {
			return
		}

		distance := (scavengerPos.X-carrionPos.X)*(scavengerPos.X-carrionPos.X) + (scavengerPos.Y-carrionPos.Y)*(scavengerPos.Y-carrionPos.Y)
		if distance < closestDistance && distance <= EatingRange*EatingRange {
			closestDistance = distance
			closestCarrion = carrion
		}
	})

	// Если нашли падаль рядом, начинаем есть
	if closestCarrion != 0 {
		world.AddEatingState(scavenger, core.EatingState{
			Target:          closestCarrion,
			EatingProgress:  0.0,
			NutritionGained: 0.0,
		})
	}
}
