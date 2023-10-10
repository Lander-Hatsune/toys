package main

import (
	"cmp"
	"flag"
	"fmt"
	"image"
	"image/color"
	"log"
	"math"
	"os"
	"runtime/pprof"
	"runtime/trace"
	"sync"
	"time"

	"github.com/go-p5/p5"
	"gonum.org/v1/gonum/spatial/r2"
)

const (
	WIDTH      = 720
	HEIGHT     = 720
	R          = 10
	G          = 0.05
	BOUNCE_F   = 0.2
	GRID_SCALE = 2 * R
	GRID_W     = WIDTH / GRID_SCALE
	GRID_H     = HEIGHT / GRID_SCALE
	NSUBDT     = 12
	MAX_N      = 814 * 2
)

type Particle struct {
	pos   r2.Vec
	v     r2.Vec
	color color.Color
}

var ps []*Particle

var grid [GRID_W][GRID_H][]*Particle

func (p Particle) draw() {
	p5.Fill(p.color)
	p5.Circle(p.pos.X, p.pos.Y, float64(R)*2)
}

var colorList []color.Color

func setup() {
	p5.StrokeWidth(0)

	p5.Canvas(WIDTH, HEIGHT)
	if *cmd == "eval" {
		f, err := os.Open(*path)
		if err != nil {
			panic("*** Color scheme invalid! ***")
		}
		defer f.Close()
		for {
			var r, g, b, a uint8
			_, err := fmt.Fscanf(f, "%v %v %v %v", &r, &g, &b, &a)
			if err != nil {
				break
			}
			colorList = append(colorList, color.RGBA{r, g, b, a})
		}
		if len(colorList) != MAX_N {
			panic("*** ncolors doesnt match nparticles! ***")
		}
	} else {
		for i := 0; i < MAX_N; i++ {
			colorList = append(colorList, color.Black)
		}
	}
	go monitor()
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

func doTrace(path string) {
	f, err := os.Create(path)
	if err != nil {
		log.Fatal(err)
	}
	trace.Start(f)
	defer trace.Stop()
	time.Sleep(time.Millisecond * 300)
}

func doCPUProfile(path string) {
	f, err := os.Create(path)
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()
	pprof.StartCPUProfile(f)
	defer pprof.StopCPUProfile()
	time.Sleep(time.Millisecond * 1000)
}

func monitor() {
	const INTERV = 500
	for {
		lastFrameCnt := p5.FrameCount()
		time.Sleep(time.Millisecond * INTERV)
		fps := (p5.FrameCount() - lastFrameCnt) * 1000 / INTERV
		log.Println("particles:", len(ps), "FPS:", fps)
		if fps <= 20 {
			if *tracePath != "" {
				doTrace(*tracePath)
				os.Exit(0)
			}
			if *cpuProfile != "" {
				doCPUProfile(*cpuProfile)
				os.Exit(0)
			}
		}
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

func solColCollision(y int) {
	for x := range grid {
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
		gx, gy := int(p.pos.X)/GRID_SCALE, int(p.pos.Y)/GRID_SCALE
		grid[gx][gy] = append(grid[gx][gy], p)
	}

	var wg sync.WaitGroup

	for y := 0; y < len(grid[0]); y += 3 {
		wg.Add(1)
		go func(y int) {
			defer wg.Done()
			solColCollision(y)
		}(y)
	}
	wg.Wait()

	for y := 1; y < len(grid[0]); y += 3 {
		wg.Add(1)
		go func(y int) {
			defer wg.Done()
			solColCollision(y)
		}(y)
	}
	wg.Wait()

	for y := 2; y < len(grid[0]); y += 3 {
		wg.Add(1)
		go func(y int) {
			defer wg.Done()
			solColCollision(y)
		}(y)
	}
	wg.Wait()
}

func doColoring() {
	time.Sleep(time.Second * 5)
	f, err := os.Open(*path)
	if err != nil {
		panic(err)
	}
	defer f.Close()
	img, _, err := image.Decode(f)
	if err != nil {
		panic(err)
	}
	of, _ := os.Create(*path + ".txt")
	defer of.Close()
	for _, p := range ps {
		p.color = img.At(int(p.pos.X), int(p.pos.Y))
		r, g, b, a := p.color.RGBA()
		fmt.Fprintf(of, "%v %v %v %v\n", uint8(r), uint8(g), uint8(b), uint8(a))
	}
}

var colorp = -1

func nextColor() color.Color {
	colorp += 1
	return colorList[colorp]
}

var colored = false
var sprayV = r2.Vec{X: 10, Y: 0}

func spray() {
	fc := p5.FrameCount()
	if fc%180 < 90 {
		sprayV = r2.Rotate(sprayV, (math.Pi / 180), r2.Vec{0, 0})
	} else {
		sprayV = r2.Rotate(sprayV, -(math.Pi / 180), r2.Vec{0, 0})
	}
	ps = append(ps,
		&Particle{
			pos:   r2.Vec{X: R * 2, Y: R * 2},
			v:     sprayV,
			color: nextColor(),
		},
	)

}

var waterfallBias = []float64{-R * 6, -R * 3, 0, R * 3, R * 6}

func waterfall() {
	fc := p5.FrameCount()
	ps = append(ps,
		&Particle{
			pos:   r2.Vec{X: WIDTH/2 + waterfallBias[fc%5], Y: R * 2},
			v:     r2.Vec{X: 0, Y: 5},
			color: nextColor(),
		},
	)
}

func draw() {
	if len(ps) < MAX_N {
		// spray()
		waterfall()
	} else {
		if !colored && *cmd == "sim" {
			colored = true
			go doColoring()
		}
	}
	p5.Background(color.Gray{Y: 200})

	for d := 0; d < NSUBDT; d++ {
		update(1 / float64(NSUBDT))
	}

	for _, p := range ps {
		p.draw()
	}
}

var cmd = flag.String("cmd", "eval", "sim/eval")
var path = flag.String("path", "demo.txt", "image path/scheme path")
var cpuProfile = flag.String("cpu-profile", "", "")
var tracePath = flag.String("trace-path", "", "")

func main() {
	flag.Parse()
	p5.Run(setup, draw)
}
