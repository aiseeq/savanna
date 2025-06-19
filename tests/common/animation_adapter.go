package common

import (
	"github.com/aiseeq/savanna/internal/animation"
	"github.com/aiseeq/savanna/internal/core"
)

// AnimationSystemAdapter адаптирует AnimationSystem к интерфейсу core.System
// Упрощённая версия AnimationManager для тестов
type AnimationSystemAdapter struct {
	rabbitSystem *animation.AnimationSystem
	wolfSystem   *animation.AnimationSystem
	resolver     *animation.AnimationResolver
}

// NewAnimationSystemAdapter создаёт новый адаптер анимационной системы для тестов
func NewAnimationSystemAdapter() *AnimationSystemAdapter {
	adapter := &AnimationSystemAdapter{
		rabbitSystem: animation.NewAnimationSystem(),
		wolfSystem:   animation.NewAnimationSystem(),
		resolver:     animation.NewAnimationResolver(),
	}

	// Регистрируем анимации для зайцев
	//nolint:gomnd // Конфигурация анимации зайцев
	adapter.rabbitSystem.RegisterAnimation(animation.AnimIdle, 2, 2.0, true, nil)
	//nolint:gomnd // Конфигурация анимации зайцев
	adapter.rabbitSystem.RegisterAnimation(animation.AnimWalk, 2, 4.0, true, nil)
	//nolint:gomnd // Конфигурация анимации зайцев
	adapter.rabbitSystem.RegisterAnimation(animation.AnimRun, 2, 12.0, true, nil)
	//nolint:gomnd // Конфигурация анимации зайцев
	adapter.rabbitSystem.RegisterAnimation(animation.AnimAttack, 2, 5.0, false, nil)
	//nolint:gomnd // Конфигурация анимации зайцев
	adapter.rabbitSystem.RegisterAnimation(animation.AnimEat, 2, 4.0, true, nil)
	//nolint:gomnd // Конфигурация анимации зайцев
	adapter.rabbitSystem.RegisterAnimation(animation.AnimDeathDying, 2, 3.0, false, nil)

	// Регистрируем анимации для волков (4 кадра для атаки волка)
	//nolint:gomnd // Конфигурация анимации волков
	adapter.wolfSystem.RegisterAnimation(animation.AnimIdle, 2, 2.0, true, nil)
	//nolint:gomnd // Конфигурация анимации волков
	adapter.wolfSystem.RegisterAnimation(animation.AnimWalk, 2, 4.0, true, nil)
	//nolint:gomnd // Конфигурация анимации волков
	adapter.wolfSystem.RegisterAnimation(animation.AnimRun, 2, 8.0, true, nil)
	//nolint:gomnd // Конфигурация анимации волков
	adapter.wolfSystem.RegisterAnimation(animation.AnimAttack, 4, 8.0, false, nil)
	//nolint:gomnd // Конфигурация анимации волков
	adapter.wolfSystem.RegisterAnimation(animation.AnimEat, 2, 4.0, true, nil)
	//nolint:gomnd // Конфигурация анимации волков
	adapter.wolfSystem.RegisterAnimation(animation.AnimDeathDying, 2, 3.0, false, nil)

	return adapter
}

// Update реализует интерфейс core.System
func (asa *AnimationSystemAdapter) Update(world *core.World, deltaTime float32) {
	// Обходим всех животных с анимациями
	world.ForEachWith(core.MaskAnimalType|core.MaskAnimation, func(entity core.EntityID) {
		animalType, ok := world.GetAnimalType(entity)
		if !ok {
			return
		}

		// Получаем анимационную систему для этого типа животного
		var animSystem *animation.AnimationSystem
		switch animalType {
		case core.TypeRabbit:
			animSystem = asa.rabbitSystem
		case core.TypeWolf:
			animSystem = asa.wolfSystem
		default:
			return // Неизвестный тип
		}

		// Определяем нужный тип анимации
		expectedAnimType := asa.resolver.ResolveAnimalAnimationType(world, entity, animalType)

		// Обновляем анимацию если нужно
		asa.updateAnimationIfNeeded(world, entity, expectedAnimType)

		// Обрабатываем кадры анимации
		asa.processAnimationUpdate(world, entity, animSystem, deltaTime)
	})
}

// updateAnimationIfNeeded обновляет тип анимации если он изменился
func (asa *AnimationSystemAdapter) updateAnimationIfNeeded(
	world *core.World,
	entity core.EntityID,
	newAnimType animation.AnimationType,
) {
	anim, ok := world.GetAnimation(entity)
	if !ok {
		return
	}

	// Проверяем нужно ли менять анимацию
	if anim.CurrentAnim != int(newAnimType) {
		// НЕ прерываем анимацию ATTACK если она играет
		if anim.CurrentAnim == int(animation.AnimAttack) && anim.Playing {
			return
		}

		// Меняем анимацию
		anim.CurrentAnim = int(newAnimType)
		anim.Frame = 0
		anim.Timer = 0
		anim.Playing = true
		world.SetAnimation(entity, anim)
	}
}

// processAnimationUpdate обрабатывает обновление кадров анимации
func (asa *AnimationSystemAdapter) processAnimationUpdate(
	world *core.World,
	entity core.EntityID,
	animSystem *animation.AnimationSystem,
	deltaTime float32,
) {
	anim, ok := world.GetAnimation(entity)
	if !ok {
		return
	}

	// Создаём компонент для системы анимации
	animComponent := animation.AnimationComponent{
		CurrentAnim: animation.AnimationType(anim.CurrentAnim),
		Frame:       anim.Frame,
		Timer:       anim.Timer,
		Playing:     anim.Playing,
		FacingRight: anim.FacingRight,
	}

	// Обновляем через систему анимации
	animSystem.Update(&animComponent, deltaTime)

	// Сохраняем обновлённое состояние
	anim.Frame = animComponent.Frame
	anim.Timer = animComponent.Timer
	anim.Playing = animComponent.Playing
	anim.FacingRight = animComponent.FacingRight
	world.SetAnimation(entity, anim)
}
