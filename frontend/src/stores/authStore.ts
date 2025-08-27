// src/stores/authStore.ts
import { writable } from 'svelte/store';

// Authentication state interface
interface AuthState {
    isAuthenticated: boolean;
    user: {
        userId: number;
        username: string;
    } | null;
    token: string | null;
}

// Initial state: localStorage'dan token okuyarak otomatik giriş yapacak
const initialAuthState: AuthState = {
    isAuthenticated: false, // Varsayılan olarak kimlik doğrulanmamış
    user: null,
    token: null,
};

export const authStore = writable<AuthState>(initialAuthState);

// Function to set authentication state after successful login
export function setAuth(userId: number, username: string, token: string) {
    // 🌟 BU SATIRI YENİDEN ETKİNLEŞTİRİLDİ: localStorage'a token kaydediliyor!
    localStorage.setItem('authToken', token);

    authStore.set({
        isAuthenticated: true,
        user: { userId, username },
        token: token,
    });
}

// Function to log out the user
export function logout() {
    // 🌟 localStorage'dan token kaldırılıyor
    localStorage.removeItem('authToken');
    authStore.set({
        isAuthenticated: false,
        user: null,
        token: null,
    });
    // Genellikle çıkış yapıldıktan sonra giriş sayfasına yönlendirme yapılır.
    window.location.href = '/';
}

// 🌟 initializeAuth fonksiyonu uygulama başladığında localStorage'daki token'ı kontrol edecek
export function initializeAuth() {
    const token = localStorage.getItem('authToken');
    if (token) {
        // Token varsa, kullanıcıyı otomatik olarak giriş yapmış kabul et.
        // Gerçek bir uygulamada, burada token'ı doğrulamanız (örneğin bir API çağrısı ile)
        // ve token'dan `userId` ve `username` gibi bilgileri çıkarmanız gerekir.
        // Şimdilik sadece yer tutucu değerler kullanıyoruz.
        console.log("localStorage'da token bulundu, otomatik giriş yapılıyor...");
        authStore.set({
            isAuthenticated: true,
            user: { userId: 0, username: 'AutoLoggedUser' }, // Token'dan gerçek kullanıcı bilgileri alınmalı
            token: token,
        });
    } else {
        console.log("localStorage'da token bulunamadı, kullanıcı kimlik doğrulanmadı.");
        // Token yoksa, varsayılan durumu ayarla
        authStore.set({
            isAuthenticated: false,
            user: null,
            token: null,
        });
    }
}

// initializeAuth'u uygulama başlangıcında sadece bir kez çağırın
initializeAuth();