// C:\Users\mucah_wi2yyc2\Desktop\mini_CMS_Desktop_App\backend\repositories\user_preference_repo.go

package repositories

import (
	"context"
	"database/sql"
	"fmt"

	"mini_CMS_Desktop_App/models" // <<<< BURADA DÜZELTİLDİ: models paketi import yolu

	"github.com/uptrace/bun"
)

type UserPreferenceRepository struct {
	db *bun.DB
}

func NewUserPreferenceRepository(db *bun.DB) *UserPreferenceRepository {
	return &UserPreferenceRepository{db: db}
}

func (r *UserPreferenceRepository) GetPreferenceByUserID(userID string) (*models.UserPreference, error) {
	var pref models.UserPreference
	err := r.db.NewSelect().
		Model(&pref).
		Where("user_id = ?", userID).
		Scan(context.Background())

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil // Tercih bulunamadı
		}
		return nil, fmt.Errorf("kullanıcı tercihi çekilirken hata: %w", err)
	}
	return &pref, nil
}

func (r *UserPreferenceRepository) SavePreference(pref *models.UserPreference) error {
	_, err := r.db.NewInsert().
		Model(pref).
		On("CONFLICT (user_id) DO UPDATE").
		Set("time_zone = EXCLUDED.time_zone").
		Set("updated_at = NOW()").
		Exec(context.Background())

	if err != nil {
		return fmt.Errorf("kullanıcı tercihi kaydedilirken/güncellenirken hata: %w", err)
	}
	return nil
}
