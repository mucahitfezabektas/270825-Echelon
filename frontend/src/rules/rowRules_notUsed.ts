// // src/rules/rowRules.ts
// import type { RowRule } from "@/rules/ruleTypes";
// import {
//     TIMELINE_PADDING_LEFT,
// } from "@/lib/constants/timelineConstants";

// export const rowRules: RowRule[] = [
//     {
//         id: "highlight-long-duty",
//         name: "Toplam uçuş süresi(FLT) 60 saati aşan satırı işaretle",
//         condition: (_groupKey, context) => {
//             const duration = context.totalDurations[_groupKey];
//             return typeof duration === "number" && duration > 60 * 60; // dakika cinsinden
//         },
//         apply: (_groupKey, ctx, baseY, rowHeight) => {
//             ctx.fillStyle = "rgba(255, 0, 0, 0.09)";
//             ctx.fillRect(
//                 0, // Start X at 0 for the left sidebar
//                 baseY,
//                 TIMELINE_PADDING_LEFT, // Width is TIMELINE_PADDING_LEFT for the left sidebar
//                 rowHeight
//             );
//         },
//     }
// ];

// src/rules/rowRules.ts
import type { RowRule } from "@/rules/ruleTypes";

// Eğer başka row rule ekleyecekseniz buraya ekleyin
export const rowRules: RowRule[] = [
  // Şu anda aktif bir kural yok
];


//Tüm uygulama aylık planda çalışmaktadır. Satırdaki herhangi bir aktivite olmayan gün sayısını hesaplayan kural yaz