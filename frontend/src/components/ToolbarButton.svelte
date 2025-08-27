<script lang="ts">
  /**
   * ToolbarButton Component
   * A reusable button component for the application's toolbar.
   * Encapsulates styling and basic click event handling for toolbar actions.
   */

  export let title: string; // The tooltip text for the button
  export let iconClass: string; // Font Awesome icon class (e.g., "fas fa-plus")
  export let onClick: () => void = () => {}; // Function to call when the button is clicked
</script>

<button
  class="toolbar-button"
  {title}
  aria-label={title}
  on:click={onClick}
>
  <i class={iconClass} aria-hidden="true"></i>
</button>

<style>
  /* Styles for the ToolbarButton component */
  .toolbar-button {
    background: #ffffff; /* White background for the button */
    border: 1px solid #dcdcdc; /* Light, subtle border */
    border-radius: 6px; /* Rounded corners */
    padding: 4px 8px; /* Internal padding */
    font-size: 1.2em; /* Size of the icon */
    line-height: 1; /* Ensures icon is centered vertically */
    cursor: pointer; /* Indicates interactivity */
    display: flex; /* Flex container to center icon */
    align-items: center;
    justify-content: center;
    /* Smooth transitions for interactive states */
    transition: background-color 0.2s ease, border-color 0.2s ease,
      transform 0.1s ease, box-shadow 0.2s ease, color 0.2s ease;
    box-shadow: 0 2px 4px rgba(0, 0, 0, 0.05); /* Subtle shadow for depth */
    color: #555; /* Default icon color */
    outline: none; /* Remove default outline on focus */
  }

  .toolbar-button:hover {
    background-color: #f0f0f0; /* Slight darkening on hover */
    border-color: #c0c0c0; /* Darker border on hover */
    transform: translateY(-2px); /* Subtle lift effect */
    box-shadow: 0 4px 8px rgba(0, 0, 0, 0.1); /* Enhanced shadow on hover */
    color: #007bff; /* Highlight icon color */
  }

  .toolbar-button:active {
    background-color: #e8e8e8; /* Darker on click */
    border-color: #b0b0b0; /* Even darker border */
    transform: translateY(0); /* Return to original position */
    box-shadow: 0 1px 2px rgba(0, 0, 0, 0.1); /* Depressed shadow */
  }

  .toolbar-button i {
    pointer-events: none; /* Ensure clicks pass through to the button */
  }
</style>