<script lang="ts">
  import { onMount, onDestroy } from "svelte";
  import type { FlightItem, FilterRow, SavedFilter } from "@/lib/types";
  import { RowType } from "@/lib/types";

  // Import all three timeline components
  import RosterTimeline from "@/components/RosterTimeline.svelte";
  import TripTimeline from "@/components/TripTimeline.svelte";
  import RotationTimeline from "@/components/RotationTimeline.svelte";

  import {
    dragStore,
    resetDragState,
    type DragState,
  } from "@/stores/dragStore";
  import { hoverStore } from "@/stores/hoverStore";
  import type { HoverInfo } from "@/lib/types";
  import { filterQueryStore } from "@/stores/filterQueryStore";
  import { get } from "svelte/store";
  import { getZoomController } from "@/stores/zoomControlStore";
  import { highlightStore } from "@/stores/highlightStore";
  import CommandHelp from "@/components/CommandHelp.svelte";

  import {
    activeTimelines,
    updateTimeline, // Import updateTimeline fonksiyonu
    removeTimeline,
    addNewEmptyTimeline,
    toggleMinimizeTimeline,
    handleGlobalSearch, // Bu fonksiyon da kullanılacak
    resetAllTimelines,
    type TimelineEntry,
  } from "@/stores/timelineManager";

  import { showContextMenu } from "@/stores/contextMenuStore";
  import { ContextMenuType } from "@/lib/ContextMenuTypes";
  import {
    getFlightItemContextMenu,
    getTimelineHeaderContextMenu,
    getTimelineRowContextMenu,
    getTimelineEmptyContextMenu,
    getTimelineRowSidebarContextMenu,
  } from "@/lib/contextMenus/index";

  let showCommandHelp = false;

  let naturalQueryInputs: Record<string, string> = {};
  let highlightInputs: Record<string, string> = {};

  let sidebarMenuOpen = false;
  let sidebarMenuX = 0;
  let sidebarMenuY = 0;
  let sidebarMenuTimelineId: string | null = null;

  let searchTerm = "";

  let currentDragState: DragState;
  const unsubscribeDragStore = dragStore.subscribe((value) => {
    currentDragState = value;
  });

  let currentHoverInfo: HoverInfo;
  const unsubscribeHoverStore = hoverStore.subscribe((value) => {
    currentHoverInfo = value;
  });

  // selectedFiltersPerTimeline state'ini koruyun, bu filtre seçimi için UI'a bağlı
  let selectedFiltersPerTimeline: Record<string, string | null> = {};
  let allFilters: SavedFilter[] = get(filterQueryStore);

  let isResizing = false;
  let resizingIndex = -1;
  let startY = 0;
  let startHeights: number[] = [];

  const unsubscribeFilters = filterQueryStore.subscribe(
    (f) => (allFilters = f)
  );

  onMount(() => {
    if (get(activeTimelines).length === 0) {
      addNewEmptyTimeline("roster");
    }

    document.addEventListener("click", handleClickOutsideDropdown);
    window.addEventListener("mousemove", handleGlobalMouseMove);
    window.addEventListener("mouseup", handleGlobalMouseUp);
    window.addEventListener("keydown", handleGlobalKeyDown);
    window.addEventListener("keyup", handleGlobalKeyUp);
  });

  onDestroy(() => {
    document.removeEventListener("click", handleClickOutsideDropdown);
    window.removeEventListener("mousemove", handleGlobalMouseMove);
    window.removeEventListener("mouseup", handleGlobalMouseUp);
    window.removeEventListener("keydown", handleGlobalKeyDown);
    window.removeEventListener("keyup", handleGlobalKeyUp);
    unsubscribeDragStore();
    unsubscribeHoverStore();
    unsubscribeFilters();
    resetAllTimelines();
    console.log(
      "TimelineContainer: onDestroy tetiklendi, zaman çizelgeleri sıfırlandı."
    );
  });

  // --- Tüm Context Menu Handler Fonksiyonları (Aynı kalmalı) ---
  function handleFlightItemContextMenu(detail: {
    flight: FlightItem;
    pageX: number;
    pageY: number;
    timelineId: string;
  }) {
    console.log("Timeline: FlightItem context menu alındı.");
    const items = getFlightItemContextMenu(detail.flight);
    showContextMenu(detail.pageX, detail.pageY, items, {
      type: ContextMenuType.TIMELINE_ITEM,
      flight: detail.flight,
      timelineId: detail.timelineId,
    });
  }

  function handleTimelineRowSidebarContextMenu({
    personId,
    rowType,
    pageX,
    pageY,
    timelineId,
  }: {
    personId: string;
    rowType: RowType;
    pageX: number;
    pageY: number;
    timelineId: string;
  }) {
    showContextMenu(
      pageX,
      pageY,
      getTimelineRowSidebarContextMenu({
        rowId: personId,
        rowType: rowType as RowType,
        timelineId,
      }),
      { rowId: personId, rowType }
    );
  }

  // --- Diğer mevcut fonksiyonlar (Yukarıdakiyle aynı kalmalı) ---

  function handleHighlightJump(timelineId: string) {
    const raw = highlightInputs[timelineId]?.trim() ?? "";
    const keys = raw.split(/[\s,;]+/).filter(Boolean);

    if (keys.length === 0) {
      highlightStore.set({ timelineId, keys: [] });
      return;
    }

    highlightStore.set({ timelineId, keys });
  }

  function handleZoom(timelineId: string, dir: "in" | "out") {
    const ctrl = getZoomController(timelineId);
    if (!ctrl) return;
    dir === "in" ? ctrl.zoomIn() : ctrl.zoomOut();
  }

  function handleCommandInsert(cmd: string) {
    searchTerm = cmd;
    showCommandHelp = false;
  }

  // ÖNEMLİ DEĞİŞİKLİK: handleNaturalQuery
  async function handleNaturalQuery(timelineId: string) {
    const input = naturalQueryInputs[timelineId]?.trim();
    let filters: Record<string, string> = {};

    if (input) {
      if (!/^\d+$/.test(input)) {
        alert(
          "Geçersiz personel ID. Lütfen sadece rakamlardan oluşan bir sicil numarası girin (Örn: 109403)."
        );
        console.warn("Geçersiz personel ID formatı.");
        return;
      }
      filters = { person_id: input };
    }

    // Artık updateTimeline'ı doğrudan çağırıyoruz ve RosterTimeline store'dan dinleyecek
    await updateTimeline(timelineId, filters, "roster");
  }

  function startResizing(index: number, event: MouseEvent) {
    isResizing = true;
    resizingIndex = index;
    startY = event.clientY;
    startHeights = get(activeTimelines).map((t) => t.heightRatio);

    window.addEventListener("mousemove", handleResizing);
    window.addEventListener("mouseup", stopResizing);
  }

  function handleResizing(event: MouseEvent) {
    if (!isResizing) return;

    const deltaY = event.clientY - startY;
    const container = document.querySelector(
      ".timelines-container"
    ) as HTMLElement;
    if (!container) return;

    const totalHeight = container.clientHeight;
    const deltaRatio = deltaY / totalHeight;

    const updated = [...get(activeTimelines)];
    const upper = updated[resizingIndex];
    const lower = updated[resizingIndex + 1];

    if (!upper || !lower) {
      console.warn("Resizing failed: Upper or lower timeline not found.");
      stopResizing();
      return;
    }

    let newUpper = Math.max(0.1, startHeights[resizingIndex] + deltaRatio);
    let newLower = Math.max(0.1, startHeights[resizingIndex + 1] - deltaRatio);

    const sum = newUpper + newLower;
    upper.heightRatio =
      (newUpper / sum) *
      (startHeights[resizingIndex] + startHeights[resizingIndex + 1]);
    lower.heightRatio =
      (newLower / sum) *
      (startHeights[resizingIndex] + startHeights[resizingIndex + 1]);

    activeTimelines.set(updated);
  }

  function stopResizing() {
    isResizing = false;
    window.removeEventListener("mousemove", handleResizing);
    window.removeEventListener("mouseup", stopResizing);
  }

  function formatDateTime(epoch: number | undefined): string {
    if (!epoch) return "-";
    return new Date(epoch).toLocaleString("tr-TR", {
      year: "numeric",
      month: "2-digit",
      day: "2-digit",
      hour: "2-digit",
      minute: "2-digit",
      hour12: false,
    });
  }

  function toggleMinimize(timelineId: string) {
    toggleMinimizeTimeline(timelineId);
  }

  function handleRemoveTimeline(idToRemove: string) {
    removeTimeline(idToRemove);
  }

  let showDropdown: Record<string, boolean> = {};
  let activeDropdownTimelineId: string | null = null;

  function toggleDropdown(timelineId: string, event: MouseEvent) {
    event.stopPropagation();
    if (activeDropdownTimelineId === timelineId) {
      activeDropdownTimelineId = null;
      if (showDropdown[timelineId]) {
        showDropdown[timelineId] = false;
      }
    } else {
      Object.keys(showDropdown).forEach((id) => (showDropdown[id] = false));
      activeDropdownTimelineId = timelineId;
      showDropdown[timelineId] = true;
    }
    showDropdown = showDropdown;
  }

  const FIELD_ALIASES: Record<string, string> = {
    person_id: "c",
    surname: "s",
    activity_code: "a",
    class: "cl",
    departure_port: "dp",
    arrival_port: "ap",
    date: "d",
    trip_id: "t",
    plane_tail_name: "pt",
    plane_cms_type: "pc",
    group_code: "gc",
    flight_position: "fp",
    flight_no: "fn",
    agreement_type: "at",
    flight_id: "fi",
  };

  function buildFiltersFromSavedFilter(
    filterRows: FilterRow[]
  ): Record<string, string> {
    const filters: Record<string, string> = {};
    for (const r of filterRows) {
      const field = r.field.trim();
      const value = r.value.trim();
      if (field && value) {
        filters[field] = value;
      }
    }
    return filters;
  }

  // ÖNEMLİ DEĞİŞİKLİK: handleDropdownAction
  function handleDropdownAction(timelineId: string, action: string) {
    console.log(`Timeline ${timelineId} Dropdown Action:`, action);
    activeDropdownTimelineId = null;
    showDropdown[timelineId] = false;
    showDropdown = showDropdown;

    if (action === "Zaman Çizelgesini Kapat") {
      handleRemoveTimeline(timelineId);
    }
    if (action === "Verileri Yenile") {
      const timeline = get(activeTimelines).find((t) => t.id === timelineId);
      if (timeline) {
        // updateTimeline artık store'u güncelliyor
        updateTimeline(
          timeline.id,
          timeline.currentSearchQuery, // Mevcut sorguyu kullan
          timeline.timelineType
        );
      }
    }
    if (action === "Yeni Zaman Çizelgesi Ekle") {
      addNewEmptyTimeline("roster");
    }
  }

  function handleClickOutsideDropdown(event: MouseEvent) {
    if (activeDropdownTimelineId) {
      const target = event.target as HTMLElement;
      const dropdownButton = document.getElementById(
        `dropdown-button-${activeDropdownTimelineId}`
      );
      const dropdownMenu = document.getElementById(
        `dropdown-menu-${activeDropdownTimelineId}`
      );

      if (
        dropdownButton &&
        !dropdownButton.contains(target) &&
        dropdownMenu &&
        !dropdownMenu.contains(target)
      ) {
        showDropdown[activeDropdownTimelineId] = false;
        activeDropdownTimelineId = null;
        showDropdown = showDropdown;
      }
    }
    if (
      sidebarMenuOpen &&
      !(event.target as HTMLElement).closest(".sidebar-menu-popup")
    ) {
      sidebarMenuOpen = false;
      sidebarMenuTimelineId = null;
    }
  }

  // ÖNEMLİ DEĞİŞİKLİK: handleSearch (Global Arama Çubuğu)
  function handleSearch() {
    // `handleGlobalSearch` zaten timelineManager store'unu güncelliyor.
    // Bu fonksiyon, tüm aktif timelinelarda arama yapacak.
    handleGlobalSearch(searchTerm);
    searchTerm = ""; // Aramadan sonra input'u temizle
  }

  function handleTimelineRowDragStart(event: CustomEvent) {
    const {
      groupKey,
      flightsInGroup,
      timelineId,
      clientX,
      clientY,
      dragOffsetY,
    } = event.detail;

    dragStore.set({
      isDraggingRow: true,
      draggedGroupKey: groupKey,
      draggedRowFlights: flightsInGroup,
      sourceTimelineId: timelineId,
      draggedRowVisualY: clientY,
      dragRowOffsetY: dragOffsetY,
    });

    document.body.style.cursor = "grabbing";
  }

  function handleGlobalMouseMove(event: MouseEvent) {
    if (currentDragState.isDraggingRow) {
      const { dragRowOffsetY } = currentDragState;
      dragStore.update((state) => ({
        ...state,
        draggedRowVisualY: event.clientY - dragRowOffsetY,
      }));

      const AUTO_SCROLL_MARGIN = 50;
      const AUTO_SCROLL_SPEED = 10;
      const scrollContainer = document.querySelector(".timelines-container");
      if (scrollContainer) {
        const rect = scrollContainer.getBoundingClientRect();
        if (
          event.clientY < rect.top + AUTO_SCROLL_MARGIN &&
          scrollContainer.scrollTop > 0
        ) {
          scrollContainer.scrollTop -= AUTO_SCROLL_SPEED;
        } else if (
          event.clientY > rect.bottom - AUTO_SCROLL_MARGIN &&
          scrollContainer.scrollTop <
            scrollContainer.scrollHeight - scrollContainer.clientHeight
        ) {
          scrollContainer.scrollTop += AUTO_SCROLL_SPEED;
        }
      }
    }
  }

  function transferFlightRow({
    sourceTimelineId,
    targetTimelineId,
    draggedRowFlights,
    draggedGroupKey,
  }: {
    sourceTimelineId: string;
    targetTimelineId: string;
    draggedRowFlights: FlightItem[];
    draggedGroupKey: string;
  }): TimelineEntry[] {
    const draggedIds = new Set(draggedRowFlights.map((f) => f.data_id));

    const currentTimelines = get(activeTimelines);

    const sourceTimelineIndex = currentTimelines.findIndex(
      (tl) => tl.id === sourceTimelineId
    );
    const targetTimelineIndex = currentTimelines.findIndex(
      (tl) => tl.id === targetTimelineId
    );

    if (sourceTimelineIndex === -1 || targetTimelineIndex === -1) {
      console.error("Kaynak veya hedef zaman çizelgesi bulunamadı.");
      return currentTimelines;
    }

    const updatedTimelines = [...currentTimelines];
    const sourceTl = { ...updatedTimelines[sourceTimelineIndex] };
    const targetTl = { ...updatedTimelines[targetTimelineIndex] };

    sourceTl.flights = sourceTl.flights.filter(
      (f) => !draggedIds.has(f.data_id)
    );

    const updatedFlightsForTarget = draggedRowFlights.map((f) => {
      const newFlight = structuredClone(f);

      newFlight.person_id = draggedGroupKey.split("_")[0];

      delete newFlight._boundingBox;
      delete (newFlight as any)._originalGroupKey;

      return newFlight;
    });

    targetTl.flights = [
      ...targetTl.flights.filter((f) => !draggedIds.has(f.data_id)),
      ...updatedFlightsForTarget,
    ];

    updatedTimelines[sourceTimelineIndex] = sourceTl;
    updatedTimelines[targetTimelineIndex] = targetTl;

    console.log(
      `✅ Satır "${draggedRowFlights[0]?.person_id}" (eski) → "${draggedGroupKey}" (yeni) ${sourceTimelineId} → ${targetTimelineId} taşındı.`
    );
    activeTimelines.set(updatedTimelines); // Store'u güncellemeyi unutmayın
    return updatedTimelines;
  }

  function handleGlobalMouseUp(event: MouseEvent) {
    if (currentDragState.isDraggingRow) {
      const { draggedGroupKey, draggedRowFlights, sourceTimelineId } =
        currentDragState;

      let targetTimelineId: string | null = null;

      const targetElement = event.target as HTMLElement;
      const targetTimelineCard = targetElement.closest(
        ".timeline-card"
      ) as HTMLElement;

      if (targetTimelineCard) {
        targetTimelineId = targetTimelineCard.dataset.timelineId || null;
      }

      if (
        draggedGroupKey &&
        draggedRowFlights &&
        sourceTimelineId &&
        targetTimelineId
      ) {
        // Drag-drop sonrası ilgili timelineların güncellenmesi.
        // Artık updateTimeline kullanıyoruz.
        transferFlightRow({
          sourceTimelineId,
          targetTimelineId,
          draggedRowFlights,
          draggedGroupKey,
        });

        // Veri taşındıktan sonra kaynak ve hedef timelineların tekrar yüklenmesini isterseniz:
        const sourceTimeline = get(activeTimelines).find(
          (t) => t.id === sourceTimelineId
        );
        if (sourceTimeline) {
          updateTimeline(
            sourceTimeline.id,
            sourceTimeline.currentSearchQuery,
            sourceTimeline.timelineType
          );
        }
        const targetTimeline = get(activeTimelines).find(
          (t) => t.id === targetTimelineId
        );
        if (targetTimeline) {
          updateTimeline(
            targetTimeline.id,
            targetTimeline.currentSearchQuery,
            targetTimeline.timelineType
          );
        }
        console.log(
          `✅ Satır "${draggedGroupKey}" ${sourceTimelineId} → ${targetTimelineId} taşındı. İlgili zaman çizelgeleri yenileniyor.`
        );
      } else {
        console.warn(
          "⚠️ Geçersiz bırakma hedefi veya eksik sürükleme bilgisi. Sürükleme iptal edildi."
        );
      }
    }

    resetDragState();
    document.body.style.cursor = "";
  }

  function handleGlobalKeyDown(event: KeyboardEvent) {
    if (event.key === "Escape" && currentDragState.isDraggingRow) {
      resetDragState();
      document.body.style.cursor = "";
      event.preventDefault();
    }

    if (event.key === "?") {
      showCommandHelp = true;
      event.preventDefault();
    }
    if (event.key === "Escape" && showCommandHelp) {
      showCommandHelp = false;
      event.preventDefault();
    }
  }

  function handleGlobalKeyUp(event: KeyboardEvent) {
    // Diğer küresel klavye olayları
  }

  // handleFlightMoved artık doğrudan timelineManager'ı güncelleyecek
  function handleFlightMoved(event: CustomEvent) {
    console.log("✈️ Tekil uçuş taşındı (handleFlightMoved):", event.detail);
    const {
      sourceTimelineId,
      flight,
      newDepartureTime,
      newArrivalTime,
      newPersonId,
      newRowType,
      originalGroupKey,
    } = event.detail;

    // flightMoved event'i ile gelen güncellenmiş bilgileri kullanarak store'u güncelle
    // Bu, tekil uçuşun hareketini handle etmek için daha detaylı bir işleme ihtiyaç duyar.
    // Sadece timeline'ı yeniden yüklemek yerine, mevcut flights dizisindeki ilgili uçuşu güncelleyebilirsiniz.
    // Ancak hızlı çözüm için, sadece ilgili timeline'ı yeniden yükleyebiliriz:
    const timeline = get(activeTimelines).find(
      (t) => t.id === sourceTimelineId
    );
    if (timeline) {
      // Yeni bir search query oluşturmak yerine, mevcut uçuşu güncelleyen bir aksiyon tanımlamanız gerekebilir
      // veya updateTimeline fonksiyonunu, mevcut verileri alıp sadece bir öğeyi değiştirecek şekilde genişletmeniz.
      // Örneğin: { partialUpdate: { flightId: flight.data_id, newPersonId, newDepartureTime, newArrivalTime } }
      updateTimeline(
        timeline.id,
        timeline.currentSearchQuery, // Mevcut sorguyu koru
        timeline.timelineType
      );
    }
  }

  function handleFlightHover(event: CustomEvent<HoverInfo>) {
    hoverStore.set(event.detail);
  }

  function handleSidebarHeaderDropdown(event: CustomEvent) {
    const { pageX, pageY, timelineId } = event.detail;
    console.log("🟢 Dropdown click:", pageX, pageY, timelineId);
    sidebarMenuOpen = true;
    sidebarMenuX = pageX;
    sidebarMenuY = pageY;
    sidebarMenuTimelineId = timelineId;
  }

  // --- Context Menu Handler Fonksiyonları (Aynı kalmalı) ---
  function handleTimelineHeaderContextMenu(detail: {
    pageX: number;
    pageY: number;
    timelineId: string;
  }) {
    console.log("Timeline: Header context menu alındı.");
    const items = getTimelineHeaderContextMenu();
    showContextMenu(detail.pageX, detail.pageY, items, {
      type: ContextMenuType.TIMELINE_HEADER,
      timelineId: detail.timelineId,
    });
  }

  function handleTimelineEmptyContextMenu(detail: {
    pageX: number;
    pageY: number;
    timelineId: string;
    coords: { x: number; y: number };
  }) {
    console.log("Timeline: Empty area context menu alındı.");
    const items = getTimelineEmptyContextMenu();
    showContextMenu(detail.pageX, detail.pageY, items, {
      type: ContextMenuType.TIMELINE_EMPTY,
      timelineId: detail.timelineId,
      coords: detail.coords,
    });
  }

  function handleTimelineRowContextMenu(detail: {
    personId: string;
    rowType: RowType;
    pageX: number;
    pageY: number;
    timelineId: string;
  }) {
    console.log(
      "Timeline: Row context menu alındı for",
      detail.personId,
      detail.rowType
    );
    const items = getTimelineRowContextMenu({
      rowId: detail.personId,
      rowType: detail.rowType,
    });
    showContextMenu(detail.pageX, detail.pageY, items, {
      type: ContextMenuType.TIMELINE_ROW,
      personId: detail.personId,
      rowType: detail.rowType,
      timelineId: detail.timelineId,
    });
  }
</script>

<div class="app-container">
  <div class="timelines-container">
    {#each $activeTimelines
      .slice()
      .sort( (a, b) => (a.isMinimized === b.isMinimized ? 0 : a.isMinimized ? 1 : -1) ) as timeline, index (timeline.id)}
      <div
        class="timeline-card"
        data-timeline-id={timeline.id}
        style={timeline.isMinimized
          ? "flex: 0 0 34px;"
          : `flex: ${timeline.heightRatio} 1 0%;`}
      >
        <div class="timeline-header">
          <div class="timeline-header-left">
            <div class="dropdown-wrapper">
              <button
                id="dropdown-button-{timeline.id}"
                class="dropdown-button"
                on:click={(e) => toggleDropdown(timeline.id, e)}
              >
                ⋮
              </button>
              {#if showDropdown[timeline.id]}
                <div id="dropdown-menu-{timeline.id}" class="dropdown-menu">
                  <button
                    on:click={() =>
                      handleDropdownAction(timeline.id, "Verileri Yenile")}
                  >
                    Verileri Yenile
                  </button>
                  <button
                    on:click={() =>
                      handleDropdownAction(
                        timeline.id,
                        "Zaman Çizelgesini Kapat"
                      )}
                  >
                    Zaman Çizelgesini Kapat
                  </button>
                  <button
                    on:click={() =>
                      handleDropdownAction(
                        timeline.id,
                        "Yeni Zaman Çizelgesi Ekle"
                      )}
                  >
                    Yeni Zaman Çizelgesi Ekle
                  </button>
                </div>
              {/if}
            </div>

            <div class="filter-dropdown-wrapper">
              <select
                bind:value={selectedFiltersPerTimeline[timeline.id]}
                on:change={async () => {
                  const filterId = selectedFiltersPerTimeline[timeline.id];
                  const filter = allFilters.find((f) => f.id === filterId);
                  if (!filter) return;

                  const filters = buildFiltersFromSavedFilter(filter.rows);
                  // updateTimeline'ı kullanarak store'u güncelle
                  await updateTimeline(
                    timeline.id,
                    filters,
                    timeline.timelineType
                  );
                  if (filters.person_id) {
                    naturalQueryInputs[timeline.id] = filters.person_id;
                  } else {
                    naturalQueryInputs[timeline.id] = "";
                  }
                }}
              >
                <option value="">Filtre Seç</option>
                {#each allFilters as f}
                  <option value={f.id}>{f.name}</option>
                {/each}
              </select>
            </div>

            <div class="toolbar-separator"></div>
            <div class="zoom-controls">
              <button on:click={() => handleZoom(timeline.id, "out")}>➖</button
              >
              <button on:click={() => handleZoom(timeline.id, "in")}>➕</button>
            </div>

            <div class="toolbar-separator"></div>

            <div class="highlight-jump-wrapper">
              <input
                type="text"
                placeholder="Sicil bul..."
                bind:value={highlightInputs[timeline.id]}
                on:keydown={(e) =>
                  e.key === "Enter" && handleHighlightJump(timeline.id)}
              />
              <button
                class="highlight-search-btn"
                on:click={() => handleHighlightJump(timeline.id)}
                title="Ara"
              >
                🔍
              </button>
            </div>
          </div>

          <div class="timeline-header-right">
            <div class="nl-query-wrapper">
              <div class="nl-query-input-group">
                <input
                  type="text"
                  placeholder="Doğal Dil ile Ara..."
                  bind:value={naturalQueryInputs[timeline.id]}
                  on:keydown={(e) => {
                    if (e.key === "Enter") handleNaturalQuery(timeline.id);
                  }}
                />
                <button
                  class="search-btn"
                  on:click={() => handleNaturalQuery(timeline.id)}
                  aria-label="Ara"
                >
                  🔍
                </button>
              </div>
            </div>
            <button
              class="min-toggle-btn"
              on:click={() => toggleMinimize(timeline.id)}
            >
              {timeline.isMinimized ? "▲" : "▼"}
            </button>
            <button
              class="remove-timeline-button"
              on:click={() => handleRemoveTimeline(timeline.id)}
            >
              X
            </button>
          </div>
        </div>

        {#if timeline.loading}
          <div class="loading-overlay">
            <div class="spinner-container">
              <div class="spinner"></div>
            </div>
            <p>Veriler yükleniyor...</p>
          </div>
        {/if}
        {#if !timeline.isMinimized}
          {#if timeline.error}
            <p class="error-message">{timeline.error}</p>
          {:else if timeline.flights.length === 0 && Object.keys(timeline.currentSearchQuery).length > 0 && !timeline.loading}
            <p class="status-message">
              Sorgunuz "{JSON.stringify(timeline.currentSearchQuery)}" için uçuş
              bulunamadı.
            </p>
          {:else if timeline.timelineType === "roster"}
            <RosterTimeline
              timelineId={timeline.id}
              on:flightMoved={handleFlightMoved}
              on:rowDragStart={handleTimelineRowDragStart}
              on:flightHover={handleFlightHover}
              on:sidebarHeaderDropdownClick={handleSidebarHeaderDropdown}
              on:flightItemContextMenu={(e) =>
                handleFlightItemContextMenu(e.detail)}
              on:timelineHeaderContextMenu={(e) =>
                handleTimelineHeaderContextMenu(e.detail)}
              on:timelineEmptyContextMenu={(e) =>
                handleTimelineEmptyContextMenu(e.detail)}
              on:rowContextMenu={(e) => handleTimelineRowContextMenu(e.detail)}
              on:rowSidebarContextMenu={(e) =>
                handleTimelineRowSidebarContextMenu(e.detail)}
            />
          {:else if timeline.timelineType === "trip"}
            <TripTimeline
              flights={timeline.flights}
              timelineId={timeline.id}
              isLoading={timeline.loading}
              loadingProgress={timeline.loadingProgress}
            />
          {:else if timeline.timelineType === "rotation"}
            <RotationTimeline
              flights={timeline.flights}
              timelineId={timeline.id}
              isLoading={timeline.loading}
              loadingProgress={timeline.loadingProgress}
            />
          {/if}
        {/if}
      </div>

      {#if index < $activeTimelines.length - 1 && !timeline.isMinimized && !$activeTimelines[index + 1].isMinimized}
        <!-- svelte-ignore a11y_no_noninteractive_element_interactions -->
        <div
          class="timeline-splitter"
          role="separator"
          aria-orientation="horizontal"
          on:mousedown={(e) => startResizing(index, e)}
        ></div>
      {/if}
    {/each}
  </div>

  {#if sidebarMenuOpen && sidebarMenuTimelineId}
    <div
      class="sidebar-menu-popup"
      style="position: fixed; top: {sidebarMenuY}px; left: {sidebarMenuX -
        120}px;"
    >
      <button
        on:click={() =>
          handleDropdownAction(sidebarMenuTimelineId!, "Verileri Yenile")}
      >
        Verileri Yenile
      </button>
      <button
        on:click={() =>
          handleDropdownAction(
            sidebarMenuTimelineId!,
            "Zaman Çizelgesini Kapat"
          )}
      >
        Zaman Çizelgesini Kapat
      </button>
      <button
        on:click={() =>
          handleDropdownAction(
            sidebarMenuTimelineId!,
            "Yeni Zaman Çizelgesi Ekle"
          )}
      >
        Yeni Zaman Çizelgesi Ekle
      </button>
    </div>
  {/if}

  <div class="timeline-screen-footer">
    <div class="info-screen-section">
      {#if currentHoverInfo?.type === "flight"}
        <div class="hover-info-box">
          <div><strong>ID:</strong> {currentHoverInfo.item.data_id}</div>
          <div>
            <strong>Başlangıç:</strong>
            {formatDateTime(currentHoverInfo.item.departure_time)}
          </div>
          <div>
            <strong>Bitiş:</strong>
            {formatDateTime(currentHoverInfo.item.arrival_time)}
          </div>
          <div><strong>Tip:</strong> {currentHoverInfo.item.activity_code}</div>

          {#if currentHoverInfo.item.flight_no}
            <div>
              <strong>Uçuş No:</strong>
              {currentHoverInfo.item.flight_no}
            </div>
          {/if}
          {#if currentHoverInfo.item.person_id}
            <div>
              <strong>Personel ID:</strong>
              {currentHoverInfo.item.person_id}
            </div>
          {/if}
          {#if currentHoverInfo.item.trip_id}
            <div>
              <strong>Sefer ID:</strong>
              {currentHoverInfo.item.trip_id}
            </div>
          {/if}
          {#if currentHoverInfo.item.plane_tail_name}
            <div>
              <strong>Kuyruk No:</strong>
              {currentHoverInfo.item.plane_tail_name}
            </div>
          {/if}
          {#if currentHoverInfo.item.flight_id}
            <div>
              <strong>Uçuş ID:</strong>
              {currentHoverInfo.item.flight_id}
            </div>
          {/if}
        </div>
      {:else if currentHoverInfo?.type === "rest"}
        <div class="hover-info-box">
          <div>
            <strong>Rest Başlangıç:</strong>
            {formatDateTime(currentHoverInfo.rest_info.rest_start)}
          </div>
          <div>
            <strong>Rest Bitiş:</strong>
            {formatDateTime(currentHoverInfo.rest_info.rest_end)}
          </div>
          <div>
            <strong>Rest Süresi:</strong>
            {#if currentHoverInfo.rest_info.rest_duration != null}
              {currentHoverInfo.rest_info.rest_duration} saat
            {:else}
              Hesaplanmadı
            {/if}
          </div>
        </div>
      {:else if currentHoverInfo?.type === "trip"}
        <div class="hover-info-box">
          <div><strong>Sefer ID:</strong> {currentHoverInfo.tripId}</div>
          <div>
            <strong>Kalkış:</strong>
            {currentHoverInfo.departure_port}
          </div>
          <div>
            <strong>Varış:</strong>
            {currentHoverInfo.arrival_port}
          </div>
          <div>
            <strong>Başlangıç Zamanı:</strong>
            {formatDateTime(currentHoverInfo.departure_time)}
          </div>
          <div>
            <strong>Bitiş Zamanı:</strong>
            {formatDateTime(currentHoverInfo.arrival_time)}
          </div>
          <div>
            <strong>Rotasyon Süresi:</strong>
            {currentHoverInfo.rotation_minutes} dakika
          </div>
          {#if currentHoverInfo.ftl_violations && currentHoverInfo.ftl_violations.length > 0}
            <div>
              <strong>FTL İhlalleri:</strong>
              {currentHoverInfo.ftl_violations.join(", ")}
            </div>
          {/if}
        </div>
      {:else}
        <div class="hover-info-placeholder">Hover Info Screen</div>
      {/if}
    </div>

    <div class="search-section">
      <div class="modern-search-box input-with-help-icon">
        <input
          type="text"
          class="modern-search-input"
          placeholder="Query Search"
          bind:value={searchTerm}
          on:keypress={(e) => e.key === "Enter" && handleSearch()}
        />

        <button
          class="input-help-icon"
          title="Komut Yardımı"
          on:click={() => (showCommandHelp = true)}
          type="button"
        >
          ?
        </button>

        <button class="modern-search-button" on:click={handleSearch}>
          Ara
        </button>
      </div>
    </div>
  </div>

  {#if currentDragState.isDraggingRow && currentDragState.draggedGroupKey}
    <div
      class="dragged-row-ghost"
      style="top: {currentDragState.draggedRowVisualY}px;"
    >
      <div class="dragged-row-content">
        <span class="dragged-row-label">
          {currentDragState.draggedGroupKey}
        </span>
        {#if currentDragState.draggedRowFlights && currentDragState.draggedRowFlights.length > 0}
          <span class="dragged-row-flight-count">
            ({currentDragState.draggedRowFlights.length} uçuş)
          </span>
        {/if}
      </div>
    </div>
  {/if}

  {#if showCommandHelp}
    <CommandHelp
      onClose={() => (showCommandHelp = false)}
      onCommandSelect={handleCommandInsert}
    />
  {/if}
</div>

<style>
  .app-container {
    display: flex;
    flex-direction: column;
    height: 100vh;
    overflow: hidden;
  }

  .timelines-container {
    display: flex;
    flex-direction: column;
    flex-grow: 1;
    overflow-y: auto;
    overflow-x: hidden;
    min-height: 0;
  }

  .timeline-card {
    border: 1px solid #ddd;
    box-shadow: 0 2px 5px rgba(0, 0, 0, 0.05);
    overflow: hidden;
    display: flex;
    flex-direction: column;
    flex-grow: 1;
    flex-shrink: 1;
    flex-basis: 0;
    min-height: 15px;
    position: relative; /* Crucial for loading-overlay positioning */
  }

  .timeline-header {
    display: flex;
    justify-content: space-between;
    align-items: center;
    padding: 2px 15px;
    background-color: #e0e0e0;
    border-bottom: 1px solid #ccc;
    flex-shrink: 0;
    cursor: grab;
  }
  .timeline-header-right {
    display: flex;
  }

  .remove-timeline-button {
    background: none;
    border: none;
    color: #dc3545;
    font-size: 1.2em;
    cursor: pointer;
    transition: color 0.2s ease;
    flex-shrink: 0;
    width: 24px;
    text-align: center;
  }

  .remove-timeline-button:hover {
    color: #a71d2a;
  }

  .status-message,
  .error-message {
    padding: 15px;
    text-align: center;
    color: #555;
    flex-grow: 1;
    display: flex;
    align-items: center;
    justify-content: center;
  }

  .error-message {
    color: #dc3545;
    background-color: #f8d7da;
    border: 1px solid #f5c6cb;
    border-radius: 4px;
    margin: 15px;
  }

  .dropdown-wrapper {
    position: relative;
    z-index: 50;
    flex-shrink: 0;
    width: 24px;
    text-align: center;
  }

  .dropdown-button {
    background: none;
    border: 1px solid #ccc;
    border-radius: 4px;
    padding: 3px 6px;
    font-size: 16px;
    cursor: pointer;
    color: #555;
    background-color: #f0f0f0;
    line-height: 1;
    display: inline-flex;
    align-items: center;
    justify-content: center;
    width: 100%;
    box-sizing: border-box;
  }

  .dropdown-button:hover {
    background-color: #e0e0e0;
    color: #333;
  }

  .dropdown-menu {
    position: absolute;
    top: 100%;
    left: 0;
    background-color: white;
    border: 1px solid #ddd;
    box-shadow: 0 2px 8px rgba(0, 0, 0, 0.15);
    border-radius: 4px;
    min-width: 150px;
    display: flex;
    flex-direction: column;
    z-index: 51;
  }

  .dropdown-menu button {
    background: none;
    border: none;
    padding: 8px 12px;
    text-align: left;
    cursor: pointer;
    font-size: 14px;
    color: #333;
    width: 100%;
  }

  .dropdown-menu button:hover {
    background-color: #f0f0f0;
  }

  .dragged-row-ghost {
    position: fixed;
    left: 0;
    width: 100%;
    height: 40px;
    background-color: rgba(0, 123, 255, 0.7);
    border: 1px solid #0056b3;
    box-shadow: 0 4px 10px rgba(0, 0, 0, 0.2);
    z-index: 9999;
    pointer-events: none;
    display: flex;
    align-items: center;
    padding-left: 10px;
    box-sizing: border-box;
  }

  .dragged-row-ghost .dragged-row-content {
    color: white;
    font-weight: bold;
    font-size: 14px;
    white-space: nowrap;
    overflow: hidden;
    text-overflow: ellipsis;
  }

  .dragged-row-ghost .dragged-row-label {
    margin-right: 5px;
  }
  .timeline-screen-footer {
    display: flex;
    justify-content: space-between;
    align-items: stretch;
    gap: 10px;

    background-color: #f8f9fa;
    border-bottom: 1px solid #e0e0e0;
    flex-shrink: 0;
  }

  .info-screen-section {
    flex: 1;
    display: flex;
    justify-content: center;
    align-items: center;
    min-width: 250px;
    height: 75px;
    padding: 5px 10px;
    font-size: 13px;
    color: #333;
    box-sizing: border-box;
    border-right: 1px solid #ccc;
  }

  .hover-info-box {
    display: grid;
    grid-template-columns: repeat(4, 1fr);
    gap: 0;
    align-items: center;
    width: 100%;
    font-size: 13px;
    color: #333;
    border-radius: 4px;
    overflow: hidden;
  }

  .hover-info-box > * {
    padding: 6px 8px;
    white-space: nowrap;
    overflow: hidden;
    text-overflow: ellipsis;
  }

  .hover-info-box > :not(:nth-child(4n)) {
    border-right: 1px solid #ccc;
  }

  .hover-info-box strong {
    text-align: right;
    padding-right: 5px;
  }

  .hover-info-placeholder {
    font-style: italic;
    color: #888;
    padding: 0 15px;
  }

  .search-section {
    display: flex;
    justify-content: flex-end;
    align-items: center;
    padding: 8px 16px;
  }

  .modern-search-box {
    display: flex;
    border: 1px solid #ccc;
    border-radius: 6px;
    overflow: hidden;
    max-width: 460px;
    width: 100%;
    background-color: white;
    box-shadow: 0 1px 2px rgba(0, 0, 0, 0.04);
    transition: border-color 0.2s ease;
  }

  .modern-search-input {
    flex: 1;
    border: none;
    padding: 10px 14px;
    font-size: 14px;
    outline: none;
    background: transparent;
    color: #333;
  }

  .modern-search-button {
    --btn: var(--thy-primary-red, #d82b2b);
    /* Destek varsa color-mix ile hover/active tonları */
    --btn-hover: color-mix(in srgb, var(--btn) 92%, black);
    --btn-active: color-mix(in srgb, var(--btn) 84%, black);
    color: white;
    border: none;
    padding: 0 20px;
    font-size: 14px;
    font-weight: 500;
    cursor: pointer;
    transition: background-color 0.2s ease-in-out;
    white-space: nowrap;
    height: 100%;
  }

  .modern-search-button:hover {
    background-color: var(
      --btn-hover
    ); /* artık #ccc değil, kırmızı tonun koyusu */
    box-shadow:
      0 12px 20px -12px rgba(0, 0, 0, 0.28),
      0 3px 8px rgba(0, 0, 0, 0.16);
  }

  .modern-search-button:active {
    background-color: var(--btn-active);
    box-shadow: 0 2px 4px rgba(0, 0, 0, 0.18) inset;
  }

  .sidebar-menu-popup {
    background: #fff;
    border: 1px solid #ddd;
    box-shadow: 0 2px 8px rgba(0, 0, 0, 0.15);
    border-radius: 6px;
    padding: 4px 0;
    min-width: 160px;
    z-index: 999;
  }

  .sidebar-menu-popup button {
    display: block;
    width: 100%;
    background: none;
    border: none;
    text-align: left;
    padding: 8px 14px;
    font-size: 14px;
    cursor: pointer;
  }

  .sidebar-menu-popup button:hover {
    background: #f5f5f5;
  }
  .filter-dropdown-wrapper {
    margin-left: 8px;
  }

  .filter-dropdown-wrapper select {
    padding: 3px 4px;
    font-size: 12px;
    border-radius: 4px;
    border: 1px solid #ccc;
    background: white;
    cursor: pointer;
  }
  .timeline-header-left {
    display: flex;
  }
  .timeline-splitter {
    height: 6px;
    background: #ccc;
    cursor: row-resize;
    user-select: none;
    flex-shrink: 0;
  }
  .timeline-splitter:hover {
    background: #999;
  }
  .min-toggle-btn {
    margin-right: 6px;
    font-size: 12px;
    background: none;
    border: none;
    cursor: pointer;
  }
  .zoom-controls {
    display: flex;
    justify-content: center;
    gap: 6px;
    margin-right: 12px;
  }
  .zoom-controls button {
    width: 24px;
    height: 24px;
    border: 1px solid #bbb;
    border-radius: 4px;
    background: #f2f2f2;
    cursor: pointer;
    display: flex;
    justify-content: center;
  }
  .toolbar-separator {
    width: 1px;
    background-color: #ccc;
    margin: 0 10px;
    height: 24px;
    align-self: center;
  }
  .nl-query-wrapper {
    display: flex;
    align-items: center;
    margin-left: 8px;
  }

  .nl-query-input-group {
    display: flex;
    border: 1px solid #ccc;
    border-radius: 6px;
    overflow: hidden;
    background: white;
  }

  .nl-query-input-group input {
    border: none;
    padding: 3px 4px;
    outline: none;
    width: 180px;
  }

  .nl-query-input-group .search-btn {
    background-color: #eee;
    border: none;
    padding: 3px 4px;
    cursor: pointer;
    font-size: 1rem;
    transition: background 0.2s;
  }

  .nl-query-input-group .search-btn:hover {
    background-color: #ddd;
  }
  .highlight-jump-wrapper {
    display: flex;
    align-items: center;
    gap: 4px;
    margin-left: 8px;
  }

  .highlight-jump-wrapper input {
    font-size: 12px;
    padding: 3px 4px;
    border: 1px solid #ccc;
    border-radius: 4px;
    width: 110px;
  }

  .highlight-search-btn {
    font-size: 14px;
    padding: 2px 6px;
    background: #eee;
    border: 1px solid #ccc;
    border-radius: 4px;
    cursor: pointer;
    transition: background 0.2s;
  }

  .highlight-search-btn:hover {
    background: #ddd;
  }
  .input-with-help-icon {
    position: relative;
    display: flex;
    align-items: center;
    width: 100%;
    max-width: 460px;
    background: white;
    border: 1px solid #ccc;
    border-radius: 6px;
    overflow: hidden;
  }

  .input-with-help-icon .modern-search-input {
    flex: 1;
    padding: 10px 34px 10px 14px;
    font-size: 14px;
    border: none;
    background: transparent;
    outline: none;
    color: #333;
  }

  .input-help-icon {
    position: fixed;
    right: 85px;
    display: inline-grid;
    place-items: center;
    width: 28px;
    height: 28px;
    border-radius: 999px;
    border: 1px solid rgba(0, 0, 0, 0.12);
    color: #555;
    font-size: 16px;
    font-weight: 700;
    cursor: pointer;
    padding: 0;
    z-index: 1;
    box-shadow: 0 1px 2px rgba(0, 0, 0, 0.08);
    transition:
      transform 0.12s ease,
      box-shadow 0.12s ease,
      background-color 0.12s ease,
      color 0.12s ease;
  }
  .input-help-icon:hover {
    background: rgba(0, 0, 0, 0.07);
    color: #111;
    box-shadow: 0 2px 6px rgba(0, 0, 0, 0.14);
  }
  .input-help-icon:active {
    transform: translateY(1px) scale(0.98);
    box-shadow: 0 1px 3px rgba(0, 0, 0, 0.12) inset;
  }
  .input-help-icon:focus-visible {
    outline: 2px solid transparent;
    box-shadow:
      0 0 0 3px rgba(25, 103, 210, 0.25),
      0 2px 6px rgba(0, 0, 0, 0.14);
  }

  .modern-search-button {
    background-color: var(--thy-primary-red);
    color: white;
    border: none;
    padding: 0 20px;
    font-size: 14px;
    font-weight: 500;
    cursor: pointer;
    transition: background-color 0.2s ease-in-out;
    white-space: nowrap;
    height: 40px;
  }

  /* --- UPDATED STYLES FOR LOADING INDICATOR (SPINNER) --- */
  .loading-overlay {
    position: absolute;
    top: 0;
    left: 0;
    width: 100%;
    height: 100%;
    background-color: rgba(255, 255, 255, 0.85); /* Slightly less transparent */
    display: flex;
    flex-direction: column; /* To stack spinner and text */
    justify-content: center;
    align-items: center;
    z-index: 10;
    padding: 20px;
    box-sizing: border-box;
    pointer-events: none; /* Allows interaction with elements behind if needed */
  }

  .spinner-container {
    display: flex;
    flex-direction: column;
    align-items: center;
    /* No need for margin-bottom here, it's handled on the spinner itself */
  }

  .spinner {
    border: 4px solid rgba(0, 0, 0, 0.1); /* Light gray border for the circle */
    width: 30px; /* Slightly smaller width */
    height: 30px; /* Slightly smaller height */
    border-radius: 50%; /* Makes it a circle */
    border-left-color: var(
      --thy-primary-red,
      #dc3545
    ); /* Corporate color for the spinning part. Added a fallback. */
    animation: spin 0.8s linear infinite; /* Faster and smoother spin */
    margin-bottom: 8px; /* Space between spinner and text */
  }

  @keyframes spin {
    0% {
      transform: rotate(0deg);
    }
    100% {
      transform: rotate(360deg);
    }
  }

  .loading-overlay p {
    font-size: 13px; /* Smaller, more subtle text */
    color: #666; /* Softer text color */
    margin-top: 0; /* Remove default paragraph margin */
  }
</style>
