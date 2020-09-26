package main

import (	
	"github.com/faiface/pixel"
	"github.com/faiface/pixel/pixelgl"
	"github.com/kbinani/screenshot"

	"github.com/supermuesli/osurender/pkg/render"
)

func run() {
	// get display dimensions
	bounds := screenshot.GetDisplayBounds(0)
	width := float64(bounds.Dx())*0.98
	height := float64(bounds.Dy())*0.98

	// initialize new canvas
	canvas := render.NewCanvas(width, height)
	canvas.ReadReplay("replay1.osr")

	maxCoordX := float32(0.0)
	maxCoordY := float32(0.0)
	for i := 0; i < len(canvas.Replay.ReplayData); i++ {
		if canvas.Replay.ReplayData[i].MosueX > maxCoordX {
			maxCoordX = canvas.Replay.ReplayData[i].MosueX
		}
		if canvas.Replay.ReplayData[i].MouseY > maxCoordY {
			maxCoordY = canvas.Replay.ReplayData[i].MouseY
		}
	}

	canvas.ScreenScale = pixel.V(height/float64(maxCoordY), height/float64(maxCoordY))
	// render loop
	//for !canvas.Win.Closed() {
	//	canvas.Tick = 0
	for i := 0; i < len(canvas.Replay.ReplayData); i++ {
		canvas.Poll()
		canvas.Draw()

		// manually enforce FPS
		<-canvas.FPS
		canvas.Tick++
	}		
	//}
}

func main() {
	pixelgl.Run(run)
}