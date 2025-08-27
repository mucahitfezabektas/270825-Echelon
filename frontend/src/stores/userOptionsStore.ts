// src/stores/userOptionsStore.ts
import { writable, derived } from "svelte/store";

/* ------------------------------------------------------------------ */
/*  Types                                                             */
/* ------------------------------------------------------------------ */
export interface GeneralPreferences {
    windowArrangement: "overlapping" | "sidebyside";
    ganttScroll: "scrollZoom" | "zoomScroll";
    showGC: boolean;
}

export interface LanguagePreferences {
    locale: string;
    showDefaultLabels: boolean;
    overrideCustomerLabels: boolean;
}

export interface OpsViewPreferences {
    startTemplate: "lastUsed" | "startWith";
    templateName: string;
}

export interface ReportingPreferences {
    urlParam: string;
    reportPath: string;
    rosterReportPath: string;
    username: string;
    password: string;
}

export interface StudioPreferences {
    startTemplate: "lastUsed" | "startWith";
    templateName: string;
    rememberZoom: boolean;
    zoomFactor: number;
    defaultRosterLevel: string;
    opsRosterLevel: string;
    saturationLevel: number;
    rosterChangeTypes: string[];
    showPatternTooltip: boolean;
    showNominationTooltip: boolean;
    showBaselineTooltip: boolean;
}

export interface TimezonePreferences {
    /** `"utc"` = her yerde UTC -  `"airport"` = seçilen havalimanının yerel saati */
    type: "utc" | "airport";
    /** Havalimanı kodu (IATA) – yalnızca `type:"airport"` ise anlamlıdır */
    selectedAirport?: string;
}

export interface Preferences {
    general: GeneralPreferences;
    language: LanguagePreferences;
    opsview: OpsViewPreferences;
    reporting: ReportingPreferences;
    studio: StudioPreferences;
    timezone: TimezonePreferences;
}

/* ------------------------------------------------------------------ */
/*  Varsayılan Ayarlar                                                */
/* ------------------------------------------------------------------ */
const defaultPreferences: Preferences = {
    general: {
        windowArrangement: "overlapping",
        ganttScroll: "scrollZoom",
        showGC: true,
    },
    language: {
        locale: "default",
        showDefaultLabels: true,
        overrideCustomerLabels: true,
    },
    opsview: {
        startTemplate: "startWith",
        templateName: "ops tiny",
    },
    reporting: {
        urlParam: "http://10.80.61.20:8080/jasperserver-pro/",
        reportPath: "/Crew_Manager/Reports",
        rosterReportPath: "/Crew_Manager/Context",
        username: "jasperadmin",
        password: "jasperadmin",
    },
    studio: {
        startTemplate: "startWith",
        templateName: "roster tiny",
        rememberZoom: false,
        zoomFactor: 5.1,
        defaultRosterLevel: "PAIRING",
        opsRosterLevel: "FLIGHT",
        saturationLevel: 0.2,
        rosterChangeTypes: [
            "assignments",
            "scheduleChanges",
            "nonFlyingAssignments",
            "fillEmptyDays",
        ],
        showPatternTooltip: false,
        showNominationTooltip: false,
        showBaselineTooltip: true,
    },
    /* 🔵 DEĞİŞEN KISIM → varsayılan artık GMT+3 (IST) */
    timezone: {
        type: "airport",
        selectedAirport: "IST",
    },
};

/* ------------------------------------------------------------------ */
/*  Airport → UTC-Offset (dakika) tablosu                             */
/* ------------------------------------------------------------------ */
export const AIRPORT_OFFSETS_MIN: Record<string, number> = {
    IST: 180, // GMT+3
    JFK: -240, // GMT-4
    LHR: 0,
    DXB: 240,
};

export function getTimezoneOffsetMin(pref: TimezonePreferences): number {
    if (pref.type === "utc") return 0;

    const airport = (pref.selectedAirport ?? "").toUpperCase(); // 🛠️ normalize et
    return AIRPORT_OFFSETS_MIN[airport || "IST"] ?? 180;

}



/* ------------------------------------------------------------------ */
/*  LocalStorage → Store                                              */
/* ------------------------------------------------------------------ */
const STORAGE_KEY = "user_preferences";

function loadPreferences(): Preferences {
    if (typeof window === "undefined") return defaultPreferences;

    try {
        const raw = localStorage.getItem(STORAGE_KEY);
        if (!raw) return defaultPreferences;
        const parsed = JSON.parse(raw);

        /* derin birleştirme – yeni alanlar eklense bile sorun yok */
        return {
            ...defaultPreferences,
            ...parsed,
            general: { ...defaultPreferences.general, ...parsed.general },
            language: { ...defaultPreferences.language, ...parsed.language },
            opsview: { ...defaultPreferences.opsview, ...parsed.opsview },
            reporting: { ...defaultPreferences.reporting, ...parsed.reporting },
            studio: {
                ...defaultPreferences.studio,
                ...parsed.studio,
                rosterChangeTypes: Array.isArray(parsed.studio?.rosterChangeTypes)
                    ? parsed.studio.rosterChangeTypes
                    : defaultPreferences.studio.rosterChangeTypes,
            },
            timezone: { ...defaultPreferences.timezone, ...parsed.timezone },
        };
    } catch (err) {
        console.error("⚠️ Prefs parse error, defaults used:", err);
        return defaultPreferences;
    }
}

/* Ana store */
export const preferencesStore = writable<Preferences>(loadPreferences());

/* Otomatik persist */
if (typeof window !== "undefined") {
    preferencesStore.subscribe((prefs) => {
        localStorage.setItem(STORAGE_KEY, JSON.stringify(prefs));
    });
}

/* ------------------------------------------------------------------ */
/*  Türetilmiş store: aktif ofset                                      */
/* ------------------------------------------------------------------ */
export const timezoneOffsetMinStore = derived(
    preferencesStore,
    ($prefs) => getTimezoneOffsetMin($prefs.timezone)
);
