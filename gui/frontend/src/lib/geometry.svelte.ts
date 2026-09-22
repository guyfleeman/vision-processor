// Shared state for the virtual field editor. A module-level $state object,
// per the project convention (this replaces a store; see CLAUDE.md).
import type {
  SSL_GeometryFieldSizeJson,
  SSL_FieldLineSegmentJson,
  SSL_FieldCircularArcJson,
} from "../proto/vision/ssl_vision_geometry_pb";

// The editable dimensions: everything the generated field size carries except
// fieldLines/fieldArcs, which are derived server-side and never hand-edited.
export type FieldConfig = Omit<
  SSL_GeometryFieldSizeJson,
  "fieldLines" | "fieldArcs"
>;

export interface OptionalFieldLines {
  goal2Goal: boolean;
  halfway: boolean;
  centerCircle: boolean;
  penalty: boolean;
}

interface FieldConfigResponse {
  field: FieldConfig;
  optionalFieldLines: OptionalFieldLines;
}

// A rulebook preset, as served by GET /api/geometry/presets -- read live
// from geometry-divA.yml/geometry-divB.yml on the Go host, not duplicated
// here. See fieldPresets.ts for what's done with one once loaded.
export interface FieldPreset extends FieldConfigResponse {
  name: string;
}

// Just enough of the full /api/geometry response to render the sketch -- not
// calib, not source, nothing else this editor doesn't need.
interface GeometryResponse {
  geometry?: {
    field?: {
      fieldLines?: SSL_FieldLineSegmentJson[];
      fieldArcs?: SSL_FieldCircularArcJson[];
    };
  };
}

export const virtualField = $state<{
  field: FieldConfig;
  optionalFieldLines: OptionalFieldLines;
  fieldLines: SSL_FieldLineSegmentJson[];
  fieldArcs: SSL_FieldCircularArcJson[];
  presets: FieldPreset[];
  loading: boolean;
  saving: boolean;
  error: string | null;
  dirty: boolean;
}>({
  field: {},
  optionalFieldLines: {
    goal2Goal: false,
    halfway: false,
    centerCircle: false,
    penalty: false,
  },
  fieldLines: [],
  fieldArcs: [],
  presets: [],
  loading: false,
  saving: false,
  error: null,
  dirty: false,
});

// Presets are loaded separately from loadVirtualField: a preset endpoint that
// can't be read (see the Go handler's graceful degradation) shouldn't stop
// the actual field config from loading, and vice versa.
export async function loadFieldPresets(): Promise<void> {
  try {
    const response = await fetch("/api/geometry/presets");
    if (response.ok) {
      virtualField.presets = (await response.json()) as FieldPreset[];
    }
  } catch {
    // Presets are a convenience, not required for the editor to work -- an
    // empty list just means the preset section has nothing to offer.
  }
}

export async function loadVirtualField(): Promise<void> {
  virtualField.loading = true;
  virtualField.error = null;

  try {
    const [configRes, geometryRes] = await Promise.all([
      fetch("/api/geometry/field"),
      fetch("/api/geometry"),
    ]);

    if (!configRes.ok) {
      throw new Error(`GET /api/geometry/field: ${String(configRes.status)}`);
    }

    const config = (await configRes.json()) as FieldConfigResponse;
    virtualField.field = config.field;
    virtualField.optionalFieldLines = config.optionalFieldLines;
    virtualField.dirty = false;

    if (geometryRes.ok) {
      const geometry = (await geometryRes.json()) as GeometryResponse;
      virtualField.fieldLines = geometry.geometry?.field?.fieldLines ?? [];
      virtualField.fieldArcs = geometry.geometry?.field?.fieldArcs ?? [];
    }
  } catch (err) {
    virtualField.error = err instanceof Error ? err.message : String(err);
  } finally {
    virtualField.loading = false;
  }
}

export async function saveVirtualField(): Promise<void> {
  virtualField.saving = true;
  virtualField.error = null;

  try {
    const response = await fetch("/api/geometry/field", {
      method: "PUT",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({
        field: virtualField.field,
        optionalFieldLines: virtualField.optionalFieldLines,
      } satisfies FieldConfigResponse),
    });

    if (!response.ok) {
      throw new Error(await response.text());
    }

    virtualField.dirty = false;

    // The saved dimensions are also the source of truth for the sketch, and
    // the server just regenerated fieldLines/fieldArcs from them.
    const geometryRes = await fetch("/api/geometry");
    if (geometryRes.ok) {
      const geometry = (await geometryRes.json()) as GeometryResponse;
      virtualField.fieldLines = geometry.geometry?.field?.fieldLines ?? [];
      virtualField.fieldArcs = geometry.geometry?.field?.fieldArcs ?? [];
    }
  } catch (err) {
    virtualField.error = err instanceof Error ? err.message : String(err);
  } finally {
    virtualField.saving = false;
  }
}
