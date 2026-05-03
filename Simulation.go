package main

import (
	"math"
	"math/rand"
)

// SimConfig хранит параметры симуляции, задаваемые пользователем.
type SimConfig struct {
	Count  int     // количество частиц [2..100]
	Speed  float64 // начальная скорость (пикс./кадр)
	Radius float64 // радиус частицы (пикс.)
}

// DefaultConfig возвращает конфигурацию по умолчанию.
func DefaultConfig() SimConfig {
	return SimConfig{
		Count:  30,
		Speed:  2.5,
		Radius: 9,
	}
}

// Simulation содержит состояние симуляции.
type Simulation struct {
	Particles  []*Particle
	Width      float64
	Height     float64
	Collisions int  // счётчик столкновений частица–частица
	Running    bool // true — симуляция запущена
	Config     SimConfig
}

// NewSimulation создаёт симуляцию с заданными параметрами.
func NewSimulation(w, h float64, cfg SimConfig) *Simulation {
	s := &Simulation{
		Width:   w,
		Height:  h,
		Running: true,
		Config:  cfg,
	}
	s.Reset(cfg)
	return s
}

// Reset пересоздаёт частицы с новой конфигурацией.
func (s *Simulation) Reset(cfg SimConfig) {
	s.Config = cfg
	s.Collisions = 0
	s.Particles = make([]*Particle, 0, cfg.Count)

	for i := 0; i < cfg.Count; i++ {
		p := s.placeParticle(cfg.Speed, cfg.Radius)
		s.Particles = append(s.Particles, p)
	}
}

// placeParticle пытается разместить частицу без перекрытия с уже созданными.
func (s *Simulation) placeParticle(speed, radius float64) *Particle {
	const maxAttempts = 300
	for attempt := 0; attempt < maxAttempts; attempt++ {
		x := radius + rand.Float64()*(s.Width-2*radius)
		y := radius + rand.Float64()*(s.Height-2*radius)

		overlaps := false
		for _, other := range s.Particles {
			if math.Hypot(other.X-x, other.Y-y) < (radius+other.R)*1.05 {
				overlaps = true
				break
			}
		}
		if !overlaps {
			return NewParticle(x, y, speed, radius)
		}
	}
	// Если не удалось разместить без перекрытия — ставим в центр области
	return NewParticle(s.Width/2, s.Height/2, speed, radius)
}

// Update выполняет один шаг физической симуляции.
func (s *Simulation) Update() {
	if !s.Running {
		return
	}

	// Перемещение всех частиц
	for _, p := range s.Particles {
		p.Move()
	}

	// Отражение от стен
	for _, p := range s.Particles {
		p.ReflectWalls(s.Width, s.Height)
	}

	// Обнаружение и разрешение парных столкновений.
	// Одновременный контакт ≥3 частиц обрабатывается последовательно для каждой пары
	// (упрощение, допустимое по разделу 1.5 анализа предметной области).
	for i := 0; i < len(s.Particles); i++ {
		for j := i + 1; j < len(s.Particles); j++ {
			if ResolveCollision(s.Particles[i], s.Particles[j]) {
				s.Collisions++
			}
		}
	}
}

// TotalKineticEnergy возвращает суммарную кинетическую энергию системы
// (в условных единицах, масса = 1).
func (s *Simulation) TotalKineticEnergy() float64 {
	var ke float64
	for _, p := range s.Particles {
		v := p.Speed()
		ke += 0.5 * v * v
	}
	return ke
}

// TogglePause переключает состояние паузы.
func (s *Simulation) TogglePause() {
	s.Running = !s.Running
}
