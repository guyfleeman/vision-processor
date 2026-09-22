package main

import (
	"os"
	"path/filepath"
	"testing"
)

func TestBootstrapGeometryFileSeedsFromPreset(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "geometry.yml")

	if err := bootstrapGeometryFile(path, "testdata/geometry.yml"); err != nil {
		t.Fatalf("bootstrapGeometryFile: %v", err)
	}

	got, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("ReadFile: %v", err)
	}

	want, err := os.ReadFile("testdata/geometry.yml")
	if err != nil {
		t.Fatalf("ReadFile preset: %v", err)
	}

	if string(got) != string(want) {
		t.Error("bootstrapped file does not match the preset byte-for-byte")
	}
}

func TestBootstrapGeometryFileLeavesExistingFileAlone(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "geometry.yml")

	if err := os.WriteFile(path, []byte("already here"), 0o644); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}

	if err := bootstrapGeometryFile(path, "testdata/geometry.yml"); err != nil {
		t.Fatalf("bootstrapGeometryFile: %v", err)
	}

	got, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("ReadFile: %v", err)
	}

	if string(got) != "already here" {
		t.Errorf("existing file was overwritten: %q", got)
	}
}
