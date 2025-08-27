<script lang="ts">
  import { logMessages, type LogEntry } from "@/stores/logStore";
  import { onDestroy, onMount } from "svelte";

  // Reference to the scrollable div to keep it scrolled to the bottom
  let logDisplayElement: HTMLElement;
  let autoScroll = true; // State to control auto-scrolling

  // Function to scroll to the bottom
  function scrollToBottom() {
    if (logDisplayElement && autoScroll) {
      logDisplayElement.scrollTop = logDisplayElement.scrollHeight;
    }
  }

  // Reactively scroll to bottom whenever logMessages change
  $: $logMessages, scrollToBottom();

  // Handle manual scrolling to disable/enable auto-scroll
  function handleScroll() {
    if (logDisplayElement) {
      const { scrollTop, scrollHeight, clientHeight } = logDisplayElement;
      // If scrolled near the bottom, re-enable auto-scroll
      if (scrollHeight - scrollTop <= clientHeight + 10) {
        // +10 for a small buffer
        autoScroll = true;
      } else {
        autoScroll = false;
      }
    }
  }

  onMount(() => {
    // Ensure it scrolls to bottom initially if content is already there
    scrollToBottom();
  });

  onDestroy(() => {
    // No specific cleanup needed for store subscription as it's reactive with $:
  });

  // Function to format timestamp for display
  function formatTimestamp(isoString: string): string {
    const date = new Date(isoString);
    return date.toLocaleTimeString("en-US", {
      hour12: false,
      hour: "2-digit",
      minute: "2-digit",
      second: "2-digit",
      fractionalSecondDigits: 3,
    });
  }

  // Get color for log level
  function getLevelClass(level: LogEntry["level"]): string {
    switch (level) {
      case "info":
        return "log-info";
      case "warn":
        return "log-warn";
      case "error":
        return "log-error";
      case "debug":
        return "log-debug";
      default:
        return "";
    }
  }
</script>

<div class="log-window">
  <div
    class="log-display"
    bind:this={logDisplayElement}
    on:scroll={handleScroll}
  >
    {#each $logMessages as logEntry (logEntry.timestamp + logEntry.message)}
      <div class="log-entry {getLevelClass(logEntry.level)}">
        <span class="log-timestamp"
          >[{formatTimestamp(logEntry.timestamp)}]</span
        >
        <span class="log-level">[{logEntry.level.toUpperCase()}]</span>
        <span class="log-message">{logEntry.message}</span>
      </div>
    {/each}
  </div>
</div>

<style>
  .log-window {
    display: flex;
    flex-direction: column;
    background-color: #2b2b2b; /* Dark background for console look */
    color: #f8f8f2; /* Light text color */
    font-family: "Cascadia Code", "Fira Code", "monospace"; /* Monospace font */
    font-size: 0.8em;
    border: 1px solid #3c3c3c;
    border-radius: 4px;
    overflow: hidden;
    height: 100%; /* Take full height of its container */
    width: 100%; /* Take full width of its container */
  }

  .log-header {
    padding: 8px 12px;
    border-bottom: 1px solid #4a4a4a;
    color: #ffffff;
    font-size: 0.9em;
    display: flex;
    justify-content: space-between;
    align-items: center;
    flex-shrink: 0;
  }

  .log-display {
    flex-grow: 1;
    overflow-y: auto;
    padding: 8px 12px;
    white-space: pre-wrap; /* Preserve whitespace and wrap text */
    word-break: break-all; /* Break long words */
    line-height: 1.4;
    scrollbar-width: thin; /* Firefox */
    scrollbar-color: #6a6a6a #2b2b2b; /* Firefox */
  }

  /* Custom scrollbar for Webkit browsers */
  .log-display::-webkit-scrollbar {
    width: 8px;
    height: 8px;
  }

  .log-display::-webkit-scrollbar-thumb {
    background-color: #6a6a6a;
    border-radius: 4px;
  }

  .log-display::-webkit-scrollbar-track {
    background-color: #2b2b2b;
  }

  .log-entry {
    margin-bottom: 2px;
  }

  .log-timestamp {
    color: #888;
    margin-right: 5px;
  }

  .log-level {
    font-weight: bold;
    margin-right: 5px;
  }

  .log-message {
    color: inherit; /* Message color will be inherited from the log-window or specific level class */
  }

  /* Log Level Specific Colors */
  .log-info .log-level {
    color: #5cb85c;
  } /* Green */
  .log-info .log-message {
    color: #f8f8f2;
  }

  .log-warn .log-level {
    color: #f0ad4e;
  } /* Orange */
  .log-warn .log-message {
    color: #f0ad4e;
  }

  .log-error .log-level {
    color: #d9534f;
  } /* Red */
  .log-error .log-message {
    color: #d9534f;
  }

  .log-debug .log-level {
    color: #5bc0de;
  } /* Light Blue */
  .log-debug .log-message {
    color: #99ccff;
  }
</style>
