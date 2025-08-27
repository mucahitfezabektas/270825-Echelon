// C:\Users\mucah_wi2yyc2\Desktop\mini_CMS_Desktop_App\backend\models\publish.go

package models

import (
	"time"

	"github.com/google/uuid"
	"github.com/uptrace/bun"
)

// ==========================
// === Model Definition ====
// ==========================

type Publish struct {
	bun.BaseModel `bun:"publishes"`

	DataID        uuid.UUID `bun:"data_id,pk,type:uuid,default:gen_random_uuid()" json:"data_id"`
	UçuşID        string    `bun:"ucus_id" json:"flight_id"`
	ActivityCode  string    `json:"activity_code"`
	Name          string    `json:"name"`
	Surname       string    `json:"surname"`
	BaseFilo      string    `json:"base_filo"`
	Class         string    `json:"class"`
	DeparturePort string    `json:"departure_port"`
	ArrivalPort   string    `json:"arrival_port"`

	DepartureTime time.Time `json:"departure_time"`
	ArrivalTime   time.Time `json:"arrival_time"`

	PersonID string `json:"person_id"`

	PlaneCmsType string `json:"plane_cms_type"`
	AircraftType string `json:"aircraft_type"`

	PlaneTailName  string    `json:"plane_tail_name"`
	TripID         string    `json:"trip_id"`
	GroupCode      string    `json:"group_code"`
	FlightPosition string    `json:"flight_position"`
	FlightNo       string    `json:"flight_no"`
	CheckinDate    time.Time `json:"checkin_date"`
	DutyStart      time.Time `json:"duty_start"`
	DutyEnd        time.Time `json:"duty_end"`
	PeriodMonth    string    `json:"period_month"`
}
