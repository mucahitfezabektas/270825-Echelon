// middleware/auth.go
package middleware

import (
	"fmt"
	"os"
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v5"
)

// JWT gizli anahtarı
var jwtSecret string

func init() {
	jwtSecret = os.Getenv("JWT_SECRET")
	if jwtSecret == "" {
		fmt.Println("UYARI: Middleware için JWT_SECRET ortam değişkeni ayarlanmadı! Varsayılan değer kullanılıyor.")
		jwtSecret = "super-secret-jwt-key-please-change-me"
	}
}

// JWTMiddleware, her korumalı isteği doğrular.
func JWTMiddleware() fiber.Handler {
	return func(c *fiber.Ctx) error {
		authHeader := c.Get("Authorization")
		if authHeader == "" {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Yetkilendirme başlığı eksik"})
		}

		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Geçersiz yetkilendirme şeması"})
		}

		tokenString := parts[1]

		token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("beklenmeyen imza metodu: %v", token.Header["alg"])
			}
			return []byte(jwtSecret), nil
		})

		if err != nil {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Geçersiz token", "details": err.Error()})
		}

		if !token.Valid {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Token geçerli değil"})
		}

		claims, ok := token.Claims.(jwt.MapClaims)
		if !ok {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Token iddiaları okunamadı"})
		}

		// Kullanıcı ID'sini Fiber Context'e ekle, böylece diğer handler'lar erişebilir
		userID, ok := claims["user_id"].(float64) // JWT claim'leri float64 olarak parse edilebilir
		if !ok {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Kullanıcı ID'si token'da bulunamadı veya geçersiz format"})
		}
		c.Locals("userID", int64(userID)) // int64'e dönüştürerek sakla

		return c.Next() // Bir sonraki middleware veya handler'a geç
	}
}

// GetUserIDFromContext, middleware tarafından ayarlanmış userID'yi alır.
func GetUserIDFromContext(c *fiber.Ctx) (int64, error) {
	userID, ok := c.Locals("userID").(int64)
	if !ok {
		return 0, fiber.NewError(fiber.StatusInternalServerError, "UserID context'te bulunamadı")
	}
	return userID, nil
}
