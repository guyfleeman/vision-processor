<script lang="ts">
  import type { ConfigCategory } from "../layout/configCategories";
  import type { VisionInstance } from "../layout/nav.svelte";

  interface Props {
    category: ConfigCategory;
    instance: VisionInstance | undefined;
  }

  let { category, instance }: Props = $props();
</script>

<section class="placeholder">
  <h2>{category.label}</h2>

  {#if instance}
    <p class="hint">
      Editing {instance.host} / cam {instance.cameraId}'s
      <code>{category.yamlKey}:</code>
      block. Not wired to a backend yet -- there is nowhere to read or write a specific
      instance's config.yml over the network. See
      <code>internal/config</code> in gui/CLAUDE.md's "Not yet built".
    </p>
  {:else}
    <p class="hint">Select a vision processor on the left first.</p>
  {/if}

  <p class="build-note">
    Contributing this panel? Replace this file with a real form for the fields
    below (see config.yml at the repo root for exact defaults/ranges).
  </p>

  <dl>
    {#each category.fields as field (field.name)}
      <div class="field">
        <dt>{field.name}</dt>
        <dd>{field.comment}</dd>
      </div>
    {/each}
  </dl>
</section>

<style>
  .placeholder {
    max-width: 640px;
  }

  h2 {
    margin: 0 0 0.5rem;
  }

  .hint {
    color: #666;
    font-size: 0.85rem;
  }

  .build-note {
    color: #a15c00;
    background: #fff6e5;
    border: 1px solid #ffe1a8;
    border-radius: 4px;
    padding: 0.5rem 0.75rem;
    font-size: 0.8rem;
  }

  dl {
    margin: 1rem 0 0;
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
