package models

import (
	"context"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/sirupsen/logrus"
)

type VirtualService struct {
	Port       int            `json:"port"`
	Algorithm  string         `json:"algorithm"`
	ServerList []*Server      `json:"serverList"`
	Logger     *logrus.Logger `json:"-"`
}

var serverMap = make(map[int]*http.Server) // Tracks running servers by port
var serverMapMutex sync.Mutex              // Ensures thread-safe access to serverMap

// Start initializes an HTTP server for the virtual service
func (vs *VirtualService) Start(getHealthyServer func(*VirtualService) (*Server, error)) {
	// Stop any existing server on the port
	StopExistingServer(vs.Port, vs.Logger)

	mux := http.NewServeMux()
	mux.HandleFunc("/", func(res http.ResponseWriter, req *http.Request) {
		server, err := getHealthyServer(vs)
		if err != nil {
			http.Error(res, "Couldn't process request: "+err.Error(), http.StatusServiceUnavailable)
			return
		}
		server.ForwardRequest(res, req)
	})

	server := &http.Server{
		Addr:    fmt.Sprintf(":%d", vs.Port),
		Handler: mux,
	}

	// Track the server in the serverMap
	serverMapMutex.Lock()
	serverMap[vs.Port] = server
	serverMapMutex.Unlock()

	vs.Logger.Infof("Starting HTTP server on port %d", vs.Port)
	go func() {
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			vs.Logger.Errorf("Error starting HTTP server on port %d: %v", vs.Port, err)
		}
	}()
}

// StartHTTPS initializes an HTTPS server for the virtual service
func (vs *VirtualService) StartHTTPS(getHealthyServer func(*VirtualService) (*Server, error)) {
	StopExistingServer(vs.Port, vs.Logger)

	cert, err := GetCertificate(vs.Port)
	if err != nil {
		vs.Logger.Warnf("HTTPS cannot start on port %d: %v", vs.Port, err)
		return
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/", func(res http.ResponseWriter, req *http.Request) {
		server, err := getHealthyServer(vs)
		if err != nil {
			http.Error(res, "No healthy servers: "+err.Error(), http.StatusServiceUnavailable)
			return
		}
		server.ForwardRequest(res, req)
	})

	server := &http.Server{
		Addr:    fmt.Sprintf(":%d", vs.Port),
		Handler: mux,
	}

	serverMapMutex.Lock()
	serverMap[vs.Port] = server
	serverMapMutex.Unlock()

	vs.Logger.Infof("Starting HTTPS server on port %d", vs.Port)
	go func() {
		if err := server.ListenAndServeTLS(cert.CertPath, cert.KeyPath); err != nil && err != http.ErrServerClosed {
			vs.Logger.Errorf("Error starting HTTPS server on port %d: %v", vs.Port, err)
		}
	}()
}

// Stops any running server on the specified port
func StopExistingServer(port int, logger *logrus.Logger) {
	serverMapMutex.Lock()
	defer serverMapMutex.Unlock()

	if existingServer, exists := serverMap[port]; exists {
		logger.Infof("Stopping existing server on port %d", port)
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := existingServer.Shutdown(ctx); err != nil {
			logger.Errorf("Error stopping server on port %d: %v", port, err)
		}
		delete(serverMap, port)
	}
}
