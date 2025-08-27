export type TimeRange = { start: number; end: number };

export function computeFlightsRange(
    flights: { departure_time?: number; arrival_time?: number }[],
    paddingRatio = 0.08 // %8 boşluk
): TimeRange | null {
    const times: number[] = [];
    for (const f of flights) {
        if (typeof f.departure_time === "number") times.push(f.departure_time);
        if (typeof f.arrival_time === "number") times.push(f.arrival_time);
    }

    if (times.length === 0) return null;

    const min = Math.min(...times);
    const max = Math.max(...times);

    if (min === max) {
        // Tek nokta verisi için küçük bir pencere aç (örn: tek bir olay)
        const pad = 60 * 60 * 1000; // 1 saatlik pad
        return { start: min - pad, end: max + pad };
    }

    const span = max - min;
    const pad = Math.floor(span * paddingRatio);
    return { start: min - pad, end: max + pad };
}

export function computeDenseFlightsRange(
    flights: { departure_time?: number; arrival_time?: number }[],
    percentileMin = 0.1,
    percentileMax = 0.9
): TimeRange | null {
    const mids: number[] = [];

    for (const f of flights) {
        if (
            typeof f.departure_time === "number" &&
            typeof f.arrival_time === "number" &&
            f.departure_time > 0 &&
            f.arrival_time > 0 &&
            f.arrival_time > f.departure_time
        ) {
            const mid = (f.departure_time + f.arrival_time) / 2;
            mids.push(mid);
        }
    }

    if (mids.length < 5) return computeFlightsRange(flights);

    mids.sort((a, b) => a - b);

    const total = mids.length;
    const lowerIndex = Math.floor(total * percentileMin);
    const upperIndex = Math.ceil(total * percentileMax);

    const trimmed = mids.slice(lowerIndex, upperIndex);
    if (trimmed.length === 0) return null;

    const min = Math.min(...trimmed);
    const max = Math.max(...trimmed);
    const span = max - min;
    const pad = Math.floor(span * 0.08);

    // 📌 Yoğun bölgenin merkezini al
    const center = (min + max) / 2;

    // 📌 Span + pad toplamını koruyarak merkezden hesapla
    const halfSpan = (span + pad * 2) / 2;

    return {
        start: center - halfSpan,
        end: center + halfSpan,
    };
}
