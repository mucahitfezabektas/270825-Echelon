<script lang="ts">
  import Timeline from "@/components/RosterTimeline.svelte";
  import { writable } from "svelte/store";

  /** 0 – 1 arası oran; 0.5 = yarı yarıya */
  const ratio = writable(0.5);

  let wrapper: HTMLDivElement;
  let dragging = false;

  function onPointerDown(event: PointerEvent) {
    dragging = true;
    /* capture pointer ⇒ uzaklaşsa bile olayı al */
    (event.target as HTMLElement).setPointerCapture(event.pointerId);
  }

  function onPointerMove(event: PointerEvent) {
    if (!dragging || !wrapper) return;

    const rect = wrapper.getBoundingClientRect();
    /* imlecin wrapper üstüne göre Y pozisyonu */
    let next = (event.clientY - rect.top) / rect.height;
    /* min–max koruması (örn. 120 px) */
    const min = 120 / rect.height;
    next = Math.min(1 - min, Math.max(min, next));

    ratio.set(next);
  }

  function onPointerUp(event: PointerEvent) {
    dragging = false;
    (event.target as HTMLElement).releasePointerCapture(event.pointerId);
  }
</script>

<div
  bind:this={wrapper}
  class="wrapper"
  on:pointermove={onPointerMove}
  on:pointerup={onPointerUp}
>
  <!-- Üst zaman çizelgesi -->
  <div class="pane" style="flex-basis: {($ratio * 100).toFixed(2)}%;">
    <Timeline timelineId="tl-1" />
  </div>

  <!-- Ayraç -->
  <div
    class="divider"
    on:pointerdown={onPointerDown}
    on:contextmenu|preventDefault
  />

  <!-- Alt zaman çizelgesi -->
  <div class="pane" style="flex-basis: {((1 - $ratio) * 100).toFixed(2)}%;">
    <Timeline timelineId="tl-2" />
  </div>
</div>

<style>
  .wrapper {
    height: 100%;
    display: flex;
    flex-direction: column;
  }
  .pane {
    min-height: 120px;
    overflow: hidden;
  }
  .divider {
    height: 6px; /* tutma kalınlığı */
    cursor: row-resize;
    background: var(--thy-light-gray, #d0d0d0);
    flex-shrink: 0;
  }
  .divider:hover {
    background: #bababa;
  }
</style>
