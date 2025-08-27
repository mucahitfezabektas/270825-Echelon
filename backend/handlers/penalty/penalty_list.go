package penalty

import (
	"context"
	"log"

	"mini_CMS_Desktop_App/db"
	"mini_CMS_Desktop_App/models"

	"github.com/gofiber/fiber/v2"
)

// ListPenalties, penalties tablosundaki tüm kayıtları listeler.
// Fiber context'i (`*fiber.Ctx`) alır.
func ListPenalties(c *fiber.Ctx) error {
	log.Println("🔍 ListPenalties çağrıldı")

	var penalties []models.Penalty

	// Tüm Penalty kayıtlarını veritabanından çek
	err := db.DB.NewSelect().
		Model(&penalties).
		Order("person_id ASC"). // Kayıtları person_id'ye göre sırala (isteğe bağlı)
		Scan(context.Background())

	if err != nil {
		log.Printf("❌ Penalties kayıtları getirilirken hata oluştu: %v", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Ceza Bilgileri kayıtları getirilirken hata oluştu."})
	}

	if len(penalties) == 0 {
		log.Println("ℹ️ Hiç Penalty kaydı bulunamadı.")
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"message": "Hiç ceza bilgisi kaydı bulunamadı."})
	}

	log.Printf("✅ %d adet Penalty kaydı başarıyla getirildi.", len(penalties))
	return c.Status(fiber.StatusOK).JSON(penalties)
}
