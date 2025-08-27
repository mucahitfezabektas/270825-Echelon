package models

import (
	"github.com/google/uuid"
	"github.com/uptrace/bun"
)

// Penalty veritabanındaki "penalties" tablosunu temsil eder.
type Penalty struct {
	bun.BaseModel `bun:"table:penalties"` // Tablo adını açıkça belirtiyoruz

	// Her tablonun DataID'si birincil anahtar olacak
	DataID uuid.UUID `bun:"data_id,pk,type:uuid,default:gen_random_uuid()" json:"data_id"`

	// Penalty Bilgileri Sütunları:
	PersonID               string `bun:"person_id" json:"person_id"`
	PersonSurname          string `bun:"person_surname" json:"person_surname"`
	PersonName             string `bun:"person_name" json:"person_name"`
	UcucuSinifi            string `bun:"ucucu_sinifi" json:"ucucu_sinifi"`
	BaseFilo               string `bun:"base_filo" json:"base_filo"`
	PenaltyCode            string `bun:"penalty_code" json:"penalty_code"`
	PenaltyCodeExplanation string `bun:"penalty_code_explanation" json:"penalty_code_explanation"`
	PenaltyStartDate       int64  `bun:"penalty_start_date" json:"penalty_start_date"` // Timestamp
	PenaltyEndDate         int64  `bun:"penalty_end_date" json:"penalty_end_date"`     // Timestamp
}
