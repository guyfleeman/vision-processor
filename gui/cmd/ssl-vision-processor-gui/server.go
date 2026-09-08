package main

import (
	"encoding/json"
	"log/slog"
	"net/http"
)

type VisionServer struct {

}

func NewVisionServer() http.Handler {
	s := &VisionServer{}

	mux := http.NewServeMux()
	s.addRoutes(mux)

	return mux
}

func (s *VisionServer) handleHealth() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		if err := json.NewEncoder(w).Encode(map[string]string{"status": "ok"}); err != nil {
			slog.Error("writing health response", "err", err)
		}
	}
}