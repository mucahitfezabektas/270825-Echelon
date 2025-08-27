// frontend/src/rules/flightRules.ts
import type { FlightItem } from "@/lib/types";
import type { FlightRule, FlightStyle } from "@/rules/ruleTypes";

/* ------------------------------------------------------------------ */
/* ----------  Yardımcılar ----------------------------------------- */
/* ------------------------------------------------------------------ */

function createFlightKey(f: FlightItem): string | undefined {
    if (!f.departure_port || !f.flight_no || !f.departure_time) return;
    const flightNo = f.flight_no.trim().replace(/^0+/, "");
    return `${f.departure_port}_${flightNo}_${f.departure_time}`;
}

function styleFor(status: "under" | "over" | "exact" | undefined): Partial<FlightStyle> {
    if (!status) return {};
    if (status === "exact") return { fillStyle: "#98c898" };
    if (status === "over") return { fillStyle: "#619861" };
    return { fillStyle: "#ccfecc" }; // under
}

/* ------------------------------------------------------------------ */
/* ----------  GLOBAL BULK CREW STATUS ----------------------------- */
/* ------------------------------------------------------------------ */

let bulkCrewStatusMap: Record<string, "under" | "over" | "exact"> = {};

export function setBulkCrewStatus(map: Record<string, "under" | "over" | "exact">) {
    bulkCrewStatusMap = map;
}

/* ------------------------------------------------------------------ */
/* ----------  FLT Kuralı (sadece bulk) ---------------------------- */
/* ------------------------------------------------------------------ */

export function createFlightRules(): FlightRule[] {
    const rule: FlightRule = {
        id: "flt-crew-match-bulk",
        name: "FLT crew ihtiyacı (bulk preload)",

        condition(flight) {
            return flight.activity_code === "FLT";
        },

        apply(flight) {
            const key = createFlightKey(flight);
            const status = key ? bulkCrewStatusMap[key] : undefined;
            return styleFor(status);
        },
    };

    return [rule];
}
