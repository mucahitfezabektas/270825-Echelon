package crew_info

import (
	"context"
	"log"

	"mini_CMS_Desktop_App/db"
	"mini_CMS_Desktop_App/models"

	"github.com/gofiber/fiber/v2"
)

// ListCrewInfo, crew_info tablosundaki tüm kayıtları listeler.
// Fiber context'i (`*fiber.Ctx`) alır.
func ListCrewInfo(c *fiber.Ctx) error {
	log.Println("🔍 ListCrewInfo çağrıldı")

	var crewInfoEntries []models.CrewInfo

	// Tüm CrewInfo kayıtlarını veritabanından çek
	err := db.DB.NewSelect().
		Model(&crewInfoEntries).
		Order("person_id ASC"). // Kayıtları person_id'ye göre sırala (isteğe bağlı)
		Scan(context.Background())

	if err != nil {
		log.Printf("❌ CrewInfo kayıtları getirilirken hata oluştu: %v", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Ekip Bilgileri kayıtları getirilirken hata oluştu."})
	}

	if len(crewInfoEntries) == 0 {
		log.Println("ℹ️ Hiç CrewInfo kaydı bulunamadı.")
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"message": "Hiç ekip bilgisi kaydı bulunamadı."})
	}

	log.Printf("✅ %d adet CrewInfo kaydı başarıyla getirildi.", len(crewInfoEntries))
	return c.Status(fiber.StatusOK).JSON(crewInfoEntries)
}
