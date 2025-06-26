package unit

import (
	"testing"

	"github.com/aiseeq/savanna/internal/core"
	"github.com/aiseeq/savanna/internal/simulation"
)

// TestAttackStateAutoCreation - Unit тест проверяющий автоматическое создание AttackState
//
// ЦЕЛЬ: Проверить что AttackSystem автоматически создает AttackState для голодных хищников
func TestAttackStateAutoCreation(t *testing.T) {
	t.Parallel()

	world := core.NewWorld(640, 640, 42)
	combatSystem := simulation.NewCombatSystem()

	// Создаём волка и зайца через правильный API
	rabbit := simulation.CreateAnimal(world, core.TypeRabbit, 300, 300)
	wolf := simulation.CreateAnimal(world, core.TypeWolf, 305, 300) // Очень близко

	t.Logf("=== ТЕСТ АВТОМАТИЧЕСКОГО СОЗДАНИЯ ATTACKSTATE ===")
	t.Logf("Заяц: entity %d (позиция 300,300)", rabbit)
	t.Logf("Волк: entity %d (позиция 305,300)", wolf)

	// Делаем волка очень голодным для гарантированной атаки
	world.SetSatiation(wolf, core.Satiation{Value: 20.0}) // 20% < 60% = очень голодный

	// Проверяем начальное состояние
	if world.HasComponent(wolf, core.MaskAttackState) {
		t.Fatal("AttackState уже существует в начале теста")
	}

	// Проверяем что у животных есть все нужные компоненты
	wolfConfig, hasWolfConfig := world.GetAnimalConfig(wolf)
	if !hasWolfConfig {
		t.Fatal("У волка нет AnimalConfig")
	}
	t.Logf("Конфигурация волка: урон=%d, дальность=%.1f, шанс=%.1f%%",
		wolfConfig.AttackDamage, wolfConfig.AttackRange, wolfConfig.HitChance*100)

	wolfBehavior, hasWolfBehavior := world.GetBehavior(wolf)
	if !hasWolfBehavior {
		t.Fatal("У волка нет Behavior")
	}
	t.Logf("Поведение волка: тип=%v, порог голода=%.1f%%", wolfBehavior.Type, wolfBehavior.SatiationThreshold)

	if wolfBehavior.Type != core.BehaviorPredator {
		t.Fatalf("Неправильный тип поведения волка: %v (ожидался Predator)", wolfBehavior.Type)
	}

	wolfHunger, _ := world.GetSatiation(wolf)
	t.Logf("Голод волка: %.1f%% (порог атаки: %.1f%%)", wolfHunger.Value, wolfBehavior.SatiationThreshold)

	if wolfHunger.Value >= wolfBehavior.SatiationThreshold {
		t.Fatalf("Волк недостаточно голодный: %.1f%% >= %.1f%%", wolfHunger.Value, wolfBehavior.SatiationThreshold)
	}

	deltaTime := float32(1.0 / 60.0)

	// Симулируем несколько обновлений, ожидая создания AttackState
	for i := 0; i < 120; i++ { // 2 секунды максимум
		combatSystem.Update(world, deltaTime)

		if world.HasComponent(wolf, core.MaskAttackState) {
			t.Logf("✅ AttackState создан на тике %d", i)

			attackState, _ := world.GetAttackState(wolf)
			t.Logf("  Цель: entity %d", attackState.Target)
			t.Logf("  Фаза: %v", attackState.Phase)

			if attackState.Target != rabbit {
				t.Errorf("Неправильная цель атаки: entity %d (ожидался %d)", attackState.Target, rabbit)
			}

			if attackState.Phase != core.AttackPhaseWindup {
				t.Errorf("Неправильная фаза атаки: %v (ожидалась Windup)", attackState.Phase)
			}

			t.Logf("✅ AttackState создан правильно!")
			return
		}

		// Логируем диагностику каждые 30 тиков (0.5 секунды)
		if i%30 == 0 {
			t.Logf("Тик %d: AttackState пока не создан", i)

			// Дополнительная диагностика
			wolfPos, _ := world.GetPosition(wolf)
			rabbitPos, _ := world.GetPosition(rabbit)
			distance := ((wolfPos.X-rabbitPos.X)*(wolfPos.X-rabbitPos.X) + (wolfPos.Y-rabbitPos.Y)*(wolfPos.Y-rabbitPos.Y))
			t.Logf("  Расстояние: %.1f пикселей", distance)

			wolfSize, _ := world.GetSize(wolf)
			rabbitSize, _ := world.GetSize(rabbit)
			maxAttackDistanceTiles := wolfSize.AttackRange + rabbitSize.Radius // в тайлах
			maxAttackDistancePixels := maxAttackDistanceTiles * 32             // конвертируем в пиксели (1 тайл = 32 пикселя)
			t.Logf("  Максимальная дальность атаки: %.1f тайла = %.1f пикселей", maxAttackDistanceTiles, maxAttackDistancePixels)

			// Проверяем расстояние для атаки
			maxDistanceSquared := maxAttackDistancePixels * maxAttackDistancePixels
			if distance > maxDistanceSquared {
				t.Logf("  ПРИЧИНА: Заяц слишком далеко для атаки (%.1f > %.1f)", distance, maxDistanceSquared)
			}
		}
	}

	t.Fatal("БАГ: AttackState не был создан за 2 секунды")
}

// TestAttackStateCreationConditions - Unit тест проверяющий условия создания AttackState
//
// ЦЕЛЬ: Проверить различные условия при которых AttackState должен или не должен создаваться
func TestAttackStateCreationConditions(t *testing.T) {
	t.Parallel()

	world := core.NewWorld(640, 640, 42)
	attackSystem := simulation.NewAttackSystem()

	t.Logf("=== ТЕСТ УСЛОВИЙ СОЗДАНИЯ ATTACKSTATE ===")

	// Тест 1: Сытый волк не должен атаковать
	t.Logf("\n--- Тест 1: Сытый волк ---")
	_ = simulation.CreateAnimal(world, core.TypeRabbit, 300, 300) // rabbit1
	wolf1 := simulation.CreateAnimal(world, core.TypeWolf, 305, 300)
	world.SetSatiation(wolf1, core.Satiation{Value: 80.0}) // 80% > 60% = сытый

	deltaTime := float32(1.0 / 60.0)
	for i := 0; i < 30; i++ { // Полсекунды
		attackSystem.Update(world, deltaTime)
	}

	if world.HasComponent(wolf1, core.MaskAttackState) {
		t.Error("БАГ: Сытый волк создал AttackState")
	} else {
		t.Logf("✅ Сытый волк правильно НЕ атакует")
	}

	// Тест 2: Голодный волк должен атаковать
	t.Logf("\n--- Тест 2: Голодный волк ---")
	_ = simulation.CreateAnimal(world, core.TypeRabbit, 400, 300) // rabbit2
	wolf2 := simulation.CreateAnimal(world, core.TypeWolf, 405, 300)
	world.SetSatiation(wolf2, core.Satiation{Value: 30.0}) // 30% < 60% = голодный

	attackStateCreated := false
	for i := 0; i < 60; i++ { // 1 секунда
		attackSystem.Update(world, deltaTime)

		if world.HasComponent(wolf2, core.MaskAttackState) {
			attackStateCreated = true
			t.Logf("✅ Голодный волк создал AttackState на тике %d", i)
			break
		}
	}

	if !attackStateCreated {
		t.Error("БАГ: Голодный волк НЕ создал AttackState")
	}

	// Тест 3: Волк слишком далеко от зайца
	t.Logf("\n--- Тест 3: Волк слишком далеко ---")
	_ = simulation.CreateAnimal(world, core.TypeRabbit, 500, 300)    // rabbit3
	wolf3 := simulation.CreateAnimal(world, core.TypeWolf, 600, 300) // 100 пикселей - далеко
	world.SetSatiation(wolf3, core.Satiation{Value: 30.0})           // Голодный

	for i := 0; i < 30; i++ { // Полсекунды
		attackSystem.Update(world, deltaTime)
	}

	if world.HasComponent(wolf3, core.MaskAttackState) {
		t.Error("БАГ: Далёкий волк создал AttackState")
	} else {
		t.Logf("✅ Далёкий волк правильно НЕ атакует")
	}

	// Тест 4: Заяц не должен атаковать (травоядное)
	t.Logf("\n--- Тест 4: Заяц (травоядное) ---")
	rabbit4 := simulation.CreateAnimal(world, core.TypeRabbit, 700, 300)
	_ = simulation.CreateAnimal(world, core.TypeRabbit, 705, 300) // rabbit5
	world.SetSatiation(rabbit4, core.Satiation{Value: 10.0})      // Очень голодный

	for i := 0; i < 30; i++ { // Полсекунды
		attackSystem.Update(world, deltaTime)
	}

	if world.HasComponent(rabbit4, core.MaskAttackState) {
		t.Error("БАГ: Заяц создал AttackState (травоядное не должно атаковать)")
	} else {
		t.Logf("✅ Заяц правильно НЕ атакует")
	}

	t.Logf("\n✅ Все тесты условий прошли успешно")
}
