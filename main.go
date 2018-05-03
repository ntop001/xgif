package main

import (
	"korok.io/korok/game"
	"korok.io/korok/gui"
	"korok.io/korok/gfx/bk"
	"korok.io/korok/gfx"
	"korok.io/korok"

	"image/gif"
	"image/draw"

	"os"
	"log"
	"image"
	"flag"
	"fmt"
)

type MainScene struct {
	img gfx.Tex2D
	p *GifPlayer
}

func (m *MainScene) OnEnter(g *game.Game) {
	m.img = m.p.Setup()
}

func (m *MainScene) Update(dt float32) {
	if dt > .5 { return } // window freeze!! skip
	m.p.Update(dt)
	gui.Image(1, gui.Rect{0 ,0, 0, 0}, m.img, nil)
}

func (*MainScene) OnExit() {

}

func main() {
	gfile := flag.Arg(0)
	// decode !
	gp, err := Load(gfile)
	if err != nil {
		log.Fatal(err)
	}
	w, h := gp.Size()
	options := korok.Options{
		Title: "Gif Player",
		Width:w,
		Height:h,
	}
	korok.Run(&options, &MainScene{p: gp})
}

type GifPlayer struct {
	canvas *image.RGBA
	gif *gif.GIF
	tex *bk.Texture2D

	frames[]float32
	index int // current index
	delay float32//
}

func Load(path string) (*GifPlayer, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	g, err := gif.DecodeAll(file)
	if err != nil {
		return nil, err
	}

	duration := float32(0)
	frames := make([]float32, len(g.Delay))
	for i, d := range g.Delay {
		frames[i] = float32(d * 10)/1000
		duration += frames[i]
	}
	gp := &GifPlayer{
		gif:g,
		frames:frames,
	}

	log.Print("Gif Size:", g.Config.Width, g.Config.Height)
	log.Print("Gif duration:", duration)
	return gp, nil
}


func (p *GifPlayer) Size() (w, h int) {
	cfg := p.gif.Config
	w, h = cfg.Width, cfg.Height
	return
}

func (p *GifPlayer) Setup() gfx.Tex2D {
	var (
		id uint16
		tex *bk.Texture2D
		w = p.gif.Config.Width
		h = p.gif.Config.Height
	)

	canvas := image.NewRGBA(image.Rect(0, 0, w, h))
	draw.Draw(canvas, canvas.Bounds(),p.gif.Image[0], image.ZP, draw.Src)
	p.canvas = canvas

	if id, tex = bk.R.AllocTexture(canvas); id != bk.InvalidId {
		p.tex = tex
	} else {
		log.Fatal("Can't create texture.")
	}
	return gfx.NewTex(id)
}

func (p *GifPlayer) Update(dt float32) {
	p.delay += dt
	if d := p.frames[p.index]; p.delay >= d {
		for ;p.delay >= d; {
			p.delay -= d
		}
		nxt := (p.index + 1) % len(p.frames)
		p.index = nxt
		p.Swap(nxt)
	}
}

// disposal:https://github.com/donatj/gifdump/blob/master/main.go
func (p *GifPlayer) Swap(i int) {
	img := p.gif.Image[i]

	sb := img.Bounds()

	// implement disposal 1 method
	for y := sb.Min.Y; y < sb.Max.Y; y++ {
		for x := sb.Min.X; x < sb.Max.X; x++ {
			c := img.At(x, y)
			_, _, _, a := c.RGBA()
			if a != 0 {
				p.canvas.Set(x, y, c)
			}
		}
	}
	// update gpu-texture buffer
	err := p.tex.Update(p.canvas, 0, 0, int32(p.gif.Config.Width), int32(p.gif.Config.Height))
	if err != nil {
		log.Fatal(err)
	}
}

func init() {
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage of %s <gif>:\n", os.Args[0])
		flag.PrintDefaults()
	}

	flag.Parse()
	if flag.NArg() != 1 {
		flag.Usage()
		os.Exit(1)
	}
}
