// src/lib/utils/canvasRegistry.ts

const canvasMap: Record<string, Set<HTMLCanvasElement>> = {};

export function registerCanvas(windowId: string, canvas: HTMLCanvasElement): boolean {
    if (!canvasMap[windowId]) canvasMap[windowId] = new Set();

    if (canvasMap[windowId].size >= 3) {
        console.warn(`⚠️ ${windowId} içinde en fazla 3 canvas kullanılabilir.`);
        return false;
    }

    canvasMap[windowId].add(canvas);
    return true;
}

export function unregisterCanvas(windowId: string, canvas: HTMLCanvasElement) {
    canvasMap[windowId]?.delete(canvas);
    if (canvasMap[windowId]?.size === 0) delete canvasMap[windowId];
}

export function getCanvasCount(windowId: string): number {
    return canvasMap[windowId]?.size ?? 0;
}
