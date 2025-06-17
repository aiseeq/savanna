package simulation

import (
	"github.com/aiseeq/savanna/internal/animation"
	"github.com/aiseeq/savanna/internal/core"
)

// AttackSystem отвечает ТОЛЬКО за атаки и нанесение урона (устраняет нарушение SRP)
type AttackSystem struct {
	attackCooldowns map[core.EntityID]float32 // Кулдауны атак
}

// NewAttackSystem создаёт новую систему атак
func NewAttackSystem() *AttackSystem {
	return &AttackSystem{
		attackCooldowns: make(map[core.EntityID]float32),
	}
}

// Update обновляет систему атак
func (as *AttackSystem) Update(world *core.World, deltaTime float32) {
	// Обновляем кулдауны атак
	as.updateAttackCooldowns(deltaTime)

	// Обрабатываем атаки хищников
	as.handlePredatorAttacks(world, deltaTime)

	// Очищаем кулдауны мертвых животных
	as.cleanupCooldowns(world)
}

// updateAttackCooldowns обновляет кулдауны атак
func (as *AttackSystem) updateAttackCooldowns(deltaTime float32) {
	for entityID, cooldown := range as.attackCooldowns {
		cooldown -= deltaTime
		if cooldown <= 0 {
			delete(as.attackCooldowns, entityID)
		} else {
			as.attackCooldowns[entityID] = cooldown
		}
	}
}

// handlePredatorAttacks обрабатывает атаки хищников (устраняет нарушение OCP)
// Работает с любыми хищниками через компонент Behavior, а не захардкоженные типы
func (as *AttackSystem) handlePredatorAttacks(world *core.World, deltaTime float32) {
	// Обрабатываем ВСЕХ животных с поведением хищника БЕЗ AttackState
	world.ForEachWith(core.MaskBehavior|core.MaskPosition|core.MaskSize, func(predator core.EntityID) {
		as.tryStartAttack(world, predator)
	})

	// Обрабатываем хищников С AttackState - обновляем фазы атаки
	world.ForEachWith(core.MaskAttackState, func(predator core.EntityID) {
		as.updateAttackState(world, predator, deltaTime)
	})
}

// findAttackTarget ищет цель для атаки (универсальная логика для любых хищников)
// Устраняет нарушение OCP - теперь работает с любыми типами через компонент Behavior
func (as *AttackSystem) findAttackTarget(world *core.World, attacker core.EntityID) core.EntityID {
	attackerPos, hasPos := world.GetPosition(attacker)
	if !hasPos {
		return 0
	}

	// Получаем поведение атакующего
	behavior, hasBehavior := world.GetBehavior(attacker)
	if !hasBehavior {
		return 0
	}

	// Получаем размеры атакующего
	attackerSize, hasAttackerSize := world.GetSize(attacker)
	if !hasAttackerSize || attackerSize.AttackRange <= 0 {
		return 0 // Не хищник или не может атаковать
	}

	// Используем дальность видения из поведения (универсально!)
	searchRadius := behavior.VisionRange

	// ПОИСК ЛЮБЫХ ТРАВОЯДНЫХ (устраняет захардкоженность TypeRabbit)
	// Ищем ближайшее животное с поведением травоядного
	var closestTarget core.EntityID
	var closestDistance float32 = searchRadius*searchRadius + 1 // За пределами радиуса

	world.ForEachWith(core.MaskBehavior|core.MaskPosition|core.MaskSize, func(candidate core.EntityID) {
		if !as.isValidHerbivoreTarget(world, attacker, candidate) {
			return
		}

		candidatePos, _ := world.GetPosition(candidate)
		dx := attackerPos.X - candidatePos.X
		dy := attackerPos.Y - candidatePos.Y
		distance := dx*dx + dy*dy

		if distance < closestDistance {
			closestDistance = distance
			closestTarget = candidate
		}
	})

	if closestTarget == 0 {
		return 0
	}

	// Проверяем что цель в радиусе атаки
	targetPos, hasTargetPos := world.GetPosition(closestTarget)
	if !hasTargetPos {
		return 0
	}

	targetSize, hasTargetSize := world.GetSize(closestTarget)
	if !hasTargetSize {
		return 0
	}

	// Проверяем возможность атаки
	if as.isWithinAttackRange(attackerPos, targetPos, attackerSize.AttackRange, targetSize.Radius) {
		return closestTarget
	}

	return 0
}

// isValidHerbivoreTarget проверяет что цель подходит для атаки (снижает сложность)
func (as *AttackSystem) isValidHerbivoreTarget(world *core.World, attacker, candidate core.EntityID) bool {
	if candidate == attacker || world.HasComponent(candidate, core.MaskCorpse) {
		return false
	}

	candidateBehavior, hasBehavior := world.GetBehavior(candidate)
	return hasBehavior && candidateBehavior.Type == core.BehaviorHerbivore
}

// tryStartAttack пытается начать атаку для хищника (упрощена через вспомогательные методы)
func (as *AttackSystem) tryStartAttack(world *core.World, predator core.EntityID) {
	// Валидация базовых условий для атаки
	if !as.canStartAttack(world, predator) {
		return
	}

	// Поиск цели и начало атаки
	target := as.findAttackTarget(world, predator)
	if target != 0 {
		as.startAttackState(world, predator, target)
	}
}

// canStartAttack проверяет может ли хищник начать атаку (извлечено из tryStartAttack)
func (as *AttackSystem) canStartAttack(world *core.World, predator core.EntityID) bool {
	// Проверяем что это хищник
	behavior, ok := world.GetBehavior(predator)
	if !ok || behavior.Type != core.BehaviorPredator {
		return false
	}

	// Проверяем возможность атаковать
	size, hasSize := world.GetSize(predator)
	if !hasSize || size.AttackRange <= 0 {
		return false
	}

	// Проверяем что не атакует уже
	if world.HasComponent(predator, core.MaskAttackState) {
		return false
	}

	// Проверяем кулдаун
	if !as.canAttack(predator) {
		return false
	}

	// Проверяем голод
	hunger, hasHunger := world.GetHunger(predator)
	return hasHunger && hunger.Value < behavior.HungerThreshold
}

// startAttackState создаёт состояние атаки и анимацию (снижает сложность)
func (as *AttackSystem) startAttackState(world *core.World, predator, target core.EntityID) {
	attackState := core.AttackState{
		Target:     target,
		Phase:      core.AttackPhaseWindup,
		PhaseTimer: 0.0,
		TotalTimer: 0.0,
		HasStruck:  false,
	}
	world.AddAttackState(predator, attackState)

	// АВТОМАТИЧЕСКИ устанавливаем анимацию атаки (если есть компонент Animation)
	if world.HasComponent(predator, core.MaskAnimation) {
		world.SetAnimation(predator, core.Animation{
			CurrentAnim: int(animation.AnimAttack),
			Frame:       AttackFrameWindup,
			Timer:       0,
			Playing:     true,
			FacingRight: true, // TODO: определять направление по положению цели
		})
	}
}

// updateAttackState обновляет состояние атаки хищника
func (as *AttackSystem) updateAttackState(world *core.World, predator core.EntityID, deltaTime float32) {
	attackState, hasAttack := world.GetAttackState(predator)
	if !hasAttack {
		return
	}

	// Обновляем таймеры
	attackState.PhaseTimer += deltaTime
	attackState.TotalTimer += deltaTime

	// Получаем анимацию для синхронизации (если есть)
	anim, hasAnim := world.GetAnimation(predator)

	if hasAnim {
		as.updateAttackStateWithAnimation(world, predator, &attackState, anim)
	} else if as.updateAttackStateWithTimer(world, predator, &attackState) {
		return // Состояние удалено
	}

	// ВАЖНО: Сохраняем обновленное состояние в конце
	world.SetAttackState(predator, attackState)
}

// updateAttackStateWithAnimation обновляет атаку по анимации (снижает сложность)
func (as *AttackSystem) updateAttackStateWithAnimation(
	world *core.World, predator core.EntityID, attackState *core.AttackState, anim core.Animation,
) {
	switch attackState.Phase {
	case core.AttackPhaseWindup:
		// КАДР WINDUP: Замах - проверяем готовность к удару
		if anim.CurrentAnim == int(animation.AnimAttack) && anim.Frame == AttackFrameStrike {
			// Переходим в фазу удара
			attackState.Phase = core.AttackPhaseStrike
			attackState.PhaseTimer = 0.0
			world.SetAttackState(predator, *attackState)
		}

	case core.AttackPhaseStrike:
		// КАДР STRIKE: Удар - наносим урон ОДИН РАЗ
		if !attackState.HasStruck {
			as.executeStrike(world, predator, attackState.Target)
			attackState.HasStruck = true
			world.SetAttackState(predator, *attackState)
		}

		// Проверяем завершение атаки
		if !anim.Playing || anim.CurrentAnim != int(animation.AnimAttack) {
			// Анимация завершена - устанавливаем кулдаун и удаляем состояние атаки
			as.setAttackCooldown(predator)
			world.RemoveAttackState(predator)
		}
	}
}

// updateAttackStateWithTimer обновляет атаку по таймерам (возвращает true если удалено)
func (as *AttackSystem) updateAttackStateWithTimer(
	world *core.World, predator core.EntityID, attackState *core.AttackState,
) bool {
	const WINDUP_DURATION = 0.08 // 5 тиков (примерно 0.083 сек)
	const STRIKE_DURATION = 0.2  // 0.2 секунды на удар

	switch attackState.Phase {
	case core.AttackPhaseWindup:
		if attackState.PhaseTimer >= WINDUP_DURATION {
			// Время перейти к удару
			attackState.Phase = core.AttackPhaseStrike
			attackState.PhaseTimer = 0.0
		}

	case core.AttackPhaseStrike:
		// Наносим урон сразу при входе в фазу Strike
		if !attackState.HasStruck {
			as.executeStrike(world, predator, attackState.Target)
			attackState.HasStruck = true
		}

		// Завершаем атаку по таймеру
		if attackState.PhaseTimer >= STRIKE_DURATION {
			// Атака завершена - устанавливаем кулдаун и удаляем состояние
			as.setAttackCooldown(predator)
			world.RemoveAttackState(predator)
			return true // Состояние удалено
		}
	}
	return false
}

// executeStrike выполняет удар - наносит урон цели (упрощена через вспомогательные методы)
func (as *AttackSystem) executeStrike(world *core.World, attacker, target core.EntityID) {
	// Валидация цели и позиций
	if !as.validateStrikeTarget(world, target) {
		return
	}

	// Валидация дальности атаки
	if !as.validateStrikeRange(world, attacker, target) {
		return
	}

	// Выполнение атаки с проверкой попадания
	as.performStrikeAttempt(world, attacker, target)
}

// validateStrikeTarget проверяет что цель валидна для атаки (извлечено из executeStrike)
func (as *AttackSystem) validateStrikeTarget(world *core.World, target core.EntityID) bool {
	return world.IsAlive(target) && !world.HasComponent(target, core.MaskCorpse)
}

// validateStrikeRange проверяет что цель в радиусе атаки (извлечено из executeStrike)
func (as *AttackSystem) validateStrikeRange(world *core.World, attacker, target core.EntityID) bool {
	attackerPos, hasAttackerPos := world.GetPosition(attacker)
	targetPos, hasTargetPos := world.GetPosition(target)
	if !hasAttackerPos || !hasTargetPos {
		return false
	}

	attackerSize, hasAttackerSize := world.GetSize(attacker)
	targetSize, hasTargetSize := world.GetSize(target)
	if !hasAttackerSize || !hasTargetSize {
		return false
	}

	return as.isWithinAttackRange(attackerPos, targetPos, attackerSize.AttackRange, targetSize.Radius)
}

// performStrikeAttempt выполняет попытку атаки с проверкой попадания (извлечено из executeStrike)
func (as *AttackSystem) performStrikeAttempt(world *core.World, attacker, target core.EntityID) {
	config, hasConfig := world.GetAnimalConfig(attacker)
	if !hasConfig {
		return // Нет конфигурации - не можем атаковать
	}

	// Проверяем шанс попадания
	rng := world.GetRNG()
	if rng.Float32() < config.HitChance {
		as.dealDamageToTarget(world, target, config.AttackDamage)
	}
}

// dealDamageToTarget наносит урон цели
func (as *AttackSystem) dealDamageToTarget(world *core.World, target core.EntityID, damage int16) {
	health, hasHealth := world.GetHealth(target)
	if !hasHealth {
		return
	}

	// Наносим урон
	health.Current -= damage
	if health.Current < 0 {
		health.Current = 0
	}

	world.SetHealth(target, health)

	// Добавляем эффект мигания
	world.AddDamageFlash(target, core.DamageFlash{
		Timer:     DamageFlashDuration,
		Duration:  DamageFlashDuration,
		Intensity: 1.0, // Максимальная интенсивность для лучшей видимости
	})

	// Если цель умерла, превращаем её в труп
	if health.Current == 0 {
		createCorpse(world, target)
	}
}

// canAttack проверяет может ли животное атаковать (нет кулдауна)
func (as *AttackSystem) canAttack(entity core.EntityID) bool {
	_, hasCooldown := as.attackCooldowns[entity]
	return !hasCooldown
}

// setAttackCooldown устанавливает кулдаун атаки
func (as *AttackSystem) setAttackCooldown(entity core.EntityID) {
	as.attackCooldowns[entity] = AttackCooldownSeconds
}

// cleanupCooldowns очищает кулдауны для мертвых животных
func (as *AttackSystem) cleanupCooldowns(world *core.World) {
	for entityID := range as.attackCooldowns {
		if !world.IsAlive(entityID) {
			delete(as.attackCooldowns, entityID)
		}
	}
}

// isWithinAttackRange проверяет находится ли цель в радиусе атаки
func (as *AttackSystem) isWithinAttackRange(
	attackerPos, targetPos core.Position, attackRange, targetRadius float32,
) bool {
	// Вычисляем квадрат расстояния между центрами
	dx := attackerPos.X - targetPos.X
	dy := attackerPos.Y - targetPos.Y
	distanceSquared := dx*dx + dy*dy

	// Максимальная дистанция атаки: радиус атаки + радиус цели
	maxAttackDistance := attackRange + targetRadius

	// Сравниваем квадраты для избежания sqrt
	return distanceSquared <= maxAttackDistance*maxAttackDistance
}
