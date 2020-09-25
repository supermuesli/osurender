package render

import (
	"fmt"
	"time"
	"os"
	"image"
	"image/png"

	"github.com/faiface/pixel"
	"github.com/faiface/pixel/pixelgl"
	"github.com/faiface/pixel/imdraw"
)

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
}

// NewCanvas prepares a new Canvas
func NewCanvas(width float64, height float64) *Canvas {
	cfg := pixelgl.WindowConfig {
		Title:  "anim8",
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
		imdraw.New(nil)
	}

	canvas.imd.Color = pixel.RGB(1, 0, 0)

	return &canvas
}

// Circle 
func (canvas *Canvas) Circle(v pixel.Vec) {
	canvas.imd.Push(v)
}

func (canvas *Canvas) buildFrame() {
	// clear screen except for canvas
	canvas.Win.Clear(colornames.Black)
	canvas.batch.Draw(canvas.Win)
	canvas.Win.Update()

	// now get canvas pixels
	canv := canvas.Win.Canvas()
	pixels := canv.Pixels()

	if canvas.curBatch < len(canvas.frames) {
		canvas.frames[canvas.curBatch] = pixels		
	} else {
		// this is so we can dump the frame without GUI as a PNG later
		canvas.frames = append(canvas.frames, pixels)
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

// Poll user input
func (canvas *Canvas) Poll() {
		
	if canvas.Win.JustPressed(pixelgl.KeyLeft) {
		if canvas.curBatch > 0 {
			canvas.buildFrame()
			canvas.curBatch--
			canvas.batch = canvas.batches[canvas.curBatch]
		}
	}
	if canvas.Win.JustPressed(pixelgl.KeyRight) {
		if canvas.curBatch < len(canvas.batches) - 1 {
			canvas.buildFrame()
			canvas.curBatch++
			canvas.batch = canvas.batches[canvas.curBatch]
		}	
	}

	// play animation at keypress P
	if canvas.Win.JustPressed(pixelgl.KeyP) {
		
		canv := canvas.Win.Canvas()
		
		// show animation at 15 FPS
		fps15 := time.Tick(time.Second/time.Duration(canvas.playbackFPS))
		for i := 0; i < len(canvas.frames); i++ {
			canv.SetPixels(canvas.frames[i])
			canvas.Win.Update()
			// note that canvas.Win.Update also calls
			// canvas.Win.UpdateInput() along with it
			
			if canvas.Win.JustPressed(pixelgl.KeyP) {
				break
			}
			<-fps15
		}
	}

	// loop at keypress L
	if canvas.Win.JustPressed(pixelgl.KeyL) {

		canv := canvas.Win.Canvas()
		skipped := false
		
		for {
			// show animation at 15 FPS
			fps15 := time.Tick(time.Second/time.Duration(canvas.playbackFPS))
			for i := 0; i < len(canvas.frames); i++ {
				canv.SetPixels(canvas.frames[i])

				canvas.Win.Update()
				// note that canvas.Win.Update also calls
				// canvas.Win.UpdateInput() along with it

				if canvas.Win.JustPressed(pixelgl.KeyL) {
					skipped = true
					break
				}	
				if canvas.Win.Pressed(pixelgl.KeyUp) {
					canvas.playbackFPS = canvas.playbackFPS + 1
				}
				if canvas.Win.Pressed(pixelgl.KeyDown) {
					canvas.playbackFPS = canvas.playbackFPS - 1
					if canvas.playbackFPS < 5 {
						canvas.playbackFPS = 5
					}
				}
				<-fps15
			}

			if skipped || canvas.Win.JustPressed(pixelgl.KeyL) {
				break
			}	
			
			if canvas.Win.Pressed(pixelgl.KeyUp) {
				canvas.playbackFPS = canvas.playbackFPS + 1
			}
			
			if canvas.Win.Pressed(pixelgl.KeyDown) {
				canvas.playbackFPS = canvas.playbackFPS - 1
				if canvas.playbackFPS < 5 {
					canvas.playbackFPS = 5
				}
			}
			<-canvas.FPS
		}		
	}


	// dump animation at keypress ENTER
	if canvas.Win.JustPressed(pixelgl.KeyEnter) {
		replayName := ""
		canvas.Dump(replayName)
	}

	if canvas.Win.JustPressed(pixelgl.KeyEscape) {
		canvas.Win.Destroy()
	}
}

// Draw renders the canvas onto the window
func (canvas *Canvas) Draw() {
	canvas.Win.Clear()
	canvas.imd.Draw()
	canvas.Win.Update()
	canvas.imd.Clear()
}
