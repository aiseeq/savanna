package animation

import "github.com/hajimehoshi/ebiten/v2"

// AnimationConfig конфигурация анимации (устраняет дублирование параметров)
type AnimationConfig struct {
	Frames   int
	FPS      float64
	Loop     bool
	AnimType AnimationType
}

// Константы анимаций (устраняет магические числа)
const (
	// Количество кадров
	IdleFrameCount   = 2
	WalkFrameCount   = 4
	RunFrameCount    = 4
	AttackFrameCount = 2
	EatFrameCount    = 2
	DeathFrameCount  = 1

	// Скорости анимаций (FPS)
	IdleFPS   = 1.0  // Медленная анимация для покоя
	WalkFPS   = 8.0  // Базовая скорость ходьбы
	RunFPS    = 12.0 // Скорость бега (быстрее ходьбы)
	AttackFPS = 6.0  // Скорость атаки
	EatFPS    = 2.0  // Скорость поедания
	DeathFPS  = 1.0  // Скорость смерти (статичная)
)

// StandardAnimationConfigs стандартные конфигурации анимаций (устраняет дублирование)
var StandardAnimationConfigs = map[AnimationType]AnimationConfig{
	AnimIdle: {
		Frames:   IdleFrameCount,
		FPS:      IdleFPS,
		Loop:     true,
		AnimType: AnimIdle,
	},
	AnimWalk: {
		Frames:   WalkFrameCount,
		FPS:      WalkFPS,
		Loop:     true,
		AnimType: AnimWalk,
	},
	AnimRun: {
		Frames:   RunFrameCount,
		FPS:      RunFPS,
		Loop:     true,
		AnimType: AnimRun,
	},
	AnimAttack: {
		Frames:   AttackFrameCount,
		FPS:      AttackFPS,
		Loop:     false, // НЕ зацикленная анимация
		AnimType: AnimAttack,
	},
	AnimEat: {
		Frames:   EatFrameCount,
		FPS:      EatFPS,
		Loop:     true,
		AnimType: AnimEat,
	},
	AnimDeathDying: {
		Frames:   DeathFrameCount,
		FPS:      DeathFPS,
		Loop:     false,
		AnimType: AnimDeathDying,
	},
}

// AnimationLoader загрузчик анимаций
type AnimationLoader struct{}

// NewAnimationLoader создаёт новый загрузчик анимаций
func NewAnimationLoader() *AnimationLoader {
	return &AnimationLoader{}
}

// LoadWolfAnimations загружает стандартные анимации волка (устраняет дублирование)
func (al *AnimationLoader) LoadWolfAnimations(animSystem *AnimationSystem, spriteImage *ebiten.Image) {
	// Определяем какие анимации нужны волку
	wolfAnimations := []AnimationType{
		AnimIdle,
		AnimWalk,
		AnimRun,
		AnimAttack,
		AnimEat,
	}

	// Загружаем каждую анимацию
	for _, animType := range wolfAnimations {
		config := StandardAnimationConfigs[animType]
		animSystem.RegisterAnimation(animType, config.Frames, float32(config.FPS), config.Loop, spriteImage)
	}
}

// LoadRabbitAnimations загружает стандартные анимации зайца (устраняет дублирование)
func (al *AnimationLoader) LoadRabbitAnimations(animSystem *AnimationSystem, spriteImage *ebiten.Image) {
	// Определяем какие анимации нужны зайцу
	rabbitAnimations := []AnimationType{
		AnimIdle,
		AnimWalk,
		AnimRun,
		AnimEat, // ИСПРАВЛЕНИЕ: зайцы тоже едят траву!
		AnimDeathDying,
	}

	// Загружаем каждую анимацию
	for _, animType := range rabbitAnimations {
		config := StandardAnimationConfigs[animType]
		animSystem.RegisterAnimation(animType, config.Frames, float32(config.FPS), config.Loop, spriteImage)
	}
}

// LoadAnimations загружает анимации с реальными спрайтами
func (al *AnimationLoader) LoadAnimations(
	wolfSystem, rabbitSystem *AnimationSystem,
	wolfSprite, rabbitSprite *ebiten.Image,
) {
	al.LoadWolfAnimations(wolfSystem, wolfSprite)
	al.LoadRabbitAnimations(rabbitSystem, rabbitSprite)
}
