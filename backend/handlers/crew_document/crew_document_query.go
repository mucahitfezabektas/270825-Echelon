package crew_document

import (
	"context"
	"log"
	"strings" // strconv artık kullanılmadığı için kaldırılabilir, ama problem değil

	"mini_CMS_Desktop_App/db"
	"mini_CMS_Desktop_App/models"

	"github.com/gofiber/fiber/v2"
)

// QueryCrewDocuments handles searchable and sortable queries for crew documents,
// returning all matching results found based on the search criteria.
// Pagination parameters (page, size) are no longer used by this function.
func QueryCrewDocuments(c *fiber.Ctx) error {
	log.Println("🔍 QueryCrewDocuments çağrıldı (arama/sıralama destekli, tüm sonuçlar döndürülüyor)")

	// Sorgu Parametrelerini Al
	// 'page' ve 'size' parametreleri artık bu fonksiyon tarafından kullanılmıyor.
	searchQuery := c.Query("search", "")
	sortBy := c.Query("sortBy", "person_id")
	sortOrder := c.Query("sortOrder", "asc")

	var crewDocuments []models.CrewDocument
	var totalCount int

	query := db.DB.NewSelect().Model(&crewDocuments)

	// 🔍 Arama (filtreleme)
	if searchQuery != "" {
		s := "%" + strings.ToLower(searchQuery) + "%"
		query.Where(
			"LOWER(COALESCE(person_id, '')) LIKE ? OR LOWER(COALESCE(person_surname, '')) LIKE ? OR LOWER(COALESCE(person_name, '')) LIKE ? OR LOWER(COALESCE(dokuman_alt_tipi, '')) LIKE ? OR LOWER(COALESCE(document_no, '')) LIKE ?",
			s, s, s, s, s,
		)
	}

	// 🧮 Toplam kayıt sayısı
	// Arama filtresi uygulandıktan sonra toplam eşleşen kayıt sayısını bulur.
	totalCount, err := query.Count(context.Background())
	if err != nil {
		log.Printf("❌ Toplam sayım hatası: %v", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error":   "Toplam kayıt alınamadı.",
			"details": err.Error(),
		})
	}

	// ⬆️ Sıralama (güvenli kontrol)
	allowedSortCols := map[string]bool{
		"person_id": true, "person_surname": true, "person_name": true,
		"dokuman_alt_tipi": true, "gecerlilik_baslangic_tarihi": true,
		"document_no": true,
	}
	if !allowedSortCols[sortBy] {
		sortBy = "person_id"
	}
	if sortOrder != "asc" && sortOrder != "desc" {
		sortOrder = "asc"
	}

	// ✅ Güvenli sıralama (string birleştirme)
	query.OrderExpr(sortBy + " " + sortOrder)

	// NOT: Sayfalama (Limit ve Offset) tamamen kaldırıldı.
	// Bu fonksiyon, arama kriterlerine uyan tüm kayıtları döndürecektir.
	err = query.Scan(context.Background())

	if err != nil {
		log.Printf("❌ Veri listeleme hatası: %v", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error":   "Veri listelenemedi.",
			"details": err.Error(),
		})
	}

	log.Printf("✅ %d adet ekip dokümanı bulundu (arama: '%s')", len(crewDocuments), searchQuery)

	return c.JSON(fiber.Map{
		"data":       crewDocuments,
		"totalCount": totalCount,
		// 'page', 'size', 'totalPages' alanları sayfalama kaldırıldığı için artık döndürülmüyor.
		// Frontend'in bu alanları beklentisini güncellediğinden emin olun.
	})
}
