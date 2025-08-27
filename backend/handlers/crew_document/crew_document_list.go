package crew_document

import (
	"context"
	"log"
	"strconv"

	"mini_CMS_Desktop_App/db"
	"mini_CMS_Desktop_App/models"

	"github.com/gofiber/fiber/v2"
)

// ListCrewDocuments provides paginated list of all crew documents without search/sort.
// This is primarily for infinite scrolling where initial full dataset is pulled.
func ListCrewDocuments(c *fiber.Ctx) error {
	log.Println("🔍 ListCrewDocuments çağrıldı (sadece sayfalı)")

	pageStr := c.Query("page", "1")
	// Sayfa başına 5000 kayıt sınırı
	sizeStr := c.Query("size", "5000")

	page, err := strconv.Atoi(pageStr)
	if err != nil || page < 1 {
		log.Printf("❌ Geçersiz sayfa numarası: %s", pageStr)
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Geçersiz sayfa numarası."})
	}

	size, err := strconv.Atoi(sizeStr)
	// Max boyutu 5000 yapmak (frontend 5000 gönderirken backend'in 100.000 dönmesi anlamsız)
	if err != nil || size < 1 || size > 10000 { // Max 5000 kayıt döndürelim
		log.Printf("❌ Geçersiz sayfa boyutu: %s. Varsayılan 5000 kullanılıyor.", sizeStr)
		size = 5000 // Geçersiz veya çok büyük boyut gelirse varsayılana çek
	}
	log.Printf("DEBUG: ListCrewDocuments'ta sayfa boyutu (size) olarak alınan değer: %d", size) // Debug logu

	offset := (page - 1) * size

	var crewDocuments []models.CrewDocument

	totalCount, err := db.DB.NewSelect().
		Model(&crewDocuments).
		Count(context.Background())
	if err != nil {
		log.Printf("❌ Ekip dokümanları toplam sayısı alınamadı: %v", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error":   "Toplam ekip dokümanı sayısı alınamadı",
			"details": err.Error(),
		})
	}

	err = db.DB.NewSelect().
		Model(&crewDocuments).
		Order("person_id ASC"). // Varsayılan bir sıralama
		Limit(size).            // Frontend'den gelen 'size' değeri kullanılır
		Offset(offset).
		Scan(context.Background())

	if err != nil {
		log.Printf("❌ Ekip dokümanları listeleme hatası: %v", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error":   "Ekip dokümanları listelenemedi",
			"details": err.Error(),
		})
	}

	log.Printf("✅ Sayfa %d için %d adet ekip dokümanı bulundu (Toplam: %d).", page, len(crewDocuments), totalCount)

	return c.JSON(fiber.Map{
		"data":       crewDocuments,
		"page":       page,
		"size":       size,
		"totalCount": totalCount,
		"totalPages": (totalCount + size - 1) / size,
	})
}
