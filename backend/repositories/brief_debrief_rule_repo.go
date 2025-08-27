package repositories

import (
	"context"
	"fmt"

	"mini_CMS_Desktop_App/models"

	"github.com/uptrace/bun"
)

type BriefDebriefRuleRepository struct {
	db *bun.DB
}

func NewBriefDebriefRuleRepository(db *bun.DB) *BriefDebriefRuleRepository {
	return &BriefDebriefRuleRepository{db: db}
}

// 🔹 Tüm kuralları getirir (öncelik ve data_id sırasına göre)
func (r *BriefDebriefRuleRepository) GetAllRules(ctx context.Context) ([]models.BriefDebriefRule, error) {
	var rules []models.BriefDebriefRule
	err := r.db.NewSelect().
		Model(&rules).
		OrderExpr("priority DESC, data_id ASC").
		Scan(ctx)
	if err != nil {
		return nil, fmt.Errorf("📛 brief/debrief kuralları alınamadı: %w", err)
	}
	return rules, nil
}

// 🔹 Filtreli eşleşen kuralları getirir (isteğe bağlı kullanım)
func (r *BriefDebriefRuleRepository) FindMatchingRules(
	ctx context.Context,
	scenarioType, aircraftType, crewType, airport string,
) ([]models.BriefDebriefRule, error) {

	var rules []models.BriefDebriefRule
	err := r.db.NewSelect().
		Model(&rules).
		Where("scenario_type = ? OR scenario_type = 'Hepsi'", scenarioType).
		Where("aircraft_type = ? OR aircraft_type IN ('Hepsi', 'BİLİNMİYOR')", aircraftType).
		Where("crew_type = ? OR crew_type = 'Hepsi'", crewType).
		Where("duty_start_airport = ? OR duty_start_airport IN ('Hepsi', 'Diğer')", airport).
		OrderExpr("priority DESC, data_id ASC").
		Scan(ctx)

	if err != nil {
		return nil, fmt.Errorf("📛 brief/debrief eşleşme sorgusu başarısız: %w", err)
	}
	return rules, nil
}
