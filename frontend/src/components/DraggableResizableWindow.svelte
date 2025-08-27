<script lang="ts">
  // onMount, onDestroy ve createEventDispatcher'ı svelte kütüphanesinden içe aktarıyoruz.
  import { createEventDispatcher, onMount, onDestroy } from "svelte";
  import interact from "@interactjs/interactjs";
  import type { Interactable } from "@interactjs/core/Interactable";
  import { get } from "svelte/store";
  import { ghostDockZone } from "@/stores/uiState"; // UI durumu mağazasını içe aktarıyoruz

  const dispatch = createEventDispatcher(); // Olay göndericiyi başlatıyoruz.

  // --- Props ---
  export let id: number; // Pencerenin benzersiz kimliği
  export let x: number; // Pencerenin X konumu
  export let y: number; // Pencerenin Y konumu
  export let width: number; // Pencerenin genişliği
  export let height: number; // Pencerenin yüksekliği
  export let title: string = "Pencere"; // Pencere başlığı
  export let active: boolean = false; // Pencerenin aktif olup olmadığı
  export let zIndex: number; // Pencerenin z-indeksi (katman sırası)
  export let minimizable: boolean = false; // Küçültülebilir olup olmadığı
  export let maximizable: boolean = false; // Büyütülebilir olup olmadığı
  export let minimized: boolean = false; // Küçültülmüş olup olmadığı
  export let dockPosition: "left" | "right" | null = null; // Bağlantı konumu (sol, sağ veya yok)

  // Pencerenin görünürlüğünü dışarıdan kontrol etmek için bir prop.
  // Bu prop, üst bileşen (örn: App.svelte) tarafından `bind:isVisible={window.isVisible}` şeklinde kullanılacaktır.
  export let isVisible: boolean = true;

  // --- Dahili Durum ---
  let windowElement: HTMLElement; // Pencere DOM elemanına referans
  let interactable: Interactable; // Interact.js örneği
  let isMaximized: boolean = false; // Pencerenin maksimize edilmiş olup olmadığı
  let resizeStartBounds = { x, y, width, height }; // Boyutlandırma başlangıcındaki sınırlar
  let previousBounds: { x: number; y: number; width: number; height: number } =
    { x, y, width, height }; // Minimize/Maksimize öncesi sınırlar
  // let previousDockBounds: { // Şu an kullanılmıyor, ileride bağlantı konumlarını hatırlamak için kullanılabilir
  //   x: number;
  //   y: number;
  //   width: number;
  //   height: number;
  // } | null = null;

  // --- Olay İşleyicileri ---
  function handleMouseDown(): void {
    // Pencere minimize edilmişse geri yükle
    if (minimized) {
      dispatch("restore", { id: id });
    }
    // Pencereyi aktif hale getir (üst bileşene bildir)
    dispatch("activate");
  }

  // Pencereyi küçültme/geri yükleme işlevi
  function toggleMinimize(): void {
    if (!minimizable) return; // Küçültülebilir değilse fonksiyonu sonlandır

    if (!minimized) {
      // Küçültülüyorsa mevcut sınırları kaydet
      previousBounds = { x, y, width, height };
      isMaximized = false; // Küçültüldüğünde maksimize durumunu sıfırla
      dispatch("minimize", { id: id }); // Üst bileşene küçültme olayını gönder
    } else {
      // Geri yükleniyorsa
      dispatch("restore", { id: id }); // Üst bileşene geri yükleme olayını gönder
    }
  }

  // Pencereyi maksimize etme/geri yükleme işlevi
  function toggleMaximize(): void {
    if (!maximizable) return; // Büyütülebilir değilse fonksiyonu sonlandır

    if (isMaximized) {
      // Maksimize edilmişse önceki boyutlara geri yükle
      dispatch("maximize", {
        id: id,
        x: previousBounds.x,
        y: previousBounds.y,
        width: previousBounds.width,
        height: previousBounds.height,
      });
      isMaximized = false; // Maksimize durumunu sıfırla
    } else {
      // Maksimize ediliyorsa mevcut boyutları kaydet
      previousBounds = { x, y, width, height };
      const parent = windowElement.parentElement;
      if (parent) {
        // Üst elemanın boyutlarını kullanarak maksimize et
        dispatch("maximize", {
          id: id,
          x: 0,
          y: 0,
          width: parent.clientWidth,
          height: parent.clientHeight,
        });
        isMaximized = true; // Maksimize edildi olarak işaretle
      }
    }
  }

  // Başlık çubuğuna çift tıklandığında maksimize etme/geri yükleme işlevi
  function handleHeaderDoubleClick(): void {
    toggleMaximize();
  }

  // PENCEREYİ KAPATMA İŞLEVİ:
  // Bu fonksiyon, pencerenin kendisini kapatmasını sağlar.
  // `isVisible = false;` ile prop'u günceller ve üst bileşene `close` olayını gönderir.
  function closeWindow(): void {
    console.log(`Window ${id} is closing.`);
    isVisible = false; // Pencerenin görünürlüğünü kapatır (bu, üst bileşendeki bind:isVisible'ı da etkiler)
    dispatch("close", { id: id }); // Üst bileşene kapanma olayını bildirir
  }

  // Klavye olaylarını işleme (örn: Escape tuşu ile kapanma)
  function handleKeydown(event: KeyboardEvent): void {
    // ESC tuşuna basıldığında ve pencere aktif, maksimize veya minimize değilse kapat
    if (event.key === "Escape" && active && !isMaximized && !minimized) {
      closeWindow(); // Kapanma fonksiyonunu çağırıyoruz.
    }
  }

  // Interact.js transformasyonlarını sıfırlama işlevi.
  // Bu, pencerenin konumu prop'lar tarafından yönetildiğinde Interact.js'in DOM üzerindeki transformlarını kaldırır.
  function resetInteractPosition(): void {
    if (windowElement) {
      windowElement.style.transform = ""; // CSS transformasyonunu sıfırla
      windowElement.removeAttribute("data-x"); // Interact.js data özniteliklerini kaldır
      windowElement.removeAttribute("data-y");
    }
  }
  export function resetInteractPositionPublic(): void {
    resetInteractPosition();
  }

  // `minimized` prop'u değiştiğinde `isMaximized` durumunu güncelle
  // Bu reaktif ifade, pencere küçültüldüğünde maksimize durumunun da sıfırlanmasını sağlar.
  $: if (minimized) {
    if (isMaximized) {
      previousBounds = { x, y, width, height }; // Eğer maksimize iken küçültülüyorsa mevcut boyutları kaydet
    }
    isMaximized = false; // Küçültüldüğünde maksimize durumunu false yap
  }

  // Svelte prop'ları (x, y, width, height) değiştiğinde Interact.js'in konumunu/boyutunu manuel olarak güncelleme.
  // Bu, `App.svelte` gibi bir üst bileşenden gelen konum/boyut değişikliklerini Interact.js'e bildirir.
  $: if (interactable && windowElement) {
    // Etkileşimlerin (sürükleme ve boyutlandırma) pencerenin durumuna göre etkinleştirilip devre dışı bırakılması
    const enableInteraction = !minimized && !isMaximized;
    interactable.draggable({ enabled: enableInteraction });
    interactable.resizable({ enabled: enableInteraction });

    interactable.styleCursor(false); // Interact.js'in fare imleci stilini yönetmesini devre dışı bırak

    // Pencerenin konumu Svelte prop'ları tarafından yönetildiğinde Interact.js'in kendi transformasyonlarını sıfırla
    resetInteractPosition();
  }

  // Bileşen DOM'a monte edildiğinde (ilk yüklendiğinde) Interact.js'i başlat
  onMount(() => {
    resetInteractPosition(); // Başlangıçta transformu sıfırla

    if (!windowElement) return; // windowElement henüz mevcut değilse (render edilmediyse) çık

    interactable = interact(windowElement)
      .draggable({
        allowFrom: ".window-header", // Sadece başlık çubuğundan sürüklemeye izin ver
        ignoreFrom: ".window-controls", // Kontrol düğmelerini sürükleme dışında tut
        listeners: {
          start(event) {
            // Eğer pencere bağlı konumdaysa (docked), bağlantıyı kes (undock olayını gönder)
            if (dockPosition) {
              dockPosition = null; // Bağlantı konumunu sıfırla
              dispatch("undock", { id }); // Üst bileşene undock olayını gönder
            }
            // Sürükleme başlangıcı olayını üst bileşene gönder
            dispatch("dragstart", {
              id,
              x,
              y,
              mouseX: event.clientX,
              mouseY: event.clientY,
            });
          },
          move(event) {
            // Pencere maksimize veya küçültülmemişse sürüklemeye devam et
            if (!isMaximized && !minimized) {
              const newX = x + event.dx;
              const parent = windowElement?.parentElement;

              if (parent) {
                const parentWidth = parent.clientWidth;
                const threshold = 30; // Kenara yaklaştığında bağlantı (dock) eşiği

                // Bağlantı bölgeleri için hayalet göstergeyi ayarla (sol veya sağ)
                if (newX <= threshold) {
                  ghostDockZone.set("left");
                } else if (newX + width >= parentWidth - threshold) {
                  ghostDockZone.set("right");
                } else {
                  ghostDockZone.set(null); // Bağlantı bölgesinde değilse sıfırla
                }
              }

              // Sürükleme olayını üst bileşene gönder
              dispatch("drag", {
                id,
                x: newX,
                y: y + event.dy,
                mouseX: event.clientX,
                mouseY: event.clientY,
              });
            }
          },
          end(event) {
            // Sürükleme bittiğinde
            if (!isMaximized && !minimized) {
              const parent = windowElement.parentElement;
              if (parent) {
                const parentWidth = parent.clientWidth;
                const ghost = get(ghostDockZone); // Hayalet bağlantı bölgesini store'dan al

                // Eğer bir bağlantı bölgesine bırakıldıysa, dock olayını gönder
                if (ghost === "left" || ghost === "right") {
                  dispatch("dock", {
                    id,
                    x: ghost === "left" ? 0 : parentWidth / 2,
                    y: 0,
                    width: parentWidth / 2,
                    height: parent.clientHeight,
                    position: ghost,
                  });
                  ghostDockZone.set(null); // Bağlantı sonrası hayalet bölgeyi sıfırla
                  return; // Dock işlemi yapıldıysa dragend olayını gönderme
                }
              }

              ghostDockZone.set(null); // Bağlantı yapılmadıysa hayalet bölgeyi temizle
              // Sürükleme bitişi olayını üst bileşene gönder
              dispatch("dragend", {
                id,
                x,
                y,
                mouseX: event.clientX,
                mouseY: event.clientY,
              });
            }
          },
        },
        modifiers: [
          interact.modifiers.restrictRect({
            restriction: "parent", // Pencereyi üst eleman sınırları içinde tut
            endOnly: true, // Sadece sürükleme sonunda kısıtlamayı uygula
          }),
        ],
      })
      .resizable({
        edges: { left: true, right: true, bottom: true, top: true }, // Tüm kenarlardan boyutlandırmaya izin ver
        listeners: {
          start(event) {
            resizeStartBounds = { x, y, width, height }; // Başlangıç boyutlarını kaydet

            dispatch("resizestart", {
              id,
              ...resizeStartBounds,
            });
          },
          move(event) {
            // Pencere maksimize veya küçültülmemişse boyutlandırmaya devam et
            if (!isMaximized && !minimized) {
              const { width: newWidth, height: newHeight } = event.rect;
              const delta = event.deltaRect;

              x += delta.left; // X konumunu güncelle
              y += delta.top; // Y konumunu güncelle

              // DOM stillerini güncelle (Interact.js'in transformu yerine, doğrudan konum ve boyutları ayarlayarak)
              windowElement.style.width = `${newWidth}px`;
              windowElement.style.height = `${newHeight}px`;
              windowElement.style.left = `${x}px`;
              windowElement.style.top = `${y}px`;

              // Boyutlandırma olayını üst bileşene gönder
              dispatch("resize", {
                id,
                width: newWidth,
                height: newHeight,
                x,
                y,
              });
            }
          },
          end(event) {
            // Boyutlandırma bittiğinde
            if (!isMaximized && !minimized) {
              // Pencerenin son DOM boyutlarını ve konumunu oku
              const finalWidth = windowElement.offsetWidth;
              const finalHeight = windowElement.offsetHeight;
              const finalX = parseFloat(windowElement.style.left);
              const finalY = parseFloat(windowElement.style.top);

              // Boyutlandırma bitişi olayını üst bileşene gönder
              dispatch("resizeend", {
                id,
                width: finalWidth,
                height: finalHeight,
                x: finalX,
                y: finalY,
              });
            }
          },
        },
        modifiers: [
          interact.modifiers.restrictSize({
            min: { width: 320, height: 220 }, // Minimum boyutları ayarla
          }),
        ],
      });
  });

  // Bileşen DOM'dan yok edildiğinde Interact.js örneğini temizle (bellek sızıntısını önlemek için)
  onDestroy(() => {
    interactable?.unset(); // Interact.js event dinleyicilerini kaldır
  });
</script>

<svelte:window on:keydown={handleKeydown} />

<!-- YENİ EKLEME: isVisible prop'u false ise tüm pencereyi DOM'dan kaldırır. -->
{#if isVisible}
  <div
    class="window {active ? 'active' : ''} {isMaximized
      ? 'maximized'
      : ''} {minimized ? 'minimized' : ''}"
    bind:this={windowElement}
    style="left: {x}px; top: {y}px; width: {width}px; height: {height}px; z-index: {zIndex};"
    on:mousedown={handleMouseDown}
    role="dialog"
    aria-labelledby="window-title-{title.replace(/\s/g, '-')}-{id}"
    aria-modal="false"
    tabindex="-1"
    data-minimized={minimized}
  >
    <div
      class="window-header"
      on:dblclick={handleHeaderDoubleClick}
      role="button"
      tabindex="0"
      aria-label="Pencere Başlığı, Maksimize etmek veya geri yüklemek için çift tıklayın"
    >
      <span
        class="window-title"
        id="window-title-{title.replace(/\s/g, '-')}-{id}">{title}</span
      >
      <div class="window-controls">
        {#if minimizable}
          <button
            class="control-button minimize"
            on:click={toggleMinimize}
            title="Minimize"
            aria-label="Pencereyi minimize et"
          >
            <i class="fas fa-minus"></i>
          </button>
        {/if}
        {#if maximizable}
          <button
            class="control-button maximize"
            on:click={toggleMaximize}
            title={isMaximized ? "Geri Yükle" : "Maksimize Et"}
            aria-label={isMaximized
              ? "Pencereyi geri yükle"
              : "Pencereyi maksimize et"}
          >
            <i class="fas {isMaximized ? 'fa-compress-alt' : 'fa-expand-alt'}"
            ></i>
          </button>
        {/if}
        <button
          class="control-button close"
          on:click={closeWindow}
          title="Kapat"
          aria-label="Pencereyi kapat"
        >
          &times;
        </button>
      </div>
    </div>
    <div class="window-content">
      <!-- Slot içeriği, doğrudan olay dinleyici eklemiyoruz. Üst bileşen (App.svelte) dinleyecek. -->
      <slot></slot>
    </div>
    {#if !isMaximized && !minimized}
      <!-- Yeniden boyutlandırma tutamaçları sadece maksimize veya minimize edilmemişse gösterilir -->
      <div class="resize-handle top-left"></div>
      <div class="resize-handle top"></div>
      <div class="resize-handle top-right"></div>
      <div class="resize-handle left"></div>
      <div class="resize-handle right"></div>
      <div class="resize-handle bottom-left"></div>
      <div class="resize-handle bottom"></div>
      <div class="resize-handle bottom-right"></div>
    {/if}
  </div>
{/if}

<style>
  @import url("https://fonts.googleapis.com/css2?family=Inter:wght@300;400;500;600;700&display=swap");

  /* Kurumsal renkler ve geçişler için CSS değişkenleri */
  :root {
    --corporate-light-gray: #f0f2f5;
    --corporate-medium-gray: #d1d9e0;
    --corporate-dark-gray: #8894a4;
    --corporate-text-color: #333d47;
    --corporate-header-bg: #e5e8ec;
    --corporate-accent-blue: #007bff;
    --corporate-red: #dc3545;
    --corporate-shadow: 0 4px 15px rgba(0, 0, 0, 0.1);
    --corporate-transition-speed: 0.2s;
    --corporate-transition-timing: ease-out;
  }

  /* Ana pencere stil tanımları */
  .window {
    position: absolute; /* Üst elemana göre konumlandır */
    background-color: var(--corporate-light-gray); /* Arka plan rengi */
    border: 1px solid var(--corporate-medium-gray); /* Kenarlık */
    box-shadow: var(--corporate-shadow); /* Gölge efekti */
    display: flex; /* İçeriği esnek kutu olarak düzenle */
    flex-direction: column; /* İçeriği dikey olarak hizala */
    overflow: hidden; /* İçerik taşmasını gizle */
    border-radius: 4px; /* Köşeleri yuvarla */
    font-family: "Inter", sans-serif; /* Yazı tipi */
    color: var(--corporate-text-color); /* Yazı rengi */
    min-width: 320px; /* Minimum genişlik */
    min-height: 220px; /* Minimum yükseklik */
    outline: none; /* Odaklandığında dış çizgiyi kaldır */
    will-change: transform, top, left, width, height; /* Performans ipucu: bu özellikler değişecek */
    transition:
      transform var(--corporate-transition-speed)
        var(--corporate-transition-timing),
      opacity var(--corporate-transition-speed)
        var(--corporate-transition-timing); /* Konum/şeffaflık için geçiş animasyonları */
    touch-action: none; /* Dokunmatik olayları devre dışı bırak (Interact.js için) */
    box-sizing: border-box; /* Kutu modelini ayarla */
  }

  /* Aktif pencere stil tanımları */
  .window.active {
    border-color: var(--corporate-dark-gray); /* Aktif kenarlık rengi */
    box-shadow:
      0 6px 20px rgba(0, 0, 0, 0.15),
      0 2px 10px rgba(0, 0, 0, 0.05); /* Aktif gölge efekti */
  }

  /* Maksimize edilmiş pencere stil tanımları */
  .window.maximized {
    top: 0 !important; /* Üstten sıfırla */
    left: 0 !important; /* Soldan sıfırla */
    width: 100% !important; /* Tam genişlik */
    height: 100% !important; /* Tam yükseklik */
    border-radius: 0; /* Köşeleri düzelt */
    box-shadow: none; /* Gölgeyi kaldır */
    transition: none; /* Geçişleri devre dışı bırak */
  }

  /* Küçültülmüş pencere stil tanımları */
  .window.minimized {
    transform: translateY(100vh); /* Ekranın altına kaydır */
    opacity: 0; /* Şeffaflaştır */
    pointer-events: none; /* Fare olaylarını engelle */
  }

  /* Pencere başlığı stil tanımları */
  .window-header {
    background-color: var(--corporate-header-bg); /* Arka plan rengi */
    border-bottom: 1px solid var(--corporate-medium-gray); /* Alt kenarlık */
    color: var(--corporate-text-color); /* Yazı rengi */
    padding: 0px 15px; /* İç boşluk */
    cursor: grab; /* Sürükleme imleci */
    display: flex; /* İçeriği esnek kutu olarak düzenle */
    justify-content: space-between; /* İçeriği yatayda yay */
    align-items: center; /* İçeriği dikeyde ortala */
    user-select: none; /* Metin seçimini engelle */
    font-size: 15px; /* Yazı tipi boyutu */
    font-weight: 500; /* Yazı tipi kalınlığı */
    flex-shrink: 0; /* Küçülmesini engelle (her zaman sabit boyutta kalır) */
  }

  /* Pencere başlığı aktifken sürükleme imleci */
  .window-header:active {
    cursor: grabbing; /* Sürükleme anında imleç */
  }

  /* Pencere başlığı metni stil tanımları */
  .window-title {
    flex-grow: 1; /* Kalan alanı kapla */
    padding-right: 15px; /* Sağdan iç boşluk */
    white-space: nowrap; /* Metnin tek satırda kalmasını sağla */
    overflow: hidden; /* Taşmayı gizle */
    text-overflow: ellipsis; /* Taşmayı üç nokta ile göster */
  }

  /* Pencere kontrol düğmeleri grubu stil tanımları */
  .window-controls {
    display: flex; /* İçeriği esnek kutu olarak düzenle */
    gap: 6px; /* Düğmeler arasında boşluk */
  }

  /* Kontrol düğmeleri genel stil tanımları */
  .control-button {
    background: none; /* Arka plan yok */
    border: none; /* Kenarlık yok */
    color: var(--corporate-dark-gray); /* Yazı rengi */
    font-size: 0.8rem; /* Yazı tipi boyutu */
    cursor: pointer; /* Fare imleci */
    width: 30px; /* Genişlik */
    height: 30px; /* Yükseklik */
    display: flex; /* İçeriği esnek kutu olarak düzenle */
    justify-content: center; /* İçeriği yatayda ortala */
    align-items: center; /* İçeriği dikeyde ortala */
    border-radius: 3px; /* Köşeleri yuvarla */
    transition:
      background-color var(--corporate-transition-speed)
        var(--corporate-transition-timing),
      color var(--corporate-transition-speed) var(--corporate-transition-timing); /* Renk ve arka plan değişimi için geçiş animasyonları */
    outline: none; /* Odaklandığında dış çizgiyi kaldır */
  }

  /* Kontrol düğmeleri üzerine gelindiğinde stil tanımları */
  .control-button:hover {
    background-color: var(--corporate-medium-gray); /* Arka plan rengi */
    color: var(--corporate-accent-blue); /* Yazı rengi */
    transform: scale(1); /* Ölçeklendirme (efekt için) */
  }

  /* Kapatma düğmesi özel stil tanımları */
  .control-button.close {
    font-weight: 400; /* Yazı tipi kalınlığı */
    font-size: 1.2rem; /* Yazı tipi boyutu */
  }

  /* Kapatma düğmesi üzerine gelindiğinde özel stil tanımları */
  .control-button.close:hover {
    background-color: var(--corporate-red); /* Kırmızı arka plan */
    color: var(--corporate-light-gray); /* Yazı rengi */
  }

  /* Küçültme ve maksimize etme düğmeleri üzerine gelindiğinde özel stil tanımları */
  .control-button.minimize:hover,
  .control-button.maximize:hover {
    background-color: var(--corporate-accent-blue); /* Mavi arka plan */
    color: var(--corporate-light-gray); /* Yazı rengi */
  }

  /* Pencere içeriği alanı stil tanımları */
  .window-content {
    flex-grow: 1; /* Kalan dikey alanı kapla */
    overflow: hidden; /* İçerik taşmasını gizle */
    background-color: var(--corporate-light-gray); /* Arka plan rengi */
    min-height: 0; /* Kritik: İçeriğin esnek kapsayıcıyı aşmasını ve uzamasını önler */
  }

  /* Boyutlandırma tutamaçları (varsayılan olarak şeffaf, üzerine gelindiğinde görünür) */
  .resize-handle {
    position: absolute; /* Pencereye göre konumlandır */
    width: 8px; /* Genişlik */
    height: 8px; /* Yükseklik */
    background-color: transparent; /* Şeffaf arka plan */
    border: none; /* Kenarlık yok */
    z-index: 2; /* Diğer elemanların üzerinde olmasını sağla */
    opacity: 0; /* Varsayılan olarak gizli */
  }

  /* Pencere üzerine gelindiğinde boyutlandırma tutamaçlarını göster */
  .window:hover .resize-handle {
    opacity: 1; /* Görünür yap */
  }

  /* Köşe boyutlandırma tutamaçları için imleçler */
  .resize-handle.top-left {
    top: 0;
    left: 0;
    cursor: nwse-resize; /* Köşe imleci */
  }
  .resize-handle.top-right {
    top: 0;
    right: 0;
    cursor: nesw-resize; /* Köşe imleci */
  }
  .resize-handle.bottom-left {
    bottom: 0;
    left: 0;
    cursor: nesw-resize; /* Köşe imleci */
  }
  .resize-handle.bottom-right {
    bottom: 0;
    right: 0;
    cursor: nwse-resize; /* Köşe imleci */
  }

  /* Kenar boyutlandırma tutamaçları için imleçler */
  .resize-handle.top {
    top: 0;
    left: 8px;
    right: 8px;
    width: calc(100% - 16px); /* Kenar boşluklarını dikkate al */
    cursor: ns-resize; /* Dikey imleç */
  }
  .resize-handle.bottom {
    bottom: 0;
    left: 8px;
    right: 8px;
    width: calc(100% - 16px);
    cursor: ns-resize; /* Dikey imleç */
  }
  .resize-handle.left {
    top: 8px;
    bottom: 8px;
    left: 0;
    height: calc(100% - 16px); /* Kenar boşluklarını dikkate al */
    cursor: ew-resize; /* Yatay imleç */
  }
  .resize-handle.right {
    top: 8px;
    bottom: 8px;
    right: 0;
    height: calc(100% - 16px);
    cursor: ew-resize; /* Yatay imleç */
  }
</style>
