package main

import (
	"github.com/aiseeq/savanna/internal/animation"
	"github.com/aiseeq/savanna/internal/constants"
	"github.com/aiseeq/savanna/internal/core"
)

// AnimationManager управляет всеми анимационными системами
// Соблюдает SRP - единственная ответственность: управление анимациями
// Соблюдает OCP - легко расширяется новыми типами животных без модификации кода
type AnimationManager struct {
	// Реестр анимационных систем по типам животных
	animalSystems map[core.AnimalType]*animation.AnimationSystem

	// Резолвер для определения типа анимации
	resolver *animation.AnimationResolver
}

// NewAnimationManager создаёт новый менеджер анимаций
func NewAnimationManager() *AnimationManager {
	return &AnimationManager{
		animalSystems: make(map[core.AnimalType]*animation.AnimationSystem),
		resolver:      animation.NewAnimationResolver(),
	}
}

// RegisterAnimalSystem регистрирует анимационную систему для типа животного
// Открыт для расширения - новые животные добавляются без изменения кода
func (am *AnimationManager) RegisterAnimalSystem(animalType core.AnimalType, system *animation.AnimationSystem) {
	am.animalSystems[animalType] = system
}

// GetAnimationSystem возвращает анимационную систему для типа животного
func (am *AnimationManager) GetAnimationSystem(animalType core.AnimalType) (*animation.AnimationSystem, bool) {
	system, exists := am.animalSystems[animalType]
	return system, exists
}

// GetResolver возвращает анимационный резолвер
func (am *AnimationManager) GetResolver() *animation.AnimationResolver {
	return am.resolver
}

// UpdateAnimalAnimations обновляет анимации всех животных в мире
func (am *AnimationManager) UpdateAnimalAnimations(world *core.World, deltaTime float32) {
	// Обходим всех животных с анимациями
	world.ForEachWith(core.MaskAnimalType|core.MaskAnimation, func(entity core.EntityID) {
		animalType, ok := world.GetAnimalType(entity)
		if !ok {
			return
		}

		// Получаем анимационную систему для этого типа животного
		animSystem, exists := am.GetAnimationSystem(animalType)
		if !exists {
			return // Нет системы для этого типа - пропускаем
		}

		// Определяем нужный тип анимации
		expectedAnimType := am.resolver.ResolveAnimalAnimationType(world, entity, animalType)

		// Обновляем анимацию если нужно
		am.updateAnimationIfNeeded(world, entity, expectedAnimType)

		// Обновляем направление анимации на основе скорости
		am.updateAnimationDirection(world, entity)

		// Обрабатываем кадры анимации
		am.processAnimationUpdate(world, entity, animSystem, deltaTime)
	})
}

// updateAnimationIfNeeded обновляет тип анимации если он изменился
func (am *AnimationManager) updateAnimationIfNeeded(
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

// updateAnimationDirection обновляет направление анимации на основе скорости
func (am *AnimationManager) updateAnimationDirection(world *core.World, entity core.EntityID) {
	anim, hasAnim := world.GetAnimation(entity)
	vel, hasVel := world.GetVelocity(entity)

	if !hasAnim || !hasVel {
		return
	}

	// Определяем направление по скорости
	if vel.X > constants.MovementThreshold {
		anim.FacingRight = true
	} else if vel.X < -constants.MovementThreshold {
		anim.FacingRight = false
	}

	world.SetAnimation(entity, anim)
}

// processAnimationUpdate обрабатывает обновление кадров анимации
func (am *AnimationManager) processAnimationUpdate(
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

// LoadAnimationsFromConfig загружает анимации из конфигурации
// Позволяет легко добавлять новые типы животных через конфигурацию
//
//nolint:unparam // error может быть добавлен в будущем для обработки ошибок загрузки
func (am *AnimationManager) LoadAnimationsFromConfig() error {
	// Создаём и регистрируем систему анимаций для зайцев
	rabbitSystem := animation.NewAnimationSystem()
	rabbitAnimations := []struct {
		name     string
		frames   int
		fps      float32
		loop     bool
		animType animation.AnimationType
	}{
		{"hare_idle", 2, 2.0, true, animation.AnimIdle},
		{"hare_walk", 2, 4.0, true, animation.AnimWalk},
		{"hare_run", 2, 12.0, true, animation.AnimRun},
		{"hare_attack", 2, 5.0, false, animation.AnimAttack},
		{"hare_eat", 2, 4.0, true, animation.AnimEat},
		{"hare_dead", 2, 3.0, false, animation.AnimDeathDying},
	}

	for _, config := range rabbitAnimations {
		rabbitSystem.RegisterAnimation(config.animType, config.frames, config.fps, config.loop, nil)
	}
	am.RegisterAnimalSystem(core.TypeRabbit, rabbitSystem)

	// Создаём и регистрируем систему анимаций для волков
	wolfSystem := animation.NewAnimationSystem()
	wolfAnimations := []struct {
		name     string
		frames   int
		fps      float32
		loop     bool
		animType animation.AnimationType
	}{
		{"wolf_idle", 2, 2.0, true, animation.AnimIdle},
		{"wolf_walk", 2, 4.0, true, animation.AnimWalk},
		{"wolf_run", 2, 8.0, true, animation.AnimRun},
		{"wolf_attack", 4, 8.0, false, animation.AnimAttack},
		{"wolf_eat", 2, 4.0, true, animation.AnimEat},
		{"wolf_dead", 2, 3.0, false, animation.AnimDeathDying},
	}

	for _, config := range wolfAnimations {
		wolfSystem.RegisterAnimation(config.animType, config.frames, config.fps, config.loop, nil)
	}
	am.RegisterAnimalSystem(core.TypeWolf, wolfSystem)

	return nil
}
