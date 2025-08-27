<!-- src/components/RestTooltip.svelte -->
<script lang="ts">
  import { hoverStore } from "@/stores/hoverStore";
  import { derived } from "svelte/store";
  import dayjs from "dayjs";

  const restHover = derived(hoverStore, ($hoverStore) => {
    if (
      $hoverStore?.rest_start &&
      $hoverStore?.rest_end &&
      $hoverStore._boundingBox
    ) {
      return {
        duration: $hoverStore.rest_duration,
        start: dayjs($hoverStore.rest_start).format("HH:mm"),
        end: dayjs($hoverStore.rest_end).format("HH:mm"),
        x: $hoverStore._boundingBox.x + $hoverStore._boundingBox.width + 10,
        y: $hoverStore._boundingBox.y,
      };
    }
    return null;
  });
</script>

{#if $restHover}
  <div
    class="absolute z-50 bg-white text-xs shadow-md border border-gray-300 px-2 py-1 rounded pointer-events-none"
    style="top: {$restHover.y - 25}px; left: {$restHover.x}px;"
  >
    <div><strong>Rest:</strong> {$restHover.duration}h</div>
    <div>{$restHover.start} → {$restHover.end}</div>
  </div>
{/if}
