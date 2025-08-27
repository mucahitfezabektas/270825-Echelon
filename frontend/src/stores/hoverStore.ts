// src/stores/hoverStore.ts

import { writable } from "svelte/store";
import type { HoverInfo } from "@/lib/types"; // <-- Buradan import etmelisiniz, kendi içinde tanımlamayın

export const hoverStore = writable<HoverInfo>(null); // Artık 'null' da kabul ediyor

export function resetHoverState() {
  hoverStore.set(null);
}