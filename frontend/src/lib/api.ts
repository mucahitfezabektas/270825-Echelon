// src/lib/api.ts

import type {
    FlightItem,
    Actual,
    SavedFilter,
    ActualsByFlightIDResponseBackend,
    ActualsByFlightIDResponseFrontend,
} from "./types";
import { RowType } from "./types";
import { logout } from "@/stores/authStore";
import { showAlert } from "@/stores/alertStore";

const API_BASE_URL = import.meta.env.VITE_API_URL;

if (!API_BASE_URL) {
    console.error("VITE_API_URL ortam değişkeni tanımlı değil! Lütfen .env dosyanızı kontrol edin.");
    throw new Error("API URL'si yapılandırılmadı.");
}

const FIELD_ALIASES_FRONTEND = {
    person_id: "c",
    surname: "s",
    activity_code: "a",
    class: "cl",
    departure_port: "dp",
    arrival_port: "ap",
    date: "d",
    trip_id: "t",
    plane_tail_name: "pt",
    plane_cms_type: "pc",
    group_code: "gc",
    flight_position: "fp",
    flight_no: "fn",
    agreement_type: "at",
    flight_id: "fi",
} as const;

function convertActualToFlightItems(data: Actual[], type: RowType): FlightItem[] {
    return data.map((item: Actual) => ({
        data_id: item.data_id,
        activity_code: item.activity_code,
        name: item.name,
        surname: item.surname,
        base_filo: item.base_filo,
        class: item.class,
        departure_port: item.departure_port,
        arrival_port: item.arrival_port,
        person_id: item.person_id,
        plane_cms_type: item.plane_cms_type,
        plane_tail_name: item.plane_tail_name,
        trip_id: item.trip_id,
        group_code: item.group_code,
        flight_position: item.flight_position,
        flight_no: item.flight_no,
        flight_id: item.flight_id,
        departure_time: new Date(item.departure_time).getTime(),
        arrival_time: new Date(item.arrival_time).getTime(),
        rest_start: item.rest_start ? new Date(item.rest_start).getTime() : undefined,
        rest_end: item.rest_end ? new Date(item.rest_end).getTime() : undefined,
        rest_duration: item.rest_duration,
        type,
    }));
}

// 🌟 Tüm isteklerde kullanılan merkezi fetch: 401/403 → otomatik logout
async function authenticatedFetch(input: RequestInfo | URL, init?: RequestInit): Promise<Response> {
    const token = localStorage.getItem("authToken");

    if (!token) {
        showAlert(
            "Giriş yapmanız gerekiyor.",
            "Yetkilendirme Hatası",
            "error",
            3000
        );
        setTimeout(() => logout(), 3000); // Alert süresi sonunda logout
        throw new Error("Giriş yapılmamış. Lütfen tekrar giriş yapın.");
    }

    const headers = new Headers(init?.headers);
    headers.set("Authorization", `Bearer ${token}`);

    const authInit: RequestInit = { ...init, headers };

    let response: Response;
    try {
        response = await fetch(input, authInit);
    } catch (e) {
        console.error("Ağ hatası:", e);
        throw e;
    }

    if (response.status === 401 || response.status === 403) {
        console.warn(`Yetkilendirme hatası (${response.status}), alert gösterilip çıkış yapılacak...`);
        showAlert(
            "Oturum süreniz doldu veya yetkilendirme başarısız oldu.",
            "Oturum Kapandı",
            "warning",
            3000
        );
        setTimeout(() => logout(), 3000); // Alert kapandıktan sonra logout
        throw new Error("Oturum süreniz doldu veya yetkilendirme başarısız oldu.");
    }

    return response;
}

// --- API Fonksiyonları ---

export async function fetchCrewTimelineData(
    filter: Record<string, string> | SavedFilter
): Promise<{ flights: FlightItem[]; total: number }> {
    const urlPath = "/api/actual/query";
    let requestUrl = `${API_BASE_URL}${urlPath}`;
    let requestOptions: RequestInit = {};

    if ("rows" in filter && "logic" in filter) {
        requestOptions = {
            method: "POST",
            headers: { "Content-Type": "application/json" },
            body: JSON.stringify(filter),
        };
        console.log("API: SavedFilter ile POST isteği gönderiliyor.");
    } else {
        const filtersRecord = filter as Record<string, string>;
        const qParamParts: string[] = [];

        for (const key in filtersRecord) {
            if (Object.prototype.hasOwnProperty.call(filtersRecord, key)) {
                const alias = FIELD_ALIASES_FRONTEND[key as keyof typeof FIELD_ALIASES_FRONTEND];
                if (alias) {
                    qParamParts.push(`${alias} ${filtersRecord[key]}`);
                } else {
                    console.warn(`Bilinmeyen filtre anahtarı: ${key}. Sorguya dahil edilmiyor.`);
                }
            }
        }

        const queryParams = new URLSearchParams();
        if (qParamParts.length > 0) {
            queryParams.append("q", qParamParts.join(" "));
        } else {
            console.warn("fetchCrewTimelineData: Hiçbir filtre sağlanmadı (Record<string, string> tipi).");
        }

        requestUrl += `?${queryParams.toString()}`;
        console.log(`API: Query parametresi ('q') ile GET isteği gönderiliyor: ${requestUrl}`);
    }

    try {
        const response = await authenticatedFetch(requestUrl, requestOptions);

        if (!response.ok) {
            const errorText = await response.text();
            console.error(
                `Ekip aktivite verisi alınamadı: ${response.status} ${response.statusText}`,
                errorText
            );
            let clientErrorMessage = `Ekip aktivite verisi alınamadı (${response.status}).`;
            try {
                const errorJson = JSON.parse(errorText);
                if ((errorJson as any).error) {
                    clientErrorMessage += ` Detay: ${(errorJson as any).error}`;
                } else {
                    clientErrorMessage += ` Ham: ${errorText}`;
                }
            } catch {
                clientErrorMessage += ` Sunucu mesajı: ${errorText}`;
            }
            throw new Error(clientErrorMessage);
        }

        const responseData: { total: number; result: Actual[] } = await response.json();

        if (!responseData || !Array.isArray(responseData.result)) {
            console.warn(`Uyarı: Ekip aktivite veri tipi beklenmeyen formatta geldi.`, responseData);
            return { flights: [], total: 0 };
        }

        const convertedFlightsData = convertActualToFlightItems(responseData.result, RowType.Actual);
        return { flights: convertedFlightsData, total: responseData.total };
    } catch (error) {
        console.error(`Ekip aktivite verisi alınırken hata oluştu:`, error);
        throw error;
    }
}

export async function getPublishDataByPersonId(personId: string): Promise<FlightItem[]> {
    try {
        const response = await authenticatedFetch(
            `${API_BASE_URL}/api/publish/query?person_id=${encodeURIComponent(personId)}`
        );
        if (!response.ok) {
            const errorText = await response.text();
            throw new Error(`Failed to fetch publish data for ${personId}: ${errorText}`);
        }
        const data: Actual[] = await response.json();
        return convertActualToFlightItems(data, RowType.Publish);
    } catch (error) {
        console.error(`Error fetching publish data for ${personId}:`, error);
        throw error;
    }
}

export async function fetchActualsByFlightID(
    flightId: string
): Promise<ActualsByFlightIDResponseFrontend> {
    const url = `${API_BASE_URL}/api/actual/by-flight-id/${encodeURIComponent(flightId)}`;
    console.log(`API: flight_id ile sorgu gönderiliyor: ${url}`);

    try {
        const response = await authenticatedFetch(url);

        if (!response.ok) {
            const errorText = await response.text();
            console.error(
                `flight_id'ye göre actual verisi alınamadı: ${response.status} ${response.statusText}`,
                errorText
            );
            let clientErrorMessage = `flight_id'ye göre actual verisi alınamadı (${response.status}).`;
            try {
                const errorJson = JSON.parse(errorText);
                if ((errorJson as any).error) {
                    clientErrorMessage += ` Detay: ${(errorJson as any).error}`;
                } else {
                    clientErrorMessage += ` Ham: ${errorText}`;
                }
            } catch {
                clientErrorMessage += ` Sunucu mesajı: ${errorText}`;
            }
            throw new Error(clientErrorMessage);
        }

        const responseData: ActualsByFlightIDResponseBackend = await response.json();

        if (!responseData || !Array.isArray(responseData.person_ids) || typeof responseData.result !== "object") {
            console.warn(`Uyarı: flight_id'ye göre Actual veri tipi beklenmeyen formatta geldi.`, responseData);
            return { total_persons_found: 0, person_ids: [], result: {} };
        }

        const transformedResult: { [personId: string]: FlightItem[] } = {};
        for (const personId in responseData.result) {
            if (Object.prototype.hasOwnProperty.call(responseData.result, personId)) {
                const actualsForPerson: Actual[] = responseData.result[personId];
                transformedResult[personId] = convertActualToFlightItems(actualsForPerson, RowType.Actual);
            }
        }

        return {
            total_persons_found: responseData.total_persons_found,
            person_ids: responseData.person_ids,
            result: transformedResult,
        };
    } catch (error) {
        console.error(`flight_id'ye göre actual verisi alınırken hata oluştu:`, error);
        throw error;
    }
}
