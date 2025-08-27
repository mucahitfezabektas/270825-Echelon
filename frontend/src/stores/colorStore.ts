// src/stores/colorStore.ts
import { writable, get } from 'svelte/store';

// Kullanıcının sağladığı kurumsal renk paleti
const corporatePalette = [
    '#0B4F6C', '#1D6787', '#348AA7', '#50A8C9', '#6EBDE5', // Profesyonel Maviler
    '#2E6A5B', '#3E8E7D', '#5CA99B', '#7EC4B8', '#9FE0D8', // Derin Yeşiller
    '#5C5751', '#7A736B', '#9B948E', '#BDB7B1', '#DDD8D3', // Nötr Toprak Tonları
    '#7D3C6A', '#9C4E8A', '#BB6BAA', '#DB8DD0', '#F0B1E6', // Canlı Morlar/Mürdüm Tonları
    '#C07830', '#D8975D', '#ECC69F', '#FFF0E0', // Sıcak Nötrler / Toprak Sarıları
    '#B02A2A', '#D03E3E', '#F05252', '#FF6D6D'  // Vurgu Kırmızılar
];

// Uygulama genelinde kullanılacak aktivite kodlarının listeleri
// Bu listeler sadece referans amaçlıdır, renk atamaları aşağıda manuel yapılmıştır.
export const dutyTypes = [
    "AE3EK", "AE3ES", "AE3JS", "AE3JV", "AE3MS",
    "AE4EK", "AE4ES", "AE4JS", "AE4JV", "AE4MS",
    "AE5EK", "AE5ES", "AE5JS", "AE5JV", "AE5MS",
    "AE6EK", "AE6ES", "AE6JS", "AE6JV", "AEGES",
    "AG1EK", "AG1ES", "AG1JS", "AG1JV", "AG1MS", "AG1ZS",
    "AG2ES", "AG2JS", "AG2MS", "AG2ZS",
    "HE3EK", "HE3ES", "HE3JS", "HE3JV", "HE3MS",
    "HE4EK", "HE4ES", "HE4JS", "HE4JV", "HE4MS",
    "HE5EK", "HE5ES", "HE5JS", "HE5JV", "HE5MS",
    "HE6EK", "HE6ES", "HE6JS", "HE6JV", "HEGES",
    "HG1ES", "HG1JS", "HG1MS", "HG1ZS",
    "HG2ES", "HG2JS", "HG2MS", "HG2ZS",
];

export const dutyCodes = [
    "CFR", "DGR", "DLA", "DLB", "DTK", "DTT", "DUS", "DYG", "DYT",
    "EBG", "EBS", "EDC", "EEF", "EES", "EKK", "EMJ", "EMM", "EPD", "EPT",
    "ERC", "ERT", "ES3", "ES7", "ESA", "ESD", "ESK", "ESP", "ESX", "ETT",
    "EUM", "EY3", "EYL", "EYQ", "EYR",
    "FLT", "BUS",
    "GLB", "GOR",
    "HAK", "HSC", "HSP",
    "MCM", "MKU", "MND", "MUA",
    "OAF", "OB1", "OB2", "OB3", "OBF", "OCM", "OCR", "OIK", "OKD", "OKE", "OKP",
    "ONE", "OOM", "OPG", "OTM",
    "S33", "S77", "SND",
    "TLM", "TMM", "TSP",
    "UDC", "UHK", "USF",
    "XEE", "XEV", "XUT", "XXM",
];

export const otherActivities = [
    "DIFFERENCE", "LAYOVER", "MINIMUM_REST", "MODIFIED_REST", "REST_OPPORTUNITY", "SIT",
    "AWAJ", "TLM", "TSP", "MCM", "MKU", "HWAJ",
    "IAB", "IAC", "IAO", "IAV", "IBB", "IBC", "IBD", "IBE", "IBG", "IBI", "IBM", "IBU",
    "IBV", "IBY", "IDB", "IDO", "IEV", "IHE", "IHI", "IHK", "II9", "III", "IIO", "IKZ",
    "IMP", "IMZ", "IOD", "IOL", "IOZ", "IPR", "IRI", "ISE", "ISF", "ISP", "IUB", "IUE",
    "IUS", "IYO", "IYZ",
    "AL2ES", "AL2JS", "AL2MS", "AL2ZS",
    "AL3ES", "AL3JS", "AL3JV", "AL3MS", "AL3ZS",
    "AL4ES", "AL4JS", "AL4MS", "AL4ZS",
    "AL5ES", "AL5JS", "AL5MS", "AL5ZS",
    "AL6MS", "AL6ZS",
    "HL2ES", "HL2JS", "HL2MS", "HL2ZS",
    "HL3ES", "HL3JS", "HL3MS", "HL3ZS",
    "HL4ES", "HL4JS", "HL4MS", "HL4ZS",
    "HL5ES", "HL5JS", "HL5MS", "HL6MS",
];


export const generalCategories = [
    { name: "Alerted", code: "ALRT", defaultColor: "#F05252" }, // Vurgu Kırmızı
    { name: "General Duty", code: "GEN_DUTY", defaultColor: "#9B948E" }, // Nötr Gri
];

// Define the type for the activityColorStore's state
interface ActivityColorStoreState {
    colors: { [key: string]: string };
    defaultColor: string;
}

// Renk store'u
export const activityColorStore = writable<ActivityColorStoreState>({
    colors: {} as { [key: string]: string },
    defaultColor: '#7A736B', // Kurumsal nötr gri - Belirsiz aktiviteler veya hata durumları için
});

// Başlangıç renklerini yüklemek için
function initializeActivityColors() {
    const initialColors: { [key: string]: string } = {
        // --- ÖZEL İSTEKLERE GÖRE YENİ ATAMALAR (En Üstte Önceliklendirildi) ---
        // "I" ile başlayanlara gri tonları
        "IAB": "#ededed", // Açık Gri
        "IAC": "#e0e0e0", // Açık Gri
        "IAO": "#a0a0a0", // Orta Gri
        "IAV": "#666666", // Koyu Gri
        "IBB": "#a0a0a0", // Çok Koyu Gri
        "IBC": "#ededed",
        "IBD": "#e0e0e0",
        "IBE": "#a0a0a0",
        "IBG": "#666666",
        "IBI": "#333333",
        "IBM": "#ededed",
        "IBU": "#e0e0e0",
        "IBV": "#a0a0a0",
        "IBY": "#666666",
        "IDB": "#333333",
        "IDO": "#ededed",
        "IEV": "#e0e0e0",
        "IHE": "#a0a0a0",
        "IHI": "#666666",
        "IHK": "#333333",
        "II9": "#ededed",
        "III": "#e0e0e0",
        "IIO": "#a0a0a0",
        "IKZ": "#666666",
        "IMP": "#333333",
        "IMZ": "#ededed",
        "IOD": "#e0e0e0",
        "IOL": "#a0a0a0",
        "IOZ": "#666666",
        "IPR": "#333333",
        "IRI": "#ededed",
        "ISE": "#e0e0e0",
        "ISF": "#a0a0a0",
        "ISP": "#666666",
        "IUB": "#333333",
        "IUE": "#ededed",
        "IUS": "#e0e0e0",
        "IYO": "#a0a0a0",
        "IYZ": "#666666",

        // FLT ve CFR'a yeşil tonları
        "FLT": "#99cc99", // Orta tonda, hafif soluk yeşil
        "CFR": "#cc99ff", // Açık, pastel yeşil

        // --- DİĞER KATEGORİLERİN MEVCUT ATAMALARI ---

        // --- A Harfi ile Başlayanlar (Profesyonel Maviler) ---
        "AE3EK": "#0B4F6C",
        "AE3ES": "#1D6787",
        "AE3JS": "#348AA7",
        "AE3JV": "#50A8C9",
        "AE3MS": "#6EBDE5",
        "AE4EK": "#0B4F6C",
        "AE4ES": "#1D6787",
        "AE4JS": "#348AA7",
        "AE4JV": "#50A8C9",
        "AE4MS": "#6EBDE5",
        "AE5EK": "#0B4F6C",
        "AE5ES": "#1D6787",
        "AE5JS": "#348AA7",
        "AE5JV": "#50A8C9",
        "AE5MS": "#6EBDE5",
        "AE6EK": "#0B4F6C",
        "AE6ES": "#1D6787",
        "AE6JS": "#348AA7",
        "AE6JV": "#50A8C9",
        "AEGES": "#6EBDE5",

        "AG1EK": "#50A8C9",
        "AG1ES": "#6EBDE5",
        "AG1JS": "#348AA7",
        "AG1JV": "#1D6787",
        "AG1MS": "#50A8C9",
        "AG1ZS": "#6EBDE5",
        "AG2ES": "#348AA7",
        "AG2JS": "#1D6787",
        "AG2MS": "#50A8C9",
        "AG2ZS": "#6EBDE5",

        "AL2ES": "#5C5751", // Nötr Toprak Tonları / Koyu Griler
        "AL2JS": "#7A736B",
        "AL2MS": "#9B948E",
        "AL2ZS": "#BDB7B1",
        "AL3ES": "#5C5751",
        "AL3JS": "#7A736B",
        "AL3JV": "#9B948E",
        "AL3MS": "#BDB7B1",
        "AL3ZS": "#5C5751",
        "AL4ES": "#7A736B",
        "AL4JS": "#9B948E",
        "AL4MS": "#BDB7B1",
        "AL4ZS": "#5C5751",
        "AL5ES": "#7A736B",
        "AL5JS": "#9B948E",
        "AL5MS": "#BDB7B1",
        "AL5ZS": "#5C5751",
        "AL6MS": "#7A736B",
        "AL6ZS": "#9B948E",
        "AWAJ": "#BDB7B1", // Nötr gri

        // --- B Harfi ile Başlayanlar (Derin Yeşiller) ---
        "BUS": "#2E6A5B",

        // --- C Harfi ile Başlayanlar (Nötr Toprak Tonları) ---
        // CFR zaten yukarıda yeniden atanmıştır.

        // --- D Harfi ile Başlayanlar (Derin Yeşiller / Sıcak Nötrler) ---
        "DGR": "#3E8E7D",
        "DLA": "#5CA99B",
        "DLB": "#7EC4B8",
        "DTK": "#C07830", // Sıcak nötr / Kahverengimsi
        "DTT": "#D8975D", // Sıcak nötr / Kahverengimsi
        "DUS": "#ECC69F", // Sıcak nötr / Açık Kahverengimsi
        "DYG": "#3E8E7D",
        "DYT": "#5CA99B",

        // --- E Harfi ile Başlayanlar (Derin Yeşiller / Sıcak Nötrler) ---
        "EBG": "#2E6A5B",
        "EBS": "#3E8E7D",
        "EDC": "#5CA99B",
        "EEF": "#7EC4B8",
        "EES": "#2E6A5B",
        "EKK": "#3E8E7D",
        "EMJ": "#5CA99B",
        "EMM": "#7EC4B8",
        "EPD": "#2E6A5B",
        "EPT": "#3E8E7D",
        "ERC": "#5CA99B",
        "ERT": "#7EC4B8",
        "ES3": "#2E6A5B",
        "ES7": "#3E8E7D",
        "ESA": "#5CA99B",
        "ESD": "#7EC4B8",
        "ESK": "#2E6A5B",
        "ESP": "#3E8E7D",
        "ESX": "#5CA99B",
        "ETT": "#7EC4B8",
        "EUM": "#2E6A5B",
        "EY3": "#C07830", // Sıcak nötr
        "EYL": "#D8975D", // Sıcak nötr
        "EYQ": "#ECC69F", // Sıcak nötr
        "EYR": "#C07830", // Sıcak nötr

        // --- F Harfi ile Başlayanlar (Profesyonel Maviler) ---
        // FLT zaten yukarıda yeniden atanmıştır.

        // --- G Harfi ile Başlayanlar (Nötr Toprak Tonları) ---
        "GLB": "#9B948E",
        "GOR": "#BDB7B1",

        // --- H Harfi ile Başlayanlar (Nötr Toprak Tonları) ---
        "HAK": "#7A736B",
        "HE3EK": "#5C5751",
        "HE3ES": "#7A736B",
        "HE3JS": "#9B948E",
        "HE3JV": "#BDB7B1",
        "HE3MS": "#5C5751",
        "HE4EK": "#7A736B",
        "HE4ES": "#9B948E",
        "HE4JS": "#BDB7B1",
        "HE4JV": "#5C5751",
        "HE4MS": "#7A736B",
        "HE5EK": "#9B948E",
        "HE5ES": "#BDB7B1",
        "HE5JS": "#5C5751",
        "HE5JV": "#7A736B",
        "HE5MS": "#9B948E",
        "HE6EK": "#BDB7B1",
        "HE6ES": "#5C5751",
        "HE6JS": "#7A736B",
        "HE6JV": "#9B948E",
        "HEGES": "#BDB7B1",

        "HG1ES": "#C07830", // Sıcak Nötrler
        "HG1JS": "#D8975D",
        "HG1MS": "#ECC69F",
        "HG1ZS": "#C07830",
        "HG2ES": "#D8975D",
        "HG2JS": "#ECC69F",
        "HG2MS": "#C07830",
        "HG2ZS": "#D8975D",

        "HL2ES": "#9B948E", // Nötr Griler (AL ile benzer)
        "HL2JS": "#BDB7B1",
        "HL2MS": "#9B948E",
        "HL2ZS": "#BDB7B1",
        "HL3ES": "#9B948E",
        "HL3JS": "#BDB7B1",
        "HL3MS": "#9B948E",
        "HL3ZS": "#BDB7B1",
        "HL4ES": "#9B948E",
        "HL4JS": "#BDB7B1",
        "HL4MS": "#9B948E",
        "HL4ZS": "#BDB7B1",
        "HL5ES": "#9B948E",
        "HL5JS": "#BDB7B1",
        "HL5MS": "#9B948E",
        "HL6MS": "#BDB7B1",

        "HSC": "#9B948E", // Nötr Toprak Tonları
        "HSP": "#BDB7B1",
        "HWAJ": "#DDD8D3", // Daha açık nötr gri

        // --- M Harfi ile Başlayanlar (Profesyonel Maviler) ---
        "MCM": "#1D6787",
        "MKU": "#348AA7",
        "MND": "#50A8C9",
        "MUA": "#6EBDE5",

        // --- O Harfi ile Başlayanlar (Derin Yeşiller / Sıcak Nötrler) ---
        "OAF": "#2E6A5B",
        "OB1": "#3E8E7D",
        "OB2": "#5CA99B",
        "OB3": "#2E6A5B",
        "OBF": "#3E8E7D",
        "OCM": "#5CA99B",
        "OCR": "#2E6A5B",
        "OIK": "#3E8E7D",
        "OKD": "#5CA99B",
        "OKE": "#C07830", // Sıcak nötr
        "OKP": "#D8975D", // Sıcak nötr
        "ONE": "#2E6A5B",
        "OOM": "#3E8E7D",
        "OPG": "#5CA99B",
        "OTM": "#2E6A5B",

        // --- S Harfi ile Başlayanlar (Nötr Toprak Tonları) ---
        "S33": "#9B948E",
        "S77": "#BDB7B1",
        "SND": "#9B948E",

        // --- T Harfi ile Başlayanlar (Sıcak Nötrler) ---
        "TLM": "#C07830",
        "TMM": "#D8975D",
        "TSP": "#ECC69F",

        // --- U Harfi ile Başlayanlar (Canlı Morlar/Mürdüm Tonları) ---
        "UDC": "#7D3C6A",
        "UHK": "#9C4E8A",
        "USF": "#BB6BAA",

        // --- X Harfi ile Başlayanlar (Nötr Toprak Tonları) ---
        "XEE": "#5C5751",
        "XEV": "#7A736B",
        "XUT": "#9B948E",
        "XXM": "#5C5751",

        // --- Genel Kategoriler (Manuel olarak atanmış özel renkler) ---
        "ALRT": "#F05252",       // Vurgu Kırmızı
        "GEN_DUTY": "#9B948E",   // Nötr Gri
        "DIFFERENCE": "#DDD8D3", // Çok açık nötr
        "LAYOVER": "#FFF0E0",    // Çok açık sıcak nötr
        "MINIMUM_REST": "#DB8DD0", // Açık mor
        "MODIFIED_REST": "#BB6BAA", // Orta mor
        "REST_OPPORTUNITY": "#6EBDE5", // Açık mavi
        "SIT": "#B02A2A", // Kırmızı vurgu
    };

    activityColorStore.set({ colors: initialColors, defaultColor: get(activityColorStore).defaultColor });
}

// Uygulama başladığında renkleri set edin
initializeActivityColors();

// Bir activity_code için rengi döndüren yardımcı fonksiyon
export function getActivityColor(activityCode: string | undefined): string {
    const storeValue = get(activityColorStore);
    if (activityCode && storeValue.colors[activityCode]) {
        return storeValue.colors[activityCode];
    }
    // Eğer belirli bir kod bulunamazsa veya 'UNK' ise varsayılan rengi döndür
    return storeValue.defaultColor;
}

// Belirli bir aktivite kodunun rengini güncelleyen fonksiyon
export function updateActivityColor(activityCode: string, newColor: string): void {
    activityColorStore.update((store: ActivityColorStoreState) => {
        return {
            ...store,
            colors: {
                ...store.colors,
                [activityCode]: newColor
            }
        };
    });
}

// Varsayılan rengi güncelleyen fonksiyon
export function updateDefaultActivityColor(newColor: string): void {
    activityColorStore.update((store: ActivityColorStoreState) => {
        return {
            ...store,
            defaultColor: newColor
        };
    });
}

// Yardımcı: Rastgele bir renk döndüren fonksiyon (paletten)
export function getRandomColorFromPalette(): string {
    const randomIndex = Math.floor(Math.random() * corporatePalette.length);
    return corporatePalette[randomIndex];
}