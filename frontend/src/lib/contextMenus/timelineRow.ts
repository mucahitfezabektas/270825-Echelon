// src/lib/contextMenus/timelineRow.ts

import type { ContextMenuItem } from "@/stores/contextMenuStore";
import type { RowType } from "@/lib/types"; // RowType'ı da dahil edin

/**
 * Belirli bir zaman çizelgesi satırı için bağlam menüsü öğelerini döndürür.
 * @param data rowId (personel ID'si) ve rowType (actual/publish) içerir.
 */
export function getTimelineRowContextMenu(data?: { rowId: string; rowType: RowType }): ContextMenuItem[] {
    const personId = data?.rowId ?? "Bilinmeyen";
    const rowTypeLabel = data?.rowType === 'actual' ? 'Actual' : data?.rowType === 'publish' ? 'Planlı' : 'Bilinmeyen';

    return [
        {
            label: `👤 ${personId} (Satır Tipi: ${rowTypeLabel})`,
            disabled: true, // Bilgi etiketi, tıklanamaz
            action: () => { }
        },
        { separator: true },
        { label: `🔍 ${personId} Satırını Filtrele`, action: () => console.log("Satır filtrelendi:", data) },
        { label: "✈️ Tüm Uçuşları Göster", action: () => console.log("Tüm uçuşlar gösterildi:", data) },
        { separator: true },
        { label: "⚙️ Satır Ayarları", action: () => alert(`Satır Ayarları için ${personId}`) },
    ];
}