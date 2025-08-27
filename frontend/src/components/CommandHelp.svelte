<script lang="ts">
  export let onClose: () => void;
  export let onCommandSelect: (cmd: string) => void;
</script>

<div class="overlay" on:click={onClose}>
  <div class="panel" on:click|stopPropagation>
    <header>
      <h2>📘 Komut Yardımı</h2>
      <p class="shortcut">
        Bu pencere <kbd>?</kbd> ile açılır, <kbd>Esc</kbd> ile kapanır.
      </p>
    </header>

    <ul class="command-list">
      <li>
        <code on:click={() => onCommandSelect("c 109403")}>c 109403</code> – İlk
        roster zaman çizelgesinde Personel ID'ye göre filtrele.
      </li>
      <li>
        <code on:click={() => onCommandSelect("1/c 109403")}>1/c 109403</code> –
        Belirli zaman çizelgesine filtre uygula.
      </li>
      <li>
        <code on:click={() => onCommandSelect("2/a FLT")}>2/a FLT</code> – İkinci
        zaman çizelgesinde FLT aktivitelerini göster.
      </li>
      <li>
        <code on:click={() => onCommandSelect("/r c 109403")}>/r c 109403</code>
        – Yeni Roster zaman çizelgesi aç.
      </li>
      <li>
        <code on:click={() => onCommandSelect("/t t TRIP-001")}
          >/t t TRIP-001</code
        > – Yeni Trip zaman çizelgesi aç.
      </li>
      <li>
        <code on:click={() => onCommandSelect("/rot a ROTATION-001")}
          >/rot a ROTATION-001</code
        > – Yeni Rotation zaman çizelgesi aç.
      </li>
      <li>
        <code on:click={() => onCommandSelect("a FLT")}>a FLT</code> – Aktivite kodu
        FLT olanları listele.
      </li>
      <li>
        <code on:click={() => onCommandSelect("dp IST ap JFK")}
          >dp IST ap JFK</code
        > – İstanbul kalkışlı, JFK varışlı uçuşları filtrele.
      </li>
      <li>
        <code on:click={() => onCommandSelect("fi TK001-IST-20240710120000")}
          >fi TK001-IST-20240710120000</code
        > – Uçuş ID’ye göre getir.
      </li>
      <li>
        <code on:click={() => onCommandSelect("cl C1")}>cl C1</code> – C1 sınıfındaki
        kayıtları göster.
      </li>
    </ul>

    <p class="note">
      <strong>Not:</strong> Eğer bir prefix (örn. <code>1/</code> veya
      <code>/t</code>) verilmezse, komut ilk (varsayılan) zaman çizelgesine
      uygulanır. Tüm komutlar boşlukla ayrılmalı ve çift sayıda yazılmalıdır
      (alan + değer).
    </p>

    <div class="actions">
      <button class="btn-close" on:click={onClose}>Kapat</button>
    </div>
  </div>
</div>

<style>
  .overlay {
    position: fixed;
    inset: 0;
    background-color: rgba(0, 0, 0, 0.5);
    display: flex;
    align-items: center;
    justify-content: center;
    z-index: 9999;
    backdrop-filter: blur(2px);
  }

  .panel {
    background-color: #ffffff;
    border-radius: 12px;
    padding: 32px;
    width: 580px;
    max-width: 90%;
    max-height: 90%;
    box-shadow: 0 8px 24px rgba(0, 0, 0, 0.2);
    overflow-y: auto;
    font-family: "Segoe UI", Tahoma, Geneva, Verdana, sans-serif;
  }

  header {
    margin-bottom: 16px;
  }

  h2 {
    margin: 0;
    font-size: 22px;
    color: #2c3e50;
  }

  .shortcut {
    font-size: 13px;
    color: #666;
    margin-top: 4px;
  }

  ul.command-list {
    list-style: none;
    padding-left: 0;
    margin: 20px 0;
  }

  .command-list li {
    margin-bottom: 10px;
    font-size: 15px;
    color: #333;
    line-height: 1.5;
  }

  code {
    font-family: "Courier New", Courier, monospace;
    background-color: #f4f4f4;
    padding: 4px 10px;
    border-radius: 6px;
    cursor: pointer;
    transition: background-color 0.2s ease;
    color: #2c3e50;
    font-size: 14px;
  }

  code:hover {
    background-color: #e2e6ea;
  }

  .note {
    background-color: #f9f9f9;
    padding: 12px 16px;
    border-left: 4px solid #007bff;
    font-size: 14px;
    color: #555;
    border-radius: 6px;
    margin-top: 24px;
  }

  .actions {
    display: flex;
    justify-content: flex-end;
    margin-top: 24px;
  }

  .btn-close {
    background-color: #007bff;
    color: white;
    padding: 8px 16px;
    font-size: 14px;
    border: none;
    border-radius: 6px;
    cursor: pointer;
    transition: background-color 0.2s ease;
  }

  .btn-close:hover {
    background-color: #0056b3;
  }
</style>
