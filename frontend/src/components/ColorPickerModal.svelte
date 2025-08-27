<script lang="ts">
  import {
    createEventDispatcher,
    onMount,
    onDestroy,
    afterUpdate,
  } from "svelte";

  export let showModal: boolean = false;
  export let currentColor: string; // Modal'a gönderilen mevcut renk
  export let activityCode: string | undefined; // Hangi aktivitenin rengi değişiyor

  let newColor: string; // Kullanıcının seçtiği yeni renk (geçici)
  let modalBackdropRef: HTMLDivElement;
  let previouslyFocusedElement: HTMLElement | null = null;

  // currentColor her değiştiğinde newColor'ı güncelle
  // Bu, modal açıldığında başlangıç rengini doğru ayarlamayı sağlar
  $: newColor = currentColor;

  // Eventleri parent bileşene göndermek için dispatcher
  const dispatch = createEventDispatcher();

  function handleSave(): void {
    dispatch("save", { code: activityCode, color: newColor });
    showModal = false; // Modalı kapat
  }

  function handleCancel(): void {
    dispatch("cancel");
    showModal = false; // Modalı kapat
  }

  // Global keydown listener for Escape key
  function handleGlobalKeyDown(event: KeyboardEvent): void {
    if (event.key === "Escape" && showModal) {
      handleCancel();
    }
  }

  // Handle keyboard events on the backdrop itself (for accessibility linter)
  function handleBackdropKeyDown(event: KeyboardEvent): void {
    // Only if the backdrop itself is focused and a relevant key is pressed
    if (event.key === "Enter" || event.key === " ") {
      event.preventDefault(); // Prevent default scroll behavior for space
      handleCancel();
    }
  }

  // --- Focus Management ---
  // Use afterUpdate to ensure the element exists before trying to focus
  // when showModal becomes true.
  afterUpdate(() => {
    if (
      showModal &&
      modalBackdropRef &&
      document.activeElement !== modalBackdropRef
    ) {
      // Store the previously focused element only when the modal opens for the first time
      // in this cycle.
      if (!previouslyFocusedElement) {
        previouslyFocusedElement = document.activeElement as HTMLElement;
      }
      modalBackdropRef.focus();
    } else if (!showModal && previouslyFocusedElement) {
      // When modal closes, return focus to the previously focused element
      previouslyFocusedElement.focus();
      previouslyFocusedElement = null; // Clear the reference
    }
  });

  // Add global keydown listener on mount, remove on destroy
  onMount(() => {
    document.addEventListener("keydown", handleGlobalKeyDown);
  });

  onDestroy(() => {
    document.removeEventListener("keydown", handleGlobalKeyDown);
  });
</script>

{#if showModal}
  <div
    class="modal-backdrop"
    role="dialog"
    aria-modal="true"
    aria-labelledby="modal-title"
    tabindex="-1"
    on:click={handleCancel}
    on:keydown={handleBackdropKeyDown}
    bind:this={modalBackdropRef}
  >
    <div class="modal-content" on:click|stopPropagation role="none">
      <div class="modal-header">
        <h2 id="modal-title">
          {activityCode === "UNK"
            ? "Varsayılan Rengi Değiştir"
            : `Rengi Değiştir: ${activityCode || "Belirsiz Aktivite"}`}
        </h2>
        <button class="close-button" on:click={handleCancel} aria-label="Kapat"
          >&times;</button
        >
      </div>
      <div class="modal-body">
        <div class="color-preview" style="background-color: {newColor};"></div>
        <input type="color" bind:value={newColor} />
        <span class="current-hex-code"
          >{newColor ? newColor.toUpperCase() : ""}</span
        >
      </div>
      <div class="modal-footer">
        <button class="cancel-btn" on:click={handleCancel}>İptal</button>
        <button class="save-btn" on:click={handleSave}>Kaydet</button>
      </div>
    </div>
  </div>
{/if}

<style>
  /* Mevcut stilleriniz aynı kalabilir */
  .modal-backdrop {
    position: absolute;
    top: 0;
    left: 0;
    width: 100%;
    height: 100%;
    background-color: rgba(0, 0, 0, 0.5);
    display: flex;
    justify-content: center;
    align-items: center;
    z-index: 1000;
    /* Klavye odaklanması için */
    outline: none;
  }

  .modal-content {
    background-color: #fff;
    padding: 20px;
    border-radius: 8px;
    box-shadow: 0 4px 15px rgba(0, 0, 0, 0.2);
    min-width: 300px;
    max-width: 90%;
    box-sizing: border-box;
    display: flex;
    flex-direction: column;
    gap: 15px;
  }

  .modal-header {
    display: flex;
    justify-content: space-between;
    align-items: center;
    border-bottom: 1px solid #eee;
    padding-bottom: 10px;
    margin-bottom: 10px;
  }

  .modal-header h2 {
    margin: 0;
    font-size: 1.2em;
    color: #333;
  }

  .close-button {
    background: none;
    border: none;
    font-size: 1.5em;
    cursor: pointer;
    color: #888;
    padding: 0 5px;
  }

  .close-button:hover {
    color: #333;
  }

  .modal-body {
    display: flex;
    flex-direction: column;
    align-items: center;
    gap: 15px;
  }

  .color-preview {
    width: 100px;
    height: 50px;
    border: 1px solid #ccc;
    border-radius: 4px;
  }

  .modal-body input[type="color"] {
    width: 100%; /* Genişliği artır */
    height: 40px; /* Yüksekliği artır */
    border: none;
    background: none;
    cursor: pointer;
  }

  .current-hex-code {
    font-family: monospace;
    font-size: 1.1em;
    color: #555;
  }

  .modal-footer {
    display: flex;
    justify-content: flex-end;
    gap: 10px;
    border-top: 1px solid #eee;
    padding-top: 15px;
    margin-top: 5px;
  }

  .cancel-btn,
  .save-btn {
    padding: 8px 15px;
    border-radius: 5px;
    cursor: pointer;
    font-weight: bold;
    transition: background-color 0.2s ease;
  }

  .cancel-btn {
    background-color: #f0f0f0;
    color: #555;
    border: 1px solid #ddd;
  }

  .cancel-btn:hover {
    background-color: #e0e0e0;
  }

  .save-btn {
    background-color: #007acc;
    color: white;
    border: 1px solid #007acc;
  }

  .save-btn:hover {
    background-color: #005f99;
  }
</style>
