<script lang="ts">
  import VirtualList from "@/modules/VirtualList.svelte";
  import tableDataLoader from "@/lib/TableDataLoader";
  import type { ActivityCode } from "@/lib/tableDataTypes";
  import { onMount } from "svelte";
  import { debounce } from "@/lib/constants/utils/utils";

  let activityCodes: ActivityCode[] = [];
  let filteredCodes: ActivityCode[] = [];
  let search = "";
  let sortKey: keyof ActivityCode = "unique_id";
  let sortAsc = true;

  const applyFilterAndSortDebounced = debounce(applyFilterAndSort, 300);

  onMount(async () => {
    if (!tableDataLoader.isLoaded) {
      await tableDataLoader.loadAll();
    }

    activityCodes = tableDataLoader.activityCodes;
    applyFilterAndSort();
  });

  function applyFilterAndSort() {
    const temp = activityCodes.filter((a) =>
      [
        a.activity_code,
        a.activity_group_code,
        a.activity_code_explanation,
      ].some((val) => val?.toLowerCase().includes(search.toLowerCase()))
    );

    temp.sort((a, b) => {
      const valA = (a[sortKey] ?? "").toString();
      const valB = (b[sortKey] ?? "").toString();
      return sortAsc ? valA.localeCompare(valB) : valB.localeCompare(valA);
    });

    filteredCodes = temp;
  }

  function toggleSort(key: keyof ActivityCode) {
    if (sortKey === key) {
      sortAsc = !sortAsc;
    } else {
      sortKey = key;
      sortAsc = true;
    }
    applyFilterAndSort();
  }
</script>

<div class="header">Aktivite Kodları</div>

<input
  class="search"
  placeholder="Ara..."
  bind:value={search}
  on:input={applyFilterAndSortDebounced}
/>

<table style="width: 100%; border-collapse: collapse;">
  <thead>
    <tr>
      <th on:click={() => toggleSort("activity_code")}>Kod</th>
      <th on:click={() => toggleSort("activity_group_code")}>Grup</th>
      <th on:click={() => toggleSort("activity_code_explanation")}>Açıklama</th>
    </tr>
  </thead>
</table>

<VirtualList
  items={filteredCodes}
  itemHeight={36}
  height={600}
  let:item
  let:index
>
  <div
    style="
      display: flex;
      padding: 8px;
      border-bottom: 1px solid #eee;
      background: {index % 2 === 0 ? '#fdfdfd' : '#f6f6f6'};
    "
  >
    <div style="flex: 1;">{item.activity_code}</div>
    <div style="flex: 1;">{item.activity_group_code}</div>
    <div style="flex: 2;">{item.activity_code_explanation}</div>
  </div>
</VirtualList>
