// backend\models\aircraft_crew_need.go
package models

import (
	"github.com/google/uuid"
	"github.com/uptrace/bun"
)

// AircraftCrewNeed veritabanındaki "aircraft_crew_need" tablosunu temsil eder.
type AircraftCrewNeed struct {
	bun.BaseModel `bun:"table:aircraft_crew_need"` // Tablo adını açıkça belirtiyoruz

	// Her tablonun DataID'si birincil anahtar olacak
	DataID uuid.UUID `bun:"data_id,pk,type:uuid,default:gen_random_uuid()" json:"data_id"`

	Actype string `bun:"actype,unique" json:"actype"` // Uçak Tipi, benzersiz olmalı

	// Mürettebat Pozisyonu Sayıları: (Örnek başlıklarınıza göre)
	C_Count  int32 `bun:"c_count" json:"c_count"`   // C pozisyonu için sayı
	P_Count  int32 `bun:"p_count" json:"p_count"`   // P pozisyonu için sayı
	J_Count  int32 `bun:"j_count" json:"j_count"`   // J pozisyonu için sayı
	EF_Count int32 `bun:"ef_count" json:"ef_count"` // EF pozisyonu için sayı
	A_Count  int32 `bun:"a_count" json:"a_count"`   // A pozisyonu için sayı
	S_Count  int32 `bun:"s_count" json:"s_count"`   // S pozisyonu için sayı
	L_Count  int32 `bun:"l_count" json:"l_count"`   // L pozisyonu için sayı
	EC_Count int32 `bun:"ec_count" json:"ec_count"` // EC pozisyonu için sayı
	T_Count  int32 `bun:"t_count" json:"t_count"`   // T pozisyonu için sayı
}
