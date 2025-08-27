package off_day_table

import (
	"context"
	"log"

	"mini_CMS_Desktop_App/db"
	"mini_CMS_Desktop_App/models" // models paketini import ettiğinizden emin olun

	"github.com/gofiber/fiber/v2"
)

// ListOffDayTable, off_day_table tablosundaki tüm kayıtları listeler.
func ListOffDayTable(c *fiber.Ctx) error {
	log.Println("🔍 ListOffDayTable çağrıldı")

	var offDayTable []models.OffDayTable
	// Tüm kayıtları veritabanından çekiyoruz
	err := db.DB.NewSelect().
		Model(&offDayTable).
		Order("work_days ASC"). // work_days'e göre sırala
		Scan(context.Background())

	if err != nil {
		log.Printf("❌ Off Day Tablosu listeleme hatası: %v", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error":   "Off Day Tablosu listelenemedi",
			"details": err.Error(),
		})
	}

	log.Printf("✅ %d adet Off Day Tablosu kaydı bulundu.", len(offDayTable))
	return c.JSON(offDayTable)
}
