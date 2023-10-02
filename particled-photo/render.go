package main

import (
	"cmp"
	"image/color"
	"math"
	"math/rand"
	"sync"
	"time"

	"github.com/go-p5/p5"
	"gonum.org/v1/gonum/spatial/r2"
)

const (
	WIDTH    = 800
	HEIGHT   = 800
	R        = 10
	G        = 0.05
	BOUNCE_F = 0.2
	GRID_W   = WIDTH / (2 * R)
	GRID_H   = HEIGHT / (2 * R)
	NSUBDT   = 16
	INIT_POS = R + 10
	INIT_V   = 5
)

type Particle struct {
	pos   r2.Vec
	v     r2.Vec
	color color.Color
}

var ps []*Particle

var grid [GRID_W][GRID_H][]*Particle

func (p Particle) draw() {
	p5.StrokeWidth(0)
	p5.Fill(p.color)
	p5.Circle(p.pos.X, p.pos.Y, float64(R)*2)
}

func setup() {
	p5.Canvas(WIDTH, HEIGHT)
	go func() {
		for i := 0; i < 600; i++ {
			ps = append(ps,
				&Particle{
					pos: r2.Vec{X: INIT_POS, Y: INIT_POS},
					v:   r2.Vec{X: INIT_V, Y: 0},
					color: color.RGBA{
						R: uint8(rand.Int()),
						G: uint8(rand.Int()),
						B: uint8(rand.Int()),
						A: 255,
					},
				},
			)
			ps = append(ps,
				&Particle{
					pos: r2.Vec{X: INIT_POS, Y: INIT_POS + R*2 + 2},
					v:   r2.Vec{X: INIT_V, Y: 0},
					color: color.RGBA{
						R: uint8(rand.Int()),
						G: uint8(rand.Int()),
						B: uint8(rand.Int()),
						A: 255,
					},
				},
			)
			time.Sleep(time.Millisecond * 100)
		}
	}()
	go perfStat()
}

func clamp[T cmp.Ordered](x, minn, maxx T) T {
	return min(max(x, minn), maxx)
}

func decompose(v, xAxis r2.Vec) r2.Vec {
	x := r2.Dot(v, xAxis) / r2.Norm(xAxis)
	y := math.Sqrt(max(r2.Norm2(v)-x*x, 0))
	if r2.Cross(v, xAxis) > 0 {
		y = -y
	}
	return r2.Vec{X: x, Y: y}
}

func compose(v, xAxis r2.Vec) r2.Vec {
	ux := r2.Unit(xAxis)
	uy := r2.Unit(r2.Rotate(xAxis, math.Pi/2, r2.Vec{X: 0, Y: 0}))
	return r2.Add(r2.Scale(v.X, ux), r2.Scale(v.Y, uy))
}

func perfStat() {
	for {
		lastFrameCnt := p5.FrameCount()
		time.Sleep(time.Second)
		println("particles:", len(ps), "FPS:", p5.FrameCount()-lastFrameCnt)

	}
}

func solCollision(x, y, x_, y_ int) {
	g, g_ := grid[x][y], grid[x_][y_]

	for _, p := range g {
		for _, p_ := range g_ {
			if p == p_ {
				continue
			}
			if r2.Norm(r2.Sub(p.pos, p_.pos)) >= R*2 {
				continue
			}
			diff := r2.Sub(p.pos, p_.pos)
			if r2.Norm(diff) <= 0.00001 {
				diff = r2.Vec{X: 0.0001, Y: 0.0001}
			}
			bias := R*2 - r2.Norm(diff)
			delta := r2.Scale(bias/2, r2.Unit(diff))
			p.pos = r2.Add(p.pos, delta)
			p_.pos = r2.Sub(p_.pos, delta)
			vDec := decompose(p.v, diff)
			v_Dec := decompose(p_.v, diff)
			vx := (vDec.X + v_Dec.X) / 2
			vDec.X, v_Dec.X = vx, vx
			p.v = compose(vDec, diff)
			p_.v = compose(v_Dec, diff)
		}
	}
}

func update(dt float64) {
	for _, p := range ps {
		p.v = r2.Add(p.v, r2.Scale(dt, r2.Vec{X: 0, Y: G}))
		p.pos = r2.Add(p.pos, r2.Scale(dt, p.v))
	}

	for x := range grid {
		for y := range grid[x] {
			grid[x][y] = make([]*Particle, 0)
		}
	}
	for _, p := range ps {
		if p.pos.X > float64(WIDTH-R) || p.pos.X < float64(R) {
			p.pos.X = clamp(p.pos.X, R, float64(WIDTH)-R)
			p.v = r2.Vec{X: -p.v.X * BOUNCE_F, Y: p.v.Y}
		}
		if p.pos.Y > float64(HEIGHT-R) || p.pos.Y < float64(R) {
			p.pos.Y = clamp(p.pos.Y, R, float64(WIDTH)-R)
			p.v = r2.Vec{X: p.v.X, Y: -p.v.Y * BOUNCE_F}
		}
		gx, gy := int(p.pos.X)/(2*R), int(p.pos.Y)/(2*R)
		grid[gx][gy] = append(grid[gx][gy], p)
	}

	var wg sync.WaitGroup

	for x := 0; x < len(grid); x += 3 {
		wg.Add(1)
		go func(x int) {
			defer wg.Done()
			for y := range grid[x] {
				for dx := -1; dx <= 1; dx++ {
					for dy := -1; dy <= 1; dy++ {
						x_, y_ := x+dx, y+dy
						if x_ >= GRID_W || x_ < 0 ||
							y_ >= GRID_H || y_ < 0 {
							continue
						}
						solCollision(x, y, x_, y_)
					}
				}
			}
		}(x)
	}

	wg.Wait()

	for x := 1; x < len(grid); x += 3 {
		wg.Add(1)
		go func(x int) {
			defer wg.Done()
			for y := range grid[x] {
				for dx := -1; dx <= 1; dx++ {
					for dy := -1; dy <= 1; dy++ {
						x_, y_ := x+dx, y+dy
						if x_ >= GRID_W || x_ < 0 ||
							y_ >= GRID_H || y_ < 0 {
							continue
						}
						solCollision(x, y, x_, y_)
					}
				}
			}
		}(x)
	}

	wg.Wait()

	for x := 2; x < len(grid); x += 3 {
		wg.Add(1)
		go func(x int) {
			defer wg.Done()
			for y := range grid[x] {
				for dx := -1; dx <= 1; dx++ {
					for dy := -1; dy <= 1; dy++ {
						x_, y_ := x+dx, y+dy
						if x_ >= GRID_W || x_ < 0 ||
							y_ >= GRID_H || y_ < 0 {
							continue
						}
						solCollision(x, y, x_, y_)
					}
				}
			}
		}(x)
	}

	wg.Wait()
}

func draw() {
	p5.Background(color.Gray{Y: 200})

	for d := 0; d < NSUBDT; d++ {
		update(1 / float64(NSUBDT))
	}

	for _, p := range ps {
		p.draw()
	}
}

func main() {
	p5.Run(setup, draw)
}