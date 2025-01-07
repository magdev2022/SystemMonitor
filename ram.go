package main

import (
	"fmt"
	"image/color"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"github.com/shirou/gopsutil/mem"
)

func DrawRAMGraph(data []float64) *fyne.Container {
	graph := container.NewWithoutLayout()
	v, _ := mem.VirtualMemory()

	background := canvas.NewRectangle(color.RGBA{0, 0, 0, 255}) // Black background
	background.SetMinSize(fyne.NewSize(300, 180))
	graph.Add(background)
	drawGraph := func() {
		graph.Objects = []fyne.CanvasObject{background} // Clear previous lines

		titleLabel := canvas.NewText("RAM", color.RGBA{255, 255, 255, 255})
		titleLabel.Move(fyne.NewPos(0, 0))
		graph.Add(titleLabel)

		ramSizeLabel := canvas.NewText(fmt.Sprintf("Total Size: %v GB", v.Total/(1024*1024*1024)), color.White)
		ramSizeLabel.Move(fyne.NewPos(120, 0))
		graph.Add(ramSizeLabel)

		valueLabel := canvas.NewText(fmt.Sprintf("%.2f%%", data[len(data)-1]), RAM_COLOR)
		valueLabel.Move(fyne.NewPos(40, 0))
		graph.Add(valueLabel)

		axisY_0 := canvas.NewText("0", color.RGBA{255, 255, 255, 200})
		axisY_0.Move(fyne.NewPos(0, 140))

		axisY_100 := canvas.NewText("100", color.RGBA{255, 255, 255, 200})
		axisY_100.Move(fyne.NewPos(0, 20))

		border := canvas.NewRectangle(color.RGBA{255, 255, 255, 0})
		border.StrokeWidth = 2
		border.StrokeColor = color.RGBA{255, 255, 255, 100}
		border.Resize(fyne.NewSize(300, 100))
		border.Move(fyne.NewPos(0, 40))
		graph.Add(axisY_0)
		graph.Add(axisY_100)
		graph.Add(border)

		for i := 1; i < len(data); i++ {
			x1 := float32((i - 1) * 30)
			y1 := float32(100-data[i-1]) + 40
			x2 := float32(i * 30)
			y2 := float32(100-data[i]) + 40
			line := canvas.NewLine(RAM_COLOR)
			line.Position1 = fyne.NewPos(x1, y1)
			line.Position2 = fyne.NewPos(x2, y2)
			graph.Add(line)
		}
		canvas.Refresh(graph)
	}
	drawGraph()
	return graph
}
