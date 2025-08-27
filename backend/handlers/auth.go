// handlers/auth.go
package handlers

import (
	"context"
	"database/sql" // ✅ Eklendi: sql.ErrNoRows için
	"fmt"
	"mini_CMS_Desktop_App/db"
	"mini_CMS_Desktop_App/models"
	"os" // Ortam değişkenleri için
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v5" // JWT kütüphanesi
	"golang.org/x/crypto/bcrypt"   // Şifre hash'leme
)

// JWT gizli anahtarı
var jwtSecret string

func init() {
	jwtSecret = os.Getenv("JWT_SECRET")
	if jwtSecret == "" {
		// Üretim ortamında bu bir hata olmalı, geliştirme için varsayılan değer
		fmt.Println("UYARI: JWT_SECRET ortam değişkeni ayarlanmadı! Varsayılan değer kullanılıyor. Bu üretim için GÜVENLİ DEĞİLDİR.")
		jwtSecret = "super-secret-jwt-key-please-change-me"
	}
}

// generateJWTToken, verilen kullanıcı ID'si için bir JWT oluşturur.
func generateJWTToken(userID int64) (string, error) {
	claims := jwt.MapClaims{
		"user_id": userID,
		"exp":     time.Now().Add(time.Hour * 24).Unix(), // Token 24 saat geçerli
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte(jwtSecret))
	if err != nil {
		return "", fmt.Errorf("token imzalanamadı: %w", err)
	}
	return tokenString, nil
}

// RegisterUserHandler yeni bir kullanıcı kaydeder.
func RegisterUserHandler(c *fiber.Ctx) error {
	var req models.UserCreateRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Geçersiz istek gövdesi", "details": err.Error()})
	}

	// Basit istemci tarafı doğrulama (daha kapsamlı kütüphane kullanılabilir)
	if req.Username == "" || req.Password == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Kullanıcı adı ve şifre boş olamaz"})
	}
	if len(req.Password) < 6 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Şifre en az 6 karakter olmalı"})
	}

	// Şifreyi hash'le
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Şifre hash'leme hatası", "details": err.Error()})
	}

	user := &models.User{
		Username: req.Username,
		Password: string(hashedPassword), // Hash'lenmiş şifreyi kaydet
		Email:    req.Email,
	}

	// Kullanıcı adının zaten var olup olmadığını kontrol et
	existingUser := new(models.User)
	err = db.DB.NewSelect().Model(existingUser).Where("username = ?", user.Username).Scan(context.Background())
	if err == nil { // Kullanıcı bulundu
		return c.Status(fiber.StatusConflict).JSON(fiber.Map{"error": "Bu kullanıcı adı zaten alınmış"})
	}
	// ✅ Değişiklik: bun.ErrNoRows yerine sql.ErrNoRows kullanıldı
	if err != sql.ErrNoRows { // Başka bir veritabanı hatası
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Veritabanı kontrol hatası", "details": err.Error()})
	}

	// Kullanıcıyı veritabanına ekle
	_, err = db.DB.NewInsert().Model(user).Exec(context.Background())
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Kullanıcı kaydedilemedi", "details": err.Error()})
	}

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{"message": "Kullanıcı başarıyla kaydedildi", "user_id": user.ID, "username": user.Username})
}

// LoginHandler kullanıcı girişi yapar ve JWT döndürür.
func LoginHandler(c *fiber.Ctx) error {
	var req models.LoginRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Geçersiz istek gövdesi", "details": err.Error()})
	}

	user := new(models.User)
	err := db.DB.NewSelect().Model(user).Where("username = ?", req.Username).Scan(context.Background())
	if err != nil {
		// ✅ Değişiklik: bun.ErrNoRows yerine sql.ErrNoRows kullanıldı
		if err == sql.ErrNoRows {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Kullanıcı adı veya şifre yanlış"})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Veritabanı sorgu hatası", "details": err.Error()})
	}

	// Hash'lenmiş şifreyi doğrula
	err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password))
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Kullanıcı adı veya şifre yanlış"})
	}

	// JWT oluştur
	token, err := generateJWTToken(user.ID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Token oluşturulamadı", "details": err.Error()})
	}

	return c.JSON(models.LoginResponse{
		Token:    token,
		Username: user.Username,
		UserID:   user.ID,
	})
}
