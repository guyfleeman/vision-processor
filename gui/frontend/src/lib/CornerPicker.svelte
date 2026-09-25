<script lang="ts">
  import { onMount } from "svelte";

  // Hacky first pass at the corner-drag picker from the calibration UI plan.
  // Web-only: no C++ changes, camera hardcoded to 0 (no camera selector yet).
  //
  // Ordering, traced from src/calib/GeomModel.cpp's cornerCalibration: only
  // the first point in line_corners matters. It must be the corner where the
  // goal line meets the touchline at the field's (minX, minY) -- the other
  // three can be in any order, since the algorithm brute-force searches every
  // valid clockwise rotation of them and keeps whichever fits best. So this
  // picker just needs the user to mark ONE corner, not sort all four.

  const camId = "0";
  const view = "geomcalib_input"; // GeomModel.cpp's calibration input snapshot

  let cacheBuster = $state(Date.now());
  let imageWidth = $state(0);
  let imageHeight = $state(0);

  // Default to an inset rectangle rather than the image edges, so all four
  // handles are visible and draggable as soon as the image loads.
  let corners = $state<{ x: number; y: number }[]>([]);

  let svgEl: SVGSVGElement | undefined = $state();
  let draggingIndex = $state<number | null>(null);

  // Which marker is the (minX, minY) corner -- see the note above. Defaults
  // to the first handle; click a different one to change it.
  let originIndex = $state(0);

  function handleImageLoad(img: HTMLImageElement): void {
    imageWidth = img.naturalWidth;
    imageHeight = img.naturalHeight;

    if (corners.length === 0) {
      const marginX = imageWidth * 0.15;
      const marginY = imageHeight * 0.15;
      corners = [
        { x: marginX, y: marginY },
        { x: imageWidth - marginX, y: marginY },
        { x: imageWidth - marginX, y: imageHeight - marginY },
        { x: marginX, y: imageHeight - marginY },
      ];
    }
  }

  // Screen pixels -> SVG user-space (== image pixel space, since viewBox is
  // set to the image's natural dimensions). Using the SVG's own CTM handles
  // however the browser has scaled it, rather than reimplementing that math.
  function toImagePoint(
    clientX: number,
    clientY: number,
  ): { x: number; y: number } {
    if (!svgEl) return { x: 0, y: 0 };

    const pt = svgEl.createSVGPoint();
    pt.x = clientX;
    pt.y = clientY;

    const ctm = svgEl.getScreenCTM();
    if (!ctm) return { x: 0, y: 0 };

    const p = pt.matrixTransform(ctm.inverse());
    return { x: p.x, y: p.y };
  }

  function startDrag(index: number, event: PointerEvent): void {
    draggingIndex = index;
    (event.target as Element).setPointerCapture(event.pointerId);
  }

  function onDrag(event: PointerEvent): void {
    if (draggingIndex === null) return;

    const p = toImagePoint(event.clientX, event.clientY);
    corners[draggingIndex] = {
      x: Math.round(Math.max(0, Math.min(imageWidth, p.x))),
      y: Math.round(Math.max(0, Math.min(imageHeight, p.y))),
    };
    savedAt = null;
  }

  function endDrag(): void {
    draggingIndex = null;
  }

  function refreshFrame(): void {
    cacheBuster = Date.now();
  }

  // The chosen origin corner first, the other three following in whatever
  // order they're already in -- correct per the note above, no sorting.
  let orderedCorners = $derived.by(() => {
    const origin = corners[originIndex];
    if (!origin) return [];

    return [origin, ...corners.filter((_, i) => i !== originIndex)];
  });

  let yamlSnippet = $derived(
    "line_corners:\n" +
      orderedCorners
        .map((c) => `- [${String(c.x)}, ${String(c.y)}]`)
        .join("\n"),
  );

  let saving = $state(false);
  let saveError = $state<string | null>(null);
  let savedAt = $state<number | null>(null);

  // The label/stroke sizes below are SVG user-space units, i.e. image
  // pixels (viewBox == the image's natural size) -- not screen pixels. They
  // must scale with image resolution the same way the handle radius already
  // does, or they shrink toward invisible on a higher-res camera feed than
  // whatever this was last tuned against.
  let handleRadius = $derived(Math.max(imageWidth, imageHeight) * 0.015);
  let handleStrokeWidth = $derived(handleRadius * 0.2);
  let labelFontSize = $derived(handleRadius * 2.5);
  let labelDy = $derived(-(handleRadius * 1.8));
  let labelStrokeWidth = $derived(handleRadius * 0.3);

  // Load whatever was last saved so a reload shows the real calibration
  // instead of always resetting to the default inset rectangle. Only applied
  // if nothing's been placed yet (corners.length === 0) and the file has a
  // full set of 4 -- a partial/absent save just leaves the default in place.
  onMount(() => {
    void loadExistingLineCorners();
  });

  async function loadExistingLineCorners(): Promise<void> {
    try {
      const response = await fetch("/api/config/line-corners");
      if (!response.ok) return;

      const data = (await response.json()) as {
        corners?: { x: number; y: number }[];
      };

      // saveLineCorners always sends the origin corner first (see
      // orderedCorners) -- what comes back is that same saved order, so the
      // origin is always index 0 here. goalSideMarker is not it: that's the
      // *original*, pre-reorder marker number, kept only as a record in
      // config.yml, and doesn't describe this already-reordered array.
      if (corners.length === 0 && data.corners?.length === 4) {
        corners = data.corners;
        originIndex = 0;
      }
    } catch {
      // Nothing saved yet, or it couldn't be read -- the default inset
      // rectangle is a fine starting point.
    }
  }

  async function saveLineCorners(): Promise<void> {
    saving = true;
    saveError = null;

    try {
      const response = await fetch("/api/config/line-corners", {
        method: "PUT",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({
          corners: orderedCorners,
          goalSideMarker: originIndex + 1,
        }),
      });

      if (!response.ok) throw new Error(await response.text());
      savedAt = Date.now();
    } catch (err) {
      saveError = err instanceof Error ? err.message : String(err);
    } finally {
      saving = false;
    }
  }
</script>

<section class="corner-picker">
  <h2>Corner picker (hacky, cam 0 only)</h2>
  <p class="hint">
    Drag the four markers onto the real field corners in the image below, then
    click the number on whichever one sits where the goal line meets the
    touchline nearest this field's (0,0) corner -- that one turns green and
    becomes first in the output. The other three can be in any order; the
    calibration algorithm works that out itself. No camera selector yet.
  </p>

  <button type="button" onclick={refreshFrame}>Refresh frame</button>

  <div
    class="overlay-container"
    onpointermove={onDrag}
    onpointerup={endDrag}
    onpointercancel={endDrag}
  >
    <img
      src={`/api/snapshot/${camId}/${view}?t=${String(cacheBuster)}`}
      alt={`cam ${camId} calibration input`}
      onload={(e: Event) => {
        handleImageLoad(e.currentTarget as HTMLImageElement);
      }}
    />

    {#if imageWidth > 0}
      <svg
        bind:this={svgEl}
        viewBox={`0 0 ${String(imageWidth)} ${String(imageHeight)}`}
        preserveAspectRatio="none"
      >
        {#each corners as corner, index (index)}
          <circle
            cx={corner.x}
            cy={corner.y}
            r={handleRadius}
            stroke-width={handleStrokeWidth}
            class="handle"
            class:origin={index === originIndex}
            onpointerdown={(e) => {
              startDrag(index, e);
            }}
          />
          <text
            x={corner.x}
            y={corner.y}
            dy={labelDy}
            font-size={labelFontSize}
            stroke-width={labelStrokeWidth}
            role="button"
            tabindex="0"
            onclick={() => {
              originIndex = index;
            }}
            onkeydown={(e) => {
              if (e.key === "Enter" || e.key === " ") originIndex = index;
            }}
          >
            {index + 1}
          </text>
        {/each}
        {#if corners.length === 4}
          <polygon
            points={corners
              .map((c) => `${String(c.x)},${String(c.y)}`)
              .join(" ")}
          />
        {/if}
      </svg>
    {/if}
  </div>

  <pre>{yamlSnippet}</pre>

  <div class="save-row">
    <button
      type="button"
      onclick={saveLineCorners}
      disabled={saving || orderedCorners.length !== 4}
    >
      {saving ? "Saving..." : "Save to config.yml"}
    </button>
    {#if savedAt}
      <span class="saved">Saved.</span>
    {/if}
    {#if saveError}
      <span class="error">Error: {saveError}</span>
    {/if}
  </div>
</section>

<style>
  .corner-picker {
    max-width: 900px;
    margin-top: 2rem;
  }

  .overlay-container {
    position: relative;
    max-width: 640px;
    touch-action: none; /* dragging must not scroll the page on touch */
  }

  .overlay-container img {
    display: block;
    width: 100%;
    height: auto;
    background: #222;
  }

  .overlay-container svg {
    position: absolute;
    inset: 0;
    width: 100%;
    height: 100%;
  }

  .handle {
    fill: orange;
    stroke: black;
    cursor: grab;
  }

  .handle.origin {
    fill: #2ecc40;
  }

  .handle:active {
    cursor: grabbing;
  }

  text {
    fill: white;
    text-anchor: middle;
    paint-order: stroke;
    stroke: black;
    cursor: pointer;
  }

  polygon {
    fill: rgba(255, 165, 0, 0.15);
    stroke: orange;
    stroke-width: 2;
    pointer-events: none;
  }

  pre {
    background: #f5f5f5;
    padding: 0.75rem;
    border-radius: 4px;
    font-size: 0.85rem;
  }

  .hint {
    color: #888;
    font-size: 0.8rem;
    font-style: italic;
  }

  .save-row {
    display: flex;
    align-items: center;
    gap: 0.5rem;
    margin-top: 0.5rem;
  }

  .saved {
    color: #1b5e20;
    font-size: 0.85rem;
  }

  .error {
    color: #b00020;
    font-size: 0.85rem;
  }
</style>
