import { writable } from "svelte/store";

export const ghostDockZone = writable<"left" | "right" | "top" | "bottom" | null>(null);

