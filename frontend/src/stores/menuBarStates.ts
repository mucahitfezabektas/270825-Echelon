// src/stores/menuBarStates.ts

import { writable } from 'svelte/store';
import { getWindowConfig, type WindowConfig } from '@/config/windowConfig';

// Menü öğeleri için basit bir arayüz tanımlayabiliriz.
// Bu, WindowConfig'in yanı sıra 'submenu' gibi özel tipleri de içerebilir.
export interface MenuItem {
    id: string;
    label: string;
    icon?: string;
    type?: "item" | "submenu";
    children?: MenuItem[];
}

// Varsayılan menü yapısını oluşturun.
// Bu, başlangıçta `CustomizeContent.svelte` içinde tanımlanan yapıya benzer olmalı.
// getWindowConfig kullanarak WindowConfig tabanlı öğeler ekleyebiliriz.
const mapConfigToMenuItem = (id: string): MenuItem | undefined => {
    const config = getWindowConfig(id);
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

const createDefaultMenuBarStructure = (): MenuItem[] => {
    return [
        {
            id: "File",
            label: "Dosya",
            type: "submenu",
            children: [
                mapConfigToMenuItem("ExitWindow"),
                mapConfigToMenuItem("RosterWindow"),
                mapConfigToMenuItem("AdminPanelWindow"),
                mapConfigToMenuItem("CustomizeWindow"),
                mapConfigToMenuItem("ColorsPanel"),
            ].filter(Boolean) as MenuItem[], // undefined olanları filtrele ve MenuItem[] olarak tipini belirle
        },
        {
            id: "Help",
            label: "Yardım",
            type: "submenu",
            children: [
                mapConfigToMenuItem("OnlineHelp"),
                mapConfigToMenuItem("AboutWindow"),
                mapConfigToMenuItem("Logs"),
            ].filter(Boolean) as MenuItem[],
        },
        // Gelecekte buraya başka varsayılan menü öğeleri ekleyebilirsiniz
    ].filter(Boolean) as MenuItem[];
};

// Menü aksiyonları store'unu varsayılan yapı ile başlatın
export const menuBarActions = writable<MenuItem[]>(createDefaultMenuBarStructure());

// Menü aksiyonlarını varsayılan durumuna sıfırlamak için fonksiyon
export function resetMenuBarActions() {
    menuBarActions.set(createDefaultMenuBarStructure());
}

// Menü aksiyonlarını dinamik olarak güncellemek için fonksiyon
// Genellikle tüm menü yapısını set etmek için kullanılır
export function updateMenuBarActions(newStructure: MenuItem[]) {
    menuBarActions.set(newStructure);
}