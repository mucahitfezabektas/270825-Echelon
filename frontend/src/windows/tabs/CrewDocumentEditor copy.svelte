<script lang="ts">
  import { onMount } from "svelte";
  import tableDataLoader from "@/lib/TableDataLoader";
  import type { ActivityCode } from "@/lib/tableDataTypes";
  import { debounce } from "@/lib/constants/utils/utils";

  const applyFilterAndSortDebounced = debounce(applyFilterAndSort, 300);

  let activityCodes: ActivityCode[] = [];
  let filteredCodes: ActivityCode[] = [];
  let isLoadingInitialData = true;
  let isProcessingTable = false;
  let search = "";
  let sortKey: keyof ActivityCode = "unique_id";
  let sortAsc = true;

  // Inline Editing Durumları
  let editingCell: { rowId: string; key: keyof ActivityCode } | null = null;
  let editedValue: string = "";

  onMount(async () => {
    if (!tableDataLoader.isLoaded) {
      await tableDataLoader.loadAll();
    }

    activityCodes = tableDataLoader.activityCodes;
    applyFilterAndSort();
    isLoadingInitialData = false;
  });

  function applyFilterAndSort() {
    isProcessingTable = true;
    setTimeout(() => {
      let tempCodes = activityCodes.filter((a) =>
        [
          a.activity_code,
          a.activity_group_code,
          a.activity_code_explanation,
        ].some((val) => val.toLowerCase().includes(search.toLowerCase()))
      );

      tempCodes.sort((a, b) => {
        const valA = (a[sortKey] ?? "").toString();
        const valB = (b[sortKey] ?? "").toString();
        return sortAsc ? valA.localeCompare(valB) : valB.localeCompare(valA);
      });

      filteredCodes = tempCodes;
      isProcessingTable = false;
    }, 0);
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

  // --- Export Data Fonksiyonları ---

  /**
   * Veriyi CSV formatında dışa aktarır.
   * @param data Dışa aktarılacak veri dizisi.
   * @param filename Dosya adı.
   */
  function exportToCsv(data: ActivityCode[], filename: string) {
    if (data.length === 0) {
      alert("Dışa aktarılacak veri bulunamadı.");
      return;
    }

    const header = Object.keys(data[0]).join(",");
    const rows = data.map((row) =>
      Object.values(row)
        .map((val) => {
          // CSV uyumluluğu için metinleri tırnak içine al ve çift tırnakları kaçır
          const processedVal = String(val).replace(/"/g, '""');
          return `"${processedVal}"`;
        })
        .join(",")
    );

    const csvContent =
      "data:text/csv;charset=utf-8," + [header, ...rows].join("\n");
    const encodedUri = encodeURI(csvContent);
    const link = document.createElement("a");
    link.setAttribute("href", encodedUri);
    link.setAttribute("download", `${filename}.csv`);
    document.body.appendChild(link);
    link.click();
    document.body.removeChild(link);
  }

  /**
   * Veriyi JSON formatında dışa aktarır.
   * @param data Dışa aktarılacak veri dizisi.
   * @param filename Dosya adı.
   */
  function exportToJson(data: ActivityCode[], filename: string) {
    if (data.length === 0) {
      alert("Dışa aktarılacak veri bulunamadı.");
      return;
    }

    const jsonContent = JSON.stringify(data, null, 2); // Pretty print JSON
    const blob = new Blob([jsonContent], { type: "application/json" });
    const url = URL.createObjectURL(blob);
    const link = document.createElement("a");
    link.href = url;
    link.setAttribute("download", `${filename}.json`);
    document.body.appendChild(link);
    link.click();
    document.body.removeChild(link);
    URL.revokeObjectURL(url); // Clean up the object URL
  }

  // --- Inline Editing Fonksiyonları ---

  /**
   * Hücre düzenleme modunu başlatır.
   * @param code The ActivityCode object for the row.
   * @param key The key of the property to edit.
   */
  function startEditing(code: ActivityCode, key: keyof ActivityCode) {
    // Only allow editing for 'activity_code_explanation' for now
    if (key === "activity_code_explanation") {
      editingCell = { rowId: code.activity_code, key };
      editedValue = (code[key] ?? "").toString();
    }
  }

  /**
   * Düzenlemeyi kaydeder ve hücreyi günceller.
   * Şu anlık sadece UI'da günceller, backend'e göndermez.
   */
  function saveEditing() {
    if (!editingCell) return;

    // find the activity in the original activityCodes array
    const originalCodeIndex = activityCodes.findIndex(
      (a) => a.activity_code === editingCell!.rowId
    );

    if (originalCodeIndex !== -1) {
      // Create a shallow copy to ensure reactivity (Svelte needs mutation or new array)
      const updatedActivityCodes = [...activityCodes];
      updatedActivityCodes[originalCodeIndex] = {
        ...updatedActivityCodes[originalCodeIndex],
        [editingCell.key]: editedValue,
      };
      activityCodes = updatedActivityCodes; // Reassign to trigger reactivity

      // Re-apply filter and sort to reflect the change in the filtered view
      // This is crucial for the edited data to be visible in the current table state.
      applyFilterAndSort();
    }

    // Reset editing state
    editingCell = null;
    editedValue = "";
  }

  /**
   * Düzenlemeyi iptal eder.
   */
  function cancelEditing() {
    editingCell = null;
    editedValue = "";
  }

  /**
   * Bir hücrenin düzenleme modunda olup olmadığını kontrol eder.
   */
  function isEditing(code: ActivityCode, key: keyof ActivityCode): boolean {
    return (
      editingCell?.rowId === code.activity_code && editingCell?.key === key
    );
  }
</script>

<div>
  <div class="header">Aktivite Kodları</div>
  <input
    placeholder="Ara..."
    bind:value={search}
    on:input={applyFilterAndSortDebounced}
    class="search"
  />

  <div class="controls">
    <button on:click={() => exportToCsv(filteredCodes, "aktivite-kodlari")}>
      CSV Aktar
    </button>
    <button on:click={() => exportToJson(filteredCodes, "aktivite-kodlari")}>
      JSON Aktar
    </button>
  </div>

  <div class="info">
    {#if isLoadingInitialData}
      Veriler yükleniyor...
    {:else}
      Toplam {filteredCodes.length} kayıt
    {/if}
  </div>

  <div class="table-container">
    {#if isLoadingInitialData}
      <div class="loader-container">
        <div class="spinner"></div>
        <p>Veriler yükleniyor, lütfen bekleyiniz...</p>
      </div>
    {:else if filteredCodes.length === 0 && !isProcessingTable}
      <div class="empty">Sonuç bulunamadı.</div>
    {:else}
      <table>
        <thead>
          <tr>
            <th on:click={() => toggleSort("activity_code")}>
              Kod
              {#if sortKey === "activity_code"}
                <span>{sortAsc ? "▲" : "▼"}</span>
              {/if}
            </th>
            <th on:click={() => toggleSort("activity_group_code")}>
              Grup
              {#if sortKey === "activity_group_code"}
                <span>{sortAsc ? "▲" : "▼"}</span>
              {/if}
            </th>
            <th on:click={() => toggleSort("activity_code_explanation")}>
              Açıklama
              {#if sortKey === "activity_code_explanation"}
                <span>{sortAsc ? "▲" : "▼"}</span>
              {/if}
            </th>
          </tr>
        </thead>
        <tbody>
          {#each filteredCodes as a, i (a.activity_code)}
            <tr class={i % 2 === 0 ? "even" : "odd"}>
              <td>{a.activity_code}</td>
              <td>{a.activity_group_code}</td>
              <td
                class="editable-cell"
                on:dblclick={() => {
                  editingCell = {
                    rowId: a.activity_code,
                    key: "activity_code_explanation",
                  };
                  editedValue = (a.activity_code_explanation ?? "").toString();
                }}
              >
                {#if isEditing(a, "activity_code_explanation")}
                  <input
                    type="text"
                    bind:value={editedValue}
                    on:blur={saveEditing}
                    on:keydown={(e) => {
                      if (e.key === "Enter") saveEditing();
                      if (e.key === "Escape") cancelEditing();
                    }}
                  />
                {:else}
                  {a.activity_code_explanation}
                {/if}
              </td>
            </tr>
          {/each}
        </tbody>
      </table>
      {#if isProcessingTable}
        <div class="processing-overlay">
          <div class="spinner small"></div>
          <p>Veriler filtreleniyor/sıralanıyor...</p>
        </div>
      {/if}
    {/if}
  </div>
</div>

<style>
  .header {
    font-weight: bold;
    font-size: 16px;
    margin-bottom: 0.5rem;
  }

  .search {
    width: 100%;
    padding: 6px;
    font-size: 13px;
    margin-bottom: 0.5rem;
    box-sizing: border-box;
    border: 1px solid #ccc;
    border-radius: 4px;
  }

  .controls {
    margin-bottom: 0.5rem;
    display: flex;
    gap: 10px;
  }

  .controls button {
    padding: 8px 15px;
    background-color: #007bff;
    color: white;
    border: none;
    border-radius: 4px;
    cursor: pointer;
    font-size: 13px;
    transition: background-color 0.2s;
  }

  .controls button:hover {
    background-color: #0056b3;
  }

  .info {
    font-size: 12px;
    margin-bottom: 0.5rem;
    color: #666;
  }

  .table-container {
    height: calc(100vh - 290px);
    overflow: auto;
    border: 1px solid #ccc;
    background: white;
    font-size: 12px;
    border-radius: 4px;
    position: relative;
  }

  table {
    width: 100%;
    border-collapse: collapse;
  }

  th,
  td {
    padding: 8px 12px;
    border: 1px solid #eee;
    text-align: left;
  }

  th {
    background: #e9e9e9;
    position: sticky;
    top: 0;
    cursor: pointer;
    white-space: nowrap;
    user-select: none;
    z-index: 2; /* Sıralama sorununu çözmek için z-index artırıldı */
  }

  th span {
    margin-left: 5px;
    font-size: 0.8em;
  }

  tr.even {
    background-color: #fdfdfd;
  }

  tr.odd {
    background-color: #f6f6f6;
  }

  .empty {
    padding: 1rem;
    color: #999;
    text-align: center;
  }

  /* Yeni Loader Stillleri */
  .loader-container,
  .processing-overlay {
    position: absolute;
    top: 0;
    left: 0;
    right: 0;
    bottom: 0;
    background: rgba(255, 255, 255, 0.9); /* Daha opak bir arka plan */
    display: flex;
    flex-direction: column; /* İçeriği dikey sırala */
    align-items: center;
    justify-content: center;
    color: #333;
    z-index: 10;
    font-size: 14px; /* Metin boyutu artırıldı */
    font-weight: bold;
    border-radius: 4px;
    gap: 10px; /* Spinner ve metin arasına boşluk */
  }

  /* Spinner */
  .spinner {
    border: 4px solid #f3f3f3; /* Light grey */
    border-top: 4px solid #007bff; /* Blue */
    border-radius: 50%;
    width: 40px; /* Büyük spinner */
    height: 40px;
    animation: spin 1s linear infinite;
  }

  .spinner.small {
    width: 20px; /* Küçük spinner */
    height: 20px;
    border-width: 3px; /* Kalınlığı azaltıldı */
  }

  @keyframes spin {
    0% {
      transform: rotate(0deg);
    }
    100% {
      transform: rotate(360deg);
    }
  }

  /* Inline Editing Styles */
  .editable-cell {
    position: relative; /* Ensure input takes full width of TD */
  }

  .editable-cell input {
    width: 100%;
    padding: 6px;
    box-sizing: border-box;
    border: 1px solid #007bff;
    border-radius: 3px;
    font-size: 12px;
  }
</style>
