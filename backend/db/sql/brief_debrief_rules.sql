-- C:\Users\mucah_wi2yyc2\Desktop\mini_CMS_Desktop_App\backend\db\sql\brief_debrief_rules.sql
CREATE TABLE
    IF NOT EXISTS brief_debrief_rules (
        id SERIAL PRIMARY KEY,
        scenario_type VARCHAR(255) NOT NULL,
        aircraft_type VARCHAR(50) NOT NULL,
        crew_type VARCHAR(50) NOT NULL,
        duty_start_airport VARCHAR(50) NOT NULL,
        brief_duration_min INTEGER NOT NULL,
        debrief_duration_min INTEGER NOT NULL,
        priority INTEGER NOT NULL DEFAULT 0, -- Higher number for higher specificity/priority
        created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
        updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
    );

-- Türk Hava Yolları El Kitabındaki Tablo-2 (Uçuş Ekibi, Kargo Hariç) için Veriler [cite: 885]
INSERT INTO
    brief_debrief_rules (
        scenario_type,
        aircraft_type,
        crew_type,
        duty_start_airport,
        brief_duration_min,
        debrief_duration_min,
        priority
    )
VALUES
    (
        'İlk sektörü görevli, yolculu uçuşlar',
        'DAR GÖVDE',
        'Uçuş Ekibi',
        'IST',
        75,
        30,
        100
    ), -- Brief: 01:15 [cite: 885], Debrief: 00:30 [cite: 1135]
    (
        'İlk sektörü görevli, yolculu uçuşlar',
        'GENİŞ GÖVDE',
        'Uçuş Ekibi',
        'IST',
        90,
        30,
        100
    ), -- Brief: 01:30 [cite: 885], Debrief: 00:30 [cite: 1135]
    (
        'İlk sektörü görevli, yolculu uçuşlar',
        'DAR GÖVDE',
        'Uçuş Ekibi',
        'ISL',
        60,
        30,
        90
    ), -- Brief: 01:00 [cite: 885], Debrief: 00:30 [cite: 1135]
    (
        'İlk sektörü görevli, yolculu uçuşlar',
        'GENİŞ GÖVDE',
        'Uçuş Ekibi',
        'ISL',
        90,
        30,
        90
    ), -- Brief: 01:30 [cite: 885], Debrief: 00:30 [cite: 1135]
    (
        'İlk sektörü görevli, yolculu uçuşlar',
        'DAR GÖVDE',
        'Uçuş Ekibi',
        'SAW',
        60,
        30,
        90
    ), -- Brief: 01:00 [cite: 885], Debrief: 00:30 [cite: 1135]
    (
        'İlk sektörü görevli, yolculu uçuşlar',
        'GENİŞ GÖVDE',
        'Uçuş Ekibi',
        'SAW',
        90,
        30,
        90
    ), -- Brief: 01:30 [cite: 885], Debrief: 00:30 [cite: 1135]
    (
        'İlk sektörü görevli, yolculu uçuşlar',
        'DAR GÖVDE',
        'Uçuş Ekibi',
        'Diğer',
        60,
        30,
        80
    ), -- Brief: 01:00 [cite: 885], Debrief: 00:30 [cite: 1135]
    (
        'İlk sektörü görevli, yolculu uçuşlar',
        'GENİŞ GÖVDE',
        'Uçuş Ekibi',
        'Diğer',
        60,
        30,
        80
    ), -- Brief: 01:00 [cite: 885], Debrief: 00:30 [cite: 1135]
    (
        'Simülatör',
        'Hepsi',
        'Uçuş Ekibi',
        'Hepsi',
        60,
        60,
        70
    ), -- Brief: 01:00 [cite: 885], Debrief: 01:00 [cite: 1135]
    (
        'İlk sektörü görevli, İntikal Uçuşları-Yolcusuz (boş/dağıtım/demo)',
        'Hepsi',
        'Uçuş Ekibi',
        'Hepsi',
        60,
        15,
        70
    ), -- Brief: 01:00 [cite: 885], Debrief: 00:15 [cite: 1135]
    (
        'İlk sektör ekip konumlandırma (pas)',
        'Hepsi',
        'Uçuş Ekibi',
        'Hepsi',
        60,
        0,
        70
    ), -- Brief: 01:00 [cite: 885], Debrief: 00:00 [cite: 1135]
    (
        'Açık Mesai',
        'Hepsi',
        'Uçuş Ekibi',
        'Hepsi',
        60,
        15,
        70
    );

-- Brief: 01:00 [cite: 885], Debrief: 00:15 [cite: 1135]
-- Türk Hava Yolları El Kitabındaki Tablo-3 (Kabin Ekibi) için Veriler [cite: 888]
INSERT INTO
    brief_debrief_rules (
        scenario_type,
        aircraft_type,
        crew_type,
        duty_start_airport,
        brief_duration_min,
        debrief_duration_min,
        priority
    )
VALUES
    (
        'İlk sektörü görevli, yolculu uçuşlar',
        'DAR GÖVDE',
        'Kabin Ekibi',
        'IST',
        75,
        30,
        100
    ), -- Brief: 01:15 [cite: 888], Debrief: 00:30 [cite: 1142]
    (
        'İlk sektörü görevli, yolculu uçuşlar',
        'GENİŞ GÖVDE',
        'Kabin Ekibi',
        'IST',
        90,
        30,
        100
    ), -- Brief: 01:30 [cite: 888], Debrief: 00:30 [cite: 1142]
    (
        'İlk sektörü görevli, yolculu uçuşlar',
        'DAR GÖVDE',
        'Kabin Ekibi',
        'ISL',
        60,
        30,
        90
    ), -- Brief: 01:00 [cite: 888], Debrief: 00:30 [cite: 1142]
    (
        'İlk sektörü görevli, yolculu uçuşlar',
        'GENİŞ GÖVDE',
        'Kabin Ekibi',
        'ISL',
        90,
        30,
        90
    ), -- Brief: 01:30 [cite: 888], Debrief: 00:30 [cite: 1142]
    (
        'İlk sektörü görevli, yolculu uçuşlar',
        'DAR GÖVDE',
        'Kabin Ekibi',
        'SAW',
        60,
        30,
        90
    ), -- Brief: 01:00 [cite: 888], Debrief: 00:30 [cite: 1142]
    (
        'İlk sektörü görevli, yolculu uçuşlar',
        'GENİŞ GÖVDE',
        'Kabin Ekibi',
        'SAW',
        90,
        30,
        90
    ), -- Brief: 01:30 [cite: 888], Debrief: 00:30 [cite: 1142]
    (
        'İlk sektörü görevli, yolculu uçuşlar',
        'DAR GÖVDE',
        'Kabin Ekibi',
        'Diğer',
        60,
        30,
        80
    ), -- Brief: 01:00 [cite: 888], Debrief: 00:30 [cite: 1142]
    (
        'İlk sektörü görevli, yolculu uçuşlar',
        'GENİŞ GÖVDE',
        'Kabin Ekibi',
        'Diğer',
        60,
        30,
        80
    ), -- Brief: 01:00 [cite: 888], Debrief: 00:30 [cite: 1142]
    (
        'İlk sektörü görevli, İntikal Uçuşları-Yolcusuz (boş/dağıtım/demo)',
        'Hepsi',
        'Kabin Ekibi',
        'Hepsi',
        60,
        15,
        70
    ), -- Brief: 01:00 [cite: 888], Debrief: 00:15 [cite: 1142]
    (
        'İlk sektör ekip konumlandırma (pas)',
        'Hepsi',
        'Kabin Ekibi',
        'Hepsi',
        60,
        0,
        70
    ), -- Brief: 01:00 [cite: 888], Debrief: 00:00 [cite: 1142]
    (
        'Açık Mesai',
        'Hepsi',
        'Kabin Ekibi',
        'Hepsi',
        60,
        15,
        70
    );

-- Brief: 01:00 [cite: 888], Debrief: 00:15 [cite: 1142]
-- Türk Hava Yolları El Kitabındaki Tablo-4 (Kargo Uçuş Ekibi) için Veriler [cite: 895]
INSERT INTO
    brief_debrief_rules (
        scenario_type,
        aircraft_type,
        crew_type,
        duty_start_airport,
        brief_duration_min,
        debrief_duration_min,
        priority
    )
VALUES
    (
        'İlk sektörü görevli',
        'Hepsi',
        'Kargo Uçuş Ekibi',
        'Hepsi',
        60,
        30,
        100
    ), -- Brief: 01:00 [cite: 895], Debrief: 00:30 [cite: 1145]
    (
        'İlk sektör ekip konumlandırma (pas)',
        'Hepsi',
        'Kargo Uçuş Ekibi',
        'Hepsi',
        60,
        0,
        90
    ), -- Brief: 01:00 [cite: 895], Debrief: 00:00 [cite: 1145]
    (
        'Açık Mesai',
        'Hepsi',
        'Kargo Uçuş Ekibi',
        'Hepsi',
        60,
        15,
        80
    );

-- Brief: 01:00 [cite: 895], Debrief: 00:15 [cite: 1145]