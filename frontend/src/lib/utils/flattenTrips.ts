// src/lib/utils/flattenTrips.ts

import type { Trip, Actual } from "@/lib/types";

export function flattenTripsToActivities(trips: Trip[]): Actual[] {
    return trips.flatMap((trip) => trip.activities);
}
