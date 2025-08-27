package services

import (
	"context"
	"log"
	"mini_CMS_Desktop_App/repositories"
	"strings"
)

// BriefDebriefCalculator hesaplayıcı yapı
type BriefDebriefCalculator struct {
	ruleRepo *repositories.BriefDebriefRuleRepository
}

// Yeni BriefDebriefCalculator oluşturur
func NewBriefDebriefCalculator(ruleRepo *repositories.BriefDebriefRuleRepository) *BriefDebriefCalculator {
	return &BriefDebriefCalculator{ruleRepo: ruleRepo}
}

// Brief ve debrief sürelerini hesaplar
func (c *BriefDebriefCalculator) GetBriefDebriefDurations(
	ctx context.Context,
	crewType string,
	dutyType string,
	aircraftType string,
	dutyStartAirport string,
) (briefMin, debriefMin int) {

	// Kuralları önceden sıralı şekilde getir
	rules, err := c.ruleRepo.GetAllRules(ctx)
	if err != nil {
		log.Printf("[BriefDebriefCalc] ❗ Kural yükleme hatası: %v — Varsayılan değerler kullanılacak.", err)
		return 60, 30
	}

	// Kuralları sırayla değerlendir (öncelik yüksek → düşük)
	for _, rule := range rules {
		if !matches(rule.CrewType, crewType) {
			continue
		}
		if !matches(rule.ScenarioType, dutyType) {
			continue
		}
		if !airportMatches(rule.DutyStartAirport, dutyStartAirport) {
			continue
		}
		if !aircraftMatches(rule.AircraftType, aircraftType) {
			continue
		}

		// Eşleşme bulundu
		log.Printf("[BriefDebriefCalc] ✅ Kural eşleşti → ID:%d, Brief:%d dk, Debrief:%d dk", rule.DataID, rule.BriefDurationMin, rule.DebriefDurationMin)
		return rule.BriefDurationMin, rule.DebriefDurationMin
	}

	log.Printf("[BriefDebriefCalc] ⚠️ Hiçbir kural eşleşmedi. Varsayılan değerler uygulanacak.")
	return 60, 30
}

// Metin karşılaştırmalarında eşleşme kontrolü
func matches(ruleVal, inputVal string) bool {
	return strings.EqualFold(ruleVal, inputVal) || strings.EqualFold(ruleVal, "Hepsi")
}

// Meydan eşleşme kontrolü (Diğer / Hepsi dahil)
func airportMatches(ruleAirport, inputAirport string) bool {
	return strings.EqualFold(ruleAirport, inputAirport) ||
		strings.EqualFold(ruleAirport, "Hepsi") ||
		strings.EqualFold(ruleAirport, "Diğer")
}

// Uçak tipi eşleşme kontrolü (bilinmiyor durumu dahil)
func aircraftMatches(ruleType, inputType string) bool {
	return strings.EqualFold(ruleType, inputType) ||
		strings.EqualFold(ruleType, "Hepsi") ||
		(strings.EqualFold(ruleType, "Uçuş Ekibi") && strings.EqualFold(inputType, "BİLİNMİYOR"))
}
