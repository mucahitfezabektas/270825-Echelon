import { writable } from "svelte/store";

export interface HighlightRequest {
    timelineId: string;
    keys: string[];
}

export const highlightStore = writable<HighlightRequest | null>(null);
