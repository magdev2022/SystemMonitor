package main

import (
	"fmt"
	"image/color"
	i_net "net"
	"strings"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/layout"
	"github.com/shirou/gopsutil/cpu"
	"github.com/shirou/gopsutil/disk"
	"github.com/shirou/gopsutil/mem"
	"github.com/shirou/gopsutil/net"
)

var CPU_COLOR color.Color
var RAM_COLOR color.Color
var DISK_COLOR color.Color
var WIFI_COLOR color.Color
var MAX_WIFI_SPEED float64

func getActiveNetInterface() string {
	interfaces, err := i_net.Interfaces()
	if err != nil {
		fmt.Println("Error:", err)
		return ""
	}

	for _, iface := range interfaces {
		// Skip interfaces that are down or loopback interfaces
		if iface.Flags&i_net.FlagUp == 0 || iface.Flags&i_net.FlagLoopback != 0 {
			continue
		}

		// Check if the interface has an IP address
		addrs, err := iface.Addrs()
		if err != nil {
			fmt.Println("Error:", err)
			continue
		}

		if len(addrs) == 0 {
			continue
		}
		// Determine if it's Wi-Fi or Ethernet
		if strings.Contains(strings.ToLower(iface.Name), "wi-fi") || strings.Contains(strings.ToLower(iface.Name), "wlan") {
			return strings.ToLower(iface.Name)
		} else if strings.Contains(strings.ToLower(iface.Name), "eth") {
			return strings.ToLower(iface.Name)
		}
	}
	return ""
}

func getSystemStats(prevDiskStats map[string]disk.IOCountersStat) (float64, float64, float64, float64, float64, float64, float64, map[string]disk.IOCountersStat) {
	// CPU Usage
	cpuUsage, _ := cpu.Percent(0, false)
	// RAM Usage
	memStats, _ := mem.VirtualMemory()
	// Disk Usage
	diskStats, _ := disk.Usage("/")
	// Network Usage
	initialStats, _ := net.IOCounters(true)

	time.Sleep(time.Millisecond * 500)

	netStats, _ := net.IOCounters(true)

	activeNetInterfaceName := getActiveNetInterface()
	netReadSpeed := 0.0
	netWriteSpeed := 0.0

	for i := 0; i < len(netStats); i++ {
		if strings.ToLower(netStats[i].Name) == activeNetInterfaceName {
			bytesReceived := netStats[i].BytesRecv - initialStats[i].BytesRecv
			mbReceived := float64(bytesReceived) / (1024 * 1024)
			netReadSpeed = mbReceived * 2
			bytesSent := netStats[i].BytesSent - initialStats[i].BytesSent
			mbSent := float64(bytesSent) / (1024 * 1024)
			netWriteSpeed = mbSent * 2
			break
		}
	}

	// Disk IO Counters for read/write usage
	ioStats, _ := disk.IOCounters()
	var readSpeed, writeSpeed float64
	for diskName, stats := range ioStats {
		readSpeed = float64(stats.ReadBytes-prevDiskStats[diskName].ReadBytes) / float64(time.Second)
		writeSpeed = float64(stats.WriteBytes-prevDiskStats[diskName].WriteBytes) / float64(time.Second)
		break
	}

	// Return read/write speeds along with other system stats
	return cpuUsage[0], memStats.UsedPercent, diskStats.UsedPercent, netReadSpeed, netWriteSpeed, readSpeed, writeSpeed, ioStats
}

func DrawGraph(title string, data []float64, graphColor color.Color) *fyne.Container {
	graph := container.NewWithoutLayout()
	background := canvas.NewRectangle(color.RGBA{0, 0, 0, 255}) // Black background
	background.SetMinSize(fyne.NewSize(300, 180))
	graph.Add(background)
	drawGraph := func() {
		graph.Objects = []fyne.CanvasObject{background} // Clear previous lines

		titleLabel := canvas.NewText(title, color.RGBA{255, 255, 255, 255})
		titleLabel.Move(fyne.NewPos(0, 0))
		graph.Add(titleLabel)

		valueLabel := canvas.NewText(fmt.Sprintf("%.2f%%", data[len(data)-1]), graphColor)
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
			line := canvas.NewLine(graphColor)
			line.Position1 = fyne.NewPos(x1, y1)
			line.Position2 = fyne.NewPos(x2, y2)
			graph.Add(line)
		}
		canvas.Refresh(graph)
	}
	drawGraph()
	return graph
}

func main() {
	a := app.New()
	w := a.NewWindow("System Monitor")

	MAX_WIFI_SPEED = 100

	CPU_COLOR = color.RGBA{40, 255, 40, 255}
	RAM_COLOR = color.RGBA{255, 40, 255, 255}
	DISK_COLOR = color.RGBA{40, 255, 255, 255}
	WIFI_COLOR = color.RGBA{255, 255, 40, 255}

	cpuData := make([]float64, 11)
	ramData := make([]float64, 11)
	diskReadData := make([]float64, 11)
	diskWriteData := make([]float64, 11)
	networkReadData := make([]float64, 11)
	networkWriteData := make([]float64, 11)

	CPU_graph := container.NewWithoutLayout()
	RAM_graph := container.NewWithoutLayout()
	DISK_graph := container.NewWithoutLayout()
	WIFI_graph := container.NewWithoutLayout()

	prevDiskStats, _ := disk.IOCounters()
	// Start the animation
	ticker := time.NewTicker(time.Second)
	go func() {
		for range ticker.C {
			cpuUsage, ramUsage, diskUsage, netReadSpeed, netWriteSpeed, diskReadSpeed, diskWriteSpeed, _ := getSystemStats(prevDiskStats)
			cpuData = append(cpuData[1:], cpuUsage)
			ramData = append(ramData[1:], ramUsage)

			networkReadData = append(networkReadData[1:], netReadSpeed)
			networkWriteData = append(networkWriteData[1:], netWriteSpeed)

			diskReadData = append(diskReadData[1:], diskReadSpeed)
			diskWriteData = append(diskWriteData[1:], diskWriteSpeed)

			CPU_graph = DrawCPUGraph(cpuData)
			RAM_graph = DrawRAMGraph(ramData)
			DISK_graph = DrawDiskGraph(diskReadData, diskWriteData, diskUsage)
			WIFI_graph = DrawWifiGraph(networkReadData, networkWriteData)
			graphContent := container.New(layout.NewVBoxLayout(), CPU_graph, RAM_graph, DISK_graph, WIFI_graph)
			w.SetContent(graphContent)
		}
	}()

	w.Resize(fyne.NewSize(500, 800))
	w.ShowAndRun()
}
