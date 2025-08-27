// C:\Users\mucah_wi2yyc2\Desktop\mini_CMS_Desktop_App\backend\models\user_preference.go

package models

import (
	"time"

	"github.com/uptrace/bun"
)

// UserPreference, kullanıcının uygulama genelindeki tercihlerini saklar.
// Şimdilik sadece zaman dilimi tercihini içeriyor.
type UserPreference struct {
	bun.BaseModel `bun:"user_preferences"`

	UserID    string    `json:"user_id" bun:"user_id,pk"`  // Tercihin ait olduğu kullanıcı ID'si
	TimeZone  string    `json:"time_zone" bun:"time_zone"` // Örneğin: "Europe/Istanbul", "UTC", "America/New_York"
	CreatedAt time.Time `json:"created_at" bun:"created_at,default:current_timestamp"`
	UpdatedAt time.Time `json:"updated_at" bun:"updated_at,default:current_timestamp"`
}

// TableName, bun ORM'in bu struct'ı 'user_preferences' tablosuyla eşleştirmesini sağlar.
func (UserPreference) TableName() string {
	return "user_preferences"
}
