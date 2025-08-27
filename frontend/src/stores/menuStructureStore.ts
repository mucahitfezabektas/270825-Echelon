import { writable, get } from 'svelte/store';

// Type definitions for menu items
export interface MenuItem {
    id: string;
    label: string;
    type: 'item' | 'submenu' | 'separator';
    icon?: string;
    children?: MenuItem[];
    actionId?: string; // If it's a direct action item (usually same as id for items)
    expanded?: boolean; // New: For tree view expand/collapse state
}

// mapConfigToMenuItem fonksiyonu, allWindowConfigs'e erişmek için dışarıdan bir getter fonksiyonu alacak şekilde güncellendi.
// Bu fonksiyon artık yalnızca app.ts veya customize.ts gibi yerlerde çağrılır.
// menuStructureStore'un varsayılan yapısını tanımlarken doğrudan kullanılır.
// NOT: Bu mapConfigToMenuItem artık sadece CustomizeWindow.svelte içinde kullanılıyor
// ve orada allWindowConfigs'e eriştiği için burada parametreye ihtiyaç duymuyor.
// Ancak buradaki tanımı dışarıdan kullanılabilirlik için koruyabiliriz.
export const mapConfigToMenuItem = (id: string, allConfigs: any[]): MenuItem | undefined => {
    const config = allConfigs.find(c => c.id === id);
    if (config) {
        return {
            id: config.id,
            label: config.title,
            icon: config.icon,
            type: "item",
        };
    }
    return undefined;
};

// isMenuItemAlreadyInStructure fonksiyonu güncel hali
export function isMenuItemAlreadyInStructure(items: MenuItem[], itemId: string): boolean {
    for (const item of items) {
        if (item.id === itemId) {
            return true;
        }
        if (item.type === "submenu" && Array.isArray(item.children)) {
            if (isMenuItemAlreadyInStructure(item.children, itemId)) {
                return true;
            }
        }
    }
    return false;
}

// Varsayılan menü yapısını doğrudan ID'ler ve basit bilgilerle tanımlayalım.
// Expanded durumları da buraya eklendi.
const defaultMenuStructure: MenuItem[] = [
    {
        id: "File",
        label: "File",
        type: "submenu",
        expanded: true, // Varsayılan olarak açık
        children: [
            { id: "OpenProjectWindow", label: "Open Project", type: "item", icon: "fas fa-folder-open" },
            { id: "SaveChangesWindow", label: "Save Changes", type: "item", icon: "fas fa-save" },
            { id: "RosterWindow", label: "Roster", type: "item", icon: "fas fa-plus" },
            { id: "AdminPanelWindow", label: "Admin Panel", type: "item", icon: "fas fa-user-cog" },
            { id: "CustomizeWindow", label: "Customize", type: "item", icon: "fas fa-sliders-h" },
            { id: "ApplicationSettings", label: "App Settings", type: "item", icon: "fas fa-cog" },
            { id: "ExitWindow", label: "Exit", type: "item", icon: "fas fa-times-circle" },
        ],
    },
    {
        id: "View",
        label: "View",
        type: "submenu",
        expanded: true, // Varsayılan olarak açık
        children: [
            { id: "CrewExplorer", label: "Crew Explorer", type: "item", icon: "fas fa-users" },
            { id: "DataExplorer", label: "Data Explorer", type: "item", icon: "fas fa-database" },
            { id: "FlightScheduleView", label: "Flight Schedule", type: "item", icon: "fas fa-calendar-check" },
        ],
    },
    {
        id: "Tools",
        label: "Tools",
        type: "submenu",
        expanded: true, // Varsayılan olarak açık
        children: [
            { id: "Toolbox", label: "Toolbox", type: "item", icon: "fas fa-toolbox" },
            { id: "ConfigurationEditor", label: "Config Editor", type: "item", icon: "fas fa-sliders-h" },
            { id: "ColorsPanel", label: "Colors Panel", type: "item", icon: "fas fa-palette" },
            { id: "UserPreferences", label: "User Preferences", type: "item", icon: "fas fa-user" },
        ],
    },
    {
        id: "Help",
        label: "Help",
        type: "submenu",
        expanded: true, // Varsayılan olarak açık
        children: [
            { id: "OnlineHelp", label: "Online Help", type: "item", icon: "fas fa-question-circle" },
            { id: "HelpManual", label: "Help Manual", type: "item", icon: "fas fa-book" },
            { id: "Logs", label: "Logs", type: "item", icon: "fas fa-clipboard-list" },
            { id: "AboutWindow", label: "About", type: "item", icon: "fas fa-info-circle" },
        ],
    },
];

export const menuStructure = writable<MenuItem[]>(defaultMenuStructure);

export function resetMenuStructureToDefaults() {
    menuStructure.set(JSON.parse(JSON.stringify(defaultMenuStructure)));
}