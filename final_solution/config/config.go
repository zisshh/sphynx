package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"

	"load-balancer/models"

	"github.com/go-redis/redis/v8"
)

var (
	VirtualServices []*models.VirtualService
	configFile      = "config/config.json"
	mu              sync.Mutex // Ensures thread-safe access to VirtualServices
)

// redis configuration
var RedisClient *redis.Client

func InitializeRedis() {
	RedisClient = redis.NewClient(&redis.Options{
		Addr: "127.0.0.1:6379", // Replace with your Redis host and port
	})
}

// user configuration
type UserConfig struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

func LoadUserConfig() (*UserConfig, error) {
	file, err := os.Open("config/user_conf.json")
	if err != nil {
		return nil, fmt.Errorf("failed to open user configuration file: %w", err)
	}
	defer file.Close()

	var config UserConfig
	if err := json.NewDecoder(file).Decode(&config); err != nil {
		return nil, fmt.Errorf("failed to decode user configuration file: %w", err)
	}

	return &config, nil
}

// CheckConfigFileExists checks if the configuration file exists
func CheckConfigFileExists() bool {
	_, err := os.Stat("config/config.json")
	return !os.IsNotExist(err) // Returns true if the file exists
}

// Initialize the configuration, including dynamic reloading
func InitConfig() {
	ReloadConfig()
	go WatchConfigChanges()
}

// Reloads configuration from the JSON file
func ReloadConfig() {
	mu.Lock()
	defer mu.Unlock()

	absConfigFile, err := filepath.Abs(configFile)
	if err != nil {
		fmt.Printf("Error getting absolute path of config file: %v\n", err)
		return
	}

	data, err := os.ReadFile(absConfigFile)
	if err != nil {
		fmt.Printf("Error reading config file: %v\n", err)
		return
	}

	var updatedServices []*models.VirtualService
	err = json.Unmarshal(data, &updatedServices)
	if err != nil {
		fmt.Printf("Error parsing config file: %v\n", err)
		return
	}

	VirtualServices = updatedServices
	fmt.Println("Configuration reloaded successfully.")
}

// Watches for changes in the config file and reloads it dynamically
func WatchConfigChanges() {
	lastModified := time.Now()
	for {
		time.Sleep(2 * time.Second)

		info, err := os.Stat(configFile)
		if err != nil {
			continue
		}

		if info.ModTime().After(lastModified) {
			lastModified = info.ModTime()
			fmt.Println("Configuration file updated. Reloading...")
			ReloadConfig()
			restartVirtualServices()
		}
	}
}

// Display the current configuration
func DisplayConfig() {
	mu.Lock()
	defer mu.Unlock()

	configBytes, err := json.MarshalIndent(VirtualServices, "", "  ")
	if err != nil {
		fmt.Printf("Error displaying config: %v\n", err)
		return
	}
	fmt.Printf("Loaded configuration:\n%s\n", string(configBytes))
}

// Restarts all virtual services after configuration reload
func restartVirtualServices() {
	mu.Lock()
	defer mu.Unlock()

	for _, vs := range VirtualServices {
		go func(vs *models.VirtualService) {
			if _, err := models.GetCertificate(vs.Port); err == nil {
				vs.StartHTTPS(nil) // Replace nil with your GetHealthyServer function if needed
			} else {
				vs.Start(nil) // Replace nil with your GetHealthyServer function if needed
			}
		}(vs)
	}
	fmt.Println("All virtual services restarted with the updated configuration.")
}
