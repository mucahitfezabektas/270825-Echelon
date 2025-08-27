import { get } from "svelte/store";
import { activeTimelines } from "@/stores/timelineManager";
import { getPublishDataByPersonId } from "@/lib/api";
import { RowType } from "@/lib/types";
import type { FlightItem } from "@/lib/types";
import { showAlert } from "@/stores/alertStore";

/**
 * Belirtilen person_id için publish verisini API'den alıp
 * actual satırın altına ekler.
 */
export async function insertPublishRowBelow(personId: string, timelineId: string) {
    const currentTimelines = get(activeTimelines);
    const timelineIndex = currentTimelines.findIndex((t) => t.id === timelineId);

    if (timelineIndex === -1) {
        console.warn(`[insertPublishRowBelow] Timeline bulunamadı: ${timelineId}`);
        return;
    }

    const timeline = currentTimelines[timelineIndex];

    // Zaten publish satırı var mı?
    const alreadyExists = timeline.flights.some(
        (item) => item.person_id === personId && item.type === RowType.Publish
    );
    if (alreadyExists) {
        console.info(`[insertPublishRowBelow] Zaten eklenmiş: ${personId} (publish)`);
        return;
    }

    // API'den publish verisini al
    let publishItems: FlightItem[] = [];
    try {
        publishItems = await getPublishDataByPersonId(personId);
        publishItems = publishItems.map((item) => ({
            ...item,
            type: RowType.Publish,
        }));
    } catch (err: any) {
        console.error(`[insertPublishRowBelow] Publish verisi alınamadı: ${personId}`, err);

        // Use your custom showAlert instead of the browser's alert
        if (err.message && typeof err.message === 'string' && err.message.includes("No Publish record found for the specified person.")) {
            showAlert(`Publish tablosunda yayın verisi bulunamadı: ${personId} için herhangi bir yayın kaydı bulunamadı.`, 'Publish Tablosunda Kayıt Bulanamadı.');
        } else {
            showAlert(`Yayın verisi alınırken bir hata oluştu: ${err.message || 'Bilinmeyen bir hata oluştu.'}`, 'API Hatası');
        }
        return;
    }

    // Actual satırın konumunu bul
    const actualIndex = timeline.flights.findIndex(
        (item) => item.person_id === personId && item.type === RowType.Actual
    );
    if (actualIndex === -1) {
        console.warn(`[insertPublishRowBelow] Actual satır bulunamadı: ${personId}`);
        return;
    }

    // Yeni flight dizisi
    const updatedFlights = [
        ...timeline.flights.slice(0, actualIndex + 1),
        ...publishItems,
        ...timeline.flights.slice(actualIndex + 1),
    ];

    // visibleRowTypes güncellemesi
    const currentVisible = timeline.visibleRowTypes?.[personId] ?? [];
    const newVisible = Array.from(new Set([...currentVisible, RowType.Actual, RowType.Publish]));

    activeTimelines.update((timelines) => {
        const updated = [...timelines];
        updated[timelineIndex] = {
            ...timeline,
            flights: updatedFlights,
            visibleRowTypes: {
                ...timeline.visibleRowTypes,
                [personId]: newVisible,
            },
        };
        return updated;
    });

    console.log(`✅ [insertPublishRowBelow] ${personId} için publish satırı eklendi.`);
}