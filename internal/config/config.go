package config

import (
	"crypto/rand"
	"encoding/hex"
	"os"
	"path/filepath"

	"github.com/google/uuid"
	"gopkg.in/yaml.v3"
)

type Config struct {
	Debug bool   `yaml:"debug"`
	UUID  string `yaml:"uuid"`
	API   struct {
		Host                  string `yaml:"host"`
		Port                  int    `yaml:"port"`
		DisableRemoteDownload bool   `yaml:"disable_remote_download"`
		UploadLimit           int64  `yaml:"upload_limit"`
	} `yaml:"api"`
	User struct {
		Username string `yaml:"username"`
		Password string `yaml:"password"`
	} `yaml:"user"`
	System struct {
		LogDirectory           string `yaml:"log_directory"`
		TmpDirectory           string `yaml:"tmp_directory"`
		Timezone               string `yaml:"timezone"`
		DiskCheckInterval      int    `yaml:"disk_check_interval"`
		ActivitySendInterval   int    `yaml:"activity_send_interval"`
		CheckPermissionsOnBoot bool   `yaml:"check_permissions_on_boot"`
		EnableLogRotate        bool   `yaml:"enable_log_rotate"`
		WebsocketLogCount      int    `yaml:"websocket_log_count"`
		SFTP                   struct{}
	} `yaml:"system"`
	BindAddress           string `yaml:"bind_address"`
	BindPort              int    `yaml:"bind_port"`
	ReadOnly              bool   `yaml:"read_only"`
	CrashDetection        struct{}
	Enabled               bool `yaml:"enabled"`
	DetectCleanExitAsCrash bool `yaml:"detect_clean_exit_as_crash"`
	Timeout               int  `yaml:"timeout"`
	JWTSecret             string `yaml:"jwt_secret"` 
}

var Current *Config

func Load() error {
	path := "/etc/nex/config.yml"
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return createDefault(path)
	}

	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}

	Current = &Config{}
	return yaml.Unmarshal(data, Current)
}

func createDefault(path string) error {
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	jwtSecret, _ := generateRandomString(32)

	cfg := Config{
		Debug: false,
		UUID:  uuid.New().String(),
		API: struct {
			Host                  string `yaml:"host"`
			Port                  int    `yaml:"port"`
			DisableRemoteDownload bool   `yaml:"disable_remote_download"`
			UploadLimit           int64  `yaml:"upload_limit"`
		}{
			Host:                  "0.0.0.0",
			Port:                  9384,
			DisableRemoteDownload: false,
			UploadLimit:           4064,
		},
		User: struct {
			Username string `yaml:"username"`
			Password string `yaml:"password"`
		}{
			Username: "admin",
			Password: "admin",
		},
		System: struct {
			LogDirectory           string `yaml:"log_directory"`
			TmpDirectory           string `yaml:"tmp_directory"`
			Timezone               string `yaml:"timezone"`
			DiskCheckInterval      int    `yaml:"disk_check_interval"`
			ActivitySendInterval   int    `yaml:"activity_send_interval"`
			CheckPermissionsOnBoot bool   `yaml:"check_permissions_on_boot"`
			EnableLogRotate        bool   `yaml:"enable_log_rotate"`
			WebsocketLogCount      int    `yaml:"websocket_log_count"`
			SFTP                   struct{}
		}{
			LogDirectory:           "/var/log/nexserver",
			TmpDirectory:           "/tmp/nexserver",
			Timezone:               "America/Sao_Paulo",
			DiskCheckInterval:      150,
			ActivitySendInterval:   60,
			CheckPermissionsOnBoot: true,
			EnableLogRotate:        true,
			WebsocketLogCount:      150,
		},
		BindAddress:            "0.0.0.0",
		BindPort:               2222,
		ReadOnly:               false,
		Enabled:                true,
		DetectCleanExitAsCrash: true,
		Timeout:                60,
		JWTSecret:              jwtSecret,
	}

	data, err := yaml.Marshal(cfg)
	if err != nil {
		return err
	}

	Current = &cfg
	return os.WriteFile(path, data, 0644)
}

func generateRandomString(length int) (string, error) {
	bytes := make([]byte, length)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}
