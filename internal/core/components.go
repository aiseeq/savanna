package core

// Базовые компоненты для симуляции экосистемы

// Position позиция сущности в мире
type Position struct {
	X, Y float32
}

// Velocity скорость движения сущности
type Velocity struct {
	X, Y float32
}

// Health здоровье сущности
type Health struct {
	Current int16 // Текущее здоровье
	Max     int16 // Максимальное здоровье
}

// Hunger уровень голода (0-100)
type Hunger struct {
	Value float32 // 0 = умирает от голода, 100 = сыт
}

// Age возраст сущности в секундах
type Age struct {
	Seconds float32
}

// AnimalType тип животного
type AnimalType uint8

const (
	TypeNone   AnimalType = iota // Отсутствует (для неживых объектов)
	TypeRabbit                   // Заяц (травоядное)
	TypeWolf                     // Волк (хищник)
	TypeGrass                    // Трава (для будущего расширения)
)

// String возвращает строковое представление типа животного
func (at AnimalType) String() string {
	switch at {
	case TypeRabbit:
		return "Rabbit"
	case TypeWolf:
		return "Wolf"
	case TypeGrass:
		return "Grass"
	default:
		return "None"
	}
}

// Size размер сущности (радиус для коллизий)
type Size struct {
	Radius float32
}

// Speed скорости движения
type Speed struct {
	Base    float32 // Базовая скорость
	Current float32 // Текущая скорость (с модификаторами)
}

// ComponentMask битовые маски для быстрой проверки наличия компонентов
type ComponentMask uint64

const (
	MaskPosition ComponentMask = 1 << iota
	MaskVelocity
	MaskHealth
	MaskHunger
	MaskAge
	MaskAnimalType
	MaskSize
	MaskSpeed
)

// HasComponent проверяет наличие компонента в маске
func (mask ComponentMask) HasComponent(component ComponentMask) bool {
	return mask&component != 0
}

// AddComponent добавляет компонент к маске
func (mask ComponentMask) AddComponent(component ComponentMask) ComponentMask {
	return mask | component
}

// RemoveComponent удаляет компонент из маски
func (mask ComponentMask) RemoveComponent(component ComponentMask) ComponentMask {
	return mask &^ component
}

// ComponentSet вспомогательная структура для работы с наборами компонентов
type ComponentSet struct {
	mask ComponentMask
}

// NewComponentSet создаёт новый набор компонентов
func NewComponentSet(components ...ComponentMask) ComponentSet {
	var mask ComponentMask
	for _, component := range components {
		mask = mask.AddComponent(component)
	}
	return ComponentSet{mask: mask}
}

// Has проверяет наличие компонента
func (cs ComponentSet) Has(component ComponentMask) bool {
	return cs.mask.HasComponent(component)
}

// HasAll проверяет наличие всех указанных компонентов
func (cs ComponentSet) HasAll(components ComponentMask) bool {
	return cs.mask&components == components
}

// Add добавляет компонент
func (cs *ComponentSet) Add(component ComponentMask) {
	cs.mask = cs.mask.AddComponent(component)
}

// Remove удаляет компонент
func (cs *ComponentSet) Remove(component ComponentMask) {
	cs.mask = cs.mask.RemoveComponent(component)
}

// Mask возвращает битовую маску
func (cs ComponentSet) Mask() ComponentMask {
	return cs.mask
}

// Clear очищает набор компонентов
func (cs *ComponentSet) Clear() {
	cs.mask = 0
}

// Предопределённые наборы компонентов для быстрых запросов

var (
	// MovingEntities сущности которые могут двигаться
	MovingEntities = NewComponentSet(MaskPosition, MaskVelocity)

	// LivingEntities живые сущности
	LivingEntities = NewComponentSet(MaskPosition, MaskHealth, MaskHunger, MaskAge, MaskAnimalType, MaskSize)

	// AnimalsEntities животные (для рендеринга и логики)
	AnimalsEntities = NewComponentSet(MaskPosition, MaskVelocity, MaskHealth, MaskHunger, MaskAnimalType, MaskSize, MaskSpeed)
)
