// src/stores/zoomControlStore.ts
import { writable, get, type Writable } from 'svelte/store';

export type ZoomController = {
    zoomIn: () => void;
    zoomOut: () => void;
    setZoom: (newZoom: number) => void;   // newZoom: px/ms ya da senin tanımın
    resetZoom: () => void;
    fitToRange: (start: number, end: number) => void;
    subscribe?: Writable<number>['subscribe'];
};

const _controllers = new Map<string, { store: Writable<number>, api: ZoomController }>();

export function registerZoomController(
    id: string,
    api: ZoomController,
    initialZoom: number = 1
): ZoomController {
    if (_controllers.has(id)) {
        console.warn(`[ZoomStore] Controller for ID "${id}" already registered. Returning existing instance.`);
        return _controllers.get(id)!.api;
    }

    const zoomStore = writable(initialZoom);

    // <<< kritik sarmalayıcı >>>
    const originalSetZoom = api.setZoom;
    api.setZoom = (newZoom: number) => {
        // önce bileşen içi mantığı çalıştır (layout/axis vb.)
        originalSetZoom(newZoom);
        // sonra store’u güncelle ki abone olan tüm yerler tetiklensin
        zoomStore.set(newZoom);
    };

    // harici abonelik
    api.subscribe = zoomStore.subscribe;

    _controllers.set(id, { store: zoomStore, api });
    return api;
}

export function getZoomController(id: string): ZoomController | undefined {
    return _controllers.get(id)?.api;
}

export function unregisterZoomController(id: string) {
    _controllers.delete(id);
}

// (opsiyonel) Lazım olursa dışarıdan store değerini okumak/kurmak için yardımcılar:
export function getZoomValue(id: string): number | undefined {
    const entry = _controllers.get(id);
    return entry ? get(entry.store) : undefined;
}
export function setZoomValue(id: string, newZoom: number) {
    const entry = _controllers.get(id);
    if (entry) entry.store.set(newZoom);
}
