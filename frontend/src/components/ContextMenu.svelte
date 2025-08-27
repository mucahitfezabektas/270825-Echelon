<script lang="ts">
  import { contextMenuStore, hideContextMenu } from "@/stores/contextMenuStore";
  import type { ContextMenuItem } from "@/stores/contextMenuStore";
  import { onMount, onDestroy } from "svelte";

  let menuElement: HTMLElement;
  let ignoreNextMouseDown = false; // Menü açıldığında bir sonraki mousedown'ı yoksaymak için bayrak

  function handleItemClick(item: ContextMenuItem, data: any) {
    console.log("ContextMenu: Menü öğesi tıklandı. Kapatılıyor.");
    if (!item.disabled && item.action) {
      item.action(data);
      hideContextMenu(); // Seçenek tıklandığında menüyü gizle
    }
  }

  // Menü açıkken dışarıdaki herhangi bir mousedown (sol tuş) olayında menüyü kapat
  function handleMouseDownOutside(event: MouseEvent) {
    // Sadece sol mousedown olaylarını dinle (sağ tıklama hariç)
    if (event.button === 0) {
      if (ignoreNextMouseDown) {
        console.log(
          "ContextMenu: İlk mousedown yoksayıldı (menü açılışından sonra)."
        );
        ignoreNextMouseDown = false; // Bir sonraki için sıfırla
        return;
      }

      // Menü elementinin varlığını ve olayın menü dışında gerçekleştiğini kontrol et
      if (menuElement && !menuElement.contains(event.target as Node)) {
        console.log("ContextMenu: Menü dışına sol tıklandı. Kapatılıyor.");
        hideContextMenu();
      }
    }
  }

  function handleKeyDown(event: KeyboardEvent) {
    if (event.key === "Escape") {
      console.log("ContextMenu: ESC tuşuna basıldı. Kapatılıyor.");
      hideContextMenu();
    }
  }

  // Store'dan gelen veriyi dinle
  $: ({ show, x, y, items, data } = $contextMenuStore);

  // 'show' durumu değiştiğinde ignoreNextMouseDown bayrağını yönet
  $: {
    if (show) {
      console.log(
        "ContextMenu: 'show' true oldu. ignoreNextMouseDown bayrağı ayarlandı."
      );
      // Menü gösterildiğinde, ignoreNextMouseDown'ı true yap.
      // Bu, menüyü açan sağ tıklamayı takip eden olası mousedown olayını yoksaymamızı sağlar.
      ignoreNextMouseDown = true;
    } else {
      console.log(
        "ContextMenu: 'show' false oldu. ignoreNextMouseDown sıfırlandı."
      );
      // Menü kapandığında, bayrağı sıfırla.
      ignoreNextMouseDown = false;
    }
  }

  onMount(() => {
    console.log("ContextMenu: Mounted. Olay dinleyicileri ekleniyor.");
    // `document` üzerinde `mousedown` olayını yakalama aşamasında dinle (`true` parametresi)
    document.addEventListener("mousedown", handleMouseDownOutside, true);
    window.addEventListener("keydown", handleKeyDown);
  });

  onDestroy(() => {
    console.log("ContextMenu: Destroyed. Olay dinleyicileri kaldırılıyor.");
    document.removeEventListener("mousedown", handleMouseDownOutside, true);
    window.removeEventListener("keydown", handleKeyDown);
  });

  $: adjustedX = x;
  $: adjustedY = y;

  $: if (menuElement && show) {
    const menuRect = menuElement.getBoundingClientRect();
    const viewportWidth = window.innerWidth;
    const viewportHeight = window.innerHeight;

    if (x + menuRect.width > viewportWidth - 10) {
      adjustedX = viewportWidth - menuRect.width - 10;
    }
    if (y + menuRect.height > viewportHeight - 10) {
      adjustedY = viewportHeight - menuRect.height - 10;
    }
    adjustedX = Math.max(0, adjustedX);
    adjustedY = Math.max(0, adjustedY);
  }
</script>

{#if show}
  <!-- svelte-ignore a11y_no_static_element_interactions -->
  <div
    bind:this={menuElement}
    class="context-menu-popup"
    style="left: {adjustedX}px; top: {adjustedY}px;"
    on:contextmenu|preventDefault|stopPropagation
    on:mousedown|stopPropagation={(e) => {
      console.log("ContextMenu: Menü popup'ına mousedown. Yayılım durduruldu.");
    }}
  >
    <ul>
      {#each items as item (item.label || (item.separator ? "separator-" + Math.random() : ""))}
        {#if item.separator}
          <hr />
        {:else}
          <li>
            <button
              on:click={() => handleItemClick(item, data)}
              disabled={item.disabled}
              class:disabled={item.disabled}
            >
              {item.label}
            </button>
          </li>
        {/if}
      {/each}
    </ul>
  </div>
{/if}

<style>
  .context-menu-popup {
    position: fixed;
    background: #f3f3f3; /* Açık gri arka plan */
    border: 1px solid #e0e0e0; /* İnce beyazımsı gri kenarlık */
    box-shadow: 0 2px 8px rgba(0, 0, 0, 0.08);
    border-radius: 8px;
    padding: 6px 0;
    min-width: 200px;
    z-index: 10000;
    cursor: default;
    user-select: none;
    display: flex;
    flex-direction: column;
    font-family: "Segoe UI", "Roboto", sans-serif;
    font-size: 14px;
    color: #2c2c2c;
    pointer-events: auto;
  }

  .context-menu-popup ul {
    list-style: none;
    padding: 0;
    margin: 0;
  }

  .context-menu-popup ul li {
    margin: 0;
    padding: 0;
  }

  .context-menu-popup ul li button {
    display: block;
    width: 100%;
    background: none;
    border: none;
    padding: 10px 16px;
    text-align: left;
    font-size: 14px;
    font-weight: 500;
    color: #2c2c2c;
    border-radius: 4px;
    cursor: pointer;
    transition: background-color 0.15s ease;
  }

  .context-menu-popup ul li button:hover:not(.disabled) {
    background-color: #e5e5e5; /* Daha koyu gri hover */
  }

  .context-menu-popup ul li button.disabled {
    color: #a0a0a0;
    cursor: not-allowed;
    opacity: 0.6;
  }

  .context-menu-popup hr {
    border: none;
    border-top: 1px solid #dddddd;
    margin: 8px 12px;
  }
</style>
