package main

import (
	"net/http"
)

func (s *VisionServer) addRoutes(mux *http.ServeMux) {
	mux.Handle("GET /api/health", s.handleHealth())
}