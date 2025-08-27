package models

import (
	"github.com/google/uuid"
	"github.com/uptrace/bun"
)

// CrewInfo veritabanındaki "crew_info" tablosunu temsil eder.
type CrewInfo struct {
	bun.BaseModel `bun:"table:crew_info"` // Tablo adını açıkça belirtiyoruz

	// Her tablonun DataID'si birincil anahtar olacak
	DataID uuid.UUID `bun:"data_id,pk,type:uuid,default:gen_random_uuid()" json:"data_id"`

	// Ekip Bilgileri Başlıkları:
	PersonID                 string `bun:"person_id" json:"person_id"`
	PersonSurname            string `bun:"person_surname" json:"person_surname"`
	PersonName               string `bun:"person_name" json:"person_name"`
	Gender                   string `bun:"gender" json:"gender"`
	Tabiiyet                 string `bun:"tabiiyet" json:"tabiiyet"`
	BaseFilo                 string `bun:"base_filo" json:"base_filo"`
	DogumTarihi              int64  `bun:"dogum_tarihi" json:"dogum_tarihi"` // Timestamp
	BaseLocation             string `bun:"base_location" json:"base_location"`
	UcucuTipi                string `bun:"ucucu_tipi" json:"ucucu_tipi"`
	OML                      string `bun:"oml" json:"oml"`                           // OML genellikle metin olabilir (örn: "A", "B", "C")
	Seniority                string `bun:"seniority" json:"seniority"`               // Seniority numarası metin de olabilir (örn: "S123")
	RankChangeDate           int64  `bun:"rank_change_date" json:"rank_change_date"` // Timestamp
	Rank                     string `bun:"rank" json:"rank"`
	AgreementType            string `bun:"agreement_type" json:"agreement_type"`
	AgreementTypeExplanation string `bun:"agreement_type_explanation" json:"agreement_type_explanation"`
	JobStartDate             int64  `bun:"job_start_date" json:"job_start_date"` // Timestamp
	JobEndDate               int64  `bun:"job_end_date" json:"job_end_date"`     // Timestamp, nullable olabilirse özel tip gerekebilir
	MarriageDate             int64  `bun:"marriage_date" json:"marriage_date"`   // Timestamp, nullable olabilirse özel tip gerekebilir
	UcucuSinifi              string `bun:"ucucu_sinifi" json:"ucucu_sinifi"`
	UcucuSinifiLastValid     string `bun:"ucucu_sinifi_last_valid" json:"ucucu_sinifi_last_valid"`
	UcucuAltTipi             string `bun:"ucucu_alt_tipi" json:"ucucu_alt_tipi"`
	PersonThyCalisiyorMu     bool   `bun:"person_thy_calisiyor_mu" json:"person_thy_calisiyor_mu"` // Boolean
	Birthplace               string `bun:"birthplace" json:"birthplace"`
	PeriodInfo               string `bun:"period_info" json:"period_info"`
	ServiceUseHomePickup     bool   `bun:"service_use_home_pickup" json:"service_use_home_pickup"` // Boolean
	ServiceUseSaw            bool   `bun:"service_use_saw" json:"service_use_saw"`                 // Boolean
	BridgeUse                bool   `bun:"bridge_use" json:"bridge_use"`                           // Boolean
}
