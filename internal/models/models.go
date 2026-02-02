package models

type StatsEvent struct {
	Event string   `json:"event"`
	Args  []string `json:"args"`
}

type SystemStats struct {
	MemoryBytes      uint64       `json:"memory_bytes"`
	MemoryLimitBytes uint64       `json:"memory_limit_bytes"`
	SwapBytes        uint64       `json:"swap_bytes"`
	SwapLimitBytes   uint64       `json:"swap_limit_bytes"`
	CpuAbsolute      float64      `json:"cpu_absolute"`
	CpuTemp          float64      `json:"cpu_temp"`
	Network          NetworkStats `json:"network"`
	Uptime           uint64       `json:"uptime"`
	State            string       `json:"state"`
	DiskBytes        uint64       `json:"disk_bytes"`
	DiskTotal        uint64       `json:"disk_total"`
	Audio            []AudioState `json:"audio"`
	Wifi             WifiState    `json:"wifi"`
	IP               string       `json:"ip"`
	Battery          BatteryState `json:"battery"`
	Volume           int          `json:"volume"`
	Backlight        int          `json:"backlight"`
}

type NetworkStats struct {
	RxBytes uint64 `json:"rx_bytes"`
	TxBytes uint64 `json:"tx_bytes"`
}

type AudioState struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	Playing   bool   `json:"playing"`
	Artist    string `json:"artist"`
	Title     string `json:"title"`
	Album     string `json:"album"`
	ArtURL    string `json:"art_url"`
	Timestamp int64  `json:"timestamp"` 
	Duration  int64  `json:"duration"`
}

type WifiState struct {
	SSID      string `json:"ssid"`
	Connected bool   `json:"connected"`
}

type BatteryState struct {
	Percentage int  `json:"percentage"`
	PluggedIn  bool `json:"plugged_in"`
}

type LoginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type LoginResponse struct {
	Token string `json:"token"`
}

type WebSocketResponse struct {
	Object string `json:"object"`
	Data   struct {
		Token  string `json:"token"`
		Socket string `json:"socket"`
	} `json:"data"`
}
