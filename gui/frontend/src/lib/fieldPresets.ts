// Rulebook preset math: scaling a preset by a percentage. The presets
// themselves are read live from /api/geometry/presets (see
// loadFieldPresets in geometry.svelte.ts) -- straight from the same
// geometry-divA.yml/geometry-divB.yml a human would open, rather than a copy
// kept here that could drift from them.
import type { FieldConfig, FieldPreset } from "./geometry.svelte";

export type { FieldPreset };

// Fields it makes sense to shrink along with the field itself. ballRadius and
// maxRobotRadius are deliberately excluded by default -- they describe real
// hardware, which is usually not scaled down just because the field is.
export const SCALABLE_KEYS: (keyof FieldConfig)[] = [
  "fieldLength",
  "fieldWidth",
  "goalWidth",
  "goalDepth",
  "goalHeight",
  "penaltyAreaDepth",
  "penaltyAreaWidth",
  "goalCenterToPenaltyMark",
  "boundaryWidth",
  "boundaryWidthGoalLine",
  "centerCircleRadius",
  "lineThickness",
];

// Applies a preset at the given percentage scale (100 = actual rule-book
// size) to a FieldConfig, returning a new object. ballRadius/maxRobotRadius
// are real hardware and always left at the preset's value.
export function scalePreset(
  preset: FieldPreset,
  scalePercent: number,
): FieldConfig {
  const factor = scalePercent / 100;
  const result: FieldConfig = { ...preset.field };

  for (const key of SCALABLE_KEYS) {
    const value = preset.field[key];
    if (typeof value === "number") {
      result[key] = Math.round(value * factor);
    }
  }

  return result;
}
