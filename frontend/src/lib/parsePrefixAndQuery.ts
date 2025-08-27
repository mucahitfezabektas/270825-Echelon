export type TimelineTarget = "1" | "2" | "r" | "t" | "rot";

export interface ParsedQuery {
    target: TimelineTarget;
    rawQuery: string;
}

/**
 * Örnek:
 *   - "1/c 109403" → { target: "1", rawQuery: "c 109403" }
 *   - "/t t TRIP-01" → { target: "t", rawQuery: "t TRIP-01" }
 *   - "a FLT" → { target: "r", rawQuery: "a FLT" }
 */
export function parsePrefixAndQuery(input: string): ParsedQuery {
    const trimmed = input.trim();

    const match = trimmed.match(/^((\d+)|\/(r|t|rot))\//i);
    if (match) {
        const full = match[0]; // "1/", "/t" gibi
        const value = match[2] || match[3]; // "1" ya da "t"
        const rest = trimmed.slice(full.length).trim();
        return {
            target: value as TimelineTarget,
            rawQuery: rest,
        };
    }

    // Prefix yoksa varsayılan olarak 'r'
    return {
        target: "r",
        rawQuery: trimmed,
    };
}
