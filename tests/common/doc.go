// Package common содержит общую инфраструктуру для тестов
// Устраняет дублирование кода создания миров, систем и сущностей в 80+ тестах
//
// Основные компоненты:
//
// 1. constants.go - Общие константы (размеры миров, семена, пороги)
// 2. world_builder.go - TestWorldBuilder для создания тестовых миров (Builder Pattern)
// 3. system_factory.go - Фабрики для создания наборов систем
// 4. simulation_utils.go - Утилиты для запуска симуляции и проверок
//
// Применяемые принципы:
// - DRY: Вынос дублированного кода в переиспользуемые компоненты
// - SOLID: Builder Pattern, Factory Pattern, Single Responsibility
// - KISS: Простые и понятные API для создания тестов
//
// Пример использования:
//
//	world, systems, entities := NewTestWorld().
//		WithLargeSize().
//		AddHungryRabbit().
//		AddHungryWolf().
//		Build()
//
//	RunSimulation(world, systems, FiveSecondTicks)
//	AssertRabbitDamaged(t, world, entities.Rabbits[0])
package common
