// frontend\src\lib\tableDataTypes.ts
export interface ActivityCode {
    unique_id: string;
    activity_code: string;
    activity_group_code: string;
    activity_code_explanation: string;
}

export interface AircraftCrewNeed {
    unique_id: string;
    actype: string;
    c_count: number;
    p_count: number;
    j_count: number;
    ef_count: number;
    a_count: number;
    s_count: number;
    l_count: number;
    ec_count: number;
    t_count: number;
}

export interface OffDayTable {
    unique_id: string;
    work_days: number;
    off_day_entitlement: number;
    distribution: string;
}

// CrewDocument kaldırıldı, bu arayüz burada olmamalı
/*
export interface CrewDocument {
    unique_id: string;
    person_id: string;
    person_surname: string;
    person_name: string;
    citizenship_number: string;
    person_type: string;
    ucucu_alt_tipi: string;
    ucucu_sinifi: string;
    base_filo: string;
    dokuman_alt_tipi: string;
    gecerlilik_baslangic_tarihi: number | null;
    gecerlilik_bitis_tarihi: number | null;
    document_no: string | null;
    dokumani_veren: string | null;
    end_date_leave_job: number | null;
    personel_thy_calisiyor_mu: boolean;
    dokuman_gecerli_mi: boolean;
    agreement_type: string;
}
*/

// frontend\src\lib\tableDataTypes.ts
export interface CrewInfo {
    unique_id: string;
    person_id: string;
    person_surname: string;
    person_name: string;
    gender: string;
    tabiiyet: string;
    base_filo: string | null;
    dogum_tarihi: number | null;
    base_location: string | null; // ⭐ Changed to string | null
    ucucu_tipi: string | null;
    oml: string | null; // ⭐ Changed to string | null
    seniority: number | null;
    rank_change_date: number | null;
    rank: string | null;
    agreement_type: string | null;
    agreement_type_explanation: string | null;
    job_start_date: number | null;
    job_end_date: number | null;
    marriage_date: number | null;
    ucucu_sinifi: string | null;
    ucucu_sinifi_last_valid: string | null;
    ucucu_alt_tipi: string | null;
    person_thy_calisiyor_mu: boolean;
    birthplace: string | null;
    period_info: string | number | null;
    service_use_home_pickup: boolean;
    service_use_saw: boolean;
    bridge_use: boolean;
}

export interface Penalty {
    unique_id: string;
    person_id: string;
    person_surname: string;
    person_name: string;
    ucucu_sinifi: string;
    base_filo: string;
    penalty_code: string;
    penalty_code_explanation: string;
    penalty_start_date: number;
    penalty_end_date: number;
}