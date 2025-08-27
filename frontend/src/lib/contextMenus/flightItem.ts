// C:\Users\mucah_wi2yyc2\Desktop\280725\frontend\src\lib\contextMenus\flightItem.ts

import type { FlightItem } from "@/lib/types";
import { RowType } from "@/lib/types";
import type { ContextMenuItem } from "@/stores/contextMenuStore";
import { showFlightCrewInNewTimeline } from "@/stores/timelineManager";

export function getFlightItemContextMenu(flight: FlightItem): ContextMenuItem[] {
    const isActualFlight =
        flight.activity_code === "FLT" && flight.type === RowType.Actual;

    return [
        {
            label: `🛫 ${flight.flight_no ?? "Bilinmeyen"} - ${flight.departure_port} → ${flight.arrival_port}`,
            disabled: true,
            action: () => { },
        },
        { separator: true },

        ...(isActualFlight
            ? [
                {
                    label: "✈️ Uçuştaki kişileri göster",
                    action: () => {
                        console.log(`Uçuştaki kişileri göster: ${flight.flight_id}`);
                        showFlightCrewInNewTimeline(flight.flight_id);
                    },
                } as ContextMenuItem,
            ]
            : []),

        {
            label: "🗑️ Bu Aktiviteyi Sil",
            action: () => {
                alert(`Silinecek aktivite: ${flight.flight_no}`);
            },
        },
    ];
}
