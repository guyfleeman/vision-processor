package main

import (
	"net/http"

	"github.com/RoboCup-SSL/ssl-vision-processor/gui/frontend"
)

func (s *VisionServer) addRoutes(mux *http.ServeMux) {
	mux.Handle("GET /api/health", s.handleHealth())
	mux.Handle("/api/", http.NotFoundHandler())
	
	mux.Handle("/", frontend.HandleFrontend())
}