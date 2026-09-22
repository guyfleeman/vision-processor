package geometry

import (
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/RoboCup-SSL/ssl-vision-processor/gui/internal/vision"
)

func testConfig() FieldConfig {
	return FieldConfig{
		FieldLength:   2160,
		FieldWidth:    1680,
		GoalWidth:     280,
		GoalDepth:     50,
		BoundaryWidth: 100,
		LineThickness: 10,
	}
}

func testOptional() OptionalLinesConfig {
	return OptionalLinesConfig{Halfway: true, Penalty: true}
}

func TestFieldConfigRoundTripsThroughGet(t *testing.T) {
	g := testGeometry(t)

	cfg := testConfig()
	opt := testOptional()

	if err := g.UpdateField(cfg, opt); err != nil {
		t.Fatalf("UpdateField: %v", err)
	}

	gotCfg, gotOpt := g.FieldConfig()
	if gotCfg != cfg {
		t.Errorf("FieldConfig() = %+v, want %+v", gotCfg, cfg)
	}

	if gotOpt != opt {
		t.Errorf("optional = %+v, want %+v", gotOpt, opt)
	}
}

func TestUpdateFieldRegeneratesMarkings(t *testing.T) {
	g := testGeometry(t)

	if err := g.UpdateField(testConfig(), testOptional()); err != nil {
		t.Fatalf("UpdateField: %v", err)
	}

	lines := g.Snapshot().GetGeometry().GetField().GetFieldLines()

	// Halfway is on, goal2goal/centercircle are off: expect the 4 mandatory
	// lines + HalfwayLine + the 6 penalty stretches, nothing else.
	if len(lines) != 11 {
		t.Fatalf("field lines = %d, want 11 (got %v)", len(lines), lines)
	}
}

func TestUpdateFieldRejectsNonsenseDimensions(t *testing.T) {
	g := testGeometry(t)

	cfg := testConfig()
	cfg.FieldLength = 0

	if err := g.UpdateField(cfg, testOptional()); err == nil {
		t.Fatal("UpdateField accepted a zero field_length")
	}
}

func TestUpdateFieldKeepsExistingCalibrations(t *testing.T) {
	g := testGeometry(t)
	g.Absorb(&vision.SSL_GeometryData{Calib: []*vision.SSL_GeometryCameraCalibration{calib(0, 400)}})

	if err := g.UpdateField(testConfig(), testOptional()); err != nil {
		t.Fatalf("UpdateField: %v", err)
	}

	if got := g.Snapshot().GetGeometry().GetCalib(); len(got) != 1 {
		t.Fatalf("calib = %v, want the pre-existing entry kept", got)
	}
}

// UpdateField persists to the file Geometry was loaded from. A second Load of
// that same file must see the new dimensions, and "models" (which UpdateField
// never touches) must have survived the round trip unchanged.
func TestUpdateFieldPersistsAndPreservesExtraKeys(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "geometry.yml")

	src, err := os.ReadFile("testdata/geometry.yml")
	if err != nil {
		t.Fatalf("ReadFile fixture: %v", err)
	}

	fixtureWithModels := string(src) + "\nmodels:\n  straight_two_phase:\n    acc_slide: -3.4\n    acc_roll: -0.45\n    k_switch: 0.64\n"
	if err := os.WriteFile(path, []byte(fixtureWithModels), 0o644); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}

	g, err := New(path)
	if err != nil {
		t.Fatalf("New: %v", err)
	}

	cfg := testConfig()
	if err := g.UpdateField(cfg, testOptional()); err != nil {
		t.Fatalf("UpdateField: %v", err)
	}

	reloaded, err := New(path)
	if err != nil {
		t.Fatalf("New (reload): %v", err)
	}

	gotCfg, _ := reloaded.FieldConfig()
	if gotCfg.FieldLength != cfg.FieldLength {
		t.Errorf("reloaded field_length = %d, want %d", gotCfg.FieldLength, cfg.FieldLength)
	}

	models := reloaded.Snapshot().GetGeometry().GetModels()
	if models.GetStraightTwoPhase().GetAccSlide() != -3.4 {
		t.Errorf("models.straight_two_phase.acc_slide did not survive the save: %v", models)
	}
}

// UpdateField must refuse to overwrite a file whose name matches a protected
// rulebook preset -- and must not mutate any in-memory state on the way to
// refusing, since a rejected save that still changed what's held in memory
// would get broadcast over multicast despite never reaching disk.
func TestUpdateFieldRefusesToOverwriteAPreset(t *testing.T) {
	src, err := os.ReadFile("testdata/geometry.yml")
	if err != nil {
		t.Fatalf("ReadFile fixture: %v", err)
	}

	path := filepath.Join(t.TempDir(), "geometry-divB.yml")
	if err := os.WriteFile(path, src, 0o644); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}

	g, err := New(path)
	if err != nil {
		t.Fatalf("New: %v", err)
	}

	before, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("ReadFile before: %v", err)
	}

	beforeCfg, _ := g.FieldConfig()

	var readOnly *ReadOnlyError
	if err := g.UpdateField(testConfig(), testOptional()); !errors.As(err, &readOnly) {
		t.Fatalf("UpdateField err = %v, want a *ReadOnlyError", err)
	}

	after, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("ReadFile after: %v", err)
	}

	if string(before) != string(after) {
		t.Error("file was modified despite the rejected save")
	}

	afterCfg, _ := g.FieldConfig()
	if afterCfg != beforeCfg {
		t.Errorf("in-memory config changed despite the rejected save: got %+v, want %+v", afterCfg, beforeCfg)
	}
}

func TestLoadPresetReadsWithoutAGeometryInstance(t *testing.T) {
	cfg, opt, err := LoadPreset("testdata/geometry.yml")
	if err != nil {
		t.Fatalf("LoadPreset: %v", err)
	}

	if cfg.FieldLength != 9000 {
		t.Errorf("field_length = %d, want 9000", cfg.FieldLength)
	}

	if !opt.Halfway {
		t.Error("halfway = false, want true (per the fixture)")
	}
}
