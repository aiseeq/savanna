package config

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

// Config содержит все настройки симуляции
type Config struct {
	World      WorldConfig      `yaml:"world"`
	Terrain    TerrainConfig    `yaml:"terrain"`
	Population PopulationConfig `yaml:"population"`
}

// WorldConfig настройки мира
type WorldConfig struct {
	Size int   `yaml:"size"` // Размер мира в тайлах
	Seed int64 `yaml:"seed"` // Seed для детерминированной генерации
}

// TerrainConfig настройки ландшафта
type TerrainConfig struct {
	WaterBodies    int `yaml:"water_bodies"`     // Количество водоёмов
	WaterRadiusMin int `yaml:"water_radius_min"` // Минимальный радиус водоёма
	WaterRadiusMax int `yaml:"water_radius_max"` // Максимальный радиус водоёма
	BushClusters   int `yaml:"bush_clusters"`    // Количество групп кустов
	BushPerCluster int `yaml:"bush_per_cluster"` // Кустов в группе
}

// PopulationConfig настройки популяций животных
type PopulationConfig struct {
	Rabbits         int `yaml:"rabbits"`           // Количество зайцев
	Wolves          int `yaml:"wolves"`            // Количество волков
	RabbitGroupSize int `yaml:"rabbit_group_size"` // Размер группы зайцев
	MinWolfDistance int `yaml:"min_wolf_distance"` // Минимальная дистанция между волками
}

// LoadConfig загружает конфигурацию из файла
func LoadConfig(filename string) (*Config, error) {
	data, err := os.ReadFile(filename)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	var config Config
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("failed to parse config file: %w", err)
	}

	return &config, nil
}

// LoadDefaultConfig загружает конфигурацию по умолчанию
func LoadDefaultConfig() *Config {
	return &Config{
		World: WorldConfig{
			Size: 50,
			Seed: 42,
		},
		Terrain: TerrainConfig{
			WaterBodies:    3,
			WaterRadiusMin: 3,
			WaterRadiusMax: 5,
			BushClusters:   7,
			BushPerCluster: 5,
		},
		Population: PopulationConfig{
			Rabbits:         30,
			Wolves:          3,
			RabbitGroupSize: 3,
			MinWolfDistance: 20,
		},
	}
}

// SaveConfig сохраняет конфигурацию в файл
func SaveConfig(config *Config, filename string) error {
	data, err := yaml.Marshal(config)
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	if err := os.WriteFile(filename, data, 0644); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}

	return nil
}

// Validate проверяет корректность конфигурации
func (c *Config) Validate() error {
	if c.World.Size < 10 || c.World.Size > 200 {
		return fmt.Errorf("world size must be between 10 and 200, got %d", c.World.Size)
	}

	if c.Terrain.WaterBodies < 0 || c.Terrain.WaterBodies > 10 {
		return fmt.Errorf("water bodies count must be between 0 and 10, got %d", c.Terrain.WaterBodies)
	}

	if c.Terrain.WaterRadiusMin < 1 || c.Terrain.WaterRadiusMax > c.World.Size/2 {
		return fmt.Errorf("invalid water radius: min=%d, max=%d", c.Terrain.WaterRadiusMin, c.Terrain.WaterRadiusMax)
	}

	if c.Terrain.WaterRadiusMin > c.Terrain.WaterRadiusMax {
		return fmt.Errorf("water radius min (%d) cannot be greater than max (%d)",
			c.Terrain.WaterRadiusMin, c.Terrain.WaterRadiusMax)
	}

	if c.Population.Rabbits < 0 || c.Population.Wolves < 0 {
		return fmt.Errorf("population counts cannot be negative: rabbits=%d, wolves=%d",
			c.Population.Rabbits, c.Population.Wolves)
	}

	if c.Population.RabbitGroupSize < 1 {
		return fmt.Errorf("rabbit group size must be at least 1, got %d", c.Population.RabbitGroupSize)
	}

	return nil
}
