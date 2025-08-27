// models/crew_document.go (Bu dosya doğru)
package models

import (
	"database/sql" // sql.NullInt64 ve sql.NullString için

	"github.com/google/uuid"
	"github.com/uptrace/bun"
)

type CrewDocument struct {
	bun.BaseModel `bun:"table:crew_documents"`
	DataID        uuid.UUID `bun:"data_id,pk,type:uuid,default:gen_random_uuid()" json:"data_id"`

	PersonID          string `bun:"person_id" json:"person_id"`
	PersonSurname     string `bun:"person_surname" json:"person_surname"`
	PersonName        string `bun:"person_name" json:"person_name"`
	CitizenshipNumber string `bun:"citizenship_number" json:"citizenship_number"`
	PersonType        string `bun:"person_type" json:"person_type"`
	UcucuAltTipi      string `bun:"ucucu_alt_tipi" json:"ucucu_alt_tipi"`
	UcucuSinifi       string `bun:"ucucu_sinifi" json:"ucucu_sinifi"`
	BaseFilo          string `bun:"base_filo" json:"base_filo"`
	DokumanAltTipi    string `bun:"dokuman_alt_tipi" json:"dokuman_alt_tipi"`
	// Tarih alanları sql.NullInt64 olacak
	GecerlilikBaslangicTarihi sql.NullInt64 `bun:"gecerlilik_baslangic_tarihi" json:"gecerlilik_baslangic_tarihi"`
	GecerlilikBitisTarihi     sql.NullInt64 `bun:"gecerlilik_bitis_tarihi" json:"gecerlilik_bitis_tarihi"`
	// String alanlar sql.NullString olacak
	DocumentNo      sql.NullString `bun:"document_no" json:"document_no"`
	DokumaniVeren   sql.NullString `bun:"dokumani_veren" json:"dokumani_veren"`
	EndDateLeaveJob sql.NullInt64  `bun:"end_date_leave_job" json:"end_date_leave_job"`

	PersonelThyCalisiyorMu bool   `bun:"personel_thy_calisiyor_mu" json:"personel_thy_calisiyor_mu"`
	DokumanGecerliMi       bool   `bun:"dokuman_gecerli_mi" json:"dokuman_gecerli_mi"`
	AgreementType          string `bun:"agreement_type" json:"agreement_type"`
}
