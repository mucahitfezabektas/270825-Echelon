// src/lib/utils/rebalanceHeights.ts
import type { TimelineEntry } from '@/stores/timelineManager'; // timelineManager'dan tipi import et

export function rebalanceHeights(timelines: TimelineEntry[]) {
    const active = timelines.filter((t) => !t.isMinimized);
    const equal = active.length ? 1 / active.length : 0;
    timelines.forEach(t => {
        if (!t.isMinimized) {
            t.heightRatio = equal;
        }
    });
}