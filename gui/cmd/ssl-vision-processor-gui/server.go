package main

import (
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"

	"github.com/RoboCup-SSL/ssl-vision-processor/gui/internal/geometry"
	"github.com/RoboCup-SSL/ssl-vision-processor/gui/internal/hub"
	"google.golang.org/protobuf/encoding/protojson"
)

type VisionServer struct {
	geometry *geometry.Geometry
	hub      *hub.Hub
	imgDir   string
}

func NewVisionServer(geom *geometry.Geometry, wsHub *hub.Hub, imgDir string) http.Handler {
	s := &VisionServer{geometry: geom, hub: wsHub, imgDir: imgDir}

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

// handleGetGeometry serves the current wrapper packet as canonical protojson,
// matching the wire encoding the rest of the league already expects.
func (s *VisionServer) handleGetGeometry() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		data, err := protojson.Marshal(s.geometry.Snapshot())
		if err != nil {
			slog.Error("marshalling geometry", "err", err)
			http.Error(w, "internal error", http.StatusInternalServerError)

			return
		}

		w.Header().Set("Content-Type", "application/json")

		if _, err := w.Write(data); err != nil {
			slog.Error("writing geometry response", "err", err)
		}
	}
}

// fieldConfigResponse is the "virtual field" shape: just the editable
// dimensions and optional-line toggles, not the full wrapper packet (calib,
// source, generated field_lines/arcs) that handleGetGeometry serves.
type fieldConfigResponse struct {
	Field              geometry.FieldConfig         `json:"field"`
	OptionalFieldLines geometry.OptionalLinesConfig `json:"optionalFieldLines"`
}

func (s *VisionServer) handleGetFieldConfig() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		field, optional := s.geometry.FieldConfig()

		w.Header().Set("Content-Type", "application/json")

		if err := json.NewEncoder(w).Encode(fieldConfigResponse{Field: field, OptionalFieldLines: optional}); err != nil {
			slog.Error("writing field config response", "err", err)
		}
	}
}

// handlePutFieldConfig edits the virtual field: dimensions and which optional
// markings exist. Regenerates the derived field lines/arcs and persists to
// geometry.yml -- see geometry.Geometry.UpdateField.
func (s *VisionServer) handlePutFieldConfig() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req fieldConfigResponse
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "malformed request body: "+err.Error(), http.StatusBadRequest)

			return
		}

		if err := s.geometry.UpdateField(req.Field, req.OptionalFieldLines); err != nil {
			var validation *geometry.ValidationError
			var readOnly *geometry.ReadOnlyError

			switch {
			case errors.As(err, &validation), errors.As(err, &readOnly):
				http.Error(w, err.Error(), http.StatusBadRequest)
			default:
				slog.Error("updating field config", "err", err)
				http.Error(w, "internal error", http.StatusInternalServerError)
			}

			return
		}

		w.WriteHeader(http.StatusNoContent)
	}
}

// fieldPresets are the rulebook defaults, read live from the same files a
// human would open (geometry-divA.yml, geometry-divB.yml at the repo root) --
// not a copy that could drift from them. See geometry.ReadOnlyFiles: saves
// through handlePutFieldConfig refuse to touch these paths.
var fieldPresets = []struct {
	Name string
	Path string
}{
	{Name: "Division A", Path: "geometry-divA.yml"},
	{Name: "Division B", Path: "geometry-divB.yml"},
}

type fieldPresetResponse struct {
	Name string `json:"name"`
	fieldConfigResponse
}

// handleGetFieldPresets serves the rulebook presets for the "start from a
// preset" feature. A preset file that can't be read (wrong working
// directory, for one) is skipped with a warning rather than failing the
// whole request -- the other preset(s) are still useful.
func (s *VisionServer) handleGetFieldPresets() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		presets := make([]fieldPresetResponse, 0, len(fieldPresets))

		for _, p := range fieldPresets {
			field, optional, err := geometry.LoadPreset(p.Path)
			if err != nil {
				slog.Warn("skipping unreadable field preset", "name", p.Name, "path", p.Path, "err", err)

				continue
			}

			presets = append(presets, fieldPresetResponse{
				Name:                p.Name,
				fieldConfigResponse: fieldConfigResponse{Field: field, OptionalFieldLines: optional},
			})
		}

		w.Header().Set("Content-Type", "application/json")

		if err := json.NewEncoder(w).Encode(presets); err != nil {
			slog.Error("writing field presets response", "err", err)
		}
	}
}
