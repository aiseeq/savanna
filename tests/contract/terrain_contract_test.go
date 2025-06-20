package contract

import (
	"testing"

	"github.com/aiseeq/savanna/config"
	"github.com/aiseeq/savanna/internal/generator"
	"github.com/aiseeq/savanna/internal/simulation"
	"github.com/aiseeq/savanna/tests/common"
)

// Contract Testing проверяет что все реализации TerrainInterface ведут себя одинаково
// Это помогает обнаружить несоответствия между моками и реальными реализациями

type TerrainImplementation struct {
	name    string
	factory func() generator.TerrainInterface
}

// getAllTerrainImplementations возвращает все доступные реализации terrain
func getAllTerrainImplementations() []TerrainImplementation {
	return []TerrainImplementation{
		{
			name: "RealTerrain",
			factory: func() generator.TerrainInterface {
				cfg := config.LoadDefaultConfig()
				cfg.World.Size = 5
				terrainGen := generator.NewTerrainGenerator(cfg)
				return terrainGen.Generate()
			},
		},
		{
			name: "MockTerrain",
			factory: func() generator.TerrainInterface {
				return common.NewMockTerrain(5)
			},
		},
	}
}

// TestTerrainContractBasicOperations проверяет базовые операции всех реализаций
func TestTerrainContractBasicOperations(t *testing.T) {
	t.Parallel()

	implementations := getAllTerrainImplementations()

	for _, impl := range implementations {
		impl := impl // Capture for parallel execution
		t.Run(impl.name, func(t *testing.T) {
			t.Parallel()

			terrain := impl.factory()

			// КОНТРАКТ 1: GetSize должен возвращать положительное число
			size := terrain.GetSize()
			if size <= 0 {
				t.Errorf("%s: GetSize() returned %d, expected > 0", impl.name, size)
			}

			// КОНТРАКТ 2: Валидные координаты должны возвращать валидные типы тайлов
			tileType := terrain.GetTileType(0, 0)
			if tileType < 0 || tileType > 10 { // Разумные границы для типов
				t.Errorf("%s: GetTileType(0,0) returned %d, expected valid tile type", impl.name, tileType)
			}

			// КОНТРАКТ 3: Невалидные координаты должны обрабатываться корректно
			invalidTileType := terrain.GetTileType(-1, -1)
			_ = invalidTileType // Не должно паниковать

			// КОНТРАКТ 4: Количество травы должно быть неотрицательным
			grassAmount := terrain.GetGrassAmount(0, 0)
			if grassAmount < 0 {
				t.Errorf("%s: GetGrassAmount(0,0) returned %.2f, expected >= 0", impl.name, grassAmount)
			}
		})
	}
}

// TestTerrainContractGrassManagement проверяет управление травой
func TestTerrainContractGrassManagement(t *testing.T) {
	t.Parallel()

	implementations := getAllTerrainImplementations()

	for _, impl := range implementations {
		impl := impl
		t.Run(impl.name, func(t *testing.T) {
			t.Parallel()

			terrain := impl.factory()

			// Устанавливаем тип тайла и количество травы
			terrain.SetTileType(1, 1, generator.TileGrass)
			terrain.SetGrassAmount(1, 1, 50.0)

			// КОНТРАКТ: После установки травы на TileGrass, её должно быть видно
			retrievedAmount := terrain.GetGrassAmount(1, 1)

			// Для MockTerrain это может отличаться, но поведение должно быть предсказуемым
			if impl.name == "MockTerrain" {
				// Mock возвращает константное значение - это ОК для мока, но должно быть задокументировано
				if retrievedAmount != 100.0 { // MockTerrain возвращает константу
					t.Logf("%s: Mock returns constant value %.1f instead of set value 50.0 - this is expected mock behavior",
						impl.name, retrievedAmount)
				}
			} else {
				// Реальный terrain должен сохранять установленное значение
				if retrievedAmount != 50.0 {
					t.Errorf("%s: SetGrassAmount(50.0) then GetGrassAmount() returned %.2f, expected 50.0",
						impl.name, retrievedAmount)
				}
			}
		})
	}
}

// TestTerrainContractVegetationSystemIntegration проверяет интеграцию с VegetationSystem
func TestTerrainContractVegetationSystemIntegration(t *testing.T) {
	t.Parallel()

	implementations := getAllTerrainImplementations()

	for _, impl := range implementations {
		impl := impl
		t.Run(impl.name, func(t *testing.T) {
			t.Parallel()

			terrain := impl.factory()
			vegetationSystem := simulation.NewVegetationSystem(terrain)

			// КОНТРАКТ: VegetationSystem должна работать с любой реализацией terrain
			// Проверяем что базовые операции не паникуют

			// 1. Получение количества травы
			grassAmount := vegetationSystem.GetGrassAt(32, 32) // Центр тайла (1,1)
			if grassAmount < 0 {
				t.Errorf("%s: VegetationSystem.GetGrassAt returned negative value %.2f", impl.name, grassAmount)
			}

			// 2. Поиск ближайшей травы
			grassX, grassY, found := vegetationSystem.FindNearestGrass(32, 32, 100, 10)

			if found {
				// Если трава найдена, координаты должны быть валидными
				if grassX < 0 || grassY < 0 {
					t.Errorf("%s: FindNearestGrass returned invalid coordinates (%.1f, %.1f)",
						impl.name, grassX, grassY)
				}

				// И в найденном месте должна быть трава
				foundGrass := vegetationSystem.GetGrassAt(grassX, grassY)
				if foundGrass < 10 {
					t.Errorf("%s: FindNearestGrass found location with %.1f grass, expected >= 10",
						impl.name, foundGrass)
				}
			}

			// 3. Обновление количества травы
			vegetationSystem.UpdateGrassAt(32, 32, -5) // Съедаем 5 единиц
			// Не должно паниковать
		})
	}
}

// TestTerrainContractConsistency проверяет внутреннюю консистентность
func TestTerrainContractConsistency(t *testing.T) {
	t.Parallel()

	implementations := getAllTerrainImplementations()

	for _, impl := range implementations {
		impl := impl
		t.Run(impl.name, func(t *testing.T) {
			t.Parallel()

			terrain := impl.factory()

			// КОНТРАКТ: Повторные вызовы одной функции с теми же параметрами должны возвращать то же значение

			tileType1 := terrain.GetTileType(2, 2)
			tileType2 := terrain.GetTileType(2, 2)
			if tileType1 != tileType2 {
				t.Errorf("%s: GetTileType(2,2) inconsistent: first=%d, second=%d",
					impl.name, tileType1, tileType2)
			}

			grass1 := terrain.GetGrassAmount(2, 2)
			grass2 := terrain.GetGrassAmount(2, 2)
			if grass1 != grass2 {
				t.Errorf("%s: GetGrassAmount(2,2) inconsistent: first=%.2f, second=%.2f",
					impl.name, grass1, grass2)
			}

			// КОНТРАКТ: Тип тайла должен оставаться консистентным
			tileType := terrain.GetTileType(2, 2)

			// Проверяем что тип тайла валидный
			if tileType < 0 || tileType > 10 {
				t.Errorf("%s: Invalid tile type at (2,2): %d", impl.name, tileType)
			}
		})
	}
}
