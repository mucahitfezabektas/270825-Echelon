<script lang="ts">
  import SmartTable from "@/windows/tabs/SmartTable.svelte";
  import tableDataLoader from "@/lib/TableDataLoader";
  import type { CrewInfo } from "@/lib/tableDataTypes";
  import { onMount } from "svelte";

  let crewInfo: CrewInfo[] = [];

  onMount(async () => {
    if (!tableDataLoader.isLoaded) {
      await tableDataLoader.loadAll();
    }
    crewInfo = tableDataLoader.crewInfo;
    console.log("Loaded CrewInfo for table (after transformation):", crewInfo); // Yüklenen veriyi konsola yazdırın
  });

  // Helper function for date/time formatting (GG/AA/YYYY HH:MM)
  const formatDateTime = (v: number | null): string => {
    if (v === null || v === 0) return "-"; // Null veya 0 değeri gelirse "-" döndür
    const date = new Date(v); // Unix timestamp (milisaniye) doğrudan kullanılır

    const formattedDate = date
      .toLocaleDateString("tr-TR", {
        day: "2-digit",
        month: "2-digit",
        year: "numeric",
      })
      .replace(/\./g, "/");

    const formattedTime = date.toLocaleTimeString("tr-TR", {
      hour: "2-digit",
      minute: "2-digit",
      hour12: false,
    });

    return `${formattedDate} ${formattedTime}`;
  };

  // Helper for boolean values
  const formatBoolean = (v: boolean | string | null): string => {
    if (v === null) return "-";
    // parseBooleanHelper TableDataLoader'da olduğu için burada string kontrolü artık gerekmeyebilir
    // Ama backend hala string gönderebiliyorsa bu kontrolü tutmak güvenli olur.
    if (typeof v === "string") {
      const lowerCaseValue = v.toLowerCase().trim();
      if (
        lowerCaseValue === "y" ||
        lowerCaseValue === "true" ||
        lowerCaseValue === "calisiyor" ||
        lowerCaseValue === "gecerli"
      ) {
        return "✅ Evet";
      } else if (
        lowerCaseValue === "n" ||
        lowerCaseValue === "false" ||
        lowerCaseValue === "calismiyor" ||
        lowerCaseValue === "gecerli degil"
      ) {
        return "❌ Hayır";
      }
      return String(v); // Tanınmayan stringler için orijinal değeri döndür
    }
    return v ? "✅ Evet" : "❌ Hayır"; // Direct boolean handling
  };

  // Null veya boş stringleri "-" olarak formatlayan genel yardımcı
  const formatStringOrNull = (v: string | number | null): string => {
    if (v === null || v === "" || v === undefined) {
      return "-";
    }
    return String(v);
  };

  const columns = [
    { key: "person_id", label: "ID", format: formatStringOrNull },
    { key: "person_name", label: "Ad", format: formatStringOrNull },
    { key: "person_surname", label: "Soyad", format: formatStringOrNull },
    { key: "gender", label: "Cinsiyet", format: formatStringOrNull },
    { key: "tabiiyet", label: "Tabiiyet", format: formatStringOrNull },
    { key: "base_filo", label: "Base Filo", format: formatStringOrNull },
    {
      key: "dogum_tarihi",
      label: "Doğum Tarihi",
      format: formatDateTime,
    },
    {
      key: "base_location",
      label: "Base Lokasyon",
      format: formatStringOrNull,
    },
    { key: "ucucu_tipi", label: "Uçucu Tipi", format: formatStringOrNull },
    { key: "oml", label: "OML", format: formatStringOrNull },
    { key: "seniority", label: "Seniority", format: formatStringOrNull }, // ⭐ formatStringOrNull kullanılır
    {
      key: "rank_change_date",
      label: "Rank Değişim Tarihi",
      format: formatDateTime,
    },
    { key: "rank", label: "Rank", format: formatStringOrNull },
    {
      key: "agreement_type",
      label: "Anlaşma Tipi",
      format: formatStringOrNull,
    },
    {
      key: "agreement_type_explanation",
      label: "Anlaşma Açıklaması",
      format: formatStringOrNull,
    },
    {
      key: "job_start_date",
      label: "İşe Başlama Tarihi",
      format: formatDateTime,
    },
    {
      key: "job_end_date",
      label: "İşten Ayrılma Tarihi",
      format: formatDateTime,
    },
    {
      key: "marriage_date",
      label: "Evlilik Tarihi",
      format: formatDateTime,
    },
    { key: "ucucu_sinifi", label: "Uçucu Sınıfı", format: formatStringOrNull },
    {
      key: "ucucu_sinifi_last_valid",
      label: "Son Sınıf Geçerliliği",
      format: formatStringOrNull,
    },
    {
      key: "ucucu_alt_tipi",
      label: "Uçucu Alt Tipi",
      format: formatStringOrNull,
    },
    {
      key: "person_thy_calisiyor_mu",
      label: "THY Çalışıyor mu?",
      format: formatBoolean,
    },
    { key: "birthplace", label: "Doğum Yeri", format: formatStringOrNull },
    { key: "period_info", label: "Dönem Bilgisi", format: formatStringOrNull },
    {
      key: "service_use_home_pickup",
      label: "Home Pickup Kullanımı?",
      format: formatBoolean,
    },
    {
      key: "service_use_saw",
      label: "SAW Kullanımı?",
      format: formatBoolean,
    },
    {
      key: "bridge_use",
      label: "Köprü Kullanımı?",
      format: formatBoolean,
    },
  ];
</script>

<SmartTable items={crewInfo} {columns} />
