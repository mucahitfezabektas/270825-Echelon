CREATE TABLE
    actuals ( -- <-- Tablo adını 'actuals' olarak değiştirdik!
        data_id UUID PRIMARY KEY DEFAULT gen_random_uuid (),
        group_code TEXT,
        activity_code TEXT,
        person_id TEXT,
        surname TEXT,
        name TEXT,
        base_filo TEXT,
        class TEXT,
        flight_position TEXT,
        flight_no TEXT,
        departure_port TEXT,
        arrival_port TEXT,
        departure_time BIGINT,
        arrival_time BIGINT,
        plane_cms_type TEXT,
        plane_tail_name TEXT,
        trip_id TEXT,
        checkin_date BIGINT,
        duty_start BIGINT,
        duty_end BIGINT,
        period_month TEXT
    );