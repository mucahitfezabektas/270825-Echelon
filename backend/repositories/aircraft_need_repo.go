package repositories

import (
	"context"
	"database/sql"
	"fmt"
)

type AircraftNeedRepo struct {
	DB *sql.DB
}

func NewAircraftNeedRepo(db *sql.DB) *AircraftNeedRepo {
	return &AircraftNeedRepo{DB: db}
}

// İlgili aircraft_type için ihtiyaç string'ini getir
func (r *AircraftNeedRepo) GetNeedsByAircraftType(ctx context.Context, aircraftType string) (string, error) {
	query := `
		SELECT needs
		FROM aircraft_crew_need
		WHERE aircraft_type = $1
		LIMIT 1
	`
	var needs string
	err := r.DB.QueryRowContext(ctx, query, aircraftType).Scan(&needs)
	if err != nil {
		return "", fmt.Errorf("GetNeedsByAircraftType query error: %w", err)
	}
	return needs, nil
}
