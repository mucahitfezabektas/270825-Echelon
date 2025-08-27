-- C:\Users\mucah_wi2yyc2\Desktop\mini_CMS_Desktop_App\backend\db\sql\trips.sql
CREATE TABLE
    IF NOT EXISTS trips (
        trip_id VARCHAR(255) PRIMARY KEY,
        crew_member_id VARCHAR(255) NOT NULL,
        first_leg_departure_time TIMESTAMP
        WITH
            TIME ZONE NOT NULL,
            last_leg_arrival_time TIMESTAMP
        WITH
            TIME ZONE NOT NULL,
            aircraft_type VARCHAR(50) NOT NULL,
            duty_start_airport VARCHAR(50) NOT NULL,
            duty_type VARCHAR(255) NOT NULL,
            crew_type VARCHAR(255) NOT NULL,
            brief_trip_type VARCHAR(255) NOT NULL, -- <<<< BURADA EKLENDİ
            debrief_trip_type VARCHAR(255) NOT NULL, -- <<<< BURADA EKLENDİ
            calculated_brief_duration_min INTEGER NOT NULL,
            calculated_debrief_duration_min INTEGER NOT NULL,
            calculated_duty_period_start TIMESTAMP
        WITH
            TIME ZONE NOT NULL,
            calculated_duty_period_end TIMESTAMP
        WITH
            TIME ZONE NOT NULL,
            calculated_duty_period_duration_min INTEGER NOT NULL,
            calculated_flight_duty_period_duration_min INTEGER NOT NULL,
            calculated_rest_period_start TIMESTAMP
        WITH
            TIME ZONE, -- Dinlenme başlaması, NULL olabilir
            calculated_rest_period_end TIMESTAMP
        WITH
            TIME ZONE, -- Dinlenme bitişi, NULL olabilir
            calculated_rest_period_duration_min INTEGER, -- Dinlenme süresi, NULL olabilir
            ftl_violations JSONB, -- []string tipini JSONB olarak saklamak için, NULL olabilir
            last_calculated_at TIMESTAMP
        WITH
            TIME ZONE DEFAULT CURRENT_TIMESTAMP,
            created_at TIMESTAMP
        WITH
            TIME ZONE DEFAULT CURRENT_TIMESTAMP,
            updated_at TIMESTAMP
        WITH
            TIME ZONE DEFAULT CURRENT_TIMESTAMP
    );

-- FTL hesaplamaları için crew_member_id ve zaman bazında hızlı arama için indeksler
-- Bu indeks, bir ekip üyesinin geçmiş görevlerini hızlıca çekmek için kritik olacaktır.
CREATE INDEX IF NOT EXISTS idx_trips_crew_member_id_departure_time ON trips (crew_member_id, first_leg_departure_time);