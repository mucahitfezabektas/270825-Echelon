import type { ContextMenuItem } from "@/stores/contextMenuStore";
import type { RowType } from "@/lib/types";
import { insertPublishRowBelow } from "@/stores/timelineActions/insertPublishRowBelow";

/**
 * Zaman çizelgesi satırının kenar çubuğu (LeftSidebar) için bağlam menüsü öğelerini döndürür.
 * Bu menü, satırın kendisiyle ilgili genel ayar ve yönetim eylemlerini içerir.
 * @param data rowId (personel ID'si), rowType (actual/publish), timelineId içerir.
 */
export function getTimelineRowSidebarContextMenu(data?: {
    rowId: string;
    rowType: RowType;
    timelineId: string;
}): ContextMenuItem[] {
    const personId = data?.rowId ?? "Bilinmeyen";
    const timelineId = data?.timelineId ?? "unknown";

    const rowTypeLabel =
        data?.rowType === "actual"
            ? "Actual"
            : data?.rowType === "publish"
                ? "Planlı"
                : "Inform";

    const menu: ContextMenuItem[] = [
        {
            label: `👤 Personel: ${personId} (${rowTypeLabel})`,
            disabled: true,
            action: () => { },
        },
        { separator: true },

        {
            label: `🔎 ${personId} Detaylarını Aç`,
            action: () => console.log(`Personel detayları açıldı: ${personId}`),
        },
        {
            label: `📋 ${personId} Verilerini Kopyala`,
            action: () => alert(`Personel ID ${personId} kopyalandı.`),
        },
    ];

    if (data?.rowType === "actual") {
        menu.push({
            label: `📂 Publish Verisi Aç`,
            action: () => {
                console.log(`📂 Publish verisi yükleniyor: ${personId}`);
                insertPublishRowBelow(personId, timelineId);
            },
        });
    }

    menu.push(
        { separator: true },
        {
            label: `📊 Performans Raporu Oluştur`,
            action: () => console.log(`Performans raporu oluşturuluyor: ${personId}`),
        },
        {
            label: `📧 Mail Gönder: ${personId}`,
            action: () => alert(`Mail gönderiliyor: ${personId}`),
        },
        { separator: true },
        {
            label: "⚙️ Kenar Çubuğu Ayarları",
            action: () => alert("Kenar Çubuğu Ayarları"),
        },
        {
            label: "⬇️ Satırı Gizle",
            action: () => console.log(`Satır gizlendi: ${personId}`),
        }
    );

    return menu;
}
