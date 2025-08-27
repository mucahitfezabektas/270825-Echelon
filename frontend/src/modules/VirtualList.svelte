<script lang="ts">
  import { onMount } from "svelte";

  export let items: any[] = [];
  export let itemHeight: number = 36;
  export let overscan: number = 5;
  export let height: number = 600;

  let container: HTMLDivElement;
  let scrollTop = 0;

  $: totalHeight = items.length * itemHeight;
  $: startIndex = Math.max(Math.floor(scrollTop / itemHeight) - overscan, 0);
  $: endIndex = Math.min(
    Math.ceil((scrollTop + height) / itemHeight) + overscan,
    items.length
  );
  $: visibleItems = items.slice(startIndex, endIndex);

  function handleScroll() {
    scrollTop = container.scrollTop;
  }

  onMount(() => {
    scrollTop = container.scrollTop;
  });
</script>

<div
  bind:this={container}
  on:scroll={handleScroll}
  style="overflow-y: auto; height: {height}px; position: relative;"
>
  <div style="height: {totalHeight}px; position: relative;">
    {#each visibleItems as item, i (startIndex + i)}
      <div
        style="
          position: absolute;
          top: {(startIndex + i) * itemHeight}px;
          height: {itemHeight}px;
          left: 0;
          right: 0;
        "
      >
        <slot {item} index={startIndex + i} />
      </div>
    {/each}
  </div>
</div>
