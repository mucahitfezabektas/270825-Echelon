// snappingStore.ts dosyanızın tamamı (güncellenmiş applySnap ile)

import { writable, get } from 'svelte/store';

// Yakalama Bölgesi Tanımı
interface SnapZone {
    x: number;      // Piksel cinsinden X koordinatı (main-area'nın sol üst köşesine göre)
    y: number;      // Piksel cinsinden Y koordinatı (main-area'nın sol üst köşesine göre)
    width: number;  // Piksel cinsinden genişlik
    height: number; // Piksel cinsinden yükseklik
    id: string;     // Bölgeyi benzersiz şekilde tanımlamak için ID
}

// Düzen Tanımı
interface LayoutZoneDefinition {
    x: number;      // Yüzdesel X koordinatı (0.0 - 1.0)
    y: number;      // Yüzdesel Y koordinatı (0.0 - 1.0)
    width: number;  // Yüzdesel genişlik (0.0 - 1.0)
    height: number; // Yüzdesel yükseklik (0.0 - 1.0)
    id: string;
}

interface Layout {
    name: string;
    zones: LayoutZoneDefinition[];
}

// Kullanılabilir Düzenler
const availableLayouts: Record<string, Layout> = {
    "grid-2x2": {
        name: "2x2 Grid",
        zones: [
            { x: 0, y: 0, width: 0.5, height: 0.5, id: "zone-0-0" },
            { x: 0.5, y: 0, width: 0.5, height: 0.5, id: "zone-0-1" },
            { x: 0, y: 0.5, width: 0.5, height: 0.5, id: "zone-1-0" },
            { x: 0.5, y: 0.5, width: 0.5, height: 0.5, id: "zone-1-1" },
        ],
    },
    "columns-3": {
        name: "3 Column",
        zones: [
            { x: 0, y: 0, width: 1 / 3, height: 1, id: "zone-col-0" },
            { x: 1 / 3, y: 0, width: 1 / 3, height: 1, id: "zone-col-1" },
            { x: 2 / 3, y: 0, width: 1 / 3, height: 1, id: "zone-col-2" },
        ],
    },
    "columns-2": {
        name: "2 Column",
        zones: [
            { x: 0, y: 0, width: 1 / 2, height: 1, id: "zone-col-right" },
            { x: 1 / 2, y: 0, width: 1 / 2, height: 1, id: "zone-col-left" },
        ],
    },
    "row-2": {
        name: "2 Row",
        zones: [
            { x: 0, y: 0, width: 1, height: 0.5, id: "zone-main" },
            { x: 0, y: 0.5, width: 1, height: 0.5, id: "zone-footer" },
        ],
    },
    "left-sidebar-main": {
        name: "Main + Sidebar(L)",
        zones: [
            { x: 0, y: 0, width: 0.25, height: 1, id: "zone-sidebar" },
            { x: 0.25, y: 0, width: 0.75, height: 1, id: "zone-main" },
        ],
    },
    "right-sidebar-main": {
        name: "Main + Sidebar(R)",
        zones: [
            { x: 0, y: 0, width: 0.75, height: 1, id: "zone-main" },
            { x: 0.75, y: 0, width: 0.25, height: 1, id: "zone-sidebar" },
        ],
    },

};


// Store'lar
export const isSnappingEnabled = writable(false);
export const selectedLayoutId = writable<keyof typeof availableLayouts>("grid-2x2");
export const snapZones = writable<SnapZone[]>([]);
export const hoveredZone = writable<string | null>(null);

// Main area'nın ekran koordinatlarındaki sol üst köşesini saklamak için
let mainAreaOffsetX = 0;
let mainAreaOffsetY = 0;

export function setMainAreaOffsets(x: number, y: number) {
    mainAreaOffsetX = x;
    mainAreaOffsetY = y;
}

// Yakalama alanlarını güncelleyen fonksiyon
export function updateSnapZones(
    mainRect: DOMRect | null
) {
    const $isSnappingEnabled = get(isSnappingEnabled);
    const $selectedLayoutId = get(selectedLayoutId);

    if (!mainRect || !$isSnappingEnabled) {
        snapZones.set([]);
        return;
    }

    const layout = availableLayouts[$selectedLayoutId];

    const newSnapZones: SnapZone[] = layout.zones.map(zoneDef => ({
        x: zoneDef.x * mainRect.width,
        y: zoneDef.y * mainRect.height,
        width: zoneDef.width * mainRect.width,
        height: zoneDef.height * mainRect.height,
        id: zoneDef.id
    }));

    snapZones.set(newSnapZones);
    console.log("Snap zones updated:", newSnapZones);
}

// Pencere sürüklendiğinde hover edilen alanı tespit et
export function detectHoveredZone(mouseX: number, mouseY: number) {
    const $snapZones = get(snapZones);
    let currentHoveredZone: string | null = null;

    // Fare koordinatlarını main-area'ya göre ayarlayın
    const relativeMouseX = mouseX - mainAreaOffsetX;
    const relativeMouseY = mouseY - mainAreaOffsetY;

    for (const zone of $snapZones) {
        if (
            relativeMouseX >= zone.x &&
            relativeMouseX <= zone.x + zone.width &&
            relativeMouseY >= zone.y &&
            relativeMouseY <= zone.y + zone.height
        ) {
            currentHoveredZone = zone.id;
            break;
        }
    }
    hoveredZone.set(currentHoveredZone);
}

export function applySnap(
    windowId: number,
    mouseX: number,
    mouseY: number,
    openWindows: any[],
    setOpenWindows: (windows: any[]) => void,
    setStatus: (message: string) => void
): boolean {
    const $hoveredZone = get(hoveredZone);
    const $snapZones = get(snapZones);
    const $selectedLayoutId = get(selectedLayoutId);

    if ($hoveredZone) {
        const snappedZone = $snapZones.find(zone => zone.id === $hoveredZone);
        if (snappedZone) {
            const updatedWindows = openWindows.map(win => {
                if (win.id === windowId) {
                    // Snaplenmiş pozisyonu hem mevcut konum hem de restore konum olarak kaydet
                    return {
                        ...win,
                        x: snappedZone.x,
                        y: snappedZone.y,
                        width: snappedZone.width,
                        height: snappedZone.height,
                        // Restore değerlerini de yeni snaplenmiş değerlere eşitle
                        restoreX: snappedZone.x,
                        restoreY: snappedZone.y,
                        restoreWidth: snappedZone.width,
                        restoreHeight: snappedZone.height,
                        active: true,
                        minimized: false // Snaplenen pencere minimize olmamalı
                    };
                }
                return win;
            });

            setOpenWindows(updatedWindows); // `App.svelte`'deki `openWindows`'u güncelle

            setStatus(
                `Pencere "${openWindows.find(w => w.id === windowId)?.title}" "${availableLayouts[$selectedLayoutId].name}" düzenine yakalandı.`
            );

            // 🔧 DOM transform temizliği
            // Bu kısım genellikle pencere bileşeninin kendi içindeki state'i
            // ve DOM manipülasyonunu yönetmesiyle daha iyi olur,
            // ancak dışarıdan zorunlu transform temizliği yapıyorsanız kalsın.
            setTimeout(() => {
                const el = document.querySelector(
                    `.window[data-window-id="${windowId}"]`
                ) as HTMLElement | null;
                if (el) {
                    el.style.transform = "";
                    el.removeAttribute("data-x");
                    el.removeAttribute("data-y");
                }
            }, 0);

            return true;
        }
    }

    return false;
}

export { availableLayouts };
export type { Layout };