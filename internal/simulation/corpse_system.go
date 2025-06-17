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

		// Уменьшаем таймер разложения
		corpseData.DecayTimer -= deltaTime
		if corpseData.DecayTimer <= 0 {
			corpsesToRemove = append(corpsesToRemove, corpse)
		} else {
			world.SetCorpse(corpse, corpseData)
		}
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
		animalType, hasType := world.GetAnimalType(entity)

		if !hasHealth || !hasType || health.Current > 0 {
			return
		}

		// Если это заяц И он уже превращен в труп - оставляем
		if animalType == core.TypeRabbit && world.HasComponent(entity, core.MaskCorpse) {
			return
		}

		// Иначе удаляем мертвое животное
		deadAnimals = append(deadAnimals, entity)
	})

	// Удаляем мертвых животных (кроме трупов зайцев)
	for _, entity := range deadAnimals {
		world.DestroyEntity(entity)
	}
}

// createCorpse превращает мёртвое животное в труп (глобальная функция для других систем)
func createCorpse(world *core.World, animal core.EntityID) {
	// Удаляем компоненты живого животного
	world.RemoveVelocity(animal)
	world.RemoveHunger(animal)
	world.RemoveSpeed(animal)

	// Добавляем компонент трупа
	world.AddCorpse(animal, core.Corpse{
		NutritionalValue: CorpseNutritionalValue,
		MaxNutritional:   CorpseNutritionalValue,
		DecayTimer:       CorpseDecayTime,
	})

	// Устанавливаем анимацию смерти если есть анимационный компонент
	if world.HasComponent(animal, core.MaskAnimation) {
		anim, hasAnim := world.GetAnimation(animal)
		if hasAnim {
			anim.CurrentAnim = int(animation.AnimDeathDying)
			anim.Frame = 0
			anim.Timer = 0
			anim.Playing = true
			world.SetAnimation(animal, anim)
		}
	}
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
		if carrionData.DecayTimer <= 0 {
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
