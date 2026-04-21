package main

import (
	"bufio"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"

	"load-balancer/config"
	"load-balancer/handlers"
	"load-balancer/healthcheck"
	"load-balancer/loadbalancing"
	"load-balancer/logging"
	"load-balancer/models"

	"github.com/gorilla/mux"
)

var (
	adminUser string
	adminPass string
)

func initializeVirtualServices() {
	for _, vs := range config.VirtualServices {
		if _, err := models.GetCertificate(vs.Port); err == nil {
			// Start HTTPS server if certificate is available
			go func(vs *models.VirtualService) {
				vs.Logger.Infof("Starting HTTPS server on port %d", vs.Port)
				vs.StartHTTPS(loadbalancing.GetHealthyServer)
			}(vs)
		} else {
			// Start HTTP server otherwise
			go func(vs *models.VirtualService) {
				vs.Logger.Infof("Starting HTTP server on port %d", vs.Port)
				vs.Start(loadbalancing.GetHealthyServer)
			}(vs)
		}
	}
}

func askForConfiguration() {
	reader := bufio.NewReader(os.Stdin)

	for {
		fmt.Print("Enter port for virtual service: ")
		portStr, _ := reader.ReadString('\n')
		portStr = strings.TrimSpace(portStr)

		fmt.Print("Enter load balancing algorithm (e.g., round_robin): ")
		algorithm, _ := reader.ReadString('\n')
		algorithm = strings.TrimSpace(algorithm)

		var serverList []*models.Server
		for {
			fmt.Print("Enter server name (or 'done' to finish): ")
			serverName, _ := reader.ReadString('\n')
			serverName = strings.TrimSpace(serverName)

			if serverName == "done" {
				break
			}

			fmt.Print("Enter server URL: ")
			serverURL, _ := reader.ReadString('\n')
			serverURL = strings.TrimSpace(serverURL)

			fmt.Print("Enter server weight: ")
			var weight int
			fmt.Scanln(&weight)

			serverList = append(serverList, models.NewServer(serverName, serverURL, weight))
		}

		config.VirtualServices = append(config.VirtualServices, &models.VirtualService{
			Port:       mustAtoi(portStr),
			Algorithm:  algorithm,
			ServerList: serverList,
		})

		fmt.Print("Add another virtual service? (y/n): ")
		another, _ := reader.ReadString('\n')
		another = strings.TrimSpace(another)

		if another != "y" {
			break
		}
	}
}

func mustAtoi(input string) int {
	result, err := strconv.Atoi(input)
	if err != nil {
		log.Fatalf("Invalid integer input: %v", err)
	}
	return result
}

func main() {
	// Load user configuration
	userConfig, err := config.LoadUserConfig()
	if err != nil {
		log.Fatalf("Error loading user configuration: %v", err)
	}
	adminUser = userConfig.Username
	adminPass = userConfig.Password

	// Check if configuration file exists
	if !config.CheckConfigFileExists() {
		fmt.Println("Configuration file not found. Enter configuration manually.")
		askForConfiguration()
	} else {
		fmt.Println("Using configuration from file.")
		config.InitConfig()
		config.DisplayConfig()
	}

	// Initialize logging for each virtual service
	for _, vs := range config.VirtualServices {
		vs.Logger = logging.InitLogging(vs)
	}

	// Start health checks for all virtual services
	healthcheck.StartHealthCheck(config.VirtualServices)

	// Start periodic certificate rotation
	models.StartCertificateRotation()

	// Initialize and start virtual services
	initializeVirtualServices()

	// Configure the HTTP router
	r := mux.NewRouter()

	// Apply middlewares
	r.Use(handlers.MiddlewareIPBlacklist)
	r.Use(handlers.BasicAuthMiddleware(adminUser, adminPass)) // Add Basic Authentication middleware

	// Define API routes
	r.HandleFunc("/access/vs", handlers.GetVirtualServices).Methods("GET")
	r.HandleFunc("/access/vs/{vs_id:[0-9]+}", handlers.GetVirtualService).Methods("GET")
	r.HandleFunc("/access/vs", handlers.CreateVirtualService).Methods("POST")
	r.HandleFunc("/access/vs/{vs_id:[0-9]+}", handlers.UpdateVirtualService).Methods("PUT")
	r.HandleFunc("/access/vs/{vs_id:[0-9]+}", handlers.DeleteVirtualService).Methods("DELETE")
	r.HandleFunc("/access/vs/ip-rules", handlers.AddIPRule).Methods("POST")
	r.HandleFunc("/access/vs/certificates", handlers.ListCertificatesHandler).Methods("GET")
	r.HandleFunc("/access/vs/certificates/generate", handlers.GenerateCertificateHandler).Methods("POST")
	r.HandleFunc("/access/vs/certificates/renew/{port:[0-9]+}", handlers.RenewCertificateHandler).Methods("POST")

	// Start the HTTP server
	log.Printf("Server running on port %d", 8080)
	log.Fatal(http.ListenAndServe(":8080", r))
}
