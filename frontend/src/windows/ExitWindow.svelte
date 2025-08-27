<script lang="ts">
  import { logout } from "@/stores/authStore";
  import { createEventDispatcher } from "svelte";
  // SvelteKit kullanıyorsanız bu satırı açın:
  // import { goto } from "$app/navigation";

  const dispatch = createEventDispatcher();
  export let windowId: string | undefined = undefined;

  async function handleConfirmLogout() {
    // 1. Önce UI modali kapat
    // Bu, SvelteKit'in yönlendirme sırasında unmount işlemi yapmasından önce
    // bileşenin kendi event'ini göndermesini sağlar.
    if (windowId) {
      dispatch("closeWindow", windowId);
    }

    // 2. Auth state'i sıfırla
    logout();

    // 3. Yönlendir
    // Düz Svelte/Web uygulamaları için:
    window.location.href = "/";
    // SvelteKit uygulamaları için bu satırı kullanın:
    // await goto("/");
  }

  function handleCancelLogout() {
    // Sadece pencereyi kapat, oturumda bir değişiklik yapma
    if (windowId) {
      dispatch("closeWindow", windowId);
    }
  }
</script>

<div class="window-content">
  <h3>Uygulamadan Çıkış</h3>
  <p>Oturumunuzu kapatmak istediğinizden emin misiniz?</p>
  <div class="button-group">
    <button class="exit-button" on:click={handleConfirmLogout}>Evet, Çık</button
    >
    <button class="cancel-button" on:click={handleCancelLogout}
      >Hayır, İptal</button
    >
  </div>
</div>

<style>
  .window-content {
    padding: 24px;
    text-align: center;
    display: flex;
    flex-direction: column;
    align-items: center;
  }

  h3 {
    font-size: 1.25rem;
    color: var(--thy-primary-red, #c00000);
    margin-bottom: 12px;
  }

  p {
    font-size: 1rem;
    margin-bottom: 24px;
    color: var(--thy-text-color, #333);
  }

  .button-group {
    display: flex;
    gap: 12px;
  }

  button {
    padding: 10px 20px;
    border-radius: 4px;
    font-weight: 500;
    font-size: 0.95rem;
    cursor: pointer;
    border: 1px solid var(--thy-medium-gray, #ccc);
    background-color: var(--thy-light-gray, #f8f8f8);
    color: var(--thy-text-color, #333);
    transition: background-color 0.2s ease-in-out;
    min-width: 100px;
  }

  .exit-button {
    background-color: var(--thy-primary-red, #c00000);
    color: white;
    border-color: var(--thy-primary-red, #c00000);
  }

  .exit-button:hover {
    background-color: #a00000;
  }

  .cancel-button:hover {
    background-color: var(--thy-medium-gray, #ddd);
  }
</style>
