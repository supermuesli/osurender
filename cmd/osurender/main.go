package main

import (
	"github.com/faiface/pixel/pixelgl"
	"github.com/kbinani/screenshot"

	"github.com/supermuesli/osurender/pkg/render"
)

func run() {
	// get display dimensions
	bounds := screenshot.GetDisplayBounds(0)
	width := float64(bounds.Dx())
	height := float64(bounds.Dy())

	// initialize new canvas
	canvas := render.NewCanvas(width, height)

	// render loop
	for !canvas.Win.Closed() {
		canvas.Poll()
		canvas.Draw()		
		
		// manually enforce FPS
		<-canvas.FPS
	}
}

func main() {
	pixelgl.Run(run)
}