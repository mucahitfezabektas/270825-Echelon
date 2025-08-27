package repositories

import (
	"context"
	"fmt"
	"time"

	// "time" // No longer needed for this specific method, but keep if other methods use it

	"mini_CMS_Desktop_App/models"

	"github.com/uptrace/bun"
)

// MinimalPublish → frontend için hafif yapı (Publish modeliyle aynı alanlara sahip)
type MinimalPublish struct {
	PersonID     string    `json:"person_id"`
	DutyStart    time.Time `json:"duty_start"`
	DutyEnd      time.Time `json:"duty_end"`
	TripID       string    `json:"trip_id"`
	ActivityCode string    `json:"activity_code"`
	FlightNo     string    `json:"flight_no"`
}

// PublishRepository, 'publishes' tablosu üzerinde CRUD ve sorgulama işlemleri sağlar.
type PublishRepository struct {
	DB *bun.DB // <--- Changed 'db' to 'DB' to make it exported
}

// NewPublishRepository, PublishRepository'nin yeni bir örneğini oluşturur.
func NewPublishRepository(db *bun.DB) *PublishRepository {
	return &PublishRepository{DB: db} // <--- Assign to DB
}

// 🔹 1. Belirli person_id için tüm Publish kayıtlarını alır.
// The 'from' and 'limit' parameters are removed as per the new requirement.
func (r *PublishRepository) GetPublishesByPersonID(ctx context.Context, personID string) ([]models.Publish, error) { // <--- Updated signature
	var publishes []models.Publish

	// Directly use r.DB (the exported field)
	query := r.DB.NewSelect(). // <--- Use r.DB
					Model(&publishes).
					Where("person_id = ?", personID).
					Order("duty_start ASC")

	// No limit or time range here, just the personID filter

	if err := query.Scan(ctx); err != nil {
		return nil, fmt.Errorf("publishes alınamadı (person_id=%s): %w", personID, err)
	}
	return publishes, nil
}

// 🔹 2. Daha hafif versiyon (sadece seçili kolonlar) - Frontend için MinimalPublish döner.
// Signature needs to be updated if GetPublishesByPersonID is changed.
// Let's assume this method also needs to be updated for "no limit or time range".
func (r *PublishRepository) GetMinimalPublishesByPersonID(ctx context.Context, personID string) ([]MinimalPublish, error) { // <--- Updated signature
	var rows []MinimalPublish

	query := r.DB.NewSelect(). // <--- Use r.DB
					Model((*models.Publish)(nil)).
					Column("person_id", "duty_start", "duty_end", "trip_id", "activity_code", "flight_no").
					Where("person_id = ?", personID).
					Order("duty_start ASC")

	if err := query.Scan(ctx, &rows); err != nil {
		return nil, fmt.Errorf("minimal publishes alınamadı (person_id=%s): %w", personID, err)
	}
	return rows, nil
}

// 🔹 3. Trip ID üzerinden Publish kayıtlarını alır.
func (r *PublishRepository) GetPublishesByTripID(ctx context.Context, tripID string) ([]models.Publish, error) {
	var publishes []models.Publish

	err := r.DB.NewSelect(). // <--- Use r.DB
					Model(&publishes).
					Where("trip_id = ?", tripID).
					Order("duty_start ASC").
					Scan(ctx)

	if err != nil {
		return nil, fmt.Errorf("trip_id=%s için publishes alınamadı: %w", tripID, err)
	}
	return publishes, nil
}

// 🔹 4. Flight No üzerinden Publish kayıtlarını alır.
func (r *PublishRepository) GetPublishesByFlightNo(ctx context.Context, flightNo string) ([]models.Publish, error) {
	var publishes []models.Publish

	err := r.DB.NewSelect(). // <--- Use r.DB
					Model(&publishes).
					Where("flight_no = ?", flightNo).
					Order("departure_time ASC").
					Scan(ctx)

	if err != nil {
		return nil, fmt.Errorf("flight_no=%s için publishes alınamadı: %w", flightNo, err)
	}
	return publishes, nil
}

// 🔹 5. Genel arama (örnek: activity_code, base_filo, port filtreleri)
// Dinamik filtreler kullanarak Publish kayıtlarını arar.
func (r *PublishRepository) SearchPublishes(ctx context.Context, filters map[string]interface{}, limit int) ([]models.Publish, error) {
	var publishes []models.Publish

	query := r.DB.NewSelect().Model(&publishes) // <--- Use r.DB

	for key, val := range filters {
		query = query.Where(fmt.Sprintf("%s = ?", key), val)
	}

	query = query.Order("duty_start ASC")

	if limit > 0 {
		query = query.Limit(limit)
	}

	if err := query.Scan(ctx); err != nil {
		return nil, fmt.Errorf("publish arama hatası: %w", err)
	}
	return publishes, nil
}
