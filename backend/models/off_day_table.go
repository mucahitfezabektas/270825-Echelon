package models

import (
	"github.com/google/uuid"
	"github.com/uptrace/bun"
)

// OffDayTable veritabanındaki "off_day_table" tablosunu temsil eder.
type OffDayTable struct {
	bun.BaseModel `bun:"table:off_day_table"` // Tablo adını açıkça belirtiyoruz

	// Her tablonun DataID'si birincil anahtar olacak
	DataID uuid.UUID `bun:"data_id,pk,type:uuid,default:gen_random_uuid()" json:"data_id"`

	// Off Day Table Sütunları:
	WorkDays          int32  `bun:"work_days" json:"work_days"`                     // INTEGER -> int32/int64
	OffDayEntitlement int32  `bun:"off_day_entitlement" json:"off_day_entitlement"` // INTEGER -> int32/int64
	Distribution      string `bun:"distribution" json:"distribution"`               // TEXT -> string
}
