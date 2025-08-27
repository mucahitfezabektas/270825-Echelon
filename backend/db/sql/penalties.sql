-- penalties.sql
-- Ceza Bilgileri (Penalties) verilerini tutar.
CREATE TABLE
    penalties (
        -- Birincil Anahtar: Go modelindeki DataID ile eşleşir
        data_id UUID PRIMARY KEY DEFAULT gen_random_uuid (),
        -- Penalty Bilgileri Sütunları:
        person_id TEXT NOT NULL, -- PersonID (Genellikle boş bırakılmaz)
        person_surname TEXT NOT NULL, -- PersonSurname (Genellikle boş bırakılmaz)
        person_name TEXT NOT NULL, -- PersonName (Genellikle boş bırakılmaz)
        ucucu_sinifi TEXT, -- UcucuSinifi
        base_filo TEXT, -- BaseFilo
        penalty_code TEXT NOT NULL, -- PenaltyCode (Ceza kodu benzersiz veya boş bırakılmaz olmalı)
        penalty_code_explanation TEXT, -- PenaltyCodeExplanation
        penalty_start_date BIGINT, -- PenaltyStartDate (Timestamp)
        penalty_end_date BIGINT -- PenaltyEndDate (Timestamp)
    );

-- Sıkça yapılan sorgulamalar için indeksler eklemek performans artışı sağlayabilir:
-- Örneğin, `person_id` ve `penalty_code` üzerinde sıkça arama yapılıyorsa ve bir kişiye ait
-- belirli bir ceza kodu aynı başlangıç tarihinde birden fazla olamazsa:
CREATE UNIQUE INDEX idx_penalties_person_penalty_date ON penalties (person_id, penalty_code, penalty_start_date);

-- Sadece `person_id` üzerinde arama yapılıyorsa:
CREATE INDEX idx_penalties_person_id ON penalties (person_id);