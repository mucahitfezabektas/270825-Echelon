package activity_code // activity_code_handler.go ile aynı paketteyiz

import (
	"context"
	"log"

	"mini_CMS_Desktop_App/db"
	"mini_CMS_Desktop_App/models"

	"github.com/gofiber/fiber/v2"
)

// ListActivityCodes, activity_codes tablosundaki tüm kayıtları listeler.
func ListActivityCodes(c *fiber.Ctx) error {
	log.Println("🔍 ListActivityCodes çağrıldı")

	var activityCodes []models.ActivityCode
	// Tüm kayıtları veritabanından çekiyoruz
	err := db.DB.NewSelect().
		Model(&activityCodes).
		Order("activity_code ASC"). // activity_code'a göre sırala
		Scan(context.Background())

	if err != nil {
		log.Printf("❌ Activity codes listeleme hatası: %v", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error":   "Aktivite kodları listelenemedi",
			"details": err.Error(),
		})
	}

	log.Printf("✅ %d adet aktivite kodu bulundu.", len(activityCodes))
	return c.JSON(activityCodes)
}
