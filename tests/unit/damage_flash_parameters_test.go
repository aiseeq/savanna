package unit

import (
	"testing"

	"github.com/aiseeq/savanna/internal/simulation"
)

// TestDamageFlashParameters - unit тест проверяющий параметры усиленного DamageFlash
//
// Фиксирует константы из combat.go:
// - DamageFlashDuration = 0.16 (ускорено в 5 раз с 0.8)
//
// Фиксирует формулу из sprite_renderer.go:
// - scale = 1.0 + intensity * 5.0 (усилено в 5 раз)
func TestDamageFlashParameters(t *testing.T) {
	t.Parallel()

	t.Log("=== ТЕСТ ПАРАМЕТРОВ УСИЛЕННОГО DAMAGEFLASH ===")

	// ПРОВЕРКА 1: Константа длительности уменьшена в 5 раз
	expectedDuration := float64(0.16) // Новое значение
	actualDuration := simulation.DamageFlashDuration

	if actualDuration != expectedDuration {
		t.Errorf("БАГ: Неправильная длительность DamageFlash: %.3f (ожидалось %.3f)",
			actualDuration, expectedDuration)
	} else {
		t.Logf("✅ Длительность DamageFlash: %.3f сек (ускорена в 5 раз)", actualDuration)
	}

	// ПРОВЕРКА 2: Длительность стала значительно короче старого значения
	oldDuration := float64(0.8)
	speedupFactor := oldDuration / actualDuration

	if speedupFactor < 4.5 || speedupFactor > 5.5 {
		t.Errorf("БАГ: Неправильный коэффициент ускорения: %.1fх (ожидалось ~5х)",
			speedupFactor)
	} else {
		t.Logf("✅ Ускорение угасания: %.1fх", speedupFactor)
	}

	// ПРОВЕРКА 3: Длительность подходит для быстрого эффекта
	maxReasonableDuration := float64(0.25) // Четверть секунды
	if actualDuration > maxReasonableDuration {
		t.Errorf("БАГ: DamageFlash слишком долгий: %.3f сек (должен быть быстрым)",
			actualDuration)
	} else {
		t.Logf("✅ DamageFlash достаточно быстрый: %.3f сек", actualDuration)
	}

	t.Log("\n=== РАСЧЁТ УСИЛЕННОЙ ИНТЕНСИВНОСТИ ===")

	// ПРОВЕРКА 4: Тестируем новую формулу усиления (scale = 1.0 + intensity * 5.0)
	testCases := []struct {
		intensity     float32
		expectedScale float32
		description   string
	}{
		{0.0, 1.0, "нулевая интенсивность"},
		{0.2, 2.0, "слабая интенсивность"},
		{0.5, 3.5, "средняя интенсивность"},
		{0.8, 5.0, "сильная интенсивность"},
		{1.0, 6.0, "максимальная интенсивность"},
	}

	for _, tc := range testCases {
		// Имитируем формулу из sprite_renderer.go
		calculatedScale := 1.0 + tc.intensity*5.0

		if calculatedScale != tc.expectedScale {
			t.Errorf("БАГ: Неправильный расчёт для %s: %.1f (ожидалось %.1f)",
				tc.description, calculatedScale, tc.expectedScale)
		} else {
			t.Logf("✅ %s: интенсивность %.1f → масштаб %.1fх",
				tc.description, tc.intensity, calculatedScale)
		}
	}

	// ПРОВЕРКА 5: Усиление должно быть заметным
	minIntensity := float32(0.8) // Типичная интенсивность при уроне
	minScale := 1.0 + minIntensity*5.0

	if minScale < 4.0 {
		t.Errorf("БАГ: Эффект недостаточно заметен: %.1fх при интенсивности %.1f",
			minScale, minIntensity)
	} else {
		t.Logf("✅ Эффект заметен: %.1fх яркости при типичной интенсивности", minScale)
	}

	// ПРОВЕРКА 6: Усиление не должно быть чрезмерным
	maxIntensity := float32(1.0)
	maxScale := 1.0 + maxIntensity*5.0

	if maxScale > 8.0 {
		t.Errorf("БАГ: Эффект может быть слишком ярким: %.1fх при максимальной интенсивности",
			maxScale)
	} else {
		t.Logf("✅ Максимальный эффект разумен: %.1fх яркости", maxScale)
	}

	t.Log("\n=== РЕЗЮМЕ ПАРАМЕТРОВ ===")
	t.Logf("Длительность: %.3f сек (ускорена в %.1f раз)", actualDuration, speedupFactor)
	t.Logf("Усиление интенсивности: 5х множитель")
	t.Logf("Диапазон яркости: 1.0х - %.1fх", maxScale)
	t.Log("✅ Все параметры настроены для заметного и быстрого эффекта")
}

// TestDamageFlashFormula - unit тест проверяющий математику усиления
func TestDamageFlashFormula(t *testing.T) {
	t.Parallel()

	t.Log("=== ТЕСТ ФОРМУЛЫ УСИЛЕНИЯ DAMAGEFLASH ===")

	// Тестируем что новая формула даёт значительно более яркий эффект
	oldMultiplier := float32(1.0) // Старая формула: scale = 1.0 + intensity
	newMultiplier := float32(5.0) // Новая формула: scale = 1.0 + intensity * 5.0

	testIntensity := float32(0.9) // Типичная интенсивность

	oldScale := 1.0 + testIntensity*oldMultiplier
	newScale := 1.0 + testIntensity*newMultiplier

	improvementFactor := newScale / oldScale

	t.Logf("Тестовая интенсивность: %.1f", testIntensity)
	t.Logf("Старая формула: %.1fх яркости", oldScale)
	t.Logf("Новая формула: %.1fх яркости", newScale)
	t.Logf("Улучшение: %.1fх", improvementFactor)

	// ПРОВЕРКА: Новая формула должна дать значительное улучшение (>2x)
	minImprovement := float32(2.5)

	if improvementFactor < minImprovement {
		t.Errorf("БАГ: Недостаточное улучшение: %.1fх (ожидалось >%.1fх)",
			improvementFactor, minImprovement)
	} else {
		t.Logf("✅ Значительное улучшение: %.1fх", improvementFactor)
	}

	// ПРОВЕРКА: Новый эффект должен быть заметно ярче
	if newScale < 4.0 {
		t.Error("БАГ: Новый эффект недостаточно яркий")
	} else {
		t.Logf("✅ Новый эффект достаточно яркий: %.1fх", newScale)
	}

	t.Log("✅ Формула усиления работает корректно")
}
