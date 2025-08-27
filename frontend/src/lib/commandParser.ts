// src/lib/commandParser.ts

const SHORT_TO_FULL_FIELD: Record<string, string> = {
    c: "person_id",
    s: "surname",
    a: "activity_code",
    cl: "class",
    dp: "departure_port",
    ap: "arrival_port",
    d: "date",
    t: "trip_id",
    pt: "plane_tail_name",
    pc: "plane_cms_type",
    gc: "group_code",
    fp: "flight_position",
    fn: "flight_no",
    at: "agreement_type",
    fi: "flight_id",
};

export function parseAbbreviatedCommand(input: string): Record<string, string> {
    const filters: Record<string, string> = {};
    const tokens = input.trim().split(/\s+/);

    for (let i = 0; i < tokens.length; i++) {
        const key = tokens[i];
        const fullKey = SHORT_TO_FULL_FIELD[key];

        if (!fullKey) continue;

        // Eğer 'd' (tarih) alanıysa ve sonrasında 2 token varsa → aralık
        if (key === "d" && i + 2 < tokens.length) {
            const val1 = tokens[i + 1];
            const val2 = tokens[i + 2];

            // ISO tarih kontrolü yapılabilir (isteğe bağlı)
            if (/\d{4}-\d{2}-\d{2}/.test(val1) && /\d{4}-\d{2}-\d{2}/.test(val2)) {
                filters[fullKey] = `${val1} ${val2}`;
                i += 2;
                continue;
            }
        }

        // Normal çiftli komutlar (key + value)
        if (i + 1 < tokens.length) {
            filters[fullKey] = tokens[i + 1];
            i++;
        }
    }

    return filters;
}
