CREATE TABLE
    off_day_table (
        data_id UUID PRIMARY KEY DEFAULT gen_random_uuid (),
        work_days INTEGER NOT NULL, -- Çalışma günleri sayısı (tam sayı)
        off_day_entitlement INTEGER NOT NULL, -- Kazanılan izin günü sayısı (tam sayı)
        distribution TEXT -- Dağılım veya açıklama (metin)
    );