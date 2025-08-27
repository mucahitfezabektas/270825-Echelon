// C:\Users\mucah_wi2yyc2\Desktop\mini_CMS_Desktop_App\backend\models\actual.go

package models

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/uptrace/bun"
)

// Aircraft Type dönüşüm haritası
var aircraftTypeMapping = map[string]string{
	"310": "DAR GÖVDE", "319": "DAR GÖVDE", "320": "DAR GÖVDE", "321": "DAR GÖVDE", "32D": "DAR GÖVDE", "32H": "DAR GÖVDE",
	"37B": "DAR GÖVDE", "3A0": "DAR GÖVDE", "3A1": "DAR GÖVDE", "3HD": "DAR GÖVDE", "3S2": "DAR GÖVDE", "3VF": "DAR GÖVDE",
	"6VF": "DAR GÖVDE", "734": "DAR GÖVDE", "737": "DAR GÖVDE", "738": "DAR GÖVDE", "739": "DAR GÖVDE", "73D": "DAR GÖVDE",
	"73E": "DAR GÖVDE", "73H": "DAR GÖVDE", "73M": "DAR GÖVDE", "73N": "DAR GÖVDE", "73P": "DAR GÖVDE", "73S": "DAR GÖVDE",
	"73V": "DAR GÖVDE", "73Z": "DAR GÖVDE", "74D": "DAR GÖVDE", "78H": "DAR GÖVDE", "78I": "DAR GÖVDE", "78L": "DAR GÖVDE",
	"78T": "DAR GÖVDE", "79Y": "DAR GÖVDE", "79Z": "DAR GÖVDE", "7A8": "DAR GÖVDE", "7B8": "DAR GÖVDE", "7C3": "DAR GÖVDE",
	"7D3": "DAR GÖVDE", "7VF": "DAR GÖVDE", "8VF": "DAR GÖVDE", "9VF": "DAR GÖVDE", "A20": "DAR GÖVDE", "A21": "DAR GÖVDE",
	"A32": "DAR GÖVDE", "A78": "DAR GÖVDE", "B32": "DAR GÖVDE", "B78": "DAR GÖVDE", "C32": "DAR GÖVDE", "D23": "DAR GÖVDE",
	"D32": "DAR GÖVDE", "D73": "DAR GÖVDE", "E20": "DAR GÖVDE", "E21": "DAR GÖVDE", "E32": "DAR GÖVDE", "G32": "DAR GÖVDE",
	"HD3": "DAR GÖVDE", "K21": "DAR GÖVDE", "L20": "DAR GÖVDE", "L21": "DAR GÖVDE", "M32": "DAR GÖVDE", "N32": "DAR GÖVDE",
	"N78": "DAR GÖVDE", "N79": "DAR GÖVDE", "R21": "DAR GÖVDE", "SA1": "DAR GÖVDE", "SC0": "DAR GÖVDE", "SL7": "DAR GÖVDE",
	"SL8": "DAR GÖVDE", "SX1": "DAR GÖVDE", "SY0": "DAR GÖVDE", "TC0": "DAR GÖVDE", "TC1": "DAR GÖVDE", "TC2": "DAR GÖVDE",
	"TC3": "DAR GÖVDE", "TC4": "DAR GÖVDE", "TC5": "DAR GÖVDE", "TC6": "DAR GÖVDE", "TC7": "DAR GÖVDE", "TC8": "DAR GÖVDE",
	"TC9": "DAR GÖVDE", "TKB": "DAR GÖVDE", "TKD": "DAR GÖVDE", "TKE": "DAR GÖVDE", "TKS": "DAR GÖVDE", "U21": "DAR GÖVDE",
	"V20": "DAR GÖVDE", "V21": "DAR GÖVDE", "VFA": "DAR GÖVDE", "VFB": "DAR GÖVDE", "VFC": "DAR GÖVDE", "VFD": "DAR GÖVDE",
	"VFE": "DAR GÖVDE", "VFM": "DAR GÖVDE", "VFQ": "DAR GÖVDE", "VFS": "DAR GÖVDE", "VFT": "DAR GÖVDE", "VFX": "DAR GÖVDE",
	"VIP": "DAR GÖVDE", "W20": "DAR GÖVDE", "W21": "DAR GÖVDE", "W31": "DAR GÖVDE", "W32": "DAR GÖVDE", "W33": "DAR GÖVDE",
	"W34": "DAR GÖVDE", "X20": "DAR GÖVDE", "X21": "DAR GÖVDE", "X32": "DAR GÖVDE", "Y20": "DAR GÖVDE", "Y21": "DAR GÖVDE",
	"Z73": "DAR GÖVDE", "ZB1": "DAR GÖVDE", "ZB2": "DAR GÖVDE", "ZB3": "DAR GÖVDE",

	"330": "GENİŞ GÖVDE", "332": "GENİŞ GÖVDE", "333": "GENİŞ GÖVDE", "33A": "GENİŞ GÖVDE", "33B": "GENİŞ GÖVDE", "33C": "GENİŞ GÖVDE",
	"33E": "GENİŞ GÖVDE", "33F": "GENİŞ GÖVDE", "33H": "GENİŞ GÖVDE", "33I": "GENİŞ GÖVDE", "33J": "GENİŞ GÖVDE", "33M": "GENİŞ GÖVDE",
	"33N": "GENİŞ GÖVDE", "33P": "GENİŞ GÖVDE", "33R": "GENİŞ GÖVDE", "33S": "GENİŞ GÖVDE", "33T": "GENİŞ GÖVDE", "33V": "GENİŞ GÖVDE",
	"33X": "GENİŞ GÖVDE", "33Y": "GENİŞ GÖVDE", "33Z": "GENİŞ GÖVDE", "340": "GENİŞ GÖVDE", "343": "GENİŞ GÖVDE", "350": "GENİŞ GÖVDE",
	"35D": "GENİŞ GÖVDE", "35E": "GENİŞ GÖVDE", "3MF": "GENİŞ GÖVDE", "74F": "GENİŞ GÖVDE", "74G": "GENİŞ GÖVDE", "74I": "GENİŞ GÖVDE",
	"777": "GENİŞ GÖVDE", "77A": "GENİŞ GÖVDE", "77B": "GENİŞ GÖVDE", "77K": "GENİŞ GÖVDE", "77M": "GENİŞ GÖVDE", "77R": "GENİŞ GÖVDE",
	"77X": "GENİŞ GÖVDE", "787": "GENİŞ GÖVDE", "789": "GENİŞ GÖVDE", "B74": "GENİŞ GÖVDE", "C74": "GENİŞ GÖVDE", "D33": "GENİŞ GÖVDE",
	"H77": "GENİŞ GÖVDE", "I77": "GENİŞ GÖVDE", "TK3": "GENİŞ GÖVDE", "TK7": "GENİŞ GÖVDE",
}

// Crew Type helper: FlightPosition'a göre ekip tipi atar
var flightCrewPositions = map[string]bool{
	"C1": true, "C2": true, "C3": true, "C4": true, "CI": true, "CN": true,
	"J1": true, "J2": true, "P1": true, "P2": true, "P3": true, "P4": true, "P5": true, "P6": true,
}

var cabinCrewPositions = map[string]bool{
	"P": true, "L": true, "V": true, "F": true, "K": true, "E": true, "N": true,
	"Y": true, "Q": true, "Z": true, "B": true, "H": true,
}

// ==========================
// === Model Definition ====
// ==========================

type Actual struct {
	bun.BaseModel `bun:"actuals"`

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

// ==========================
// === Utility Functions ====
// ==========================

// ParseTimeFromUnixMs: Unix milisaniye stringini time.Time'a dönüştürür.
// Dışarıdan erişilebilir (fonksiyon adı büyük harfle başlıyor).
func ParseTimeFromUnixMs(value string) (time.Time, error) {
	ts, err := strconv.ParseInt(value, 10, 64)
	if err != nil {
		return time.Time{}, fmt.Errorf("zaman (Unix Ms) ayrıştırılamadı '%s': %w", value, err)
	}
	return time.Unix(ts/1000, (ts%1000)*int64(time.Millisecond)), nil
}

func ParseTimeFromDMYHMS(raw string) (*time.Time, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return nil, fmt.Errorf("boş tarih değeri")
	}

	// 🔄 İzin verilen formatlar
	formats := []string{
		"02.01.2006 15:04:05",
		"02/01/2006 15:04:05", // <- gelen veri bu olabilir
	}

	for _, layout := range formats {
		if t, err := time.ParseInLocation(layout, raw, time.Local); err == nil {
			return &t, nil
		}
	}

	return nil, fmt.Errorf("geçersiz format: '%s'", raw)
}

func (a *Actual) SetPlaneCmsType(cmsType string) {
	a.PlaneCmsType = cmsType
	a.AircraftType = GetAircraftTypeFromCmsType(cmsType)
}

// ==========================
// === Classification Rules ===
// ==========================

func GetAircraftTypeFromCmsType(cmsType string) string {
	if aircraftType, ok := aircraftTypeMapping[cmsType]; ok {
		return aircraftType
	}
	return "BİLİNMİYOR"
}

func GetDutyTypeFromActual(actual *Actual) string {
	if actual.FlightPosition == "DH" {
		return "Konumlandırma"
	}
	switch actual.GroupCode {
	case "SIM":
		return "Simülatör"
	case "GT":
		if strings.EqualFold(actual.ActivityCode, "BUS") || strings.EqualFold(actual.ActivityCode, "OAF") {
			return "Konumlandırma"
		}
		return "Diğer Görev"
	case "OTH":
		return "Açık Mesai"
	case "FLT":
		return "Yolculu Uçuşlar"
	}
	return "Diğer Görev"
}

func GetCrewTypeFromFlightPosition(flightPosition string) string {
	normalized := strings.ToUpper(flightPosition)
	if flightCrewPositions[normalized] {
		return "Uçuş Ekibi"
	}
	if cabinCrewPositions[normalized] {
		return "Kabin Ekibi"
	}
	return "Uçuş Ekibi"
}
