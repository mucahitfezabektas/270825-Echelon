<script lang="ts">
  import SmartTable from "@/windows/tabs/SmartTable.svelte";
  import tableDataLoader from "@/lib/TableDataLoader";
  import type { CrewDocument } from "@/lib/tableDataTypes";
  import { onMount } from "svelte";

  let documents: CrewDocument[] = [];

  onMount(async () => {
    if (!tableDataLoader.isLoaded) {
      await tableDataLoader.loadAll();
    }
    documents = tableDataLoader.crewDocuments;
  });

  // Helper function for date/time formatting (copy from PenaltyTable.svelte)
  const formatDateTime = (v: number | null): string => {
    if (v === null || v === 0) return ""; // Null veya 0 değeri gelirse boş döndür
    const date = new Date(v); // Unix timestamp (milisaniye) doğrudan kullanılır

    const formattedDate = date
      .toLocaleDateString("tr-TR", {
        day: "2-digit",
        month: "2-digit",
        year: "numeric",
      })
      .replace(/\./g, "/"); // Noktaları slash ile değiştir

    const formattedTime = date.toLocaleTimeString("tr-TR", {
      hour: "2-digit",
      minute: "2-digit",
      hour12: false, // 24 saat formatı için
    });

    return `${formattedDate} ${formattedTime}`;
  };

  // Helper function for boolean formatting
  const formatBoolean = (v: boolean | string | null): string => {
    if (v === null) return "";
    if (typeof v === "string") {
      v =
        v.toLowerCase() === "true" ||
        v.toLowerCase() === "calisiyor" ||
        v.toLowerCase() === "gecerli";
    }
    return v ? "✅ Evet" : "❌ Hayır";
  };

  const columns = [
    { key: "person_id", label: "ID" },
    { key: "person_name", label: "Ad" },
    { key: "person_surname", label: "Soyad" },
    { key: "citizenship_number", label: "TCKN" },
    { key: "person_type", label: "Tip" },
    { key: "ucucu_alt_tipi", label: "Alt Tip" },
    { key: "ucucu_sinifi", label: "Sınıf" },
    { key: "base_filo", label: "Filo" },
    { key: "dokuman_alt_tipi", label: "Belge Tipi" },
    {
      key: "gecerlilik_baslangic_tarihi",
      label: "Geçerlilik Başlangıç",
      format: formatDateTime, // Helper'ı kullan
    },
    {
      key: "gecerlilik_bitis_tarihi",
      label: "Geçerlilik Bitiş",
      format: formatDateTime, // Helper'ı kullan
    },
    {
      key: "document_no",
      label: "Belge No",
      format: (v: string | null) => v || "", // Null ise boş string göster
    },
    {
      key: "dokumani_veren",
      label: "Veren",
      format: (v: string | null) => v || "", // Null ise boş string göster
    },
    {
      key: "end_date_leave_job",
      label: "İşten Ayrılma",
      format: formatDateTime, // Tarih olduğu için helper'ı kullan
    },
    {
      key: "personel_thy_calisiyor_mu",
      label: "Çalışıyor mu?",
      format: formatBoolean, // Helper'ı kullan
    },
    {
      key: "dokuman_gecerli_mi",
      label: "Geçerli mi?",
      format: formatBoolean, // Helper'ı kullan
    },
    { key: "agreement_type", label: "Anlaşma" },
  ];
</script>

<SmartTable items={documents} {columns} />
