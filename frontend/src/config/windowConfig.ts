// src/config/windowConfig.ts

import type { SvelteComponent } from "svelte";

// Import all your window components
import EmptyWindowContent from "@/components/EmptyWindowContent.svelte";
import RosterWindow from "@/windows/OperationWindow.svelte";
import OpenProjectWindow from "@/windows/OpenProjectWindow.svelte";
import SaveChangesWindow from "@/windows/SaveChangesWindow.svelte";
import ExitWindow from "@/windows/ExitWindow.svelte";
import AdminPanelWindow from "@/windows/AdminPanelWindow.svelte";
import CustomizeWindow from "@/windows/CustomizeWindow.svelte";
import AboutWindow from "@/windows/AboutWindow.svelte";
import LogWindow from "@/windows/LogWindow.svelte";
import ColorSampler from "@/windows/ColorSamplerWindow.svelte";
import UserOptions from "@/windows/UserOptionsWindow.svelte";
import UnifiedTableEditor from "@/windows/UnifiedTableEditor.svelte";
import FilterQueryWindow from "@/windows/FilterQueryWindow.svelte";

// Define the structure for a window configuration entry
export interface WindowConfig {
    id: string; // Unique identifier for the window/action
    title: string; // Default title for the window
    component?: any | null; // Made optional and can be null (for actions without a dedicated window)
    type: 'item'; // Explicitly define 'type' for WindowConfig items as 'item'
    initialWidth: number;
    initialHeight: number;
    minimizable?: boolean;
    maximizable?: boolean;
    icon: string; // Icon class for toolbars/menus
}

// Define the map of all available windows and their configurations
export const allWindowConfigs: WindowConfig[] = [
    {
        id: "OperationWindow",
        title: "Operation Window",
        component: RosterWindow,
        initialWidth: 900,
        initialHeight: 700,
        minimizable: true,
        maximizable: true,
        icon: "fas fa-plus",
        type: "item", // Added type
    },
    {
        id: "LogWindow",
        title: "Application Log",
        icon: "fas fa-clipboard-list",
        component: LogWindow,
        initialWidth: 600,
        initialHeight: 400,
        minimizable: true,
        maximizable: true,
        type: "item", // Added type
    },
    {
        id: "OpenProjectWindow",
        title: "Open Project",
        component: OpenProjectWindow,
        initialWidth: 600,
        initialHeight: 400,
        minimizable: false,
        maximizable: false,
        icon: "fas fa-folder-open",
        type: "item", // Added type
    },
    {
        id: "SaveChangesWindow",
        title: "Save Changes",
        component: SaveChangesWindow,
        initialWidth: 500,
        initialHeight: 350,
        minimizable: false,
        maximizable: false,
        icon: "fas fa-save",
        type: "item", // Added type
    },
    {
        id: "ExitWindow",
        title: "Exit",
        component: ExitWindow,
        initialWidth: 450,
        initialHeight: 250,
        minimizable: false,
        maximizable: false,
        icon: "fas fa-times-circle",
        type: "item", // Added type
    },
    {
        id: "AdminPanelWindow",
        title: "Admin Panel",
        component: AdminPanelWindow,
        initialWidth: 1000,
        initialHeight: 750,
        minimizable: true,
        maximizable: true,
        icon: "fas fa-user-cog",
        type: "item", // Added type
    },
    {
        id: "CustomizeWindow",
        title: "Customize",
        component: CustomizeWindow,
        initialWidth: 700,
        initialHeight: 500,
        minimizable: true,
        maximizable: true,
        icon: "fas fa-sliders-h",
        type: "item", // Added type
    },
    {
        id: "AboutWindow",
        title: "About",
        component: AboutWindow,
        initialWidth: 600,
        initialHeight: 450,
        minimizable: false,
        maximizable: false,
        icon: "fas fa-info-circle",
        type: "item", // Added type
    },
    {
        id: "ApplicationSettings",
        title: "Uygulama Ayarları",
        component: EmptyWindowContent,
        initialWidth: 600,
        initialHeight: 400,
        minimizable: false,
        maximizable: false,
        icon: "fas fa-cog",
        type: "item", // Added type
    },
    {
        id: "ColorSamplerWindow",
        title: "Color Sampler",
        component: ColorSampler,
        initialWidth: 600,
        initialHeight: 800,
        minimizable: true,
        maximizable: true,
        icon: "fas fa-palette",
        type: "item", // Added type
    },

    {
        id: "UserOptionsWindow",
        title: "User Options",
        component: UserOptions,
        initialWidth: 600,
        initialHeight: 800,
        minimizable: true,
        maximizable: false,
        icon: "fas fa-users-cog",
        type: "item", // Added type
    },

    {
        id: "UnifiedTableEditorWindow",
        title: "Unified Table Editor",
        component: UnifiedTableEditor,
        initialWidth: 600,
        initialHeight: 800,
        minimizable: true,
        maximizable: true,
        icon: "fas fa-database",
        type: "item", // Added type
    },

    {
        id: "FilterQueryWindow",
        title: "Filter Query Window",
        component: FilterQueryWindow,
        initialWidth: 650,
        initialHeight: 800,
        minimizable: true,
        maximizable: true,
        icon: "fas fa-filter",
        type: "item", // Added type
    },

    // --- Yer Tutucu (Placeholder) Pencereler ---
    { id: "FlightScheduler", title: "Uçuş Planlayıcı", component: EmptyWindowContent, initialWidth: 850, initialHeight: 650, minimizable: true, maximizable: true, icon: "fas fa-plane", type: "item" },
    { id: "CrewManagement", title: "Mürettebat Yönetimi", component: EmptyWindowContent, initialWidth: 950, initialHeight: 700, minimizable: true, maximizable: true, icon: "fas fa-users-cog", type: "item" },
    { id: "AircraftMaintenance", title: "Uçak Bakımı", component: EmptyWindowContent, initialWidth: 900, initialHeight: 600, minimizable: true, maximizable: true, icon: "fas fa-tools", type: "item" },
    { id: "Reports", title: "Raporlar", component: EmptyWindowContent, initialWidth: 750, initialHeight: 550, minimizable: true, maximizable: true, icon: "fas fa-chart-line", type: "item" },
    { id: "Notifications", title: "Bildirimler", component: EmptyWindowContent, initialWidth: 500, initialHeight: 300, minimizable: true, maximizable: false, icon: "fas fa-bell", type: "item" },
    { id: "HelpManual", title: "Yardım Kılavuzu", component: EmptyWindowContent, initialWidth: 800, initialHeight: 600, minimizable: true, maximizable: true, icon: "fas fa-book", type: "item" },
    { id: "UserPreferences", title: "Kullanıcı Tercihleri", component: EmptyWindowContent, initialWidth: 550, initialHeight: 450, minimizable: false, maximizable: false, icon: "fas fa-user", type: "item" },
    { id: "Alerts", title: "Uyarılar", component: EmptyWindowContent, initialWidth: 500, initialHeight: 300, minimizable: true, maximizable: false, icon: "fas fa-exclamation-triangle", type: "item" },
    { id: "BackgroundProcesses", title: "Arkaplan İşlemleri", component: EmptyWindowContent, initialWidth: 700, initialHeight: 400, minimizable: true, maximizable: false, icon: "fas fa-cogs", type: "item" },
    { id: "ChangeTimeZone", title: "Saat Dilimi Değiştir", component: EmptyWindowContent, initialWidth: 400, initialHeight: 250, minimizable: false, maximizable: false, icon: "fas fa-clock", type: "item" },
    { id: "ConfigurationEditor", title: "Konfigürasyon Düzenleyici", component: EmptyWindowContent, initialWidth: 700, initialHeight: 550, minimizable: true, maximizable: true, icon: "fas fa-sliders-h", type: "item" },
    { id: "Course", title: "Kurs", component: EmptyWindowContent, initialWidth: 600, initialHeight: 500, minimizable: true, maximizable: false, icon: "fas fa-book", type: "item" },
    { id: "CourseGrade", title: "Kurs Notu", component: EmptyWindowContent, initialWidth: 600, initialHeight: 400, minimizable: true, maximizable: false, icon: "fas fa-graduation-cap", type: "item" },
    { id: "CrewExplorer", title: "Mürettebat Gezgini", component: EmptyWindowContent, initialWidth: 800, initialHeight: 600, minimizable: true, maximizable: true, icon: "fas fa-users", type: "item" },
    { id: "DataExplorer", title: "Veri Gezgini", component: EmptyWindowContent, initialWidth: 900, initialHeight: 700, minimizable: true, maximizable: true, icon: "fas fa-database", type: "item" },
    { id: "DeskDefinition", title: "Masa Tanımı", component: EmptyWindowContent, initialWidth: 700, initialHeight: 500, minimizable: true, maximizable: false, icon: "fas fa-clipboard-list", type: "item" },
    { id: "DisplayFRMSAttributes", title: "FRMS Özellikleri Göster", component: EmptyWindowContent, initialWidth: 750, initialHeight: 500, minimizable: true, maximizable: true, icon: "fas fa-chart-bar", type: "item" },
    { id: "FillEmptyDays", title: "Boş Günleri Doldur", component: EmptyWindowContent, initialWidth: 400, initialHeight: 250, minimizable: false, maximizable: false, icon: "fas fa-calendar-alt", type: "item" },
    { id: "FlightInTrouble", title: "Sorunlu Uçuş", component: EmptyWindowContent, initialWidth: 600, initialHeight: 400, minimizable: true, maximizable: false, icon: "fas fa-plane-slash", type: "item" },
    { id: "FlightRequirementsSettings", title: "Uçuş Gereksinim Ayarları", component: EmptyWindowContent, initialWidth: 700, initialHeight: 500, minimizable: true, maximizable: false, icon: "fas fa-plane-departure", type: "item" },
    { id: "FlightScheduleView", title: "Uçuş Takvim Görünümü", component: EmptyWindowContent, initialWidth: 900, initialHeight: 650, minimizable: true, maximizable: true, icon: "fas fa-calendar-check", type: "item" },
    { id: "GlobalDelete", title: "Global Silme", component: EmptyWindowContent, initialWidth: 500, initialHeight: 300, minimizable: false, maximizable: false, icon: "fas fa-trash-alt", type: "item" },
    { id: "HotelManager", title: "Otel Yöneticisi", component: EmptyWindowContent, initialWidth: 700, initialHeight: 550, minimizable: true, maximizable: true, icon: "fas fa-hotel", type: "item" },
    { id: "Logs", title: "Loglar", component: EmptyWindowContent, initialWidth: 800, initialHeight: 600, minimizable: true, maximizable: true, icon: "fas fa-clipboard-list", type: "item" },
    { id: "MasterCrewList", title: "Ana Mürettebat Listesi", component: EmptyWindowContent, initialWidth: 850, initialHeight: 650, minimizable: true, maximizable: true, icon: "fas fa-address-book", type: "item" },
    { id: "MessagingBroadcast", title: "Mesaj Yayınlama", component: EmptyWindowContent, initialWidth: 600, initialHeight: 400, minimizable: false, maximizable: false, icon: "fas fa-broadcast-tower", type: "item" },
    { id: "Toolbox", title: "Araç Kutusu", component: EmptyWindowContent, initialWidth: 400, initialHeight: 300, minimizable: true, maximizable: false, icon: "fas fa-toolbox", type: "item" },
    { id: "NonFlyingAssignments", title: "Uçuş Dışı Görevler", component: EmptyWindowContent, initialWidth: 700, initialHeight: 500, minimizable: true, maximizable: false, icon: "fas fa-ban", type: "item" },
    { id: "OperationalSimulator", title: "Operasyonel Simülatör", component: EmptyWindowContent, initialWidth: 900, initialHeight: 700, minimizable: true, maximizable: true, icon: "fas fa-plane", type: "item" },
    { id: "OnlineHelp", title: "Çevrimiçi Yardım", component: EmptyWindowContent, initialWidth: 800, initialHeight: 600, minimizable: true, maximizable: true, icon: "fas fa-question-circle", type: "item" },
];

// Helper to easily get a config by ID
export function getWindowConfig(id: string): WindowConfig | undefined {
    return allWindowConfigs.find((config) => config.id === id);
}

// Helper to get component map for App.svelte
export const getComponentMap = (): Record<string, new (...args: any[]) => SvelteComponent> => {
    const map: Record<string, new (...args: any[]) => SvelteComponent> = {};
    for (const config of allWindowConfigs) {
        // Ensure component is not null or undefined before assigning
        if (config.component) {
            map[config.id] = config.component;
        }
    }
    return map;
};