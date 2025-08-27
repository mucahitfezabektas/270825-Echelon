<script lang="ts">
  import SmartTable from "@/windows/tabs/SmartTable.svelte";
  import tableDataLoader from "@/lib/TableDataLoader";
  import type { AircraftCrewNeed } from "@/lib/tableDataTypes"; // tableDataTypes.ts modelinizi de kontrol edeceğiz
  import { onMount } from "svelte";

  let aircraftCrewNeed: AircraftCrewNeed[] = [];

  onMount(async () => {
    if (!tableDataLoader.isLoaded) {
      await tableDataLoader.loadAll();
    }
    aircraftCrewNeed = tableDataLoader.aircraftCrewNeed;
    console.log(aircraftCrewNeed);
  });
</script>

<SmartTable
  items={aircraftCrewNeed}
  columns={[
    { key: "actype", label: "Aircraft Type" },
    // ⭐ BURADAKİ KEY'LERİ GELEN JSON VERİSİNDEKİ İSİMLERLE EŞLEŞTİRDİK (snake_case) ⭐
    { key: "c_count", label: "C" },
    { key: "p_count", label: "P" },
    { key: "j_count", label: "J" },
    { key: "ef_count", label: "EF" },
    { key: "a_count", label: "A" },
    { key: "s_count", label: "S" },
    { key: "l_count", label: "L" },
    { key: "ec_count", label: "EC" },
    { key: "t_count", label: "T" },
  ]}
/>
