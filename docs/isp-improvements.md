# ISP Улучшения в Savanna ECS

## Принцип разделения интерфейсов (Interface Segregation Principle)

Цель ISP рефакторинга: создать узкоспециализированные интерфейсы, чтобы системы зависели только от методов, которые они реально используют.

## До рефакторинга

```go
// ❌ НАРУШЕНИЕ ISP: Слишком широкий интерфейс
type SimulationAccess interface {
    ComponentReader // 15+ методов чтения
    ComponentWriter // 20+ методов записи  
    QueryProvider   // 5+ методов запросов
    SpatialQueries  // 4+ методов пространственного поиска
    WorldInfo       // 4+ метода состояния мира
}

// HungerSystem использует только 3 метода из 45+
func (hs *HungerSystem) Update(world SimulationAccess, deltaTime float32) {
    world.GetHunger(entity)    // ИСПОЛЬЗУЕТСЯ
    world.SetHunger(entity)    // ИСПОЛЬЗУЕТСЯ
    world.ForEachWith(mask)    // ИСПОЛЬЗУЕТСЯ
    // world.GetAnimation()    // НЕ ИСПОЛЬЗУЕТСЯ
    // world.AddAttackState()  // НЕ ИСПОЛЬЗУЕТСЯ
    // world.QueryInRadius()   // НЕ ИСПОЛЬЗУЕТСЯ
    // ... еще 40+ неиспользуемых методов
}
```

## После рефакторинга

```go
// ✅ СОБЛЮДЕНИЕ ISP: Узкоспециализированные интерфейсы

// HungerSystemAccess содержит ТОЛЬКО нужные методы
type HungerSystemAccess interface {
    GetHunger(EntityID) (Hunger, bool)  // Чтение голода
    GetSize(EntityID) (Size, bool)      // Для расчёта скорости голода
    SetHunger(EntityID, Hunger) bool    // Изменение голода
    ForEachWith(ComponentMask, QueryFunc) // Итерация
}

// HungerSystem зависит только от того, что использует
func (hs *HungerSystem) Update(world HungerSystemAccess, deltaTime float32) {
    world.GetHunger(entity)    // ВСЕ методы используются
    world.SetHunger(entity)    // ВСЕ методы используются  
    world.ForEachWith(mask)    // ВСЕ методы используются
}
```

## Преимущества ISP улучшений

### 1. Ясность зависимостей
- **До**: HungerSystem зависит от 45+ методов, используя только 4
- **После**: HungerSystem зависит от 4 методов, используя все 4

### 2. Легкость тестирования
```go
// ✅ Простой мок для тестирования HungerSystem
type MockHungerSystemAccess struct {
    hungers map[EntityID]Hunger
}

func (m *MockHungerSystemAccess) GetHunger(id EntityID) (Hunger, bool) {
    h, ok := m.hungers[id]
    return h, ok
}
// Только 4 метода вместо 45+
```

### 3. Принцип наименьшего удивления (PoLA)
- Имена интерфейсов ясно указывают на назначение
- `HungerSystemAccess` - только для работы с голодом
- `GrassSearchSystemAccess` - только для поиска травы

### 4. Соответствие SOLID принципам
- **S (SRP)**: Каждая система имеет одну ответственность
- **I (ISP)**: Системы зависят только от нужных интерфейсов
- **D (DIP)**: Системы зависят от абстракций, не от конкретики

## Созданные узкоспециализированные интерфейсы

1. **HungerSystemAccess** - управление голодом (4 метода)
2. **GrassSearchSystemAccess** - поиск травы (7 методов)  
3. **StarvationDamageSystemAccess** - урон от голода (4 метода)
4. **HungerSpeedModifierSystemAccess** - влияние голода на скорость (5 методов)

## Обратная совместимость

Сохранены широкие интерфейсы для сложных систем:
- `SimulationAccess` - для систем, которым нужно много функций
- `ECSAccess` - deprecated, для обратной совместимости

## Результат

- ✅ Соблюдение ISP принципа
- ✅ Улучшенная тестируемость
- ✅ Ясность архитектуры
- ✅ Обратная совместимость
- ✅ Простота понимания зависимостей