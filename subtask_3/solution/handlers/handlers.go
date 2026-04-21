package handlers

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strconv"
	"strings"

	"load-balancer/config"
	"load-balancer/loadbalancing"
	"load-balancer/logging"
	"load-balancer/models"
	"load-balancer/utils"

	"github.com/dgrijalva/jwt-go"
	"github.com/gorilla/mux"
)

var serverMap = make(map[int]*http.Server) // Map to track running servers by port
var blacklistedIPs = map[string]bool{}     // Store blacklisted IPs in a map

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

// AddIPRule handles adding IP rules
func AddIPRule(w http.ResponseWriter, r *http.Request) {
	var rule map[string]string
	if err := json.NewDecoder(r.Body).Decode(&rule); err != nil {
		http.Error(w, "Invalid rule format", http.StatusBadRequest)
		return
	}

	ip, ipExists := rule["ip"]
	action, actionExists := rule["rule"]

	// Validate input fields
	if !ipExists || !actionExists {
		http.Error(w, "Missing 'rule' or 'ip' in payload", http.StatusBadRequest)
		return
	}

	if action == "block" {
		blacklistedIPs[ip] = true
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]string{
			"status": "success",
			"ip":     ip,
			"action": action,
		})
	} else {
		http.Error(w, "Unsupported action", http.StatusBadRequest)
	}
}

// MiddlewareIPBlacklist blocks requests from blacklisted IPs
func MiddlewareIPBlacklist(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Get the IP address of the client
		clientIP := getClientIP(r)

		if blacklistedIPs[clientIP] {
			http.Error(w, "Forbidden: Your IP has been blocked", http.StatusForbidden)
			return
		}

		// Proceed to the next handler if the IP is not blocked
		next.ServeHTTP(w, r)
	})
}

// getClientIP retrieves the client's IP address from the request
func getClientIP(r *http.Request) string {
	ip := r.RemoteAddr
	if idx := strings.LastIndex(ip, ":"); idx != -1 {
		ip = ip[:idx] // Remove the port number from the address
	}
	return ip
}

func MiddlewareJWT(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" || len(authHeader) < 8 || authHeader[:7] != "Bearer " {
			http.Error(w, "Unauthorized: Missing or invalid Authorization header", http.StatusUnauthorized)
			return
		}

		tokenStr := authHeader[7:]
		token, err := jwt.Parse(tokenStr, func(token *jwt.Token) (interface{}, error) {
			return []byte("my_secret_key"), nil
		})

		if err != nil || !token.Valid {
			http.Error(w, "Unauthorized: Invalid or expired token", http.StatusUnauthorized)
			return
		}

		next.ServeHTTP(w, r)
	})
}

// BasicAuthMiddleware applies Basic Authentication to all routes
func BasicAuthMiddleware(username, password string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Extract Authorization header
			authHeader := r.Header.Get("Authorization")
			if authHeader == "" || !strings.HasPrefix(authHeader, "Basic ") {
				w.Header().Set("WWW-Authenticate", `Basic realm="Restricted"`)
				http.Error(w, "Unauthorized: Missing or invalid Authorization header", http.StatusUnauthorized)
				return
			}

			// Decode Basic Auth credentials
			payload, err := base64.StdEncoding.DecodeString(authHeader[len("Basic "):])
			if err != nil {
				http.Error(w, "Unauthorized: Invalid Authorization header", http.StatusUnauthorized)
				return
			}

			parts := strings.SplitN(string(payload), ":", 2)
			if len(parts) != 2 || parts[0] != username || parts[1] != password {
				w.Header().Set("WWW-Authenticate", `Basic realm="Restricted"`)
				http.Error(w, "Unauthorized: Invalid credentials", http.StatusUnauthorized)
				return
			}

			// Credentials verified, proceed to next handler
			next.ServeHTTP(w, r)
		})
	}
}

// GenerateCertificateHandler generates a new SSL certificate and attaches it to a virtual service
func GenerateCertificateHandler(w http.ResponseWriter, r *http.Request) {
	var payload struct {
		CommonName string `json:"commonName"`
		Port       int    `json:"port"`
		Days       int    `json:"days"`
	}

	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		http.Error(w, "Invalid JSON payload", http.StatusBadRequest)
		return
	}

	// Generate the certificate
	certPath, keyPath, err := models.GenerateCertificate(payload.Port, payload.CommonName, payload.Days)
	if err != nil {
		http.Error(w, "Certificate generation failed: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Attach the certificate to the VirtualService
	err = models.UploadCertificate(certPath, keyPath, payload.Port)
	if err != nil {
		http.Error(w, "Failed to attach certificate: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Restart the VirtualService with HTTPS
	for _, vs := range config.VirtualServices {
		if vs.Port == payload.Port {
			go func(vs *models.VirtualService) {
				vs.Logger.Infof("Restarting HTTPS server on port %d with new certificate", vs.Port)
				vs.StartHTTPS(loadbalancing.GetHealthyServer)
			}(vs)
			break
		}
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]string{
		"status": "success",
		"cert":   certPath,
		"key":    keyPath,
		"port":   strconv.Itoa(payload.Port),
	})
}

// renew certificate
func RenewCertificateHandler(w http.ResponseWriter, r *http.Request) {
	portStr := mux.Vars(r)["port"]
	port, err := strconv.Atoi(portStr)
	if err != nil {
		http.Error(w, "Invalid port", http.StatusBadRequest)
		return
	}

	// Check if the certificate exists
	cert := models.Certificates[port]
	if cert == nil {
		http.Error(w, "Certificate not found for port "+portStr, http.StatusNotFound)
		return
	}

	// Renew the certificate
	certPath, keyPath, err := models.RenewCertificate(port, "example.com", 365)
	if err != nil {
		http.Error(w, "Failed to renew certificate: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Update the certificate in the system
	err = models.UploadCertificate(certPath, keyPath, port)
	if err != nil {
		http.Error(w, "Failed to update certificate: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Restart the VirtualService with the renewed certificate
	for _, vs := range config.VirtualServices {
		if vs.Port == port {
			go func(vs *models.VirtualService) {
				vs.Logger.Infof("Restarting HTTPS server on port %d with renewed certificate", vs.Port)
				vs.StartHTTPS(loadbalancing.GetHealthyServer)
			}(vs)
			break
		}
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{
		"status": "renewed",
		"port":   portStr,
	})
}

// list certificates
func ListCertificatesHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	certDetails := make(map[int]map[string]string)
	for port, cert := range models.Certificates {
		certDetails[port] = map[string]string{
			"certPath": cert.CertPath,
			"keyPath":  cert.KeyPath,
		}
	}
	json.NewEncoder(w).Encode(certDetails)
}

// rate limiting middleware
func MiddlewareSlidingWindowRateLimiting(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		portStr := mux.Vars(r)["vs_id"]

		if portStr == "" {
			next.ServeHTTP(w, r)
			return
		}

		port, err := strconv.Atoi(portStr)
		if err != nil {
			http.Error(w, "Invalid port format", http.StatusBadRequest)
			return
		}

		vs := getVirtualServiceByPort(port)
		if vs == nil {
			http.Error(w, "Virtual service not found", http.StatusNotFound)
			return
		}

		// Apply sliding window rate limiting
		if !utils.AllowRequestSlidingWindow(config.RedisClient, port, 60, vs.RateLimit) {
			http.Error(w, vs.Message, vs.StatusCode)
			return
		}

		next.ServeHTTP(w, r)
	})
}

// Helper to fetch the latest VirtualService
func getVirtualServiceByPort(port int) *models.VirtualService {
	for _, vs := range config.VirtualServices {
		if vs.Port == port {
			return vs
		}
	}
	return nil
}

// rate limiting handler
func UpdateRateLimit(w http.ResponseWriter, r *http.Request) {
	var payload struct {
		Port       int    `json:"port"`
		RateLimit  int    `json:"rate_limit"`
		StatusCode int    `json:"status_code"`
		Message    string `json:"message"`
	}

	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		return
	}

	updated := false
	for _, vs := range config.VirtualServices {
		if vs.Port == payload.Port {
			// Update rate limit configuration
			vs.RateLimit = payload.RateLimit
			vs.StatusCode = payload.StatusCode
			vs.Message = payload.Message
			updated = true

			// Clear Redis keys for the rate-limiting bucket
			redisKey := fmt.Sprintf("rate_limit:%d", payload.Port)
			err := config.RedisClient.Del(context.Background(), redisKey).Err()
			if err != nil {
				fmt.Printf("Failed to reset Redis key %s: %v\n", redisKey, err)
			} else {
				fmt.Printf("Redis key %s cleared\n", redisKey)
			}

			// Restart the service with new rate limits
			fmt.Printf("Restarting VirtualService on port %d with updated rate limits\n", payload.Port)
			models.StopExistingServer(vs.Port, vs.Logger)

			if _, err := models.GetCertificate(vs.Port); err == nil {
				go vs.StartHTTPS(loadbalancing.GetHealthyServer)
			} else {
				go vs.Start(loadbalancing.GetHealthyServer)
			}
			break
		}
	}

	if !updated {
		http.Error(w, "Virtual Service not found", http.StatusNotFound)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte(fmt.Sprintf("Rate limits updated and Redis keys reset for port %d", payload.Port)))
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
