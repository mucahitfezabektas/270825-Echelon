<script lang="ts">
  // Değişiklikleri Kaydet'e özel mantık buraya gelecek
  let saveMessage: string = "Kaydedilecek bekleyen değişiklikler bulunuyor.";
  let showConfirmation: boolean = false;

  function confirmSave() {
    saveMessage = "Tüm değişiklikler başarıyla kaydedildi!";
    showConfirmation = true;
    setTimeout(() => {
      // Bu pencereyi otomatik kapatmak isterseniz:
      // window.close() veya bir event ile ana bileşenden kapatma tetikleme
    }, 2000);
  }
</script>

<div class="window-content-wrapper">
  <h2>Değişiklikleri Kaydet</h2>
  <p>{saveMessage}</p>
  {#if !showConfirmation}
    <button on:click={confirmSave} class="primary-button">Şimdi Kaydet</button>
    <button
      on:click={() => alert("Kaydetme işlemi iptal edildi.")}
      class="secondary-button">İptal</button
    >
  {:else}
    <p>Kaydetme işlemi tamamlandı.</p>
  {/if}
</div>

<style>
  .window-content-wrapper {
    padding: 1.5rem;
    font-family: "Inter", sans-serif;
    color: var(--thy-text-color);
    text-align: center;
  }
  h2 {
    color: var(--thy-primary-red);
    margin-top: 0;
    margin-bottom: 1rem;
  }
  p {
    margin-bottom: 1.5rem;
    font-size: 1.1em;
  }
  .primary-button,
  .secondary-button {
    border: none;
    padding: 10px 20px;
    border-radius: 4px;
    cursor: pointer;
    font-size: 1rem;
    font-weight: 500;
    transition: background-color 0.2s ease-in-out;
    margin: 0 0.5rem;
  }
  .primary-button {
    background-color: var(--thy-primary-red);
    color: var(--thy-white);
  }
  .primary-button:hover {
    background-color: #a00000;
  }
  .secondary-button {
    background-color: var(--thy-medium-gray);
    color: var(--thy-text-color);
  }
  .secondary-button:hover {
    background-color: var(--thy-dark-gray);
    color: var(--thy-white);
  }
</style>
