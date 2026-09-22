<script lang="ts">
  import type { ConfigCategory } from "../layout/configCategories";
  import type { VisionInstance } from "../layout/nav.svelte";
  import CornerPicker from "../CornerPicker.svelte";

  interface Props {
    category: ConfigCategory;
    instance: VisionInstance | undefined;
  }

  let { category, instance }: Props = $props();
</script>

<section class="geometry">
  <h2>Geometry</h2>

  {#if instance}
    <p class="hint">
      {instance.host} / cam {instance.cameraId}. The numeric settings below
      (config.yml's <code>geometry:</code> block) aren't wired to a backend yet; the
      corner picker further down already is (writes to whichever config.yml you tell
      it to, by hand -- see its own hint text).
    </p>
  {:else}
    <p class="hint">Select a vision processor on the left first.</p>
  {/if}

  <dl>
    {#each category.fields as field (field.name)}
      <div class="field">
        <dt>{field.name}</dt>
        <dd>{field.comment}</dd>
      </div>
    {/each}
  </dl>

  <CornerPicker />
</section>

<style>
  .geometry {
    max-width: 900px;
  }

  h2 {
    margin: 0 0 0.5rem;
  }

  .hint {
    color: #666;
    font-size: 0.85rem;
  }

  dl {
    margin: 1rem 0;
  }

  .field {
    display: flex;
    gap: 1rem;
    padding: 0.35rem 0;
    border-bottom: 1px solid #eee;
    font-size: 0.85rem;
  }

  dt {
    flex: 0 0 12rem;
    font-family: monospace;
    font-weight: 600;
  }

  dd {
    margin: 0;
    color: #555;
  }

  code {
    background: #eee;
    padding: 0.1rem 0.3rem;
    border-radius: 3px;
  }
</style>
