// src/stores/timelineManager.ts

import { writable, get } from 'svelte/store';
import { fetchActualsByFlightID, fetchCrewTimelineData } from '@/lib/api'; // Güncellenmiş API fonksiyonunu import et
import type { FlightItem, SavedFilter } from '@/lib/types'; // SavedFilter'ı import et
import { parseAbbreviatedCommand } from '@/lib/commandParser';
import { rebalanceHeights } from '@/lib/utils/rebalanceHeights'; // Yeni yardımcı fonksiyon
import type { RowType } from "@/lib/types";

// 👇 Yeni Importlar: ZoomController ve zaman aralığı hesaplama
import { getZoomController } from "@/stores/zoomControlStore";
import { computeFlightsRange } from "@/lib/timeRange";


// TimelineEntry arayüzü
// Bu arayüz zaten burada tanımlı olduğu için, Timeline.svelte'den kaldırıldı ve buradan dışa aktarılıyor.
export interface TimelineEntry { // Exported for use in Timeline.svelte
    id: string;
    flights: FlightItem[];
    loading: boolean;
    error: string | null;
    currentSearchQuery: Record<string, string>; // Artık sadece Record<string, string> olarak kalmalı, SavedFilter'ı ayrıştırmalısınız
    timelineType: "roster" | "trip" | "rotation"; // Farklı zaman çizelgesi tipleri
    heightRatio: number;
    isMinimized: boolean;
    loadingProgress: number;

    // 👇 Yeni eklenecek alan
    visibleRowTypes: Record<string, RowType[]>;
}

// Aktif zaman çizelgeleri için Svelte store'u
export const activeTimelines = writable<TimelineEntry[]>([]);
let timelineIdCounter = 0; // Her timeline için benzersiz ID sağlamak amacıyla

export function resetAllTimelines() {
    activeTimelines.set([]); // Store'u boş bir diziye ayarla
    timelineIdCounter = 0;   // ID sayacını da sıfırla
    console.log("✅ Tüm zaman çizelgeleri sıfırlandı.");
}

// Yardımcı fonksiyon: Yeni boş bir zaman çizelgesi oluşturur
function createEmptyTimelineEntry(
    type: "roster" | "trip" | "rotation" = "roster"
): TimelineEntry {
    timelineIdCounter++;
    return {
        id: `timeline-${timelineIdCounter}`,
        flights: [],
        loading: false,
        error: null,
        currentSearchQuery: {}, // Başlangıçta boş Record<string, string>
        timelineType: type,
        heightRatio: 0.5,
        isMinimized: false,
        loadingProgress: 0,
        visibleRowTypes: {}, // 👈 Buraya eklendi
    };
}


export function updateTimelineFlights(id: string, flights: FlightItem[]) {
    const grouped: Record<string, RowType[]> = {};

    for (const f of flights) {
        const key = f.person_id;
        if (!grouped[key]) grouped[key] = [];
        if (!grouped[key].includes(f.type as RowType)) {
            grouped[key].push(f.type as RowType);
        }

    }

    activeTimelines.update((timelines) =>
        timelines.map((t) =>
            t.id === id
                ? {
                    ...t,
                    flights,
                    visibleRowTypes: grouped, // 👈 Eklenen satır
                }
                : t
        )
    );
}


// Zaman çizelgesini güncelleme veya yeni bir sorgu başlatma fonksiyonu
export async function updateTimeline(
    targetTimelineId: string,
    filters: Record<string, string>, // Sadece Record<string, string> almalı
    type: "roster" | "trip" | "rotation" = "roster", // Varsayılan tip
    options?: { fitToData?: boolean } // 👈 Yeni opsiyonel parametre
) {
    const fitToData = options?.fitToData !== false; // Varsayılan olarak true

    // Önce store'u bulalım ve güncelleyelim
    activeTimelines.update(currentTimelines => {
        let timelineToUpdate = currentTimelines.find(t => t.id === targetTimelineId);

        if (!timelineToUpdate) {
            timelineToUpdate = createEmptyTimelineEntry(type);
            currentTimelines = [...currentTimelines, timelineToUpdate];
            rebalanceHeights(currentTimelines);
        }

        // currentSearchQuery'yi gelen filtre ile güncelle
        timelineToUpdate.currentSearchQuery = filters;
        timelineToUpdate.timelineType = type;

        timelineToUpdate.loading = true;
        timelineToUpdate.error = null;
        timelineToUpdate.loadingProgress = 0;

        const indexToUpdate = currentTimelines.findIndex(t => t.id === timelineToUpdate!.id);
        if (indexToUpdate !== -1) {
            currentTimelines[indexToUpdate] = timelineToUpdate;
        } else {
            currentTimelines.push(timelineToUpdate);
        }
        return [...currentTimelines];
    });

    // Store güncellendikten sonra, güncellenen timeline objesini tekrar al
    const timelinesNow = get(activeTimelines);
    const timelineEntry = timelinesNow.find(t => t.id === targetTimelineId);
    if (!timelineEntry) {
        console.error(`⛔️ Güncellenecek zaman çizelgesi bulunamadı: ${targetTimelineId}`);
        console.warn("Aktif timeline ID'ler:", timelinesNow.map(t => t.id));
        return;
    }

    let progressInterval: NodeJS.Timeout | undefined;

    // Eğer filtre objesi boş değilse (yani bir sorgu varsa) ilerleme çubuğunu başlat
    const isFilterEmpty = Object.keys(filters).length === 0;

    if (!isFilterEmpty) {
        progressInterval = setInterval(() => {
            activeTimelines.update(currentTimelines => {
                const updated = [...currentTimelines];
                const idx = updated.findIndex(t => t.id === targetTimelineId);
                if (idx !== -1) {
                    updated[idx] = {
                        ...updated[idx],
                        loadingProgress: Math.min(updated[idx].loadingProgress + 5, 95),
                    };
                }
                return updated;
            });
        }, 100);
    } else {
        // Filtre boşsa, progress bar'ı hemen %100 yap ve verileri temizle
        if (progressInterval) clearInterval(progressInterval);
        activeTimelines.update(currentTimelines => {
            const updated = [...currentTimelines];
            const idx = updated.findIndex(t => t.id === targetTimelineId);
            if (idx !== -1) {
                updated[idx] = {
                    ...updated[idx],
                    loading: false,
                    flights: [],
                    error: null,
                    loadingProgress: 100,
                };
            }
            return updated;
        });
        // 👇 Eğer filtre boşsa ve veri temizlendiyse, görünümü de sıfırla
        if (fitToData) { // Sadece fitToData aktifse resetle
            const ctrl = getZoomController(targetTimelineId);
            if (ctrl?.resetZoom) {
                console.log(`[TimelineManager] ${targetTimelineId} için filtre boş, zoom resetleniyor.`);
                ctrl.resetZoom();
            }
        }
        return;
    }

    try {
        const { flights: fetchedFlights } = await fetchCrewTimelineData(filters);

        if (progressInterval) clearInterval(progressInterval);

        // Flights'ı güncelledikten sonra visibleRowTypes'ı da hesapla
        const grouped: Record<string, RowType[]> = {};
        for (const f of fetchedFlights) {
            const key = f.person_id;
            if (!grouped[key]) grouped[key] = [];
            if (!grouped[key].includes(f.type as RowType)) {
                grouped[key].push(f.type as RowType);
            }
        }

        activeTimelines.update(currentTimelines => {
            const updated = [...currentTimelines];
            const idx = updated.findIndex(t => t.id === targetTimelineId);
            if (idx !== -1) {
                updated[idx] = {
                    ...updated[idx],
                    flights: fetchedFlights,
                    loading: false,
                    error: null,
                    loadingProgress: 100,
                    visibleRowTypes: grouped, // 👈 Güncellendi
                };
            }
            return updated;
        });

        // 👇 Yeni: Veri çekildikten sonra zaman aralığını hesapla ve uygula
        if (fitToData && fetchedFlights.length > 0) {
            const range = computeFlightsRange(fetchedFlights);
            if (range) {
                const ctrl = getZoomController(targetTimelineId);
                if (ctrl?.fitToRange) {
                    console.log(`[TimelineManager] ${targetTimelineId} için fitToRange çağrılıyor: ${new Date(range.start).toLocaleString()} - ${new Date(range.end).toLocaleString()}`);
                    ctrl.fitToRange(range.start, range.end);
                } else {
                    console.warn(`[TimelineManager] ${targetTimelineId} için ZoomController veya fitToRange metodu bulunamadı.`);
                }
            } else {
                console.info(`[TimelineManager] ${targetTimelineId} için uçuşlar için zaman aralığı bulunamadı, fitToData uygulanmıyor.`);
            }
        } else if (fitToData && fetchedFlights.length === 0) {
            // Hiç uçuş gelmediyse ve fitToData aktifse, görünümü resetle
            const ctrl = getZoomController(targetTimelineId);
            if (ctrl?.resetZoom) {
                console.log(`[TimelineManager] ${targetTimelineId} için uçuş bulunamadı, zoom resetleniyor.`);
                ctrl.resetZoom();
            }
        }

    } catch (err: any) {
        console.error(`Veri çekilirken hata oluştu (Zaman Çizelgesi ${targetTimelineId}):`, err);

        if (progressInterval) clearInterval(progressInterval);

        activeTimelines.update(currentTimelines => {
            const updated = [...currentTimelines];
            const idx = updated.findIndex(t => t.id === targetTimelineId);
            if (idx !== -1) {
                updated[idx] = {
                    ...updated[idx],
                    flights: [],
                    loading: false,
                    error: err?.message ?? "Veriler getirilemedi.",
                    loadingProgress: 0,
                };
            }
            return updated;
        });
    }
}


// Zaman çizelgesi kaldırma
export function removeTimeline(idToRemove: string) {
    activeTimelines.update(currentTimelines => {
        const updated = currentTimelines.filter(t => t.id !== idToRemove);
        rebalanceHeights(updated); // Kaldırıldıktan sonra yükseklikleri dengele
        return updated;
    });
}


// Yeni boş zaman çizelgesi ekleme
export function addNewEmptyTimeline(type: "roster" | "trip" | "rotation" = "roster") {
    activeTimelines.update(currentTimelines => {
        const newTimeline = createEmptyTimelineEntry(type);
        const updated = [...currentTimelines, newTimeline];
        rebalanceHeights(updated);
        return updated;
    });
    // Yeni bir timeline eklendiğinde de fitToData uygulanması gerekirse (örn: boş bir "şimdi" aralığına sığdırmak için)
    // buraya da bir `updateTimeline` çağrısı eklenebilir. Ancak genellikle veriler yüklendiğinde fit edilmesi tercih edilir.
}


// Zaman çizelgesini minimize etme/maksime etme
export function toggleMinimizeTimeline(timelineId: string) {
    activeTimelines.update(currentTimelines => {
        const updated = currentTimelines.map(t =>
            t.id === timelineId ? { ...t, isMinimized: !t.isMinimized } : t
        );
        // Minimized olanları en alta taşı
        const reordered = [
            ...updated.filter((t) => !t.isMinimized),
            ...updated.filter((t) => t.isMinimized),
        ];
        rebalanceHeights(reordered);
        return reordered;
    });
}


// Genel arama ve prefix işleme (Timeline.svelte'den buraya taşındı)
export async function handleGlobalSearch(searchTerm: string) { // async keyword added
    const trimmedSearchTerm = searchTerm.trim();
    const currentTimelines = get(activeTimelines);

    // Eğer arama terimi boşsa, tüm timeline'ları temizle ve çık
    if (!trimmedSearchTerm) {
        for (const timeline of currentTimelines) {
            await updateTimeline(timeline.id, {}, timeline.timelineType, { fitToData: true }); // 👇 fitToData: true eklendi
        }
        return;
    }

    let targetTimelineId: string | null = null;
    let newTimelineType: "roster" | "trip" | "rotation" = "roster";
    let commandString = trimmedSearchTerm;
    let timelineIndexToTarget: number | null = null;

    // 1. Prefix Kontrolü
    const match = trimmedSearchTerm.match(/^(\d+)\/(.*)|^(\/[a-z]+)\s*(.*)/);

    if (match) {
        if (match[1]) {
            // Sayısal prefix (örn: 1/, 2/)
            const index = parseInt(match[1], 10) - 1;
            if (index >= 0 && index < currentTimelines.length) {
                timelineIndexToTarget = index;
                commandString = match[2].trim();
                newTimelineType = currentTimelines[index].timelineType; // Hedef timeline'ın tipini koru
            } else {
                alert("Belirtilen numarada zaman çizelgesi bulunamadı.");
                return;
            }
        } else if (match[3]) {
            // Özel prefix (örn: /t, /r, /rot)
            const prefix = match[3];
            commandString = match[4].trim();

            if (prefix === "/t") {
                newTimelineType = "trip";
            } else if (prefix === "/r") {
                newTimelineType = "roster";
            } else if (prefix === "/rot") {
                newTimelineType = "rotation";
            } else {
                alert(`Geçersiz prefix: ${prefix}`);
                return;
            }
            // Özel prefix kullanıldığında her zaman yeni bir timeline eklenecek.
            await addNewEmptyTimeline(newTimelineType);
            targetTimelineId = get(activeTimelines)[get(activeTimelines).length - 1].id;
        }
    } else {
        // Prefix yoksa, varsayılan olarak ilk roster timeline'ı hedeflenir.
        if (currentTimelines.length === 0 || currentTimelines[0].timelineType !== "roster") {
            await addNewEmptyTimeline("roster");
            timelineIndexToTarget = 0;
        } else {
            timelineIndexToTarget = 0;
        }
    }

    // Eğer targetTimelineId henüz belirlenmediyse (yani sayısal prefix ile bir timeline hedeflendiyse)
    if (!targetTimelineId && timelineIndexToTarget !== null) {
        targetTimelineId = get(activeTimelines)[timelineIndexToTarget].id;
    }

    // Filtreleri ayrıştır
    const parsedFilters = parseAbbreviatedCommand(commandString);

    if (Object.keys(parsedFilters).length === 0 && commandString !== "") {
        alert("Geçersiz komut. Örn: c 109403 veya /t t TRIP-001");
        return;
    }

    if (targetTimelineId) {
        // Hedeflenen timeline'ın mevcut filtrelerini al
        const targetTimeline = get(activeTimelines).find(t => t.id === targetTimelineId);
        let combinedFilters: Record<string, string> = {};

        if (targetTimeline) {
            combinedFilters = { ...targetTimeline.currentSearchQuery, ...parsedFilters };
        } else {
            combinedFilters = parsedFilters;
        }

        // Güncellenmiş filtrelerle updateTimeline çağrısı
        await updateTimeline(targetTimelineId, combinedFilters, newTimelineType, { fitToData: true }); // 👇 fitToData: true eklendi
    }
}

/**
 * Belirli bir zaman çizelgesini mevcut filtreleriyle yeniden çizer.
 * Bu fonksiyon, bir zaman çizelgesinin verilerini yenilemek veya
 * görselini güncellemek için kullanılabilir.
 * @param timelineId Yeniden çizilecek zaman çizelgesinin ID'si.
 */
export async function redrawTimeline(timelineId: string) {
    const currentTimelines = get(activeTimelines);
    const timelineToRedraw = currentTimelines.find(t => t.id === timelineId);

    if (timelineToRedraw) {
        console.log(`Zaman çizelgesi ${timelineId} yeniden çiziliyor...`);
        // updateTimeline'ı, zaman çizelgesinin mevcut filtreleri ve tipiyle çağırıyoruz.
        await updateTimeline(timelineToRedraw.id, timelineToRedraw.currentSearchQuery, timelineToRedraw.timelineType, { fitToData: true }); // 👇 fitToData: true eklendi
    } else {
        console.warn(`Yeniden çizilecek zaman çizelgesi bulunamadı: ${timelineId}`);
    }
}


// 🌟 YENİ FONKSİYON: Uçuş ekibini yeni bir zaman çizelgesinde gösterme
export async function showFlightCrewInNewTimeline(flightId: string) {
    const newTimelineEntry = createEmptyTimelineEntry("roster"); // Yeni bir roster timeline oluştur
    const newTimelineId = newTimelineEntry.id;

    // Yeni timeline'ı hemen aktif timeline'lar listesine ekle ve yükleniyor durumunu ayarla
    activeTimelines.update(currentTimelines => {
        const updated = [...currentTimelines, { ...newTimelineEntry, loading: true, loadingProgress: 5 }];
        rebalanceHeights(updated); // Yükseklikleri yeniden dengele
        return updated;
    });

    let progressInterval: NodeJS.Timeout | undefined;
    progressInterval = setInterval(() => {
        activeTimelines.update(currentTimelines => {
            const updated = [...currentTimelines];
            const idx = updated.findIndex(t => t.id === newTimelineId);
            if (idx !== -1) {
                updated[idx] = {
                    ...updated[idx],
                    loadingProgress: Math.min(updated[idx].loadingProgress + 5, 95),
                };
            }
            return updated;
        });
    }, 100);


    try {
        // flight_id'ye göre ilgili tüm Actual verilerini çek
        const { person_ids, result: groupedActuals } = await fetchActualsByFlightID(flightId);

        if (progressInterval) clearInterval(progressInterval);

        const allCrewFlights: FlightItem[] = [];
        for (const personId of person_ids) {
            if (groupedActuals[personId]) {
                allCrewFlights.push(...groupedActuals[personId]);
            }
        }

        // Yeni timeline'ın verilerini güncelle
        activeTimelines.update(currentTimelines => {
            const updated = [...currentTimelines];
            const idx = updated.findIndex(t => t.id === newTimelineId);
            if (idx !== -1) {
                updated[idx] = {
                    ...updated[idx],
                    flights: allCrewFlights,
                    loading: false,
                    error: null,
                    loadingProgress: 100,
                    // Yeni eklenen visibleRowTypes'ı da burada güncellemeliyiz
                    visibleRowTypes: allCrewFlights.reduce((acc, f) => {
                        if (!acc[f.person_id!]) acc[f.person_id!] = [];
                        if (!acc[f.person_id!].includes(f.type as RowType)) {
                            acc[f.person_id!].push(f.type as RowType);
                        }
                        return acc;
                    }, {} as Record<string, RowType[]>),
                };
            }
            return updated;
        });
        console.log(`Zaman çizelgesi ${newTimelineId} için ${person_ids.length} kişinin verileri yüklendi.`);

        // 👇 Yeni: Veri çekildikten sonra zaman aralığını hesapla ve uygula
        if (allCrewFlights.length > 0) {
            const range = computeFlightsRange(allCrewFlights);
            if (range) {
                const ctrl = getZoomController(newTimelineId);
                if (ctrl?.fitToRange) {
                    console.log(`[TimelineManager] ${newTimelineId} için fitToRange çağrılıyor: ${new Date(range.start).toLocaleString()} - ${new Date(range.end).toLocaleString()}`);
                    ctrl.fitToRange(range.start, range.end);
                } else {
                    console.warn(`[TimelineManager] ${newTimelineId} için ZoomController veya fitToRange metodu bulunamadı.`);
                }
            } else {
                console.info(`[TimelineManager] ${newTimelineId} için uçuşlar için zaman aralığı bulunamadı, fitToData uygulanmıyor.`);
            }
        } else {
            // Hiç uçuş gelmediyse, görünümü resetle
            const ctrl = getZoomController(newTimelineId);
            if (ctrl?.resetZoom) {
                console.log(`[TimelineManager] ${newTimelineId} için uçuş bulunamadı, zoom resetleniyor.`);
                ctrl.resetZoom();
            }
        }

    } catch (err: any) {
        if (progressInterval) clearInterval(progressInterval);
        console.error(`Uçuş ekibi verisi çekilirken hata oluştu (${flightId}):`, err);

        activeTimelines.update(currentTimelines => {
            const updated = [...currentTimelines];
            const idx = updated.findIndex(t => t.id === newTimelineId);
            if (idx !== -1) {
                updated[idx] = {
                    ...updated[idx],
                    flights: [],
                    loading: false,
                    error: err?.message ?? "Uçuş ekibi verileri getirilemedi.",
                    loadingProgress: 0,
                };
            }
            return updated;
        });
    }
}