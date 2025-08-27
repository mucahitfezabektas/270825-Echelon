// C:\Users\mucah_wi2yyc2\Desktop\mini_CMS_Desktop_App\backend\handlers\user_preference_handler.go

package user_preference

import (
	"log"

	"mini_CMS_Desktop_App/models"
	"mini_CMS_Desktop_App/repositories"

	"github.com/gofiber/fiber/v2"
)

// UserPreferenceHandler, kullanıcı tercihleri ile ilgili API isteklerini yönetir.
type UserPreferenceHandler struct {
	repo *repositories.UserPreferenceRepository
}

// NewUserPreferenceHandler, UserPreferenceHandler'ın yeni bir örneğini oluşturur.
func NewUserPreferenceHandler(repo *repositories.UserPreferenceRepository) *UserPreferenceHandler {
	return &UserPreferenceHandler{repo: repo}
}

// SetUserPreference, bir kullanıcının tercihini kaydeder veya günceller.
func (h *UserPreferenceHandler) SetUserPreference(c *fiber.Ctx) error {
	var pref models.UserPreference
	if err := c.BodyParser(&pref); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Geçersiz istek gövdesi", "details": err.Error()})
	}

	// Kullanıcı ID'sini Fiber Context'ten almalısınız (Auth middleware'i sonrası)
	// Şimdilik varsayılan bir kullanıcı ID'si veya request body'den alalım.
	// pref.UserID = c.Locals("userID").(string) // Eğer bir Auth middleware'iniz varsa
	if pref.UserID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "UserID boş olamaz"})
	}

	if err := h.repo.SavePreference(&pref); err != nil {
		log.Printf("Hata: Kullanıcı tercihi kaydedilirken sorun: %v", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Kullanıcı tercihi kaydedilemedi", "details": err.Error()})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{"message": "Kullanıcı tercihi başarıyla kaydedildi"})
}

// GetUserPreference, bir kullanıcının tercihini çeker.
func (h *UserPreferenceHandler) GetUserPreference(c *fiber.Ctx) error {
	userID := c.Query("user_id") // Query parametresinden kullanıcı ID'si al
	if userID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "user_id parametresi gerekli"})
	}

	pref, err := h.repo.GetPreferenceByUserID(userID)
	if err != nil {
		log.Printf("Hata: Kullanıcı tercihi çekilirken sorun: %v", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Kullanıcı tercihi çekilemedi", "details": err.Error()})
	}
	if pref == nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"message": "Kullanıcı tercihi bulunamadı"})
	}

	return c.Status(fiber.StatusOK).JSON(pref)
}
