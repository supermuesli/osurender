package render

import (
	"fmt"
	"io/ioutil"
	"time"
	"os"
	"image"
	"image/png"

	"github.com/faiface/pixel"
	"github.com/faiface/pixel/pixelgl"
	"github.com/faiface/pixel/imdraw"
	"github.com/Mempler/rplpa"
)

const (
	cursorTrailBufferSize = 50 // has to be divisible by 2
	cursorSize = 25
)

func (canvas *Canvas) ReadReplay(path string) {
	buf, err := ioutil.ReadFile(path)
	if err != nil {
		panic(err)
	}

	replay, err := rplpa.ParseReplay(buf)
	if err != nil {
		panic(err)
	}

	canvas.Replay = replay
}

// Canvas 
type Canvas struct {
	Win *pixelgl.Window
	
	title string
	width float64
	height float64

	// set FPS
	FPS <-chan time.Time
	
	// Circleing/polling/framebuffer attributes
	frames [][]uint8

	imd *imdraw.IMDraw

	Replay *rplpa.Replay

	Tick int

	cursorTrailBuffer *ringBuffer
	cursorTrailBufferSize int
	cursorColor [2]pixel.RGBA

	ScreenScale pixel.Vec
}

// NewCanvas prepares a new Canvas
func NewCanvas(width float64, height float64) *Canvas {
	cfg := pixelgl.WindowConfig {
		Title:  "osurender",
		Bounds: pixel.R(0, 0, width, height),
		VSync:  false,
	}

	win, err := pixelgl.NewWindow(cfg)
	if err != nil {
		panic(err)
	}

	win.Canvas().SetSmooth(true)
	win.SetCursorVisible(false)


	canvas := Canvas {
		win,
		cfg.Title,
		width,
		height,
		time.Tick(time.Second / 60),
		[][]uint8{},
		imdraw.New(nil),
		nil,
		0,
		&ringBuffer{[cursorTrailBufferSize]pixel.Vec{}, 0},
		cursorTrailBufferSize,
		[2]pixel.RGBA{pixel.RGB(1, 1, 0),pixel.RGB(0, 0.5, 1)},
		pixel.V(1, 1),
	}

	return &canvas
}

type ringBuffer struct {
	vs [cursorTrailBufferSize]pixel.Vec
	index int
}

func (r *ringBuffer) add(v pixel.Vec) {
	if r.index < cursorTrailBufferSize-2 {
		r.vs[r.index] = v
		r.index++
	} else {
		for i := 0; i < r.index; i++ {
			r.vs[i] = r.vs[i+1]
		}
		r.vs[r.index] = v
	}
}

func (r *ringBuffer) clear() {
	r.index = 0
	r.vs = [cursorTrailBufferSize]pixel.Vec{}
}

// Cursor 
func (canvas *Canvas) Cursor(v pixel.Vec, size float64) {
	canvas.imd.Push(v)
	canvas.imd.Circle(size, 0)
	canvas.imd.Draw(canvas.Win)
}

func (canvas *Canvas) Trail() {
	canvas.imd.Clear()
	canvas.imd.EndShape = imdraw.RoundEndShape

	// draw cursor trail
	for i := 1; i < cursorTrailBufferSize-1 ; i++ {
		canvas.imd.Color = pixel.RGB(float64(i)/float64(cursorTrailBufferSize*1.3), 0, float64(i)/float64(cursorTrailBufferSize*1.3))
		canvas.imd.Push(canvas.cursorTrailBuffer.vs[i])
		canvas.imd.Push(canvas.cursorTrailBuffer.vs[i-1])
		canvas.imd.Line(float64(cursorSize + i*(cursorSize)/(cursorTrailBufferSize)))
	}
}

// Dump saves the animation as a set of PNGs using `replayName` as the naming prefix
func (canvas *Canvas) Dump(replayName string) {
	if _, err := os.Stat(replayName); os.IsNotExist(err) {
		os.Mkdir(replayName, 0700)
	}
	
	for i := 0; i < len(canvas.frames); i++ {
		img := image.NewRGBA(image.Rectangle{image.Point{0, 0}, image.Point{int(canvas.width), int(canvas.height)}})
		img.Pix = canvas.frames[i]

		file, err := os.Create(fmt.Sprintf(replayName + "/%s%06d.png", replayName, i))
		if err != nil {
			panic(err)
		}

		if err := png.Encode(file, img); err != nil {
			file.Close()
			panic(err)
		}
	}
}

func (canvas *Canvas) DrawCursor() {
	// update cursor trail buffer interpolation points
	v := pixel.V(float64(canvas.Replay.ReplayData[canvas.Tick].MosueX), float64(canvas.Replay.ReplayData[canvas.Tick].MouseY) - canvas.Win.GetPos().Y)
	fmt.Println(v.ScaledXY(canvas.ScreenScale))
	canvas.cursorTrailBuffer.add(v.ScaledXY(canvas.ScreenScale))
	

	// draw cursor trail
	canvas.Trail()
	
	// set cursor color
	if canvas.Replay.ReplayData[canvas.Tick].KeyPressed.LeftClick || canvas.Replay.ReplayData[canvas.Tick].KeyPressed.RightClick {
		canvas.imd.Color = canvas.cursorColor[0]
	}

	// draw cursor at parsed location
	canvas.Cursor(v.ScaledXY(canvas.ScreenScale), float64(cursorSize))
}

// Poll user input
func (canvas *Canvas) Poll() {

	// play animation at keypress P
	if canvas.Win.JustPressed(pixelgl.KeyP) {
		
		for {
			canvas.Win.UpdateInput()

			if canvas.Win.JustPressed(pixelgl.KeyP) {
				break
			}

			if canvas.Win.Pressed(pixelgl.KeyLeft) {
				if canvas.Tick > 0 {
					canvas.Tick--
					canvas.Draw()
				}
			}

			if canvas.Win.Pressed(pixelgl.KeyRight) {
				if canvas.Tick < len(canvas.Replay.ReplayData) {
					canvas.Tick++
					canvas.Draw()
				}
			}

			<-canvas.FPS
		}
	}


	// dump animation at keypress ENTER
	if canvas.Win.JustPressed(pixelgl.KeyEnter) {
		replayName := "arsch"
		canvas.Dump(replayName)
	}

	if canvas.Win.JustPressed(pixelgl.KeyEscape) {
		canvas.Win.Destroy()
	}
}

// Draw renders the canvas onto the window
func (canvas *Canvas) Draw() {
	canvas.Win.Clear(pixel.RGB(0, 0, 0))
	canvas.DrawCursor()
	canvas.imd.Draw(canvas.Win)
	canvas.Win.Update()
	canvas.imd.Clear()
}
