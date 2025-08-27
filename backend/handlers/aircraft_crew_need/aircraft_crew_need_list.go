package aircraft_crew_need

import (
	"context"
	"log"

	"mini_CMS_Desktop_App/db"
	"mini_CMS_Desktop_App/models"

	"github.com/gofiber/fiber/v2"
)

// ListAircraftCrewNeed, aircraft_crew_need tablosundaki tüm kayıtları listeler.
// Fiber context'i (`*fiber.Ctx`) alır.
func ListAircraftCrewNeed(c *fiber.Ctx) error {
	log.Println("🔍 ListAircraftCrewNeed çağrıldı")

	var aircraftCrewNeedEntries []models.AircraftCrewNeed

	// Tüm AircraftCrewNeed kayıtlarını veritabanından çek
	err := db.DB.NewSelect().
		Model(&aircraftCrewNeedEntries).
		Order("actype ASC"). // Kayıtları actype'a göre sırala (isteğe bağlı)
		Scan(context.Background())

	if err != nil {
		log.Printf("❌ Aircraft Crew Need kayıtları getirilirken hata oluştu: %v", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Uçak Ekip İhtiyacı kayıtları getirilirken hata oluştu."})
	}

	if len(aircraftCrewNeedEntries) == 0 {
		log.Println("ℹ️ Hiç Aircraft Crew Need kaydı bulunamadı.")
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"message": "Hiç uçak ekip ihtiyacı kaydı bulunamadı."})
	}

	log.Printf("✅ %d adet Aircraft Crew Need kaydı başarıyla getirildi.", len(aircraftCrewNeedEntries))
	return c.Status(fiber.StatusOK).JSON(aircraftCrewNeedEntries)
}
