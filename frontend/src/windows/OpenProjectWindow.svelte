<script lang="ts">
  // Mevcut Projeyi Aç'a özel mantık buraya gelecek
  // Örneğin, bir dosya seçici veya proje listesi
  let selectedProject: string = "";
  const projects = [
    "Proje A - Roster 2024",
    "Proje B - Uçuş Planlama",
    "Proje C - Ekip Yönetimi",
  ];

  function openProject() {
    if (selectedProject) {
      alert(`${selectedProject} projesi açılıyor...`);
      // Gerçek bir proje açma işlemi burada yapılabilir
    } else {
      alert("Lütfen bir proje seçin.");
    }
  }
</script>

<div class="window-content-wrapper">
  <h2>Mevcut Projeyi Aç</h2>
  <p>Açmak istediğiniz projeyi seçin:</p>
  <select bind:value={selectedProject}>
    <option value="" disabled>Proje Seçin</option>
    {#each projects as project}
      <option value={project}>{project}</option>
    {/each}
  </select>
  <button on:click={openProject} class="primary-button">Projeyi Aç</button>
</div>

<style>
  .window-content-wrapper {
    padding: 1.5rem;
    font-family: "Inter", sans-serif;
    color: var(--thy-text-color);
  }
  h2 {
    color: var(--thy-secondary-blue);
    margin-top: 0;
    margin-bottom: 1rem;
  }
  select {
    width: 100%;
    padding: 8px 10px;
    border: 1px solid var(--thy-border-color);
    border-radius: 4px;
    font-size: 1rem;
    margin-bottom: 1rem;
    box-sizing: border-box;
  }
  .primary-button {
    background-color: var(--thy-secondary-blue);
    color: var(--thy-white);
    border: none;
    padding: 10px 20px;
    border-radius: 4px;
    cursor: pointer;
    font-size: 1rem;
    font-weight: 500;
    transition: background-color 0.2s ease-in-out;
  }
  .primary-button:hover {
    background-color: #003a6a; /* Mavi tonu */
  }
</style>
