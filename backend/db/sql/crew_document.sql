-- crew_documents.sql
CREATE TABLE
    crew_documents (
        -- Birincil Anahtar: Go modelindeki DataID ile eşleşir
        data_id UUID PRIMARY KEY DEFAULT gen_random_uuid (),
        -- Doküman Bilgileri Sütunları:
        person_id TEXT NOT NULL, -- PersonID: Artık UNIQUE değil
        person_surname TEXT NOT NULL,
        person_name TEXT NOT NULL,
        citizenship_number TEXT,
        person_type TEXT,
        ucucu_alt_tipi TEXT,
        ucucu_sinifi TEXT,
        base_filo TEXT,
        dokuman_alt_tipi TEXT NOT NULL,
        -- Tarih alanları: Go modelindeki sql.NullInt64'e karşılık gelen BIGINT
        gecerlilik_baslangic_tarihi BIGINT,
        gecerlilik_bitis_tarihi BIGINT,
        -- String alanları: Go modelindeki sql.NullString'e karşılık gelen TEXT
        document_no TEXT, -- UNIQUE değil, NULL da olabilir.
        dokumani_veren TEXT,
        end_date_leave_job BIGINT,
        -- Boolean alanlar:
        personel_thy_calisiyor_mu BOOLEAN DEFAULT FALSE NOT NULL,
        dokuman_gecerli_mi BOOLEAN DEFAULT FALSE NOT NULL,
        agreement_type TEXT
    );

-- Sadece hız için indeksler kalsın, UNIQUE kısıtlamalar kaldırıldı
CREATE INDEX idx_crew_documents_person_id ON crew_documents (person_id);

CREATE INDEX idx_crew_documents_document_no ON crew_documents (document_no);

CREATE INDEX idx_crew_documents_validity_dates ON crew_documents (
    gecerlilik_baslangic_tarihi,
    gecerlilik_bitis_tarihi
);