<script lang="ts">
  import { createEventDispatcher, onMount, onDestroy } from "svelte";

  export let isOpen: boolean = false;
  export let message: string = "";
  export let title: string = "Bilgilendirme"; // Default title for info/success
  export let duration: number = 3000; // Otomatik kapanma süresi (ms cinsinden)
  export let type: "info" | "success" | "warning" | "error" = "info"; // Bildirim tipi

  const dispatch = createEventDispatcher();
  let timeoutId: ReturnType<typeof setTimeout> | null = null; // Timeout ID'sini tutmak için

  // isOpen değiştiğinde veya bileşen mount edildiğinde çalışır
  $: if (isOpen) {
    startAutoClose();
  } else {
    clearAutoClose();
  }

  onMount(() => {
    // Bileşen mount edildiğinde de otomatik kapanmayı başlat
    if (isOpen) {
      startAutoClose();
    }
  });

  onDestroy(() => {
    // Bileşen yok edildiğinde timeout'u temizle
    clearAutoClose();
  });

  function startAutoClose() {
    clearAutoClose(); // Mevcut timeout varsa temizle
    if (duration > 0) {
      timeoutId = setTimeout(() => {
        close();
      }, duration);
    }
  }

  function clearAutoClose() {
    if (timeoutId) {
      clearTimeout(timeoutId);
      timeoutId = null;
    }
  }

  function close() {
    isOpen = false;
    clearAutoClose(); // Kapatılırken de timeout'u temizle
    dispatch("close");
  }

  // Tipine göre ikon veya stil belirleme
  function getTypeClass() {
    switch (type) {
      case "success":
        return "modal-success";
      case "warning":
        return "modal-warning";
      case "error":
        return "modal-error";
      default:
        return "modal-info";
    }
  }

  function getTypeIcon() {
    switch (type) {
      case "success":
        return "✔️"; // Unicode check mark
      case "warning":
        return "⚠️"; // Unicode warning sign
      case "error":
        return "❌"; // Unicode cross mark
      default:
        return "ℹ️"; // Unicode info sign
    }
  }
</script>

{#if isOpen}
  <div class="modal-backdrop-toast" on:click|self={close}></div>
  <div class="modal-container-toast {getTypeClass()}">
    <div class="modal-header-toast">
      <span class="modal-icon">{getTypeIcon()}</span>
      <h2>{title}</h2>
      <button class="close-button-toast" on:click={close}>&times;</button>
    </div>
    <div class="modal-body-toast">
      <p>{message}</p>
    </div>
  </div>
{/if}

<style>
  .modal-backdrop-toast {
    position: fixed;
    top: 0;
    left: 0;
    width: 100%;
    height: 100%;
    /* Hafif bir arka plan karartması, isteğe bağlı */
    background-color: rgba(0, 0, 0, 0.1);
    z-index: 1000;
    pointer-events: none; /* Arka plandaki tıklamaları engelleme */
  }

  .modal-container-toast {
    position: fixed;
    top: 20px; /* Üstten biraz boşluk */
    left: 50%;
    transform: translateX(-50%); /* Yatayda ortalama */
    background-color: white;
    padding: 15px 20px;
    border-radius: 8px;
    box-shadow: 0 4px 12px rgba(0, 0, 0, 0.2);
    z-index: 1001;
    min-width: 300px;
    max-width: 90%;
    display: flex;
    flex-direction: column;
    gap: 10px;
    animation:
      fadeInDown 0.3s ease-out forwards,
      fadeOutUp 0.3s ease-in forwards var(--duration-delay);
    will-change: transform, opacity;
    box-sizing: border-box; /* Padding'in genişliğe dahil olmasını sağla */
  }

  .modal-header-toast {
    display: flex;
    align-items: center;
    gap: 10px;
    font-weight: bold;
    font-size: 1.1em;
    color: #333;
  }

  .modal-header-toast h2 {
    margin: 0;
    font-size: 1em; /* Başlığı daha küçük yap */
    flex-grow: 1;
  }

  .modal-icon {
    font-size: 1.5em; /* İkon boyutu */
    line-height: 1;
  }

  .close-button-toast {
    background: none;
    border: none;
    font-size: 1.5em;
    cursor: pointer;
    color: #aaa;
    padding: 0;
    line-height: 1;
  }

  .close-button-toast:hover {
    color: #666;
  }

  .modal-body-toast {
    font-size: 0.9em;
    color: #555;
  }

  /* Tipine göre stil güncellemeleri */
  .modal-info {
    border-left: 5px solid #2196f3;
  } /* Mavi */
  .modal-success {
    border-left: 5px solid #4caf50;
  } /* Yeşil */
  .modal-warning {
    border-left: 5px solid #ffc107;
  } /* Sarı */
  .modal-error {
    border-left: 5px solid #f44336;
  } /* Kırmızı */

  .modal-info .modal-icon {
    color: #2196f3;
  }
  .modal-success .modal-icon {
    color: #4caf50;
  }
  .modal-warning .modal-icon {
    color: #ffc107;
  }
  .modal-error .modal-icon {
    color: #f44336;
  }

  /* Animasyonlar */
  @keyframes fadeInDown {
    from {
      opacity: 0;
      transform: translate(-50%, -20px);
    }
    to {
      opacity: 1;
      transform: translate(-50%, 0);
    }
  }

  @keyframes fadeOutUp {
    from {
      opacity: 1;
      transform: translate(-50%, 0);
    }
    to {
      opacity: 0;
      transform: translate(-50%, -20px);
    }
  }
</style>
