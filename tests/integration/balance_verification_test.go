package integration

import (
	"testing"

	"github.com/aiseeq/savanna/internal/core"
	"github.com/aiseeq/savanna/internal/simulation"
	"github.com/aiseeq/savanna/tests/common"
)

// TestBalanceVerification проверяет общую стабильность экосистемы за 15 секунд симуляции
//
//nolint:revive // function-length: Финальный тест стабильности всей экосистемы
func TestBalanceVerification(t *testing.T) {
	t.Parallel()

	t.Logf("=== ТЕСТ СТАБИЛЬНОСТИ ЭКОСИСТЕМЫ ===")
	t.Logf("ЦЕЛЬ: Убедиться что экосистема стабильна в течение 15 секунд")

	// Создаём реалистичный мир
	worldSize := float32(50 * 32) // 50x50 тайлов = 1600x1600 пикселей
	world := core.NewWorld(worldSize, worldSize, 12345)

	// Используем централизованный системный менеджер
	systemManager := common.CreateTestSystemManager(worldSize)

	// Создаём экосистему с разными животными
	animals := make([]core.EntityID, 0, 20)

	// 12 зайцев в разных местах карты
	for i := 0; i < 12; i++ {
		x := float32(200 + (i%3)*400) // 3 колонки
		y := float32(200 + (i/3)*300) // 4 ряда
		rabbit := simulation.CreateAnimal(world, core.TypeRabbit, x, y)
		animals = append(animals, rabbit)
	}

	// 8 волков в центральной области
	for i := 0; i < 8; i++ {
		x := float32(600 + (i%2)*400) // 2 колонки
		y := float32(600 + (i/2)*200) // 4 ряда
		wolf := simulation.CreateAnimal(world, core.TypeWolf, x, y)
		animals = append(animals, wolf)
	}

	t.Logf("Создана экосистема: 12 зайцев + 8 волков = %d животных", len(animals))

	// Записываем начальное состояние
	initialStats := gatherEcosystemStats(world, animals)
	t.Logf("Начальное состояние:")
	t.Logf("  Живых зайцев: %d", initialStats.AliveRabbits)
	t.Logf("  Живых волков: %d", initialStats.AliveWolves)
	t.Logf("  Средний голод зайцев: %.1f%%", initialStats.AvgRabbitHunger)
	t.Logf("  Средний голод волков: %.1f%%", initialStats.AvgWolfHunger)
	t.Logf("  Средне здоровье зайцев: %.1f%%", initialStats.AvgRabbitHealth)
	t.Logf("  Средне здоровье волков: %.1f%%", initialStats.AvgWolfHealth)

	// Симулируем 15 секунд (900 тиков)
	const simulationTicks = 900
	const logInterval = 180 // Логируем каждые 3 секунды

	deltaTime := float32(1.0 / 60.0)

	t.Logf("\n--- НАЧАЛО СИМУЛЯЦИИ НА %d ТИКОВ (15 СЕКУНД) ---", simulationTicks)

	for tick := 0; tick < simulationTicks; tick++ {
		world.Update(deltaTime)
		systemManager.Update(world, deltaTime)

		// Логируем состояние каждые 3 секунды
		if tick%logInterval == 0 {
			stats := gatherEcosystemStats(world, animals)
			t.Logf("\n--- СОСТОЯНИЕ НА ТИКЕ %d (%.1f сек) ---", tick, float32(tick)/60.0)
			t.Logf("  Живых зайцев: %d/%d", stats.AliveRabbits, stats.TotalRabbits)
			t.Logf("  Живых волков: %d/%d", stats.AliveWolves, stats.TotalWolves)
			t.Logf("  Трупов: %d", stats.Corpses)
			t.Logf("  Средний голод зайцев: %.1f%%", stats.AvgRabbitHunger)
			t.Logf("  Средний голод волков: %.1f%%", stats.AvgWolfHunger)
			t.Logf("  Средне здоровье зайцев: %.1f%%", stats.AvgRabbitHealth)
			t.Logf("  Средне здоровье волков: %.1f%%", stats.AvgWolfHealth)

			// Проверки стабильности во время симуляции
			if stats.AliveRabbits == 0 {
				t.Logf("⚠️  Все зайцы мертвы на тике %d", tick)
			}
			if stats.AliveWolves == 0 {
				t.Logf("⚠️  Все волки мертвы на тике %d", tick)
			}

			// Проверка на аномальные значения
			if stats.AvgRabbitHunger > 100 || stats.AvgWolfHunger > 100 {
				t.Errorf("❌ Аномальные значения голода: зайцы %.1f%%, волки %.1f%%",
					stats.AvgRabbitHunger, stats.AvgWolfHunger)
			}
			if stats.AvgRabbitHealth > 100 || stats.AvgWolfHealth > 100 {
				t.Errorf("❌ Аномальные значения здоровья: зайцы %.1f%%, волки %.1f%%",
					stats.AvgRabbitHealth, stats.AvgWolfHealth)
			}
		}
	}

	// Финальный анализ
	finalStats := gatherEcosystemStats(world, animals)
	t.Logf("\n--- ФИНАЛЬНОЕ СОСТОЯНИЕ (15 СЕКУНД) ---")
	t.Logf("  Живых зайцев: %d/%d (%.1f%% выживаемость)",
		finalStats.AliveRabbits, finalStats.TotalRabbits,
		float32(finalStats.AliveRabbits)/float32(finalStats.TotalRabbits)*100)
	t.Logf("  Живых волков: %d/%d (%.1f%% выживаемость)",
		finalStats.AliveWolves, finalStats.TotalWolves,
		float32(finalStats.AliveWolves)/float32(finalStats.TotalWolves)*100)
	t.Logf("  Трупов: %d", finalStats.Corpses)

	// КРИТЕРИИ СТАБИЛЬНОСТИ

	// 1. Не должно быть полного вымирания одного вида
	if finalStats.AliveRabbits == 0 {
		t.Errorf("❌ НЕСТАБИЛЬНОСТЬ: Полное вымирание зайцев за 15 секунд")
	} else {
		t.Logf("✅ Зайцы выжили: %d особей", finalStats.AliveRabbits)
	}

	if finalStats.AliveWolves == 0 {
		t.Logf("⚠️  Все волки вымерли (может быть нормально если зайцы выжили)")
	} else {
		t.Logf("✅ Волки выжили: %d особей", finalStats.AliveWolves)
	}

	// 2. Значения голода и здоровья должны быть в разумных пределах
	if finalStats.AvgRabbitHunger < 0 || finalStats.AvgRabbitHunger > 100 {
		t.Errorf("❌ ОШИБКА: Аномальный голод зайцев: %.1f%%", finalStats.AvgRabbitHunger)
	}
	if finalStats.AvgWolfHunger < 0 || finalStats.AvgWolfHunger > 100 {
		t.Errorf("❌ ОШИБКА: Аномальный голод волков: %.1f%%", finalStats.AvgWolfHunger)
	}
	if finalStats.AvgRabbitHealth < 0 || finalStats.AvgRabbitHealth > 100 {
		t.Errorf("❌ ОШИБКА: Аномальное здоровье зайцев: %.1f%%", finalStats.AvgRabbitHealth)
	}
	if finalStats.AvgWolfHealth < 0 || finalStats.AvgWolfHealth > 100 {
		t.Errorf("❌ ОШИБКА: Аномальное здоровье волков: %.1f%%", finalStats.AvgWolfHealth)
	}

	// 3. Должны быть признаки активности экосистемы
	if finalStats.Corpses == 0 {
		t.Logf("⚠️  Нет трупов - возможно недостаточно взаимодействий")
	} else {
		t.Logf("✅ Экосистема активна: %d трупов", finalStats.Corpses)
	}

	// 4. Сравнение с начальным состоянием
	rabbitSurvival := float32(finalStats.AliveRabbits) / float32(initialStats.AliveRabbits) * 100
	wolfSurvival := float32(finalStats.AliveWolves) / float32(initialStats.AliveWolves) * 100

	t.Logf("\n--- ИТОГОВЫЙ АНАЛИЗ СТАБИЛЬНОСТИ ---")
	t.Logf("Выживаемость зайцев: %.1f%% (%d -> %d)",
		rabbitSurvival, initialStats.AliveRabbits, finalStats.AliveRabbits)
	t.Logf("Выживаемость волков: %.1f%% (%d -> %d)",
		wolfSurvival, initialStats.AliveWolves, finalStats.AliveWolves)

	// ОБЩАЯ ОЦЕНКА СТАБИЛЬНОСТИ
	stableEcosystem := true

	if finalStats.AliveRabbits == 0 && finalStats.AliveWolves == 0 {
		t.Errorf("❌ ЭКОСИСТЕМА НЕСТАБИЛЬНА: Полное вымирание всех видов")
		stableEcosystem = false
	}

	if finalStats.AvgRabbitHunger < 0 || finalStats.AvgRabbitHunger > 100 ||
		finalStats.AvgWolfHunger < 0 || finalStats.AvgWolfHunger > 100 {
		t.Errorf("❌ ЭКОСИСТЕМА НЕСТАБИЛЬНА: Аномальные значения голода")
		stableEcosystem = false
	}

	if finalStats.AvgRabbitHealth < 0 || finalStats.AvgRabbitHealth > 100 ||
		finalStats.AvgWolfHealth < 0 || finalStats.AvgWolfHealth > 100 {
		t.Errorf("❌ ЭКОСИСТЕМА НЕСТАБИЛЬНА: Аномальные значения здоровья")
		stableEcosystem = false
	}

	if stableEcosystem {
		t.Logf("✅ ЭКОСИСТЕМА СТАБИЛЬНА: Все показатели в норме")
	}

	t.Logf("\n=== ТЕСТ СТАБИЛЬНОСТИ ЗАВЕРШЕН ===")
}

// EcosystemStats структура для хранения статистики экосистемы
type EcosystemStats struct {
	TotalRabbits    int
	TotalWolves     int
	AliveRabbits    int
	AliveWolves     int
	Corpses         int
	AvgRabbitHunger float32
	AvgWolfHunger   float32
	AvgRabbitHealth float32
	AvgWolfHealth   float32
}

// gatherEcosystemStats собирает статистику экосистемы
func gatherEcosystemStats(world *core.World, animals []core.EntityID) EcosystemStats {
	stats := EcosystemStats{}

	rabbitHungerSum := float32(0)
	wolfHungerSum := float32(0)
	rabbitHealthSum := float32(0)
	wolfHealthSum := float32(0)

	for _, animal := range animals {
		animalType, hasType := world.GetAnimalType(animal)
		if !hasType {
			continue
		}

		isAlive := world.IsAlive(animal)

		if animalType == core.TypeRabbit {
			stats.TotalRabbits++
			if isAlive {
				stats.AliveRabbits++

				// Считаем средние значения только для живых
				if hunger, hasHunger := world.GetSatiation(animal); hasHunger {
					rabbitHungerSum += hunger.Value
				}
				if health, hasHealth := world.GetHealth(animal); hasHealth {
					rabbitHealthSum += float32(health.Current) / float32(health.Max) * 100
				}
			}
		} else if animalType == core.TypeWolf {
			stats.TotalWolves++
			if isAlive {
				stats.AliveWolves++

				// Считаем средние значения только для живых
				if hunger, hasHunger := world.GetSatiation(animal); hasHunger {
					wolfHungerSum += hunger.Value
				}
				if health, hasHealth := world.GetHealth(animal); hasHealth {
					wolfHealthSum += float32(health.Current) / float32(health.Max) * 100
				}
			}
		}

		// Считаем трупы
		if world.HasComponent(animal, core.MaskCorpse) {
			stats.Corpses++
		}
	}

	// Вычисляем средние значения
	if stats.AliveRabbits > 0 {
		stats.AvgRabbitHunger = rabbitHungerSum / float32(stats.AliveRabbits)
		stats.AvgRabbitHealth = rabbitHealthSum / float32(stats.AliveRabbits)
	}
	if stats.AliveWolves > 0 {
		stats.AvgWolfHunger = wolfHungerSum / float32(stats.AliveWolves)
		stats.AvgWolfHealth = wolfHealthSum / float32(stats.AliveWolves)
	}

	return stats
}
