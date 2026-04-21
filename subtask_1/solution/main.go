package main

import (
	"log"
	"net/http"

	"load-balancer/config"
	"load-balancer/handlers"
	"load-balancer/healthcheck"
	"load-balancer/loadbalancing"
	"load-balancer/logging"

	"github.com/gorilla/mux"
)

func main() {
	// Initialize configuration
	config.InitConfig()

	// Display loaded configuration
	config.DisplayConfig()

	// Initialize logging for each virtual service
	for _, vs := range config.VirtualServices {
		vs.Logger = logging.InitLogging(vs)
	}

	// Start health checks
	healthcheck.StartHealthCheck(config.VirtualServices)

	for _, vs := range config.VirtualServices {
		go vs.Start(loadbalancing.GetHealthyServer)
	}

	// Initialize router
	r := mux.NewRouter()
	r.HandleFunc("/access/vs", handlers.GetVirtualServices).Methods("GET")
	r.HandleFunc("/access/vs/{vs_id:[0-9]+}", handlers.GetVirtualService).Methods("GET")
	r.HandleFunc("/access/vs", handlers.CreateVirtualService).Methods("POST")
	r.HandleFunc("/access/vs/{vs_id:[0-9]+}", handlers.UpdateVirtualService).Methods("PUT")
	r.HandleFunc("/access/vs/{vs_id:[0-9]+}", handlers.DeleteVirtualService).Methods("DELETE")

	// Start HTTP server
	log.Fatal(http.ListenAndServe(":8080", r))

	select {} // Block forever
}
