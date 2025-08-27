<script lang="ts">
  import VirtualList from "@/modules/VirtualList.svelte";
  import { writable, derived, get } from "svelte/store";

  // ====== Props ======
  export let items: Record<string, any>[] = [];
  export let columns: {
    key: string;
    label: string;
    width?: number; // px (varsayılan: eşit pay)
    type?: "text" | "number" | "date";
    format?: (value: any, row?: any) => string;
    editable?: boolean;
    comparator?: (a: any, b: any, asc: boolean) => number;
  }[] = [];

  // ====== Local state ======
  const itemsStore = writable(items);
  $: itemsStore.set(items);

  const search = writable("");
  const sortStack = writable<{ key: string; asc: boolean }[]>([]);
  const colWidths = writable<number[]>([]);

  const activeCell = writable<{ row: number; col: number } | null>(null);
  const editingCell = writable<{ row: number; col: number } | null>(null);

  // Scroll containers
  let headerScrollEl: HTMLDivElement;
  let bodyScrollEl: HTMLDivElement;

  // İlk yükte kolon genişlikleri
  $: if (get(colWidths).length === 0) {
    colWidths.set(columns.map((c) => c.width ?? -1)); // -1: auto
  }

  // ====== Helpers ======
  function normalizeStr(v: any) {
    return String(v ?? "").toLocaleLowerCase();
  }
  function parseNumber(v: any) {
    if (typeof v === "number") return v;
    const n = Number(String(v).replace(",", "."));
    return Number.isNaN(n) ? Number.NEGATIVE_INFINITY : n;
  }
  function parseDate(v: any) {
    if (typeof v === "number") return v;
    const t = Date.parse(v);
    return Number.isNaN(t) ? Number.NEGATIVE_INFINITY : t;
  }

  function defaultCompare(
    a: any,
    b: any,
    type: "text" | "number" | "date" | undefined,
    asc: boolean
  ) {
    let r = 0;
    if (type === "number") r = parseNumber(a) - parseNumber(b);
    else if (type === "date") r = parseDate(a) - parseDate(b);
    else
      r = normalizeStr(a).localeCompare(normalizeStr(b), undefined, {
        numeric: true,
        sensitivity: "base",
      });
    return asc ? r : -r;
  }

  function stableSort<T>(arr: T[], cmp: (a: T, b: T) => number) {
    return arr
      .map((v, i) => ({ v, i }))
      .sort((A, B) => {
        const r = cmp(A.v, B.v);
        return r !== 0 ? r : A.i - B.i;
      })
      .map((e) => e.v);
  }

  function toggleSort(key: string, shiftKey = false) {
    sortStack.update((stack) => {
      const idx = stack.findIndex((s) => s.key === key);
      if (!shiftKey) {
        if (idx === 0) return [{ key, asc: !stack[0].asc }];
        return [{ key, asc: true }];
      } else {
        if (idx >= 0) {
          const newStack = [...stack];
          newStack[idx] = { key, asc: !newStack[idx].asc };
          return newStack;
        } else {
          return [...stack, { key, asc: true }];
        }
      }
    });
  }

  const filtered = derived(
    [itemsStore, search, sortStack],
    ([$items, $search, $sortStack]) => {
      let result = $items;

      const s = $search.trim().toLocaleLowerCase();
      if (s) {
        result = result.filter((row) =>
          columns.some((col) => normalizeStr(row[col.key]).includes(s))
        );
      }

      if ($sortStack.length > 0) {
        result = stableSort(result, (a, b) => {
          for (const srt of $sortStack) {
            const col = columns.find((c) => c.key === srt.key)!;
            const cmp = col.comparator
              ? col.comparator(a[col.key], b[col.key], srt.asc)
              : defaultCompare(a[col.key], b[col.key], col.type, srt.asc);
            if (cmp !== 0) return cmp;
          }
          return 0;
        });
      }

      return result;
    }
  );

  function onHeaderClick(e: MouseEvent, key: string) {
    toggleSort(key, e.shiftKey);
  }

  // ====== Column resize ======
  let headerRef: HTMLDivElement;
  let bodyRef: HTMLDivElement;

  function startResize(e: PointerEvent, colIndex: number) {
    (e.target as HTMLElement).setPointerCapture(e.pointerId);
    const startX = e.clientX;
    const startWidths = get(colWidths).slice();

    function onMove(ev: PointerEvent) {
      const delta = ev.clientX - startX;
      colWidths.update((ws) => {
        const w = Math.max(
          60,
          (startWidths[colIndex] > 0
            ? startWidths[colIndex]
            : getAutoWidth(colIndex)) + delta
        );
        ws[colIndex] = w;
        return [...ws];
      });
      // genişlik değişiminde de hizala
      queueMicrotask(syncFromBody);
    }
    function onUp() {
      (e.target as HTMLElement).releasePointerCapture(e.pointerId);
      window.removeEventListener("pointermove", onMove);
      window.removeEventListener("pointerup", onUp);
    }
    window.addEventListener("pointermove", onMove);
    window.addEventListener("pointerup", onUp);
  }

  function getAutoWidth(colIndex: number) {
    const total = headerRef?.clientWidth ?? 1000;
    const fixed = get(colWidths)
      .filter((w) => w > 0)
      .reduce((a, b) => a + b, 0);
    const remainingCols =
      columns.filter((_, i) => get(colWidths)[i] <= 0).length || 1;
    return Math.max(120, Math.floor((total - fixed) / remainingCols));
  }

  // Grid template columns string
  const gridTemplate = derived(colWidths, ($ws) =>
    $ws.map((w, i) => (w > 0 ? `${w}px` : `${getAutoWidth(i)}px`)).join(" ")
  );

  // gridTemplate değişince bir sonraki frame’de senkronu uygula
  $: {
    // Svelte DOM güncellemesi sonrası çalışsın
    requestAnimationFrame(() => syncFromBody());
  }

  function cellClass(r: number, c: number) {
    const ac = get(activeCell);
    const ed = get(editingCell);
    let cls = "cell";
    if (ac && ac.row === r && ac.col === c) cls += " cell--active";
    if (ed && ed.row === r && ed.col === c) cls += " cell--editing";
    return cls;
  }

  function onCellClick(rowIdx: number, colIdx: number) {
    activeCell.set({ row: rowIdx, col: colIdx });
  }

  function onCellDblClick(rowIdx: number, colIdx: number) {
    const col = columns[colIdx];
    if (col?.editable) editingCell.set({ row: rowIdx, col: colIdx });
  }

  function onKeyDown(e: KeyboardEvent) {
    const ac = get(activeCell);
    if (!ac) return;
    const maxRow = get(filtered).length - 1;
    const maxCol = columns.length - 1;

    if (e.key === "Enter") {
      const col = columns[ac.col];
      if (col?.editable) editingCell.set({ row: ac.row, col: ac.col });
      else activeCell.set({ row: Math.min(ac.row + 1, maxRow), col: ac.col });
      e.preventDefault();
    } else if (e.key === "ArrowDown") {
      activeCell.set({ row: Math.min(ac.row + 1, maxRow), col: ac.col });
      e.preventDefault();
    } else if (e.key === "ArrowUp") {
      activeCell.set({ row: Math.max(ac.row - 1, 0), col: ac.col });
      e.preventDefault();
    } else if (e.key === "ArrowRight") {
      activeCell.set({ row: ac.row, col: Math.min(ac.col + 1, maxCol) });
      e.preventDefault();
    } else if (e.key === "ArrowLeft") {
      activeCell.set({ row: ac.row, col: Math.max(ac.col - 1, 0) });
      e.preventDefault();
    } else if ((e.ctrlKey || e.metaKey) && e.key.toLowerCase() === "c") {
      const row = get(filtered)[ac.row];
      const line = columns.map((c) => String(row[c.key] ?? "")).join("\t");
      navigator.clipboard.writeText(line);
    }
  }

  function commitEdit(rowIdx: number, colIdx: number, value: string) {
    const key = columns[colIdx].key;
    itemsStore.update((arr) => {
      const copy = arr.slice();
      const row = { ...copy[rowIdx], [key]: value };
      copy[rowIdx] = row;
      return copy;
    });
    editingCell.set(null);
  }

  // ====== Scroll Sync (one-way: body -> header) ======
  function syncFromBody() {
    if (!headerScrollEl || !bodyScrollEl) return;
    headerScrollEl.scrollLeft = bodyScrollEl.scrollLeft;
  }
</script>

<!-- Toolbar -->
<div class="toolbar" on:keydown={onKeyDown} tabindex="0">
  <input
    class="search"
    placeholder="Ara..."
    on:input={(e) => search.set((e.target as HTMLInputElement).value)}
  />
  <div class="info">
    Toplam: {get(itemsStore).length} — Görüntülenen: {get(filtered).length}
  </div>
</div>

<!-- Scroll Sync Container -->
<div class="table-wrap">
  <!-- HEADER (scroll container) -->
  <div class="thead-scroll" bind:this={headerScrollEl}>
    <div
      class="thead"
      bind:this={headerRef}
      style={`grid-template-columns:${$gridTemplate};`}
    >
      {#each columns as col, i}
        <div class="th" on:click={(e) => onHeaderClick(e, col.key)}>
          <div class="th__label">
            {col.label}
            {#if get(sortStack).some((s) => s.key === col.key)}
              {#each get(sortStack) as s, idx}
                {#if s.key === col.key}
                  <span class="sort-badge"
                    >{s.asc ? "▲" : "▼"}{get(sortStack).length > 1
                      ? ` ${idx + 1}`
                      : ""}</span
                  >
                {/if}
              {/each}
            {/if}
          </div>
          <div
            class="th__resizer"
            title="Sütunu yeniden boyutlandır"
            on:pointerdown={(e) => startResize(e, i)}
          />
        </div>
      {/each}
    </div>
  </div>

  <!-- BODY (scroll container) -->
  <div class="tbody-scroll" bind:this={bodyScrollEl} on:scroll={syncFromBody}>
    <div class="tbody" bind:this={bodyRef} style={`--grid:${$gridTemplate};`}>
      {#if $filtered.length > 0}
        <VirtualList
          items={$filtered}
          itemHeight={40}
          height={700}
          let:item
          let:index
        >
          <div class="row" style={`grid-template-columns:${$gridTemplate};`}>
            {#each columns as col, ci}
              <div
                class={cellClass(index, ci)}
                on:click={() => onCellClick(index, ci)}
                on:dblclick={() => onCellDblClick(index, ci)}
              >
                {#if get(editingCell)?.row === index && get(editingCell)?.col === ci && col.editable}
                  <input
                    class="cell-editor"
                    value={String(item[col.key] ?? "")}
                    on:blur={(e) =>
                      commitEdit(
                        index,
                        ci,
                        (e.target as HTMLInputElement).value
                      )}
                    on:keydown={(e) => {
                      if (e.key === "Enter")
                        commitEdit(
                          index,
                          ci,
                          (e.target as HTMLInputElement).value
                        );
                      if (e.key === "Escape") editingCell.set(null);
                      e.stopPropagation();
                    }}
                    autofocus
                  />
                {:else if col.format}
                  {col.format(item[col.key], item)}
                {:else}
                  {item[col.key]}
                {/if}
              </div>
            {/each}
          </div>
        </VirtualList>
      {:else}
        <p style="padding: 16px;">Sonuç bulunamadı.</p>
      {/if}
    </div>
  </div>
</div>

<style>
  .toolbar {
    display: flex;
    gap: 12px;
    align-items: center;
    margin-bottom: 6px;
  }
  .search {
    flex: 1;
    padding: 8px 12px;
    font-size: 13px;
    border: 1px solid #dadce0;
    border-radius: 4px;
  }
  .info {
    font-size: 12px;
    color: #444;
  }

  .table-wrap {
    border: 1px solid #dadce0;
    border-radius: 6px;
    overflow: hidden;
  }

  /* Header scroll container: scrollbar gizli, ama scrollLeft uygulanabilir */
  .thead-scroll {
    overflow-x: auto; /* programatik scroll için açık */
    overflow-y: hidden;
    position: sticky;
    top: 0;
    z-index: 2;
    background: #f8f9fa;
    border-bottom: 1px solid #dadce0;
    scrollbar-width: none;
  }
  .thead-scroll::-webkit-scrollbar {
    display: none;
  }

  /* İçerik gerçek genişliği alsın ki yatay kaydırma mümkün olsun */
  .thead {
    display: grid;
    user-select: none;
    width: max-content; /* <<< kritik */
    min-width: max-content; /* bazı tarayıcılar için güvenlik */
  }

  .th {
    display: flex;
    align-items: center;
    justify-content: space-between;
    padding: 6px 8px;
    border-right: 1px solid #dadce0;
    background: #f1f3f4;
    position: relative;
    white-space: nowrap;
  }
  .th:last-child {
    border-right: none;
  }
  .th__label {
    display: inline-flex;
    align-items: center;
    gap: 6px;
    cursor: pointer;
  }
  .sort-badge {
    font-size: 11px;
    color: #5f6368;
  }
  .th__resizer {
    width: 6px;
    cursor: col-resize;
    align-self: stretch;
    position: absolute;
    right: 0;
    top: 0;
    bottom: 0;
  }

  .tbody-scroll {
    overflow-x: auto;
    overflow-y: hidden;
  }

  .tbody {
    /* grid-template-columns inline geliyor */
  }

  .row {
    display: grid;
    align-items: stretch;
    height: 40px;
    background: #fff;
    width: max-content; /* <<< kritik: satır da gerçek genişlikte olsun */
    min-width: max-content;
  }
  .row:nth-child(even) {
    background: #f8f9fa;
  }
  .row:hover {
    background: #e8f0fe;
  }

  .cell {
    padding: 6px 8px;
    border-right: 1px solid #e0e0e0;
    border-bottom: 1px solid #e0e0e0;
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
    outline: none;
  }
  .cell:last-child {
    border-right: none;
  }

  .cell--active {
    box-shadow: inset 0 0 0 2px #1a73e8;
    background: #eef4ff;
  }
  .cell--editing {
    background: #fffef0;
  }

  .cell-editor {
    width: 100%;
    height: 100%;
    border: none;
    outline: none;
    font: inherit;
    background: #fff;
  }
</style>
