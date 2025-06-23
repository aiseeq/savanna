package animation

import "github.com/aiseeq/savanna/internal/core"

// AnimationManager менеджер анимаций для всех животных (устраняет дублирование логики обновления)
type AnimationManager struct {
	resolver *AnimationResolver
	systems  map[core.AnimalType]*AnimationSystem
}

// NewAnimationManager создаёт новый менеджер анимаций
func NewAnimationManager(wolfSystem, rabbitSystem *AnimationSystem) *AnimationManager {
	return &AnimationManager{
		resolver: NewAnimationResolver(),
		systems: map[core.AnimalType]*AnimationSystem{
			core.TypeWolf:   wolfSystem,
			core.TypeRabbit: rabbitSystem,
		},
	}
}

// Update реализует интерфейс core.System для системы анимаций
func (am *AnimationManager) Update(world *core.World, deltaTime float32) {
	am.UpdateAllAnimations(world, deltaTime)
}

// UpdateAllAnimations обновляет анимации всех животных
func (am *AnimationManager) UpdateAllAnimations(world *core.World, deltaTime float32) {
	world.ForEachWith(core.MaskAnimalType|core.MaskAnimation, func(entity core.EntityID) {
		am.updateAnimalAnimation(world, entity, deltaTime)
	})
}

// updateAnimalAnimation обновляет анимацию одного животного (общая логика)
func (am *AnimationManager) updateAnimalAnimation(world *core.World, entity core.EntityID, deltaTime float32) {
	animalType, ok := world.GetAnimalType(entity)
	if !ok {
		return
	}

	anim, hasAnim := world.GetAnimation(entity)
	if !hasAnim {
		return
	}

	// Определяем какая анимация должна играть
	newAnimType := am.resolver.ResolveAnimalAnimationType(world, entity, animalType)
	animSystem, hasSystem := am.systems[animalType]
	if !hasSystem {
		return
	}

	// КРИТИЧЕСКОЕ ПРАВИЛО: НЕ прерываем анимацию ATTACK пока она играет!
	if anim.CurrentAnim != int(newAnimType) {
		if anim.CurrentAnim == int(AnimAttack) && anim.Playing {
			// Анимация атаки должна доиграться до конца
		} else {
			// Переключаем на новую анимацию
			anim.CurrentAnim = int(newAnimType)
			anim.Frame = 0
			anim.Timer = 0
			anim.Playing = true
			world.SetAnimation(entity, anim)
		}
	}

	// Обновляем направление взгляда на основе скорости
	am.updateFacingDirection(world, entity, &anim)

	// Обновляем анимацию через систему
	animComponent := AnimationComponent{
		CurrentAnim: AnimationType(anim.CurrentAnim),
		Frame:       anim.Frame,
		Timer:       anim.Timer,
		Playing:     anim.Playing,
		FacingRight: anim.FacingRight,
	}

	animSystem.Update(&animComponent, deltaTime)

	// Сохраняем состояние обратно в мир
	anim.Frame = animComponent.Frame
	anim.Timer = animComponent.Timer
	anim.Playing = animComponent.Playing
	anim.FacingRight = animComponent.FacingRight
	world.SetAnimation(entity, anim)
}

// updateFacingDirection обновляет направление взгляда животного на основе движения
func (am *AnimationManager) updateFacingDirection(world *core.World, entity core.EntityID, anim *core.Animation) {
	velocity, hasVel := world.GetVelocity(entity)
	if !hasVel {
		return
	}

	// Обновляем направление взгляда только если животное движется
	if velocity.X*velocity.X+velocity.Y*velocity.Y > SpeedThresholds.Idle {
		anim.FacingRight = velocity.X >= 0
	}
}

// GetAnimationSystem возвращает систему анимации для типа животного
func (am *AnimationManager) GetAnimationSystem(animalType core.AnimalType) (*AnimationSystem, bool) {
	system, exists := am.systems[animalType]
	return system, exists
}
