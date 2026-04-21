package handlers

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strconv"

	"load-balancer/config"
	"load-balancer/loadbalancing"
	"load-balancer/logging"
	"load-balancer/models"

	"github.com/gorilla/mux"
)

var serverMap = make(map[int]*http.Server) // Map to track running servers by port

// Get all virtual services
func GetVirtualServices(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	if len(config.VirtualServices) == 0 {
		http.Error(w, "No virtual services available", http.StatusNotFound)
		return
	}
	if err := json.NewEncoder(w).Encode(config.VirtualServices); err != nil {
		http.Error(w, "Error encoding response: "+err.Error(), http.StatusInternalServerError)
		return
	}
}

// Get a specific virtual service by Port (used as ID here)
func GetVirtualService(w http.ResponseWriter, r *http.Request) {
	portStr := mux.Vars(r)["vs_id"]
	port, err := strconv.Atoi(portStr)
	if err != nil {
		http.Error(w, "Invalid ID format", http.StatusBadRequest)
		return
	}

	for _, vs := range config.VirtualServices {
		if vs.Port == port {
			w.Header().Set("Content-Type", "application/json")
			if err := json.NewEncoder(w).Encode(vs); err != nil {
				http.Error(w, "Error encoding response: "+err.Error(), http.StatusInternalServerError)
				return
			}
			return
		}
	}

	http.Error(w, fmt.Sprintf("Virtual service with port %d not found", port), http.StatusNotFound)
}

// Create a new virtual service
func CreateVirtualService(w http.ResponseWriter, r *http.Request) {
	var vs models.VirtualService
	if err := json.NewDecoder(r.Body).Decode(&vs); err != nil {
		log.Printf("Error decoding request body: %v", err)
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		return
	}

	vs.Logger = logging.InitLogging(&vs)
	log.Printf("Creating virtual service on port %d with algorithm %s", vs.Port, vs.Algorithm)

	// Validate and initialize servers
	for _, server := range vs.ServerList {
		server.Health = server.CheckHealth() // Force initial health check
		loadbalancing.UpdateHealthState(server.URL, server.Health)
		log.Printf("Server %s initialized with health status: %v", server.URL, server.Health)
	}

	// Stop any existing server on the same port
	if existingServer, exists := serverMap[vs.Port]; exists {
		log.Printf("Stopping existing server on port %d", vs.Port)
		go func() {
			if err := existingServer.Shutdown(nil); err != nil {
				log.Printf("Error stopping server on port %d: %v", vs.Port, err)
			}
		}()
	}

	// Add to virtual services and start the server
	config.VirtualServices = append(config.VirtualServices, &vs)
	go func() {
		log.Printf("Starting server on port %d", vs.Port)
		vs.Start(loadbalancing.GetHealthyServer)
	}()

	serverMap[vs.Port] = &http.Server{
		Addr:    fmt.Sprintf(":%d", vs.Port),
		Handler: http.DefaultServeMux,
	}

	log.Printf("Virtual service created successfully on port %d", vs.Port)
	w.WriteHeader(http.StatusCreated)
}

// Update a virtual service
func UpdateVirtualService(w http.ResponseWriter, r *http.Request) {
	portStr := mux.Vars(r)["vs_id"]
	port, err := strconv.Atoi(portStr)
	if err != nil {
		http.Error(w, "Invalid ID format", http.StatusBadRequest)
		return
	}

	var updatedVS models.VirtualService
	err = json.NewDecoder(r.Body).Decode(&updatedVS)
	if err != nil {
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		return
	}

	// Initialize logger for the updated virtual service
	updatedVS.Logger = logging.InitLogging(&updatedVS)

	// Stop existing server if running on the same port
	if existingServer, exists := serverMap[port]; exists {
		updatedVS.Logger.Warnf("Stopping existing server on port %d", port)
		go func() {
			if shutdownErr := existingServer.Shutdown(nil); shutdownErr != nil {
				updatedVS.Logger.Errorf("Error shutting down server on port %d: %v", port, shutdownErr)
			}
		}()
	}

	// Update or add the virtual service
	found := false
	for i, vs := range config.VirtualServices {
		if vs.Port == port {
			config.VirtualServices[i] = &updatedVS
			found = true
			break
		}
	}

	if !found {
		config.VirtualServices = append(config.VirtualServices, &updatedVS)
		updatedVS.Logger.Infof("Virtual service on port %d not found; creating a new one", port)
	}

	// Restart the virtual service with updated configuration
	go func() {
		mux := http.NewServeMux()
		mux.HandleFunc("/", func(res http.ResponseWriter, req *http.Request) {
			ForwardRequest(res, req, &updatedVS)
		})

		server := &http.Server{
			Addr:    fmt.Sprintf(":%d", updatedVS.Port),
			Handler: mux,
		}
		serverMap[updatedVS.Port] = server // Track the server

		updatedVS.Logger.Infof("Starting server on port %d", updatedVS.Port)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			updatedVS.Logger.Errorf("Error starting server on port %d: %v", updatedVS.Port, err)
		}
	}()

	updatedVS.Logger.Infof("Virtual service on port %d updated successfully", updatedVS.Port)
	w.WriteHeader(http.StatusOK)
}

// Delete a virtual service
func DeleteVirtualService(w http.ResponseWriter, r *http.Request) {
	portStr := mux.Vars(r)["vs_id"]
	port, err := strconv.Atoi(portStr)
	if err != nil {
		http.Error(w, "Invalid ID format", http.StatusBadRequest)
		return
	}

	for i, vs := range config.VirtualServices {
		if vs.Port == port {
			config.VirtualServices = append(config.VirtualServices[:i], config.VirtualServices[i+1:]...)
			w.WriteHeader(http.StatusNoContent)
			return
		}
	}
	http.Error(w, "Virtual service not found", http.StatusNotFound)
}

func ForwardRequest(res http.ResponseWriter, req *http.Request, vs *models.VirtualService) {
	server, err := loadbalancing.GetHealthyServer(vs)
	if err != nil {
		vs.Logger.Warnf("No healthy hosts available for virtual service on port %d: %v", vs.Port, err)
		http.Error(res, "No healthy hosts available: "+err.Error(), http.StatusServiceUnavailable)
		return
	}
	server.Connections++
	defer func() { server.Connections-- }()

	vs.Logger.Infof("Forwarding request to: %s", server.URL)
	proxyURL, err := url.Parse(server.URL)
	if err != nil {
		http.Error(res, "Invalid server URL: "+err.Error(), http.StatusInternalServerError)
		return
	}

	proxy := httputil.NewSingleHostReverseProxy(proxyURL)
	proxy.ServeHTTP(res, req)
}
