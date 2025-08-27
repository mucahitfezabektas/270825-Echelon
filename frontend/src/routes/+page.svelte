<script lang="ts">
  import { onMount, onDestroy } from "svelte";
  import Login from "@/components/Login.svelte";
  import AppLayout from "@/windows/MainScreen.svelte";
  import SplashScreen from "@/components/SplashScreen.svelte";

  import { authStore, initializeAuth, logout } from "@/stores/authStore";
  import { setStatus } from "@/stores/appStates";
  import {
    hideContextMenu,
    contextMenuStore,
    showContextMenu,
    type ContextMenuItem,
  } from "@/stores/contextMenuStore";
  import ContextMenu from "@/components/ContextMenu.svelte";
  import { showAlert } from "@/stores/alertStore"; // <-- YENİ EKLEME: Özel alert fonksiyonunu import et

  // Yükleme ve başlangıç durumu değişkenleri
  let initializing = true;
  let splashMessage = "Uygulama başlatılıyor...";
  let isLoadingData = false;

  // Context menü store'una abone ol
  let currentContextMenuState = $contextMenuStore;
  const unsubscribeContextMenu = contextMenuStore.subscribe((value) => {
    currentContextMenuState = value;
  });

  // Doğrudan authStore'dan isAuthenticated değerini takip et
  $: isAuthenticated = $authStore.isAuthenticated;

  onMount(() => {
    // initializeAuth() çağrısı authStore.ts içinde otomatik olarak yapılıyor.
    // Eğer sadece bir kez App.svelte yüklendiğinde tetiklenmesini istiyorsanız
    // buraya çağırabilirsiniz, ancak authStore.ts'deki çağrı genellikle yeterlidir.
    // initializeAuth(); // Eğer authStore.ts'de initializeAuth() çağrısını yorum satırı yaparsanız, buraya ekleyebilirsiniz.

    // Splash ekranını belli bir süre sonra gizle
    setTimeout(() => {
      initializing = false;
    }, 2000);

    // Global tıklamaları dinleyerek context menüyü kapatma
    document.body.addEventListener("mousedown", handleGlobalMouseDown, true);
    document.body.addEventListener("contextmenu", handleGlobalRightClick, true);
  });

  onDestroy(() => {
    // Component yok edildiğinde event listener'ları temizle
    document.body.removeEventListener("mousedown", handleGlobalMouseDown, true);
    document.body.removeEventListener(
      "contextmenu",
      handleGlobalRightClick,
      true
    );
    unsubscribeContextMenu();
  });

  async function onLoginSuccess() {
    isLoadingData = true;
    setStatus("Giriş başarılı! Veriler yükleniyor...");

    // Giriş başarılı olduktan sonra verileri yükle
    // tableDataLoader.loadAll(); // Örnek veri yüklemesi
    // preloadCrewMatchStatuses(); // Örnek veri yüklemesi

    // Simülasyon: Veri yükleme süresi
    await new Promise((res) => setTimeout(res, 1000));

    setStatus("Veriler yüklendi. TK CMS");
    isLoadingData = false;
  }

  // Global MOUSE DOWN olayı ile menü kapatma
  function handleGlobalMouseDown(event: MouseEvent) {
    if (!currentContextMenuState.show) {
      return;
    }

    const contextMenuElement = document.getElementById("context-menu-popup");

    if (
      contextMenuElement &&
      event.target instanceof Node &&
      !contextMenuElement.contains(event.target)
    ) {
      console.log("+page.svelte: Menü dışına tıklandı, kapatılıyor.");
      hideContextMenu();
    }
  }

  // Global SAĞ TIKLAMA olayı ile genel context menü açma
  function handleGlobalRightClick(event: MouseEvent) {
    console.log(
      "+page.svelte: handleGlobalRightClick tetiklendi. Hedef:",
      event.target
    );

    event.preventDefault();

    const targetElement = event.target as HTMLElement;

    if (targetElement.closest('[data-context-menu-trigger="true"]')) {
      console.log(
        "+page.svelte: Hedef özel bir menü tetikleyicisi. Genel menü açılmıyor."
      );
      return;
    }

    if (currentContextMenuState.show) {
      console.log("+page.svelte: Mevcut menü açık, genel alanda kapatılıyor.");
      hideContextMenu();
      return;
    }

    const x = event.pageX;
    const y = event.pageY;

    const generalMenuItems: ContextMenuItem[] = [
      {
        label: "Ayarlar",
        action: () => showAlert("Ayarlar penceresi açılacak!", "Ayarlar"),
      }, // <-- alert() yerine showAlert()
      { separator: true },
      {
        label: "Yardım",
        action: () => showAlert("Yardım penceresi açılacak!", "Yardım"),
      }, // <-- alert() yerine showAlert()
      {
        label: "Çıkış Yap",
        action: () => {
          console.log("Çıkış yapılıyor...");
          logout(); // `authStore`'dan import edilen `logout` fonksiyonu
        },
      },
    ];

    showContextMenu(x, y, generalMenuItems, { context: "global" });
  }
</script>

{#if initializing}
  <SplashScreen message={splashMessage} />
{:else if !isAuthenticated}
  {#if isLoadingData}
    <SplashScreen message="Veriler yükleniyor…" />
  {:else}
    <Login on:loginSuccess={onLoginSuccess} />
  {/if}
{:else}
  <AppLayout />
{/if}

<ContextMenu />

<style>
  :global(body) {
    margin: 0;
    padding: 0;
    box-sizing: border-box;
    overflow: hidden;
    font-family: "Inter", sans-serif;
    color: var(--thy-text-color);
    background-color: var(--thy-light-gray);
  }
</style>
