// src/rules/ruleTypes.ts
import type { AircraftCrewNeed } from "@/lib/tableDataTypes";
import type { FlightItem } from "@/lib/types";

export interface RuleContext {
  timelineType: string;
  groupKey: string;
  totalDurations: Record<string, number>;
  groupedItems?: Record<string, FlightItem[]>; // opsiyonel
  rangeStart?: number;
  rangeEnd?: number;
  assignedCrewMap: Record<string, { person_id: string; role: string }[]>;
  aircraftCrewNeed: AircraftCrewNeed[];
  invalidate?: () => void;   // repaint tetikleyici (opsiyonel)
}

export type FlightStyle = {
  fillStyle?: string;
  strokeStyle?: string;
  fontStyle?: string;
};

export interface FlightRule {
  id: string;
  name: string;
  condition: (flight: FlightItem, context: RuleContext) => boolean;
  apply: (flight: FlightItem, context: RuleContext) => Partial<FlightStyle>;
}


export interface RowRule {
  id: string;
  name: string;
  condition: (groupKey: string, context: RuleContext) => boolean;
  apply: (
    groupKey: string,
    ctx: CanvasRenderingContext2D,
    baseY: number,
    rowHeight: number,
    canvasWidth: number
  ) => void;
}

// ruleTypes.ts
export interface RowStats {
  totalEmptyDays: number;
  reducedEmptyDays: number; // IBB + IBC çıkarılmış hali
}


export interface RowMetrics {
  groupKey: string;
  workedDays: number;
  usedOffDays: number;
  entitlement: number;
  rawDistribution?: string;
}