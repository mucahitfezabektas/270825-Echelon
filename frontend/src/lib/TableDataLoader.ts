// frontend\src\lib\TableDataLoader.ts
import type { ActivityCode, AircraftCrewNeed, CrewInfo, OffDayTable, Penalty } from "@/lib/tableDataTypes";
import { logInfo, logWarn } from "@/stores/logStore";
import { get } from "svelte/store";
import { authStore } from "@/stores/authStore";

const BASE_URL = "http://localhost:8080";

class TableDataLoader {
    private static instance: TableDataLoader;

    private _data = {
        activityCodes: [] as ActivityCode[],
        crewInfo: [] as CrewInfo[],
        penalties: [] as Penalty[],
        offDayTable: [] as OffDayTable[],
        aircraftCrewNeed: [] as AircraftCrewNeed[],
    };

    private _loaded = false;
    private _loadingPromise: Promise<void> | null = null;

    private constructor() {
        logInfo("🔁 TableDataLoader singleton initialized.");
    }

    static getInstance(): TableDataLoader {
        if (!TableDataLoader.instance) {
            TableDataLoader.instance = new TableDataLoader();
        }
        return TableDataLoader.instance;
    }

    private getNullStringValue(obj: any): string | null {
        if (obj === null || obj === undefined || (typeof obj === 'string' && obj.trim() === '')) {
            return null;
        }
        if (typeof obj === 'object' && 'Valid' in obj) {
            return obj.Valid ? obj.String : null;
        }
        return String(obj);
    }

    private getNullIntValue(obj: any): number | null {
        if (obj === null || obj === undefined || (typeof obj === 'string' && obj.trim() === '')) {
            return null;
        }
        if (typeof obj === 'object' && 'Valid' in obj) {
            return obj.Valid ? obj.Int64 : null;
        }
        return typeof obj === 'number' ? obj : null;
    }

    private parseBooleanHelper(value: any): boolean {
        if (typeof value === 'boolean') return value;
        if (typeof value === 'string') {
            const lowerCaseValue = value.toLowerCase().trim();
            return lowerCaseValue === 'true' || lowerCaseValue === 'calisiyor' || lowerCaseValue === 'gecerli' || lowerCaseValue === 'y';
        }
        return false;
    }

    private parseNumericHelper(value: any): number | null {
        if (typeof value === 'number') return value;
        if (typeof value === 'string') {
            const cleanedValue = value.replace(',', '.').trim();
            const parsed = parseFloat(cleanedValue);
            if (!isNaN(parsed)) {
                return parsed;
            }
        }
        return null;
    }

    // ⭐ GÜNCELLENDİ: transformCrewInfoData - Tüm alanlar için getNullStringValue veya getNullIntValue kullanıldı
    private transformCrewInfoData(item: any): CrewInfo {
        // Console log'u kaldırıldı veya yorum satırı yapıldı, aksi halde çok fazla çıktıya neden olabilir.
        // console.log("Transforming CrewInfo item (raw):", item); 

        return {
            unique_id: String(item.data_id || ''), // DataID'nin string olduğundan emin olun
            person_id: String(item.person_id || ''),
            person_surname: String(item.person_surname || ''),
            person_name: String(item.person_name || ''),
            gender: String(item.gender || ''),
            tabiiyet: String(item.tabiiyet || ''),

            base_filo: this.getNullStringValue(item.base_filo),
            dogum_tarihi: this.getNullIntValue(item.dogum_tarihi),
            base_location: this.getNullStringValue(item.base_location), // getNullStringValue ile işlendi
            ucucu_tipi: this.getNullStringValue(item.ucucu_tipi),
            oml: this.getNullStringValue(item.oml), // getNullStringValue ile işlendi
            seniority: this.parseNumericHelper(item.seniority), // Zaten doğru
            rank_change_date: this.getNullIntValue(item.rank_change_date),
            rank: this.getNullStringValue(item.rank),
            agreement_type: this.getNullStringValue(item.agreement_type),
            agreement_type_explanation: this.getNullStringValue(item.agreement_type_explanation),
            job_start_date: this.getNullIntValue(item.job_start_date),
            job_end_date: this.getNullIntValue(item.job_end_date),
            marriage_date: this.getNullIntValue(item.marriage_date),
            ucucu_sinifi: this.getNullStringValue(item.ucucu_sinifi),
            ucucu_sinifi_last_valid: this.getNullStringValue(item.ucucu_sinifi_last_valid),
            ucucu_alt_tipi: this.getNullStringValue(item.ucucu_alt_tipi),
            person_thy_calisiyor_mu: this.parseBooleanHelper(item.person_thy_calisiyor_mu),
            birthplace: this.getNullStringValue(item.birthplace),
            period_info: this.getNullStringValue(item.period_info), // period_info da null olabilir
            service_use_home_pickup: this.parseBooleanHelper(item.service_use_home_pickup),
            service_use_saw: this.parseBooleanHelper(item.service_use_saw),
            bridge_use: this.parseBooleanHelper(item.bridge_use),
        };
    }

    async loadAll(): Promise<void> {
        if (this._loaded) {
            logInfo("📦 TableDataLoader: Veriler zaten yüklendi.");
            return;
        }

        if (this._loadingPromise) {
            return this._loadingPromise;
        }

        const authToken = get(authStore).token;
        if (!authToken) {
            logWarn("❌ TableDataLoader: Veri yüklemek için yetkilendirme token'ı mevcut değil. Yükleme durduruldu.");
            return Promise.reject("Yetkilendirme token'ı eksik.");
        }


        this._loadingPromise = (async () => {
            logInfo("📥 TableDataLoader: Referans verileri yükleniyor...");

            try {
                const crewInfoRaw: any[] = await this.fetch<any[]>("/api/crew-info/list", "Ekip_Bilgileri.xlsx");
                this._data.crewInfo = crewInfoRaw.map(item => this.transformCrewInfoData(item));
                console.log("Transformed CrewInfo:", this._data.crewInfo);

                const [a, p, o, n] = await Promise.all([
                    this.fetch<ActivityCode>("/api/activity-codes/list", "Aktivite_Kodlari.xlsx"),
                    this.fetch<Penalty>("/api/penalties/list", "Penalty.xlsx"),
                    this.fetch<OffDayTable>("/api/off-day-table/list", "Bos_Gun_Table.xlsx"),
                    this.fetch<AircraftCrewNeed>("/api/aircraft-crew-need/list", "Aircraft_Crew_Need.xlsx"),
                ]);

                this._data.activityCodes = a;
                this._data.penalties = p;
                this._data.offDayTable = o;
                this._data.aircraftCrewNeed = n;
                
                this._loaded = true;

                logInfo("✅ Tüm veriler başarıyla yüklendi.");
            } catch (err) {
                logWarn(`🚨 Veri yükleme hatası: ${err}`);
                this._loaded = false; 
                throw err;
            } finally {
                this._loadingPromise = null;
            }
        })();

        return this._loadingPromise;
    }

    private async fetch<T>(url: string, filename: string): Promise<T[]> {
        const authToken = get(authStore).token;
        if (!authToken) {
            logWarn(`❌ [${filename}] yüklenemedi: Yetkilendirme token'ı mevcut değil.`);
            return [];
        }

        try {
            logInfo(`🔎 [${filename}] verisi isteniyor: ${BASE_URL}${url}...`);
            const res = await fetch(`${BASE_URL}${url}`, {
                cache: "no-store",
                headers: {
                    'Authorization': `Bearer ${authToken}`
                }
            });
            if (!res.ok) {
                const errorText = await res.text();
                let errorMessage = errorText;
                try {
                    const errorData = JSON.parse(errorText);
                    errorMessage = errorData.error || errorData.message || errorText;
                } catch (jsonError) {
                    console.warn("API yanıtı JSON formatında değil, ham metin kullanılıyor:", errorText);
                }
                logWarn(`❌ [${filename}] HTTP ${res.status}: ${errorMessage}`);
                return [];
            }

            const json = await res.json();
            if (!Array.isArray(json)) {
                logWarn(`❌ [${filename}] yüklenirken beklenmeyen yanıt formatı: JSON bir dizi değil.`, json);
                return [];
            }
            
            logInfo(`✅ [${filename}] yüklendi (${json.length} kayıt)`);
            return json;
        } catch (err) {
            logWarn(`❌ [${filename}] yüklenirken ağ veya parse hatası: ${err}`);
            return [];
        }
    }

    get activityCodes(): ActivityCode[] {
        return this._data.activityCodes;
    }

    get aircraftCrewNeed(): AircraftCrewNeed[] {
        return this._data.aircraftCrewNeed;
    }

    get crewInfo(): CrewInfo[] {
        return this._data.crewInfo;
    }

    get penalties(): Penalty[] {
        return this._data.penalties;
    }

    get offDayTable(): OffDayTable[] {
        return this._data.offDayTable;
    }

    get isLoaded(): boolean {
        return this._loaded;
    }
}

const loader = TableDataLoader.getInstance();
export default loader;