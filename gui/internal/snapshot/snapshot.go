// Package snapshot serves the debug images the C++ vision processor writes to
// disk as <cam_id>.<view>.{jpg,png}, via tmp-then-rename so reads never see a
// torn frame.
//
// This assumes the Go host and the vision processor share a filesystem, which
// only holds for a single-host setup; it does not work for a remote instance.
package snapshot

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"sort"
)

// filename matches "<cam_id>.<view>.<ext>", e.g. "0.raw.jpg". Both captured
// groups are also validated against this same character set before use in a
// filepath.Glob, since Glob treats *, ?, and [ as metacharacters and an
// unvalidated view name could turn a lookup into a directory listing.
var filename = regexp.MustCompile(`^(\d+)\.([a-zA-Z0-9_-]+)\.(?:jpg|jpeg|png)$`)

// Entry identifies one available snapshot. Field names match what the
// frontend already expects.
type Entry struct {
	CamID string `json:"cam_id"`
	View  string `json:"view"`
}

// HandleList serves the set of snapshots currently on disk in dir.
func HandleList(dir string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		entries, err := list(dir)
		if err != nil {
			slog.Error("listing snapshots", "dir", dir, "err", err)
			http.Error(w, "internal error", http.StatusInternalServerError)

			return
		}

		w.Header().Set("Content-Type", "application/json")

		if err := json.NewEncoder(w).Encode(entries); err != nil {
			slog.Error("writing snapshot list response", "err", err)
		}
	}
}

func list(dir string) ([]Entry, error) {
	files, err := os.ReadDir(dir)
	if os.IsNotExist(err) {
		return []Entry{}, nil
	}

	if err != nil {
		return nil, err
	}

	seen := make(map[Entry]struct{})

	for _, f := range files {
		m := filename.FindStringSubmatch(f.Name())
		if m == nil {
			continue
		}

		seen[Entry{CamID: m[1], View: m[2]}] = struct{}{}
	}

	entries := make([]Entry, 0, len(seen))
	for e := range seen {
		entries = append(entries, e)
	}

	// os.ReadDir order is directory-entry order, not usefully sorted for a UI,
	// and map iteration is randomised on top of that -- fix an order so the
	// grid does not reshuffle itself every 5s poll.
	sort.Slice(entries, func(i, j int) bool {
		if entries[i].CamID != entries[j].CamID {
			return entries[i].CamID < entries[j].CamID
		}

		return entries[i].View < entries[j].View
	})

	return entries, nil
}

// validSegment matches one path-value component of a snapshot request: digits
// for a camera id, or the same character set the filename itself uses for a
// view name.
var validSegment = regexp.MustCompile(`^[a-zA-Z0-9_-]+$`)

// HandleGet serves the most recently written {camID}.{view}.* file in dir.
func HandleGet(dir string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		camID := r.PathValue("camID")
		view := r.PathValue("view")

		if !validSegment.MatchString(camID) || !validSegment.MatchString(view) {
			http.NotFound(w, r)

			return
		}

		path, err := latest(dir, camID, view)
		if err != nil {
			slog.Error("finding snapshot", "cam", camID, "view", view, "err", err)
			http.Error(w, "internal error", http.StatusInternalServerError)

			return
		}

		if path == "" {
			http.NotFound(w, r)

			return
		}

		http.ServeFile(w, r, path)
	}
}

// latest returns the most recently modified file matching dir/camID.view.*, or
// "" if none exists.
func latest(dir, camID, view string) (string, error) {
	matches, err := filepath.Glob(filepath.Join(dir, camID+"."+view+".*"))
	if err != nil {
		return "", err
	}

	var (
		newest     string
		newestTime int64
	)

	for _, path := range matches {
		info, err := os.Stat(path)
		if err != nil {
			continue
		}

		if t := info.ModTime().UnixNano(); t > newestTime {
			newest, newestTime = path, t
		}
	}

	return newest, nil
}
