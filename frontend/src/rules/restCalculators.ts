import dayjs from "dayjs";
import type { FlightItem } from "@/lib/types";

const POST_FLIGHT_BUFFER_MIN = 0;
const MIN_REST_HOURS = 10;


export function attachRestInfoToSingleFlight(flight: FlightItem, allFlights: FlightItem[]): { rest_start: string, rest_end: string, rest_duration: number } | undefined {
    const relevantFlights = allFlights.filter(f => f.trip_id === flight.trip_id);
    const sorted = relevantFlights.sort((a, b) => a.departure_time - b.departure_time);
    const lastFLT = [...sorted].reverse().find(f => f.activity_code === "FLT");

    if (!lastFLT || lastFLT.data_id !== flight.data_id) return undefined;

    const ugsH = (flight.arrival_time - flight.departure_time) / 3_600_000;
    const dutyEnd = dayjs(flight.arrival_time).add(POST_FLIGHT_BUFFER_MIN, "minute");
    const minRestH = Math.max(MIN_REST_HOURS, ugsH);
    const restEnd = dutyEnd.add(minRestH, "hour");

    return {
        rest_start: dutyEnd.toISOString(),
        rest_end: restEnd.toISOString(),
        rest_duration: +minRestH.toFixed(2),
    };
}



export function attachRestInfo(items: FlightItem[]): FlightItem[] {
    const byTrip = new Map<string, FlightItem[]>();

    // 🔹 Uçuşları trip_id'ye göre grupla
    for (const item of items) {
        if (!item.trip_id) continue; // trip_id olmayanlar dışarıda bırakılabilir
        if (!byTrip.has(item.trip_id)) {
            byTrip.set(item.trip_id, []);
        }
        byTrip.get(item.trip_id)!.push(item);
    }

    const result: FlightItem[] = [];

    for (const tripFlights of byTrip.values()) {
        // 🔹 Saat sırasına göre sırala (emin olmak için)
        const sorted = tripFlights.sort((a, b) => a.departure_time - b.departure_time);

        const lastFLT = [...sorted].reverse().find(f => f.activity_code === "FLT");

        for (const flight of sorted) {
            if (flight === lastFLT) {
                const ugsH = (flight.arrival_time - flight.departure_time) / 3_600_000;
                const dutyEnd = dayjs(flight.arrival_time).add(POST_FLIGHT_BUFFER_MIN, "minute");
                const minRestH = Math.max(MIN_REST_HOURS, ugsH);
                const restEnd = dutyEnd.add(minRestH, "hour");

                result.push({
                    ...flight,
                    rest_start: dutyEnd.toISOString(),
                    rest_end: restEnd.toISOString(),
                    rest_duration: +minRestH.toFixed(2),
                });
            } else {
                result.push({ ...flight });
            }
        }
    }

    return result;
}
