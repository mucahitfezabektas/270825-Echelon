// C:\Users\mucah_wi2yyc2\Desktop\mini_CMS_Desktop_App\backend\models\brief_debrief_rule.go

package models

import "github.com/uptrace/bun" // bun paketi eklendi

// brief ve debrief süreleri için tanımlanan kuralları temsil eder.
type BriefDebriefRule struct {
	bun.BaseModel `bun:"brief_debrief_rules"` // `brief_debrief_rules` tablosu ile ilişkilendirildi

	DataID             int    `json:"data_id" bun:"data_id,pk,autoincrement"`          // Otomatik artan ID
	ScenarioType       string `json:"scenario_type" bun:"scenario_type"`               // "İlk sektörü görevli, yolculu uçuşlar", "Simülatör" vb.
	AircraftType       string `json:"aircraft_type" bun:"aircraft_type"`               // "DAR GÖVDE", "GENİŞ GÖVDE", "Hepsi", "Uçuş Ekibi"
	CrewType           string `json:"crew_type" bun:"crew_type"`                       // "Uçuş Ekibi", "Kabin Ekibi", "Kargo Uçuş Ekibi"
	DutyStartAirport   string `json:"duty_start_airport" bun:"duty_start_airport"`     // "IST", "ISL", "SAW", "Diğer", "Hepsi"
	BriefDurationMin   int    `json:"brief_duration_min" bun:"brief_duration_min"`     // Briefing süresi dakika cinsinden
	DebriefDurationMin int    `json:"debrief_duration_min" bun:"debrief_duration_min"` // Debriefing süresi dakika cinsinden
	Priority           int    `json:"priority" bun:"priority,default:0"`               // Kural eşleşmesi için öncelik (yüksek sayı = yüksek öncelik)
}

// TableName, bun ORM'in bu struct'ı 'brief_debrief_rules' tablosuyla eşleştirmesini sağlar.
func (BriefDebriefRule) TableName() string {
	return "brief_debrief_rules"
}
