<script lang="ts">
  import { onMount } from "svelte";
  import { topic } from "./lib/wrapper-bus";
  import Shell from "./lib/layout/Shell.svelte";
  import { nav } from "./lib/layout/nav.svelte";

  let subscribed = $state(false);
  const wrapperPacket = topic<Record<string, unknown>>("wrapper_packet.out");

  // Same-origin: the Go host serves /api itself, alongside /ws and the frontend.
  interface Snapshot {
    cam_id: string;
    view: string;
  }
  let snapshots = $state<Snapshot[]>([]);
  let cacheBuster = $state(0);

  // flat/gradient/blob are ball-detection debug views (see src/main.cpp) --
  // color-tuning territory, not relevant to a geometry calibration.
  const COLOR_ONLY_VIEWS = new Set(["flat", "gradient", "blob"]);

  let visibleSnapshots = $derived(
    nav.selectedCategoryId === "geometry"
      ? snapshots.filter((s) => !COLOR_ONLY_VIEWS.has(s.view))
      : snapshots,
  );

  async function refreshSnapshotList(): Promise<void> {
    try {
      const response = await fetch("/api/snapshots");
      if (response.ok) snapshots = (await response.json()) as Snapshot[];
    } catch {
      // ignore: next tick will retry
    }
  }

  onMount(() => {
    void refreshSnapshotList();
    const listId = setInterval(() => void refreshSnapshotList(), 5000);
    const imgId = setInterval(() => {
      cacheBuster = Date.now();
    }, 1000);
    return () => {
      clearInterval(listId);
      clearInterval(imgId);
    };
  });

  function toggleSubscribe(): void {
    subscribed = !subscribed;
  }
</script>

<Shell>
  {#snippet below()}
    <!--
      Everything in this snippet predates the two-column shell (see issue
      #18) and doesn't belong to any config category -- it's a stand-in for
      the mockup's Video + Debug Console panels, which aren't built yet
      (WHEP/live video is still undecided; the Debug Console is a live log
      viewer over internal/hub that hasn't been started). Lift it into those
      once they exist rather than growing it further here.
    -->
    <p class="below-hint">
      Temporary debug utilities -- see the note in this file's source.
    </p>

    {#if nav.selectedCategoryId !== "field"}
      <section>
        <h2>Snapshots</h2>
        {#if visibleSnapshots.length === 0}
          <p class="hint">No images in img/ yet.</p>
        {:else}
          <div class="grid">
            {#each visibleSnapshots as snap (`${snap.cam_id}.${snap.view}`)}
              <figure>
                <img
                  src={`/api/snapshot/${snap.cam_id}/${snap.view}?t=${String(cacheBuster)}`}
                  alt={`cam ${snap.cam_id} ${snap.view}`}
                />
                <figcaption>cam {snap.cam_id} / {snap.view}</figcaption>
              </figure>
            {/each}
          </div>
        {/if}
      </section>
    {/if}

    <section>
      <button onclick={toggleSubscribe}>
        {subscribed ? "Unsubscribe" : "Subscribe to wrapper_packet.out"}
      </button>

      {#if subscribed}
        {#if $wrapperPacket}
          <pre>{JSON.stringify($wrapperPacket, null, 2)}</pre>
        {:else}
          <p class="hint">Waiting for first frame…</p>
        {/if}
      {/if}
    </section>
  {/snippet}
</Shell>

<style>
  .below-hint {
    color: #a15c00;
    background: #fff6e5;
    border: 1px solid #ffe1a8;
    border-radius: 4px;
    padding: 0.4rem 0.75rem;
    font-size: 0.8rem;
    margin: 0 0 1rem;
  }

  .grid {
    display: grid;
    grid-template-columns: 1fr 1fr;
    gap: 0.75rem;
  }

  figure {
    margin: 0;
  }

  img {
    width: 100%;
    height: auto;
    display: block;
    background: #f5f5f5;
  }

  figcaption {
    font-size: 0.75rem;
    color: #666;
    margin-top: 0.25rem;
  }

  h2 {
    font-size: 1rem;
    margin: 0 0 0.5rem;
  }

  pre {
    background: #f5f5f5;
    padding: 1rem;
    border-radius: 4px;
    overflow: auto;
    max-height: 70vh;
    font-size: 0.8rem;
  }

  .hint {
    color: #666;
    font-style: italic;
  }
</style>
