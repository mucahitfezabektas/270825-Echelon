<script lang="ts">
  import { createEventDispatcher, onMount, onDestroy } from "svelte";
  import { get } from "svelte/store";
  import { filterQueryStore, filterActions } from "@/stores/filterQueryStore";
  import type { FilterRow, SavedFilter } from "@/lib/types";
  import { tick } from "svelte";

  const FIELD_OPTIONS: { value: string; label: string; placeholder: string }[] =
    [
      { value: "person_id", label: "Person ID (c)", placeholder: "ÖR: 109403" },
      { value: "surname", label: "Surname (s)", placeholder: "ÖR: YILMAZ" },
      {
        value: "activity_code",
        label: "Activity Code (a)",
        placeholder: "ÖR: FLT",
      },
      { value: "class", label: "Class (cl)", placeholder: "ÖR: A" },
      {
        value: "departure_port",
        label: "Departure Port (dp)",
        placeholder: "ÖR: IST",
      },
      {
        value: "arrival_port",
        label: "Arrival Port (ap)",
        placeholder: "ÖR: AYT",
      },
      {
        value: "date",
        label: "Date (d)",
        placeholder: "ÖR: 2024-06-15 veya 2024-06-15 TO 2024-06-30",
      },
      { value: "trip_id", label: "Trip ID (t)", placeholder: "ÖR: TRIP_12345" },
      {
        value: "plane_tail_name",
        label: "Tail Name (pt)",
        placeholder: "ÖR: TC-JHL",
      },
      {
        value: "plane_cms_type",
        label: "CMS Type (pc)",
        placeholder: "ÖR: A320",
      },
      {
        value: "group_code",
        label: "Group Code (gc)",
        placeholder: "ÖR: GRP01",
      },
      {
        value: "flight_position",
        label: "Flight Pos (fp)",
        placeholder: "ÖR: CPT",
      },
      {
        value: "flight_no",
        label: "Flight No (fn)",
        placeholder: "ÖR: TK1234",
      },
      {
        value: "agreement_type",
        label: "Agreement Type (at)",
        placeholder: "ÖR: FULL",
      },
      {
        value: "ucus_id",
        label: "Flight ID (fi)",
        placeholder: "ÖR: TK001-IST-20240710120000",
      },
    ];

  const dispatch = createEventDispatcher();

  /* --- store aboneliği --- */
  let filters: SavedFilter[] = [];
  const unsubscribe = filterQueryStore.subscribe((v) => (filters = v));
  onDestroy(unsubscribe);

  /* --- local state --- */
  let workingRows: FilterRow[] = [];
  let selectedFilterId: string | null = null;
  let editingFilterNameId: string | null = null;

  let inputRef: HTMLInputElement | null = null;

  // FIXED: Provide a default value and handle undefined correctly
  let currentFilterLogic: "AND" | "OR" = "AND";

  /* --- lifecycle --- */
  onMount(() => {
    if (filters.length === 0) {
      const id = filterActions.addFilter("Yeni Filtre 1");
      selectedFilterId = id;
    } else {
      selectedFilterId = filters[0].id;
    }
    loadSelectedFilterRows();
  });

  $: if (editingFilterNameId && inputRef) {
    tick().then(() => {
      inputRef?.focus();
      inputRef?.select(); // tüm metni seçmek için
    });
  }

  /* --- reactive: filtre değişince satırları yükle & eskisini kaydet --- */
  $: {
    const current = filters.find((f) => f.id === selectedFilterId);
    if (current) {
      workingRows = current.rows.map((r) => ({ ...r }));
      // FIXED: Use nullish coalescing operator to provide a default if current.logic is undefined
      currentFilterLogic = current.logic ?? "AND";
    } else {
      workingRows = [];
      currentFilterLogic = "AND"; // Reset to default if no filter is selected
    }
  }

  /* --- helpers --- */
  function loadSelectedFilterRows() {
    const sel = filters.find((f) => f.id === selectedFilterId);
    workingRows = sel ? sel.rows.map((r) => ({ ...r })) : [];
    // FIXED: Use nullish coalescing operator here as well
    currentFilterLogic = sel?.logic ?? "AND";
  }

  function saveWorkingRowsToSelectedFilter() {
    if (selectedFilterId)
      filterActions.updateFilterRows(
        selectedFilterId,
        workingRows.map((r) => ({ ...r })) // <- her satırı klonla
      );
  }

  function addFilter() {
    saveWorkingRowsToSelectedFilter();
    selectedFilterId = filterActions.addFilter();
    editingFilterNameId = selectedFilterId;
  }

  function deleteFilter() {
    if (!selectedFilterId) return;
    filterActions.deleteFilter(selectedFilterId);
    selectedFilterId = filters[0]?.id ?? null; // Silindikten sonra ilk filtreye geç veya null yap
    editingFilterNameId = null;
  }

  function addRow() {
    workingRows = [
      ...workingRows,
      { field: FIELD_OPTIONS[0].value, operator: "=", value: "" },
    ];
    saveWorkingRowsToSelectedFilter();
  }

  function deleteRow(i: number) {
    workingRows = workingRows.filter((_, idx) => idx !== i);
    saveWorkingRowsToSelectedFilter();
  }

  function setLogic(val: "AND" | "OR") {
    if (!selectedFilterId) return;
    filterActions.updateFilterLogic(selectedFilterId, val);
    currentFilterLogic = val; // Update local reactive variable to reflect change
  }

  function handleFilterNameChange(e: Event, f: SavedFilter) {
    // This function is generally not needed if using bind:value directly on f.name
    // The blur and keydown handlers already manage saving changes to the store.
  }

  /* --- footer actions --- */
  const handleOk = () => {
    saveWorkingRowsToSelectedFilter();
    const currentFilter = filters.find((f) => f.id === selectedFilterId);
    if (currentFilter) {
      dispatch("ok", { savedFilter: currentFilter });
    } else {
      dispatch("ok", { savedFilter: null }); // Send null if no filter is selected
    }
  };

  const handleApply = () => {
    saveWorkingRowsToSelectedFilter();
    const currentFilter = filters.find((f) => f.id === selectedFilterId);
    if (currentFilter) {
      dispatch("apply", { savedFilter: currentFilter });
    } else {
      dispatch("apply", { savedFilter: null }); // Send null if no filter is selected
    }
  };

  const handleCancel = () => dispatch("cancel");
</script>

<div class="body">
  <div class="list-pane">
    <div class="toolbar">
      <button title="Yeni filtre ekle" on:click={addFilter}>
        <span class="icon">➕</span>
      </button>
      <button title="Seçili filtreyi sil" on:click={deleteFilter}>
        <span class="icon">🗑️</span>
      </button>
    </div>

    <ul class="filter-list" role="listbox">
      {#each filters as f (f.id)}
        <li
          role="option"
          aria-selected={f.id === selectedFilterId}
          class:selected={f.id === selectedFilterId}
          tabindex="0"
          on:click={() => {
            selectedFilterId = f.id;
            editingFilterNameId = null; // Exit editing mode when clicking another filter
          }}
          on:keydown={(e) => e.key === "Enter" && (selectedFilterId = f.id)}
          on:dblclick={() => (editingFilterNameId = f.id)}
        >
          {#if editingFilterNameId === f.id}
            <input
              type="text"
              bind:this={inputRef}
              bind:value={f.name}
              on:blur={() => {
                editingFilterNameId = null;
                filterActions.updateFilterName(f.id, f.name); // Save change to store
              }}
              on:keydown={(e) => {
                if (e.key === "Enter") {
                  editingFilterNameId = null;
                  filterActions.updateFilterName(f.id, f.name);
                }
              }}
            />
          {:else}
            {f.name}
          {/if}
        </li>
      {/each}
    </ul>
  </div>

  <div class="settings-pane">
    {#if selectedFilterId}
      <div class="rows-container">
        {#each workingRows as row, i (i)}
          <div class="row">
            <select
              class="field-select"
              bind:value={row.field}
              on:change={saveWorkingRowsToSelectedFilter}
            >
              {#each FIELD_OPTIONS as opt}
                <option value={opt.value}>{opt.label}</option>
              {/each}
            </select>

            <select
              class="op-select"
              bind:value={row.operator}
              on:change={saveWorkingRowsToSelectedFilter}
            >
              <option value="=">=</option>
              <option value="!=">≠</option>
              <option value=">">&gt;</option>
              <option value="<">&lt;</option>
              <option value="LIKE">LIKE</option>
            </select>
            <input
              type="text"
              placeholder={FIELD_OPTIONS.find((opt) => opt.value === row.field)
                ?.placeholder ?? "Değer"}
              bind:value={row.value}
              on:blur={saveWorkingRowsToSelectedFilter}
              on:keydown={(e) => {
                // Save on Enter key press as well
                if (e.key === "Enter") {
                  saveWorkingRowsToSelectedFilter();
                  (e.target as HTMLInputElement).blur(); // Remove focus
                }
              }}
            />

            <button
              class="row-del"
              title="Satırı Kaldır"
              on:click={() => deleteRow(i)}
            >
              <span class="icon">✕</span>
            </button>
          </div>
        {/each}
      </div>

      <div class="row-toolbar">
        <button class="add-condition-btn" on:click={addRow}>
          <span class="icon">➕</span> Koşul Ekle
        </button>
        <label class="logic-selector">
          Mantık:
          <select
            bind:value={currentFilterLogic}
            on:change={(e) => {
              const target = e.target as HTMLSelectElement;
              const val = target.value as "AND" | "OR";
              setLogic(val);
            }}
          >
            <option value="AND">VE (AND)</option>
            <option value="OR">VEYA (OR)</option>
          </select>
        </label>
      </div>
    {:else}
      <p class="no-filter-selected">
        <span class="icon">🤔</span> Lütfen sol panelden bir filtre seçin veya yeni
        bir filtre ekleyin.
      </p>
    {/if}
  </div>
</div>

<div class="footer">
  <button class="btn btn-primary" on:click={handleOk}>Uygula ve Kapat</button>
  <button class="btn btn-secondary" on:click={handleApply}>Uygula</button>
  <button class="btn btn-cancel" on:click={handleCancel}>İptal</button>
</div>

<style>
  /* --- Genel Renk Paleti ve Temel Stiller (Profesyonel ve Soğuk Tonlar) --- */
  :root {
    --primary-blue: #3498db;
    --dark-blue: #2980b9;
    --light-blue-bg: #eaf2f6; /* Açık mavi arka plan */
    --text-color-dark: #2c3e50; /* Koyu gri/mavi */
    --text-color-medium: #5b7c99; /* Orta gri/mavi */
    --border-light: #d1d8dc; /* Hafif kenarlık */
    --border-medium: #abbdc7; /* Orta kenarlık */
    --shadow-subtle: rgba(44, 62, 80, 0.1); /* Hafif gölge */
    --success-color: #28a745;
    --cancel-color: #dc3545;
  }

  /*
    WARNING FIXED:
    The 'body' selector is moved to a global CSS file
    (e.g., src/app.css or src/global.css) and imported at your app's root.
    Svelte scopes styles, so 'body' in a component's <style> tag won't apply globally.
  */
  /* body {
    font-family: 'Segoe UI', Tahoma, Geneva, Verdana, sans-serif;
    font-size: 14px;
    color: var(--text-color-dark);
  } */

  /* --- Ana Düzen --- */
  .body {
    display: flex;
    height: 100%; /* Parent'tan gelen yüksekliği tam kullan */
    min-height: 400px; /* Minimum yükseklik */
    border: 1px solid var(--border-medium);
    border-radius: 8px;
    overflow: hidden; /* İçerik taşmasını engelle */
    background-color: #ffffff; /* Beyaz iç arka plan */
    box-shadow: 0 4px 15px var(--shadow-subtle);
  }

  /* --- Sol Panel: Filtre Listesi --- */
  .list-pane {
    width: 220px; /* Daha geniş liste paneli */
    display: flex;
    flex-direction: column;
    border-right: 1px solid var(--border-light);
    background-color: #f8fbfd; /* Çok açık mavi/gri arka plan */
  }

  .toolbar {
    display: flex;
    gap: 8px; /* Butonlar arası boşluk */
    padding: 8px;
    background: #eef5f8; /* Hafif farklı toolbar rengi */
    border-bottom: 1px solid var(--border-light);
    justify-content: flex-start; /* Butonları sola hizala */
  }

  .toolbar button {
    display: flex;
    align-items: center;
    justify-content: center;
    width: 32px; /* Daha büyük butonlar */
    height: 32px;
    border: 1px solid var(--border-medium);
    background: #ffffff;
    border-radius: 6px; /* Hafif yuvarlak köşeler */
    cursor: pointer;
    font-size: 1.2em;
    color: var(--text-color-dark);
    transition: all 0.2s ease;
    box-shadow: 0 1px 2px rgba(0, 0, 0, 0.05);
  }

  .toolbar button:hover {
    background-color: var(--light-blue-bg);
    border-color: var(--primary-blue);
    color: var(--primary-blue);
  }

  .filter-list {
    flex: 1;
    margin: 0;
    padding: 0;
    list-style: none;
    overflow-y: auto;
    background-color: #fdfefe; /* Liste arka planı */
  }

  .filter-list li {
    padding: 10px 15px; /* Daha fazla padding */
    border-bottom: 1px solid #f6f9fb; /* Çok açık bir çizgi */
    display: flex;
    align-items: center;
    cursor: pointer;
    transition: background-color 0.2s ease;
  }

  .filter-list li:hover {
    background-color: #f0f5f8; /* Hafif hover efekti */
  }

  .filter-list li.selected {
    background: var(--light-blue-bg); /* Seçili filtre arka planı */
    font-weight: 600; /* Kalın font */
    color: var(--dark-blue); /* Koyu mavi metin */
    border-left: 4px solid var(--primary-blue); /* Solunda vurgu çizgisi */
    padding-left: 11px; /* Çizgi nedeniyle kayma */
  }

  .filter-list input[type="text"] {
    width: 100%;
    padding: 6px 8px;
    border: 1px solid var(--border-medium);
    border-radius: 4px;
    font-size: 0.95em;
    background-color: #ffffff;
    outline: none;
  }

  /* --- Sağ Panel: Ayarlar --- */
  .settings-pane {
    flex: 1;
    display: flex;
    flex-direction: column;
    padding: 15px; /* Artırılmış padding */
    gap: 15px; /* Öğeler arası boşluk */
    background-color: #ffffff;
  }

  .rows-container {
    flex: 1;
    overflow-y: auto;
    border: 1px solid var(--border-light);
    border-radius: 6px;
    padding: 15px; /* İç padding */
    background-color: #fefefe;
    box-shadow: inset 0 1px 3px rgba(0, 0, 0, 0.05); /* İç gölge */
  }

  .row {
    display: flex;
    align-items: center;
    gap: 10px; /* Daha fazla boşluk */
    margin-bottom: 12px; /* Satırlar arası dikey boşluk */
  }

  .row:last-child {
    margin-bottom: 0; /* Son satırın altında boşluk olmasın */
  }

  .row select,
  .row input[type="text"] {
    padding: 8px 10px; /* Daha fazla padding */
    border: 1px solid var(--border-medium);
    border-radius: 5px; /* Hafif yuvarlak köşeler */
    font-size: 0.95em;
    color: var(--text-color-dark);
    background-color: #ffffff;
    outline: none;
    transition:
      border-color 0.2s ease,
      box-shadow 0.2s ease;
  }

  .row select:focus,
  .row input[type="text"]:focus {
    border-color: var(--primary-blue);
    box-shadow: 0 0 0 2px rgba(52, 152, 219, 0.2); /* Odaklandığında mavi glow */
  }

  .row .field-select {
    flex: 2; /* Genişliği ayarla, oransal olarak */
    min-width: 150px;
    max-width: 200px;
  }

  .row .op-select {
    flex: 0 0 70px; /* Sabit genişlik */
    text-align: center;
  }

  .row input[type="text"] {
    flex: 3; /* Geri kalan alanı kapla */
    min-width: 100px;
  }

  .row-del {
    background: none;
    border: none;
    color: var(--cancel-color); /* Kırmızı çarpı */
    cursor: pointer;
    font-size: 1.3em; /* Daha büyük */
    padding: 5px; /* Tıklama alanı */
    transition: color 0.2s ease;
  }

  .row-del:hover {
    color: #a71d2a; /* Daha koyu kırmızı */
  }

  .row-toolbar {
    display: flex;
    justify-content: space-between;
    align-items: center;
    padding-top: 15px;
    border-top: 1px solid #f0f0f0; /* Hafif üst çizgi */
  }

  .add-condition-btn {
    display: flex;
    align-items: center;
    gap: 5px;
    background-color: #6c757d; /* Gri ton */
    color: white;
    border: none;
    padding: 8px 15px;
    border-radius: 6px;
    cursor: pointer;
    font-size: 0.95em;
    transition: background-color 0.2s ease;
  }

  .add-condition-btn:hover {
    background-color: #5a6268; /* Daha koyu gri */
  }

  .logic-selector {
    display: flex;
    align-items: center;
    gap: 8px;
    font-weight: 500;
    color: var(--text-color-medium);
  }

  .logic-selector select {
    padding: 6px 10px;
    border: 1px solid var(--border-medium);
    border-radius: 5px;
    font-size: 0.95em;
    background-color: #ffffff;
    color: var(--text-color-dark);
    outline: none;
    cursor: pointer;
  }

  .no-filter-selected {
    display: flex;
    flex-direction: column;
    align-items: center;
    justify-content: center;
    height: 100%;
    color: var(--text-color-medium);
    font-style: italic;
    text-align: center;
    gap: 10px;
  }

  .no-filter-selected .icon {
    font-size: 2.5em;
    color: var(--primary-blue);
  }

  /* --- Altbilgi --- */
  .footer {
    display: flex;
    justify-content: flex-end;
    gap: 10px; /* Butonlar arası boşluk */
    border-top: 1px solid var(--border-light);
    padding: 15px; /* Daha fazla padding */
    background: #f8fbfd; /* Footer arka planı */
    border-radius: 0 0 8px 8px; /* Alt köşeleri yuvarlak yap (eğer tüm pencere köşeliyse) */
  }

  .btn {
    padding: 10px 20px;
    border: none;
    border-radius: 6px;
    cursor: pointer;
    font-size: 1em;
    font-weight: 600;
    transition:
      background-color 0.2s ease,
      box-shadow 0.2s ease;
    box-shadow: 0 2px 5px rgba(0, 0, 0, 0.1);
  }

  .btn-primary {
    background-color: var(--primary-blue);
    color: white;
  }

  .btn-primary:hover {
    background-color: var(--dark-blue);
    box-shadow: 0 4px 10px rgba(0, 0, 0, 0.2);
  }

  .btn-secondary {
    background-color: #6c757d; /* Gri ton */
    color: white;
  }

  .btn-secondary:hover {
    background-color: #5a6268;
    box-shadow: 0 4px 10px rgba(0, 0, 0, 0.2);
  }

  .btn-cancel {
    background-color: #f8f8f8; /* Açık gri */
    color: var(--text-color-dark);
    border: 1px solid var(--border-medium);
  }

  .btn-cancel:hover {
    background-color: #e0e0e0;
    box-shadow: 0 4px 10px rgba(0, 0, 0, 0.15);
  }

  /* İkonlar için genel stil */
  .icon {
    line-height: 1; /* Satır yüksekliğini sıfırla */
  }
</style>
