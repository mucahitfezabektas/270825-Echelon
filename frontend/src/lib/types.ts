// src/lib/types.ts

// ==============================
// Backend Entegrasyon Tipleri
// ==============================

// ✅ Actual tipi (backend'deki activity ile birebir uyumlu)
// Bu interface, backend'den gelen ham veriyi temsil eder.
export interface Actual {
    data_id: string;
    activity_code: string;
    name: string;
    surname: string;
    base_filo: string;
    class: string;
    departure_port: string;
    arrival_port: string;
    departure_time: string; // API'den ISO 8601 string olarak gelir
    arrival_time: string;   // API'den ISO 8601 string olarak gelir
    person_id: string;
    plane_cms_type: string;
    plane_tail_name: string;
    trip_id: string;
    group_code: string;
    flight_position: string;
    flight_no: string;
    agreement_type: string;
    checkin_date: string;
    duty_start: string;
    duty_end: string;
    period_month: string;
    flight_id: string;

    // Backend'den gelen opsiyonel dinlenme süresi alanları (string olarak)
    rest_start?: string;
    rest_end?: string;
    rest_duration?: number; // Backend'den number olarak gelebilir
}

// ==============================
// Timeline-Specific Tipler
// ==============================

// Timeline satırının türünü belirtmek için
export enum RowType {
    Actual = 'actual',
    Publish = 'publish',
}

export interface BoundingBox {
    x: number;
    y: number;
    width: number;
    height: number;
}

// Uçuş verisi (FlightItem) - Frontend'in kullandığı ve zaman damgalarını number'a dönüştürdüğü tip
// Bu interface, timeline'da kullanılacak işlenmiş veriyi temsil eder.
export interface FlightItem {
    data_id: string;
    activity_code: string;
    name: string;
    surname: string;
    base_filo: string;
    class: string;
    departure_port: string;
    arrival_port: string;
    departure_time: number; // Frontend'de Unix timestamp'e (milliseconds) dönüştürülmüş
    arrival_time: number;   // Frontend'de Unix timestamp'e (milliseconds) dönüştürülmüş
    person_id: string;
    plane_cms_type: string;
    plane_tail_name: string;
    trip_id: string;
    group_code: string;
    flight_position: string;
    flight_no: string;
    flight_id: string;

    // Yeni eklenen alan: Bu öğenin Actual mı Publish mi olduğunu belirtir
    type: 'actual' | 'publish';

    // Frontend runtime özellikleri (canvas içi pozisyonlama ve sürükleme)
    _boundingBox?: BoundingBox;
    _originalGroupKey?: string; // Sürükleme için orijinal grubu tutar

    // Dinlenme Süresi Hesaplaması Sabitleri (number olmalı)
    rest_start?: number; // Frontend'de Unix timestamp'e (milliseconds) dönüştürülmüş
    rest_end?: number;   // Frontend'de Unix timestamp'e (milliseconds) dönüştürülmüş
    rest_duration?: number; // Saat cinsinden
}

export type FlightDataType = 'roster' | 'trip' | 'rotation';
export type StatusMessageType = string;

// --- Düzeltildi: HoverInfo birleşim tipi güncellendi ve 'null' eklendi ---
export type HoverInfo =
    | { type: "flight"; item: FlightItem; rest_info?: never; }
    | { // Yeni eklenen "trip" tipi
        type: "trip";
        tripId: string;
        x: number;
        y: number;
        departure_port: string;
        arrival_port: string;
        departure_time: number;
        arrival_time: number;
        rotation_minutes: number;
        ftl_violations: string[];
        item?: never;
        rest_info?: never;
    }
    | {
        type: "rest";
        rest_info: {
            rest_start: number; // Düzeltildi: string yerine number
            rest_end: number;   // Düzeltildi: string yerine number
            rest_duration: number;
            x: number;
            y: number;
        };
        item?: never;
    }
    | { type: "none"; item?: never; rest_info?: never }
    | null; // Burası eklendi: null değeri de kabul edilsin diye

// This interface is needed for the dragStore and is used in Timeline.svelte
export interface DragState {
    isDraggingRow: boolean;
    draggedGroupKey: string | null;
    draggedRowFlights: FlightItem[];
    sourceTimelineId: string | null;
    draggedRowVisualY: number;
    dragRowOffsetY: number;
}


// Sağ tık menüsü event'i için payload
export interface RowContextMenuEvent {
    personId: string;
    rowType: RowType; // Hangi satır tipine tıklandığını belirtir
    pageX: number;
    pageY: number;
    timelineId: string;
}


// ==============================
// Pencere (Window) ile İlgili Tipler
// ==============================

export interface WindowBounds {
    x: number;
    y: number;
    width: number;
    height: number;
}

export interface WindowDragEvent {
    x: number;
    y: number;
}

export interface WindowResizeEvent extends WindowBounds { }

export interface WindowEvents {
    activate: CustomEvent<void>;
    minimize: CustomEvent<void>;
    maximize: CustomEvent<WindowBounds>;
    restore: CustomEvent<WindowBounds>;
    close: CustomEvent<void>;
    dragStart: CustomEvent<WindowDragEvent>;
    drag: CustomEvent<WindowDragEvent>;
    dragEnd: CustomEvent<WindowDragEvent>;
    resizeStart: CustomEvent<WindowResizeEvent>;
    resize: CustomEvent<WindowResizeEvent>;
    resizeEnd: CustomEvent<WindowResizeEvent>;
}

// ==============================
// UI/UX ile İlgili Tipler
// ==============================

export type DockPosition = "left" | "right" | "top" | "bottom" | null;

// ==============================
// Filter Query ile İlgili Tipler
// ==============================

export type FilterOperator = '=' | '!=' | '>' | '<' | 'LIKE';

export type FilterRow = {
    field: string;
    operator: FilterOperator;
    value: string;
};

export type SavedFilter = {
    id: string;
    name: string;
    rows: FilterRow[];
    logic: 'AND' | 'OR'; // Go backend'in beklediği gibi, opsiyonel değil
};

// ✅ Backend’ten gelen ham yanıt (Actual[])
export interface ActualsByFlightIDResponseBackend {
    total_persons_found: number;
    person_ids: string[];
    result: {
        [personId: string]: Actual[]; // <-- HAM veri: Actual[]
    };
}

// ✅ Frontend’de dönüştürülmüş yanıt (FlightItem[])
export interface ActualsByFlightIDResponseFrontend {
    total_persons_found: number;
    person_ids: string[];
    result: {
        [personId: string]: FlightItem[]; // <-- DÖNÜŞMÜŞ veri: FlightItem[]
    };
}