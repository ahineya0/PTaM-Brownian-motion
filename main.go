// TODO (финальная версия):
//   справочная система CHM, вызов по F1
//   поле ввода точного числа частиц
//   тест 10-минутной непрерывной работы
//   корректная обработка ≥3 одновременных столкновений

package main

import (
	"fmt"
	"image/color"
	"log"
	"os"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	"github.com/hajimehoshi/ebiten/v2/text"
	"github.com/hajimehoshi/ebiten/v2/vector"

	"golang.org/x/image/font"
	"golang.org/x/image/font/opentype"
)

// Размеры окна fddf
const (
	WindowWidth  = 960
	WindowHeight = 660

	// Ширина правой панели управления
	PanelWidth = 220

	// Отступы внутри панели
	PanelPad = 14

	// Размеры области симуляции
	SimOffX = 0
	SimOffY = 0
	SimW    = WindowWidth - PanelWidth
	SimH    = WindowHeight
)

var face font.Face

func loadFont() {
	data, err := os.ReadFile("assets/DejaVuSans.ttf")
	if err != nil {
		log.Fatal(err)
	}

	tt, err := opentype.Parse(data)
	if err != nil {
		log.Fatal(err)
	}

	face, err = opentype.NewFace(tt, &opentype.FaceOptions{
		Size:    14,
		DPI:     72,
		Hinting: font.HintingFull,
	})
	if err != nil {
		log.Fatal(err)
	}
}

// Game реализует интерфейс ebiten.Game.
type Game struct {
	sim *Simulation

	// Ползунки панели управления
	sliderCount  *Slider
	sliderSpeed  *Slider
	sliderRadius *Slider

	// Кнопки
	btnPause *Button
	btnReset *Button
	// TODO (прототип): btnHelp открывает CHM-справку через os/exec

	// Флаг отображения векторов скоростей
	showVectors bool

	// Счётчик FPS (среднее за последние 60 кадров)
	fpsTick int
	fps     float64
}

// NewGame инициализирует игру.
func NewGame() *Game {
	cfg := DefaultConfig()
	sim := NewSimulation(SimW, SimH, cfg)

	px := SimW + PanelPad
	pw := PanelWidth - PanelPad*2

	g := &Game{
		sim:          sim,
		sliderCount:  NewSlider(px, 80, pw, "Частицы", 2, 100, cfg.Count),
		sliderSpeed:  NewSlider(px, 150, pw, "Скорость", 1, 8, int(cfg.Speed*2)),
		sliderRadius: NewSlider(px, 220, pw, "Радиус", 4, 18, int(cfg.Radius)),
		btnPause:     NewButton(px, 280, pw, 32, "[ ПАУЗА ]"),
		btnReset:     NewButton(px, 322, pw, 32, "[ СБРОС ]"),
	}
	return g
}

// currentConfig собирает конфигурацию из текущих значений ползунков.
func (g *Game) currentConfig() SimConfig {
	return SimConfig{
		Count:  g.sliderCount.Value,
		Speed:  float64(g.sliderSpeed.Value) / 2.0,
		Radius: float64(g.sliderRadius.Value),
	}
}

// Update вызывается Ebitengine каждый тик (60 раз/с).
func (g *Game) Update() error {
	// Клавиша ESC — завершение программы (§4.1 ТЗ: корректное завершение)
	if ebiten.IsKeyPressed(ebiten.KeyEscape) {
		return ebiten.Termination
	}

	// Пробел — пауза/продолжение
	if inpututil.IsKeyJustPressed(ebiten.KeySpace) {
		g.sim.TogglePause()
		if g.sim.Running {
			g.btnPause.Label = "[ Pause ]"
		} else {
			g.btnPause.Label = "[ Start ]"
		}
	}

	// V — показать/скрыть векторы скоростей
	if inpututil.IsKeyJustPressed(ebiten.KeyV) {
		g.showVectors = !g.showVectors
	}

	// F1 TODO (прототип) открыть CHM-справку
	if inpututil.IsKeyJustPressed(ebiten.KeyF1) {
		// TODO: exec.Command("hh.exe", "brownian.chm").Start() — Windows
		// TODO: xdg-open / gio open — Linux
		log.Println("[TODO] Справочная система CHM будет реализована в финальной версии")
	}

	// Обновляем ползунки
	countChanged := g.sliderCount.Update()
	speedChanged := g.sliderSpeed.Update()
	radiusChanged := g.sliderRadius.Update()

	// При изменении числа частиц или радиуса — полный сброс
	if countChanged || radiusChanged {
		g.sim.Reset(g.currentConfig())
	}

	// При изменении скорости — масштабируем скорости без сброса позиций
	if speedChanged {
		newSpeed := float64(g.sliderSpeed.Value) / 2.0
		for _, p := range g.sim.Particles {
			s := p.Speed()
			if s > 1e-9 {
				p.VX = p.VX / s * newSpeed
				p.VY = p.VY / s * newSpeed
			}
		}
	}

	// Кнопки
	if g.btnPause.Update() {
		g.sim.TogglePause()
		if g.sim.Running {
			g.btnPause.Label = "[ Pause ]"
		} else {
			g.btnPause.Label = "[ Start ]"
		}
	}
	if g.btnReset.Update() {
		g.sim.Reset(g.currentConfig())
		g.sim.Running = true
		g.btnPause.Label = "[ Pause ]"
	}

	// Шаг физики
	g.sim.Update()

	// FPS
	g.fpsTick++
	if g.fpsTick >= 30 {
		g.fps = ebiten.ActualFPS()
		g.fpsTick = 0
	}

	return nil
}

// Draw вызывается Ebitengine для отрисовки кадра.
func (g *Game) Draw(screen *ebiten.Image) {
	// Область симуляции
	DrawSimField(screen, SimOffX, SimOffY, SimW, SimH)
	DrawParticles(screen, g.sim, SimOffX, SimOffY)
	if g.showVectors {
		DrawVelocityVectors(screen, g.sim, SimOffX, SimOffY, 6)
	}

	// Панель управления
	g.drawPanel(screen)
}

// drawPanel рисует правую панель управления.
func (g *Game) drawPanel(screen *ebiten.Image) {
	px := float32(SimW)
	pw := float32(PanelWidth)
	ph := float32(WindowHeight)

	// Фон панели
	vector.DrawFilledRect(screen, px, 0, pw, ph, colorPanelBg, false)
	vector.StrokeLine(screen, px, 0, px, ph, 1.5, colorBorder, false)

	x := SimW + PanelPad
	y := 14

	// Заголовок
	text.Draw(screen, "БРОУНОВСКОЕ ДВИЖЕНИЕ", face, x, y, color.White)
	y += 18
	text.Draw(screen, "──────────────────", face, x, y, color.White)
	y += 20

	// Ползунки с подписями и текущими значениями
	text.Draw(screen, fmt.Sprintf("Частицы: %d", g.sliderCount.Value), face, x, y, color.White)
	y += 14
	g.sliderCount.Y = y
	g.sliderCount.X = x
	g.sliderCount.Draw(screen)
	y += 36

	speed := float64(g.sliderSpeed.Value) / 2.0
	text.Draw(screen, fmt.Sprintf("Скорость: %.1f пикс./кадр", speed), face, x, y, color.White)
	y += 14
	g.sliderSpeed.Y = y
	g.sliderSpeed.X = x
	g.sliderSpeed.Draw(screen)
	y += 36

	text.Draw(screen, fmt.Sprintf("Радиус: %d пикс.", g.sliderRadius.Value), face, x, y, color.White)
	y += 14
	g.sliderRadius.Y = y
	g.sliderRadius.X = x
	g.sliderRadius.Draw(screen)
	y += 46

	// Кнопки
	g.btnPause.X = x
	g.btnPause.Y = y
	g.btnPause.Draw(screen)
	y += 42

	g.btnReset.X = x
	g.btnReset.Y = y
	g.btnReset.Draw(screen)
	y += 50

	// Разделитель
	text.Draw(screen, "──────────────────", face, x, y, color.White)
	y += 18

	// Статистика
	text.Draw(screen, "СТАТИСТИКА", face, x, y, color.White)
	y += 18

	ke := g.sim.TotalKineticEnergy()
	text.Draw(screen, fmt.Sprintf("Частиц:       %d", len(g.sim.Particles)), face, x, y, color.White)
	y += 16
	text.Draw(screen, fmt.Sprintf("Столкновений: %d", g.sim.Collisions), face, x, y, color.White)
	y += 16
	text.Draw(screen, fmt.Sprintf("Кин. энергия: %.1f", ke), face, x, y, color.White)
	y += 16
	text.Draw(screen, fmt.Sprintf("FPS:          %.0f", g.fps), face, x, y, color.White)
	y += 30

	// Статус паузы
	status := "▶ работает"
	if !g.sim.Running {
		status = "⏸ пауза"
	}
	text.Draw(screen, "Состояние: "+status, face, x, y, color.White)
	y += 30

	// Разделитель
	text.Draw(screen, "──────────────────", face, x, y, color.White)
	y += 18

	// Горячие клавиши
	text.Draw(screen, "УПРАВЛЕНИЕ", face, x, y, color.White)
	y += 18
	text.Draw(screen, "Пробел  — пауза", face, x, y, color.White)
	y += 14
	text.Draw(screen, "V       — векторы", face, x, y, color.White)
	y += 14
	text.Draw(screen, "F1      — справка*", face, x, y, color.White)
	y += 14
	text.Draw(screen, "Escape  — выход", face, x, y, color.White)
	y += 24

	// Примечание о прototипе
	text.Draw(screen, "* TODO: CHM-справка", face, x, y, color.White)
	y += 14
	text.Draw(screen, "  (финальная версия)", face, x, y, color.White)
}

// Layout возвращает логические размеры экрана (Ebitengine §Layout).
func (g *Game) Layout(outsideWidth, outsideHeight int) (int, int) {
	return WindowWidth, WindowHeight
}

func main() {
	ebiten.SetWindowSize(WindowWidth, WindowHeight)
	ebiten.SetWindowTitle("Броуновское движение — РГР по программированию")
	ebiten.SetWindowResizingMode(ebiten.WindowResizingModeEnabled)

	// Запрет вертикального синхросигнала для более высокого FPS при отладке:
	// ebiten.SetVsyncEnabled(false)
	loadFont()
	if err := ebiten.RunGame(NewGame()); err != nil && err != ebiten.Termination {
		log.Fatal(err)
	}
}
