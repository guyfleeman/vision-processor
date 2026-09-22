<script lang="ts">
  import { onMount } from "svelte";
  import {
    virtualField,
    loadVirtualField,
    loadFieldPresets,
    saveVirtualField,
    saveVirtualFieldAs,
    loadVirtualFieldFrom,
  } from "./geometry.svelte";
  import FieldSketch from "./FieldSketch.svelte";
  import { scalePreset } from "./fieldPresets";
  import { computeFieldSlice } from "./fieldSplit";

  // "Half field" is a data-entry convenience, not a wire concept: the backend
  // (SSL_GeometryFieldSize.field_length) only ever means the FULL field.
  // virtualField.field.fieldLength keeps that meaning always; this toggle
  // just changes what the length input shows and how it's written back, so
  // switching modes never silently mutates already-loaded data. See
  // gui/CLAUDE.md and src/CameraModel.cpp's visibleFieldExtentEstimation for
  // why length (not width) is what gets halved.
  let fieldLayout = $state<"full" | "half">("full");

  // camAmount here is informational only -- it belongs in each vision
  // processor's own config.yml (SSL_VPConfigGeometry.camera_amount), which
  // this host does not yet read, write, or push to any instance.
  const CAMERA_COUNT_OPTIONS: Record<"full" | "half", number[]> = {
    full: [1, 2, 4],
    half: [1, 2],
  };
  let cameraCount = $state(1);
  let cameraId = $state(0);

  let cameraAmount = $derived(fieldLayout === "half" ? 2 : cameraCount);

  // Which one this instance is, out of cameraAmount -- clamped so a stale
  // selection (e.g. picked "camera 3 of 4" then switched to 2 cameras) can't
  // point past the end.
  let clampedCameraId = $derived(Math.min(cameraId, cameraAmount - 1));

  // The FieldSketch highlight: which portion of the field cameraId is
  // responsible for out of cameraAmount, so the rest of the markings (e.g.
  // the other goal's penalty box) render dimmed.
  let fieldSlice = $derived(
    computeFieldSlice(
      clampedCameraId,
      cameraAmount,
      virtualField.field.fieldLength ?? 0,
      virtualField.field.fieldWidth ?? 0,
    ),
  );

  let halfLength = $derived(
    Math.round((virtualField.field.fieldLength ?? 0) / 2),
  );

  function setHalfLength(value: number): void {
    virtualField.field.fieldLength = value * 2;
    markDirty();
  }

  function setFieldLayout(layout: "full" | "half"): void {
    fieldLayout = layout;

    const options = CAMERA_COUNT_OPTIONS[layout];
    if (!options.includes(cameraCount)) {
      cameraCount = options[0] ?? 1;
    }
  }

  let selectedPresetName = $state("Division B");
  let scalePercent = $state(100);

  function applyPreset(): void {
    const preset = virtualField.presets.find(
      (p) => p.name === selectedPresetName,
    );
    if (!preset) return;

    // Ball/robot size are real hardware, not scaled with a shrunk field --
    // scalePreset always leaves them at the preset's own value.
    virtualField.field = scalePreset(preset, scalePercent);
    markDirty();
  }

  let saveAsPath = $state("");
  let loadPath = $state("");

  function handleSaveAs(): void {
    if (!saveAsPath) return;
    void saveVirtualFieldAs(saveAsPath);
  }

  function handleLoad(): void {
    if (!loadPath) return;
    if (virtualField.dirty && !confirm("Discard unsaved changes?")) return;
    void loadVirtualFieldFrom(loadPath);
  }

  // Dimension fields, paired with their FieldConfig key and a label. Order
  // matches how someone would actually measure a field: outer boundary first,
  // then the goal, then the penalty area, then the smaller markings.
  // fieldLength is handled separately above (half/full aware).
  const dimensionFields: {
    key: keyof typeof virtualField.field;
    label: string;
  }[] = [
    { key: "fieldWidth", label: "Field width" },
    { key: "boundaryWidth", label: "Boundary width" },
    { key: "boundaryWidthGoalLine", label: "Boundary width (goal line)" },
    { key: "goalWidth", label: "Goal width" },
    { key: "goalDepth", label: "Goal depth" },
    { key: "goalHeight", label: "Goal height" },
    { key: "penaltyAreaDepth", label: "Penalty area depth" },
    { key: "penaltyAreaWidth", label: "Penalty area width" },
    { key: "centerCircleRadius", label: "Center circle radius" },
    { key: "lineThickness", label: "Line thickness" },
    { key: "ballRadius", label: "Ball radius" },
    { key: "maxRobotRadius", label: "Max robot radius" },
  ];

  const optionalLineFields: {
    key: keyof typeof virtualField.optionalFieldLines;
    label: string;
  }[] = [
    { key: "halfway", label: "Halfway line" },
    { key: "goal2Goal", label: "Center line (goal-to-goal)" },
    { key: "centerCircle", label: "Center circle" },
    { key: "penalty", label: "Penalty area" },
  ];

  onMount(() => {
    void loadVirtualField();
    void loadFieldPresets();
  });

  function markDirty(): void {
    virtualField.dirty = true;
  }

  function handleSave(): void {
    void saveVirtualField();
  }
</script>

<section class="field-editor">
  <div class="title-row">
    <h2>Virtual field</h2>
    <span class="path">{virtualField.path || "(unsaved)"}</span>
  </div>

  {#if virtualField.error}
    <p class="error">Error: {virtualField.error}</p>
  {/if}

  <div class="layout">
    <form
      onsubmit={(e) => {
        e.preventDefault();
        handleSave();
      }}
    >
      <fieldset disabled={virtualField.loading}>
        <legend>File</legend>
        <div class="file-actions">
          <button
            type="submit"
            disabled={virtualField.saving || !virtualField.dirty}
          >
            {virtualField.saving ? "Saving..." : "Save"}
          </button>
          {#if virtualField.dirty && !virtualField.saving}
            <span class="hint">*new changes</span>
          {/if}
        </div>

        <label>
          Save as
          <span>
            <input type="text" bind:value={saveAsPath} placeholder="path.yml" />
            <button type="button" onclick={handleSaveAs} disabled={!saveAsPath}
              >Save as</button
            >
          </span>
        </label>

        <label>
          Load
          <span>
            <input type="text" bind:value={loadPath} placeholder="path.yml" />
            <button type="button" onclick={handleLoad} disabled={!loadPath}
              >Load</button
            >
          </span>
        </label>
      </fieldset>

      <fieldset disabled={virtualField.loading}>
        <legend>Field layout</legend>

        <label class="radio-row">
          <span>
            <input
              type="radio"
              name="fieldLayout"
              checked={fieldLayout === "full"}
              onchange={() => {
                setFieldLayout("full");
              }}
            />
            Full field
          </span>
          <span>
            <input
              type="radio"
              name="fieldLayout"
              checked={fieldLayout === "half"}
              onchange={() => {
                setFieldLayout("half");
              }}
            />
            Half field
          </span>
        </label>

        {#if fieldLayout === "half"}
          <label>
            Half length (full: {virtualField.field.fieldLength ?? 0}mm)
            <input
              type="number"
              value={halfLength}
              oninput={(e) => {
                setHalfLength(e.currentTarget.valueAsNumber);
              }}
            />
          </label>
        {/if}

        <label>
          Cameras (camera_amount: {cameraAmount})
          <select
            value={cameraCount}
            onchange={(e) => {
              cameraCount = Number(e.currentTarget.value);
            }}
          >
            {#each CAMERA_COUNT_OPTIONS[fieldLayout] as count (count)}
              <option value={count}>{count}</option>
            {/each}
          </select>
        </label>

        <label>
          This camera (camera_id)
          <select
            value={clampedCameraId}
            onchange={(e) => {
              cameraId = Number(e.currentTarget.value);
            }}
          >
            {#each Array.from(Array(cameraAmount).keys()) as id (id)}
              <option value={id}>{id}</option>
            {/each}
          </select>
        </label>
      </fieldset>

      <fieldset disabled={virtualField.loading}>
        <legend>Start from a rulebook preset</legend>
        {#if virtualField.presets.length === 0}
          <p class="hint">No presets available.</p>
        {/if}
        <label>
          Division
          <select bind:value={selectedPresetName}>
            {#each virtualField.presets as preset (preset.name)}
              <option value={preset.name}>{preset.name}</option>
            {/each}
          </select>
        </label>
        <label>
          Scale to
          <span>
            <input type="number" bind:value={scalePercent} min="1" max="200" />
            %
          </span>
        </label>
        <button type="button" onclick={applyPreset}>
          Fill in dimensions from this preset
        </button>
      </fieldset>

      <fieldset disabled={virtualField.loading}>
        <legend>Dimensions</legend>
        {#each dimensionFields as { key, label } (key)}
          <label>
            {label}
            <input
              type="number"
              bind:value={virtualField.field[key]}
              oninput={markDirty}
            />
          </label>
        {/each}
      </fieldset>

      <fieldset disabled={virtualField.loading}>
        <legend>Markings present on this field</legend>
        {#each optionalLineFields as { key, label } (key)}
          <label class="checkbox">
            <input
              type="checkbox"
              bind:checked={virtualField.optionalFieldLines[key]}
              onchange={markDirty}
            />
            {label}
          </label>
        {/each}
      </fieldset>
    </form>

    <div class="sketch">
      <FieldSketch
        fieldLength={virtualField.field.fieldLength ?? 0}
        fieldWidth={virtualField.field.fieldWidth ?? 0}
        lines={virtualField.fieldLines}
        arcs={virtualField.fieldArcs}
        slice={fieldSlice}
      />
    </div>
  </div>
</section>

<style>
  .field-editor {
    max-width: 900px;
  }

  .title-row {
    display: flex;
    align-items: baseline;
    gap: 0.75rem;
    margin-bottom: 0.5rem;
  }

  .title-row h2 {
    margin: 0;
  }

  .path {
    font-family: monospace;
    font-size: 0.8rem;
    color: #666;
  }

  .layout {
    display: grid;
    grid-template-columns: 1fr 1fr;
    gap: 1.5rem;
    align-items: start;
  }

  fieldset {
    border: 1px solid #ddd;
    border-radius: 4px;
    margin-bottom: 1rem;
  }

  label {
    display: flex;
    justify-content: space-between;
    gap: 0.5rem;
    margin: 0.4rem 0;
    font-size: 0.85rem;
  }

  label.checkbox {
    justify-content: flex-start;
  }

  label.radio-row {
    flex-direction: column;
    align-items: flex-start;
    gap: 0.2rem;
  }

  label.radio-row span {
    display: flex;
    align-items: center;
    gap: 0.4rem;
    font-weight: normal;
  }

  input[type="number"] {
    width: 6rem;
  }

  input[type="text"] {
    width: 10rem;
  }

  .file-actions {
    display: flex;
    align-items: center;
    gap: 0.5rem;
    margin: 0.4rem 0;
  }

  .error {
    color: #b00020;
  }

  .hint {
    color: #888;
    font-size: 0.8rem;
    font-style: italic;
  }
</style>
