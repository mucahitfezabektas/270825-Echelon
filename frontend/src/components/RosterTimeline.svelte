<script lang="ts">
  import { onMount, onDestroy, createEventDispatcher } from "svelte";
  import type { FlightItem, RowContextMenuEvent, HoverInfo } from "@/lib/types";
  import { RowType } from "@/lib/types";

  import { dragStore, resetDragState } from "@/stores/dragStore";
  import FlightRowOverlay from "@/components/FlightRowOverlay.svelte";
  import { getActivityColor } from "@/stores/colorStore";
  import { FTL_VIOLATION_COLORS } from "@/lib/config/ruleConfig";

  import { timezoneOffsetMinStore } from "@/stores/userOptionsStore";
  import { highlightStore } from "@/stores/highlightStore";

  import { get } from "svelte/store";
  import { activeTimelines } from "@/stores/timelineManager";

  import {
    ROW_HEIGHT,
    ROW_GAP,
    HEADER_HEIGHT,
    TIMELINE_PADDING_LEFT,
    FLIGHT_ITEM_VERTICAL_PADDING,
    FLIGHT_ITEM_MIN_WIDTH,
    FLIGHT_TEXT_OFFSET_X,
    FLIGHT_TEXT_OFFSET_Y,
    MIN_ZOOM,
    MAX_ZOOM,
    DRAG_OVER_COLOR,
    COLLISION_COLOR,
    DRAGGING_ITEM_ALPHA,
  } from "@/lib/constants/timelineConstants";
  import {
    registerZoomController,
    unregisterZoomController,
    type ZoomController, // ZoomController tipini import edin
  } from "@/stores/zoomControlStore";
  import { computeDenseFlightsRange } from "@/lib/timeRange";

  export let locale: string = "tr-TR";
  export let timelineId: string;

  export let isLoading: boolean = false;
  export let loadingProgress: number = 0;

  type SafeDateFormat = Intl.DateTimeFormatOptions & {
    hour?: "2-digit" | "numeric";
    minute?: "2-digit" | "numeric";
    day?: "2-digit" | "numeric";
    weekday?: "short" | "narrow" | "long";
    hourCycle?: "h11" | "h12" | "h23" | "h24";
  };

  const dispatch = createEventDispatcher<{
    flightSelected: FlightItem;
    flightMoved: {
      flight: FlightItem;
      newDepartureTime: number;
      newArrivalTime: number;
      newPersonId: string;
      newRowType: RowType;
      sourceTimelineId: string;
      originalGroupKey: string;
    };
    flightHover: HoverInfo;
    sidebarHeaderDropdownClick: {
      pageX: number;
      pageY: number;
      timelineId: string;
    };
    rowContextMenu: RowContextMenuEvent;
    rowDragStart: {
      groupKey: string;
      flightsInGroup: FlightItem[];
      timelineId: string;
      clientX: number;
      clientY: number;
      dragOffsetY: number;
    };
    flightItemContextMenu: {
      flight: FlightItem;
      pageX: number;
      pageY: number;
      timelineId: string;
    };
    timelineHeaderContextMenu: {
      pageX: number;
      pageY: number;
      timelineId: string;
    };
    timelineEmptyContextMenu: {
      pageX: number;
      pageY: number;
      timelineId: string;
      coords: { x: number; y: number };
    };
    rowSidebarContextMenu: {
      personId: string;
      rowType: RowType;
      pageX: number;
      pageY: number;
      timelineId: string;
    };
  }>();

  let highlightedKeys: Set<string> = new Set();
  let canvas: HTMLCanvasElement;
  let canvasContainerRef: HTMLDivElement;
  let ctx: CanvasRenderingContext2D;

  let groupedFlights: Record<string, FlightItem[]> = {};
  let pixelsPerHour = 5; // Başlangıç piksel/saat değeri
  let spacerHeight = 0;

  let draggedFlight: FlightItem | null = null;
  let dragStartX: number = 0;
  let dragStartY: number = 0;
  let dragOffsetX: number = 0;
  let dragOffsetY: number = 0;
  let isItemDragging: boolean = false;
  let currentDropTargetRowKey: string | null = null;
  let isDroppingOverCollision: boolean = false;

  // Timeline'ın kendi dahili başlangıç ve bitiş zamanları
  let timelineStart: number;
  let timelineEnd: number;

  let isInitialLoad = true;

  let isDragging = false;
  let lastMouseX = 0;
  let hoveredFlight: FlightItem | null = null;
  let clickedFlight: FlightItem | null = null;
  let mouseCanvasX = 0;
  let mouseCanvasY = 0;
  let currentCursor: "grab" | "grabbing" | "pointer" | "not-allowed" = "grab";
  let needsRedraw = false;
  let nowLineInterval: NodeJS.Timeout;

  let scrollY = 0;
  let containerWidth = 0;
  let containerHeight = 0;
  let totalContentHeight = 0;

  let isModifierKeyDown = false;
  let isMKeyDown = false; // 'M' tuşu için yeni durum değişkeni

  let totalDurations: Record<string, number> = {};
  let lastGroupKeys: string[] = [];
  let rowFTLViolations: Record<string, string[]> = {};

  let isGlobalRowDragging: boolean = false;
  let globalDraggedGroupKey: string | null = null;
  let globalSourceTimelineId: string | null = null;
  let timeline;
  let flights: FlightItem[] = [];
  let visibleRowTypes: Record<string, RowType[]> = {};

  // ZoomController'ı tutmak için yerel değişken
  let currentZoomController: ZoomController;

  const unsubscribeDragStore = dragStore.subscribe((state) => {
    isGlobalRowDragging = state.isDraggingRow;
    globalDraggedGroupKey = state.draggedGroupKey;
    globalSourceTimelineId = state.sourceTimelineId;
  });
  type InteractionMode = "zoom" | "pan" | "none";

  const DROPDOWN_ICON = {
    size: 12,
    pad: 6,
  };

  const unsubscribeHighlight = highlightStore.subscribe((req) => {
    if (!req || req.timelineId !== timelineId) return;
    highlightedKeys = new Set(req.keys);
    scrollToFirstHighlighted();
    requestAnimationFrame(drawTimeline);
  });

  // ========== REAKTİF BLOKLAR ==========

  // Reaktif olarak bu timeline'ı store'dan al
  $: {
    timeline = $activeTimelines.find((t) => t.id === timelineId);
    flights = timeline?.flights ?? [];
    visibleRowTypes = timeline?.visibleRowTypes ?? {};
    // isLoading ve loadingProgress props olarak gelir ama burda reaktif olarak da izleyebiliriz:
    isLoading = timeline?.loading ?? false;
    loadingProgress = timeline?.loadingProgress ?? 0;
  }

  $: visibleRows = Object.keys(groupedFlights)
    .sort((a, b) => {
      const [personIdA, typeA] = a.split("_");
      const [personIdB, typeB] = b.split("_");

      if (personIdA !== personIdB) {
        return personIdA.localeCompare(personIdB);
      }
      if (typeA === RowType.Actual && typeB === RowType.Publish) return -1;
      if (typeA === RowType.Publish && typeB === RowType.Actual) return 1;
      return 0;
    })
    .slice(firstVisibleRowIndex, lastVisibleRowIndex + 1);

  $: {
    const numRows = Object.keys(groupedFlights).length;
    const contentH =
      HEADER_HEIGHT + numRows * (ROW_HEIGHT + ROW_GAP) + ROW_GAP * 2;

    spacerHeight = Math.max(0, contentH - containerHeight);
    totalContentHeight = contentH;

    scrollY = Math.min(scrollY, totalContentHeight - containerHeight);
    scrollY = Math.max(0, scrollY);
  }

  // Uçuşlar değiştiğinde veya ilk yüklendiğinde zaman aralığını ayarla
  $: if (flights && visibleRowTypes) {
    groupedFlights = groupFlightsByPersonIdAndType(flights, visibleRowTypes);
    rowFTLViolations = {};

    const denseRange = computeDenseFlightsRange(flights, 0.1, 0.9);
    if (denseRange) {
      timelineStart = denseRange.start;
      timelineEnd = denseRange.end;
    } else {
      const now = Date.now();
      timelineStart = now - 6 * 60 * 60 * 1000;
      timelineEnd = now + 6 * 60 * 60 * 1000;
    }

    isInitialLoad = false;
    scheduleRedraw();
  }

  // Bu reaktif blok, timelineStart, timelineEnd veya container boyutları değiştiğinde çizimi tetikler.
  // Pan ve zoom işlemleri sırasında bu değişkenler değiştiği için çizimin güncellenmesini sağlar.
  $: if (
    ctx &&
    containerWidth > 0 &&
    containerHeight > 0 &&
    timelineStart != undefined &&
    timelineEnd != undefined
  ) {
    scheduleRedraw();
  }

  // ========== YARDIMCI FONKSİYONLAR ==========

  // Zoom fonksiyonları (ZoomController arayüzünü uygular)
  function zoomIn() {
    const centerTime = timelineStart + (timelineEnd - timelineStart) / 2;
    const newPixelsPerHour = Math.min(pixelsPerHour * 1.2, MAX_ZOOM);

    const oldSpanMs = timelineEnd - timelineStart;
    const newSpanMs = oldSpanMs * (pixelsPerHour / newPixelsPerHour);

    timelineStart = centerTime - newSpanMs / 2;
    timelineEnd = centerTime + newSpanMs / 2;
    pixelsPerHour = newPixelsPerHour;
    scheduleRedraw();
  }

  function zoomOut() {
    const centerTime = timelineStart + (timelineEnd - timelineStart) / 2;
    const newPixelsPerHour = Math.max(pixelsPerHour / 1.2, MIN_ZOOM);

    const oldSpanMs = timelineEnd - timelineStart;
    const newSpanMs = oldSpanMs * (pixelsPerHour / newPixelsPerHour);

    timelineStart = centerTime - newSpanMs / 2;
    timelineEnd = centerTime + newSpanMs / 2;
    pixelsPerHour = newPixelsPerHour;
    scheduleRedraw();
  }

  function setZoom(newZoom: number) {
    const oldPixelsPerHour = pixelsPerHour;
    pixelsPerHour = Math.max(MIN_ZOOM, Math.min(newZoom, MAX_ZOOM));

    const centerTime = timelineStart + (timelineEnd - timelineStart) / 2;
    const currentSpanMs =
      ((containerWidth - TIMELINE_PADDING_LEFT) / oldPixelsPerHour) *
      60 *
      60 *
      1000;
    const newSpanMs =
      ((containerWidth - TIMELINE_PADDING_LEFT) / pixelsPerHour) *
      60 *
      60 *
      1000;

    timelineStart = centerTime - newSpanMs / 2;
    timelineEnd = centerTime + newSpanMs / 2;

    scheduleRedraw();
  }

  function resetZoom() {
    pixelsPerHour = 5; // Varsayılan pixelsPerHour değeri

    const range = computeDenseFlightsRange(flights, 0.1, 0.9); // %10–%90 yoğunluk aralığı
    if (range) {
      fitToRange(range.start, range.end);
    } else {
      const now = Date.now();
      timelineStart = now - 6 * 60 * 60 * 1000;
      timelineEnd = now + 6 * 60 * 60 * 1000;
    }

    scheduleRedraw();
  }

  // Bu, zoomControlStore'dan çağrılacak asıl fit-to-range metodudur.
  // Bu metot timeline'ın kendi başlangıç ve bitiş zamanını ayarlar.
  function fitToRange(start: number, end: number) {
    console.log(
      `[RosterTimeline:${timelineId}] fitToRange çağrıldı: ${new Date(
        start
      ).toLocaleString()} - ${new Date(end).toLocaleString()}`
    );
    timelineStart = start;
    timelineEnd = end;

    // containerWidth'e göre pixelsPerHour'ı ayarla ki aralık tam otursun
    pixelsPerHour =
      (containerWidth - TIMELINE_PADDING_LEFT) /
      ((end - start) / (1000 * 60 * 60));

    // Min ve Max zoom sınırları içinde kalmasını sağla
    pixelsPerHour = Math.max(MIN_ZOOM, Math.min(pixelsPerHour, MAX_ZOOM));

    scheduleRedraw();
  }

  function getCurrentInteractionMode(
    event: MouseEvent | WheelEvent
  ): InteractionMode {
    if (event.ctrlKey) return "zoom";
    if (event.shiftKey) return "pan";
    return "none";
  }

  function debounce<T extends (...args: any[]) => void>(
    func: T,
    delay: number
  ): (...args: Parameters<T>) => void {
    let timeout: NodeJS.Timeout;
    return function (this: ThisParameterType<T>, ...args: Parameters<T>) {
      const context = this;
      clearTimeout(timeout);
      timeout = setTimeout(() => func.apply(context, args), delay);
    };
  }

  function timeToX(timestamp: number): number {
    return Math.round(
      TIMELINE_PADDING_LEFT +
        ((timestamp - timelineStart) / (1000 * 60 * 60)) * pixelsPerHour
    );
  }

  function xToTime(x: number): number {
    return (
      timelineStart +
      ((x - TIMELINE_PADDING_LEFT) / pixelsPerHour) * 60 * 60 * 1000
    );
  }

  function groupFlightsByPersonIdAndType(
    items: FlightItem[],
    visibleTypesMap: Record<string, RowType[]>
  ): Record<string, FlightItem[]> {
    const map: Record<string, FlightItem[]> = {};
    const uniquePersonIds = new Set<string>();

    for (const item of items) {
      if (!item.person_id) continue;
      uniquePersonIds.add(item.person_id);
    }

    for (const personId of uniquePersonIds) {
      const typesToShowForPerson = visibleTypesMap[personId] || [
        RowType.Actual,
      ];

      const personFlights = items.filter((item) => item.person_id === personId);

      if (typesToShowForPerson.includes(RowType.Actual)) {
        const actualItems = personFlights.filter(
          (item) => item.type === RowType.Actual
        );
        const groupKey = `${personId}_${RowType.Actual}`;
        if (!map[groupKey]) {
          map[groupKey] = [];
        }
        map[groupKey].push(...actualItems);
      }
      if (typesToShowForPerson.includes(RowType.Publish)) {
        const publishItems = personFlights.filter(
          (item) => item.type === RowType.Publish
        );
        const groupKey = `${personId}_${RowType.Publish}`;
        if (!map[groupKey]) {
          map[groupKey] = [];
        }
        map[groupKey].push(...publishItems);
      }
    }

    const sortedKeys = Object.keys(map).sort((a, b) => {
      const [personIdA, typeA] = a.split("_");
      const [personIdB, typeB] = b.split("_");

      if (personIdA !== personIdB) {
        return personIdA.localeCompare(personIdB);
      }
      if (typeA === RowType.Actual && typeB === RowType.Publish) return -1;
      if (typeA === RowType.Publish && typeB === RowType.Actual) return 1;
      return 0;
    });

    const sortedMap: Record<string, FlightItem[]> = {};
    for (const key of sortedKeys) {
      sortedMap[key] = map[key];
      if (sortedMap[key]) {
        sortedMap[key].sort((a, b) => a.departure_time - b.departure_time);
      }
    }

    return sortedMap;
  }

  function getTruncatedText(
    context: CanvasRenderingContext2D,
    text: string,
    maxWidth: number
  ): string {
    const ellipsis = "...";
    const ellipsisWidth = context.measureText(ellipsis).width;
    let textWidth = context.measureText(text).width;

    if (textWidth <= maxWidth) {
      return text;
    }

    let truncatedText = text;
    while (textWidth >= maxWidth - ellipsisWidth && truncatedText.length > 0) {
      truncatedText = truncatedText.substring(0, truncatedText.length - 1);
      textWidth = context.measureText(truncatedText).width;
    }

    return truncatedText + ellipsis;
  }

  function scheduleRedraw() {
    if (!needsRedraw) {
      updateVisibleRowRange();
      needsRedraw = true;
      requestAnimationFrame(() => {
        needsRedraw = false;
        drawTimeline();
      });
    }
  }

  const debouncedScheduleRedraw = debounce(scheduleRedraw, 50);

  function checkCollision(
    movingFlight: FlightItem,
    targetGroupKey: string,
    newDepartureTime: number,
    newArrivalTime: number
  ): boolean {
    const itemsInTargetGroup = groupedFlights[targetGroupKey];
    if (!itemsInTargetGroup) return false;

    for (const existingFlight of itemsInTargetGroup) {
      if (existingFlight.data_id === movingFlight.data_id) continue;

      const existingDepTime = existingFlight.departure_time;
      const existingArrTime = existingFlight.arrival_time;

      if (
        newDepartureTime < existingArrTime &&
        newArrivalTime > existingDepTime
      ) {
        return true;
      }
    }
    return false;
  }

  function drawGridLinesOnly(
    timelineStart: number,
    timelineEnd: number,
    drawHeight: number
  ) {
    ctx.save();
    ctx.beginPath();
    ctx.rect(
      TIMELINE_PADDING_LEFT,
      0,
      canvas.width - TIMELINE_PADDING_LEFT,
      drawHeight
    );
    ctx.clip();

    const { minor: minorStepMs, major: majorStepMs } =
      getStepSizes(pixelsPerHour);

    const alignedStart = Math.floor(timelineStart / minorStepMs) * minorStepMs;

    for (
      let t = alignedStart;
      t <= timelineEnd + majorStepMs;
      t += minorStepMs
    ) {
      const x = timeToX(t);
      if (x < TIMELINE_PADDING_LEFT || x > canvas.width) continue;

      const isMajorTick = (t / minorStepMs) % (majorStepMs / minorStepMs) === 0;

      ctx.strokeStyle = isMajorTick ? "#999" : "#ccc";
      ctx.lineWidth = isMajorTick ? 1.5 : 1;
      ctx.beginPath();
      ctx.moveTo(x + 0.5, 0);
      ctx.lineTo(x + 0.5, drawHeight);
      ctx.stroke();
    }

    ctx.restore();
  }

  function getStepSizes(pixelsPerHour: number): {
    minor: number;
    major: number;
    minorFmt: SafeDateFormat;
    majorFmt: SafeDateFormat;
  } {
    const MIN = 60 * 1000;
    const HOUR = 60 * MIN;
    const DAY = 24 * HOUR;

    if (pixelsPerHour > 800) {
      return {
        minor: 5 * MIN,
        major: 15 * MIN,
        minorFmt: {
          hour: "2-digit",
          minute: "2-digit",
          hourCycle: "h23",
        },
        majorFmt: {
          hour: "2-digit",
          minute: "2-digit",
          hourCycle: "h23",
        },
      };
    } else if (pixelsPerHour > 400) {
      return {
        minor: 15 * MIN,
        major: 60 * MIN,
        minorFmt: {
          hour: "2-digit",
          minute: "2-digit",
          hourCycle: "h23",
        },
        majorFmt: {
          hour: "2-digit",
          hourCycle: "h23",
        },
      };
    } else if (pixelsPerHour > 200) {
      return {
        minor: 30 * MIN,
        major: 2 * HOUR,
        minorFmt: {
          hour: "2-digit",
          hourCycle: "h23",
        },
        majorFmt: {
          hour: "2-digit",
          hourCycle: "h23",
        },
      };
    } else if (pixelsPerHour > 150) {
      return {
        minor: 1 * HOUR,
        major: 3 * HOUR,
        minorFmt: {
          hour: "2-digit",
          hourCycle: "h23",
        },
        majorFmt: {
          hour: "2-digit",
          hourCycle: "h23",
        },
      };
    } else if (pixelsPerHour > 100) {
      return {
        minor: 3 * HOUR,
        major: 6 * HOUR,
        minorFmt: {
          hour: "2-digit",
          hourCycle: "h23",
        },
        majorFmt: {
          hour: "2-digit",
          hourCycle: "h23",
        },
      };
    } else if (pixelsPerHour > 50) {
      return {
        minor: 6 * HOUR,
        major: 12 * HOUR,
        minorFmt: {
          hour: "2-digit",
          hourCycle: "h23",
        },
        majorFmt: {
          hour: "2-digit",
          hourCycle: "h23",
        },
      };
    } else if (pixelsPerHour > 20) {
      return {
        minor: 12 * HOUR,
        major: 1 * DAY,
        minorFmt: {
          weekday: "short",
          day: "numeric",
        },
        majorFmt: {
          day: "numeric",
          month: "short",
        },
      };
    } else {
      return {
        minor: 1 * DAY,
        major: 7 * DAY,
        minorFmt: {
          day: "numeric",
        },
        majorFmt: {
          month: "short",
          year: "numeric",
        },
      };
    }
  }

  function drawTimeLabels(
    timelineStart: number,
    timelineEnd: number,
    headerHeight: number
  ) {
    ctx.save();

    ctx.fillStyle = "#e0e0e0";
    ctx.fillRect(
      TIMELINE_PADDING_LEFT,
      0,
      canvas.width - TIMELINE_PADDING_LEFT,
      headerHeight
    );

    ctx.beginPath();
    ctx.rect(
      TIMELINE_PADDING_LEFT,
      0,
      canvas.width - TIMELINE_PADDING_LEFT,
      headerHeight
    );
    ctx.clip();

    const showYearLabels = pixelsPerHour < 2.5;
    const showMonthLabels = pixelsPerHour < 80;

    const DAY_MS = 24 * 60 * 60 * 1000;
    const { minor: minorStepMs } = getStepSizes(pixelsPerHour);

    const yLayerH = showYearLabels ? headerHeight * 0.33 : 0;
    const mLayerH = showMonthLabels ? headerHeight * 0.33 : 0;
    const LABEL_PAD = 4;

    ctx.textBaseline = "top";
    ctx.textAlign = "left";

    const alignedStart = Math.floor(timelineStart / minorStepMs) * minorStepMs;
    let lastYear = -1;
    let lastMonth = -1;
    let lastLabelRightEdge = -Infinity;

    for (
      let t = alignedStart;
      t <= timelineEnd + minorStepMs;
      t += minorStepMs
    ) {
      const x = timeToX(t);
      if (x < TIMELINE_PADDING_LEFT || x > canvas.width) continue;

      ctx.beginPath();
      ctx.moveTo(x + 0.5, 0);
      ctx.lineTo(x + 0.5, headerHeight);
      ctx.strokeStyle = "#c6c6c6";
      ctx.lineWidth = 1;
      ctx.stroke();

      const offset = get(timezoneOffsetMinStore);
      const date = new Date(t - offset * 60 * 1000); // Apply offset for display
      const year = date.getFullYear();
      const month = date.getMonth();
      const weekday = date.getDay();

      if (showYearLabels && year !== lastYear) {
        ctx.font = "bold 12px sans-serif";
        ctx.fillStyle = "#111";
        ctx.fillText(String(year), x + LABEL_PAD, 0);
        lastYear = year;
      }

      if (showMonthLabels && month !== lastMonth) {
        ctx.font = "11px sans-serif";
        ctx.fillStyle = "#333";
        const monthLabel = date.toLocaleString(locale, { month: "short" });
        ctx.fillText(monthLabel, x + LABEL_PAD, yLayerH);
        lastMonth = month;
      }

      ctx.font = "10px sans-serif";
      let label = "";

      if (minorStepMs >= DAY_MS) {
        const dayNum = date.getDate();
        const weekdayLabel = date.toLocaleString(locale, { weekday: "short" });
        label = `${dayNum} ${weekdayLabel}`;
        ctx.fillStyle = weekday === 0 ? "darkred" : "#555";
      } else {
        const hour = date.getHours().toString().padStart(2, "0");
        const minute = date.getMinutes().toString().padStart(2, "0");
        label = `${hour}.${minute}`;
        ctx.fillStyle = "#666";
      }

      const textWidth = ctx.measureText(label).width;
      const labelLeft = x + LABEL_PAD;
      const labelRight = labelLeft + textWidth;

      if (labelLeft > lastLabelRightEdge + 4) {
        ctx.fillText(label, labelLeft, yLayerH + mLayerH);
        lastLabelRightEdge = labelRight;
      }
    }

    ctx.restore();
  }

  function drawNowLine(drawHeight: number) {
    ctx.save();
    ctx.beginPath();
    ctx.rect(
      TIMELINE_PADDING_LEFT,
      0,
      canvas.width - TIMELINE_PADDING_LEFT,
      drawHeight
    );
    ctx.clip();

    const x = timeToX(Date.now());
    if (x >= TIMELINE_PADDING_LEFT && x <= canvas.width) {
      ctx.strokeStyle = "red";
      ctx.lineWidth = 2;
      ctx.beginPath();
      ctx.moveTo(x, 0);
      ctx.lineTo(x, drawHeight);
      ctx.stroke();
      ctx.lineWidth = 1;
    }
    ctx.restore();
  }

  let firstVisibleRowIndex = 0;
  let lastVisibleRowIndex = 0;

  function updateVisibleRowRange() {
    const groupKeys = Object.keys(groupedFlights).sort((a, b) => {
      const [personIdA, typeA] = a.split("_");
      const [personIdB, typeB] = b.split("_");

      if (personIdA !== personIdB) {
        return personIdA.localeCompare(personIdB);
      }
      if (typeA === RowType.Actual && typeB === RowType.Publish) return -1;
      if (typeA === RowType.Publish && typeB === RowType.Actual) return 1;
      return 0;
    });
    const numRows = groupKeys.length;

    const buffer = 10;

    firstVisibleRowIndex = Math.max(
      0,
      Math.floor((scrollY - HEADER_HEIGHT) / (ROW_HEIGHT + ROW_GAP)) - buffer
    );

    lastVisibleRowIndex = Math.min(
      numRows - 1,
      Math.ceil(
        (scrollY + containerHeight - HEADER_HEIGHT) / (ROW_HEIGHT + ROW_GAP)
      ) + buffer
    );
  }

  const SIDEBAR = {
    padX: 0,
    padY: 4,
    fontReg: "10px sans-serif",
    fontBold: "600 10px sans-serif",
    cols: 3,
    rows: 2,
  };

  function paintCellBackground(
    x: number,
    y: number,
    w: number,
    h: number,
    baseFill: string,
    alertFill?: string
  ) {
    ctx.fillStyle = alertFill ?? baseFill;
    ctx.fillRect(x, y, w, h);
  }

  function drawLeftSidebar(
    groupKeys: string[],
    firstIdx: number,
    lastIdx: number
  ) {
    ctx.save();
    ctx.textAlign = "left";
    ctx.textBaseline = "middle";

    const cellW = (TIMELINE_PADDING_LEFT - SIDEBAR.padX * 2) / SIDEBAR.cols;
    const cellH = (ROW_HEIGHT - SIDEBAR.padY * 2) / SIDEBAR.rows;

    const maxRow = Math.min(lastIdx + 1, groupKeys.length);
    for (let i = firstIdx; i < maxRow; i++) {
      const key = groupKeys[i];
      if (!key) continue;

      const [personId, rowType] = key.split("_");

      const yRow = HEADER_HEIGHT + i * (ROW_HEIGHT + ROW_GAP);
      let zebraFill = i % 2 === 0 ? "#f4f4f4" : "#fcfcfc";
      let rowFill = rowType === RowType.Publish ? "#e0e7ee" : zebraFill;

      const hasFTLViolations =
        rowFTLViolations[key] && rowFTLViolations[key].length > 0;

      if (highlightedKeys.has(personId)) {
        rowFill = "rgba(255, 215, 0, 0.35)";
      } else if (
        isGlobalRowDragging &&
        globalSourceTimelineId !== timelineId &&
        currentDropTargetRowKey === key
      ) {
        rowFill = DRAG_OVER_COLOR;
      } else if (isItemDragging && currentDropTargetRowKey === key) {
        rowFill = isDroppingOverCollision ? COLLISION_COLOR : DRAG_OVER_COLOR;
      } else if (hasFTLViolations) {
        rowFill = FTL_VIOLATION_COLORS.ROW_BACKGROUND_DARK;
      }

      // Row arkaplanı çizimi
      ctx.fillStyle = rowFill;
      ctx.fillRect(0, yRow, TIMELINE_PADDING_LEFT, ROW_HEIGHT);

      // ✅ RowType.Publish satırına sol çizgi ekle
      if (rowType === RowType.Publish) {
        ctx.strokeStyle = "#1a73e8"; // Google-mavisi tonu
        ctx.lineWidth = 3;
        ctx.beginPath();
        ctx.moveTo(3, yRow); // 0 yerine 3 px içeriden başlat
        ctx.lineTo(3, yRow + ROW_HEIGHT);
        ctx.stroke();
      }

      // Bilgiler
      const items = groupedFlights[key];
      const firstItem = items?.[0];
      const durationM = totalDurations[key] ?? 0;
      const durH = (durationM / 60).toFixed(2);

      const cells: (string | null)[] = [
        firstItem?.name ?? "",
        firstItem?.class ?? "-",
        `Uçuş: ${durH}s`,
        firstItem?.surname ?? "-",
        `${personId} (${rowType === RowType.Actual ? "ACT" : "PLN"})`,
        `Görev: ${durH}s`,
      ];

      ctx.font = SIDEBAR.fontReg;
      for (let idx = 0; idx < cells.length; idx++) {
        const col = idx % SIDEBAR.cols;
        const row = Math.floor(idx / SIDEBAR.cols);

        const baseX = SIDEBAR.padX + col * cellW;
        const baseY = yRow + SIDEBAR.padY + row * cellH;

        let alertFill: string | undefined;
        if (hasFTLViolations && (idx === 2 || idx === 5)) {
          alertFill = FTL_VIOLATION_COLORS.ALERT_CELL_BACKGROUND;
        }

        if (alertFill)
          paintCellBackground(
            baseX,
            yRow + SIDEBAR.padY + row * cellH,
            cellW,
            cellH,
            rowFill,
            alertFill
          );

        ctx.font = idx === 0 ? SIDEBAR.fontBold : SIDEBAR.fontReg;
        const txt = getTruncatedText(ctx, cells[idx]!, cellW - 4);
        ctx.fillStyle = "#333";
        ctx.fillText(txt, baseX + 2, baseY + cellH / 2);
      }

      if (hasFTLViolations) {
        ctx.font = "bold 14px sans-serif";
        ctx.fillStyle = "red";
        ctx.textAlign = "right";
        ctx.textBaseline = "top";
        ctx.fillText(
          FTL_VIOLATION_COLORS.WARNING_ICON,
          TIMELINE_PADDING_LEFT - 5,
          yRow + 5
        );
      }
    }

    ctx.restore();
  }

  function drawTimeline() {
    if (!ctx || !canvas) return;

    calculateTotalDurations();

    const groupKeys = Object.keys(groupedFlights).sort((a, b) => {
      const [personIdA, typeA] = a.split("_");
      const [personIdB, typeB] = b.split("_");

      if (personIdA !== personIdB) {
        return personIdA.localeCompare(personIdB);
      }
      if (typeA === RowType.Actual && typeB === RowType.Publish) return -1;
      if (typeA === RowType.Publish && typeB === RowType.Actual) return 1;
      return 0;
    });

    canvas.width = containerWidth;
    canvas.height = containerHeight;
    ctx.clearRect(0, 0, canvas.width, containerHeight);

    ctx.fillStyle = "#f0f0f0";
    ctx.fillRect(0, 0, TIMELINE_PADDING_LEFT, containerHeight);

    ctx.fillStyle = "#e0e0e0";
    ctx.fillRect(0, 0, containerWidth, HEADER_HEIGHT);

    ctx.save();
    ctx.translate(0, -scrollY);

    drawLeftSidebar(groupKeys, firstVisibleRowIndex, lastVisibleRowIndex);

    ctx.textAlign = "left";
    ctx.font = "12px sans-serif";

    for (let i = firstVisibleRowIndex; i <= lastVisibleRowIndex; i++) {
      const key = groupKeys[i];
      const [personId, rowType] = key.split("_");
      const baseY = HEADER_HEIGHT + i * (ROW_HEIGHT + ROW_GAP);
      const rowHeight = ROW_HEIGHT;
      let hasFTLViolations =
        rowFTLViolations[key] && rowFTLViolations[key].length > 0;

      let rowBackgroundFill = i % 2 === 0 ? "#f8f8f8" : "#ffffff";
      if (rowType === RowType.Publish) {
        rowBackgroundFill = i % 2 === 0 ? "#f0f5fa" : "#e0e7ee";
      }

      if (highlightedKeys.has(personId)) {
        rowBackgroundFill = "rgba(255, 215, 0, 0.25)";
      } else if (
        isGlobalRowDragging &&
        globalSourceTimelineId !== timelineId &&
        currentDropTargetRowKey === key
      ) {
        rowBackgroundFill = DRAG_OVER_COLOR;
      } else if (isItemDragging && currentDropTargetRowKey === key) {
        rowBackgroundFill = isDroppingOverCollision
          ? COLLISION_COLOR
          : DRAG_OVER_COLOR;
      } else if (hasFTLViolations) {
        rowBackgroundFill = FTL_VIOLATION_COLORS.ROW_BACKGROUND_LIGHT;
      }

      ctx.fillStyle = rowBackgroundFill;
      ctx.fillRect(
        TIMELINE_PADDING_LEFT,
        baseY,
        canvas.width - TIMELINE_PADDING_LEFT,
        ROW_HEIGHT
      );

      ctx.fillStyle = "#222";
      ctx.strokeStyle = "#eee";
      ctx.beginPath();
      ctx.moveTo(0, baseY + ROW_HEIGHT);
      ctx.lineTo(canvas.width, baseY + ROW_HEIGHT);
      ctx.stroke();
    }

    ctx.save();
    ctx.globalAlpha = 0.3;
    const maxDrawHeight = scrollY + containerHeight + 200;
    drawGridLinesOnly(timelineStart, timelineEnd, maxDrawHeight);
    ctx.restore();

    for (let i = firstVisibleRowIndex; i <= lastVisibleRowIndex; i++) {
      const key = groupKeys[i];
      const baseY = HEADER_HEIGHT + i * (ROW_HEIGHT + ROW_GAP);
      const rowHeight = ROW_HEIGHT;
      const itemsForThisGroup = groupedFlights[key];
      if (!itemsForThisGroup) continue;

      // DRAG KORUMALI BoundingBox (sürüklenen item'ın x/y'sini EZME)
      for (const item of itemsForThisGroup) {
        const isDragged =
          isItemDragging && draggedFlight?.data_id === item.data_id;

        if (isDragged && item._boundingBox) {
          // mousemove'da set edilmiş x/y'yi KORU, sadece w/h'yi güncelle
          const startX = item._boundingBox.x;
          const endX = timeToX(item.arrival_time);
          const width = Math.max(endX - startX, FLIGHT_ITEM_MIN_WIDTH);

          item._boundingBox = {
            ...item._boundingBox,
            width,
            height: rowHeight - 2 * FLIGHT_ITEM_VERTICAL_PADDING,
          };
        } else {
          // normal akış
          const x = timeToX(item.departure_time);
          const endX = timeToX(item.arrival_time);
          const width = Math.max(endX - x, FLIGHT_ITEM_MIN_WIDTH);

          item._boundingBox = {
            x,
            y: baseY + FLIGHT_ITEM_VERTICAL_PADDING,
            width,
            height: rowHeight - 2 * FLIGHT_ITEM_VERTICAL_PADDING,
          };
        }
      }

      for (let j = 0; j < itemsForThisGroup.length; j++) {
        const item = itemsForThisGroup[j];

        if (
          (item.data_id === draggedFlight?.data_id && isItemDragging) ||
          (isGlobalRowDragging &&
            globalDraggedGroupKey === key &&
            globalSourceTimelineId === timelineId)
        )
          continue;

        if (!item._boundingBox) continue;

        ctx.save();
        ctx.beginPath();
        ctx.rect(
          TIMELINE_PADDING_LEFT,
          baseY,
          canvas.width - TIMELINE_PADDING_LEFT,
          rowHeight
        );
        ctx.clip();

        const baseColor = getActivityColor(item.activity_code);
        const itemColor = baseColor;

        ctx.fillStyle =
          hoveredFlight?.data_id === item.data_id
            ? "#ffc107"
            : clickedFlight?.data_id === item.data_id
              ? "#007bff"
              : itemColor;

        const drawX = Math.max(item._boundingBox.x, TIMELINE_PADDING_LEFT);
        const actualDrawingWidth =
          item._boundingBox.width -
          (TIMELINE_PADDING_LEFT - item._boundingBox.x > 0
            ? TIMELINE_PADDING_LEFT - item._boundingBox.x
            : 0);

        ctx.fillRect(
          drawX,
          item._boundingBox.y,
          Math.min(actualDrawingWidth, canvas.width - drawX),
          item._boundingBox.height
        );

        if (
          item.rest_start != null &&
          item.rest_end != null &&
          item.rest_duration != null
        ) {
          const restStartX = timeToX(item.rest_start);
          const restEndX = timeToX(item.rest_end);
          const restWidth = restEndX - restStartX;

          const restBoxY = item._boundingBox.y + item._boundingBox.height - 4;
          ctx.fillStyle = "#9ec9ff";
          ctx.fillRect(restStartX, restBoxY, Math.max(restWidth, 2), 4);
        }

        const textToDraw =
          item.activity_code === "FLT"
            ? String(item.flight_no)
            : item.activity_code;

        const boxX = item._boundingBox.x;
        const boxWidth = item._boundingBox.width;
        const boxRight = boxX + boxWidth;

        if (boxRight > 0 && boxX < canvas.width) {
          const centerX = item._boundingBox.x + item._boundingBox.width / 2;
          const centerY = item._boundingBox.y + item._boundingBox.height / 2;

          const textWidth = ctx.measureText(textToDraw).width;

          if (textWidth < item._boundingBox.width - 4) {
            ctx.fillStyle = "black";
            ctx.font = "10px sans-serif";
            ctx.textAlign = "center";
            ctx.textBaseline = "middle";
            ctx.fillText(textToDraw, centerX, centerY);
          }

          const drawX = Math.max(item._boundingBox.x, TIMELINE_PADDING_LEFT);
          const actualDrawingWidthForText =
            item._boundingBox.width -
            (TIMELINE_PADDING_LEFT - item._boundingBox.x > 0
              ? TIMELINE_PADDING_LEFT - item._boundingBox.x
              : 0);

          const portsFit = actualDrawingWidthForText > 100;
          if (portsFit) {
            const depTime = new Date(item.departure_time).toLocaleTimeString(
              locale,
              {
                hour: "2-digit",
                minute: "2-digit",
                hourCycle: "h23",
              }
            );
            const arrTime = new Date(item.arrival_time).toLocaleTimeString(
              locale,
              {
                hour: "2-digit",
                minute: "2-digit",
                hourCycle: "h23",
              }
            );

            ctx.font = "bold 8px sans-serif";
            ctx.fillStyle = "black";
            ctx.textBaseline = "bottom";

            ctx.textAlign = "left";
            ctx.fillText(
              depTime,
              drawX + 2,
              item._boundingBox.y + item._boundingBox.height - 11
            );

            ctx.textAlign = "right";
            ctx.fillText(
              arrTime,
              drawX + actualDrawingWidthForText - 2,
              item._boundingBox.y + item._boundingBox.height - 11
            );

            ctx.font = "bold 9px sans-serif";
            ctx.fillStyle = "#444";
            ctx.textBaseline = "bottom";

            ctx.textAlign = "left";
            ctx.fillText(
              item.departure_port ?? "",
              drawX + 2,
              item._boundingBox.y + item._boundingBox.height - 2
            );

            ctx.textAlign = "right";
            ctx.fillText(
              item.arrival_port ?? "",
              drawX + actualDrawingWidthForText - 2,
              item._boundingBox.y + item._boundingBox.height - 2
            );
          }
        }

        const nextItem = itemsForThisGroup[j + 1];
        if (
          nextItem &&
          nextItem._boundingBox &&
          item.trip_id &&
          nextItem.trip_id === item.trip_id
        ) {
          if (item.arrival_time <= nextItem.departure_time) {
            const startPointX = item._boundingBox.x + item._boundingBox.width;
            const startPointY =
              item._boundingBox.y + item._boundingBox.height / 2;
            const endPointX = nextItem._boundingBox.x;
            const endPointY =
              nextItem._boundingBox.y + nextItem._boundingBox.height / 2;

            ctx.strokeStyle = "#669966";
            ctx.lineWidth = 5;
            ctx.beginPath();
            ctx.moveTo(startPointX, startPointY);
            ctx.lineTo(endPointX, endPointY);
            ctx.stroke();

            ctx.fillStyle = "#669966";
            const rectWidth = 4;
            const rectHeight = 10;

            ctx.fillRect(
              startPointX - rectWidth / 2,
              startPointY - rectHeight / 2,
              rectWidth,
              rectHeight
            );

            ctx.fillRect(
              endPointX - rectWidth / 2,
              endPointY - rectHeight / 2,
              rectWidth,
              rectHeight
            );
          }
        }
        ctx.restore();
      }
    }

    ctx.restore();

    ctx.strokeStyle = "#999";
    ctx.lineWidth = 1;
    ctx.beginPath();
    ctx.moveTo(TIMELINE_PADDING_LEFT, 0);
    ctx.lineTo(TIMELINE_PADDING_LEFT, containerHeight);
    ctx.stroke();

    ctx.beginPath();
    ctx.moveTo(0, HEADER_HEIGHT);
    ctx.lineTo(containerWidth, HEADER_HEIGHT);
    ctx.stroke();

    drawLeftSidebarHeader();
    drawTimeLabels(timelineStart, timelineEnd, HEADER_HEIGHT);
    drawNowLine(containerHeight);

    if (isItemDragging && draggedFlight && draggedFlight._boundingBox) {
      ctx.save();
      ctx.globalAlpha = DRAGGING_ITEM_ALPHA;

      const drawX = Math.max(
        draggedFlight._boundingBox.x,
        TIMELINE_PADDING_LEFT
      );
      const drawY = draggedFlight._boundingBox.y - scrollY;

      const actualDrawingWidth =
        draggedFlight._boundingBox.width -
        (TIMELINE_PADDING_LEFT - draggedFlight._boundingBox.x > 0
          ? TIMELINE_PADDING_LEFT - draggedFlight._boundingBox.x
          : 0);

      ctx.fillStyle = isDroppingOverCollision
        ? COLLISION_COLOR
        : clickedFlight?.data_id === draggedFlight.data_id
          ? "#007bff"
          : "#28a745";

      ctx.fillRect(
        drawX,
        drawY,
        Math.min(actualDrawingWidth, canvas.width - drawX),
        draggedFlight._boundingBox.height
      );

      const textToDraw = draggedFlight.activity_code;
      if (
        draggedFlight._boundingBox.width >
          ctx.measureText(textToDraw).width + 2 * FLIGHT_TEXT_OFFSET_X &&
        drawX + ctx.measureText(textToDraw).width + FLIGHT_TEXT_OFFSET_X <
          canvas.width
      ) {
        ctx.fillStyle = "black";
        ctx.font = "10px sans-serif";
        ctx.textAlign = "left";
        ctx.fillText(
          textToDraw,
          drawX + FLIGHT_TEXT_OFFSET_X,
          drawY + FLIGHT_TEXT_OFFSET_Y
        );
      }

      ctx.restore();
    }
  }

  function calculateTotalDurations() {
    const groupKeys = Object.keys(groupedFlights).sort((a, b) => {
      const [personIdA, typeA] = a.split("_");
      const [personIdB, typeB] = b.split("_");

      if (personIdA !== personIdB) {
        return personIdA.localeCompare(personIdB);
      }
      if (typeA === RowType.Actual && typeB === RowType.Publish) return -1;
      if (typeA === RowType.Publish && typeB === RowType.Actual) return 1;
      return 0;
    });

    if (JSON.stringify(groupKeys) === JSON.stringify(lastGroupKeys)) return;

    lastGroupKeys = groupKeys;

    totalDurations = {};
    for (const key of groupKeys) {
      const activities = groupedFlights[key];
      totalDurations[key] =
        activities?.reduce((acc, act) => {
          if (
            act.group_code === "FLT" &&
            typeof act.departure_time === "number" &&
            typeof act.arrival_time === "number"
          ) {
            const depTime = act.departure_time;
            const arrTime = act.arrival_time;
            return acc + (arrTime - depTime) / 60_000;
          }
          return acc;
        }, 0) ?? 0;
    }

    console.log("🧮 totalDurations updated:", totalDurations);
  }

  function scrollToFirstHighlighted() {
    const groupKeys = Object.keys(groupedFlights).sort((a, b) => {
      const [personIdA, typeA] = a.split("_");
      const [personIdB, typeB] = b.split("_");

      if (personIdA !== personIdB) {
        return personIdA.localeCompare(personIdB);
      }
      if (typeA === RowType.Actual && typeB === RowType.Publish) return -1;
      if (typeA === RowType.Publish && typeB === RowType.Actual) return 1;
      return 0;
    });
    const firstIdx = groupKeys.findIndex((k) =>
      highlightedKeys.has(k.split("_")[0])
    );
    if (firstIdx === -1) return;

    const targetY = HEADER_HEIGHT + firstIdx * (ROW_HEIGHT + ROW_GAP);
    const container = canvasContainerRef;
    if (container) {
      container.scrollTo({ top: targetY, behavior: "smooth" });
    }
  }

  function isInDropdownHeader(x: number, y: number): boolean {
    const width = TIMELINE_PADDING_LEFT;
    const iconX = width - DROPDOWN_ICON.size - DROPDOWN_ICON.pad;
    const iconY = (HEADER_HEIGHT - DROPDOWN_ICON.size) / 2;

    const isInside =
      x >= iconX &&
      x <= iconX + DROPDOWN_ICON.size &&
      y >= iconY &&
      y <= iconY + DROPDOWN_ICON.size;
    return isInside;
  }

  function drawLeftSidebarHeader() {
    const width = TIMELINE_PADDING_LEFT;
    const height = HEADER_HEIGHT;

    const leftWidth = width * 0.5;
    const rightWidth = width - leftWidth;
    const halfHeight = height / 2;

    const uniquePersonIds = new Set(
      Object.keys(groupedFlights).map((key) => key.split("_")[0])
    );
    const rowCount = uniquePersonIds.size;
    const itemCount = Object.values(groupedFlights).reduce(
      (acc, items) => acc + items.length,
      0
    );

    ctx.save();

    ctx.fillStyle = "#eaeaea";
    ctx.fillRect(0, 0, width, height);

    ctx.strokeStyle = "#000";
    ctx.lineWidth = 0.2;
    ctx.beginPath();
    ctx.moveTo(leftWidth, 0);
    ctx.lineTo(leftWidth, height);
    ctx.stroke();

    ctx.beginPath();
    ctx.moveTo(leftWidth, halfHeight);
    ctx.lineTo(width, halfHeight);
    ctx.stroke();

    ctx.font = "bold 12px sans-serif";
    ctx.fillStyle = "#000";
    ctx.textAlign = "center";
    ctx.textBaseline = "middle";
    ctx.fillText("ROSTER", leftWidth / 2, height / 2);

    ctx.font = "11px sans-serif";
    ctx.fillStyle = "#666";
    ctx.fillText(`${rowCount} row`, leftWidth + rightWidth / 2, halfHeight / 2);
    ctx.fillText(
      `${itemCount.toLocaleString(locale)} item`,
      leftWidth + rightWidth / 2,
      halfHeight + halfHeight / 2
    );

    if (highlightedKeys && highlightedKeys.size > 0) {
      const label = [...highlightedKeys].join(", ");
      ctx.font = "10px sans-serif";
      ctx.fillStyle = "#c00";
      ctx.textAlign = "left";
      ctx.textBaseline = "bottom";
      ctx.fillText(`Highlight: ${label}`, 4, height - 2);
    }

    ctx.restore();
  }

  function handleWheel(event: WheelEvent) {
    if (isGlobalRowDragging) {
      event.preventDefault();
      return;
    }

    const mode = getCurrentInteractionMode(event);

    if (mode === "zoom" || mode === "pan") {
      event.preventDefault();

      const rect = canvas.getBoundingClientRect();
      const mouseX = event.clientX - rect.left;

      if (mode === "zoom") {
        const centerTime = xToTime(mouseX);
        const delta = Math.sign(event.deltaY);
        const zoomFactor = 1.1;

        pixelsPerHour =
          delta > 0
            ? Math.max(pixelsPerHour / zoomFactor, MIN_ZOOM)
            : Math.min(pixelsPerHour * zoomFactor, MAX_ZOOM);

        timelineStart =
          centerTime -
          ((mouseX - TIMELINE_PADDING_LEFT) / pixelsPerHour) * 60 * 60 * 1000;
        timelineEnd =
          timelineStart +
          ((containerWidth - TIMELINE_PADDING_LEFT) / pixelsPerHour) *
            60 *
            60 *
            1000;
      } else if (mode === "pan") {
        const panAmount = event.deltaY * 5;
        const deltaMs = (panAmount / pixelsPerHour) * 60 * 60 * 1000;

        timelineStart += deltaMs;
        timelineEnd += deltaMs;
      }

      scheduleRedraw();
    }
  }

  function handleMouseDown(event: MouseEvent) {
    if (isGlobalRowDragging) return;

    const rect = canvas.getBoundingClientRect();
    const clientX = event.clientX - rect.left;
    const clientY = event.clientY - rect.top;

    const adjustedMouseYForContent = clientY + scrollY;
    const isClickInSidebar = clientX < TIMELINE_PADDING_LEFT;
    const isClickInTimelineContent = clientX >= TIMELINE_PADDING_LEFT;

    const groupKeys = Object.keys(groupedFlights).sort((a, b) => {
      const [personIdA, typeA] = a.split("_");
      const [personIdB, typeB] = b.split("_");
      if (personIdA !== personIdB) return personIdA.localeCompare(personIdB);
      if (typeA === RowType.Actual && typeB === RowType.Publish) return -1;
      if (typeA === RowType.Publish && typeB === RowType.Actual) return 1;
      return 0;
    });

    const rowClickedIndex = Math.floor(
      (adjustedMouseYForContent - HEADER_HEIGHT) / (ROW_HEIGHT + ROW_GAP)
    );

    const clickedRowKey =
      rowClickedIndex >= 0 && rowClickedIndex < groupKeys.length
        ? groupKeys[rowClickedIndex]
        : null;

    // --- DRAG / PAN HANDLING (Left-click) ---
    // Only allow flight/row drag if 'M' key is pressed.
    // Allow panning always (if not dragging an item or row).

    // Row Drag (only with M key)
    if (isClickInSidebar && clickedRowKey) {
      if (!isMKeyDown) return; // Require 'M' for row drag

      const activitiesInGroup = groupedFlights[clickedRowKey];
      dispatch("rowDragStart", {
        groupKey: clickedRowKey,
        flightsInGroup: activitiesInGroup,
        timelineId: timelineId,
        clientX: event.clientX,
        clientY: event.clientY,
        dragOffsetY:
          adjustedMouseYForContent -
          (HEADER_HEIGHT + rowClickedIndex * (ROW_HEIGHT + ROW_GAP)),
      });
      currentCursor = "grabbing";
      event.preventDefault();
      window.addEventListener("mousemove", handleMouseMove);
      window.addEventListener("mouseup", handleMouseUp);
      return;
    }

    // Timeline content: Flight Item Drag or Pan
    if (isClickInTimelineContent) {
      let foundFlightToDrag = false;
      const numRows = groupKeys.length;

      const firstVisibleRowIndex = Math.max(
        0,
        Math.floor(
          (adjustedMouseYForContent - HEADER_HEIGHT) / (ROW_HEIGHT + ROW_GAP)
        ) - 10
      );
      const lastVisibleRowIndex = Math.min(
        numRows - 1,
        Math.ceil(
          (adjustedMouseYForContent + containerHeight - HEADER_HEIGHT) /
            (ROW_HEIGHT + ROW_GAP)
        ) + 10
      );

      for (let i = firstVisibleRowIndex; i <= lastVisibleRowIndex; i++) {
        const key = groupKeys[i];
        const items = groupedFlights[key];
        if (items) {
          for (const item of items) {
            const bbox = item._boundingBox;
            if (
              bbox &&
              clientX >= bbox.x &&
              clientX <= bbox.x + bbox.width &&
              adjustedMouseYForContent >= bbox.y &&
              adjustedMouseYForContent <= bbox.y + bbox.height
            ) {
              if (isMKeyDown) {
                // Only start drag if 'M' key is down
                draggedFlight = item;
                isItemDragging = true;
                dragOffsetX = clientX - bbox.x;
                dragOffsetY = adjustedMouseYForContent - bbox.y;
                (draggedFlight as any)._originalGroupKey = key;
                currentCursor = "grabbing";
              }
              clickedFlight = item; // Clicked/selected regardless of 'M' key
              scheduleRedraw();
              foundFlightToDrag = true;
              break;
            }
          }
        }
        if (foundFlightToDrag) break;
      }

      if (foundFlightToDrag) {
        // If an item was found AND M was down, then we're dragging.
        // Otherwise, it was just a click to select, no mousemove listener needed for dragging.
        if (isMKeyDown) {
          window.addEventListener("mousemove", handleMouseMove);
          window.addEventListener("mouseup", handleMouseUp);
        }
        return;
      } else {
        // If no flight was clicked, it's a pan operation (always allowed for left-click)
        const mode = getCurrentInteractionMode(event); // Zoom/pan with Ctrl/Shift
        if (mode === "pan" || event.button === 0) {
          // Normal left-click or Shift+Click for pan
          isDragging = true;
          lastMouseX = event.clientX;
          currentCursor = "grab";
          window.addEventListener("mousemove", handleMouseMove);
          window.addEventListener("mouseup", handleMouseUp);
          return;
        }
      }
    }

    // Default cursor if no specific action initiated
    currentCursor = "grab";
  }

  // --- handleMouseMove fonksiyonu: Panning ve hover algılama ---
  function handleMouseMove(event: MouseEvent) {
    if (!canvas || !canvasContainerRef) return;

    // Global Row Dragging (from another timeline) - always allowed
    if (isGlobalRowDragging) {
      currentCursor = "grabbing";

      const rect = canvasContainerRef.getBoundingClientRect();
      const mouseXRelativeToContainer = event.clientX - rect.left;
      const mouseYRelativeToContainer = event.clientY - rect.top;

      if (
        mouseXRelativeToContainer >= 0 &&
        mouseXRelativeToContainer <= rect.width &&
        mouseYRelativeToContainer >= 0 &&
        mouseYRelativeToContainer <= rect.height
      ) {
        const adjustedMouseYForContent = mouseYRelativeToContainer + scrollY;
        const groupKeys = Object.keys(groupedFlights).sort((a, b) => {
          const [personIdA, typeA] = a.split("_");
          const [personIdB, typeB] = b.split("_");

          if (personIdA !== personIdB) {
            return personIdA.localeCompare(personIdB);
          }
          if (typeA === RowType.Actual && typeB === RowType.Publish) return -1;
          if (typeA === RowType.Publish && typeB === RowType.Actual) return 1;
          return 0;
        });
        const currentMouseRowIndex = Math.floor(
          (adjustedMouseYForContent - HEADER_HEIGHT) / (ROW_HEIGHT + ROW_GAP)
        );

        let newDropTargetRowKey: string | null = null;
        if (
          currentMouseRowIndex >= 0 &&
          currentMouseRowIndex < groupKeys.length
        ) {
          newDropTargetRowKey = groupKeys[currentMouseRowIndex];
        }

        if (currentDropTargetRowKey !== newDropTargetRowKey) {
          currentDropTargetRowKey = newDropTargetRowKey;
          scheduleRedraw();
        }
      } else {
        if (currentDropTargetRowKey !== null) {
          currentDropTargetRowKey = null;
          scheduleRedraw();
        }
      }
      return;
    }

    const rect = canvas.getBoundingClientRect();
    const clientX = event.clientX - rect.left;
    const clientY = event.clientY - rect.top;

    mouseCanvasX = clientX;
    mouseCanvasY = clientY + scrollY;

    // Item Dragging (requires M key to be down)
    if (isItemDragging && draggedFlight && draggedFlight._boundingBox) {
      if (!isMKeyDown) {
        // Stop dragging if M key is released mid-drag
        handleMouseUp(); // Simulate mouse up to finalize or cancel drag
        return;
      }

      draggedFlight._boundingBox.x = clientX - dragOffsetX;
      currentCursor = "grabbing";

      const groupKeys = Object.keys(groupedFlights).sort((a, b) => {
        const [personIdA, typeA] = a.split("_");
        const [personIdB, typeB] = b.split("_");

        if (personIdA !== personIdB) {
          return personIdA.localeCompare(personIdB);
        }
        if (typeA === RowType.Actual && typeB === RowType.Publish) return -1;
        if (typeA === RowType.Publish && typeB === RowType.Actual) return 1;
        return 0;
      });
      const currentMouseRowIndex = Math.floor(
        (mouseCanvasY - HEADER_HEIGHT) / (ROW_HEIGHT + ROW_GAP)
      );
      let newDropTargetRowKey: string | null = null;
      let newDropTargetRowY: number | null = null;

      if (
        currentMouseRowIndex >= 0 &&
        currentMouseRowIndex < groupKeys.length
      ) {
        newDropTargetRowKey = groupKeys[currentMouseRowIndex];
        newDropTargetRowY =
          HEADER_HEIGHT +
          currentMouseRowIndex * (ROW_HEIGHT + ROW_GAP) +
          FLIGHT_ITEM_VERTICAL_PADDING;
      }

      if (newDropTargetRowY !== null) {
        draggedFlight._boundingBox.y = newDropTargetRowY;
      }

      let newIsDroppingOverCollision = false;
      if (newDropTargetRowKey && draggedFlight._boundingBox) {
        const newDepartureTime = xToTime(draggedFlight._boundingBox.x);
        const flightDuration =
          draggedFlight.arrival_time - draggedFlight.departure_time;
        const newArrivalTime = newDepartureTime + flightDuration;

        const [draggedPersonId, draggedType] = (
          draggedFlight._originalGroupKey || ""
        ).split("_");
        const [targetPersonId, targetType] = newDropTargetRowKey.split("_");

        if (draggedType !== targetType) {
          newIsDroppingOverCollision = true;
        } else {
          newIsDroppingOverCollision = checkCollision(
            draggedFlight,
            newDropTargetRowKey,
            newDepartureTime,
            newArrivalTime
          );
        }
      }

      if (
        clientX < TIMELINE_PADDING_LEFT ||
        (newDropTargetRowKey && newIsDroppingOverCollision)
      ) {
        currentCursor = "not-allowed";
      } else {
        currentCursor = "grabbing";
      }

      if (
        currentDropTargetRowKey !== newDropTargetRowKey ||
        isDroppingOverCollision !== newIsDroppingOverCollision
      ) {
        currentDropTargetRowKey = newDropTargetRowKey;
        isDroppingOverCollision = newIsDroppingOverCollision;
        scheduleRedraw();
      } else {
        debouncedScheduleRedraw();
      }

      const AUTO_SCROLL_MARGIN = 50;
      const AUTO_SCROLL_SPEED = 10;

      if (clientY < AUTO_SCROLL_MARGIN && scrollY > 0) {
        scrollY = Math.max(0, scrollY - AUTO_SCROLL_SPEED);
        scheduleRedraw();
      } else if (
        clientY > containerHeight - AUTO_SCROLL_MARGIN &&
        scrollY < totalContentHeight - containerHeight
      ) {
        scrollY = Math.min(
          totalContentHeight - containerHeight,
          scrollY + AUTO_SCROLL_SPEED
        );
        scheduleRedraw();
      }
    } else if (isDragging) {
      // Panning (no M key required for initiation or movement)
      const deltaX = event.clientX - lastMouseX;
      lastMouseX = event.clientX;

      const deltaMs = (deltaX / pixelsPerHour) * 60 * 60 * 1000;

      timelineStart -= deltaMs;
      timelineEnd =
        timelineStart +
        ((containerWidth - TIMELINE_PADDING_LEFT) / pixelsPerHour) *
          60 *
          60 *
          1000;

      debouncedScheduleRedraw();
    } else {
      // Hover effects: active when not dragging or panning
      let foundHover = false;
      let newHoveredFlight: FlightItem | null = null;

      if (mouseCanvasX >= TIMELINE_PADDING_LEFT) {
        const groupKeys = Object.keys(groupedFlights).sort((a, b) => {
          const [personIdA, typeA] = a.split("_");
          const [personIdB, typeB] = b.split("_");

          if (personIdA !== personIdB) {
            return personIdA.localeCompare(personIdB);
          }
          if (typeA === RowType.Actual && typeB === RowType.Publish) return -1;
          if (typeA === RowType.Publish && typeB === RowType.Actual) return 1;
          return 0;
        });
        const numRows = groupKeys.length;
        const adjustedClientY = clientY + scrollY;

        const firstVisibleRowIndex = Math.max(
          0,
          Math.floor(
            (adjustedClientY - HEADER_HEIGHT) / (ROW_HEIGHT + ROW_GAP)
          ) - 10
        );
        const lastVisibleRowIndex = Math.min(
          numRows - 1,
          Math.ceil(
            (adjustedClientY + containerHeight - HEADER_HEIGHT) /
              (ROW_HEIGHT + ROW_GAP)
          ) + 10
        );

        for (let i = firstVisibleRowIndex; i <= lastVisibleRowIndex; i++) {
          const key = groupKeys[i];
          const items = groupedFlights[key];
          if (!items) continue;

          for (const item of items) {
            const bbox = item._boundingBox;
            if (!bbox) continue;

            if (
              item.rest_duration != null &&
              item.rest_start != null &&
              item.rest_end != null
            ) {
              const restStartX = timeToX(item.rest_start);
              const restEndX = timeToX(item.rest_end);
              const restWidth = restEndX - restStartX;

              const restY = bbox.y + bbox.height - 4;
              const restHeight = 4;

              const isHoveringRestBar =
                mouseCanvasX >= restStartX &&
                mouseCanvasX <= restStartX + restWidth &&
                mouseCanvasY >= restY &&
                mouseCanvasY <= restY + restHeight;

              if (isHoveringRestBar) {
                dispatch("flightHover", {
                  type: "rest",
                  rest_info: {
                    rest_start: item.rest_start,
                    rest_end: item.rest_end,
                    rest_duration: item.rest_duration,
                    x: mouseCanvasX,
                    y: restY,
                  },
                });
                return;
              }
            }

            if (
              mouseCanvasX >= bbox.x &&
              mouseCanvasX <= bbox.x + bbox.width &&
              mouseCanvasY >= bbox.y &&
              mouseCanvasY <= bbox.y + bbox.height
            ) {
              newHoveredFlight = item;
              foundHover = true;
              break;
            }
          }

          if (foundHover) break;
        }
      }

      if (hoveredFlight?.data_id !== newHoveredFlight?.data_id) {
        hoveredFlight = newHoveredFlight;
        dispatch(
          "flightHover",
          newHoveredFlight ? { type: "flight", item: newHoveredFlight } : null
        );
        scheduleRedraw();
      } else if (!newHoveredFlight && hoveredFlight) {
        dispatch("flightHover", null);
        hoveredFlight = null;
        scheduleRedraw();
      }

      currentCursor =
        clientX < TIMELINE_PADDING_LEFT
          ? "grab"
          : foundHover
            ? "pointer"
            : "grab";
    }
  }

  function handleMouseUp() {
    if (isGlobalRowDragging) {
      currentDropTargetRowKey = null;
      scheduleRedraw();
      return;
    }

    if (isItemDragging && draggedFlight) {
      isItemDragging = false;
      // If 'M' key was released during dragging, reset cursor and state
      if (!isMKeyDown) currentCursor = "grab";

      currentDropTargetRowKey = null;
      isDroppingOverCollision = false;
      dispatch("flightHover", null);

      const newDepartureTime = xToTime(draggedFlight._boundingBox!.x);
      const flightDuration =
        draggedFlight.arrival_time - draggedFlight.departure_time;
      const newArrivalTime = newDepartureTime + flightDuration;

      const groupKeys = Object.keys(groupedFlights).sort((a, b) => {
        const [personIdA, typeA] = a.split("_");
        const [personIdB, typeB] = b.split("_");

        if (personIdA !== personIdB) {
          return personIdA.localeCompare(personIdB);
        }
        if (typeA === RowType.Actual && typeB === RowType.Publish) return -1;
        if (typeA === RowType.Publish && typeB === RowType.Actual) return 1;
        return 0;
      });
      const newRowIndex = Math.floor(
        (mouseCanvasY - HEADER_HEIGHT) / (ROW_HEIGHT + ROW_GAP)
      );

      let newGroupKey: string | undefined;
      if (newRowIndex >= 0 && newRowIndex < groupKeys.length) {
        newGroupKey = groupKeys[newRowIndex];
      }

      const isInvalidDropArea = mouseCanvasX < TIMELINE_PADDING_LEFT;

      const [draggedPersonId, draggedType] = (
        draggedFlight._originalGroupKey || ""
      ).split("_");
      let targetType = newGroupKey ? newGroupKey.split("_")[1] : null;

      const droppedOnWrongTypeRow = newGroupKey && draggedType !== targetType;

      if (
        isInvalidDropArea ||
        droppedOnWrongTypeRow ||
        (newGroupKey &&
          checkCollision(
            draggedFlight,
            newGroupKey,
            newDepartureTime,
            newArrivalTime
          ))
      ) {
        console.warn(
          "Flight overlaps with another flight in the target group or dropped in an invalid area/row type! Drop cancelled."
        );
        draggedFlight = null;
        clickedFlight = null;
        scheduleRedraw();
      } else if (newGroupKey) {
        const [targetPersonId, targetRowType] = newGroupKey.split("_");

        dispatch("flightMoved", {
          flight: draggedFlight,
          newDepartureTime,
          newArrivalTime,
          newPersonId: targetPersonId,
          newRowType: targetRowType as RowType,
          sourceTimelineId: timelineId,
          originalGroupKey: (draggedFlight as any)._originalGroupKey,
        });

        draggedFlight = null;
        clickedFlight = null;
        scheduleRedraw();
      } else {
        console.log("Flight dropped outside target row.");
        draggedFlight = null;
        clickedFlight = null;
        scheduleRedraw();
      }

      window.removeEventListener("mousemove", handleMouseMove);
      window.removeEventListener("mouseup", handleMouseUp);
    } else if (isDragging) {
      isDragging = false;
      currentCursor = "grab";

      window.removeEventListener("mousemove", handleMouseMove);
      window.removeEventListener("mouseup", handleMouseUp);
    }
  }

  function handleClick(event: MouseEvent) {
    // Prevent click actions if an item is being dragged, modifier keys are down (for zoom/pan), or global row dragging is active
    if (isItemDragging || isModifierKeyDown || isGlobalRowDragging) return;

    const rect = canvas.getBoundingClientRect();
    const rawX = event.clientX - rect.left;
    const rawY = event.clientY - rect.top;

    // Handle sidebar header dropdown click
    if (rawY < HEADER_HEIGHT && rawX < TIMELINE_PADDING_LEFT) {
      dispatch("sidebarHeaderDropdownClick", {
        pageX: event.pageX,
        pageY: event.pageY,
        timelineId: timelineId,
      });
      return;
    }

    const clickX = rawX;
    const clickY = rawY + scrollY;

    // If click is in the sidebar, deselect any flight
    if (clickX < TIMELINE_PADDING_LEFT) {
      clickedFlight = null;
      scheduleRedraw();
      return;
    }

    let foundClicked = false;
    const groupKeys = Object.keys(groupedFlights).sort((a, b) => {
      const [personIdA, typeA] = a.split("_");
      const [personIdB, typeB] = b.split("_");

      if (personIdA !== personIdB) {
        return personIdA.localeCompare(personIdB);
      }
      if (typeA === RowType.Actual && typeB === RowType.Publish) return -1;
      if (typeA === RowType.Publish && typeB === RowType.Actual) return 1;
      return 0;
    });
    const numRows = groupKeys.length;

    // Optimize search for clicked flight only within visible/relevant rows
    const firstVisibleRowIndex = Math.max(
      0,
      Math.floor((clickY - HEADER_HEIGHT) / (ROW_HEIGHT + ROW_GAP)) - 10
    );
    const lastVisibleRowIndex = Math.min(
      numRows - 1,
      Math.ceil(
        (clickY + containerHeight - HEADER_HEIGHT) / (ROW_HEIGHT + ROW_GAP)
      ) + 10
    );

    for (let i = firstVisibleRowIndex; i <= lastVisibleRowIndex; i++) {
      const key = groupKeys[i];
      const items = groupedFlights[key];
      if (!items) continue;

      for (const item of items) {
        const bbox = item._boundingBox;
        if (
          bbox &&
          clickX >= bbox.x &&
          clickX <= bbox.x + bbox.width &&
          clickY >= bbox.y &&
          clickY <= bbox.y + bbox.height
        ) {
          clickedFlight = item;
          dispatch("flightSelected", item);
          foundClicked = true;
          scheduleRedraw();
          break;
        }
      }
    }

    if (!foundClicked && clickedFlight !== null) {
      clickedFlight = null;
      scheduleRedraw();
    }
  }

  function handleMouseLeaveCanvas() {
    if (!isItemDragging) {
      hoveredFlight = null;
      dispatch("flightHover", null);
      scheduleRedraw();
    }
  }

  let resizeObserver: ResizeObserver;

  function handleKeyDown(event: KeyboardEvent) {
    if (event.ctrlKey || event.shiftKey) {
      isModifierKeyDown = true;
    }
    if (event.key === "m" || event.key === "M") {
      isMKeyDown = true;
      // If we're not currently dragging something, set cursor to grab
      if (!isDragging && !isItemDragging && !isGlobalRowDragging) {
        currentCursor = "grab";
      }
    }
    if (event.key === "Escape" && isItemDragging) {
      draggedFlight = null;
      isItemDragging = false;
      clickedFlight = null;
      currentDropTargetRowKey = null;
      isDroppingOverCollision = false;
      currentCursor = "grab";
      dispatch("flightHover", null);
      scheduleRedraw();
      event.preventDefault();
    }
  }

  function handleKeyUp(event: KeyboardEvent) {
    if (!event.ctrlKey && !event.shiftKey) {
      isModifierKeyDown = false;
    }
    if (event.key === "m" || event.key === "M") {
      isMKeyDown = false;
      // If no dragging is active, reset cursor to grab
      if (!isDragging && !isItemDragging && !isGlobalRowDragging) {
        currentCursor = "grab";
      }
    }
  }

  onMount(() => {
    ctx = canvas.getContext("2d")!;

    containerWidth = canvasContainerRef.clientWidth;
    containerHeight = canvasContainerRef.clientHeight;

    // ZoomController'ı kaydet
    currentZoomController = registerZoomController(timelineId, {
      zoomIn,
      zoomOut,
      setZoom,
      resetZoom,
      fitToRange,
    });

    // İlk yüklemede zaman aralığını ve piksel/saat değerini belirle
    // (Bu mantık artık reactive $flights bloğunda daha merkezi olarak yönetiliyor)
    // Ancak onMount sırasında initial çizimi tetiklemek için:
    scheduleRedraw();

    resizeObserver = new ResizeObserver((entries) => {
      for (let entry of entries) {
        if (entry.target === canvasContainerRef) {
          const newWidth = entry.contentRect.width;
          const newHeight = entry.contentRect.height;
          if (containerWidth !== newWidth || containerHeight !== newHeight) {
            containerWidth = newWidth;
            containerHeight = newHeight;
            updateVisibleRowRange();
            // Genişlik değiştiğinde piksel/saat oranını koruyarak bitiş zamanını ayarla
            timelineEnd =
              timelineStart +
              ((containerWidth - TIMELINE_PADDING_LEFT) / pixelsPerHour) *
                60 *
                60 *
                1000;
            scheduleRedraw();
          }
        }
      }
    });

    resizeObserver.observe(canvasContainerRef);

    nowLineInterval = setInterval(() => {
      scheduleRedraw();
    }, 1000);

    canvasContainerRef.addEventListener("scroll", handleScroll);
    canvas.addEventListener("wheel", handleWheel, { passive: false });
    canvas.addEventListener("mousedown", handleMouseDown);
    canvas.addEventListener("mousemove", handleMouseMove);
    canvas.addEventListener("mouseup", handleMouseUp);
    canvas.addEventListener("click", handleClick);
    canvas.addEventListener("mouseleave", handleMouseLeaveCanvas);

    window.addEventListener("keydown", handleKeyDown);
    window.addEventListener("keyup", handleKeyUp);
  });

  onDestroy(() => {
    if (resizeObserver) {
      resizeObserver.disconnect();
    }

    if (canvas) {
      canvas.removeEventListener("wheel", handleWheel);
      canvas.removeEventListener("mousedown", handleMouseDown);
      canvas.removeEventListener("mousemove", handleMouseMove);
      canvas.removeEventListener("mouseup", handleMouseUp);
      canvas.removeEventListener("click", handleClick);
      canvas.removeEventListener("mouseleave", handleMouseLeaveCanvas);
    }
    canvasContainerRef.removeEventListener("scroll", handleScroll);
    window.removeEventListener("keydown", handleKeyDown);
    window.removeEventListener("keyup", handleKeyUp);

    unregisterZoomController(timelineId);
    unsubscribeHighlight();

    clearInterval(nowLineInterval);
    unsubscribeDragStore();
  });

  function handleScroll() {
    scrollY = canvasContainerRef.scrollTop;
    updateVisibleRowRange();
    scheduleRedraw();
  }

  function darkenColor(hex: string, percent: number) {
    let f = parseInt(hex.slice(1), 16),
      t = percent < 0 ? 0 : 255,
      p = percent < 0 ? percent * -1 : percent,
      R = f >> 16,
      G = (f >> 8) & 0x00ff,
      B = f & 0x0000ff;
    return (
      "#" +
      (
        0x1000000 +
        (Math.round((t - R) * p) + R) * 0x10000 +
        (Math.round((t - G) * p) + G) * 0x100 +
        (Math.round((t - B) * p) + B)
      )
        .toString(16)
        .slice(1)
    );
  }
</script>

<div class="canvas-container" bind:this={canvasContainerRef}>
  <canvas
    bind:this={canvas}
    style="cursor: {currentCursor};"
    on:wheel={handleWheel}
    on:mousedown={handleMouseDown}
    on:mousemove={handleMouseMove}
    on:mouseup={handleMouseUp}
    on:click={handleClick}
    on:mouseleave={handleMouseLeaveCanvas}
    data-context-menu-trigger="true"
    on:contextmenu|preventDefault|stopPropagation={(e) => {
      console.log(
        "RosterTimeline: Canvas üzerinde on:contextmenu yakalandı. Özel menü mantığı başlatılıyor."
      );

      const rect = canvas.getBoundingClientRect();
      const clientX = e.clientX - rect.left;
      const clientY = e.clientY - rect.top;

      const adjustedMouseYForContent = clientY + scrollY;

      const groupKeys = Object.keys(groupedFlights).sort((a, b) => {
        const [personIdA, typeA] = a.split("_");
        const [personIdB, typeB] = b.split("_");
        if (personIdA !== personIdB) return personIdA.localeCompare(personIdB);
        if (typeA === RowType.Actual && typeB === RowType.Publish) return -1;
        if (typeA === RowType.Publish && typeB === RowType.Actual) return 1;
        return 0;
      });

      const rowClickedIndex = Math.floor(
        (adjustedMouseYForContent - HEADER_HEIGHT) / (ROW_HEIGHT + ROW_GAP)
      );

      let clickedRowKey: string | null = null;
      if (rowClickedIndex >= 0 && rowClickedIndex < groupKeys.length) {
        clickedRowKey = groupKeys[rowClickedIndex];
      }

      // 1) TIMELINE HEADER
      if (
        clientY >= 0 &&
        clientY < HEADER_HEIGHT &&
        clientX >= TIMELINE_PADDING_LEFT
      ) {
        console.log(
          "RosterTimeline: TIMELINE_HEADER context menu dispatch ediliyor."
        );
        dispatch("timelineHeaderContextMenu", {
          pageX: e.pageX,
          pageY: e.pageY,
          timelineId,
        });
        return;
      }

      // 2) SOL SİDEBAR
      if (clientX < TIMELINE_PADDING_LEFT) {
        if (clickedRowKey) {
          const [personId, rowType] = clickedRowKey.split("_");
          console.log(
            "RosterTimeline: TIMELINE_ROW_SIDEBAR context menu dispatch ediliyor."
          );
          dispatch("rowSidebarContextMenu", {
            personId,
            rowType: rowType as RowType,
            pageX: e.pageX,
            pageY: e.pageY,
            timelineId,
          });
          return;
        }
      }

      // 3) İÇERİK ALANI (ITEM veya BOŞ SATIR)
      if (clientX >= TIMELINE_PADDING_LEFT) {
        let foundFlightForContextMenu = false;

        const searchStartRowIndex = Math.max(0, rowClickedIndex - 5);
        const searchEndRowIndex = Math.min(
          groupKeys.length - 1,
          rowClickedIndex + 5
        );

        for (let i = searchStartRowIndex; i <= searchEndRowIndex; i++) {
          const key = groupKeys[i];
          const items = groupedFlights[key];
          if (!items) continue;

          for (const item of items) {
            const bbox = item._boundingBox;
            if (
              bbox &&
              clientX >= bbox.x &&
              clientX <= bbox.x + bbox.width &&
              adjustedMouseYForContent >= bbox.y &&
              adjustedMouseYForContent <= bbox.y + bbox.height
            ) {
              console.log(
                "RosterTimeline: TIMELINE_ITEM context menu dispatch ediliyor."
              );
              dispatch("flightItemContextMenu", {
                flight: item,
                pageX: e.pageX,
                pageY: e.pageY,
                timelineId,
              });
              // ✅ ITEM bulundu: asla fall-through olmasın
              return;
            }
          }
        }

        // (Ek güvenlik) Eğer bir şekilde işaretlendi ise yine de dön
        if (foundFlightForContextMenu) return;

        // 4) BOŞ SATIR
        if (clickedRowKey) {
          const [personId, rowType] = clickedRowKey.split("_");
          console.log(
            "RosterTimeline: TIMELINE_ROW (Zaman Çizelgesi Boş Alanı) context menu dispatch ediliyor."
          );
          dispatch("rowContextMenu", {
            personId,
            rowType: rowType as RowType,
            pageX: e.pageX,
            pageY: e.pageY,
            timelineId,
          });
          return;
        }
      }

      // 5) GENEL BOŞ ALAN
      console.log(
        "RosterTimeline: TIMELINE_EMPTY context menu dispatch ediliyor."
      );
      dispatch("timelineEmptyContextMenu", {
        pageX: e.pageX,
        pageY: e.pageY,
        timelineId,
        coords: { x: clientX, y: clientY + scrollY },
      });
      return;
    }}
  ></canvas>

  <div class="spacer" style="height: {totalContentHeight}px;"></div>

  {#each visibleRows as rowKey, index (rowKey)}
    <FlightRowOverlay
      {rowKey}
      top={HEADER_HEIGHT +
        (firstVisibleRowIndex + index) * (ROW_HEIGHT + ROW_GAP) -
        scrollY}
      height={ROW_HEIGHT}
      isActiveDropTarget={currentDropTargetRowKey === rowKey}
      isCollision={isDroppingOverCollision &&
        currentDropTargetRowKey === rowKey}
      on:rowSidebarContextMenu={(e) => {
        const [personId, rowType] = e.detail.rowKey.split("_");

        dispatch("rowSidebarContextMenu", {
          personId,
          rowType: rowType as RowType,
          pageX: e.detail.pageX,
          pageY: e.detail.pageY,
          timelineId,
        });
        console.log(timelineId);
      }}
    />
  {/each}

  {#if isLoading}
    <div class="timeline-loader">
      <div class="progress-bar-container">
        <div class="progress-bar" style="width: {loadingProgress}%;"></div>
      </div>
      <p>Veriler yükleniyor...</p>
    </div>
  {/if}
</div>

<style>
  .canvas-container {
    width: 100%;
    height: 100%;
    overflow-y: auto;
    overflow-x: hidden;
    position: relative;
    background-color: #f5f5f5;
    flex-grow: 1;
    min-height: 0;
  }
  .spacer {
    position: absolute;
    top: 0;
    left: 0;
    width: 1px;
    pointer-events: none;
    height: 100%;
    min-height: 0;
  }

  canvas {
    position: sticky;
    top: 0;
    left: 0;
    width: 100%;
    height: 100%;
    z-index: 1;
    pointer-events: auto;
  }

  .timeline-loader {
    position: absolute;
    top: 0;
    left: 0;
    width: 100%;
    height: 100%;
    background-color: rgba(255, 255, 255, 0.85);
    display: flex;
    flex-direction: column;
    justify-content: center;
    align-items: center;
    z-index: 100;
    pointer-events: none;
  }

  .timeline-loader p {
    margin-top: 15px;
    font-size: 1.1em;
    color: #555;
  }

  .progress-bar-container {
    width: 80%;
    max-width: 300px;
    height: 8px;
    background-color: #e0e0e0;
    border-radius: 4px;
    overflow: hidden;
  }

  .progress-bar {
    height: 100%;
    background-color: #007bff;
    width: 0%;
    transition: width 0.1s ease-out;
  }
</style>
