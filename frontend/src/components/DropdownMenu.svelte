<script lang="ts">
  import { createEventDispatcher, onMount, onDestroy } from "svelte";

  const dispatch = createEventDispatcher();

  export let items: { label: string; action: string }[] = [];
  export let show: boolean = false; // Dışarıdan kontrol edilebilir görünürlük
  export let positionTop: number = 0;
  export let positionLeft: number = 0;

  let dropdownRef: HTMLDivElement;

  function handleButtonClick(event: MouseEvent) {
    // show = !show; // Bu bileşenin kendi toggle'ı olmayacak, dışarıdan yönetilecek
    // Dispatch'i burada tetiklemiyoruz, bu buton sadece dropdown'ı göster/gizle eylemini yapar.
    // Parent component'in toggleDropdown fonksiyonu handleButtonClick yerine geçecektir.
    // Bu component'in kullanımı aşağıdaki TimelineItem'da gösterilecek.
  }

  function handleMenuItemClick(action: string) {
    dispatch("select", action); // Seçilen eylemi parent'a gönder
    show = false; // Menüyü kapat
  }

  // Dışarı tıklama olayını dinle
  function handleClickOutside(event: MouseEvent) {
    if (show && dropdownRef && !dropdownRef.contains(event.target as Node)) {
      show = false; // Menüyü kapat
      dispatch("close"); // Parent'a kapandığını bildir
    }
  }

  onMount(() => {
    document.addEventListener("click", handleClickOutside, true); // true ile capture aşamasında dinle
  });

  onDestroy(() => {
    document.removeEventListener("click", handleClickOutside, true);
  });

  // `show` değişkeni dışarıdan değiştiğinde, menüyü otomatik olarak kapatmak için
  // bu reaktif ifadeye ihtiyacımız yok çünkü `handleClickOutside` zaten bunu yapıyor.
  // Ancak `show` dışarıdan `false` olarak ayarlanırsa, menü kapanacaktır.
</script>

<div
  class="dropdown-wrapper"
  bind:this={dropdownRef}
  style="top: {positionTop}px; left: {positionLeft}px;"
>
  {#if show}
    <div class="dropdown-menu">
      {#each items as item}
        <button on:click={() => handleMenuItemClick(item.action)}>
          {item.label}
        </button>
      {/each}
    </div>
  {/if}
</div>

<style>
  .dropdown-wrapper {
    position: fixed; /* Parent component'ten alınan pozisyona göre sabitlenecek */
    z-index: 500; /* Canvas'tan yüksek, tooltip'ten düşük (varsayım) */
    /* top ve left style prop'ları ile dinamik olarak ayarlanır */
  }

  .dropdown-menu {
    position: absolute; /* Kendi wrapper'ına göre konumlanır */
    top: 100%; /* Butonun hemen altında başlar */
    left: 0;
    background-color: white;
    border: 1px solid #ddd;
    box-shadow: 0 2px 8px rgba(0, 0, 0, 0.15);
    border-radius: 4px;
    min-width: 150px;
    display: flex;
    flex-direction: column;
    z-index: 501; /* Wrapper'ın üzerinde */
  }

  .dropdown-menu button {
    background: none;
    border: none;
    padding: 8px 12px;
    text-align: left;
    cursor: pointer;
    font-size: 14px;
    color: #333;
    width: 100%;
  }

  .dropdown-menu button:hover {
    background-color: #f0f0f0;
  }
</style>
