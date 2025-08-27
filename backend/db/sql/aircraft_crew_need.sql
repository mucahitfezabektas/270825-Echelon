-- aircraft_crew_need.sql
CREATE TABLE
    aircraft_crew_need (
        -- Birincil Anahtar: Go modelindeki DataID ile eşleşir
        data_id UUID PRIMARY KEY DEFAULT gen_random_uuid (),
        -- Uçak Tipi: Genellikle benzersizdir
        actype TEXT NOT NULL UNIQUE, -- Uçak Tipi (örn: TC1, TC2), benzersiz ve boş bırakılamaz
        -- Mürettebat Pozisyonu Sayıları:
        c_count INTEGER NOT NULL DEFAULT 0, -- C pozisyonu için sayı
        p_count INTEGER NOT NULL DEFAULT 0, -- P pozisyonu için sayı
        j_count INTEGER NOT NULL DEFAULT 0, -- J pozisyonu için sayı
        ef_count INTEGER NOT NULL DEFAULT 0, -- EF pozisyonu için sayı
        a_count INTEGER NOT NULL DEFAULT 0, -- A pozisyonu için sayı
        s_count INTEGER NOT NULL DEFAULT 0, -- S pozisyonu için sayı
        l_count INTEGER NOT NULL DEFAULT 0, -- L pozisyonu için sayı
        ec_count INTEGER NOT NULL DEFAULT 0, -- EC pozisyonu için sayı
        t_count INTEGER NOT NULL DEFAULT 0 -- T pozisyonu için sayı
    );