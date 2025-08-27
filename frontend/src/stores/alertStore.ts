// src/stores/alertStore.ts
import { writable } from 'svelte/store';

interface AlertState {
    isOpen: boolean;
    message: string;
    title?: string;
    duration?: number; // YENİ: Otomatik kapanma süresi (ms)
    type?: 'info' | 'success' | 'warning' | 'error'; // YENİ: Bildirim tipi
}

const defaultAlertState: AlertState = {
    isOpen: false,
    message: '',
    title: 'Bilgilendirme', // Varsayılan başlığı 'Bilgilendirme' olarak değiştirmek daha mantıklı
    duration: 3000, // Varsayılan 3 saniye
    type: 'info' // Varsayılan tip 'info'
};

export const alertStore = writable<AlertState>(defaultAlertState);

// showAlert fonksiyonu artık `type` ve `duration` parametrelerini kabul ediyor
export function showAlert(
    message: string,
    title?: string,
    type: 'info' | 'success' | 'warning' | 'error' = 'info', // Varsayılan 'info'
    duration: number = 3000 // Varsayılan 3000ms
) {
    alertStore.set({
        isOpen: true,
        message,
        title: title || defaultAlertState.title,
        type,
        duration
    });
}

export function hideAlert() {
    alertStore.set({ ...defaultAlertState, isOpen: false });
}