<script lang="ts">
  import { onMount } from "svelte";
  import tableDataLoader from "@/lib/TableDataLoader";
  import ActivityCodeEditor from "./tabs/ActivityCodeEditor.svelte";
  import CrewInfoEditor from "./tabs/CrewInfoEditor.svelte";
  import PenaltyEditor from "./tabs/PenaltyEditor.svelte";
  import OffDayTable from "./tabs/OffDayTable.svelte";
  import AircraftCrewNeedEditor from "./tabs/AircraftCrewNeedEditor.svelte";

  const tabs = [
    "Activity Codes",
    "Off Day Table",
    "Crew Info",
    "Penalties",
    "Aircraft Crew Need",
  ];
  let activeTab = 0;

  onMount(async () => {
    if (!tableDataLoader.isLoaded) {
      await tableDataLoader.loadAll();
    }
  });
</script>

<div class="tabs">
  {#each tabs as tab, i}
    <button class:selected={i === activeTab} on:click={() => (activeTab = i)}>
      {tab}
    </button>
  {/each}
</div>

<div class="tab-content">
  {#if activeTab === 0}
    <ActivityCodeEditor />
  {:else if activeTab === 1}
    <OffDayTable />
  {:else if activeTab === 2}
    <CrewInfoEditor />
  {:else if activeTab === 3}
    <PenaltyEditor />
  {:else if activeTab === 4}
    <AircraftCrewNeedEditor />
  {/if}
</div>

<style>
  .tabs {
    display: flex;
    border-bottom: 1px solid #ccc;
    margin-bottom: 10px;
  }

  .tabs button {
    padding: 8px 16px;
    cursor: pointer;
    background: none;
    border: none;
    border-bottom: 2px solid transparent;
    font-weight: bold;
  }

  .tabs button.selected {
    border-bottom: 2px solid #007bff;
    color: #007bff;
  }

  .tab-content {
    margin-top: 1rem;
  }
</style>
