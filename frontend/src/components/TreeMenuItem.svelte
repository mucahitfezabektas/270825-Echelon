<script lang="ts">
  import type { MenuItem } from "@/stores/menuStructureStore";
  import type { WindowConfig } from "@/config/windowConfig";
  import { createEventDispatcher, type EventDispatcher } from "svelte";

  // Define dispatched events for TreeMenuItem
  interface TreeMenuItemEvents {
    toggleExpand: string; // The detail is the item ID (string)
  }

  const dispatch: EventDispatcher<TreeMenuItemEvents> = createEventDispatcher();

  export let item: MenuItem;
  export let selectedMenuAction: MenuItem | null;
  export let selectAction: (
    action: WindowConfig | MenuItem | null,
    type: "toolbar" | "menu" | "available"
  ) => void;

  let isExpandedState: boolean =
    item.expanded !== undefined ? item.expanded : true;
  // Keep local state in sync with prop for reactivity
  $: isExpandedState = item.expanded !== undefined ? item.expanded : true;

  function toggleExpand(e: MouseEvent | KeyboardEvent) {
    // Added KeyboardEvent type
    e.stopPropagation(); // Prevent item selection when clicking the toggle icon
    if (item.type === "submenu") {
      dispatch("toggleExpand", item.id); // Dispatch event to parent to update model
    }
  }

  // Handle keyboard events for the toggle icon
  function handleToggleKeyDown(e: KeyboardEvent) {
    if (e.key === "Enter" || e.key === " ") {
      e.preventDefault(); // Prevent default scroll/click
      toggleExpand(e);
    }
  }
</script>

<li class="menu-item-wrapper">
  <div
    class="menu-item {item.type === 'submenu'
      ? 'menu-group-label'
      : ''} {selectedMenuAction?.id === item.id ? 'selected' : ''}"
    data-item-id={item.id}
    data-item-type={item.type || "item"}
    role="button"
    tabindex="0"
    on:click={() => selectAction(item, "menu")}
    on:keydown={(e) => {
      if (e.key === "Enter" || e.key === " ") {
        e.preventDefault();
        selectAction(item, "menu");
      }
    }}
  >
    {#if item.type === "submenu"}
      <i
        class="tree-toggle-icon fas {isExpandedState
          ? 'fa-minus-square'
          : 'fa-plus-square'}"
        on:click={toggleExpand}
        on:keydown={handleToggleKeyDown}
        role="button"
        tabindex="0"
        aria-expanded={isExpandedState}
        aria-controls="submenu-{item.id}"
      ></i>
      <i class="tree-folder-icon"></i>
      <span class="item-label">{item.label || item.id}</span>
    {:else if item.type === "separator"}
      <div class="menu-separator-visual"></div>
    {:else}
      <span class="tree-toggle-icon-placeholder"></span>
      <i class="{item.icon || ''} tree-item-icon"></i>
      <span class="item-label">{item.label || item.id}</span>
    {/if}
  </div>

  {#if item.type === "submenu" && isExpandedState}
    <ul class="nested-menu-list" id="submenu-{item.id}" role="group">
      {#each item.children || [] as child (child.id)}
        <svelte:self
          item={child as MenuItem}
          {selectedMenuAction}
          {selectAction}
          on:toggleExpand={(e) => dispatch("toggleExpand", e.detail)}
        />
      {/each}
    </ul>
  {/if}
</li>

<style>
  :root {
    --thy-tree-line-color: #c0c0c0;
    --thy-darker-gray: #6b7280;
    --thy-secondary-blue: #2563eb;
    --thy-text-color: #374151;
    --thy-medium-gray: #e0e6ed;
    --thy-selection-blue: #c4d9f6;
    --thy-border-blue: #a0c2e3;
    --thy-light-gray: #f0f2f5;
    --thy-white: #ffffff;
  }

  .menu-item-wrapper {
    margin-bottom: 1px;
    user-select: none;
    position: relative;
  }

  .menu-item-wrapper .menu-item {
    padding-left: 25px;
    position: relative;
    z-index: 1;
    display: flex;
    align-items: center;
    gap: 6px;
    font-size: 12px;
    padding: 4px 8px;
    border-radius: 2px;
    cursor: pointer;
    transition: background-color 0.15s ease;
  }

  .menu-item-wrapper .menu-item:hover {
    background-color: var(--thy-light-gray);
  }

  .menu-item.selected {
    background-color: var(--thy-selection-blue);
    border: 1px solid var(--thy-border-blue);
    padding: 3px 7px;
  }

  /* Tree lines using pseudo-elements */
  .menu-item-wrapper::before,
  .menu-item-wrapper::after {
    content: "";
    position: absolute;
    left: 12px;
    border-left: 1px solid var(--thy-tree-line-color);
    z-index: 0;
  }

  .menu-item-wrapper::before {
    top: 0;
    height: 100%;
  }

  .menu-item-wrapper::after {
    top: 13px;
    width: 10px;
    border-bottom: 1px solid var(--thy-tree-line-color);
  }

  /* Nested ul for the tree lines */
  .nested-menu-list {
    list-style: none;
    padding-left: 0;
    margin: 0;
  }

  /* Tree icons */
  .tree-toggle-icon {
    position: absolute;
    left: 5px;
    font-size: 0.9em;
    color: var(--thy-darker-gray);
    cursor: pointer;
    z-index: 2;
  }

  .tree-toggle-icon-placeholder {
    width: 15px;
    flex-shrink: 0;
  }

  .tree-folder-icon,
  .tree-item-icon {
    position: absolute;
    left: 18px;
    font-size: 1.1em;
    width: 1.2em;
    text-align: center;
    z-index: 1;
  }

  .tree-folder-icon {
    color: var(--thy-secondary-blue);
  }

  .tree-item-icon {
    color: var(--thy-text-color);
  }

  .item-label {
    margin-left: 15px;
    white-space: nowrap;
    overflow: hidden;
    text-overflow: ellipsis;
    flex-grow: 1;
  }

  .menu-group-label {
    font-weight: 600;
  }

  .menu-separator-visual {
    width: 100%;
    height: 1px;
    background-color: var(--thy-medium-gray);
    margin: 4px 0;
    opacity: 0.7;
    border: none;
  }
</style>
