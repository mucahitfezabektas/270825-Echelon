<script lang="ts">
  import interact from "@interactjs/interactjs";
  import { onMount, onDestroy } from "svelte";
  import type { Interactable } from "@interactjs/core/Interactable";
  import { toolbarActions, resetToolbarActions } from "@/stores/toolbarStates";
  import {
    allWindowConfigs,
    getWindowConfig,
    type WindowConfig, // Ensure WindowConfig is imported
  } from "@/config/windowConfig";
  import { createEventDispatcher, type EventDispatcher } from "svelte";
  import { get } from "svelte/store";
  import {
    menuStructure,
    resetMenuStructureToDefaults,
    isMenuItemAlreadyInStructure,
    type MenuItem,
  } from "@/stores/menuStructureStore";

  import TreeMenuItem from "@/components/TreeMenuItem.svelte";

  // Define dispatched events for CustomizeWindow
  interface CustomizeWindowEvents {
    close: undefined;
  }

  const dispatch: EventDispatcher<CustomizeWindowEvents> =
    createEventDispatcher();

  // Helper function to map a WindowConfig ID to a menu item format.
  const mapConfigToMenuItem = (id: string): MenuItem | undefined => {
    const config = getWindowConfig(id);
    if (config) {
      return {
        id: config.id,
        label: config.title,
        icon: config.icon,
        type: "item", // Make sure MenuItem type also aligns with WindowConfig.type if it's an item
      };
    }
    return undefined;
  };

  // Reactive state for selected tab
  let selectedTab: "toolbars" | "menu" = "toolbars";

  // HTML element references for Interact.js bind:this
  let toolbarDefaultElement: HTMLElement;
  let menuStructureElement: HTMLElement;
  let actionsListElement: HTMLElement;

  // Define SeparatorItem and ToolbarItem types
  type SeparatorItem = { id: string; type: "separator" };
  // ToolbarItem can be a WindowConfig (representing an action) or a SeparatorItem
  // WindowConfig now *must* have a 'type' property defined in its interface (e.g., type: "item")
  type ToolbarItem = WindowConfig | SeparatorItem;

  let stagedToolbarActions: ToolbarItem[] = [];
  let initialToolbarActions: ToolbarItem[] = [];
  let selectedToolbarAction: ToolbarItem | null = null; // Can be a WindowConfig or SeparatorItem

  let currentMenuStructure: MenuItem[] = [];
  let initialMenuStructure: MenuItem[] = [];

  // Variables to hold the currently selected item in each list
  let selectedMenuAction: MenuItem | null = null;
  let selectedAvailableAction: WindowConfig | null = null; // This should only be WindowConfig

  // Interact.js interactable instances
  let interactableToolbarTarget: Interactable;
  let interactableMenuTarget: Interactable;
  let interactableActionItems: Interactable;
  let interactableToolbarItems: Interactable;
  let interactableMenuItems: Interactable;

  // --- Helper Functions for Menu Tree Manipulation ---

  /**
   * Clears all active selections in the customize panel.
   */
  function clearSelections(): void {
    selectedToolbarAction = null;
    selectedMenuAction = null;
    selectedAvailableAction = null;
  }

  /**
   * Recursive helper to find a menu item by ID.
   * @param items The current array of menu items to search in.
   * @param id The ID of the item to find.
   * @returns The MenuItem if found, otherwise undefined.
   */
  function findMenuItem(items: MenuItem[], id: string): MenuItem | undefined {
    for (const item of items) {
      if (item.id === id) {
        return item;
      }
      if (item.type === "submenu" && item.children) {
        const found = findMenuItem(item.children, id);
        if (found) return found;
      }
    }
    return undefined;
  }

  /**
   * Recursive helper to remove a menu item by ID, returning a new array.
   * This ensures immutability for Svelte's reactivity.
   * @param items The current array of menu items.
   * @param id The ID of the item to remove.
   * @returns A new array with the item removed, or the original array if not found.
   */
  function removeMenuItem(items: MenuItem[], id: string): MenuItem[] {
    return items.reduce((acc: MenuItem[], item) => {
      if (item.id === id) {
        return acc; // Skip this item (remove it)
      }
      if (item.type === "submenu" && item.children) {
        const updatedChildren = removeMenuItem(item.children, id);
        // Only create a new submenu object if children actually changed
        if (updatedChildren !== item.children) {
          acc.push({ ...item, children: updatedChildren });
        } else {
          acc.push(item);
        }
      } else {
        acc.push(item);
      }
      return acc;
    }, []);
  }

  function addSeparatorToToolbar() {
    const separatorId = `separator-${Date.now()}`;
    stagedToolbarActions = [
      ...stagedToolbarActions,
      { id: separatorId, type: "separator" },
    ];
  }

  /**
   * Recursive helper to find a menu item's parent array and its index within that array.
   * @param items The current array of menu items to search in.
   * @param id The ID of the item to find.
   * @param currentSearchArray The array currently being searched (for correct parentArray assignment).
   * @returns An object containing the parent array and the index within it, or null/undefined if not found.
   */
  function findMenuItemParentAndIndex(
    items: MenuItem[],
    id: string,
    currentSearchArray: MenuItem[] | null = null
  ): { parentArray: MenuItem[] | null; indexInParent: number } {
    const arrayToSearch = currentSearchArray || items;
    for (let i = 0; i < arrayToSearch.length; i++) {
      const item = arrayToSearch[i];
      if (item.id === id) {
        return { parentArray: arrayToSearch, indexInParent: i };
      }
      if (item.type === "submenu" && item.children) {
        const found = findMenuItemParentAndIndex(
          item.children,
          id,
          item.children
        );
        if (found.parentArray) return found;
      }
    }
    return { parentArray: null, indexInParent: -1 };
  }

  /**
   * Recursive helper to update the menu structure with a new children array for a specific parent.
   * Ensures immutability by creating new objects/arrays at each level affected.
   * @param tree The full menu tree.
   * @param parentId The ID of the parent whose children array needs to be updated (null for root).
   * @param newChildren The new children array to set.
   * @returns A new menu tree with the updated children array.
   */
  function updateMenuStructureWithNewChildren(
    tree: MenuItem[],
    parentId: string | null,
    newChildren: MenuItem[]
  ): MenuItem[] {
    if (parentId === null) {
      return newChildren; // Root level is being updated
    }
    return tree.map((node) => {
      if (node.id === parentId && node.type === "submenu") {
        return { ...node, children: newChildren };
      }
      if (node.type === "submenu" && node.children) {
        const updatedChildren = updateMenuStructureWithNewChildren(
          node.children,
          parentId,
          newChildren
        );
        if (updatedChildren !== node.children) {
          return { ...node, children: updatedChildren };
        }
      }
      return node;
    });
  }

  /**
   * Helper to find the parent MenuItem given its children array reference.
   * @param fullTree The entire menu structure.
   * @param targetArray The children array whose parent we are looking for.
   * @returns The parent MenuItem, or undefined if not found (e.g., if targetArray is the root).
   */
  function findParentOfArray(
    fullTree: MenuItem[],
    targetArray: MenuItem[]
  ): MenuItem | undefined {
    for (const item of fullTree) {
      if (item.type === "submenu" && item.children === targetArray) {
        return item;
      }
      if (item.type === "submenu" && item.children) {
        const found = findParentOfArray(item.children, targetArray);
        if (found) return found;
      }
    }
    return undefined;
  }

  /**
   * Inserts a menu item into the tree at a specified parent and index.
   * @param tree The full menu tree.
   * @param parentId The ID of the parent to insert into (null for root).
   * @param itemToInsert The MenuItem to insert.
   * @param insertIndex The index at which to insert.
   * @returns A new menu tree with the item inserted.
   */
  function findParentAndInsertItem(
    tree: MenuItem[],
    parentId: string | null,
    itemToInsert: MenuItem,
    insertIndex: number
  ): MenuItem[] {
    if (parentId === null) {
      const newTree = [...tree];
      newTree.splice(insertIndex, 0, itemToInsert);
      return newTree;
    }

    return tree.map((node) => {
      if (node.id === parentId && node.type === "submenu") {
        const newChildren = [...(node.children || [])];
        newChildren.splice(insertIndex, 0, itemToInsert);
        return { ...node, children: newChildren };
      }
      if (node.type === "submenu" && node.children) {
        const updatedChildren = findParentAndInsertItem(
          node.children,
          parentId,
          itemToInsert,
          insertIndex
        );
        if (updatedChildren !== node.children) {
          return { ...node, children: updatedChildren };
        }
      }
      return node;
    });
  }

  /**
   * Recursively removes a menu item from the tree and returns the updated tree.
   * @param tree The full menu tree.
   * @param itemId The ID of the item to remove.
   * @returns A new menu tree with the item removed.
   */
  function removeMenuItemFromTree(
    tree: MenuItem[],
    itemId: string
  ): MenuItem[] {
    const newTree = tree.filter((item) => item.id !== itemId);
    if (newTree.length === tree.length) {
      // Item not found at current level, check children
      return tree.map((item) => {
        if (item.type === "submenu" && item.children) {
          const updatedChildren = removeMenuItemFromTree(item.children, itemId);
          if (updatedChildren !== item.children) {
            return { ...item, children: updatedChildren };
          }
        }
        return item;
      });
    }
    return newTree;
  }

  // --- End Helper Functions ---

  // Refined selectAction to use specific types for `action` parameter
  function selectAction(
    action: ToolbarItem | MenuItem | WindowConfig | null,
    type: "toolbar" | "menu" | "available"
  ) {
    clearSelections(); // Clear all selections before setting a new one

    if (type === "toolbar") {
      selectedToolbarAction = action as ToolbarItem; // Could be WindowConfig or SeparatorItem
    } else if (type === "menu") {
      selectedMenuAction = action as MenuItem;
    } else if (type === "available") {
      selectedAvailableAction = action as WindowConfig;
    }
  }

  // Generic function to move an item within a flat array
  function moveItemInArray<T extends { id: string }>(
    arr: T[],
    itemId: string,
    direction: "up" | "down"
  ): T[] {
    const index = arr.findIndex((item) => item.id === itemId);
    if (index === -1) return arr; // Item not found

    const newArr = [...arr]; // Copy the array to ensure reactivity
    const itemToMove = newArr[index];

    if (direction === "up") {
      if (index === 0) return arr; // Already at the top
      newArr.splice(index, 1); // Remove from current position
      newArr.splice(index - 1, 0, itemToMove); // Insert at new position
    } else {
      // direction === 'down'
      if (index === newArr.length - 1) return arr; // Already at the bottom
      newArr.splice(index, 1);
      newArr.splice(index + 1, 0, itemToMove);
    }
    return newArr;
  }

  function addSelectedToToolbar(): void {
    if (selectedAvailableAction) {
      // Check if it's already in stagedToolbarActions (regardless of type)
      if (
        !stagedToolbarActions.some((a) => a.id === selectedAvailableAction?.id)
      ) {
        // Here, selectedAvailableAction is always WindowConfig, so it's safe to add
        stagedToolbarActions = [
          ...stagedToolbarActions,
          selectedAvailableAction,
        ];
        selectedToolbarAction = selectedAvailableAction; // selectedToolbarAction can be WindowConfig
        selectedAvailableAction = null; // Clear available selection after adding
      }
    }
  }

  function removeSelectedFromToolbar(): void {
    if (selectedToolbarAction) {
      stagedToolbarActions = stagedToolbarActions.filter(
        (a) => a.id !== selectedToolbarAction?.id
      );
      selectedToolbarAction = null; // Clear toolbar selection after removing
    }
  }

  function handleMoveToolbarItem(direction: "up" | "down") {
    if (selectedToolbarAction) {
      stagedToolbarActions = moveItemInArray(
        stagedToolbarActions,
        selectedToolbarAction.id,
        direction
      );
      // Ensure the selected item remains selected after reordering
      const reselected = stagedToolbarActions.find(
        (a) => a.id === selectedToolbarAction?.id
      );
      if (reselected) {
        selectedToolbarAction = reselected;
      }
    }
  }

  function handleRevertToDefaults() {
    resetToolbarActions();
    // Assuming toolbarActions store contains WindowConfig items with type: "item"
    // Filter and cast to ensure type safety for stagedToolbarActions
    stagedToolbarActions = get(toolbarActions).map((item) => ({
      ...item,
      type: "item",
    })) as ToolbarItem[];
    clearSelections(); // Clear all selections
    console.log("Toolbar actions reverted to defaults.");
  }

  function handleApplyChanges() {
    toolbarActions.set([...stagedToolbarActions]); // artık separator'ları da içeriyor
    menuStructure.set([...currentMenuStructure]);
    console.log("Applied toolbar actions:", stagedToolbarActions);
    console.log("Applied menu structure:", currentMenuStructure);
    dispatch("close");
  }

  function handleCancelChanges() {
    stagedToolbarActions = [...initialToolbarActions];
    currentMenuStructure = JSON.parse(JSON.stringify(initialMenuStructure)); // Deep copy to revert
    clearSelections(); // Clear all selections
    console.log("Changes cancelled. State reverted to initial.");
    dispatch("close");
  }

  function addSelectedToMenu(): void {
    if (
      selectedAvailableAction &&
      !isMenuItemAlreadyInStructure(
        currentMenuStructure,
        selectedAvailableAction.id
      )
    ) {
      const newMenuItem = mapConfigToMenuItem(selectedAvailableAction.id);
      if (newMenuItem) {
        const itemToAdd = { ...newMenuItem, expanded: false }; // Default to collapsed

        // If a submenu is selected, add the new item into that submenu
        if (selectedMenuAction && selectedMenuAction.type === "submenu") {
          const addIntoSubmenu = (
            items: MenuItem[],
            targetId: string
          ): MenuItem[] => {
            return items.map((item) => {
              if (item.id === targetId && item.type === "submenu") {
                return {
                  ...item,
                  children: [...(item.children || []), itemToAdd],
                  expanded: true, // Expand parent submenu when adding
                };
              }
              if (item.type === "submenu" && item.children) {
                return {
                  ...item,
                  children: addIntoSubmenu(item.children, targetId),
                };
              }
              return item;
            });
          };
          currentMenuStructure = addIntoSubmenu(
            currentMenuStructure,
            selectedMenuAction.id
          );
        }
        // If no submenu is selected, or a regular item is selected, add to the root level
        else {
          currentMenuStructure = [...currentMenuStructure, itemToAdd];
        }
        selectedMenuAction = itemToAdd;
        selectedAvailableAction = null; // Clear available selection after adding
      }
    } else if (selectedAvailableAction) {
      alert(
        `"${selectedAvailableAction.title}" is already in the menu structure.`
      );
    }
  }

  function removeSelectedFromMenu(): void {
    if (selectedMenuAction) {
      currentMenuStructure = removeMenuItemFromTree(
        currentMenuStructure,
        selectedMenuAction.id
      );
      selectedMenuAction = null; // Clear menu selection after removing
    }
  }

  function handleAddSubmenu(): void {
    const submenuLabel = prompt("Please enter the name for the new submenu:");
    if (submenuLabel === null) {
      return; // User cancelled
    }
    const trimmedLabel = submenuLabel.trim();
    if (trimmedLabel === "") {
      alert("Submenu name cannot be empty.");
      return;
    }

    // Generate a simple ID from the label, or use a more robust UUID for production
    const newSubmenuId = `submenu-${trimmedLabel.replace(/\s+/g, "-").toLowerCase()}`;

    if (isMenuItemAlreadyInStructure(currentMenuStructure, newSubmenuId)) {
      alert(
        "A menu item with this name already exists. Please choose a different name."
      );
      return;
    }

    const newSubmenu: MenuItem = {
      id: newSubmenuId,
      label: trimmedLabel,
      type: "submenu", // Submenu has type "submenu"
      children: [],
      expanded: true, // New submenus are expanded by default
    };

    if (selectedMenuAction && selectedMenuAction.type === "submenu") {
      const addIntoSubmenu = (
        items: MenuItem[],
        targetId: string
      ): MenuItem[] => {
        return items.map((item) => {
          if (item.id === targetId && item.type === "submenu") {
            return {
              ...item,
              children: [...(item.children || []), newSubmenu],
              expanded: true,
            };
          }
          if (item.type === "submenu" && item.children) {
            return {
              ...item,
              children: addIntoSubmenu(item.children, targetId),
            };
          }
          return item;
        });
      };
      currentMenuStructure = addIntoSubmenu(
        currentMenuStructure,
        selectedMenuAction.id
      );
    } else {
      currentMenuStructure = [...currentMenuStructure, newSubmenu];
    }

    selectedMenuAction = newSubmenu;
    selectedAvailableAction = null;
  }

  function handleToggleMenuExpand(
    itemId: string,
    currentTree: MenuItem[],
    forceExpand?: boolean // Alt menüyü zorla açmak için opsiyonel parametre
  ): MenuItem[] {
    const toggleExpansion = (
      items: MenuItem[],
      targetId: string
    ): MenuItem[] => {
      return items.map((item) => {
        if (item.id === targetId && item.type === "submenu") {
          // Eğer forceExpand belirtildiyse onu kullan, yoksa mevcut durumu tersine çevir
          const newExpandedState =
            forceExpand !== undefined ? forceExpand : !item.expanded;
          return { ...item, expanded: newExpandedState };
        }
        if (item.type === "submenu" && item.children) {
          const updatedChildren = toggleExpansion(item.children, targetId);
          if (updatedChildren !== item.children) {
            return {
              ...item,
              children: updatedChildren,
            };
          }
        }
        return item;
      });
    };
    return toggleExpansion(currentTree, itemId);
  }

  function handleMoveMenuItem(direction: "up" | "down") {
    if (!selectedMenuAction) return;

    const itemIdToMove = selectedMenuAction.id;

    // 1. Mevcut menü yapısının derin bir kopyasını oluştur.
    let newTree = JSON.parse(JSON.stringify(currentMenuStructure));

    // 2. Öğenin *newTree* içindeki mevcut konumunu ve ebeveynini bul.
    const {
      parentArray: currentParentArrayInNewTree,
      indexInParent: currentItemIndexInNewTree,
    } = findMenuItemParentAndIndex(newTree, itemIdToMove);

    if (!currentParentArrayInNewTree || currentItemIndexInNewTree === -1) {
      console.warn(
        "Seçilen menü öğesi veya onun newTree içindeki ebeveyn dizisi bulunamadı."
      );
      return;
    }

    const itemToMove = currentParentArrayInNewTree[currentItemIndexInNewTree];

    // Öğenin mevcut dizisinin ebeveyninin ID'sini belirle. Eğer kök öğesiyse null olur.
    const currentParentOfItem = findParentOfArray(
      newTree,
      currentParentArrayInNewTree
    );
    const currentParentOfItemId = currentParentOfItem?.id || null;

    // 3. Öğeyi mevcut konumundan newTree'den kaldır.
    newTree = removeMenuItemFromTree(newTree, itemIdToMove);

    let targetParentId: string | null = null;
    let insertIndex: number;

    if (direction === "up") {
      if (currentItemIndexInNewTree === 0) {
        // Öğenin mevcut dizisindeki ilk öğe ise
        if (currentParentOfItem && currentParentOfItem.type === "submenu") {
          // Durum 1: Bir alt menünün ilk çocuğu -> alt menünün ebeveyn seviyesine taşı (alt menüden önce)
          const {
            parentArray: grandparentArray,
            indexInParent: parentIndexInGrandparent,
          } = findMenuItemParentAndIndex(
            newTree,
            currentParentOfItemId as string
          ); // currentParentOfItem'ın varlığını kontrol ettik, bu yüzden string olarak cast edebiliriz

          if (grandparentArray) {
            targetParentId =
              findParentOfArray(newTree, grandparentArray)?.id || null;
            insertIndex = parentIndexInGrandparent;
          } else {
            console.error(
              "Alt menü için büyük ebeveyn dizisi bulunamadı, bu beklenmedik bir durum."
            );
            return;
          }
        } else {
          // Durum 2: Kök seviyesindeki veya ebeveyni olmayan bir dizideki ilk öğe -> yukarı taşınamaz
          currentMenuStructure = newTree; // Değişiklik olmasa bile tepkiselliği tetikle
          return;
        }
      } else {
        // Mevcut dizisindeki ilk öğe değil
        const itemAboveId =
          currentParentArrayInNewTree[currentItemIndexInNewTree - 1].id;
        const itemAboveInNewTree = findMenuItem(newTree, itemAboveId); // *newTree*'den öğeyi al

        if (itemAboveInNewTree && itemAboveInNewTree.type === "submenu") {
          // Durum 3: Doğrudan yukarıdaki alt menünün içine taşı
          targetParentId = itemAboveInNewTree.id;
          insertIndex = itemAboveInNewTree.children?.length || 0; // Alt menünün çocuklarının sonuna ekle
          // Hedef alt menüyü zorla aç
          newTree = handleToggleMenuExpand(
            itemAboveInNewTree.id,
            newTree,
            true
          );
        } else {
          // Durum 4: Mevcut ebeveyn dizisi içinde normal şekilde taşı
          targetParentId = currentParentOfItemId;
          insertIndex = currentItemIndexInNewTree - 1; // Aynı dizide bir konum yukarı taşı
        }
      }
    } else {
      // direction === "down"
      if (
        currentItemIndexInNewTree ===
        currentParentArrayInNewTree.length - 1
      ) {
        // Öğenin mevcut dizisindeki son öğe ise
        if (currentParentOfItem && currentParentOfItem.type === "submenu") {
          // Durum 5: Bir alt menünün son çocuğu -> alt menünün ebeveyn seviyesine taşı (alt menüden sonra)
          const {
            parentArray: grandparentArray,
            indexInParent: parentIndexInGrandparent,
          } = findMenuItemParentAndIndex(
            newTree,
            currentParentOfItemId as string
          );

          if (grandparentArray) {
            targetParentId =
              findParentOfArray(newTree, grandparentArray)?.id || null;
            insertIndex = parentIndexInGrandparent + 1;
          } else {
            console.error(
              "Alt menü için büyük ebeveyn dizisi bulunamadı, bu beklenmedik bir durum."
            );
            return;
          }
        } else {
          // Durum 6: Kök seviyesindeki veya ebeveyni olmayan bir dizideki son öğe -> aşağı taşınamaz
          currentMenuStructure = newTree; // Tepkiselliği tetikle
          return;
        }
      } else {
        // Mevcut dizisindeki son öğe değil
        const itemBelowId =
          currentParentArrayInNewTree[currentItemIndexInNewTree + 1].id;
        const itemBelowInNewTree = findMenuItem(newTree, itemBelowId); // *newTree*'den öğeyi al

        if (itemBelowInNewTree && itemBelowInNewTree.type === "submenu") {
          // Durum 7: Doğrudan aşağıdaki alt menünün içine taşı
          targetParentId = itemBelowInNewTree.id;
          insertIndex = 0; // Alt menünün çocuklarının başına ekle
          // Hedef alt menüyü zorla aç
          newTree = handleToggleMenuExpand(
            itemBelowInNewTree.id,
            newTree,
            true
          );
        } else {
          // Durum 8: Mevcut ebeveyn dizisi içinde normal şekilde taşı
          targetParentId = currentParentOfItemId;
          insertIndex = currentItemIndexInNewTree + 1; // Aynı dizide bir konum aşağı taşı
        }
      }
    }

    // Öğeyi yeni konumuna ekle
    newTree = findParentAndInsertItem(
      newTree,
      targetParentId,
      itemToMove,
      insertIndex
    );

    // Svelte'in tepkisel durumunu yeni ağaçla güncelle
    currentMenuStructure = newTree;
    // Odağı korumak için öğeyi yeniden seç
    selectedMenuAction =
      findMenuItem(currentMenuStructure, itemIdToMove) || null;
    // No need to explicitly re-assign currentMenuStructure = currentMenuStructure; here,
    // as it's already updated by the helper functions, triggering reactivity.
  }

  function handleRevertMenuToDefaults() {
    resetMenuStructureToDefaults();
    currentMenuStructure = get(menuStructure); // This is already a deep copy from the store's default
    clearSelections(); // Clear all selections
    console.log("Menu structure reverted to defaults.");
  }

  onMount(() => {
    initialToolbarActions = get(toolbarActions); // artık ToolbarItem[]
    stagedToolbarActions = [...initialToolbarActions];

    const currentMenuData = get(menuStructure);
    currentMenuStructure = JSON.parse(JSON.stringify(currentMenuData));
    initialMenuStructure = JSON.parse(JSON.stringify(currentMenuData));

    if (
      !toolbarDefaultElement ||
      !actionsListElement ||
      !menuStructureElement
    ) {
      console.warn(
        "Required DOM elements for Interact.js not found. Initialization aborted."
      );
      return;
    }

    // --- Interact.js Configuration ---

    // Draggable for "Available Actions"
    interactableActionItems = interact(".action-item").draggable({
      inertia: true,
      autoScroll: true,
      listeners: {
        start(event) {
          const originalTarget = event.target as HTMLElement;
          originalTarget.classList.add("dragging");

          const clone = originalTarget.cloneNode(true) as HTMLElement;
          clone.classList.add("dragging-clone");
          clone.style.position = "fixed";
          clone.style.zIndex = "9999";
          document.body.appendChild(clone);
          const rect = originalTarget.getBoundingClientRect();
          clone.style.left = `${rect.left}px`;
          clone.style.top = `${rect.top}px`;
          clone.dataset.x = "0";
          clone.dataset.y = "0";

          originalTarget.style.opacity = "0.5";
          originalTarget.style.visibility = "hidden";

          event.drag.target = clone; // Drag the clone
        },
        move(event) {
          const target = event.target as HTMLElement;
          const x = parseFloat(target.dataset.x || "0") + event.dx;
          const y = parseFloat(target.dataset.y || "0") + event.dy;
          target.style.transform = `translate(${x}px, ${y}px)`;
          target.dataset.x = x.toString();
          target.dataset.y = y.toString();
        },
        end(event) {
          const draggedItemId = (event.target as HTMLElement).dataset.itemId;
          const original = document.querySelector(
            `.action-item[data-item-id="${draggedItemId}"]:not(.dragging-clone)`
          ) as HTMLElement;
          if (original) {
            original.style.opacity = "1";
            original.style.visibility = "visible";
          }
          if (
            (event.target as HTMLElement).classList.contains("dragging-clone")
          ) {
            (event.target as HTMLElement).remove();
          }
          (event.target as HTMLElement).classList.remove("dragging");
          (event.target as HTMLElement).style.transform = "translate(0px, 0px)";
          delete (event.target as HTMLElement).dataset.x;
          delete (event.target as HTMLElement).dataset.y;
        },
      },
    });

    // Dropzone for "Toolbar"
    interactableToolbarTarget = interact(toolbarDefaultElement).dropzone({
      accept: ".action-item",
      overlap: 0.75,
      ondropactivate: function (event) {
        event.target.classList.add("drop-active");
      },
      ondragenter: function (event) {
        event.relatedTarget.classList.add("can-drop");
      },
      ondragleave: function (event) {
        event.relatedTarget.classList.remove("can-drop");
      },
      ondrop: function (event) {
        const droppedItemId = (event.relatedTarget as HTMLElement).dataset
          .itemId;
        const droppedAction = allWindowConfigs.find(
          (config) => config.id === droppedItemId
        );

        if (selectedTab === "toolbars" && droppedAction) {
          // IMPORTANT: droppedAction is WindowConfig. When adding it to stagedToolbarActions,
          // ensure its 'type' property (now "item") is present.
          if (!stagedToolbarActions.some((a) => a.id === droppedAction.id)) {
            stagedToolbarActions = [
              ...stagedToolbarActions,
              { ...droppedAction, type: "item" },
            ];
            selectedToolbarAction = { ...droppedAction, type: "item" };
            selectedAvailableAction = null;
          } else {
            console.log("Item already in toolbar:", droppedAction.id);
            alert(`"${droppedAction.title}" is already in the toolbar.`);
          }
        }
      },
      ondropdeactivate: function (event) {
        event.target.classList.remove("drop-active");
        (event.relatedTarget as HTMLElement)?.classList.remove("can-drop");
      },
    });

    // Dropzone for "Menu Structure" (root level)
    interactableMenuTarget = interact(menuStructureElement).dropzone({
      accept: ".action-item",
      overlap: 0.75,
      ondropactivate: function (event) {
        event.target.classList.add("drop-active");
      },
      ondragenter: function (event) {
        event.relatedTarget.classList.add("can-drop");
      },
      ondragleave: function (event) {
        event.relatedTarget.classList.remove("can-drop");
      },
      ondrop: function (event) {
        const droppedItemId = (event.relatedTarget as HTMLElement).dataset
          .itemId;
        const droppedAction = allWindowConfigs.find(
          (config) => config.id === droppedItemId
        );

        if (selectedTab === "menu" && droppedAction) {
          if (
            !isMenuItemAlreadyInStructure(
              currentMenuStructure,
              droppedAction.id
            )
          ) {
            const newMenuItem = mapConfigToMenuItem(droppedAction.id);
            if (newMenuItem) {
              const itemToAdd = { ...newMenuItem, expanded: false };
              // Default behavior for dropping onto the main menu box: add to root
              currentMenuStructure = [...currentMenuStructure, itemToAdd];
              selectedMenuAction = itemToAdd;
              selectedAvailableAction = null;
            }
          } else {
            console.log("Item already in menu structure:", droppedAction.id);
            alert(`"${droppedAction.title}" is already in the menu.`);
          }
        }
      },
      ondropdeactivate: function (event) {
        event.target.classList.remove("drop-active");
        (event.relatedTarget as HTMLElement)?.classList.remove("can-drop");
      },
    });

    // Draggable for "Toolbar Items" (for reordering within toolbar)
    interactableToolbarItems = interact(".toolbar-item").draggable({
      inertia: true,
      autoScroll: true,
      listeners: {
        start(event) {
          event.target.classList.add("dragging");
          // Optionally create a clone for smooth dragging without distorting original layout
          // This requires more complex logic for insertion point prediction
        },
        move(event) {
          // If not cloning, just move the original element (can be jumpy)
          // For true smooth reordering, use a clone and then update array on 'end'
        },
        end(event) {
          event.target.classList.remove("dragging");
          // Reordering within the toolbar is handled by the buttons for simplicity
          // For D&D reordering here, you'd need a dropzone on the toolbar items container
          // and calculate the insertion index based on drop position.
        },
      },
    });

    // Draggable for "Menu Items" (for reordering within menu tree)
    interactableMenuItems = interact(".menu-item").draggable({
      inertia: true,
      autoScroll: {
        container: menuStructureElement,
        speed: 300,
        interval: 100,
      },
      listeners: {
        start(event) {
          const originalTarget = event.target as HTMLElement;
          originalTarget.classList.add("dragging");

          const clone = originalTarget.cloneNode(true) as HTMLElement;
          clone.classList.add("dragging-clone");
          clone.style.position = "fixed";
          clone.style.zIndex = "9999";
          document.body.appendChild(clone);
          const rect = originalTarget.getBoundingClientRect();
          clone.style.left = `${rect.left}px`;
          clone.style.top = `${rect.top}px`;
          clone.dataset.x = "0";
          clone.dataset.y = "0";

          originalTarget.style.opacity = "0.5";
          originalTarget.style.visibility = "hidden";

          event.drag.target = clone;
        },
        move(event) {
          const target = event.target as HTMLElement;
          const x = parseFloat(target.dataset.x || "0") + event.dx;
          const y = parseFloat(target.dataset.y || "0") + event.dy;
          target.style.transform = `translate(${x}px, ${y}px)`;
          target.dataset.x = x.toString();
          target.dataset.y = y.toString();
        },
        end(event) {
          const draggedItemId = (event.target as HTMLElement).dataset.itemId;
          const originalElement = document.querySelector(
            `.menu-item[data-item-id="${draggedItemId}"]:not(.dragging-clone)`
          ) as HTMLElement;

          if (originalElement) {
            originalElement.style.opacity = "1";
            originalElement.style.visibility = "visible";
            originalElement.classList.remove("dragging");
          }

          if (
            (event.target as HTMLElement).classList.contains("dragging-clone")
          ) {
            (event.target as HTMLElement).remove();
          }

          const dropzoneElement = event.dropzone?.element as
            | HTMLElement
            | undefined;
          const dropzoneItemId = dropzoneElement?.dataset.itemId; // The ID of the item being dropped ONTO

          if (draggedItemId) {
            // It's crucial to get the item's original state *before* removal from its old spot
            // and then re-insert it. `findMenuItem` ensures we get the correct item object.
            // We need to find the item from the *current* state, not the initial.
            const movedItem = findMenuItem(currentMenuStructure, draggedItemId);
            if (!movedItem) return;

            // Remove the item from its original position first
            currentMenuStructure = removeMenuItemFromTree(
              currentMenuStructure,
              draggedItemId
            );

            let targetParentId: string | null = null;
            let insertIndex: number;

            if (dropzoneElement === menuStructureElement) {
              // Dropped onto the main menu box, add to root
              targetParentId = null;
              insertIndex = currentMenuStructure.length; // Default to end
              const dropYInContainer =
                event.client.y -
                menuStructureElement.getBoundingClientRect().top;

              // Find the correct insertion index at the root level
              const rootItems = Array.from(
                menuStructureElement.querySelectorAll(
                  ".menu-tree-box > ul > .menu-item-wrapper > .menu-item"
                )
              );
              for (let i = 0; i < rootItems.length; i++) {
                const el = rootItems[i] as HTMLElement;
                const rect = el.getBoundingClientRect();
                const elMidYInContainer =
                  rect.top -
                  menuStructureElement.getBoundingClientRect().top +
                  rect.height / 2;

                if (dropYInContainer < elMidYInContainer) {
                  insertIndex = i;
                  break;
                }
              }
            } else if (dropzoneItemId) {
              // Dropped onto a specific menu item
              const targetMenuItem = findMenuItem(
                currentMenuStructure, // Search in the updated structure after removal
                dropzoneItemId
              );
              if (targetMenuItem) {
                if (targetMenuItem.type === "submenu") {
                  // Dropped onto a submenu, add as a child
                  targetParentId = targetMenuItem.id;
                  insertIndex = targetMenuItem.children?.length || 0;
                  // Ensure the target submenu is expanded
                  currentMenuStructure = handleToggleMenuExpand(
                    targetMenuItem.id,
                    currentMenuStructure,
                    true // Force expand
                  );
                } else {
                  // Dropped onto a regular item, insert before or after it in its parent array
                  const { parentArray, indexInParent } =
                    findMenuItemParentAndIndex(
                      currentMenuStructure, // Search in the updated structure after removal
                      dropzoneItemId
                    );
                  if (parentArray) {
                    targetParentId =
                      findParentOfArray(currentMenuStructure, parentArray)
                        ?.id || null;
                    const rect = dropzoneElement.getBoundingClientRect();
                    const dropYRelativeToTarget = event.client.y - rect.top;
                    if (dropYRelativeToTarget < rect.height / 2) {
                      // Insert before
                      insertIndex = indexInParent;
                    } else {
                      // Insert after
                      insertIndex = indexInParent + 1;
                    }
                  } else {
                    // Fallback for root-level items if parentArray is null unexpectedly
                    targetParentId = null;
                    insertIndex = currentMenuStructure.length;
                  }
                }
              } else {
                console.warn(
                  "Dropped onto an unrecognized menu item:",
                  dropzoneItemId
                );
                return;
              }
            } else {
              console.warn("Dropzone not recognized or no item ID.");
              return;
            }

            // Insert the item into its new position
            currentMenuStructure = findParentAndInsertItem(
              currentMenuStructure,
              targetParentId,
              movedItem,
              insertIndex
            );

            // Re-select the moved item
            selectedMenuAction =
              findMenuItem(currentMenuStructure, draggedItemId) || null;
            // No need to explicitly re-assign currentMenuStructure = currentMenuStructure; here,
            // as it's already updated by the helper functions, triggering reactivity.
          }
        },
      },
    });

    // Dropzone for individual menu items (to allow dropping *onto* a submenu or *between* items)
    interact(".menu-item").dropzone({
      accept: ".menu-item", // Accept other menu items
      overlap: 0.1, // Small overlap for hover
      ondropactivate: function (event) {
        event.target.classList.add("drop-target-active");
      },
      ondragenter: function (event) {
        // Prevent dropping an item onto itself
        if (
          event.target.dataset.itemId === event.relatedTarget.dataset.itemId
        ) {
          event.preventDefault();
          return;
        }
        event.target.classList.add("drop-target-hover");
      },
      ondragleave: function (event) {
        event.target.classList.remove("drop-target-hover");
      },
      ondropdeactivate: function (event) {
        event.target.classList.remove("drop-target-active");
        event.target.classList.remove("drop-target-hover");
      },
    });
  });

  onDestroy(() => {
    interactableToolbarTarget?.unset();
    interactableMenuTarget?.unset();
    interactableActionItems?.unset();
    interactableToolbarItems?.unset();
    interactableMenuItems?.unset();
  });

  // Re-render when currentMenuStructure changes
  $: currentMenuStructure;
</script>

<svelte:head>
  <link
    rel="stylesheet"
    href="https://cdnjs.cloudflare.com/ajax/libs/font-awesome/5.15.4/css/all.min.css"
  />
</svelte:head>

<div class="customize-container">
  <div class="tabs">
    <button
      class="tab-button {selectedTab === 'toolbars' ? 'active' : ''}"
      on:click={() => (selectedTab = "toolbars")}>Toolbar</button
    >
    <button
      class="tab-button {selectedTab === 'menu' ? 'active' : ''}"
      on:click={() => (selectedTab = "menu")}>Menubar</button
    >
  </div>

  <div class="tab-content">
    {#if selectedTab === "toolbars"}
      <div class="customize-panel customize-panel-toolbars">
        <div class="list-section toolbar-selected-actions">
          <h4>Toolbar Items</h4>

          <div class="toolbar-list-container">
            <div class="toolbar-items-box" bind:this={toolbarDefaultElement}>
              {#each stagedToolbarActions as action (action.id)}
                {#if action.type === "separator"}
                  <div
                    class="toolbar-item separator {selectedToolbarAction?.id ===
                    action.id
                      ? 'selected'
                      : ''}"
                    on:click={() => selectAction(action, "toolbar")}
                    on:keydown={(e) =>
                      e.key === "Enter" && selectAction(action, "toolbar")}
                    tabindex="0"
                    role="button"
                    data-item-id={action.id}
                  >
                    <i class="fas fa-minus"></i> Separator
                  </div>
                {:else if action.type === "item"}
                  <div
                    class="toolbar-item {selectedToolbarAction?.id === action.id
                      ? 'selected'
                      : ''}"
                    on:click={() => selectAction(action, "toolbar")}
                    on:keydown={(e) =>
                      e.key === "Enter" && selectAction(action, "toolbar")}
                    tabindex="0"
                    role="button"
                    data-item-id={action.id}
                  >
                    <i class={action.icon}></i>
                    {action.title}
                  </div>
                {/if}
              {/each}
            </div>
          </div>
        </div>

        <div class="action-buttons toolbar-action-buttons">
          <button on:click={addSeparatorToToolbar}> Add Separator </button>

          <button
            on:click={addSelectedToToolbar}
            disabled={!selectedAvailableAction}
            ><i class="fas fa-angle-left"></i> Add</button
          >
          <button
            on:click={removeSelectedFromToolbar}
            disabled={!selectedToolbarAction}
            >Remove <i class="fas fa-angle-right"></i></button
          >
          <button
            class="up-down-button"
            on:click={() => handleMoveToolbarItem("up")}
            disabled={!selectedToolbarAction ||
              stagedToolbarActions.indexOf(selectedToolbarAction) === 0}
            title="Move Up"><i class="fas fa-arrow-up"></i> Up</button
          >
          <button
            class="up-down-button"
            on:click={() => handleMoveToolbarItem("down")}
            disabled={!selectedToolbarAction ||
              stagedToolbarActions.indexOf(selectedToolbarAction) ===
                stagedToolbarActions.length - 1}
            title="Move Down"><i class="fas fa-arrow-down"></i> Down</button
          >
          <button class="revert-button" on:click={handleRevertToDefaults}
            >Revert to defaults</button
          >
        </div>

        <div class="list-section toolbar-all-actions">
          <h4>Actions</h4>
          <div class="actions-list-box" bind:this={actionsListElement}>
            {#each allWindowConfigs
              .filter((action) => !stagedToolbarActions.some((a) => a.id === action.id))
              .sort( (a, b) => a.title.localeCompare(b.title) ) as action (action.id)}
              <div
                class="action-item {selectedAvailableAction?.id === action.id
                  ? 'selected'
                  : ''}"
                on:click={() => selectAction(action, "available")}
                on:keydown={(e) => {
                  if (e.key === "Enter" || e.key === " ") {
                    e.preventDefault();
                    selectAction(action, "available");
                  }
                }}
                data-item-id={action.id}
                data-item-type="item"
                role="button"
                tabindex="0"
              >
                <i class={action.icon}></i>
                {action.title}
              </div>
            {/each}
          </div>
        </div>
      </div>
    {:else if selectedTab === "menu"}
      <div class="customize-panel customize-panel-menu">
        <div class="list-section menu-structure-section">
          <h4>Menubar Structure</h4>
          <div class="menu-structure-container">
            <div class="menu-tree-box" bind:this={menuStructureElement}>
              <ul>
                {#each currentMenuStructure as item (item.id)}
                  <TreeMenuItem
                    {item}
                    {selectedMenuAction}
                    {selectAction}
                    on:toggleExpand={(e) =>
                      (currentMenuStructure = handleToggleMenuExpand(
                        e.detail,
                        currentMenuStructure
                      ))}
                  />
                {/each}
              </ul>
            </div>
          </div>
        </div>

        <div class="action-buttons menu-action-buttons">
          <button class="small-btn" on:click={handleAddSubmenu}
            >Add Submenu</button
          >
          <button
            on:click={addSelectedToMenu}
            disabled={!selectedAvailableAction}
            ><i class="fas fa-angle-left"></i> Add</button
          >
          <button
            class="small-btn"
            on:click={removeSelectedFromMenu}
            disabled={!selectedMenuAction}
            >Remove <i class="fas fa-angle-right"></i></button
          >

          <button
            class="up-down-button"
            on:click={() => handleMoveMenuItem("up")}
            disabled={!selectedMenuAction ||
              (findMenuItemParentAndIndex(
                currentMenuStructure,
                selectedMenuAction.id
              ).indexInParent === 0 &&
                !findParentOfArray(
                  currentMenuStructure,
                  findMenuItemParentAndIndex(
                    currentMenuStructure,
                    selectedMenuAction.id
                  ).parentArray || []
                ))}
            title="Move Up"
          >
            <i class="fas fa-arrow-up"></i> Up
          </button>
          <button
            class="up-down-button"
            on:click={() => handleMoveMenuItem("down")}
            disabled={!selectedMenuAction ||
              (findMenuItemParentAndIndex(
                currentMenuStructure,
                selectedMenuAction.id
              ).indexInParent ===
                (findMenuItemParentAndIndex(
                  currentMenuStructure,
                  selectedMenuAction.id
                ).parentArray?.length || 0) -
                  1 &&
                !findParentOfArray(
                  currentMenuStructure,
                  findMenuItemParentAndIndex(
                    currentMenuStructure,
                    selectedMenuAction.id
                  ).parentArray || []
                ))}
            title="Move Down"
          >
            <i class="fas fa-arrow-down"></i> Down
          </button>
          <button class="revert-button" on:click={handleRevertMenuToDefaults}
            >Revert to defaults</button
          >
        </div>

        <div class="list-section menu-all-actions">
          <h4>Actions</h4>
          <div class="actions-list-box">
            {#each allWindowConfigs
              .filter((action) => !isMenuItemAlreadyInStructure(currentMenuStructure, action.id))
              .sort( (a, b) => a.title.localeCompare(b.title) ) as action (action.id)}
              <div
                class="action-item {selectedAvailableAction?.id === action.id
                  ? 'selected'
                  : ''}"
                on:click={() => selectAction(action, "available")}
                on:keydown={(e) => {
                  if (e.key === "Enter" || e.key === " ") {
                    e.preventDefault();
                    selectAction(action, "available");
                  }
                }}
                data-item-id={action.id}
                data-item-type="item"
                role="button"
                tabindex="0"
              >
                <i class={action.icon}></i>
                {action.title}
              </div>
            {/each}
          </div>
        </div>
      </div>
    {/if}
  </div>

  <div class="customize-footer">
    <button class="action-btn ok-btn" on:click={handleApplyChanges}
      >Apply</button
    >
    <button class="action-btn cancel-btn" on:click={handleCancelChanges}
      >Cancel</button
    >
  </div>
</div>

<style>
  @import url("https://fonts.googleapis.com/css2?family=Inter:wght@300;400;500;600;700&display=swap");

  /* Basic Styles */
  :root {
    --thy-off-white: #f5f7fa;
    --thy-white: #ffffff;
    --thy-light-gray: #f0f2f5;
    --thy-medium-gray: #e0e6ed;
    --thy-dark-gray: #cbd5e1;
    --thy-darker-gray: #6b7280;
    --thy-text-color: #374151;
    --thy-secondary-blue: #2563eb;
    --thy-primary-red: #ef4444;
    --thy-shadow-medium: 0 4px 12px rgba(0, 0, 0, 0.1);
    --thy-border-radius: 4px;
    --thy-selection-blue: #c4d9f6; /* Light blue for selection */
    --thy-border-blue: #a0c2e3; /* Border for selected items */
    --thy-tree-line-color: #c0c0c0; /* Ağaç çizgileri için renk */
  }

  .customize-container {
    padding: 15px;
    background-color: var(--thy-off-white);
    font-family: "Inter", sans-serif;
    color: var(--thy-text-color);
    display: flex;
    flex-direction: column;
    height: 100%;
    box-sizing: border-box;
    font-size: 13px; /* Genel font boyutunu düşür */
  }

  .tabs {
    display: flex;
    border-bottom: 1px solid var(--thy-medium-gray);
    margin-bottom: 15px;
  }

  .tab-button {
    background-color: var(--thy-light-gray);
    border: 1px solid var(--thy-medium-gray);
    border-bottom: none;
    padding: 6px 12px; /* Düğme padding'ini küçült */
    cursor: pointer;
    font-weight: 500;
    color: var(--thy-darker-gray);
    border-top-left-radius: var(--thy-border-radius);
    border-top-right-radius: var(--thy-border-radius);
    margin-right: 3px; /* Düğmeler arası boşluğu küçült */
    transition:
      background-color 0.2s ease,
      color 0.2s ease;
    outline: none;
    font-size: 12px; /* Tab buton font boyutunu küçült */
  }

  .tab-button:hover {
    background-color: var(--thy-medium-gray);
  }

  .tab-button.active {
    background-color: var(--thy-white);
    border-bottom: 1px solid var(--thy-white);
    color: var(--thy-text-color);
  }

  .tab-content {
    flex-grow: 1;
    overflow: hidden;
    background-color: var(--thy-white);
    border: 1px solid var(--thy-medium-gray);
    border-top-left-radius: var(--thy-border-radius);
    border-top-right-radius: var(--thy-border-radius);
    padding: 15px;
    display: flex;
  }

  .customize-panel {
    display: flex;
    width: 100%;
    gap: 10px; /* Paneller arası boşluğu küçült */
  }

  /* Specific panel layouts */
  .customize-panel-toolbars,
  .customize-panel-menu {
    display: flex;
    flex: 1; /* Alana yayılsın */
  }

  .list-section {
    flex: 1;
    display: flex;
    flex-direction: column;
    border: 1px solid var(--thy-medium-gray);
    border-radius: var(--thy-border-radius);
    overflow: hidden;
  }

  .list-header-with-buttons {
    display: flex;
    justify-content: space-between;
    align-items: center;
    padding: 8px 10px;
    background-color: var(--thy-light-gray);
    border-bottom: 1px solid var(--thy-medium-gray);
    flex-shrink: 0;
  }

  .header-buttons {
    display: flex;
    gap: 5px;
  }

  .list-section h4 {
    margin-top: 0;
    margin-bottom: 0;
    color: var(--thy-text-color);
    font-weight: 600;
    padding: 8px 10px;
    background-color: var(--thy-light-gray);
    flex-shrink: 0;
  }

  .toolbar-list-container,
  .menu-structure-container {
    background-color: var(--thy-white);
    flex-grow: 1;
    display: flex;
    flex-direction: column;
    overflow: hidden;
  }

  .toolbar-name-display,
  .menu-bar-label {
    background-color: var(--thy-light-gray);
    border-bottom: 1px solid var(--thy-medium-gray);
    padding: 5px 10px;
    font-weight: 500;
    color: var(--thy-darker-gray);
    font-size: 12px;
    flex-shrink: 0;
  }

  .toolbar-items-box,
  .actions-list-box,
  .menu-tree-box {
    flex-grow: 1;
    overflow-y: auto;
    padding: 3px;
  }

  .action-item,
  .toolbar-item {
    padding: 4px 8px;
    margin-bottom: 1px;
    border-radius: 2px;
    cursor: pointer;
    user-select: none;
    display: flex;
    align-items: center;
    gap: 6px;
    transition:
      background-color 0.15s ease,
      border-color 0.15s ease;
    font-size: 12px;
    line-height: 1.4;
  }

  .action-item i,
  .toolbar-item i {
    font-size: 1.1em;
    width: 1.2em;
    text-align: center;
  }

  .action-item:hover,
  .toolbar-item:hover {
    background-color: var(--thy-light-gray);
  }

  .action-item.selected,
  .toolbar-item.selected {
    background-color: var(--thy-selection-blue);
    border: 1px solid var(--thy-border-blue);
    padding: 3px 7px;
  }

  /* Drag and Drop Feedback */
  .drop-active {
    background-color: #e6f7ff;
    border-color: #91d5ff;
  }

  .can-drop {
    box-shadow: 0 0 0 2px var(--thy-secondary-blue);
  }

  .dragging {
    opacity: 0.5;
    cursor: grabbing !important;
  }

  .dragging-clone {
    background-color: var(--thy-white);
    border: 1px solid var(--thy-border-blue);
    box-shadow: var(--thy-shadow-medium);
    pointer-events: none;
    position: absolute;
    z-index: 9999;
    width: auto;
    white-space: nowrap;
    padding: 4px 8px;
    border-radius: 2px;
    display: flex;
    align-items: center;
    gap: 6px;
    font-size: 12px;
  }

  /* Drop target feedback for menu items */
  .menu-item.drop-target-hover {
    background-color: var(--thy-selection-blue);
    border: 1px solid var(--thy-border-blue);
  }

  .menu-item.drop-target-active {
    outline: 2px dashed var(--thy-secondary-blue);
    outline-offset: -2px;
  }

  /* Action Buttons Column */
  .action-buttons {
    display: flex;
    flex-direction: column;
    justify-content: center;
    align-items: center;
    min-width: 90px;
    gap: 5px;
    flex-shrink: 0;
  }

  .action-buttons button {
    background-color: var(--thy-light-gray);
    border: 1px solid var(--thy-medium-gray);
    padding: 16px 8px;
    border-radius: 3px;
    cursor: pointer;
    font-weight: 500;
    color: var(--thy-text-color);
    transition:
      background-color 0.2s ease,
      border-color 0.2s ease;
    width: 80px;
    text-align: center;
    outline: none;
    font-size: 12px;
    display: flex;
    align-items: center;
    justify-content: center;
    gap: 4px;
    height: 24px;
  }

  .action-buttons button i {
    font-size: 0.8em;
  }

  .action-buttons button:hover:not(:disabled) {
    background-color: var(--thy-medium-gray);
  }

  .action-buttons button:disabled {
    opacity: 0.5;
    cursor: not-allowed;
  }
  .action-buttons .revert-button {
    margin-top: 20px;
    background-color: #fcebeb;
    border-color: #fcc2c2;
    color: var(--thy-primary-red);
  }

  .action-buttons .revert-button:hover:not(:disabled) {
    background-color: #fddcdc;
    border-color: #fbafaf;
  }

  /* Menu Specific Styles (Tree Structure) - Global styles for the tree */
  .menu-tree-box ul {
    list-style: none;
    padding-left: 0;
    margin: 0;
  }

  .customize-footer {
    display: flex;
    justify-content: flex-end;
    padding-top: 10px;
    gap: 8px;
  }

  .action-btn {
    background-color: var(--thy-light-gray);
    border: 1px solid var(--thy-medium-gray);
    padding: 6px 15px;
    border-radius: var(--thy-border-radius);
    cursor: pointer;
    font-weight: 500;
    color: var(--thy-text-color);
    transition: background-color 0.2s ease;
    outline: none;
    font-size: 12px;
  }

  .action-btn:hover {
    background-color: var(--thy-medium-gray);
  }

  .ok-btn {
    background-color: var(--thy-secondary-blue);
    color: var(--thy-white);
    border-color: var(--thy-secondary-blue);
  }

  .ok-btn:hover {
    background-color: #1e40af;
  }

  .cancel-btn {
    background-color: var(--thy-white);
    color: var(--thy-text-color);
  }

  .menu-separator-visual {
    width: 100%;
    height: 1px;
    background-color: var(--thy-medium-gray);
    margin: 4px 0;
    opacity: 0.7;
    border: none;
  }

  .toolbar-separator {
    width: 2px;
    height: 60%;
    background-color: var(--thy-medium-gray);
    margin: 0 0.6rem;
  }
</style>
