// frontend/src/stores/logStore.ts
import { writable } from 'svelte/store';

// Define the type for a log entry
export interface LogEntry {
    timestamp: string; // ISO string for sorting and display
    level: 'info' | 'warn' | 'error' | 'debug';
    message: string;
}

// Create a writable store for log entries
// It will hold an array of LogEntry objects
export const logMessages = writable<LogEntry[]>([]);

/**
 * Adds a new log message to the store.
 * @param level The log level (e.g., 'info', 'warn', 'error').
 * @param message The log message string.
 */
export function addLog(level: LogEntry['level'], message: string) {
    const timestamp = new Date().toISOString(); // Get current timestamp
    const newEntry: LogEntry = { timestamp, level, message };

    logMessages.update(currentLogs => {
        // You might want to limit the number of log entries to prevent performance issues
        const MAX_LOG_ENTRIES = 500; // Keep the last 500 entries
        const updatedLogs = [...currentLogs, newEntry];
        if (updatedLogs.length > MAX_LOG_ENTRIES) {
            return updatedLogs.slice(updatedLogs.length - MAX_LOG_ENTRIES);
        }
        return updatedLogs;
    });

    // Optionally, also log to the browser console for easy debugging during development
    if (import.meta.env.DEV) { // Check if in development mode
        const consoleMethod = console[level] || console.log;
        consoleMethod(`[APP LOG] [${level.toUpperCase()}] ${timestamp}: ${message}`);
    }
}

// Helper functions for specific log levels
export const logInfo = (message: string) => addLog('info', message);
export const logWarn = (message: string, rawData?: any) => addLog('warn', message);
export const logError = (message: string) => addLog('error', message);
export const logDebug = (message: string) => addLog('debug', message);

/**
 * Clears all log messages from the store.
 */
export function clearLogs() {
    logMessages.set([]);
}