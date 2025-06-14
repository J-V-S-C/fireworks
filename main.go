package main

import (
	"fmt"
	"math"
	"math/rand"
	"strings"
	"sync"
	"time"
)

const w = 190
const h = 40

var rockets []Rocket
var particles []Particle
var bombs []Bomb
var mu sync.Mutex
var screen [][]string

type Rocket struct {
	x, y    float64
	vx      float64
	char    rune
	visible bool
}

type Bomb struct {
	x, y     float64
	vy       float64
	char     rune
	exploded bool
}

type Particle struct {
	x, y   float64
	vx, vy float64
	life   int
	char   rune
	color  string
}

func ChooseColor() string {
	colorCodes := []string{
		"\033[38;5;94m",  // marrom ferrugem
		"\033[38;5;130m", // laranja queimado
		"\033[38;5;124m", // vermelho escuro
		"\033[38;5;136m", // ocre (terra)
		"\033[38;5;240m", // cinza escuro (fumaça)
		"\033[38;5;235m", // carvão quase preto
		"\033[38;5;226m", // amarelo poeira
	}
	return colorCodes[rand.Intn(len(colorCodes))]
}

func ChooseMyParticle() rune {
	shapes := []rune{
		'·', '•', '∙', '⁘', '⁙',
		'░', '▒', '▓', '▖', '▘', '▝', '▗',
		'^', '~', '˄', '▲', '∆',
	}
	return shapes[rand.Intn(len(shapes))]
}

func CreateParticle(x float64, y float64) Particle {
	angle := rand.Float64() * 2 * 3.1415 // 0 a 2π
	speed := rand.Float64() * 2          // 0 a 2

	vx := speed * math.Cos(angle)
	vy := speed * math.Sin(angle)
	return Particle{
		x:     x,
		y:     y,
		vx:    vx,
		vy:    vy,
		life:  10 + rand.Intn(10),
		char:  ChooseMyParticle(),
		color: ChooseColor(),
	}
}

func CreateRocket(y float64) Rocket {
	return Rocket{
		x:       0,
		y:       y,
		vx:      0.8,
		char:    '🙮',
		visible: true,
	}
}

func CreateBomb(x float64, y float64) Bomb {
	return Bomb{
		x:        x,
		y:        y,
		vy:       -0.7,
		char:     '𜱣',
		exploded: false,
	}
}

func (p *Particle) updateParticle() {
	p.x += p.vx
	p.y += p.vy
	p.vy += 0.1
	p.life--

	if p.life < 5 {
		p.vx *= 0.9
		p.vy *= 0.9
	}
}

func (r *Rocket) updateRocket() {
	r.x += r.vx
	if r.x >= w-1 {
		r.visible = false
	}
}

func (b *Bomb) updateBomb() {
	b.y -= b.vy
	if b.y >= h-(2+rand.Float64()*3) {
		b.exploded = true
	}
}

func ClearScreen() {
	screen = make([][]string, h)
	for i := range screen {
		screen[i] = make([]string, w)
		for j := range screen[i] {
			if i == 0 || i == h-1 {
				screen[i][j] = "-"
			} else if j == 0 || j == w-1 {
				screen[i][j] = "|"
			} else {
				screen[i][j] = " "
			}
		}
	}

	screen[0][0] = "┌"
	screen[0][w-1] = "┐"
	screen[h-1][0] = "└"
	screen[h-1][w-1] = "┘"
}

func UpdateRockets() {
	for i := range rockets {
		r := &rockets[i]
		r.updateRocket()

		if r.visible {
			x := int(r.x)
			y := int(r.y)
			if y >= 0 && y < h && x >= 0 && x < w {
				screen[int(r.y)][int(r.x)] = string(r.char)
			}

			if rand.Float64() < 0.005 {
				bombs = append(bombs, CreateBomb(r.x, r.y))
			}

		}
	}
}

func UpdateParticles() {
	for i := range particles {
		p := &particles[i]
		p.updateParticle()
		x := int(p.x)
		y := int(p.y)
		if p.life > 0 && y >= 0 && y < h && x >= 0 && x < w {
			screen[y][x] = p.color + string(p.char) + "\033[0m"
		}
	}
}

func UpdateBombs() {
	for i := range bombs {
		b := &bombs[i]
		b.updateBomb()

		if b.exploded {
			newParticles := make([]Particle, 50)
			for i := range newParticles {
				newParticles[i] = CreateParticle(b.x, b.y)
			}
			particles = append(particles, newParticles...)
		} else {
			x := int(b.x)
			y := int(b.y)
			if y >= 0 && y < h && x >= 0 && x < w {
				screen[int(b.y)][int(b.x)] = string(b.char)
			}
		}
	}
}

func RipRockets() {
	var aliveRockets []Rocket
	for _, r := range rockets {
		if r.visible {
			aliveRockets = append(aliveRockets, r)
		}

	}
	rockets = aliveRockets
}

func RipParticles() {
	var aliveParticles []Particle
	for _, p := range particles {
		if p.life > 0 {
			aliveParticles = append(aliveParticles, p)
		}
	}
	particles = aliveParticles

}

func RipBombs() {
	var aliveBombs []Bomb
	for _, b := range bombs {
		if !b.exploded {
			aliveBombs = append(aliveBombs, b)
		}
		if b.exploded {
			continue
		}
	}
	bombs = aliveBombs
}

func Launch(wg *sync.WaitGroup) {
	defer wg.Done()

	for {
		mu.Lock()
		y := float64(rand.Intn(h / 2))
		rockets = append(rockets, CreateRocket(y))
		mu.Unlock()

		time.Sleep(time.Duration(1+rand.Intn(6)) * time.Second)
	}
}

func Render(wg *sync.WaitGroup) {
	defer wg.Done()

	for {
		mu.Lock()

		ClearScreen()
		UpdateRockets()
		UpdateBombs()
		UpdateParticles()
		RipRockets()
		RipBombs()
		RipParticles()

		mu.Unlock()

		fmt.Print("\033[H\033[2J")

		for _, row := range screen {
			fmt.Println(strings.Join(row, ""))
		}

		time.Sleep(70 * time.Millisecond)
	}
}

func main() {
	var wg sync.WaitGroup
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go Launch(&wg)
	}
	wg.Add(1)
	go Render(&wg)

	wg.Wait()
}
