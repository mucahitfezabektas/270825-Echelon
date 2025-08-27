<script lang="ts">
  import CustomAlertModal from "@/components/CustomAlertModal.svelte";
  import { alertStore, hideAlert } from "@/stores/alertStore";
  $: ({ isOpen, message, title } = $alertStore);
</script>

<slot />

<CustomAlertModal {isOpen} {message} {title} on:close={hideAlert} />

<style>
  :global(body) {
    margin: 0;
    padding: 0;
    box-sizing: border-box;
    overflow: hidden;
    font-family: "Inter", sans-serif;
    color: var(--thy-text-color);
    background-color: var(--thy-light-gray);
  }
</style>
