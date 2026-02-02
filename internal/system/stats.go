package system

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"math"
	"net"
	"nex-server/internal/models"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/shirou/gopsutil/v3/cpu"
	"github.com/shirou/gopsutil/v3/disk"
	"github.com/shirou/gopsutil/v3/host"
	"github.com/shirou/gopsutil/v3/mem"
	psnet "github.com/shirou/gopsutil/v3/net"
)

func GetSystemStats(audio *MediaController) (*models.StatsEvent, error) {
	vm, _ := mem.VirtualMemory()
	sw, _ := mem.SwapMemory()
	cpus, _ := cpu.Percent(0, false)
	totalCpu := 0.0
	if len(cpus) > 0 {
		totalCpu = cpus[0]
	}

	uptime, _ := host.Uptime()
	
	netIO, _ := psnet.IOCounters(false)
	var rx, tx uint64
	if len(netIO) > 0 {
		rx = netIO[0].BytesRecv
		tx = netIO[0].BytesSent
	}

	diskStat, _ := disk.Usage("/")
	
	ip := getLocalIP()
	
	batteryState := getBatteryState()
	wifiState := getWifiState()
	
	audioStates := audio.GetAllStatus()
	
	cpuTemp := getCpuTemp()

	stats := models.SystemStats{
		MemoryBytes:      vm.Used,
		MemoryLimitBytes: vm.Total,
		SwapBytes:        sw.Used,
		SwapLimitBytes:   sw.Total,
		CpuAbsolute:      totalCpu,
		CpuTemp:          cpuTemp,
		Network: models.NetworkStats{
			RxBytes: rx,
			TxBytes: tx,
		},
		Uptime:    uptime,
		State:     "running",
		DiskBytes: diskStat.Used,
		DiskTotal: diskStat.Total,
		IP:        ip,
		Battery:   batteryState,
		Wifi:      wifiState,
		Audio:     audioStates,
		Volume:    getVolume(), 
		Backlight: getBacklight(), 
	}

	statsJson, err := json.Marshal(stats)
	if err != nil {
		return nil, err
	}

	return &models.StatsEvent{
		Event: "stats",
		Args:  []string{string(statsJson)},
	}, nil
}

func getLocalIP() string {
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		return ""
	}
	for _, address := range addrs {
		if ipnet, ok := address.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
			if ipnet.IP.To4() != nil {
				return ipnet.IP.String()
			}
		}
	}
	return ""
}

func getBatteryState() models.BatteryState {
	out, err := exec.Command("upower", "-i", "/org/freedesktop/UPower/devices/DisplayDevice").Output()
	if err != nil {
		return models.BatteryState{Percentage: 100, PluggedIn: true}
	}
	
	str := string(out)
	plugged := strings.Contains(str, "state:               charging") || strings.Contains(str, "state:               fully-charged")
	
	var pct int
	lines := strings.Split(str, "\n")
	for _, line := range lines {
		if strings.Contains(line, "percentage:") {
			parts := strings.Fields(line)
			if len(parts) >= 2 {
				fmt.Sscanf(strings.Trim(parts[1], "%"), "%d", &pct)
			}
		}
	}

	return models.BatteryState{
		Percentage: pct,
		PluggedIn:  plugged,
	}
}

func getWifiState() models.WifiState {
	out, err := exec.Command("iwgetid", "-r").Output()
	ssid := strings.TrimSpace(string(out))
	if err != nil || ssid == "" {
		out, err = exec.Command("nmcli", "-t", "-f", "active,ssid", "dev", "wifi").Output()
		if err == nil {
			lines := strings.Split(string(out), "\n")
			for _, line := range lines {
				if strings.HasPrefix(line, "yes:") {
					ssid = strings.TrimPrefix(line, "yes:")
					break
				}
			}
		}
	}
	
	return models.WifiState{
		SSID:      ssid,
		Connected: ssid != "",
	}
}

func getVolume() int {
	uid := getRealUID()
	username := getUsernameFromUID(uid)
	
	cmd := exec.Command("runuser", "-u", username, "--", "wpctl", "get-volume", "@DEFAULT_AUDIO_SINK@")
	cmd.Env = append(os.Environ(), fmt.Sprintf("XDG_RUNTIME_DIR=/run/user/%d", uid))
	out, err := cmd.Output()
	if err == nil {
		str := strings.TrimSpace(string(out))
		if strings.HasPrefix(str, "Volume:") {
			str = strings.TrimPrefix(str, "Volume:")
			str = strings.TrimSpace(str)
			parts := strings.Fields(str)
			if len(parts) > 0 {
				val, err := strconv.ParseFloat(parts[0], 64)
				if err == nil {
					return int(val * 100)
				}
			}
		}
	}
	
	cmd = exec.Command("runuser", "-u", username, "--", "pactl", "get-sink-volume", "@DEFAULT_SINK@")
	cmd.Env = append(os.Environ(), fmt.Sprintf("XDG_RUNTIME_DIR=/run/user/%d", uid))
	out, err = cmd.Output()
	if err == nil {
		str := string(out)
		parts := strings.Split(str, "/")
		if len(parts) >= 2 {
			volStr := strings.TrimSpace(parts[1])
			volStr = strings.TrimRight(volStr, "%")
			vol, err := strconv.Atoi(volStr)
			if err == nil {
				return vol
			}
		}
	}
	
	return 0
}

func getUsernameFromUID(uid int) string {
	data, err := ioutil.ReadFile("/etc/passwd")
	if err != nil {
		return fmt.Sprintf("#%d", uid)
	}
	
	lines := strings.Split(string(data), "\n")
	for _, line := range lines {
		parts := strings.Split(line, ":")
		if len(parts) >= 3 {
			if parts[2] == strconv.Itoa(uid) {
				return parts[0]
			}
		}
	}
	
	return fmt.Sprintf("#%d", uid)
}

func getRealUID() int {
	sudoUID := os.Getenv("SUDO_UID")
	if sudoUID != "" {
		if uid, err := strconv.Atoi(sudoUID); err == nil {
			return uid
		}
	}
	
	uid := os.Getuid()
	if uid == 0 {
		files, _ := ioutil.ReadDir("/run/user")
		for _, f := range files {
			if f.IsDir() && f.Name() != "0" {
				if id, err := strconv.Atoi(f.Name()); err == nil && id >= 1000 {
					return id
				}
			}
		}
		return 1000
	}
	return uid
}

func getBacklight() int {
	files, err := ioutil.ReadDir("/sys/class/backlight")
	if err != nil || len(files) == 0 {
		return 100
	}

	dir := filepath.Join("/sys/class/backlight", files[0].Name())
	
	maxBytes, err := ioutil.ReadFile(filepath.Join(dir, "max_brightness"))
	if err != nil {
		return 100
	}
	
	actualBytes, err := ioutil.ReadFile(filepath.Join(dir, "brightness"))
	if err != nil {
		return 100
	}

	max, _ := strconv.Atoi(strings.TrimSpace(string(maxBytes)))
	actual, _ := strconv.Atoi(strings.TrimSpace(string(actualBytes)))

	if max == 0 {
		return 100
	}

	return int((float64(actual) / float64(max)) * 100)
}

func getCpuTemp() float64 {
	temps, err := host.SensorsTemperatures()
	if err != nil {
		return 0.0
	}
	for _, temp := range temps {
		if strings.HasPrefix(temp.SensorKey, "coretemp") || strings.HasPrefix(temp.SensorKey, "k10temp") {
			return temp.Temperature
		}
	}
	if len(temps) > 0 {
		return temps[0].Temperature
	}
	return 0.0
}

func Round(val float64) int {
    if val < 0 {
        return int(val - 0.5)
    }
    return int(val + 0.5)
}

func ToFixed(num float64, precision int) float64 {
    output := math.Pow(10, float64(precision))
    return float64(Round(num * output)) / output
}
