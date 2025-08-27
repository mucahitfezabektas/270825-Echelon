<script lang="ts">
  import ToolbarButton from "@/components/ToolbarButton.svelte";
  import DraggableResizableWindow from "@/components/DraggableResizableWindow.svelte";
  import { statusMessage, setStatus } from "@/stores/appStates";
  import { toolbarActions } from "@/stores/toolbarStates";
  import DockSplitter from "@/components/DockSplitter.svelte";
  import {
    getWindowConfig,
    getComponentMap,
    type WindowConfig,
  } from "@/config/windowConfig";
  import { menuStructure, type MenuItem } from "@/stores/menuStructureStore";
  import { logInfo, logWarn } from "@/stores/logStore"; // Import logInfo for logging window actions

  import {
    isSnappingEnabled,
    selectedLayoutId,
    snapZones,
    hoveredZone,
    updateSnapZones,
    detectHoveredZone,
    applySnap,
    availableLayouts,
    setMainAreaOffsets,
  } from "@/stores/snappingStore";

  import type { SvelteComponent } from "svelte";
  import { onMount, onDestroy } from "svelte";
  import { ghostDockZone } from "@/stores/uiState";
  import { timezoneOffsetMinStore } from "@/stores/userOptionsStore";

  let now = new Date();
  let offsetMin = 0;
  let offsetHour = 0;
  let formattedUTC = "";

  onMount(() => {
    const timer = setInterval(() => {
      now = new Date();
      updateTime();
    }, 1000);
    return () => clearInterval(timer);
  });

  

  timezoneOffsetMinStore.subscribe((val) => {
    offsetMin = val;
    console.log(offsetMin)
    offsetHour = offsetMin / 60;
    updateTime();
  });

  function updateTime() {
    formattedUTC = new Intl.DateTimeFormat("tr-TR", {
      timeZone: "Etc/GMT-3",
      year: "numeric",
      month: "2-digit",
      day: "2-digit",
      hour: "2-digit",
      minute: "2-digit",
      second: "2-digit",
      hour12: false,
    })
      .format(now)
      .replace(",", "");
  }

  type DynamicSvelteComponent = new (...args: any[]) => SvelteComponent;

  interface AppWindow {
    id: number;
    x: number;
    y: number;
    width: number;
    height: number;
    title: string;
    active: boolean;
    zIndex: number;
    component: DynamicSvelteComponent;
    props?: Record<string, any>;
    minimizable?: boolean;
    maximizable?: boolean;
    minimized: boolean;
    restoreX?: number;
    restoreY?: number;
    restoreWidth?: number;
    restoreHeight?: number;
    dockPosition?: "left" | "right" | null;
    previousDockBounds?: {
      x: number;
      y: number;
      width: number;
      height: number;
    };
  }

  let openWindows: AppWindow[] = [];
  let nextWindowId = 0;
  let nextZIndex = 100;
  let activeWindowId: number | null = null;

  let activeMenuId: string | null = null;

  const componentMap = getComponentMap();

  let mainAreaRef: HTMLDivElement | null = null;
  let mainAreaObserver: ResizeObserver | null = null;
  let windowRefMap: Record<
    number,
    InstanceType<typeof DraggableResizableWindow>
  > = {};
  let tabsBarVisible = false; // Başlangıçta gizli olsun
  let showToggleButton = false; // Butonun görünürlüğünü kontrol eden yeni değişken

  let appWrapperRef: HTMLDivElement | null = null; // app-wrapper için referans

  onMount(() => {
    document.addEventListener("click", () => (activeMenuId = null));

    if (mainAreaRef) {
      mainAreaObserver = new ResizeObserver((entries) => {
        if (entries[0] && entries[0].target) {
          const rect = entries[0].target.getBoundingClientRect();
          setMainAreaOffsets(rect.left, rect.top);
          updateSnapZones(rect);
        }
      });
      mainAreaObserver.observe(mainAreaRef);
    }

    // Fare hareketini dinleyerek sağ alt köşeyi algıla
    document.addEventListener("mousemove", handleMouseMove);

    // Log that the application has started
    setStatus("Uygulama başlatıldı.");
    logInfo("Application started.");
  });

  onDestroy(() => {
    document.removeEventListener("click", () => (activeMenuId = null));

    if (mainAreaObserver && mainAreaRef) {
      mainAreaObserver.disconnect();
    }
    document.removeEventListener("mousemove", handleMouseMove);
  });

  // Fare hareketini dinleyen fonksiyon
  function handleMouseMove(event: MouseEvent) {
    if (!appWrapperRef) return;

    const rect = appWrapperRef.getBoundingClientRect();
    const distanceFromRight = rect.right - event.clientX;
    const distanceFromBottom = rect.bottom - event.clientY;

    const triggerZone = 100; // Sağ ve alttan bu kadar piksel içinde olursa tetikle

    if (distanceFromRight < triggerZone && distanceFromBottom < triggerZone) {
      showToggleButton = true;
    } else {
      showToggleButton = false;
    }
  }

  $: if ($isSnappingEnabled || $selectedLayoutId) {
    if (mainAreaRef) {
      updateSnapZones(mainAreaRef.getBoundingClientRect());
    }
  }
  $: dockedLeft = openWindows.find((w) => w.dockPosition === "left");
  $: dockedRight = openWindows.find((w) => w.dockPosition === "right");

  /** GMT etiketine tıklanınca çalışır  */
  function openUserOptionsTimezone() {
    // windowConfig içinde “UserOptionsWindow” tanımlı diye varsayıyoruz
    const cfg = getWindowConfig("UserOptionsWindow");
    if (!cfg) return;

    /*  props içine başlangıç sekmesini ilet →  
      UserOptionsWindow.svelte içinde `export let activeMenuItem = "general";`
      gibi bir değişken olduğunu varsayıyoruz.                    */
    openNewWindow(
      cfg.component,
      cfg.title,
      cfg.initialWidth,
      cfg.initialHeight,
      cfg.minimizable ?? false,
      cfg.maximizable ?? false,
      { activeMenuItem: "timezone" } // 🔹 EKLENEN satır
    );
  }

  function handleDockResize(deltaX: number) {
    if (!dockedLeft || !dockedRight) return;

    const parentWidth = window.innerWidth;
    const minWidth = 200;

    let newLeftWidth = dockedLeft.width + deltaX;
    newLeftWidth = Math.min(
      Math.max(minWidth, newLeftWidth),
      parentWidth - minWidth
    );
    const newRightWidth = parentWidth - newLeftWidth;

    openWindows = openWindows.map((w) => {
      if (w.id === dockedLeft.id) return { ...w, width: newLeftWidth };
      if (w.id === dockedRight.id)
        return { ...w, x: newLeftWidth, width: newRightWidth };
      return w;
    });
  }

  function handleDockReset() {
    if (!dockedLeft || !dockedRight) return;
    const mainAreaHeight =
      mainAreaRef?.getBoundingClientRect().height || window.innerHeight;

    const parentWidth = window.innerWidth;
    const halfWidth = parentWidth / 2;

    openWindows = openWindows.map((w) => {
      if (w.id === dockedLeft.id)
        return { ...w, x: 0, width: halfWidth, height: mainAreaHeight };
      if (w.id === dockedRight.id)
        return { ...w, x: halfWidth, width: halfWidth, height: mainAreaHeight };
      return w;
    });
  }

  function handleDock(detail: {
    id: number;
    x: number;
    y: number;
    width: number;
    height: number;
    position: "left" | "right";
  }) {
    const parentWidth = window.innerWidth;
    const mainAreaHeight =
      mainAreaRef?.getBoundingClientRect().height || window.innerHeight;
    const otherSide = detail.position === "left" ? "right" : "left";

    const oppositeWindow = openWindows.find(
      (w) => w.dockPosition === otherSide
    );
    const oppositeWidth = oppositeWindow?.width ?? null;
    const remainingWidth = oppositeWidth
      ? parentWidth - oppositeWidth
      : parentWidth / 2;

    openWindows = openWindows.map((win) => {
      if (win.id === detail.id) {
        const newWidth = oppositeWidth ? remainingWidth : parentWidth / 2;
        const newX = detail.position === "left" ? 0 : parentWidth - newWidth;

        return {
          ...win,
          x: newX,
          y: 0,
          width: newWidth,
          height: mainAreaHeight, // 🔥 sadece main-area yüksekliği kadar
          dockPosition: detail.position,
          previousDockBounds: {
            x: win.x,
            y: win.y,
            width: win.width,
            height: win.height,
          },
        };
      }

      if (oppositeWindow && win.id === oppositeWindow.id) {
        return {
          ...win,
          x: win.dockPosition === "left" ? 0 : parentWidth - win.width,
          y: 0,
          width: win.width,
          height: mainAreaHeight,
        };
      }

      return win;
    });

    ghostDockZone.set(null);
    const title = openWindows.find((w) => w.id === detail.id)?.title;
    setStatus(`"${title}" penceresi ${detail.position} kenarına sabitlendi.`);
    logInfo(`Docked window: "${title}" (${detail.position})`);
  }

  function handleUndock({ id }: { id: number }) {
    const windowData = openWindows.find((w) => w.id === id);
    if (!windowData?.previousDockBounds) return;

    openWindows = openWindows.map((w) =>
      w.id === id
        ? {
            ...w,
            ...w.previousDockBounds, // geri yükle
            dockPosition: null,
            previousDockBounds: undefined,
          }
        : w
    );

    const title = windowData.title;
    setStatus(`"${title}" penceresi sabitlemeden çıkarıldı.`);
    logInfo(`Undocked window: "${title}"`);
  }

  function openNewWindow(
    component: DynamicSvelteComponent,
    title: string,
    initialWidth: number,
    initialHeight: number,
    shouldBeMinimizable: boolean,
    shouldBeMaximizable: boolean,
    props?: Record<string, any>
  ) {
    const newId = nextWindowId++;
    const currentZIndex = nextZIndex++;

    const mainAreaRect = mainAreaRef?.getBoundingClientRect();
    let startX = 0;
    let startY = 0;

    if (mainAreaRect) {
      startX = Math.max(
        0,
        (mainAreaRect.width - initialWidth) / 2 - mainAreaRect.left
      );
      startY = Math.max(
        0,
        (mainAreaRect.height - initialHeight) / 2 - mainAreaRect.top
      );
    } else {
      const viewportWidth = document.body.clientWidth;
      const viewportHeight = document.body.clientHeight;
      startX = Math.max(0, (viewportWidth - initialWidth) / 2);
      startY = Math.max(0, (viewportHeight - initialHeight) / 2);
    }

    const newWindow: AppWindow = {
      id: newId,
      component,
      title,
      x: startX,
      y: startY,
      width: initialWidth,
      height: initialHeight,
      active: true,
      zIndex: currentZIndex,
      minimizable: shouldBeMinimizable,
      maximizable: shouldBeMaximizable,
      minimized: false,
      props: props,
    };

    openWindows = openWindows.map((win) => ({ ...win, active: false }));
    openWindows = [...openWindows, newWindow];
    activeWindowId = newId;

    setStatus(`Pencere "${title}" açıldı.`);
    logInfo(`Opened window: "${title}" (ID: ${newId})`); // Log window open event
  }

  function activateWindow(id: number) {
    const targetWindow = openWindows.find((win) => win.id === id);

    if (!targetWindow) return;

    if (targetWindow.active && !targetWindow.minimized) {
      return;
    }

    const currentZIndex = nextZIndex++;

    openWindows = openWindows.map((win) => {
      if (win.id === id) {
        return {
          ...win,
          active: true,
          zIndex: currentZIndex,
          minimized: false,
          x: win.restoreX !== undefined ? win.restoreX : win.x,
          y: win.restoreY !== undefined ? win.restoreY : win.y,
          width: win.restoreWidth !== undefined ? win.restoreWidth : win.width,
          height:
            win.restoreHeight !== undefined ? win.restoreHeight : win.height,
        };
      } else {
        return { ...win, active: false };
      }
    });
    activeWindowId = id;
    const windowTitle = openWindows.find((w) => w.id === id)?.title;
    setStatus(`Pencere "${windowTitle}" aktif edildi.`);
    logInfo(`Activated window: "${windowTitle}" (ID: ${id})`); // Log window activation
  }

  function closeWindow(id: number) {
    const closedWindowTitle = openWindows.find((w) => w.id === id)?.title;
    openWindows = openWindows.filter((win) => win.id !== id);

    if (activeWindowId === id) {
      activeWindowId =
        openWindows.length > 0 ? openWindows[openWindows.length - 1].id : null;
      if (activeWindowId !== null) {
        activateWindow(activeWindowId);
      }
    }
    setStatus(`Pencere "${closedWindowTitle}" kapatıldı.`);
    logInfo(`Closed window: "${closedWindowTitle}" (ID: ${id})`); // Log window close event
  }

  function toggleMenu(menuId: string, event: MouseEvent) {
    activeMenuId = activeMenuId === menuId ? null : menuId;
    event.stopPropagation();
  }

  function handleMenuItemSelect(actionId: string, event: MouseEvent) {
    setStatus(`Menü aksiyonu: '${actionId}' seçildi.`);
    logInfo(`Menu item selected: '${actionId}'`); // Log menu item selection
    activeMenuId = null;
    event.stopPropagation();

    const config = getWindowConfig(actionId);
    if (config) {
      openNewWindow(
        config.component,
        config.title,
        config.initialWidth,
        config.initialHeight,
        config.minimizable || false,
        config.maximizable || false
      );
    } else {
      console.log(
        `Bilinmeyen menü aksiyonu veya yapılandırma bulunamadı: ${actionId}`
      );
      logWarn(`Unknown menu action or config not found: ${actionId}`); // Log unknown menu action
    }
  }

  function handleGlobalKeydown(event: KeyboardEvent) {
    if (event.key === "Escape" && activeMenuId !== null) {
      activeMenuId = null;
    }
  }

  function handleToolbarButtonClick(actionId: string, label: string) {
    logInfo(`Toolbar button clicked: "${label}" (ID: ${actionId})`); // Log toolbar button click
    const config = getWindowConfig(actionId);
    if (config) {
      openNewWindow(
        config.component,
        config.title,
        config.initialWidth,
        config.initialHeight,
        config.minimizable || false,
        config.maximizable || false
      );
    } else {
      setStatus(
        `"${label}" (ID: ${actionId}) aksiyonu bilinmiyor veya henüz tanımlanmadı.`
      );
      logWarn(`Unknown toolbar action or config not found: ${actionId}`); // Log unknown toolbar action
    }
  }

  function handleMinimize(detail: { id: number }) {
    const windowId = detail.id;
    const currentWindow = openWindows.find((win) => win.id === windowId);
    if (currentWindow) {
      openWindows = openWindows.map((win) =>
        win.id === windowId
          ? {
              ...win,
              minimized: true,
              active: false,
              restoreX: win.x,
              restoreY: win.y,
              restoreWidth: win.width,
              restoreHeight: win.height,
            }
          : win
      );
      setStatus(`Pencere "${currentWindow.title}" minimize edildi.`);
      logInfo(`Minimized window: "${currentWindow.title}" (ID: ${windowId})`); // Log minimize
    }
  }

  function handleRestore(detail: { id: number }) {
    const windowId = detail.id;
    const currentWindow = openWindows.find((win) => win.id === windowId);
    if (currentWindow) {
      openWindows = openWindows.map((win) =>
        win.id === windowId
          ? {
              ...win,
              minimized: false,
            }
          : win
      );
      activateWindow(windowId);
      setStatus(`Pencere "${currentWindow.title}" geri yüklendi.`);
      logInfo(`Restored window: "${currentWindow.title}" (ID: ${windowId})`); // Log restore
    }
  }

  function handleMaximize(detail: {
    id: number;
    x: number;
    y: number;
    width: number;
    height: number;
  }) {
    const windowId = detail.id;
    const currentWindow = openWindows.find((win) => win.id === windowId);
    if (currentWindow) {
      openWindows = openWindows.map((win) =>
        win.id === windowId
          ? {
              ...win,
              x: detail.x,
              y: detail.y,
              width: detail.width,
              height: detail.height,
              minimized: false,
            }
          : win
      );
      setStatus(`Pencere "${currentWindow.title}" maksimize edildi.`);
      logInfo(`Maximized window: "${currentWindow.title}" (ID: ${windowId})`); // Log maximize
    }
  }

  function handleWindowDrag(
    event: CustomEvent<{
      id: number;
      x: number;
      y: number;
      mouseX: number;
      mouseY: number;
    }>
  ) {
    const windowId = event.detail.id;
    openWindows = openWindows.map((win) =>
      win.id === windowId
        ? { ...win, x: event.detail.x, y: event.detail.y }
        : win
    );
    detectHoveredZone(event.detail.mouseX, event.detail.mouseY);
  }

  function handleWindowDragEnd(
    event: CustomEvent<{
      id: number;
      x: number;
      y: number;
      mouseX: number;
      mouseY: number;
    }>
  ) {
    console.log("⛳ Drag END:", event.detail); // ✅ Bu log geliyor mu?

    const windowId = event.detail.id;
    const mouseX = event.detail.mouseX;
    const mouseY = event.detail.mouseY;

    applySnap(
      windowId,
      mouseX,
      mouseY,
      openWindows,
      (updatedWindows: AppWindow[]) => {
        openWindows = updatedWindows;
      },
      setStatus
    );
    hoveredZone.set(null);
  }

  function handleWindowResize(
    event: CustomEvent<{
      id: number;
      width: number;
      height: number;
      x: number;
      y: number;
    }>
  ) {
    const windowId = event.detail.id;
    openWindows = openWindows.map((win) =>
      win.id === windowId
        ? {
            ...win,
            width: event.detail.width,
            height: event.detail.height,
            x: event.detail.x,
            y: event.detail.y,
          }
        : win
    );
  }

  function renderMenuItems(items: MenuItem[]): MenuItem[] {
    return items
      .map((item) => {
        if (item.type === "submenu") {
          return {
            ...item,
            children: item.children ? renderMenuItems(item.children) : [],
          };
        } else if (item.type === "item") {
          const config = getWindowConfig(item.id);
          return {
            ...item,
            label: config?.title || item.label || item.id,
            icon: config?.icon || item.icon,
            actionId: item.id,
          };
        } else if (item.type === "separator") {
          return { ...item };
        }
        return null;
      })
      .filter(Boolean) as MenuItem[];
  }

  $: renderedMenuStructure = renderMenuItems($menuStructure);
</script>

<svelte:window on:keydown={handleGlobalKeydown} />

<div class="app-wrapper" bind:this={appWrapperRef as HTMLDivElement}>
  <header class="menu-bar">
    <nav class="menu-actions">
      {#each renderedMenuStructure as topLevelItem (topLevelItem.id)}
        <div class="menu-item-wrapper">
          <button
            class="menu-button"
            on:click={(e) => toggleMenu(topLevelItem.id, e)}
            aria-haspopup="true"
            aria-expanded={activeMenuId === topLevelItem.id}
          >
            {topLevelItem.label}
          </button>
          {#if topLevelItem.type === "submenu" && topLevelItem.children && activeMenuId === topLevelItem.id}
            <ul class="submenu">
              {#each topLevelItem.children as subItem (subItem.id)}
                {#if subItem.type === "item"}
                  <li>
                    <button
                      class="submenu-item"
                      on:click={(e) =>
                        handleMenuItemSelect(subItem.actionId || subItem.id, e)}
                    >
                      {#if subItem.icon}<i class={subItem.icon}></i>{/if}
                      {subItem.label}
                    </button>
                  </li>
                {:else if subItem.type === "submenu"}
                  <li class="submenu-group">
                    <button
                      class="submenu-group-header"
                      on:click|stopPropagation
                    >
                      <i class="fas fa-folder"></i>
                      {subItem.label}
                    </button>
                    <ul class="nested-submenu">
                      {#each subItem.children || [] as nestedSubItem (nestedSubItem.id)}
                        {#if nestedSubItem.type === "item"}
                          <li>
                            <button
                              class="submenu-item"
                              on:click={(e) =>
                                handleMenuItemSelect(
                                  nestedSubItem.actionId || nestedSubItem.id,
                                  e
                                )}
                            >
                              {#if nestedSubItem.icon}<i
                                  class={nestedSubItem.icon}
                                ></i>{/if}
                              {nestedSubItem.label}
                            </button>
                          </li>
                        {:else if nestedSubItem.type === "separator"}
                          <li class="submenu-separator"></li>
                        {/if}
                      {/each}
                    </ul>
                  </li>
                {:else if subItem.type === "separator"}
                  <li class="submenu-separator"></li>
                {/if}
              {/each}
            </ul>
          {/if}
        </div>
      {/each}
    </nav>
    <div class="snap-controls">
      <label>
        <input type="checkbox" bind:checked={$isSnappingEnabled} />
        Snap
      </label>
      <select bind:value={$selectedLayoutId}>
        {#each Object.entries(availableLayouts) as [id, layout]}
          <option value={id}>{layout.name}</option>
        {/each}
      </select>
    </div>
  </header>

  <div class="toolbar">
    {#each $toolbarActions as action (action.id)}
      {#if action.type === "separator"}
        <div class="toolbar-separator"></div>
      {:else}
        <ToolbarButton
          title={action.title}
          iconClass={action.icon}
          onClick={() => handleToolbarButtonClick(action.id, action.title)}
        />
      {/if}
    {/each}
  </div>

  <main class="main-area" bind:this={mainAreaRef as HTMLDivElement}>
    {#if $isSnappingEnabled && mainAreaRef}
      {#each $snapZones as zone (zone.id)}
        <div
          class="snap-zone"
          class:hovered={$hoveredZone === zone.id}
          style="left: {zone.x}px; top: {zone.y}px; width: {zone.width}px; height: {zone.height}px;"
        ></div>
      {/each}
    {/if}

    {#each openWindows as window (window.id)}
      <DraggableResizableWindow
        bind:this={windowRefMap[window.id]}
        id={window.id}
        x={window.x}
        y={window.y}
        width={window.width}
        height={window.height}
        title={window.title}
        active={window.active}
        zIndex={window.zIndex}
        minimizable={window.minimizable}
        maximizable={window.maximizable}
        minimized={window.minimized}
        on:activate={() => activateWindow(window.id)}
        on:close={() => closeWindow(window.id)}
        on:drag={handleWindowDrag}
        on:dragend={handleWindowDragEnd}
        on:resize={handleWindowResize}
        on:minimize={() => handleMinimize({ id: window.id })}
        on:maximize={(e) =>
          handleMaximize({
            id: window.id,
            x: e.detail.x,
            y: e.detail.y,
            width: e.detail.width,
            height: e.detail.height,
          })}
        on:restore={() => handleRestore({ id: window.id })}
        on:dock={(e) =>
          handleDock({
            id: window.id,
            x: e.detail.x,
            y: e.detail.y,
            width: e.detail.width,
            height: e.detail.height,
            position: e.detail.position,
          })}
        dockPosition={window.dockPosition}
        on:undock={(e) => handleUndock(e.detail)}
      >
        <svelte:component this={window.component} {...window.props || {}} />
      </DraggableResizableWindow>
    {/each}
  </main>

  <div
    class="tabs-bar"
    class:empty={openWindows.length === 0}
    class:hidden={!tabsBarVisible}
  >
    {#each openWindows as window (window.id)}
      <div
        class="tab-item"
        class:active={window.active}
        class:minimized={window.minimized}
        on:click={() => activateWindow(window.id)}
        on:keydown={(e: KeyboardEvent) => {
          if (e.key === "Enter" || e.key === " ") {
            e.preventDefault();
            activateWindow(window.id);
          }
        }}
        aria-label={`Pencereyi aç: ${window.title}`}
        role="tab"
        tabindex="0"
      >
        <span class="tab-title">{window.title}</span>
        <button
          class="tab-close-button"
          on:click|stopPropagation={() => closeWindow(window.id)}
          aria-label={`${window.title} penceresini kapat`}
        >
          <i class="fas fa-times"></i>
        </button>
      </div>
    {/each}
    {#if openWindows.length === 0}
      <span class="no-tabs-message">Açık pencere yok.</span>
    {/if}

    {#if $ghostDockZone === "left"}
      <div class="ghost-dock-zone left"></div>
    {/if}
    {#if $ghostDockZone === "right"}
      <div class="ghost-dock-zone right"></div>
    {/if}
    {#if dockedLeft && dockedRight}
      <DockSplitter
        x={dockedLeft.x + dockedLeft.width - 3}
        height={window.innerHeight}
        on:resizeSplit={(e) => handleDockResize(e.detail.deltaX)}
        on:resetSplit={handleDockReset}
      />
    {/if}
  </div>

  <button
    class="tabs-toggle-button"
    class:visible={showToggleButton || tabsBarVisible}
    on:click={() => (tabsBarVisible = !tabsBarVisible)}
    aria-label="Sekme çubuğunu aç/kapat"
    title={tabsBarVisible ? "Sekme çubuğunu gizle" : "Sekme çubuğunu göster"}
  >
    <i class="fas fa-chevron-up" class:rotated={!tabsBarVisible}></i>
  </button>

  <footer class="status-bar">
    <span class="status-text">
      <i class="fas fa-info-circle status-icon" aria-hidden="true"></i>
      {$statusMessage}
    </span>
    <div class="footer-right-section">
      <button
        class="log-button"
        on:click={() =>
          handleToolbarButtonClick("LogWindow", "Application Log")}
      >
        <i class="fas fa-clipboard-list"></i> Log
      </button>
      <div class="status-text">
        <div class="vertical-separator"></div>

        <!-- GMT ifadesini “button-benzeri” yapıyoruz -->
        <span
          class="label gmt-link"
          role="button"
          tabindex="0"
          title="Tercihler › Time Zone"
          on:click={openUserOptionsTimezone}
          on:keydown={(e) => {
            if (e.key === "Enter" || e.key === " ") {
              e.preventDefault();
              openUserOptionsTimezone();
            }
          }}
        >
          GMT{offsetHour >= 0 ? `+${offsetHour}` : offsetHour}
        </span>

        <div class="vertical-separator"></div>
        <span class="time">{formattedUTC}</span>
      </div>
    </div>
  </footer>
</div>

<style>
  @import url("https://fonts.googleapis.com/css2?family=Inter:wght@300;400;500;600;700&display=swap");
  @import url("https://cdnjs.cloudflare.com/ajax/libs/font-awesome/5.15.4/css/all.min.css");

  /* THY Kurumsal Renkleri */
  :root {
    --thy-primary-red: #c00000;
    --thy-secondary-blue: #004b85;
    --thy-light-blue: #e0f2f7;
    --thy-white: #ffffff;
    --thy-black: #000000;
    --thy-text-color: #333333;
    --thy-dark-gray: #666666;
    --thy-medium-gray: #dddddd;
    --thy-light-gray: #f8f8f8;
    --thy-off-white: #fefefe;
    --thy-border-color: #cccccc;

    /* Gölge ve Geçişler */
    --thy-shadow-light: 0 1px 3px rgba(0, 0, 0, 0.08);
    --thy-shadow-medium: 0 2px 8px rgba(0, 0, 0, 0.12);
    --thy-shadow-strong: 0 4px 12px rgba(0, 0, 0, 0.15);
    --thy-transition-speed: 0.2s;
    --thy-transition-timing: ease-in-out;

    /* Tabs Bar hidden state variables */
    --tabs-bar-height-visible: 30px;
    --tabs-bar-height-hidden: 0px; /* Bu değer 0 olarak ayarlandı */
    --tabs-bar-padding-visible: 0 0.5rem;
    --tabs-bar-padding-hidden: 0;
  }

  :global(body) {
    margin: 0;
    padding: 0;
    box-sizing: border-box;
    overflow: hidden;
    font-family: "Inter", sans-serif;
    color: var(--thy-text-color);
    background-color: var(--thy-light-gray);
  }

  .app-wrapper {
    height: 100vh;
    display: flex;
    flex-direction: column;
    font-size: 14px;
    background-color: var(--thy-off-white);
    box-shadow: var(--thy-shadow-strong);
    overflow: hidden;
    position: relative; /* Butonu konumlandırmak için eklendi */
  }

  /* Ana Başlık (Menu Bar) Stilleri - THY Kırmızısı */
  .menu-bar {
    height: 30px;
    background: var(--thy-primary-red);
    color: var(--thy-white);
    display: flex;
    align-items: center;
    padding: 0 1.5rem;
    box-shadow: var(--thy-shadow-medium);
    flex-shrink: 0;
    z-index: 10;
    justify-content: space-between;
  }

  .menu-actions {
    display: flex;
    gap: 0.8rem;
  }

  .menu-item-wrapper {
    position: relative;
    height: 100%;
    display: flex;
    align-items: center;
  }

  .menu-button {
    background: transparent;
    color: var(--thy-white);
    border: none;
    padding: 0.2rem 0.5rem;
    border-radius: 4px;
    cursor: pointer;
    font-size: 13px;
    font-weight: 500;
    transition:
      background-color var(--thy-transition-speed) var(--thy-transition-timing),
      color var(--thy-transition-speed) var(--thy-transition-timing),
      box-shadow var(--thy-transition-speed) var(--thy-transition-timing);
    outline: none;
    height: 100%;
  }

  .menu-button:hover {
    background-color: rgba(255, 255, 255, 0.2);
    color: var(--thy-white);
    box-shadow: inset 0 0 0 1px rgba(255, 255, 255, 0.3);
  }

  .menu-button:active {
    background-color: rgba(255, 255, 255, 0.3);
  }

  /* Submenu Styles */
  .submenu {
    position: absolute;
    top: 100%;
    left: 0;
    background-color: var(--thy-white);
    border: 1px solid var(--thy-medium-gray);
    box-shadow: var(--thy-shadow-medium);
    list-style: none;
    margin: 0;
    padding: 5px 0;
    min-width: 180px;
    z-index: 999;
    border-radius: 4px;
  }

  .submenu li {
    margin: 0;
    padding: 0;
  }

  .submenu-item {
    display: flex;
    align-items: center;
    gap: 8px;
    width: 100%;
    text-align: left;
    padding: 8px 15px;
    background: none;
    border: none;
    color: var(--thy-text-color);
    cursor: pointer;
    font-size: 13px;
    white-space: nowrap;
    transition: background-color 0.15s ease;
    outline: none;
  }

  .submenu-item:hover {
    background-color: var(--thy-light-gray);
  }

  .submenu-separator {
    height: 1px;
    background-color: var(--thy-medium-gray);
    margin: 5px 0;
  }

  .submenu-group {
    position: relative;
  }

  .submenu-group-header {
    display: flex;
    align-items: center;
    gap: 8px;
    width: 100%;
    text-align: left;
    padding: 8px 15px;
    background: none;
    border: none;
    color: var(--thy-text-color);
    cursor: default;
    font-size: 13px;
    font-weight: 600;
    white-space: nowrap;
    outline: none;
  }

  .submenu-group-header i {
    margin-right: 5px;
  }

  .nested-submenu {
    list-style: none;
    padding: 0;
    margin-left: 20px;
  }

  /* Toolbar */
  .toolbar {
    height: 40px;
    background: var(--thy-white);
    display: flex;
    align-items: center;
    gap: 0.8rem;
    padding: 0 1.5rem;
    border-bottom: 1px solid var(--thy-medium-gray);
    box-shadow: var(--thy-shadow-light);
    flex-shrink: 0;
    z-index: 5;
  }

  .toolbar-separator {
    width: 1px;
    height: 60%;
    background-color: var(--thy-medium-gray);
    margin: 0 0.6rem;
  }

  /* Main Area */
  .main-area {
    flex-grow: 1;
    background: var(--thy-light-gray);
    padding: 1.5rem;
    display: flex;
    position: relative;
    align-items: flex-start;
    justify-content: flex-start;
    overflow: hidden;
  }

  /* --- Pencere Yakalama (Snap Zones) Stilleri --- */
  .snap-zone {
    position: absolute;
    border: 2px dashed rgba(0, 0, 0, 0.2);
    background-color: rgba(0, 0, 0, 0.05);
    z-index: 1;
    pointer-events: none;
    transition: all 0.1s ease-out;
  }

  .snap-zone.hovered {
    background-color: rgba(0, 75, 133, 0.2);
    border-color: var(--thy-secondary-blue);
  }

  .snap-controls {
    display: flex;
    align-items: center;
    gap: 15px;
    color: var(--thy-white);
    font-size: 13px;
  }

  .snap-controls label {
    display: flex;
    align-items: center;
    gap: 5px;
    cursor: pointer;
  }

  .snap-controls input[type="checkbox"] {
    cursor: pointer;
  }

  .snap-controls select {
    padding: 2px 5px;
    border-radius: 4px;
    border: 1px solid rgba(255, 255, 255, 0.3);
    background-color: rgba(255, 255, 255, 0.1);
    color: var(--thy-white);
    font-size: 13px;
    outline: none;
    cursor: pointer;
  }

  .snap-controls select option {
    background-color: var(--thy-primary-red);
    color: var(--thy-white);
  }

  /* Tabs Bar */
  .tabs-bar {
    background-color: var(--thy-light-gray);
    display: flex;
    flex-wrap: nowrap;
    overflow-x: auto;
    -webkit-overflow-scrolling: touch;
    box-shadow: none;
    height: var(--tabs-bar-height-visible);
    flex-shrink: 0;
    padding: var(--tabs-bar-padding-visible);
    align-items: flex-end;
    border-top: none;
    transition: all var(--thy-transition-speed) var(--thy-transition-timing);
    position: relative; /* Bu hala gerekli olabilir ama butonu dışarı taşıdığımız için daha az önemli */
  }

  .tabs-bar::-webkit-scrollbar {
    height: 6px;
  }

  .tabs-bar::-webkit-scrollbar-thumb {
    background-color: rgba(0, 0, 0, 0.2);
    border-radius: 3px;
  }

  .tabs-bar::-webkit-scrollbar-track {
    background: transparent;
  }

  .tabs-bar.hidden {
    height: var(--tabs-bar-height-hidden); /* Yeni değer */
    padding: var(--tabs-bar-padding-hidden);
    overflow: hidden;
    background-color: transparent;
    border-top: none;
    align-items: center;
    pointer-events: none; /* Gizliyken tıklamayı engelle */
  }

  /* Force open on hover kaldırıldı, artık butona bağlı */
  /* .tabs-bar.hidden.force-open {
    height: var(--tabs-bar-height-visible);
    padding: var(--tabs-bar-padding-visible);
    background-color: var(--thy-light-gray);
    overflow-x: auto;
  } */

  .tab-item {
    display: flex;
    align-items: center;
    padding: 4px 8px;
    background-color: var(--thy-medium-gray);
    border: 1px solid var(--thy-border-color);
    border-bottom: none;
    border-radius: 6px 6px 0 0;
    cursor: pointer;
    font-size: 13px;
    color: var(--thy-dark-gray);
    margin-right: 2px;
    white-space: nowrap;
    text-overflow: ellipsis;
    overflow: hidden;
    max-width: 180px;
    transition:
      background-color var(--thy-transition-speed) var(--thy-transition-timing),
      color var(--thy-transition-speed) var(--thy-transition-timing),
      box-shadow var(--thy-transition-speed) var(--thy-transition-timing);
    position: relative;
  }

  .tab-item:hover {
    background-color: var(--thy-light-gray);
    color: var(--thy-text-color);
  }

  .tab-item.active {
    background-color: var(--thy-white);
    color: var(--thy-primary-red);
    border-color: var(--thy-primary-red);
    font-weight: 600;
    z-index: 1;
    margin-bottom: -1px;
    padding-bottom: 5px;
  }

  /* Style for minimized tabs */
  .tab-item.minimized {
    background-color: var(--thy-light-blue);
    border-color: var(--thy-secondary-blue);
    color: var(--thy-secondary-blue);
    font-style: italic;
    font-weight: 500;
    opacity: 0.8;
  }

  .tab-item.minimized:hover {
    background-color: var(--thy-light-blue);
    color: var(--thy-secondary-blue);
    opacity: 1;
  }

  .tab-item.minimized .tab-close-button {
    color: var(--thy-secondary-blue);
  }
  .tab-item.minimized .tab-close-button:hover {
    background-color: rgba(0, 75, 133, 0.1);
    color: var(--thy-secondary-blue);
  }

  .tab-title {
    margin-right: 8px;
    overflow: hidden;
    text-overflow: ellipsis;
  }

  .tab-close-button {
    background: none;
    border: none;
    color: var(--thy-dark-gray);
    font-size: 0.8em;
    cursor: pointer;
    padding: 2px;
    border-radius: 2px;
    transition:
      background-color var(--thy-transition-speed) var(--thy-transition-timing),
      color var(--thy-transition-speed) var(--thy-transition-timing);
    display: flex;
    align-items: center;
    justify-content: center;
  }

  .tab-close-button:hover {
    background-color: rgba(0, 0, 0, 0.1);
    color: var(--thy-black);
  }

  .tab-item.active .tab-close-button {
    color: var(--thy-primary-red);
  }

  .tab-item.active .tab-close-button:hover {
    background-color: rgba(192, 0, 0, 0.1);
    color: var(--thy-primary-red);
  }

  .no-tabs-message {
    color: var(--thy-dark-gray);
    font-style: italic;
    padding: 0 1rem;
    display: flex;
    align-items: center;
    height: 100%;
  }

  /* Durum Çubuğu (Status Bar) Stilleri - THY Kırmızısı */
  .status-bar {
    height: 32px;
    background: var(--thy-primary-red);
    color: var(--thy-white);
    font-size: 13px;
    display: flex;
    justify-content: space-between; /* To push content to ends */
    align-items: center;
    padding: 0 20px;
    box-shadow: 0 -2px 6px rgba(0, 0, 0, 0.25);
    flex-shrink: 0;
    font-weight: 400;
    z-index: 1; /* Tabs barın üstünde kalması için */
  }

  .status-text {
    display: flex;
    align-items: center;
    gap: 8px;
  }

  .status-icon {
    color: var(--thy-white);
    font-size: 1.1em;
  }

  .footer-right-section {
    display: flex;
    align-items: center;
    gap: 15px; /* Spacing between elements on the right */
  }

  .log-button {
    background: transparent;
    border: 1px solid rgba(255, 255, 255, 0.3);
    color: var(--thy-white);
    padding: 4px 10px;
    border-radius: 4px;
    cursor: pointer;
    font-size: 12px;
    font-weight: 500;
    display: flex;
    align-items: center;
    gap: 5px;
    transition:
      background-color 0.2s ease,
      border-color 0.2s ease;
    outline: none;
  }

  .log-button:hover {
    background-color: rgba(255, 255, 255, 0.15);
    border-color: var(--thy-white);
  }

  .log-button:active {
    background-color: rgba(255, 255, 255, 0.25);
  }

  .company-logo {
    height: 24px; /* Adjust size as needed */
    width: auto;
    vertical-align: middle;
    border-radius: 2px; /* Slightly rounded corners for the logo */
  }
  .ghost-dock-zone {
    position: absolute;
    top: 0;
    bottom: 0;
    width: 50%;
    background-color: rgba(0, 123, 255, 0.1);
    z-index: 1;
    pointer-events: none;
    transition: background-color 0.2s ease;
  }

  .ghost-dock-zone.left {
    left: 0;
  }

  .ghost-dock-zone.right {
    right: 0;
  }

  /* Tabs Toggle Button: app-wrapper'a göre konumlandırıldı */
  .tabs-toggle-button {
    position: absolute;
    right: 16px;
    bottom: 33px;
    width: 28px;
    height: 28px;
    background: var(--thy-primary-red); /* THY Mavisi olsun */
    border: 1px solid var(--thy-white);
    border-radius: 50%;
    cursor: pointer;
    display: flex;
    align-items: center;
    justify-content: center;
    box-shadow: var(--thy-shadow-medium);
    transition: all 0.3s ease-in-out; /* Görünürlük ve transform için geçiş */
    z-index: 100; /* Her şeyin üstünde olması için */
    opacity: 0; /* Başlangıçta gizli */
    visibility: hidden; /* Click events almasını engelle */
  }

  .tabs-toggle-button.visible {
    opacity: 1;
    visibility: visible;
  }

  .tabs-toggle-button:hover {
    background-color: var(--thy-primary-red); /* Hover'da kırmızı olsun */
    box-shadow: var(--thy-shadow-strong);
  }

  .tabs-toggle-button i {
    transition: transform 0.3s ease;
    color: var(--thy-white); /* İkon beyaz olsun */
    font-size: 1.2em; /* İkonu biraz büyüt */
  }

  .tabs-toggle-button i.rotated {
    transform: rotate(180deg); /* Açıkken yukarı, kapalıyken aşağı ok */
  }

  /* Responsive Ayarlamalar */
  @media (max-width: 768px) {
    .menu-bar,
    .toolbar,
    .status-bar,
    .tabs-bar {
      padding-left: 1rem;
      padding-right: 1rem;
    }
    .menu-actions {
      gap: 0.6rem;
    }
    .menu-button {
      font-size: 12px;
      padding: 0.2rem 0.4rem;
    }
    .toolbar {
      gap: 0.6rem;
    }
    .main-area {
      padding: 1rem;
    }
    .tabs-bar {
      --tabs-bar-height-visible: 28px; /* Adjust for smaller screens */
    }
    .tab-item {
      font-size: 12px;
      padding: 4px 8px;
      max-width: 150px;
    }
    .tab-title {
      margin-right: 5px;
    }
    .tab-close-button {
      font-size: 0.7em;
    }
    .footer-right-section {
      gap: 10px; /* Adjust spacing for smaller screens */
    }
    .log-button {
      font-size: 11px;
      padding: 3px 8px;
    }
    .tabs-toggle-button {
      right: 10px;
      bottom: 10px;
      width: 32px;
      height: 32px;
    }
  }

  @media (max-width: 480px) {
    .menu-bar {
      flex-direction: column;
      height: auto;
      padding-top: 5px;
      padding-bottom: 5px;
      align-items: flex-start;
    }
    .menu-actions {
      margin-top: 0;
      flex-wrap: wrap;
      justify-content: center;
      width: 100%;
      gap: 0.4rem;
    }
    .toolbar {
      flex-wrap: wrap;
      height: auto;
      padding-top: 5px;
      padding-bottom: 5px;
      gap: 0.6rem;
      justify-content: center;
    }
    .toolbar-separator {
      display: none;
    }
    .main-area {
      padding: 0.8rem;
    }
    .tabs-bar {
      --tabs-bar-height-visible: 26px; /* Further adjust for mobile */
      padding: 0 0.2rem;
    }
    .tab-item {
      font-size: 11px;
      padding: 3px 6px;
      max-width: 120px;
    }
    .tab-close-button {
      font-size: 0.7em;
    }
    .status-bar {
      flex-direction: column;
      height: auto;
      padding-top: 3px;
      padding-bottom: 3px;
      gap: 3px;
      font-size: 12px;
    }
    .footer-right-section {
      width: 100%; /* Take full width on small screens */
      justify-content: center; /* Center items in footer */
      flex-wrap: wrap;
    }
    .company-logo {
      height: 20px; /* Smaller logo for mobile */
    }
    .tabs-toggle-button {
      right: 8px;
      bottom: 8px;
      width: 30px;
      height: 30px;
    }
  }
  /* dikey ayraç */
  .vertical-separator {
    width: 1px;
    height: 16px; /* isteğe göre 20-24px yapabilirsiniz */
    background-color: var(--thy-white);
    opacity: 0.6;
  }

  .gmt-link {
    cursor: pointer;
  }
  .gmt-link:focus-visible {
    outline: 2px dashed var(--thy-white);
    outline-offset: 2px;
  }
  .time {
    min-width: 140px;
  }
</style>
