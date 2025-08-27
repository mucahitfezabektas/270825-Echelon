<script lang="ts">
  import { createEventDispatcher } from "svelte";

  export let rowKey: string;
  export let top: number;
  export let height: number = 40;
  export let isActiveDropTarget: boolean = false;
  export let isCollision: boolean = false;

  const dispatch = createEventDispatcher<{
    rowSidebarContextMenu: {
      rowKey: string;
      pageX: number;
      pageY: number;
    };
  }>();
</script>

<div
  class="timeline-row"
  class:active-drop-target={isActiveDropTarget}
  class:collision-target={isCollision}
  role="row"
  aria-label={`Uçuş satırı ${rowKey}`}
  style="top: {top}px; height: {height}px;"
  data-key={rowKey}
  tabindex="-1"
>
  <div
    class="row-label"
    role="group"
    aria-label="Row overlay for {rowKey}"
    on:contextmenu|stopPropagation={(e) => {
      e.preventDefault(); // 🔒 Tarayıcı menüsünü engelle
      dispatch("rowSidebarContextMenu", {
        rowKey,
        pageX: e.pageX,
        pageY: e.pageY,
      });
    }}
  ></div>
</div>

<style>
  .timeline-row {
    position: absolute;
    left: 0;
    right: 0;
    z-index: 2;
    pointer-events: none;
    display: flex;
    align-items: center;
    font-size: 11px;
    user-select: none;
    transition: background-color 0.1s ease-out;
  }

  .row-label {
    pointer-events: auto;
    padding-left: 8px;
    display: flex;
    gap: 6px;
    background-color: transparent;
    width: 200px;
    height: 100%;
  }

  .timeline-row.active-drop-target {
    background-color: rgba(0, 123, 255, 0.1);
  }

  .timeline-row.collision-target {
    background-color: rgba(220, 53, 69, 0.1);
  }

  .row-label:hover {
    background-color: rgba(0, 123, 255, 0.05);
  }
</style>
