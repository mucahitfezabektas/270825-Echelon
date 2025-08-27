<script lang="ts">
  import SmartTable from "@/windows/tabs/SmartTable.svelte";
  import tableDataLoader from "@/lib/TableDataLoader";
  import type { ActivityCode } from "@/lib/tableDataTypes";
  import { onMount } from "svelte";

  let activityCodes: ActivityCode[] = [];

  onMount(async () => {
    if (!tableDataLoader.isLoaded) {
      await tableDataLoader.loadAll();
    }
    activityCodes = tableDataLoader.activityCodes;
    console.log(activityCodes)
  });
</script>

<SmartTable
  items={activityCodes}
  columns={[
    { key: "activity_code", label: "Kod" },
    { key: "activity_group_code", label: "Grup" },
    { key: "activity_code_explanation", label: "Açıklama" },
  ]
  }
/>
