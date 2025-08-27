// src/lib/dragStore.ts
import { writable } from 'svelte/store';
import type { FlightItem } from '@/lib/types'; // FlightItem tipinizi buraya import edin ve güncel olduğundan emin olun

export interface DragState {
    isDraggingRow: boolean; // Bir satırın mı, yoksa bir öğenin mi sürüklendiğini belirtir
    draggedGroupKey: string | null; // Sürüklenen satırın anahtarı
    draggedRowFlights: FlightItem[] | null; // Sürüklenen satıra ait tüm uçuş öğeleri
    sourceTimelineId: string | null; // Satırın hangi zaman çizelgesinden geldiği
    draggedRowVisualY: number; // Sürüklenen satırın görsel Y pozisyonu (Ekran koordinatında)
    dragRowOffsetY: number; // Farenin satırın başından olan Y ofseti
}

const initialDragState: DragState = {
    isDraggingRow: false,
    draggedGroupKey: null,
    draggedRowFlights: null,
    sourceTimelineId: null,
    draggedRowVisualY: 0,
    dragRowOffsetY: 0,
};

export const dragStore = writable<DragState>(initialDragState);

// Sürükleme durumunu sıfırlayan bir fonksiyon
export function resetDragState() {
    dragStore.set(initialDragState);
}