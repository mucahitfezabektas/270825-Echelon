<script lang="ts">
  import SmartTable from "@/windows/tabs/SmartTable.svelte";
  import tableDataLoader from "@/lib/TableDataLoader";
  import type { OffDayTable } from "@/lib/tableDataTypes";
  import { onMount } from "svelte";

  let offDayTable: OffDayTable[] = [];

  onMount(async () => {
    if (!tableDataLoader.isLoaded) {
      await tableDataLoader.loadAll();
    }
    offDayTable = tableDataLoader.offDayTable;
  });
</script>

<SmartTable
  items={offDayTable}
  columns={[
    { key: "work_days", label: "Çalışılan Gün" },
    { key: "off_day_entitlement", label: "Boş Gün Hakedişi" },
    { key: "distribution", label: "Dağılım" },
  ]}
/>
