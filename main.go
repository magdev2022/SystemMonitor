package main

import (
	"fmt"
	"image/color"
	i_net "net"
	"strings"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/theme"
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
	cpuUsage, _ := cpu.Percent(0, false)
	memStats, _ := mem.VirtualMemory()
	diskStats, _ := disk.Usage("/")
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

	ioStats, _ := disk.IOCounters()
	var readSpeed, writeSpeed float64
	for diskName, stats := range ioStats {
		readSpeed = float64(stats.ReadBytes-prevDiskStats[diskName].ReadBytes) / float64(time.Second)
		writeSpeed = float64(stats.WriteBytes-prevDiskStats[diskName].WriteBytes) / float64(time.Second)
		break
	}

	return cpuUsage[0], memStats.UsedPercent, diskStats.UsedPercent, netReadSpeed, netWriteSpeed, readSpeed, writeSpeed, ioStats
}

func main() {
	a := app.New()
	w := a.NewWindow("System Monitor")

	a.Settings().SetTheme(theme.DarkTheme())

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
