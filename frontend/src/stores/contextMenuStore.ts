// src/stores/contextMenuStore.ts

import { writable } from 'svelte/store';
import type { SvelteComponent } from 'svelte';

// Menü seçeneği tipini tanımla
export type ContextMenuItem =
    | {
        label: string;
        action: (data?: any) => void;
        disabled?: boolean;
        separator?: false; // Açıkça false olarak işaretliyoruz veya hiç belirtmiyoruz
    }
    | {
        separator: true;
        label?: string; // Separator için label ve action opsiyonel olabilir
        action?: (data?: any) => void;
        disabled?: boolean;
    };

// Context menünün durumunu tanımla
export interface ContextMenuState {
    show: boolean;
    x: number;
    y: number;
    items: ContextMenuItem[];
    // Menüye özgü veri (örneğin, sağ tıklanan FlightItem, RowType vb.)
    data?: any;
}

// Varsayılan boş state
const initialState: ContextMenuState = {
    show: false,
    x: 0,
    y: 0,
    items: [],
    data: null,
};

export const contextMenuStore = writable<ContextMenuState>(initialState);

// Menüyü göstermek için yardımcı fonksiyon
export function showContextMenu(x: number, y: number, items: ContextMenuItem[], data?: any) {
    console.log("ContextMenu: Menü gösteriliyor.", { x, y, items, data });
    contextMenuStore.set({ show: true, x, y, items, data });
}

// Menüyü gizlemek için yardımcı fonksiyon
export function hideContextMenu() {
    console.log("ContextMenu: Menü gizleniyor.");
    contextMenuStore.set(initialState);
}