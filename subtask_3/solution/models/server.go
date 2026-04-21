package models

import (
    "net/http"
    "net/http/httputil"
    "net/url"
)

type Server struct {
    Name         string `json:"name"`
    URL          string `json:"url"`
    Weight       int    `json:"weight"`
    Health       bool   `json:"health"`
    Connections  int    `json:"connections"`
}

func NewServer(name, url string, weight int) *Server {
    return &Server{
        Name:        name,
        URL:         url,
        Weight:      weight,
        Health:      true, // Assume server is healthy initially
        Connections: 0,
    }
}

func (s *Server) CheckHealth() bool {
    resp, err := http.Head(s.URL)
	if err != nil {
		s.Health = false
		return s.Health
	}
	if resp.StatusCode != http.StatusOK {
		s.Health = false
		return s.Health
	}
	s.Health = true
	return s.Health
}

func (s *Server) ForwardRequest(res http.ResponseWriter, req *http.Request) {
    proxyURL, _ := url.Parse(s.URL)
    proxy := httputil.NewSingleHostReverseProxy(proxyURL)
    proxy.ServeHTTP(res, req)
}