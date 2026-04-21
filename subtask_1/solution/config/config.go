package config

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"load-balancer/models"
)

var VirtualServices []*models.VirtualService

func InitConfig() {
	configFile := "config/config.json"
	absConfigFile, err := filepath.Abs(configFile)
	if err != nil {
		fmt.Printf("Error getting absolute path of config file: %v\n", err)
		os.Exit(1)
	}

	if _, err := os.Stat(absConfigFile); os.IsNotExist(err) {
		fmt.Println("No config file found, defaulting to empty configuration for testing.")
		VirtualServices = []*models.VirtualService{}
		return
	}

	// Automatically use the config file if it exists
	data, err := os.ReadFile(absConfigFile)
	if err != nil {
		fmt.Printf("Error reading config file: %v\n", err)
		os.Exit(1)
	}

	err = json.Unmarshal(data, &VirtualServices)
	if err != nil {
		fmt.Printf("Error parsing config file: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("Config file successfully loaded.")
}

func readConfigFromInput() {
	reader := bufio.NewReader(os.Stdin)
	for {
		fmt.Print("Enter port: ")
		portStr, _ := reader.ReadString('\n')
		port, _ := strconv.Atoi(strings.TrimSpace(portStr))

		fmt.Print("Enter algorithm: ")
		algorithm, _ := reader.ReadString('\n')
		algorithm = strings.TrimSpace(algorithm)

		var serverList []*models.Server
		for {
			fmt.Print("Enter server name (or 'done' to finish): ")
			name, _ := reader.ReadString('\n')
			name = strings.TrimSpace(name)
			if name == "done" {
				break
			}

			fmt.Print("Enter server URL: ")
			url, _ := reader.ReadString('\n')
			url = strings.TrimSpace(url)

			fmt.Print("Enter server weight: ")
			weightStr, _ := reader.ReadString('\n')
			weight, _ := strconv.Atoi(strings.TrimSpace(weightStr))

			serverList = append(serverList, models.NewServer(name, url, weight))
		}

		VirtualServices = append(VirtualServices, &models.VirtualService{
			Port:       port,
			Algorithm:  algorithm,
			ServerList: serverList,
			Logger:     nil, // Logger will be initialized later
		})

		fmt.Print("Add another virtual service? (y/n): ")
		another, _ := reader.ReadString('\n')
		another = strings.TrimSpace(another)
		if another != "y" {
			break
		}
	}
}

func DisplayConfig() {
	configBytes, err := json.MarshalIndent(VirtualServices, "", "  ")
	if err != nil {
		fmt.Printf("Error displaying config: %v\n", err)
		return
	}
	fmt.Printf("Loaded configuration:\n%s\n", string(configBytes))
}
