// C:\Users\mucah_wi2yyc2\Desktop\mini_CMS_Desktop_App\backend\models\trip.go

package models

import "time"

// Trip struct'ı, bir ekip üyesinin belirli bir trip_id altındaki tüm aktivitelerini ve FTL hesaplama sonuçlarını barındırır.
type Trip struct {
	// Veritabanı sütunları
	TripID       string   `json:"trip_id" bun:"trip_id,pk"`
	CrewMemberID string   `json:"crew_member_id" bun:"crew_member_id"`
	Activities   []Actual `json:"activities" bun:"activities,type:jsonb"`

	// Trip'in genel özellikleri (brief/debrief ve FTL hesaplamaları için gerekli özet bilgiler)
	FirstLegDepartureTime time.Time `json:"first_leg_departure_time" bun:"first_leg_departure_time"`
	LastLegArrivalTime    time.Time `json:"last_leg_arrival_time" bun:"last_leg_arrival_time"`
	DutyStartAirport      string    `json:"duty_start_airport" bun:"duty_start_airport"`
	DutyType              string    `json:"duty_type" bun:"duty_type"` // Trip'in genel görev tipi (korunabilir veya kaldırılabilir)
	CrewType              string    `json:"crew_type" bun:"crew_type"`

	// <<<< BURADA EKLENDİ: Brief ve Debrief için spesifik görev tipleri
	BriefTripType   string `json:"brief_trip_type" bun:"brief_trip_type"`     // Briefing fazındaki aktivite tipi (örn. "Yolculu Uçuşlar", "Simülatör", "Konumlandırma")
	DebriefTripType string `json:"debrief_trip_type" bun:"debrief_trip_type"` // Debriefing fazındaki aktivite tipi (örn. "Yolculu Uçuşlar", "Konumlandırma", "Açık Mesai")

	BriefAircraftType   string `json:"brief_aircraft_type" bun:"brief_aircraft_type"`
	DebriefAircraftType string `json:"debrief_aircraft_type" bun:"debrief_aircraft_type"`

	// Hesaplanan Brief/Debrief Süreleri (Dakika Cinsinden)
	CalculatedBriefDurationMin   int `json:"calculated_brief_duration_min" bun:"calculated_brief_duration_min"`
	CalculatedDebriefDurationMin int `json:"calculated_debrief_duration_min" bun:"calculated_debrief_duration_min"`

	// Hesaplanan Görev Süresi Detayları (Duty Period)
	CalculatedDutyPeriodStart       time.Time `json:"calculated_duty_period_start" bun:"calculated_duty_period_start"`
	CalculatedDutyPeriodEnd         time.Time `json:"calculated_duty_period_end" bun:"calculated_duty_period_end"`
	CalculatedDutyPeriodDurationMin int       `json:"calculated_duty_period_duration_min" bun:"calculated_duty_period_duration_min"`

	// Hesaplanan Uçuş Görev Süresi (UGS - Flight Duty Period) Detayları
	CalculatedFlightDutyPeriodDurationMin int `json:"calculated_flight_duty_period_duration_min" bun:"calculated_flight_duty_period_duration_min"`

	// Hesaplanan Dinlenme Süresi Detayları (ÖNCEKİ görev ile bu görev arasındaki)
	CalculatedRestPeriodStart       time.Time `json:"calculated_rest_period_start,omitempty" bun:"calculated_rest_period_start,null"`
	CalculatedRestPeriodEnd         time.Time `json:"calculated_rest_period_end,omitempty" bun:"calculated_rest_period_end,null"`
	CalculatedRestPeriodDurationMin int       `json:"calculated_rest_period_duration_min,omitempty" bun:"calculated_rest_period_duration_min,null"`

	// FTL İhlalleri (birden fazla ihlal olabilir), JSONB olarak saklanacak
	FTLViolations []string `json:"ftl_violations" bun:"ftl_violations,type:jsonb,null"`

	// Oluşturulma ve Güncellenme zamanları (bun.BaseModel'den gelmiyorsa)
	LastCalculatedAt time.Time `json:"last_calculated_at" bun:"last_calculated_at"`
	CreatedAt        time.Time `json:"created_at" bun:"created_at,default:current_timestamp"`
	UpdatedAt        time.Time `json:"updated_at" bun:"updated_at,default:current_timestamp"`
}

// TableName, bun ORM'in bu struct'ı 'trips' tablosuyla eşleştirmesini sağlar.
func (Trip) TableName() string {
	return "trips"
}
