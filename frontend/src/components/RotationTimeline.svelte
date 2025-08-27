<script lang="ts">
  import { onMount, onDestroy, createEventDispatcher } from "svelte";
  import type { FlightItem, Actual } from "@/lib/types"; // Actual type is no longer directly used for props here
  import { dragStore, resetDragState } from "@/stores/dragStore";
  import FlightRowOverlay from "@/components/FlightRowOverlay.svelte";
  import { getActivityColor } from "@/stores/colorStore";
  import { FTL_VIOLATION_COLORS } from "@/lib/config/ruleConfig";

  import { timezoneOffsetMinStore } from "@/stores/userOptionsStore";
  import { highlightStore } from "@/stores/highlightStore";

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
  import { get } from "svelte/store";
  import {
    registerZoomController,
    unregisterZoomController,
  } from "@/stores/zoomControlStore";

  export let flights: FlightItem[] = [];
  export let locale: string = "tr-TR";
  export let timelineId: string; // Unique ID from Timeline.svelte

  export let isLoading: boolean = false;
  export let loadingProgress: number = 0; // 0-100 arası

  type SafeDateFormat = Intl.DateTimeFormatOptions & {
    hour?: "2-digit" | "numeric";
    minute?: "2-digit" | "numeric";
    day?: "2-digit" | "numeric";
    weekday?: "short" | "narrow" | "long";
    hourCycle?: "h11" | "h12" | "h23" | "h24";
  };

  const dispatch = createEventDispatcher();
  const rangeStart = new Date(2025, 4, 1).getTime(); // 4 = May (0-based)
  const rangeEnd = new Date(2025, 4, 31, 23, 59, 59, 999).getTime();

  let highlightedKeys: Set<string> = new Set();
  let canvas: HTMLCanvasElement;
  let canvasContainerRef: HTMLDivElement;
  let ctx: CanvasRenderingContext2D;

  let groupedFlights: Record<string, FlightItem[]> = {};
  let pixelsPerHour = 5;
  let spacerHeight = 0;

  // --- Drag-and-Drop Current State Variables ---
  let draggedFlight: FlightItem | null = null;
  let dragStartX: number;
  let dragStartY: number;
  let dragOffsetX: number;
  let dragOffsetY: number;
  let isItemDragging: boolean = false;
  let currentDropTargetRowKey: string | null = null;
  let isDroppingOverCollision: boolean = false;

  let start: number;
  let end: number;
  let isInitialLoad = true;

  let isDragging = false; // Timeline pan
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

  let isModifierKeyDown = false; // Ctrl or Shift key pressed?

  let totalDurations: Record<string, number> = {};
  let lastGroupKeys: string[] = [];
  // rowFTLViolations remains empty as FTL calculation is currently backend-side and not passed to FlightItem
  let rowFTLViolations: Record<string, string[]> = {};

  // Track global drag state
  let isGlobalRowDragging: boolean = false;
  let globalDraggedGroupKey: string | null = null;
  let globalSourceTimelineId: string | null = null;
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

  $: visibleRows = Object.keys(groupedFlights)
    .sort()
    .slice(firstVisibleRowIndex, lastVisibleRowIndex + 1);

  // Recalculate total content height and adjust scrollY
  $: {
    const numRows = Object.keys(groupedFlights).length;
    const contentH =
      HEADER_HEIGHT + numRows * (ROW_HEIGHT + ROW_GAP) + ROW_GAP * 2;

    spacerHeight = Math.max(0, contentH - containerHeight);
    totalContentHeight = contentH; // Spacer handles the scrollable height.

    scrollY = Math.min(scrollY, totalContentHeight - containerHeight);
    scrollY = Math.max(0, scrollY);
  }

  $: {
    groupedFlights = groupFlightsByPersonId(flights);
    rowFTLViolations = {}; // Set to empty as FTL violations are not directly attached to FlightItem currently.

    if (isInitialLoad || flights.length === 0) {
      const allTimestamps = flights.flatMap((f) => [
        f.departure_time, // FlightItem timestamps are numbers
        f.arrival_time, // FlightItem timestamps are numbers
      ]);

      if (allTimestamps.length > 0) {
        start = Math.min(...allTimestamps) - 2 * 3600 * 1000;
        end = Math.max(...allTimestamps) + 2 * 3600 * 1000;
      } else {
        const now = Date.now();
        start = now - 6 * 60 * 60 * 1000;
        end = now + 6 * 60 * 60 * 1000;
      }
      isInitialLoad = false;
    }
    scheduleRedraw();
  }

  $: if (containerWidth > 0 && ctx && containerHeight > 0) {
    scheduleRedraw();
  }

  function zoomIn() {
    const newVal = Math.min(pixelsPerHour * 1.2, MAX_ZOOM);
    pixelsPerHour = newVal;

    const centerTime = (start + end) / 2;
    start =
      centerTime -
      ((containerWidth - TIMELINE_PADDING_LEFT) / 2 / newVal) * 60 * 60 * 1000;
    end =
      centerTime +
      ((containerWidth - TIMELINE_PADDING_LEFT) / 2 / newVal) * 60 * 60 * 1000;
  }

  function zoomOut() {
    const newVal = Math.max(pixelsPerHour / 1.2, MIN_ZOOM);
    pixelsPerHour = newVal;

    const centerTime = (start + end) / 2;
    start =
      centerTime -
      ((containerWidth - TIMELINE_PADDING_LEFT) / 2 / newVal) * 60 * 60 * 1000;
    end =
      centerTime +
      ((containerWidth - TIMELINE_PADDING_LEFT) / 2 / newVal) * 60 * 60 * 1000;
  }

  function getCurrentInteractionMode(
    event: MouseEvent | WheelEvent
  ): InteractionMode {
    if (event.ctrlKey) return "zoom";
    if (event.shiftKey) return "pan";
    return "none";
  }

  /**
   * Helper function to debounce any function.
   */
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

  /**
   * Converts a timestamp to an x-coordinate on the canvas.
   */
  function timeToX(timestamp: number): number {
    return Math.round(
      TIMELINE_PADDING_LEFT +
        ((timestamp - start) / (1000 * 60 * 60)) * pixelsPerHour
    );
  }

  /**
   * Converts an x-coordinate on the canvas to a timestamp.
   */
  function xToTime(x: number): number {
    return (
      start + ((x - TIMELINE_PADDING_LEFT) / pixelsPerHour) * 60 * 60 * 1000
    );
  }

  /**
   * Groups flights by person_id.
   * Now works with FlightItem[] directly.
   */
  function groupFlightsByPersonId(
    items: FlightItem[]
  ): Record<string, FlightItem[]> {
    const map: Record<string, FlightItem[]> = {};
    for (const item of items) {
      if (!item.person_id) continue;
      const groupKey = item.person_id; // group by person_id
      if (!map[groupKey]) {
        map[groupKey] = [];
      }
      map[groupKey].push(item);
    }
    return map;
  }

  // Helper function to truncate text to a given width
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

  /**
   * Schedules a redraw of the timeline using requestAnimationFrame to optimize performance.
   */
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

  /**
   * Checks if a moving flight would collide with existing flights in a target group.
   * Works with FlightItem.
   */
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

  /**
   * Draws only the vertical grid lines on the canvas.
   */
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

  /**
   * Draws only the time labels in the fixed header area.
   */
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

      const offset = get(timezoneOffsetMinStore); // current offset
      const date = new Date(t - offset * 60 * 1000);
      const year = date.getFullYear();
      const month = date.getMonth();
      const weekday = date.getDay();

      // === Vertical line ===
      ctx.beginPath();
      ctx.moveTo(x + 0.5, 0);
      ctx.lineTo(x + 0.5, headerHeight);
      ctx.strokeStyle = "#c6c6c6";
      ctx.lineWidth = 1;
      ctx.stroke();

      // === YEAR ===
      if (showYearLabels && year !== lastYear) {
        ctx.font = "bold 12px sans-serif";
        ctx.fillStyle = "#111";
        ctx.fillText(String(year), x + LABEL_PAD, 0);
        lastYear = year;
      }

      // === MONTH ===
      if (showMonthLabels && month !== lastMonth) {
        ctx.font = "11px sans-serif";
        ctx.fillStyle = "#333";
        const monthLabel = date.toLocaleString(locale, { month: "short" });
        ctx.fillText(monthLabel, x + LABEL_PAD, yLayerH);
        lastMonth = month;
      }

      // === DAY / HOUR ===
      ctx.font = "10px sans-serif";
      let label = "";

      if (minorStepMs >= DAY_MS) {
        // === DAY VIEW: "5 Fri" ===
        const dayNum = date.getDate(); // 5
        const weekdayLabel = date.toLocaleString(locale, { weekday: "short" }); // "Fri"
        label = `${dayNum} ${weekdayLabel}`;
        ctx.fillStyle = weekday === 0 ? "darkred" : "#555";
      } else {
        // === HOUR VIEW: "03.00" ===
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

  /**
   * Draws the "now" line indicating the current time.
   */
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
    const groupKeys = Object.keys(groupedFlights).sort();
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

  // ====== Constants & Helpers ======
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

  // ====== Main Function ======
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

    // ---- Row Loop ----
    const maxRow = Math.min(lastIdx + 1, groupKeys.length);
    for (let i = firstIdx; i < maxRow; i++) {
      const key = groupKeys[i]; // This will be crew_member_id
      if (!key) continue;

      const yRow = HEADER_HEIGHT + i * (ROW_HEIGHT + ROW_GAP);
      const zebraFill = i % 2 === 0 ? "#f4f4f4" : "#fcfcfc";

      let rowFill = zebraFill;
      let hasFTLViolations =
        rowFTLViolations[key] && rowFTLViolations[key].length > 0; // Will be false with current setup

      if (highlightedKeys.has(key)) {
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
        // FTL violation highlighting for the row (currently will not trigger)
        rowFill = FTL_VIOLATION_COLORS.ROW_BACKGROUND_DARK;
      }

      // Entire row background
      ctx.fillStyle = rowFill;
      ctx.fillRect(0, yRow, TIMELINE_PADDING_LEFT, ROW_HEIGHT);

      // --- Prepare cell data ---
      const items = groupedFlights[key];
      const firstItem = items?.[0]; // Use firstItem for crew info
      const durationM = totalDurations[key] ?? 0;
      const durH = (durationM / 60).toFixed(2);

      const cells: (string | null)[] = [
        firstItem?.name ?? "",
        firstItem?.class ?? "-",
        `Uçuş: ${durH}s`, // Total Flight Hours
        firstItem?.surname ?? "-",
        `${key}`, // Display Person ID
        `Görev: ${durH}s`, // Total Duty Hours (same as flight duration for now, can be fetched from backend)
      ];

      // --- Draw cells ---
      ctx.font = SIDEBAR.fontReg;
      for (let idx = 0; idx < cells.length; idx++) {
        const col = idx % SIDEBAR.cols;
        const row = Math.floor(idx / SIDEBAR.cols);

        const baseX = SIDEBAR.padX + col * cellW;
        const baseY = yRow + SIDEBAR.padY + row * cellH;

        // Alert background (if FTL violation) - currently will not trigger
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
            rowFill, // Pass rowFill as baseFill
            alertFill
          );

        // Text
        ctx.font = idx === 0 ? SIDEBAR.fontBold : SIDEBAR.fontReg;
        const txt = getTruncatedText(ctx, cells[idx]!, cellW - 4);
        ctx.fillStyle = "#333";
        ctx.fillText(txt, baseX + 2, baseY + cellH / 2);
      }

      // FTL Violation Icon/Indicator (top right corner) - currently will not trigger
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

  /**
   * Main function to draw the entire timeline.
   */
  function drawTimeline() {
    if (!ctx || !canvas) return;

    calculateTotalDurations();

    const groupKeys = Object.keys(groupedFlights).sort();

    canvas.width = containerWidth;
    canvas.height = containerHeight;
    ctx.clearRect(0, 0, canvas.width, containerHeight);

    /* ==== LEFT & TOP FIXED AREAS ==== */
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
      const baseY = HEADER_HEIGHT + i * (ROW_HEIGHT + ROW_GAP);
      let hasFTLViolations =
        rowFTLViolations[key] && rowFTLViolations[key].length > 0; // Currently will not trigger

      if (highlightedKeys.has(key)) {
        ctx.fillStyle = "rgba(255, 215, 0, 0.25)";
      } else if (
        isGlobalRowDragging &&
        globalSourceTimelineId !== timelineId &&
        currentDropTargetRowKey === key
      ) {
        ctx.fillStyle = DRAG_OVER_COLOR;
      } else if (isItemDragging && currentDropTargetRowKey === key) {
        ctx.fillStyle = isDroppingOverCollision
          ? COLLISION_COLOR
          : DRAG_OVER_COLOR;
      } else if (hasFTLViolations) {
        ctx.fillStyle = FTL_VIOLATION_COLORS.ROW_BACKGROUND_LIGHT; // Currently will not trigger
      } else {
        ctx.fillStyle = i % 2 === 0 ? "#f8f8f8" : "#ffffff";
      }

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
    drawGridLinesOnly(start, end, maxDrawHeight);

    ctx.restore();

    for (let i = firstVisibleRowIndex; i <= lastVisibleRowIndex; i++) {
      const key = groupKeys[i];
      const baseY = HEADER_HEIGHT + i * (ROW_HEIGHT + ROW_GAP);
      const rowHeight = ROW_HEIGHT;
      const itemsForThisGroup = groupedFlights[key];
      if (!itemsForThisGroup) continue;

      // 1️⃣ Calculate BoundingBox (before drawing)
      for (const item of itemsForThisGroup) {
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

      // 3️⃣ DRAW FLIGHT BOXES
      for (const item of itemsForThisGroup) {
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

        // Determine style directly
        const baseColor = getActivityColor(item.activity_code);
        ctx.fillStyle =
          hoveredFlight?.data_id === item.data_id
            ? "#ffc107"
            : clickedFlight?.data_id === item.data_id
              ? "#007bff"
              : baseColor;

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

        // 🔹 REST PERIOD
        // rest_start and rest_end are now numbers (timestamps) in FlightItem
        if (
          item.rest_start != null &&
          item.rest_end != null &&
          item.rest_duration != null
        ) {
          const restStartX = timeToX(item.rest_start); // Directly use as number
          const restEndX = timeToX(item.rest_end); // Directly use as number
          const restWidth = restEndX - restStartX;

          const restBoxY = item._boundingBox.y + item._boundingBox.height - 4;
          ctx.fillStyle = "#9ec9ff"; // Rest period color
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
          const drawX = Math.max(item._boundingBox.x, TIMELINE_PADDING_LEFT);
          const actualDrawingWidth =
            item._boundingBox.width -
            (TIMELINE_PADDING_LEFT - item._boundingBox.x > 0
              ? TIMELINE_PADDING_LEFT - item._boundingBox.x
              : 0);

          const fixedCenterX =
            (item._boundingBox.x +
              item._boundingBox.x +
              item._boundingBox.width) /
            2;
          const fixedCenterY =
            item._boundingBox.y + item._boundingBox.height / 2 + 1;

          const textWidth = ctx.measureText(textToDraw).width;

          if (textWidth < actualDrawingWidth - 4) {
            ctx.fillStyle = "black";
            ctx.font = "10px sans-serif";
            ctx.textAlign = "center";
            ctx.textBaseline = "middle";
            ctx.fillText(textToDraw, fixedCenterX, fixedCenterY);
          }

          const portsFit = actualDrawingWidth > 100;
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
              drawX + actualDrawingWidth - 2,
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
              drawX + actualDrawingWidth - 2,
              item._boundingBox.y + item._boundingBox.height - 2
            );
          }
        }

        ctx.restore();
      }
    }

    ctx.restore(); // Restore from scroll translation

    // These elements should remain fixed relative to the canvas viewport
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
    drawTimeLabels(start, end, HEADER_HEIGHT);
    drawNowLine(containerHeight);

    // DRAG BOX
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
    const groupKeys = Object.keys(groupedFlights).sort();

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
            const depTime = act.departure_time; // Already a number
            const arrTime = act.arrival_time; // Already a number
            return acc + (arrTime - depTime) / 60_000;
          }
          return acc;
        }, 0) ?? 0;
    }

    console.log("🧮 totalDurations updated:", totalDurations);
  }

  function scrollToFirstHighlighted() {
    const groupKeys = Object.keys(groupedFlights).sort();
    const firstIdx = groupKeys.findIndex((k) => highlightedKeys.has(k));
    if (firstIdx === -1) return;

    const targetY = firstIdx * (ROW_HEIGHT + ROW_GAP);
    const container = canvas.parentElement;
    if (container) {
      container.scrollTo({ top: targetY, behavior: "smooth" });
    }
  }

  function isInDropdownHeader(x: number, y: number): boolean {
    console.log("Checking isInDropdownHeader. x:", x, "y:", y);
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

    const rowCount = Object.keys(groupedFlights).length;
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
    ctx.fillText("ROTATION", leftWidth / 2, height / 2);

    ctx.font = "11px sans-serif";
    ctx.fillStyle = "#666";
    ctx.fillText(`${rowCount} row`, leftWidth + rightWidth / 2, halfHeight / 2);
    ctx.fillText(
      `${itemCount.toLocaleString(locale)} item`, // using locale for item count
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

  /**
   * Handles mouse wheel events for zooming and panning.
   */
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

        start =
          centerTime -
          ((mouseX - TIMELINE_PADDING_LEFT) / pixelsPerHour) * 60 * 60 * 1000;
        end =
          start +
          ((containerWidth - TIMELINE_PADDING_LEFT) / pixelsPerHour) *
            60 *
            60 *
            1000;
      } else if (mode === "pan") {
        const panAmount = event.deltaY * 5;
        const deltaMs = (panAmount / pixelsPerHour) * 60 * 60 * 1000;

        start += deltaMs;
        end += deltaMs;
      }

      debouncedScheduleRedraw();
    }
  }

  /**
   * Handles mouse down events for initiating dragging (panning).
   * Works with FlightItem.
   */
  function handleMouseDown(event: MouseEvent) {
    if (isGlobalRowDragging) return;

    const rect = canvas.getBoundingClientRect();
    const clientX = event.clientX - rect.left;
    const clientY = event.clientY - rect.top;

    mouseCanvasX = clientX;
    mouseCanvasY = clientY;

    const adjustedMouseYForContent = clientY + scrollY;

    // INITIATE ROW DRAGGING
    if (mouseCanvasX < TIMELINE_PADDING_LEFT) {
      const rowClickedIndex = Math.floor(
        (adjustedMouseYForContent - HEADER_HEIGHT) / (ROW_HEIGHT + ROW_GAP)
      );
      const groupKeys = Object.keys(groupedFlights).sort();

      if (rowClickedIndex >= 0 && rowClickedIndex < groupKeys.length) {
        const draggedGroupKey = groupKeys[rowClickedIndex];
        const activitiesInGroup = groupedFlights[draggedGroupKey]; // FlightItem[]

        dispatch("rowDragStart", {
          groupKey: draggedGroupKey,
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
    }

    // CURRENT ITEM DRAGGING LOGIC
    if (mouseCanvasX > TIMELINE_PADDING_LEFT) {
      let foundFlightToDrag = false;
      const groupKeys = Object.keys(groupedFlights).sort();
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
              mouseCanvasX >= bbox.x &&
              mouseCanvasX <= bbox.x + bbox.width &&
              adjustedMouseYForContent >= bbox.y &&
              adjustedMouseYForContent <= bbox.y + bbox.height
            ) {
              draggedFlight = item;
              isItemDragging = true;
              dragStartX = mouseCanvasX;
              dragStartY = mouseCanvasY;
              dragOffsetX = mouseCanvasX - bbox.x;
              dragOffsetY = adjustedMouseYForContent - bbox.y;

              (draggedFlight as any)._originalGroupKey = key; // Store original groupKey

              currentCursor = "grabbing";
              clickedFlight = item;
              scheduleRedraw();
              foundFlightToDrag = true;
              break;
            }
          }
        }
        if (foundFlightToDrag) break;
      }

      if (foundFlightToDrag) {
        window.addEventListener("mousemove", handleMouseMove);
        window.addEventListener("mouseup", handleMouseUp);
      } else {
        const mode = getCurrentInteractionMode(event);
        if (mode === "pan" || event.button === 0) {
          isDragging = true;
          lastMouseX = event.clientX;
          currentCursor = "grab";
          window.addEventListener("mousemove", handleMouseMove);
          window.addEventListener("mouseup", handleMouseUp);
        }
      }
    } else {
      currentCursor = "grab";
    }
  }

  /**
   * Handles mouse move events for panning and hover detection.
   * Works with FlightItem.
   */
  function handleMouseMove(event: MouseEvent) {
    if (!canvas || !canvasContainerRef) return;

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
        const groupKeys = Object.keys(groupedFlights).sort();
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

    if (isItemDragging && draggedFlight && draggedFlight._boundingBox) {
      draggedFlight._boundingBox.x = clientX - dragOffsetX;

      currentCursor = "grabbing";

      const groupKeys = Object.keys(groupedFlights).sort();
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

        newIsDroppingOverCollision = checkCollision(
          draggedFlight,
          newDropTargetRowKey,
          newDepartureTime,
          newArrivalTime
        );
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
      const deltaX = event.clientX - lastMouseX;
      lastMouseX = event.clientX;

      const deltaMs = (deltaX / pixelsPerHour) * 60 * 60 * 1000;

      start -= deltaMs;
      end =
        start +
        ((containerWidth - TIMELINE_PADDING_LEFT) / pixelsPerHour) *
          60 *
          60 *
          1000;

      debouncedScheduleRedraw();
    } else {
      let foundHover = false;
      let newHoveredFlight: FlightItem | null = null;

      if (mouseCanvasX >= TIMELINE_PADDING_LEFT) {
        const groupKeys = Object.keys(groupedFlights).sort();
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

            // REST BAR HOVER CHECK
            if (
              item.rest_duration != null &&
              item.rest_start != null &&
              item.rest_end != null
            ) {
              const restStartX = timeToX(item.rest_start); // Already number
              const restEndX = timeToX(item.rest_end); // Already number
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
                    rest_start: item.rest_start, // Keep as number for hover info
                    rest_end: item.rest_end, // Keep as number for hover info
                    rest_duration: item.rest_duration,
                    x: mouseCanvasX,
                    y: restY,
                  },
                });
                return;
              }
            }

            // FLIGHT BOX HOVER CHECK
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

      // HOVER DISPATCH
      if (hoveredFlight?.data_id !== newHoveredFlight?.data_id) {
        hoveredFlight = newHoveredFlight;
        dispatch(
          "flightHover",
          newHoveredFlight ? { type: "flight", item: newHoveredFlight } : null
        );
        scheduleRedraw();
      } else if (!newHoveredFlight && hoveredFlight) {
        // If mouse leaves a flight
        dispatch("flightHover", null);
        hoveredFlight = null; // Reset hovered flight
        scheduleRedraw();
      }

      // CURSOR UPDATE
      currentCursor =
        clientX < TIMELINE_PADDING_LEFT
          ? "grab"
          : foundHover
            ? "pointer"
            : "grab";
    }
  }

  /**
   * Handles mouse up events to end dragging.
   * Works with FlightItem.
   */
  function handleMouseUp() {
    if (isGlobalRowDragging) {
      currentDropTargetRowKey = null;
      scheduleRedraw();
      return;
    }

    if (isItemDragging && draggedFlight) {
      isItemDragging = false;
      currentCursor = "grab";
      currentDropTargetRowKey = null;
      isDroppingOverCollision = false;
      dispatch("flightHover", null); // Clear hover on mouse up

      const newDepartureTime = xToTime(draggedFlight._boundingBox!.x);
      const flightDuration =
        draggedFlight.arrival_time - draggedFlight.departure_time;
      const newArrivalTime = newDepartureTime + flightDuration;

      const groupKeys = Object.keys(groupedFlights).sort();
      const newRowIndex = Math.floor(
        (mouseCanvasY - HEADER_HEIGHT) / (ROW_HEIGHT + ROW_GAP)
      );

      let newGroupKey: string | undefined;
      if (newRowIndex >= 0 && newRowIndex < groupKeys.length) {
        newGroupKey = groupKeys[newRowIndex];
      }

      const isInvalidDropArea = mouseCanvasX < TIMELINE_PADDING_LEFT;

      if (
        isInvalidDropArea ||
        (newGroupKey &&
          checkCollision(
            draggedFlight,
            newGroupKey,
            newDepartureTime,
            newArrivalTime
          ))
      ) {
        console.warn(
          "Flight overlaps with another flight in the target group or dropped in an invalid area! Drop cancelled."
        );
        draggedFlight = null;
        clickedFlight = null;
        scheduleRedraw();
      } else if (newGroupKey) {
        dispatch("flightMoved", {
          flight: draggedFlight,
          newDepartureTime,
          newArrivalTime,
          newGroupKey,
          sourceTimelineId: timelineId, // Indicate which timeline it was moved from
          originalGroupKey: (draggedFlight as any)._originalGroupKey,
        });

        draggedFlight = null;
        clickedFlight = null;
        scheduleRedraw(); // Still call here to clear the dragged ghost immediately
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
    if (isItemDragging || isModifierKeyDown || isGlobalRowDragging) return;

    const rect = canvas.getBoundingClientRect();
    const rawX = event.clientX - rect.left;
    const rawY = event.clientY - rect.top;

    // Check for click in sidebar header dropdown area
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

    if (clickX < TIMELINE_PADDING_LEFT) {
      clickedFlight = null;
      scheduleRedraw();
      return;
    }

    let foundClicked = false;
    const groupKeys = Object.keys(groupedFlights).sort();
    const numRows = groupKeys.length;

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
      if (items) {
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
            dispatch("flightSelected", item); // Dispatch selected flight
            foundClicked = true;
            scheduleRedraw();
            break;
          }
        }
      }
      if (foundClicked) break;
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
      if (!isDragging && !isItemDragging && !isGlobalRowDragging) {
        currentCursor = "grab";
      }
    }
  }

  onMount(() => {
    ctx = canvas.getContext("2d")!;

    containerWidth = canvasContainerRef.clientWidth;
    containerHeight = canvasContainerRef.clientHeight;

    const allTimestamps = flights.flatMap((f) => [
      f.departure_time,
      f.arrival_time,
    ]);

    if (allTimestamps.length > 0) {
      start = Math.min(...allTimestamps) - 2 * 3600 * 1000;
      end = Math.max(...allTimestamps) + 2 * 3600 * 1000;
    } else {
      const now = Date.now();
      start = now - 6 * 60 * 60 * 1000;
      end = now + 6 * 60 * 60 * 1000;
    }
    isInitialLoad = false;

    resizeObserver = new ResizeObserver((entries) => {
      for (let entry of entries) {
        if (entry.target === canvasContainerRef) {
          const newWidth = entry.contentRect.width;
          const newHeight = entry.contentRect.height;
          if (containerWidth !== newWidth || containerHeight !== newHeight) {
            containerWidth = newWidth;
            containerHeight = newHeight;
            updateVisibleRowRange();
            end =
              start +
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
    registerZoomController(timelineId, { zoomIn, zoomOut });
    scheduleRedraw();
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
</script>

<div class="canvas-container" bind:this={canvasContainerRef}>
  <canvas bind:this={canvas} style="cursor: {currentCursor};"></canvas>

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
    height: 100%; /* Will be overridden by inline style */
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
