package repositories

import (
	"context"
	"mini_CMS_Desktop_App/models"

	"github.com/uptrace/bun"
)

type OpenTripRepo struct {
	DB *bun.DB
}

func NewOpenTripRepo(db *bun.DB) *OpenTripRepo {
	return &OpenTripRepo{DB: db}
}

// GetFLTActivities
// actual tablosundan activity_code = 'FLT' olan kayıtları çeker
func (r *OpenTripRepo) GetFLTActivities(ctx context.Context, period string) ([]models.Actual, error) {
	var items []models.Actual
	err := r.DB.NewSelect().
		Model(&items).
		Where("activity_code = ?", "FLT").
		Where("period_month = ?", period).
		Scan(ctx)
	return items, err
}

// GetNeedsByAircraftType
// aircraft_crew_need tablosundan ilgili aircraft_type için ihtiyaçları map olarak döner
func (r *OpenTripRepo) GetNeedsByAircraftType(ctx context.Context, aircraftType string) (map[string]int, error) {
	type needRow struct {
		C  int `bun:"c_count"`
		P  int `bun:"p_count"`
		J  int `bun:"j_count"`
		EF int `bun:"ef_count"`
		A  int `bun:"a_count"`
		S  int `bun:"s_count"`
		L  int `bun:"l_count"`
		EC int `bun:"ec_count"`
		T  int `bun:"t_count"`
	}

	var row needRow
	err := r.DB.NewSelect().
		Table("aircraft_crew_need").
		Column("c_count", "p_count", "j_count", "ef_count", "a_count", "s_count", "l_count", "ec_count", "t_count").
		Where("actype = ?", aircraftType).
		Limit(1).
		Scan(ctx, &row)
	if err != nil {
		return nil, err
	}

	return map[string]int{
		"C":  row.C,
		"P":  row.P,
		"J":  row.J,
		"EF": row.EF,
		"A":  row.A,
		"S":  row.S,
		"L":  row.L,
		"EC": row.EC,
		"T":  row.T,
	}, nil
}
