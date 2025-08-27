<script lang="ts">
  import SmartTable from "@/windows/tabs/SmartTable.svelte";
  import tableDataLoader from "@/lib/TableDataLoader";
  import type { Penalty } from "@/lib/tableDataTypes";
  import { onMount } from "svelte";

  let penalties: Penalty[] = [];

  onMount(async () => {
    if (!tableDataLoader.isLoaded) {
      await tableDataLoader.loadAll();
    }
    penalties = tableDataLoader.penalties;
  });

  const columns = [
    { key: "person_id", label: "ID" },
    { key: "person_name", label: "Ad" },
    { key: "person_surname", label: "Soyad" },
    { key: "ucucu_sinifi", label: "Sınıf" },
    { key: "base_filo", label: "Filo" },
    { key: "penalty_code", label: "Ceza Kodu" },
    { key: "penalty_code_explanation", label: "Açıklama" },
    {
      key: "penalty_start_date",
      label: "Başlangıç",
      format: (v: number) => {
        if (!v) return "";
        const date = new Date(v); // Unix timestamp (milisaniye) doğrudan kullanılır

        // Tarihi GG/AA/YYYY formatında al
        const formattedDate = date
          .toLocaleDateString("tr-TR", {
            day: "2-digit",
            month: "2-digit",
            year: "numeric",
          })
          .replace(/\./g, "/"); // Noktaları slash ile değiştir

        // Saati SS:DD formatında al
        const formattedTime = date.toLocaleTimeString("tr-TR", {
          hour: "2-digit",
          minute: "2-digit",
          hour12: false, // 24 saat formatı için
        });

        return `${formattedDate} ${formattedTime}`; // İki formatı birleştir
      },
    },
    {
      key: "penalty_end_date",
      label: "Bitiş",
      format: (v: number) => {
        if (!v) return "";
        const date = new Date(v); // Unix timestamp (milisaniye) doğrudan kullanılır

        // Tarihi GG/AA/YYYY formatında al
        const formattedDate = date
          .toLocaleDateString("tr-TR", {
            day: "2-digit",
            month: "2-digit",
            year: "numeric",
          })
          .replace(/\./g, "/"); // Noktaları slash ile değiştir

        // Saati SS:DD formatında al
        const formattedTime = date.toLocaleTimeString("tr-TR", {
          hour: "2-digit",
          minute: "2-digit",
          hour12: false, // 24 saat formatı için
        });

        return `${formattedDate} ${formattedTime}`; // İki formatı birleştir
      },
    },
  ];
</script>

<SmartTable items={penalties} {columns} />
