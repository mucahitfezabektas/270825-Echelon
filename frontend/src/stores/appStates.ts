// src/stores/appStates.ts
/**
 * Svelte Store for Application State
 * Manages global application-level state, such as status messages.
 */
import { writable } from 'svelte/store';

// Writable store for the application's status message.
// Initialized with a default "Status: Ready" message.
export const statusMessage = writable<string>("Status: Ready");

// Timer ID to clear the status message after a duration.
let statusTimer: ReturnType<typeof setTimeout> | null = null;

/**
 * Updates the global status message.
 * The message will revert to "" after a specified duration.
 * If a new message is set before the previous timer expires, the old timer is cleared.
 * @param message The new status message to display.
 * @param duration The time in milliseconds after which the message reverts to "Ready".
 * Defaults to 3000ms (3 seconds).
 * Set to 0 or negative to keep the message indefinitely.
 */
export function setStatus(message: string, duration: number = 3000) {
    // Clear any existing timer to prevent premature reversion of the status.
    if (statusTimer) {
        clearTimeout(statusTimer);
        statusTimer = null;
    }

    statusMessage.set(message); // Set the new message.

    if (duration > 0) {
        // Set a new timer to revert to "Status: Ready" after the specified duration.
        statusTimer = setTimeout(() => {
            statusMessage.set("Status: Ready");
            statusTimer = null; // Clear the timer reference once it has fired.
        }, duration);
    }
}