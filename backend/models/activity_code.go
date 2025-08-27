package models

import (
	"github.com/google/uuid"
	"github.com/uptrace/bun"
)

type ActivityCode struct {
	bun.BaseModel `bun:"table:activity_codes"`

	DataID uuid.UUID `bun:"data_id,pk,type:uuid,default:gen_random_uuid()" json:"data_id"`

	ActivityCode            string `bun:"activity_code,unique" json:"activity_code"`
	ActivityGroupCode       string `bun:"activity_group_code" json:"activity_group_code"`
	ActivityCodeExplanation string `bun:"activity_code_explanation" json:"activity_code_explanation"`
}
