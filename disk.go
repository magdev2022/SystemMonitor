package main

import (
	"fmt"
	"image/color"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"github.com/shirou/gopsutil/disk"
)

func DrawDiskGraph(r_Data []float64, w_Data []float64, diskUsage float64) *fyne.Container {
	graph := container.NewWithoutLayout()
	background := canvas.NewRectangle(color.RGBA{0, 0, 0, 255}) // Black background
	background.SetMinSize(fyne.NewSize(300, 180))
	graph.Add(background)
	r_color := color.RGBA{20, 245, 230, 255} //read speed graph color
	w_color := color.RGBA{120, 229, 40, 255} //write speed graph color

	drawGraph := func() {
		graph.Objects = []fyne.CanvasObject{background} // Clear previous lines

		titleLabel := canvas.NewText("Disk", color.RGBA{255, 255, 255, 255})
		titleLabel.Move(fyne.NewPos(0, 0))
		graph.Add(titleLabel)

		w_speedLabel := canvas.NewText(fmt.Sprintf("Write %.2fMB/s", w_Data[len(w_Data)-1]), w_color)
		w_speedLabel.Move(fyne.NewPos(40, 0))
		graph.Add(w_speedLabel)

		r_speedLabel := canvas.NewText(fmt.Sprintf("Read %.2fMB/s", r_Data[len(r_Data)-1]), r_color)
		r_speedLabel.Move(fyne.NewPos(150, 0))
		graph.Add(r_speedLabel)

		partitions, err := disk.Partitions(false)
		if err == nil {
			h := 0
			for _, partition := range partitions {
				// Get usage statistics for each partition
				usage, err := disk.Usage(partition.Mountpoint)
				if err == nil {
					// Print partition info
					partNameLabel := canvas.NewText(partition.Device, color.White)
					partNameLabel.TextSize = 11
					partNameLabel.Move(fyne.NewPos(320, float32(40)+float32(h)))
					graph.Add(partNameLabel)

					partTotalLabel := canvas.NewText(fmt.Sprintf("Total: %v GB", usage.Total/(1024*1024*1024)), color.RGBA{20, 220, 20, 255})
					partTotalLabel.TextSize = 11
					partTotalLabel.Move(fyne.NewPos(320+float32(len(partition.Device)*10), float32(40)+float32(h)))
					graph.Add(partTotalLabel)

					partUsedLabel := canvas.NewText(fmt.Sprintf("Used: %v GB", usage.Used/(1024*1024*1024)), color.RGBA{220, 20, 20, 255})
					partUsedLabel.TextSize = 11
					partUsedLabel.Move(fyne.NewPos(420, float32(40)+float32(h)))
					graph.Add(partUsedLabel)

					h = h + 20
				}
			}
		}

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

		for i := 1; i < len(r_Data); i++ {
			x1 := float32((i - 1) * 30)
			y1 := float32(100-r_Data[i-1]) + 40
			x2 := float32(i * 30)
			y2 := float32(100-r_Data[i]) + 40
			line := canvas.NewLine(r_color)
			line.Position1 = fyne.NewPos(x1, y1)
			line.Position2 = fyne.NewPos(x2, y2)
			graph.Add(line)
		}
		for i := 1; i < len(w_Data); i++ {
			x1 := float32((i - 1) * 30)
			y1 := float32(100-w_Data[i-1]) + 40
			x2 := float32(i * 30)
			y2 := float32(100-w_Data[i]) + 40
			line := canvas.NewLine(w_color)
			line.Position1 = fyne.NewPos(x1, y1)
			line.Position2 = fyne.NewPos(x2, y2)
			graph.Add(line)
		}
		canvas.Refresh(graph)
	}
	drawGraph()
	return graph
}
