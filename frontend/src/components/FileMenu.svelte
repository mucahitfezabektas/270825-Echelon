<script lang="ts">
  /**
   * FileMenu Component
   * Represents a dropdown menu for the "File" actions in the menu bar.
   * Handles item selection and provides a basic animation.
   */

  import { createEventDispatcher } from 'svelte';
  import { fly } from 'svelte/transition'; // Svelte transition for smooth entry
  import { quintOut } from 'svelte/easing'; // Easing function for the transition

  const dispatch = createEventDispatcher(); // Used to emit custom events

  export let show: boolean = false; // Controls the visibility of the dropdown

  /**
   * Handles a click on a menu item.
   * Dispatches a 'select' event with the action name and closes the menu.
   * @param action The string identifier for the selected action (e.g., 'new', 'open').
   */
  function handleClick(action: string) {
    dispatch('select', action);
    show = false; // Close menu after selection
  }

  /**
   * Handles clicks outside the dropdown menu to close it.
   * This is crucial for proper dropdown behavior.
   * @param event The DOM click event.
   */
  function handleOuterClick(event: MouseEvent) {
    if (show && event.target && !(event.target as HTMLElement).closest('.menu-item-wrapper')) {
      show = false;
    }
  }

  // Reactive statement: Add/remove click listener on the document body
  // whenever the 'show' property changes. This ensures the menu closes
  // when clicking anywhere else on the page.
  $: if (show) {
    document.addEventListener('click', handleOuterClick);
  } else {
    document.removeEventListener('click', handleOuterClick);
  }
</script>

{#if show}
  <div class="menu-dropdown" transition:fly={{ y: -10, duration: 150, easing: quintOut }}>
    <button class="dropdown-item" on:click={() => handleClick('new')} aria-label="New File">
      <i class="fas fa-file-alt dropdown-icon" aria-hidden="true"></i> New
    </button>
    <button class="dropdown-item" on:click={() => handleClick('open')} aria-label="Open File">
      <i class="fas fa-folder-open dropdown-icon" aria-hidden="true"></i> Open...
    </button>
    <button class="dropdown-item" on:click={() => handleClick('save')} aria-label="Save File">
      <i class="fas fa-save dropdown-icon" aria-hidden="true"></i> Save
    </button>
    <div class="dropdown-divider"></div> <button class="dropdown-item" on:click={() => handleClick('exit')} aria-label="Exit Application">
      <i class="fas fa-power-off dropdown-icon" aria-hidden="true"></i> Exit
    </button>
  </div>
{/if}

<style>
  /* Styles for the dropdown menu container */
  .menu-dropdown {
    position: absolute;
    top: 100%; /* Position right below the parent menu button */
    left: 0;
    background-color: #ffffff; /* White background */
    border: 1px solid #e0e0e0; /* Subtle border */
    border-radius: 6px; /* Rounded corners */
    box-shadow: 0 4px 12px rgba(0, 0, 0, 0.15); /* Soft shadow for depth */
    min-width: 180px; /* Minimum width for the dropdown */
    z-index: 100; /* Ensure it appears above other elements */
    padding: 8px 0; /* Vertical padding inside the dropdown */
  }

  /* Styles for individual dropdown menu items */
  .dropdown-item {
    background: none; /* No background by default */
    border: none; /* No border */
    color: #333; /* Dark text color */
    padding: 10px 15px; /* Internal padding */
    width: 100%; /* Full width within the dropdown */
    text-align: left; /* Align text to the left */
    cursor: pointer; /* Indicate interactivity */
    font-size: 14px; /* Standard font size */
    display: flex; /* Use flexbox for icon and text alignment */
    align-items: center;
    gap: 10px; /* Space between icon and text */
    /* Smooth transitions for hover states */
    transition: background-color 0.15s ease, color 0.15s ease;
  }

  .dropdown-item:hover {
    background-color: #f0f0f0; /* Light highlight on hover */
    color: #007bff; /* Primary blue text color on hover */
  }

  .dropdown-item:focus {
    outline: 2px solid #007bff; /* Clear focus outline for accessibility */
    outline-offset: -2px; /* Keep outline inside the button */
  }

  /* Styles for icons within dropdown items */
  .dropdown-icon {
    font-size: 1em; /* Icon size relative to text */
    width: 20px; /* Fixed width for consistent alignment of icons */
    text-align: center;
    color: #666; /* Default icon color */
  }

  .dropdown-item:hover .dropdown-icon {
    color: #007bff; /* Change icon color on item hover */
  }

  /* Styles for horizontal divider within the dropdown */
  .dropdown-divider {
    height: 1px;
    background-color: #e0e0e0; /* Light gray line */
    margin: 8px 0; /* Vertical margin around the divider */
  }
</style>