<script lang="ts">
  import { onDestroy, tick } from "svelte";
  import {
    activityColorStore,
    getActivityColor,
    updateActivityColor,
    updateDefaultActivityColor,
    dutyTypes,
    dutyCodes,
    otherActivities,
    generalCategories,
  } from "@/stores/colorStore";
  import ColorPickerModal from "@/components/ColorPickerModal.svelte"; // Yolu kontrol edin!

  // State değişkenleri
  let activityColors: { [key: string]: string };
  let defaultActivityColor: string;

  // Hangi renk kutusunun seçili olduğunu tutan değişken
  let selectedColorCode: string | undefined = undefined;
  // Modal'ın görünürlüğünü kontrol eden değişken
  let showColorPickerModal: boolean = false;
  // Modal'a gönderilecek mevcut renk
  let colorForModal: string = "#000000"; // Başlangıçta varsayılan bir renk

  // Arama ve vurgulama ile ilgili state'ler
  let searchTerm: string = ""; // Arama metni
  let highlightedCode: string | undefined = undefined; // Kalıcı olarak vurgulanacak kod
  let blinkingCode: string | undefined = undefined; // O anda yanıp sönen kod
  let blinkTimer: NodeJS.Timeout | null = null; // Yanıp sönme zamanlayıcısı

  // DOM referansları
  let colorItemRefs: { [key: string]: HTMLElement } = {}; // Renk öğelerinin DOM referansları

  // Store'a abone ol
  const unsubscribe = activityColorStore.subscribe((value) => {
    activityColors = value.colors;
    defaultActivityColor = value.defaultColor;
  });

  // Renk kutusuna tıklandığında tetiklenecek fonksiyon
  function handleColorBoxClick(
    code: string | undefined,
    event: MouseEvent
  ): void {
    selectedColorCode = code;
    colorForModal = getActivityColor(code); // Mevcut rengi al
    showColorPickerModal = true; // Modalı aç
  }

  // Modal'dan "Kaydet" eventi geldiğinde tetiklenecek fonksiyon
  function handleColorSave(
    event: CustomEvent<{ code: string | undefined; color: string }>
  ): void {
    const { code, color } = event.detail;
    if (code === "UNK") {
      updateDefaultActivityColor(color);
    } else if (code) {
      updateActivityColor(code, color);
    }
    showColorPickerModal = false; // Modalı kapatmayı unutmayın!
    selectedColorCode = undefined; // Seçimi temizle
  }

  // Modal'dan "İptal" eventi geldiğinde tetiklenecek fonksiyon
  function handleColorCancel(): void {
    showColorPickerModal = false; // Modalı kapatmayı unutmayın!
    selectedColorCode = undefined; // Seçimi temizle
  }

  // Arama butonuna basıldığında veya Enter'a basıldığında tetiklenecek fonksiyon
  async function handleSearch() {
    // Önceki zamanlayıcıları ve vurguları temizle
    if (blinkTimer) {
      clearInterval(blinkTimer);
      blinkTimer = null;
    }
    blinkingCode = undefined;
    highlightedCode = undefined;

    if (!searchTerm) {
      return; // Arama terimi yoksa bir şey yapma
    }

    const lowerCaseSearchTerm = searchTerm.toLowerCase();

    // Tüm kategorilerde arama yap ve ilk eşleşmeyi bul
    const allCodes = [
      ...dutyTypes,
      ...dutyCodes,
      ...otherActivities,
      ...generalCategories.map((item) => item.code),
      "UNK", // Varsayılanı da ara
    ];

    let foundCode: string | undefined;
    for (const code of allCodes) {
      const generalItem = generalCategories.find((item) => item.code === code);
      if (
        generalItem &&
        generalItem.name.toLowerCase().includes(lowerCaseSearchTerm)
      ) {
        foundCode = code;
        break;
      }
      if (code.toLowerCase().includes(lowerCaseSearchTerm)) {
        foundCode = code;
        break;
      }
    }

    if (foundCode) {
      blinkingCode = foundCode; // Yanıp sönmek üzere ayarla

      await tick(); // DOM'un güncellenmesini bekle

      const targetElement = colorItemRefs[foundCode];
      if (targetElement) {
        targetElement.scrollIntoView({
          behavior: "smooth",
          block: "center",
        });

        let blinkCount = 0;
        const maxBlinks = 6; // 3 saniye boyunca 0.5 saniyede bir (açık/kapalı) -> 6 geçiş

        // Blink efektini başlat
        blinkTimer = setInterval(() => {
          if (blinkingCode === foundCode) {
            // Yanıp sönen kod hala bu ise
            blinkingCode = undefined; // Kapat
          } else {
            blinkingCode = foundCode; // Aç
          }
          blinkCount++;

          if (blinkCount >= maxBlinks) {
            clearInterval(blinkTimer as NodeJS.Timeout);
            blinkTimer = null;
            blinkingCode = undefined; // Blink'i tamamen durdur
            highlightedCode = foundCode; // Kalıcı vurguyu başlat
          }
        }, 500); // Her 0.5 saniyede bir durumu değiştir
      }
    } else {
      console.log("Eşleşme bulunamadı.");
    }
  }

  // Enter tuşuna basıldığında aramayı tetikle
  function handleKeyDown(event: KeyboardEvent) {
    if (event.key === "Enter") {
      handleSearch();
    }
  }

  // searchTerm değiştiğinde (temizlendiğinde) vurguyu kaldır
  $: if (!searchTerm && (highlightedCode || blinkingCode)) {
    highlightedCode = undefined;
    blinkingCode = undefined;
    if (blinkTimer) {
      clearInterval(blinkTimer);
      blinkTimer = null;
    }
  }

  onDestroy(() => {
    unsubscribe();
    if (blinkTimer) {
      clearInterval(blinkTimer);
    }
  });
</script>

<div class="color-sampler-content-inner">
  <div class="search-controls">
    <input
      type="text"
      placeholder="Kodu veya adı ara..."
      bind:value={searchTerm}
      on:keydown={handleKeyDown}
      class="search-input"
      aria-label="Renk kodu veya adıyla ara"
    />
    <button on:click={handleSearch} class="search-button" aria-label="Ara"
      >Ara</button
    >
  </div>

  <div class="category-section">
    <div class="category-header">Duty Type</div>
    <div class="category-list">
      {#each dutyTypes as code}
        <div
          class="color-item"
          bind:this={colorItemRefs[code]}
          class:blinking={blinkingCode === code}
          class:highlighted={highlightedCode === code && searchTerm !== ""}
        >
          <span class="code-text">{code}</span>
          <button
            type="button"
            class="color-box"
            style="background-color: {activityColors[code] || '#000'};"
            on:click={(e) => handleColorBoxClick(code, e)}
            title="Rengi değiştir: {code}"
            aria-label="Rengi {code} için değiştir"
          ></button>
        </div>
      {/each}
    </div>
  </div>

  <div class="category-section">
    <div class="category-header">Duty Code</div>
    <div class="category-list">
      {#each dutyCodes as code}
        <div
          class="color-item"
          bind:this={colorItemRefs[code]}
          class:blinking={blinkingCode === code}
          class:highlighted={highlightedCode === code && searchTerm !== ""}
        >
          <span class="code-text">{code}</span>
          <button
            type="button"
            class="color-box"
            style="background-color: {activityColors[code] || '#000'};"
            on:click={(e) => handleColorBoxClick(code, e)}
            title="Rengi değiştir: {code}"
            aria-label="Rengi {code} için değiştir"
          ></button>
        </div>
      {/each}
    </div>
  </div>

  <div class="category-section">
    <div class="category-header">Other</div>
    <div class="category-list">
      {#each otherActivities as code}
        <div
          class="color-item"
          bind:this={colorItemRefs[code]}
          class:blinking={blinkingCode === code}
          class:highlighted={highlightedCode === code && searchTerm !== ""}
        >
          <span class="code-text">{code}</span>
          <button
            type="button"
            class="color-box"
            style="background-color: {activityColors[code] || '#000'};"
            on:click={(e) => handleColorBoxClick(code, e)}
            title="Rengi değiştir: {code}"
            aria-label="Rengi {code} için değiştir"
          ></button>
        </div>
      {/each}
    </div>
  </div>

  <div class="category-section">
    <div class="category-header">General</div>
    <div class="category-list">
      {#each generalCategories as item}
        <div
          class="color-item"
          bind:this={colorItemRefs[item.code]}
          class:blinking={blinkingCode === item.code}
          class:highlighted={highlightedCode === item.code && searchTerm !== ""}
        >
          <span class="code-text">{item.name}</span>
          <button
            type="button"
            class="color-box"
            style="background-color: {activityColors[item.code] || '#000'};"
            on:click={(e) => handleColorBoxClick(item.code, e)}
            title="Rengi değiştir: {item.name}"
            aria-label="Rengi {item.name} için değiştir"
          ></button>
        </div>
      {/each}
      <div
        class="color-item"
        bind:this={colorItemRefs["UNK"]}
        class:blinking={blinkingCode === "UNK"}
        class:highlighted={highlightedCode === "UNK" && searchTerm !== ""}
      >
        <span class="code-text">Unknown (Default)</span>
        <button
          type="button"
          class="color-box"
          style="background-color: {defaultActivityColor};"
          on:click={(e) => handleColorBoxClick("UNK", e)}
          title="Varsayılan rengi değiştir"
          aria-label="Varsayılan rengi değiştir"
        ></button>
      </div>
    </div>
  </div>
</div>

<ColorPickerModal
  bind:showModal={showColorPickerModal}
  bind:currentColor={colorForModal}
  activityCode={selectedColorCode}
  on:save={handleColorSave}
  on:cancel={handleColorCancel}
/>

<style>
  /* Arama kontrol alanı için yeni stiller */
  .search-controls {
    display: flex;
    gap: 10px;
    padding: 10px;
    background-color: #f9f9f9;
    border-bottom: 1px solid #e0e0e0;
    flex-shrink: 0;
    align-items: center;
  }

  .search-input {
    flex-grow: 1; /* Mevcut alanı kapla */
    padding: 8px 12px;
    border: 1px solid #ccc;
    border-radius: 4px;
    font-size: 14px;
    box-sizing: border-box;
  }

  .search-input::placeholder {
    color: #888;
  }

  .search-button {
    padding: 8px 15px;
    border: 1px solid #007bff;
    border-radius: 4px;
    background-color: #007bff;
    color: white;
    cursor: pointer;
    font-size: 14px;
    transition:
      background-color 0.2s,
      border-color 0.2s;
  }

  .search-button:hover {
    background-color: #0056b3;
    border-color: #0056b3;
  }

  /* Vurgulama stili */
  .color-item.highlighted {
    background-color: #fcf8e3; /* Çok açık, yumuşak bir bej/sarı tonu */
    /* border: 1px solid #ffbf00;
    box-shadow: 0 0 5px rgba(255, 187, 0, 0.5); */
    padding: 3px 0; /* Boşluk ayarı */
  }

  /* Yanıp sönme stili */
  .color-item.blinking {
    animation: blink-animation 0.7s infinite alternate ease-in-out; /* Daha yavaş ve akıcı yanıp sönme */
    background-color: #fff9c4; /* Daha pastel bir sarı tonu */
    border: 1px solid #ffecb3; /* Açık sarı kenarlık */
  }

  @keyframes blink-animation {
    from {
      opacity: 1;
      transform: scale(1);
    }
    to {
      opacity: 0.9;
    }
  }

  /* Mevcut stilleriniz */
  .color-sampler-content-inner {
    padding: 10px;
    height: 100%;
    overflow-y: auto;
    display: flex;
    flex-direction: column;
    gap: 15px;
    box-sizing: border-box;
  }

  .category-section {
    background-color: #ffffff;
    border: 1px solid #e0e0e0;
    border-radius: 3px;
    overflow: hidden;
    flex-shrink: 0;
  }

  .category-header {
    background-color: #f0f0f0;
    padding: 8px 10px;
    font-weight: bold;
    border-bottom: 1px solid #e0e0e0;
    color: #444;
  }

  .category-list {
    padding: 5px 10px;
  }

  .color-item {
    display: flex;
    justify-content: space-between;
    align-items: center;
    padding: 4px 0;
    border-bottom: 1px dashed #f0f0f0;
    transition:
      background-color 0.3s ease,
      border 0.3s ease,
      box-shadow 0.3s ease,
      opacity 0.3s ease; /* Opasite geçişi eklendi */
  }

  .color-item:last-child {
    border-bottom: none;
  }

  .code-text {
    font-size: 12px;
    color: #333;
  }

  .color-box {
    width: 80px;
    height: 18px;
    border: 1px solid #c0c0c0;
    border-radius: 2px;
    flex-shrink: 0;
    cursor: pointer;
    transition: transform 0.1s ease-out;
    background: none;
    padding: 0;
    margin: 0;
    box-sizing: border-box;
  }

  .color-box:hover {
    transform: scale(1.05);
  }
</style>
