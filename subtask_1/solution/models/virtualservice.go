package models

import (
	"fmt"
	"net/http"

	"github.com/sirupsen/logrus"
)

type VirtualService struct {
	Port       int            `json:"port"`
	Algorithm  string         `json:"algorithm"`
	ServerList []*Server      `json:"serverList"`
	Logger     *logrus.Logger `json:"-"`
}

func (vs *VirtualService) Start(getHealthyServer func(*VirtualService) (*Server, error)) {
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

	vs.Logger.Infof("Starting server on port %d", vs.Port)
	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		vs.Logger.Fatalf("Error starting server on port %d: %v", vs.Port, err)
	}
}


