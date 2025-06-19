package simulation

import (
	"github.com/aiseeq/savanna/internal/animation"
	"github.com/aiseeq/savanna/internal/core"
)

// CorpseSystem отвечает ТОЛЬКО за управление трупами (устраняет нарушение SRP)
type CorpseSystem struct{}

// NewCorpseSystem создаёт новую систему трупов
func NewCorpseSystem() *CorpseSystem {
	return &CorpseSystem{}
}

// Update обновляет систему трупов и падали
func (cs *CorpseSystem) Update(world *core.World, deltaTime float32) {
	// Обновляем разложение трупов
	cs.updateCorpseDecay(world, deltaTime)

	// Обновляем разложение падали
	cs.updateCarrionDecay(world, deltaTime)

	// Обрабатываем смерть животных
	cs.handleAnimalDeaths(world)
}

// updateCorpseDecay обновляет разложение трупов
func (cs *CorpseSystem) updateCorpseDecay(world *core.World, deltaTime float32) {
	var corpsesToRemove []core.EntityID

	world.ForEachWith(core.MaskCorpse, func(corpse core.EntityID) {
		corpseData, hasCorpse := world.GetCorpse(corpse)
		if !hasCorpse {
			return
		}

		// ИСПРАВЛЕНИЕ: Труп гниёт только если его НЕ едят активно
		// Ищем животных которые едят этот труп
		isBeingEaten := false
		world.ForEachWith(core.MaskEatingState, func(predator core.EntityID) {
			eatingState, hasEating := world.GetEatingState(predator)
			if hasEating && eatingState.Target == corpse {
				isBeingEaten = true
			}
		})

		// Уменьшаем таймер разложения только если труп НЕ поедается
		if !isBeingEaten {
			corpseData.DecayTimer -= deltaTime
		}

		// ИСПРАВЛЕНИЕ: Труп исчезает когда питательность = 0 ИЛИ таймер = 0
		if corpseData.NutritionalValue <= 0 || corpseData.DecayTimer <= 0 {
			corpsesToRemove = append(corpsesToRemove, corpse)
		} else {
			world.SetCorpse(corpse, corpseData)
		}
		// Если труп поедается - таймер НЕ уменьшается (консервация)
	})

	// Удаляем разложившиеся трупы
	for _, corpse := range corpsesToRemove {
		world.DestroyEntity(corpse)
	}
}

// handleAnimalDeaths обрабатывает смерть животных
// Зайцы, убитые хищниками, превращаются в трупы через createCorpse()
// Остальные мертвые животные (волки от голода и т.д.) просто удаляются
func (cs *CorpseSystem) handleAnimalDeaths(world *core.World) {
	var deadAnimals []core.EntityID

	world.ForEachWith(core.MaskHealth|core.MaskAnimalType, func(entity core.EntityID) {
		health, hasHealth := world.GetHealth(entity)
		_, hasType := world.GetAnimalType(entity)

		if !hasHealth || !hasType || health.Current > 0 {
			return
		}

		// ИСПРАВЛЕНИЕ: Если животное уже превращено в труп - оставляем
		// Теперь работает для ЛЮБЫХ животных, не только зайцев
		if world.HasComponent(entity, core.MaskCorpse) {
			return
		}

		// Иначе удаляем мертвое животное которое НЕ является трупом
		deadAnimals = append(deadAnimals, entity)
	})

	// Удаляем мертвых животных (кроме тех что уже стали трупами)
	for _, entity := range deadAnimals {
		world.DestroyEntity(entity)
	}
}

// createCorpse превращает мёртвое животное в труп (глобальная функция для других систем)
func createCorpse(world *core.World, animal core.EntityID) {
	CreateCorpseAndGetID(world, animal)
}

// CreateCorpseAndGetID превращает мёртвое животное в труп НА МЕСТЕ, сохраняя анимацию
func CreateCorpseAndGetID(world *core.World, animal core.EntityID) core.EntityID {
	// ИСПРАВЛЕНИЕ: НЕ уничтожаем животное, а превращаем его в труп на месте

	// Добавляем компонент трупа
	world.AddCorpse(animal, core.Corpse{
		NutritionalValue: CorpseNutritionalValue,
		MaxNutritional:   CorpseNutritionalValue,
		DecayTimer:       CorpseDecayTime,
	})

	// Переключаем анимацию на смерть и ОСТАНАВЛИВАЕМ на последнем кадре
	if world.HasComponent(animal, core.MaskAnimation) {
		world.SetAnimation(animal, core.Animation{
			CurrentAnim: int(animation.AnimDeathDying),
			Frame:       1,     // ПОСЛЕДНИЙ кадр анимации смерти (застывает)
			Timer:       999.0, // Большой таймер чтобы не переключалась
			Playing:     false, // НЕ играет - застыла на последнем кадре
			FacingRight: true,
		})
	}

	// ВАЖНО: Удаляем компоненты которые делают его "живым животным"
	// Но оставляем Position, Animation, AnimalType для правильного рендеринга
	if world.HasComponent(animal, core.MaskVelocity) {
		world.RemoveVelocity(animal)
	}
	if world.HasComponent(animal, core.MaskBehavior) {
		world.RemoveBehavior(animal)
	}
	if world.HasComponent(animal, core.MaskSize) {
		world.RemoveSize(animal)
	}
	if world.HasComponent(animal, core.MaskHunger) {
		world.RemoveHunger(animal)
	}
	// Оставляем Health=0 для индикации что это труп

	return animal // Возвращаем ТОТ ЖЕ EntityID - животное стало трупом
}

// updateCarrionDecay обновляет разложение падали
func (cs *CorpseSystem) updateCarrionDecay(world *core.World, deltaTime float32) {
	var carrionToRemove []core.EntityID

	world.ForEachWith(core.MaskCarrion, func(carrion core.EntityID) {
		carrionData, hasCarrion := world.GetCarrion(carrion)
		if !hasCarrion {
			return
		}

		// Уменьшаем таймер разложения
		carrionData.DecayTimer -= deltaTime

		// ИСПРАВЛЕНИЕ: Падаль теряет питательность во время естественного гниения
		// Скорость потери питательности: полная потеря за время DecayTime
		nutritionLossPerSecond := carrionData.MaxNutritional / CorpseDecayTime
		carrionData.NutritionalValue -= nutritionLossPerSecond * deltaTime

		// Не даем питательности стать отрицательной
		if carrionData.NutritionalValue < 0 {
			carrionData.NutritionalValue = 0
		}

		// ИСПРАВЛЕНИЕ: Падаль исчезает когда питательность = 0 ИЛИ таймер = 0
		if carrionData.NutritionalValue <= 0 || carrionData.DecayTimer <= 0 {
			carrionToRemove = append(carrionToRemove, carrion)
		} else {
			world.SetCarrion(carrion, carrionData)
		}
	})

	// Удаляем разложившуюся падаль
	for _, carrion := range carrionToRemove {
		world.DestroyEntity(carrion)
	}
}
