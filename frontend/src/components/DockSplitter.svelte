<script lang="ts">
  import { createEventDispatcher } from "svelte";
  export let x = 0;
  export let height = 0;

  const dispatch = createEventDispatcher();
  let dragging = false;
  let hovered = false;

  function onMouseDown() {
    dragging = true;

    function onMouseMove(e: MouseEvent) {
      if (dragging) {
        dispatch("resizeSplit", { deltaX: e.movementX });
      }
    }

    function onMouseUp() {
      dragging = false;
      window.removeEventListener("mousemove", onMouseMove);
      window.removeEventListener("mouseup", onMouseUp);
    }

    window.addEventListener("mousemove", onMouseMove);
    window.addEventListener("mouseup", onMouseUp);
  }

  function onDoubleClick() {
    dispatch("resetSplit");
  }
</script>

<div
  class="dock-splitter {hovered ? 'hovered' : ''}"
  style="left: {x}px; height: {height}px;"
  on:mousedown={onMouseDown}
  on:dblclick={onDoubleClick}
  on:mouseenter={() => (hovered = true)}
  on:mouseleave={() => (hovered = false)}
/>

<style>
  .dock-splitter {
    position: absolute;
    top: 0;
    width: 6px;
    cursor: col-resize;
    z-index: 1000;
    background-color: transparent;
    transition:
      background-color 0.2s,
      width 0.2s;
  }

  .dock-splitter.hovered {
    background-color: rgba(0, 0, 0, 0.25);
    width: 12px;
    margin-left: -3px; /* center */
  }
</style>
