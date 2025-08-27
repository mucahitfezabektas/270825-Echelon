// C:\Users\mucah_wi2yyc2\Desktop\mini_CMS_Desktop_App\backend\repositories\trip_repo.go

package repositories

import (
	"context"
	"database/sql" // sql.ErrNoRows için hala gerekli, sql.NullTime artık doğrudan kullanılmasa da kalsın
	"fmt"
	"log"
	"time"

	"mini_CMS_Desktop_App/models"

	"github.com/uptrace/bun"
)

type TripRepository struct {
	db *bun.DB
}

func NewTripRepository(db *bun.DB) *TripRepository {
	return &TripRepository{db: db}
}

func (r *TripRepository) SaveTrip(trip *models.Trip) error {
	// violationsJSON ve activitiesJSON değişkenleri kaldırıldı, bun Model() ile otomatik halledilir.

	// <<<< BURADA DÜZELTİLDİ: restStart ve restEnd değişkenleri ve atamaları kaldırıldı
	// bun, models.Trip'teki time.Time alanlarının bun:"...,null" tag'i sayesinde null değerleri otomatik yönetir.

	_, err := r.db.NewInsert().
		Model(trip). // bun'ın Model() metodu JSONB ve time.Time alanlarını otomatik işler
		On("CONFLICT (trip_id) DO UPDATE").
		Set("crew_member_id = EXCLUDED.crew_member_id").
		Set("first_leg_departure_time = EXCLUDED.first_leg_departure_time").
		Set("last_leg_arrival_time = EXCLUDED.last_leg_arrival_time").
		Set("duty_start_airport = EXCLUDED.duty_start_airport").
		Set("duty_type = EXCLUDED.duty_type").
		Set("crew_type = EXCLUDED.crew_type").
		Set("brief_trip_type = EXCLUDED.brief_trip_type").
		Set("debrief_trip_type = EXCLUDED.debrief_trip_type").
		Set("brief_aircraft_type = EXCLUDED.brief_aircraft_type").
		Set("debrief_aircraft_type = EXCLUDED.debrief_aircraft_type").
		Set("calculated_brief_duration_min = EXCLUDED.calculated_brief_duration_min").
		Set("calculated_debrief_duration_min = EXCLUDED.calculated_debrief_duration_min").
		Set("calculated_duty_period_start = EXCLUDED.calculated_duty_period_start").
		Set("calculated_duty_period_end = EXCLUDED.calculated_duty_period_end").
		Set("calculated_duty_period_duration_min = EXCLUDED.calculated_duty_period_duration_min").
		Set("calculated_flight_duty_period_duration_min = EXCLUDED.calculated_flight_duty_period_duration_min").
		// calculated_rest_period_start/end/duration_min ve ftl_violations/activities için manuel Set'ler de kaldırılabilir.
		// bun'ın Model() metodu bunları otomatik halleder.
		Set("calculated_rest_period_start = EXCLUDED.calculated_rest_period_start").
		Set("calculated_rest_period_end = EXCLUDED.calculated_rest_period_end").
		Set("calculated_rest_period_duration_min = EXCLUDED.calculated_rest_period_duration_min").
		Set("ftl_violations = EXCLUDED.ftl_violations").
		Set("activities = EXCLUDED.activities").
		Set("last_calculated_at = EXCLUDED.last_calculated_at").
		Set("updated_at = NOW()").
		Exec(context.Background())

	if err != nil {
		log.Printf("Hata: Trip %s kaydedilirken/güncellenirken sorun oluştu: %v", trip.TripID, err)
		return fmt.Errorf("trip kaydedilirken/güncellenirken hata: %w", err)
	}
	return nil
}

// GetTripByID (mevcut hali, değişmedi)
func (r *TripRepository) GetTripByID(tripID string) (*models.Trip, error) {
	var trip models.Trip
	err := r.db.NewSelect().
		Model(&trip).
		Where("trip_id = ?", tripID).
		Scan(context.Background())

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		log.Printf("Hata: Trip ID %s çekilirken sorun oluştu: %v", tripID, err)
		return nil, fmt.Errorf("trip ID çekilirken hata: %w", err)
	}
	// bun, Nullable alanları ve JSONB alanlarını otomatik olarak yönetir,
	// bu yüzden burada manuel unmarshal/valid kontrolüne gerek kalmaz.
	return &trip, nil
}

// GetTripsByCrewMemberID (mevcut hali, değişmedi)
func (r *TripRepository) GetTripsByCrewMemberID(crewMemberID string, fromTime time.Time) ([]models.Trip, error) {
	var trips []models.Trip
	err := r.db.NewSelect().
		Model(&trips).
		Where("crew_member_id = ?", crewMemberID).
		Where("first_leg_departure_time >= ?", fromTime).
		Order("first_leg_departure_time ASC").
		Scan(context.Background())

	if err != nil {
		return nil, fmt.Errorf("ekip üyesi %s için tripler çekilirken hata: %w", crewMemberID, err)
	}
	// bun, Nullable alanları ve JSONB alanlarını otomatik olarak yönetir.
	return trips, nil
}
