// backend\handlers\actual_list.go
package handlers

import (
	"mini_CMS_Desktop_App/db"
	"mini_CMS_Desktop_App/models"
	"strings"
	"time" // time paketi eklendi

	"github.com/gofiber/fiber/v2"
)

func ListActualData(c *fiber.Ctx) error {
	var results []models.Actual

	month := c.Query("month")
	limit := c.QueryInt("limit", 500)
	if limit > 10000 {
		limit = 10000
	}
	page := c.QueryInt("page", 1)
	if page < 1 {
		page = 1
	}
	offset := (page - 1) * limit
	all := c.Query("all") == "true"

	query := db.DB.NewSelect().
		Model(&results).
		Order("departure_time ASC")

	month = strings.TrimSpace(month)

	countQuery := db.DB.NewSelect().Model((*models.Actual)(nil))
	if month != "" {
		query = query.Where("period_month = ?", month)
		countQuery = countQuery.Where("period_month = ?", month)
	}

	total, err := countQuery.Count(c.Context())
	if err != nil {
		return c.Status(500).JSON(fiber.Map{
			"error":   "Kayıt sayısı alınamadı",
			"details": err.Error(),
		})
	}

	if all {
		if total > 5000 {
			return c.Status(400).JSON(fiber.Map{
				"error": "Çok fazla kayıt. Lütfen sayfalama veya filtreleme kullanın.",
				"total": total,
			})
		}
		err = query.Scan(c.Context())
	} else {
		err = query.Limit(limit).Offset(offset).Scan(c.Context())
	}

	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(fiber.Map{
		"total":  total,
		"page":   page,
		"limit":  limit,
		"result": results,
	})
}

func PreviewActualData(c *fiber.Ctx) error {
	month := strings.TrimSpace(c.Query("month"))
	if month == "" {
		return c.Status(400).JSON(fiber.Map{"error": "month parametresi gerekli"})
	}

	var total int
	total, err := db.DB.NewSelect().
		Model((*models.Actual)(nil)).
		Where("period_month = ?", month).
		Count(c.Context())

	if err != nil {
		return c.Status(500).JSON(fiber.Map{
			"error":   "Toplam kayıt alınamadı",
			"details": err.Error(),
		})
	}

	var lastUploadedTime time.Time // `uploadedAt` yerine `lastUploadedTime` ve tipi `time.Time` olarak düzeltildi
	if total > 0 {
		err = db.DB.NewSelect().
			ColumnExpr("MAX(checkin_date)"). // MAX(checkin_date) zaten time.Time döndürür
			Model((*models.Actual)(nil)).
			Where("period_month = ?", month).
			Scan(c.Context(), &lastUploadedTime) // Doğrudan time.Time'a scan yapıyoruz

		if err != nil {
			return c.Status(500).JSON(fiber.Map{
				"error":   "Yüklenme tarihi alınamadı",
				"details": err.Error(),
			})
		}
	}

	return c.JSON(fiber.Map{
		"exists":      total > 0,
		"total":       total,
		"uploaded_at": lastUploadedTime, // time.Time objesini olduğu gibi döndürebiliriz, JSON marshalling halleder.
	})
}
