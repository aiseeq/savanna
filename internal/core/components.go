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

// Satiation уровень сытости (0-100)
type Satiation struct {
	Value float32 // 0 = умирает от голода, 100 = сыт
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

// Size размер сущности (радиус для коллизий и атак)
type Size struct {
	Radius      float32 // Радиус коллизий
	AttackRange float32 // Дальность атаки (0 для мирных животных)
}

// Speed скорости движения
type Speed struct {
	Base    float32 // Базовая скорость
	Current float32 // Текущая скорость (с модификаторами)
}

// Animation анимационный компонент
type Animation struct {
	CurrentAnim int     // Текущий тип анимации (из animation.AnimationType)
	Frame       int     // Текущий кадр (0-based)
	Timer       float32 // Таймер для смены кадров
	Playing     bool    // Проигрывается ли анимация
	FacingRight bool    // Смотрит ли вправо (для отражения спрайта)
}

// DamageFlash эффект мигания при получении урона
type DamageFlash struct {
	Timer     float32 // Оставшееся время эффекта
	Duration  float32 // Общая длительность эффекта
	Intensity float32 // Интенсивность мигания (0.0-1.0)
}

// Corpse компонент для мёртвых тел
type Corpse struct {
	NutritionalValue float32 // Сколько питательности осталось
	MaxNutritional   float32 // Изначальная питательность
	DecayTimer       float32 // Время до разложения (исчезновения)
}

// Carrion падаль (брошенная хищником туша)
type Carrion struct {
	NutritionalValue float32  // Сколько питательности осталось
	MaxNutritional   float32  // Изначальная питательность
	DecayTimer       float32  // Время до разложения (исчезновения)
	AbandonedBy      EntityID // Кто бросил эту падаль
}

// EatingTargetType тип цели поедания (устраняет магическое значение Target=0)
type EatingTargetType uint8

const (
	EatingTargetGrass  EatingTargetType = iota // Поедание травы
	EatingTargetAnimal                         // Поедание животного/трупа
)

// EatingState состояние поедания
type EatingState struct {
	Target          EntityID         // Кого/что едим: EntityID животного или 0 для поедания травы
	TargetType      EatingTargetType // Тип цели: трава или животное
	EatingProgress  float32          // Прогресс поедания (0-1)
	NutritionGained float32          // Сколько уже получили питательности
}

// AttackPhase фаза атаки
type AttackPhase uint8

const (
	AttackPhaseWindup AttackPhase = iota // Замах (кадр 0)
	AttackPhaseStrike                    // Удар (кадр 1)
)

// String возвращает строковое представление фазы атаки
func (ap AttackPhase) String() string {
	switch ap {
	case AttackPhaseWindup:
		return "Windup"
	case AttackPhaseStrike:
		return "Strike"
	default:
		return "Unknown"
	}
}

// AttackState состояние атаки
type AttackState struct {
	Target     EntityID    // Цель атаки
	Phase      AttackPhase // Текущая фаза атаки
	PhaseTimer float32     // Время в текущей фазе
	TotalTimer float32     // Общее время атаки
	HasStruck  bool        // Был ли нанесен удар в этой атаке
}

// BehaviorType тип поведения животного
type BehaviorType uint8

const (
	BehaviorNone      BehaviorType = iota // Нет поведения
	BehaviorHerbivore                     // Травоядное (ищет траву, убегает от хищников)
	BehaviorPredator                      // Хищник (охотится на других животных)
	// УДАЛЕНО: BehaviorScavenger - не используется в игре (нет падальщиков)
)

// String возвращает строковое представление типа поведения
func (bt BehaviorType) String() string {
	switch bt {
	case BehaviorHerbivore:
		return "Herbivore"
	case BehaviorPredator:
		return "Predator"
	// УДАЛЕНО: BehaviorScavenger
	default:
		return "None"
	}
}

// AnimalConfig конфигурация животного (заменяет захардкоженные константы)
// Устраняет нарушения SOLID: теперь поведение НЕ зависит от типа животного
type AnimalConfig struct {
	// Базовые параметры
	BaseRadius float32 // Базовый радиус (от него выводятся все остальные размеры)
	MaxHealth  int16   // Максимальное здоровье
	BaseSpeed  float32 // Базовая скорость движения

	// Размеры и дистанции (выводятся от BaseRadius через множители)
	CollisionRadius float32 // Радиус коллизий (обычно = BaseRadius)
	AttackRange     float32 // Дальность атаки (BaseRadius * множитель)
	VisionRange     float32 // Дальность видения (BaseRadius * множитель)

	// Поведение
	SatiationThreshold float32 // При какой сытости начинает искать еду
	FleeThreshold      float32 // Дистанция на которой убегает от угрозы

	// Множители скорости в разных состояниях
	SearchSpeed    float32 // Множитель скорости при поиске еды (0.8)
	WanderingSpeed float32 // Множитель скорости при блуждании (0.7)
	ContentSpeed   float32 // Множитель скорости в покое (0.3)

	// Таймеры поведения
	MinDirectionTime float32 // Минимальное время случайного движения
	MaxDirectionTime float32 // Максимальное время случайного движения

	// Боевые характеристики
	AttackDamage   int16   // Урон атаки
	AttackCooldown float32 // Кулдаун между атаками
	HitChance      float32 // Шанс попадания (0.0-1.0)
}

// Behavior поведение животного
type Behavior struct {
	Type               BehaviorType // Тип поведения
	DirectionTimer     float32      // Таймер смены направления движения
	SatiationThreshold float32      // При какой сытости начинает искать еду
	FleeThreshold      float32      // Дистанция на которой убегает от угрозы
	SearchSpeed        float32      // Множитель скорости при поиске еды (0.8)
	WanderingSpeed     float32      // Множитель скорости при блуждании (0.7)
	ContentSpeed       float32      // Множитель скорости в покое (0.3)
	VisionRange        float32      // Дальность видения
	MinDirectionTime   float32      // Минимальное время случайного движения
	MaxDirectionTime   float32      // Максимальное время случайного движения
}

// ComponentMask битовые маски для быстрой проверки наличия компонентов
type ComponentMask uint64

const (
	MaskPosition ComponentMask = 1 << iota
	MaskVelocity
	MaskHealth
	MaskSatiation
	MaskAnimalType
	MaskSize
	MaskSpeed
	MaskAnimation
	MaskDamageFlash
	MaskCorpse
	MaskCarrion
	MaskEatingState
	MaskAttackState
	MaskBehavior
	MaskAnimalConfig
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
	LivingEntities = NewComponentSet(MaskPosition, MaskHealth, MaskSatiation, MaskAnimalType, MaskSize)

	// AnimalsEntities животные (для рендеринга и логики)
	AnimalsEntities = NewComponentSet(
		MaskPosition, MaskVelocity, MaskHealth, MaskSatiation,
		MaskAnimalType, MaskSize, MaskSpeed, MaskAnimalConfig,
	)
)
