-- crew_info.sql
-- Ekip Bilgileri (Crew Info) verilerini tutar.
CREATE TABLE
    crew_info (
        -- Birincil Anahtar: Go modelindeki DataID ile eşleşir
        data_id UUID PRIMARY KEY DEFAULT gen_random_uuid (),
        -- Ekip Bilgileri Sütunları:
        person_id TEXT NOT NULL, -- PersonID
        person_surname TEXT NOT NULL, -- PersonSurname
        person_name TEXT NOT NULL, -- PersonName
        gender TEXT, -- Gender
        tabiiyet TEXT, -- Tabiiyet (Uygulamaya göre NOT NULL yapılabilir)
        base_filo TEXT, -- BaseFilo
        dogum_tarihi BIGINT, -- Doğum Tarihi (Timestamp)
        base_location TEXT, -- BaseLocation
        ucucu_tipi TEXT, -- UcucuTipi
        oml TEXT, -- OML
        seniority TEXT, -- Seniority
        rank_change_date BIGINT, -- RankChangeDate (Timestamp)
        rank TEXT, -- Rank
        agreement_type TEXT, -- AgreementType
        agreement_type_explanation TEXT, -- AgreementTypeExplanation
        job_start_date BIGINT, -- JobStartDate (Timestamp)
        job_end_date BIGINT, -- JobEndDate (Timestamp, NULLable olabilir)
        marriage_date BIGINT, -- MarriageDate (Timestamp, NULLable olabilir)
        ucucu_sinifi TEXT, -- UcucuSinifi
        ucucu_sinifi_last_valid TEXT, -- UcucuSinifiLastValid
        ucucu_alt_tipi TEXT, -- UcucuAltTipi
        person_thy_calisiyor_mu BOOLEAN, -- PersonThyCalisiyorMu
        birthplace TEXT, -- Birthplace
        period_info TEXT, -- PeriodInfo
        service_use_home_pickup BOOLEAN, -- ServiceUseHomePickup
        service_use_saw BOOLEAN, -- ServiceUseSaw
        bridge_use BOOLEAN -- BridgeUse
    );

-- Sıkça yapılan sorgulamalar için indeksler eklemek performans artışı sağlayabilir:
-- Örneğin, `person_id` üzerinde sıkça arama yapılıyorsa ve benzersizse:
CREATE UNIQUE INDEX idx_crew_info_person_id ON crew_info (person_id);

-- İsim ve soyisim kombinasyonu üzerinde arama yapılıyorsa:
CREATE INDEX idx_crew_info_name_surname ON crew_info (person_surname, person_name);