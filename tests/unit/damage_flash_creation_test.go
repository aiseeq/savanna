package unit

import (
	"testing"

	"github.com/aiseeq/savanna/internal/core"
	"github.com/aiseeq/savanna/internal/simulation"
)

// TestDamageFlashCreationInAttackSystem - Unit тест проверяющий создание DamageFlash в AttackSystem
//
// ЦЕЛЬ: Проверить что DamageFlash компонент создается при нанесении урона
func TestDamageFlashCreationInAttackSystem(t *testing.T) {
	t.Parallel()

	world := core.NewWorld(640, 640, 42)
	attackSystem := simulation.NewAttackSystem()

	// Создаём тестовые сущности
	attacker := world.CreateEntity()
	target := world.CreateEntity()

	// Настраиваем атакующего как волка
	world.AddPosition(attacker, core.Position{X: 100, Y: 100})
	world.AddSize(attacker, core.Size{
		Radius:      24.0, // Радиус волка в пикселях
		AttackRange: 28.8, // Атака волка в пикселях (0.9 тайла * 32 пикселя/тайл)
	})
	world.AddBehavior(attacker, core.Behavior{
		Type:               core.BehaviorPredator,
		VisionRange:        160.0, // 5 тайлов в пикселях
		SatiationThreshold: 60.0,
	})
	world.AddAnimalConfig(attacker, core.AnimalConfig{
		AttackDamage: 25,
		HitChance:    1.0, // 100% попадания для теста
	})

	// Настраиваем цель как зайца
	world.AddPosition(target, core.Position{X: 105, Y: 100}) // Очень близко
	world.AddSize(target, core.Size{
		Radius:      16.0, // Радиус зайца в пикселях
		AttackRange: 0,
	})
	world.AddBehavior(target, core.Behavior{
		Type: core.BehaviorHerbivore,
	})
	world.AddHealth(target, core.Health{
		Max:     50,
		Current: 50,
	})

	t.Logf("=== ТЕСТ СОЗДАНИЯ DAMAGEFLASH ===")
	t.Logf("Атакующий: entity %d (позиция 100,100)", attacker)
	t.Logf("Цель: entity %d (позиция 105,100)", target)

	// Создаём AttackState вручную (имитируем что атака уже начата)
	world.AddAttackState(attacker, core.AttackState{
		Target:     target,
		Phase:      core.AttackPhaseStrike, // Прямо в фазу удара
		PhaseTimer: 0.0,
		TotalTimer: 0.0,
		HasStruck:  false,
	})

	// Добавляем анимацию на кадр удара
	world.AddAnimation(attacker, core.Animation{
		CurrentAnim: 6, // AnimAttack = 6
		Frame:       1, // Кадр удара
		Timer:       0,
		Playing:     true,
		FacingRight: true,
	})

	// Проверяем начальное состояние
	if world.HasComponent(target, core.MaskDamageFlash) {
		t.Fatal("DamageFlash уже существует до атаки")
	}

	initialHealth, _ := world.GetHealth(target)
	t.Logf("Начальное здоровье цели: %d", initialHealth.Current)

	// Выполняем обновление системы атак
	deltaTime := float32(1.0 / 60.0)
	attackSystem.Update(world, deltaTime)

	// КРИТИЧЕСКАЯ ПРОВЕРКА 1: Урон должен быть нанесен
	finalHealth, _ := world.GetHealth(target)
	damageTaken := initialHealth.Current - finalHealth.Current

	if damageTaken <= 0 {
		t.Errorf("Урон не был нанесен. Здоровье: %d -> %d", initialHealth.Current, finalHealth.Current)

		// Диагностика
		if attackState, hasAttack := world.GetAttackState(attacker); hasAttack {
			t.Errorf("ДИАГНОСТИКА AttackState: фаза=%v, HasStruck=%v", attackState.Phase, attackState.HasStruck)
		}

		t.Fatal("Тест невозможен без урона")
	}

	t.Logf("✅ Урон нанесен: %d единиц (%d -> %d)", damageTaken, initialHealth.Current, finalHealth.Current)

	// КРИТИЧЕСКАЯ ПРОВЕРКА 2: DamageFlash должен быть создан
	if !world.HasComponent(target, core.MaskDamageFlash) {
		t.Fatal("БАГ ОБНАРУЖЕН: DamageFlash НЕ создан при нанесении урона!")
	}

	// КРИТИЧЕСКАЯ ПРОВЕРКА 3: Параметры DamageFlash должны быть корректными
	flash, hasFlash := world.GetDamageFlash(target)
	if !hasFlash {
		t.Fatal("DamageFlash компонент есть, но не читается")
	}

	t.Logf("✅ DamageFlash создан успешно")
	t.Logf("  Таймер: %.3f секунды", flash.Timer)
	t.Logf("  Длительность: %.3f секунды", flash.Duration)
	t.Logf("  Интенсивность: %.3f", flash.Intensity)

	// Проверяем разумность параметров
	if flash.Timer <= 0 {
		t.Errorf("БАГ: Таймер DamageFlash <= 0: %.3f", flash.Timer)
	}

	if flash.Duration <= 0 {
		t.Errorf("БАГ: Длительность DamageFlash <= 0: %.3f", flash.Duration)
	}

	if flash.Intensity <= 0 || flash.Intensity > 1.0 {
		t.Errorf("БАГ: Интенсивность DamageFlash вне диапазона [0,1]: %.3f", flash.Intensity)
	}

	// Ожидаемое значение таймера (из константы DamageFlashDuration)
	expectedDuration := float32(0.16) // Из combat.go
	tolerance := float32(0.01)

	if flash.Timer < expectedDuration-tolerance || flash.Timer > expectedDuration+tolerance {
		t.Errorf("БАГ: Неправильная длительность DamageFlash: %.3f (ожидалось ~%.3f)",
			flash.Timer, expectedDuration)
	}

	t.Logf("✅ Все проверки DamageFlash пройдены успешно")
}

// TestDamageFlashUpdateInDamageSystem - Unit тест проверяющий обновление DamageFlash в DamageSystem
//
// ЦЕЛЬ: Проверить что DamageFlash правильно угасает и удаляется
func TestDamageFlashUpdateInDamageSystem(t *testing.T) {
	t.Parallel()

	world := core.NewWorld(640, 640, 42)
	damageSystem := simulation.NewDamageSystem()

	// Создаём тестовую сущность с DamageFlash
	entity := world.CreateEntity()
	initialTimer := float32(0.16) // 160ms

	world.AddDamageFlash(entity, core.DamageFlash{
		Timer:     initialTimer,
		Duration:  initialTimer,
		Intensity: 1.0,
	})

	t.Logf("=== ТЕСТ ОБНОВЛЕНИЯ DAMAGEFLASH ===")
	t.Logf("Начальный таймер: %.3f секунды", initialTimer)

	deltaTime := float32(1.0 / 60.0) // 16.67ms

	// Фаза 1: Проверяем несколько обновлений
	for i := 0; i < 5; i++ {
		damageSystem.Update(world, deltaTime)

		if !world.HasComponent(entity, core.MaskDamageFlash) {
			t.Logf("DamageFlash исчез на итерации %d", i)
			break
		}

		flash, _ := world.GetDamageFlash(entity)
		t.Logf("Итерация %d: таймер=%.3f, интенсивность=%.3f", i, flash.Timer, flash.Intensity)

		// Проверяем что таймер уменьшается
		expectedTimer := initialTimer - float32(i+1)*deltaTime
		tolerance := float32(0.001)

		if flash.Timer < expectedTimer-tolerance || flash.Timer > expectedTimer+tolerance {
			t.Errorf("БАГ: Неправильный таймер на итерации %d: %.3f (ожидалось %.3f)",
				i, flash.Timer, expectedTimer)
		}

		// Проверяем что интенсивность уменьшается пропорционально
		expectedIntensity := flash.Timer / flash.Duration
		if flash.Intensity < expectedIntensity-tolerance || flash.Intensity > expectedIntensity+tolerance {
			t.Errorf("БАГ: Неправильная интенсивность на итерации %d: %.3f (ожидалось %.3f)",
				i, flash.Intensity, expectedIntensity)
		}
	}

	// Фаза 2: Симулируем полное угасание
	for i := 0; i < 20; i++ {
		damageSystem.Update(world, deltaTime)

		if !world.HasComponent(entity, core.MaskDamageFlash) {
			t.Logf("✅ DamageFlash правильно исчез на итерации %d (через %.3f сек)",
				i, float32(i)*deltaTime)
			return
		}
	}

	t.Error("БАГ: DamageFlash не исчез за разумное время (20 итераций)")
}
