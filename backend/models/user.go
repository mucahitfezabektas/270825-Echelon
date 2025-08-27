// models/user.go
package models

import (
	"time"

	"github.com/uptrace/bun" // Bun ORM için
)

// User, veritabanındaki kullanıcı bilgilerini temsil eder.
type User struct {
	bun.BaseModel `bun:"table:users"` // 'users' adında bir tabloya eşlenecek

	ID        int64     `bun:"id,pk,autoincrement" json:"id"`                      // Primary key, otomatik artan
	Username  string    `bun:"username,unique,notnull" json:"username"`            // Kullanıcı adı, benzersiz ve boş olamaz
	Password  string    `bun:"password_hash,notnull" json:"-"`                     // Hash'lenmiş şifre. JSON'dan hariç tutulur.
	Email     string    `bun:"email,unique" json:"email"`                          // Opsiyonel: E-posta adresi
	CreatedAt time.Time `bun:"created_at,notnull,default:now()" json:"created_at"` // Kayıt tarihi
	UpdatedAt time.Time `bun:"updated_at,notnull,default:now()" json:"updated_at"` // Güncelleme tarihi
}

// UserCreateRequest, yeni kullanıcı kaydı için gelen isteği temsil eder.
type UserCreateRequest struct {
	Username string `json:"username" validate:"required,min=3,max=30"` // Doğrulama kuralları eklenebilir
	Password string `json:"password" validate:"required,min=6,max=50"`
	Email    string `json:"email" validate:"omitempty,email"`
}

// LoginRequest, kullanıcı giriş isteğini temsil eder.
type LoginRequest struct {
	Username string `json:"username" validate:"required"`
	Password string `json:"password" validate:"required"`
}

// LoginResponse, başarılı giriş sonrası dönen yanıtı temsil eder.
type LoginResponse struct {
	Token    string `json:"token"`
	Username string `json:"username"`
	UserID   int64  `json:"user_id"`
}
