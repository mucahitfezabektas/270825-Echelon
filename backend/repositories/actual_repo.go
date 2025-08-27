package repositories

import (
	"context"
	"fmt"
	"time"

	"mini_CMS_Desktop_App/models"

	"github.com/uptrace/bun"
)

// MinimalActual → frontend için hafif yapı
type MinimalActual struct {
	PersonID     string    `json:"person_id"`
	DutyStart    time.Time `json:"duty_start"`
	DutyEnd      time.Time `json:"duty_end"`
	TripID       string    `json:"trip_id"`
	ActivityCode string    `json:"activity_code"`
	FlightNo     string    `json:"flight_no"`
}

type ActualRepository struct {
	db *bun.DB
}

func NewActualRepository(db *bun.DB) *ActualRepository {
	return &ActualRepository{db: db}
}

// 🔹 1. Belirli person_id için duty_start >= X
func (r *ActualRepository) GetActualsByPersonID(ctx context.Context, personID string, from time.Time, limit int) ([]models.Actual, error) {
	var actuals []models.Actual

	query := r.db.NewSelect().
		Model(&actuals).
		Where("person_id = ?", personID).
		Where("duty_start >= ?", from).
		Order("duty_start ASC")

	if limit > 0 {
		query = query.Limit(limit)
	}

	if err := query.Scan(ctx); err != nil {
		return nil, fmt.Errorf("actuals alınamadı (person_id=%s): %w", personID, err)
	}
	return actuals, nil
}

// 🔹 2. Daha hafif versiyon (sadece seçili kolonlar)
func (r *ActualRepository) GetMinimalActualsByPersonID(ctx context.Context, personID string, from time.Time, limit int) ([]MinimalActual, error) {
	var rows []MinimalActual

	query := r.db.NewSelect().
		Model((*models.Actual)(nil)).
		Column("person_id", "duty_start", "duty_end", "trip_id", "activity_code", "flight_no").
		Where("person_id = ?", personID).
		Where("duty_start >= ?", from).
		Order("duty_start ASC")

	if limit > 0 {
		query = query.Limit(limit)
	}

	if err := query.Scan(ctx, &rows); err != nil {
		return nil, fmt.Errorf("minimal actuals alınamadı (person_id=%s): %w", personID, err)
	}
	return rows, nil
}

// 🔹 3. Trip ID üzerinden actualları al
func (r *ActualRepository) GetActualsByTripID(ctx context.Context, tripID string) ([]models.Actual, error) {
	var actuals []models.Actual

	err := r.db.NewSelect().
		Model(&actuals).
		Where("trip_id = ?", tripID).
		Order("duty_start ASC").
		Scan(ctx)

	if err != nil {
		return nil, fmt.Errorf("trip_id=%s için actuals alınamadı: %w", tripID, err)
	}
	return actuals, nil
}

// 🔹 4. Flight No üzerinden actualları al
func (r *ActualRepository) GetActualsByFlightNo(ctx context.Context, flightNo string) ([]models.Actual, error) {
	var actuals []models.Actual

	err := r.db.NewSelect().
		Model(&actuals).
		Where("flight_no = ?", flightNo).
		Order("departure_time ASC").
		Scan(ctx)

	if err != nil {
		return nil, fmt.Errorf("flight_no=%s için actuals alınamadı: %w", flightNo, err)
	}
	return actuals, nil
}

// 🔹 5. Genel arama (örnek: activity_code, base_filo, port filtreleri)
func (r *ActualRepository) SearchActuals(ctx context.Context, filters map[string]interface{}, limit int) ([]models.Actual, error) {
	var actuals []models.Actual

	query := r.db.NewSelect().Model(&actuals)

	for key, val := range filters {
		query = query.Where(fmt.Sprintf("%s = ?", key), val)
	}

	query = query.Order("duty_start ASC")

	if limit > 0 {
		query = query.Limit(limit)
	}

	if err := query.Scan(ctx); err != nil {
		return nil, fmt.Errorf("actual arama hatası: %w", err)
	}
	return actuals, nil
}
