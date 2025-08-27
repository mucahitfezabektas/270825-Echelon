import type { RowRule, RuleContext, RowMetrics } from "@/rules/ruleTypes";
import type { OffDayTable } from "@/lib/tableDataTypes";
import tableDataLoader from "@/lib/TableDataLoader";
import dayjs from "dayjs";

const NON_WORKING_CODES = ["IHI", "IMZ", "III", "UHK", "IHK", "UDM", "IUS", "IPR"];
const OFF_DAY_CODES = [
  ...NON_WORKING_CODES,
  "IAC", "IAV", "IBB", "IBC", "IBG", "IBE", "IBI", "IBM", "IBU", "IBV", "IBY", "IOZ"
];

export const rowMetrics: Record<string, RowMetrics> = {};

export const rowRules: RowRule[] = [
  {
    id: "off-day-entitlement",
    name: "Hakediş[kullanılan boş gün]",
    condition: (_groupKey, context) =>
      !!(context.groupedItems && context.rangeStart && context.rangeEnd),
    apply: (groupKey, ctx, baseY, rowHeight, canvasWidth) => {
      const context = (ctx as any).__ruleContext as RuleContext;
      const { groupedItems, rangeStart, rangeEnd } = context;
      const items = groupedItems?.[groupKey] || [];

      const startDate = dayjs(rangeStart);
      const endDate = dayjs(rangeEnd);
      const totalDays = endDate.diff(startDate, "day");

      const dayMap = new Map<string, Set<string>>();
      for (let i = 0; i <= totalDays; i++) {
        const d = startDate.add(i, "day").format("YYYY-MM-DD");
        dayMap.set(d, new Set());
      }

      for (const item of items) {
        const actCode = item.activity_code;
        const start = dayjs(item.departure_time).startOf("day");
        const end = dayjs(item.arrival_time).startOf("day");
        const diffDays = Math.max(0, end.diff(start, "day"));

        for (let d = 0; d <= diffDays; d++) {
          const key = start.add(d, "day").format("YYYY-MM-DD");
          if (dayMap.has(key)) {
            dayMap.get(key)!.add(actCode);
          }
        }
      }

      let workDayCount = 0;
      let usedOffDayCount = 0;
      const debugLogs: string[] = [];

      for (const [dateStr, codeSet] of dayMap.entries()) {
        const codes = Array.from(codeSet);

        const isOffDay = codes.some((code) => OFF_DAY_CODES.includes(code)) || codes.length === 0;
        const isWorkDay = codes.length > 0 && codes.every((code) => !NON_WORKING_CODES.includes(code));

        if (isOffDay) usedOffDayCount++;
        if (isWorkDay) workDayCount++;

        const status = [
          isOffDay ? "boş gün ✅" : "boş gün ❌",
          isWorkDay ? "çalışılmış gün ✅" : "çalışılmış gün ❌"
        ];
        debugLogs.push(`${dateStr} - ${JSON.stringify(codes)} → ${status.join(", ")}`);
      }

      const offDayTable = tableDataLoader.offDayTable;
      const matched = offDayTable
        .slice()
        .sort((a: OffDayTable, b: OffDayTable) => b.work_days - a.work_days)
        .find((row: OffDayTable) => workDayCount >= row.work_days);

      const entitlement = matched?.off_day_entitlement ?? 0;
      const rawDistribution = matched?.distribution;

      const metrics: RowMetrics = {
        groupKey,
        workedDays: workDayCount,
        usedOffDays: usedOffDayCount,
        entitlement,
        rawDistribution,
      };

      // if (!rowMetrics[groupKey]) {
      //   console.groupCollapsed(`📆 Günlük Detayları: ${groupKey}`);
      //   debugLogs.forEach((log) => console.log("rowRules.ts:", log));
      //   console.groupEnd();
      //   console.log("📊 ROW METRICS:", metrics);
      // }

      rowMetrics[groupKey] = metrics;
    }
  }
];
