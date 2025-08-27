<script lang="ts">
  import { onMount } from "svelte";
  import tableDataLoader from "@/lib/TableDataLoader";
  import type { Penalty } from "@/lib/tableDataTypes"; // Penalty arayüzünü import edin
  import { debounce } from "@/lib/constants/utils/utils";

  const applyFilterAndSortDebounced = debounce(applyFilterAndSort, 300);

  let penaltyList: Penalty[] = [];
  let filteredPenaltyList: Penalty[] = [];
  let isLoadingInitialData = true;
  let isProcessingTable = false;
  let search = "";
  // penalty_code'u varsayılan sıralama anahtarı olarak belirledik
  let sortKey: keyof Penalty = "unique_id";
  let sortAsc = true;

  // Inline Düzenleme Durumları
  let editingCell: { rowId: string; key: keyof Penalty } | null = null;
  let editedValue: string = "";

  onMount(async () => {
    if (!tableDataLoader.isLoaded) {
      await tableDataLoader.loadAll();
    }

    penaltyList = tableDataLoader.penalties; // TableDataLoader'dan penaltyList verilerini atayın
    applyFilterAndSort();
    isLoadingInitialData = false;
  });

  function applyFilterAndSort() {
    isProcessingTable = true;
    setTimeout(() => {
      let tempPenalties = penaltyList.filter((p) =>
        [
          // Aranabilir alanları Penalty arayüzüne göre güncelleyin
          p.person_id,
          p.person_surname,
          p.person_name,
          p.ucucu_sinifi,
          p.base_filo,
          p.penalty_code,
          p.penalty_code_explanation,
          // Tarih alanlarını string'e çevirerek arayabilirsiniz
          String(p.penalty_start_date),
          String(p.penalty_end_date),
        ]
          .filter(Boolean) // null veya undefined değerleri filtrele
          .some((val) =>
            String(val).toLowerCase().includes(search.toLowerCase())
          )
      );

      tempPenalties.sort((a, b) => {
        const valA = (a[sortKey] ?? "").toString();
        const valB = (b[sortKey] ?? "").toString();
        // Tarih ve sayısal alanlar için özel sıralama gerekebilir, şu an string olarak karşılaştırır.
        return sortAsc ? valA.localeCompare(valB) : valB.localeCompare(valA);
      });

      filteredPenaltyList = tempPenalties;
      isProcessingTable = false;
    }, 0);
  }

  function toggleSort(key: keyof Penalty) {
    if (sortKey === key) {
      sortAsc = !sortAsc;
    } else {
      sortKey = key;
      sortAsc = true;
    }
    applyFilterAndSort();
  }

  // Tarihleri okunabilir string formatına çevirmek için yardımcı fonksiyon
  function formatDate(timestamp: number | null): string {
    if (timestamp === null || timestamp === 0) return "";
    const date = new Date(timestamp); // timestamp'in milisaniye cinsinden olduğunu varsayıyoruz
    return date.toLocaleDateString("tr-TR"); // Türkiye formatı
  }

  // --- Veri Dışa Aktarma Fonksiyonları ---

  /**
   * Veriyi CSV formatında dışa aktarır.
   * @param data Dışa aktarılacak veri dizisi.
   * @param filename Dosya adı.
   */
  function exportToCsv(data: Penalty[], filename: string) {
    if (data.length === 0) {
      alert("Dışa aktarılacak veri bulunamadı.");
      return;
    }

    // CSV başlığını Penalty arayüzüne göre oluşturun
    const header = Object.keys(data[0]).join(",");
    const rows = data.map((row) =>
      Object.values(row)
        .map((val) => {
          // Tarih alanlarını dışa aktarırken formatlayın
          if (
            typeof val === "number" &&
            (val === row.penalty_start_date || val === row.penalty_end_date)
          ) {
            val = formatDate(val);
          }
          // CSV uyumluluğu için metinleri tırnak içine al ve çift tırnakları kaçır
          const processedVal = String(val ?? "").replace(/"/g, '""');
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
  function exportToJson(data: Penalty[], filename: string) {
    if (data.length === 0) {
      alert("Dışa aktarılacak veri bulunamadı.");
      return;
    }

    const jsonContent = JSON.stringify(data, null, 2); // JSON'u güzelce yazdır
    const blob = new Blob([jsonContent], { type: "application/json" });
    const url = URL.createObjectURL(blob);
    const link = document.createElement("a");
    link.href = url;
    link.setAttribute("download", `${filename}.json`);
    document.body.appendChild(link);
    link.click();
    document.body.removeChild(link);
    URL.revokeObjectURL(url); // Nesne URL'ini temizle
  }

  // --- Satır İçi Düzenleme Fonksiyonları ---

  /**
   * Hücre düzenleme modunu başlatır.
   * @param penalty Sıradaki Penalty nesnesi.
   * @param key Düzenlenecek özelliğin anahtarı.
   */
  function startEditing(penalty: Penalty, key: keyof Penalty) {
    // Düzenlenebilir alanları burada belirtin
    // person_id muhtemelen backend tarafından yönetildiğinden düzenlenebilir yapılmamalıdır.
    if (
      key === "person_surname" ||
      key === "person_name" ||
      key === "ucucu_sinifi" ||
      key === "base_filo" ||
      key === "penalty_code" ||
      key === "penalty_code_explanation"
    ) {
      // rowId olarak benzersiz bir anahtar sağlamak için person_id kullanıyoruz.
      // Ancak Svelte'in 'each' bloğu için unique key ihtiyacını karşıladık.
      editingCell = { rowId: penalty.person_id, key };
      editedValue = (penalty[key] ?? "").toString();
    } else if (key === "penalty_start_date" || key === "penalty_end_date") {
      editingCell = { rowId: penalty.person_id, key };
      const dateVal = penalty[key] ? new Date(penalty[key] as number) : null;
      editedValue = dateVal ? dateVal.toISOString().split("T")[0] : ""; // 'YYYY-MM-DD'
    }
  }

  /**
   * Düzenlemeyi kaydeder ve hücreyi günceller.
   * Şu anlık sadece UI'da günceller, backend'e göndermez.
   */
  function saveEditing() {
    if (!editingCell) return;

    // Orijinal penaltyList dizisindeki ilgili cezayı bulmak için person_id kullanıyoruz.
    // Eğer birden fazla aynı person_id varsa, findIndex sadece ilkini bulup günceller.
    // Backend'de person_id'lerin benzersiz olduğundan emin olmalısınız.
    const originalPenaltyIndex = penaltyList.findIndex(
      (p) => p.person_id === editingCell!.rowId
    );

    if (originalPenaltyIndex !== -1) {
      const updatedPenaltyList = [...penaltyList];
      let valueToSave: string | number | null = editedValue;

      // Eğer düzenlenen alan bir tarih alanı ise, değeri timestamp'e çevir
      if (
        editingCell.key === "penalty_start_date" ||
        editingCell.key === "penalty_end_date"
      ) {
        const date = new Date(editedValue);
        valueToSave = isNaN(date.getTime()) ? null : date.getTime();
      } else if (editedValue === "") {
        // Boş bırakılan string alanlar için null kaydet
        valueToSave = null;
      }

      updatedPenaltyList[originalPenaltyIndex] = {
        ...updatedPenaltyList[originalPenaltyIndex],
        [editingCell.key]: valueToSave,
      };
      penaltyList = updatedPenaltyList; // Reaktiviteyi tetiklemek için yeniden ata

      // Filtrelenmiş görünümde değişikliği yansıtmak için filtreyi ve sıralamayı yeniden uygula
      applyFilterAndSort();
    }

    // Düzenleme durumunu sıfırla
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
  function isEditing(penalty: Penalty, key: keyof Penalty): boolean {
    return (
      editingCell?.rowId === penalty.person_id && editingCell?.key === key
    );
  }
</script>

<div>
  <div class="header">Ceza Bilgileri</div>
  <input
    placeholder="Ara..."
    bind:value={search}
    on:input={applyFilterAndSortDebounced}
    class="search"
  />

  <div class="controls">
    <button on:click={() => exportToCsv(filteredPenaltyList, "ceza-bilgileri")}>
      CSV Aktar
    </button>
    <button
      on:click={() => exportToJson(filteredPenaltyList, "ceza-bilgileri")}
    >
      JSON Aktar
    </button>
  </div>

  <div class="info">
    {#if isLoadingInitialData}
      Veriler yükleniyor...
    {:else}
      Toplam {filteredPenaltyList.length} kayıt
    {/if}
  </div>

  <div class="table-container">
    {#if isLoadingInitialData}
      <div class="loader-container">
        <div class="spinner"></div>
        <p>Veriler yükleniyor, lütfen bekleyiniz...</p>
      </div>
    {:else if filteredPenaltyList.length === 0 && !isProcessingTable}
      <div class="empty">Sonuç bulunamadı.</div>
    {:else}
      <table>
        <thead>
          <tr>
            <th on:click={() => toggleSort("person_id")}>
              Ceza ID {#if sortKey === "person_id"}<span
                  >{sortAsc ? "▲" : "▼"}</span
                >{/if}
            </th>
            <th on:click={() => toggleSort("person_id")}>
              Personel ID {#if sortKey === "person_id"}<span
                  >{sortAsc ? "▲" : "▼"}</span
                >{/if}
            </th>
            <th on:click={() => toggleSort("person_surname")}>
              Soyadı {#if sortKey === "person_surname"}<span
                  >{sortAsc ? "▲" : "▼"}</span
                >{/if}
            </th>
            <th on:click={() => toggleSort("person_name")}>
              Adı {#if sortKey === "person_name"}<span
                  >{sortAsc ? "▲" : "▼"}</span
                >{/if}
            </th>
            <th on:click={() => toggleSort("ucucu_sinifi")}>
              Uçucu Sınıfı {#if sortKey === "ucucu_sinifi"}<span
                  >{sortAsc ? "▲" : "▼"}</span
                >{/if}
            </th>
            <th on:click={() => toggleSort("base_filo")}>
              Base Filo {#if sortKey === "base_filo"}<span
                  >{sortAsc ? "▲" : "▼"}</span
                >{/if}
            </th>
            <th on:click={() => toggleSort("penalty_code")}>
              Ceza Kodu {#if sortKey === "penalty_code"}<span
                  >{sortAsc ? "▲" : "▼"}</span
                >{/if}
            </th>
            <th on:click={() => toggleSort("penalty_code_explanation")}>
              Ceza Açıklaması {#if sortKey === "penalty_code_explanation"}<span
                  >{sortAsc ? "▲" : "▼"}</span
                >{/if}
            </th>
            <th on:click={() => toggleSort("penalty_start_date")}>
              Başlangıç Tarihi {#if sortKey === "penalty_start_date"}<span
                  >{sortAsc ? "▲" : "▼"}</span
                >{/if}
            </th>
            <th on:click={() => toggleSort("penalty_end_date")}>
              Bitiş Tarihi {#if sortKey === "penalty_end_date"}<span
                  >{sortAsc ? "▲" : "▼"}</span
                >{/if}
            </th>
          </tr>
        </thead>
        <tbody>
          {#each filteredPenaltyList as p, i (`${p.person_id}-${i}`)}
            <tr class={i % 2 === 0 ? "even" : "odd"}>
              <td>{p.person_id}</td>
              <td>{p.person_id}</td>
              <td
                class="editable-cell"
                on:dblclick={() => startEditing(p, "person_surname")}
              >
                {#if isEditing(p, "person_surname")}
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
                  {p.person_surname}
                {/if}
              </td>
              <td
                class="editable-cell"
                on:dblclick={() => startEditing(p, "person_name")}
              >
                {#if isEditing(p, "person_name")}
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
                  {p.person_name}
                {/if}
              </td>
              <td
                class="editable-cell"
                on:dblclick={() => startEditing(p, "ucucu_sinifi")}
              >
                {#if isEditing(p, "ucucu_sinifi")}
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
                  {p.ucucu_sinifi}
                {/if}
              </td>
              <td
                class="editable-cell"
                on:dblclick={() => startEditing(p, "base_filo")}
              >
                {#if isEditing(p, "base_filo")}
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
                  {p.base_filo}
                {/if}
              </td>
              <td
                class="editable-cell"
                on:dblclick={() => startEditing(p, "penalty_code")}
              >
                {#if isEditing(p, "penalty_code")}
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
                  {p.penalty_code}
                {/if}
              </td>
              <td
                class="editable-cell"
                on:dblclick={() => startEditing(p, "penalty_code_explanation")}
              >
                {#if isEditing(p, "penalty_code_explanation")}
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
                  {p.penalty_code_explanation}
                {/if}
              </td>
              <td
                class="editable-cell"
                on:dblclick={() => startEditing(p, "penalty_start_date")}
              >
                {#if isEditing(p, "penalty_start_date")}
                  <input
                    type="date"
                    bind:value={editedValue}
                    on:blur={saveEditing}
                    on:keydown={(e) => {
                      if (e.key === "Enter") saveEditing();
                      if (e.key === "Escape") cancelEditing();
                    }}
                  />
                {:else}
                  {formatDate(p.penalty_start_date)}
                {/if}
              </td>
              <td
                class="editable-cell"
                on:dblclick={() => startEditing(p, "penalty_end_date")}
              >
                {#if isEditing(p, "penalty_end_date")}
                  <input
                    type="date"
                    bind:value={editedValue}
                    on:blur={saveEditing}
                    on:keydown={(e) => {
                      if (e.key === "Enter") saveEditing();
                      if (e.key === "Escape") cancelEditing();
                    }}
                  />
                {:else}
                  {formatDate(p.penalty_end_date)}
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
  /* Ortak stiller */
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
    z-index: 2;
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

  /* Yükleyici Stillleri */
  .loader-container,
  .processing-overlay {
    position: absolute;
    top: 0;
    left: 0;
    right: 0;
    bottom: 0;
    background: rgba(255, 255, 255, 0.9);
    display: flex;
    flex-direction: column;
    align-items: center;
    justify-content: center;
    color: #333;
    z-index: 10;
    font-size: 14px;
    font-weight: bold;
    border-radius: 4px;
    gap: 10px;
  }

  /* Spinner */
  .spinner {
    border: 4px solid #f3f3f3;
    border-top: 4px solid #007bff;
    border-radius: 50%;
    width: 40px;
    height: 40px;
    animation: spin 1s linear infinite;
  }

  .spinner.small {
    width: 20px;
    height: 20px;
    border-width: 3px;
  }

  @keyframes spin {
    0% {
      transform: rotate(0deg);
    }
    100% {
      transform: rotate(360deg);
    }
  }

  /* Satır İçi Düzenleme Stilleri */
  .editable-cell {
    position: relative;
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
