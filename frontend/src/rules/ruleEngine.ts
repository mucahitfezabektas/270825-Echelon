// frontend/src/rules/ruleEngine.ts
import type { RuleContext, FlightRule, FlightStyle } from "@/rules/ruleTypes";
import type { FlightItem } from "@/lib/types";
import { createFlightRules, setBulkCrewStatus } from "./flightRules";
import { rowRules } from "./rowRules";

export class RuleEngine {
    private flightRules: FlightRule[];

    constructor(private context: RuleContext) {
        this.flightRules = createFlightRules();
    }

    getFinalStyle(flight: FlightItem): FlightStyle {
        let final: FlightStyle = {};
        for (const r of this.flightRules) {
            if (r.condition(flight, this.context)) {
                Object.assign(final, r.apply(flight, this.context));
            }
        }
        return final;
    }

    applyRowRules(
        groupKey: string,
        ctx: CanvasRenderingContext2D,
        baseY: number,
        rowHeight: number,
        canvasWidth: number
    ) {
        (ctx as any).__ruleContext = this.context;
        for (const r of rowRules) {
            if (r.condition(groupKey, this.context)) {
                r.apply(groupKey, ctx, baseY, rowHeight, canvasWidth);
            }
        }
    }
}

export async function preloadCrewMatchStatuses() {
    try {
        const res = await fetch("http://localhost:8080/api/crew-status");
        if (!res.ok) throw new Error("Status preload failed");
        const data = await res.json();
        setBulkCrewStatus(data);
        console.log("✅ crew-status bulk yüklendi:", Object.keys(data).length);
    } catch (e) {
        console.warn("⚠️ crew-status preload hatası:", e);
    }
}
