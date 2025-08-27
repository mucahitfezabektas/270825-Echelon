<script lang="ts">
  import Timeline from "@/components/TimelineContainer.svelte";
</script>

<div class="window-content-wrapper">
  <!-- Timeline -->
  <svelte:component this={Timeline} />
</div>

<style>
  .window-content-wrapper {
    padding: 0;
    font-family: "Inter", sans-serif;
    color: var(--thy-text-color);
    height: 100%;
    display: flex;
    flex-direction: column;
  }
  :global(.window-content-wrapper > .svelte-component) {
    flex-grow: 1;
    overflow: auto;
    padding: 0 1.5rem 1.5rem 1.5rem;
  }
</style>
