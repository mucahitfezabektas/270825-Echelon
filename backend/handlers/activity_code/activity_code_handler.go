package activity_code

import (
	"context"
	"encoding/csv" // CSV işlemleri için kalacak
	"fmt"
	"io"
	"log"
	"path/filepath" // Dosya uzantısını almak için eklendi
	"strings"

	"mini_CMS_Desktop_App/db"
	"mini_CMS_Desktop_App/models"

	"github.com/gofiber/fiber/v2"
	"github.com/xuri/excelize/v2" // ✅ Excelize kütüphanesi eklendi
)

// ImportActivityCodeData, activity_codes tablosuna hem CSV hem de XLSX verisi aktarır.
func ImportActivityCodeData(c *fiber.Ctx) error { // Fonksiyon adı güncellendi
	log.Println("🔍 ImportActivityCodeData çağrıldı")

	fileHeader, err := c.FormFile("file") // Frontend'den gelen form alanı adı hala "file" olmalı
	if err != nil {
		log.Printf("❌ Dosya alınamadı: %v", err)
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": fmt.Sprintf("Dosya alınamadı: %v", err)})
	}

	file, err := fileHeader.Open()
	if err != nil {
		log.Printf("❌ Dosya açılamadı: %v", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": fmt.Sprintf("Dosya açılamadı: %v", err)})
	}
	defer file.Close()

	var activityCodes []models.ActivityCode
	recordCount := 0
	lineNum := 0 // Başlık satırından sonraki satırları takip etmek için

	// Dosya uzantısına göre okuma stratejisi belirle
	fileExtension := strings.ToLower(filepath.Ext(fileHeader.Filename))

	switch fileExtension {
	case ".csv":
		log.Println("Handling CSV file...")
		reader := csv.NewReader(file)
		reader.Comma = ','
		reader.FieldsPerRecord = -1
		reader.LazyQuotes = true

		// Başlık satırını oku ve atla
		header, err := reader.Read()
		if err != nil {
			if err == io.EOF {
				log.Println("⚠️ CSV dosyası boş veya sadece başlık satırı içeriyor.")
				return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "CSV dosyası boş veya hiç veri satırı içermiyor."})
			}
			log.Printf("❌ CSV başlık satırı okunamadı: %v", err)
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": fmt.Sprintf("CSV başlık satırı okunamadı: %v", err)})
		}
		log.Printf("📌 CSV Header: %s\n", strings.Join(header, ","))

		for {
			lineNum++
			record, err := reader.Read()
			if err == io.EOF {
				break
			}
			if err != nil {
				log.Printf("❌ CSV satırı okuma hatası (Satır %d): %v\n", lineNum+1, err)
				return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": fmt.Sprintf("CSV satırı okuma hatası (Satır %d): %v", lineNum+1, err)})
			}

			// Boş satırları atla
			if len(record) == 0 || strings.Join(record, "") == "" {
				continue
			}

			if len(record) < 3 {
				log.Printf("⚠️ Satır %d atlandı: Yetersiz sütun sayısı (%d yerine 3 bekleniyor)\n", lineNum+1, len(record))
				continue
			}

			activityCode := models.ActivityCode{
				ActivityCode:            strings.TrimSpace(record[0]),
				ActivityGroupCode:       strings.TrimSpace(record[1]),
				ActivityCodeExplanation: strings.TrimSpace(record[2]),
			}
			activityCodes = append(activityCodes, activityCode)
			recordCount++
		}

	case ".xlsx":
		log.Println("Handling XLSX file...")
		// Excel dosyasını aç
		f, err := excelize.OpenReader(file)
		if err != nil {
			log.Printf("❌ Excel dosyası açılamadı: %v", err)
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": fmt.Sprintf("Excel dosyası açılamadı: %v", err)})
		}

		// ✅ YENİ: Excel dosyasındaki tüm sayfa adlarını al
		sheetList := f.GetSheetList()
		if len(sheetList) == 0 {
			log.Println("⚠️ Excel dosyasında hiç sayfa bulunamadı.")
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Excel dosyasında hiç sayfa bulunamadı."})
		}
		// ✅ İlk sayfayı seç
		sheetName := sheetList[0]
		log.Printf("ℹ️ Okunan Excel sayfası: %s\n", sheetName)

		rows, err := f.GetRows(sheetName) // ✅ İlk sayfadaki tüm satırları oku
		if err != nil {
			log.Printf("❌ Excel sayfasından satırlar okunamadı: %v", err)
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": fmt.Sprintf("Excel sayfasından veri okunamadı: %v", err)})
		}

		if len(rows) < 2 { // Başlık satırı + en az bir veri satırı
			log.Println("⚠️ Excel dosyası boş veya sadece başlık satırı içeriyor.")
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Excel dosyası boş veya hiç veri satırı içermiyor."})
		}

		// Başlık satırını atla (rows[0])
		header := rows[0]
		log.Printf("📌 XLSX Header: %s\n", strings.Join(header, ","))

		for i, row := range rows {
			if i == 0 { // Başlık satırını atla
				continue
			}
			lineNum = i + 1 // Gerçek Excel satır numarası (1 tabanlı)

			// Boş satırları atla (tüm hücreleri boşsa)
			if len(row) == 0 || strings.Join(row, "") == "" {
				continue
			}

			// Eğer satırda yeterli sütun yoksa uyarı ver ve atla
			if len(row) < 3 {
				log.Printf("⚠️ Satır %d atlandı: Yetersiz sütun sayısı (%d yerine 3 bekleniyor)\n", lineNum, len(row))
				continue
			}

			activityCode := models.ActivityCode{
				ActivityCode:            strings.TrimSpace(row[0]),
				ActivityGroupCode:       strings.TrimSpace(row[1]),
				ActivityCodeExplanation: strings.TrimSpace(row[2]),
			}
			activityCodes = append(activityCodes, activityCode)
			recordCount++
		}

	default:
		log.Printf("❌ Desteklenmeyen dosya uzantısı: %s", fileExtension)
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Desteklenmeyen dosya tipi. Lütfen .csv veya .xlsx dosyası yükleyin."})
	}

	if recordCount == 0 {
		log.Println("⚠️ Dosyada işlenecek hiç veri satırı bulunamadı (başlık hariç).")
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Dosyada boş veya hiç veri satırı içermiyor."})
	}

	log.Printf("🚀 %d adet activity_code kaydı veritabanına ekleniyor...\n", recordCount)

	// Frontend'den gelen `reset=true` parametresiyle tüm tabloyu temizleme mantığı
	if c.QueryBool("reset", false) {
		log.Println("🚀 'reset=true' parametresi algılandı, mevcut aktivite kodları temizleniyor...")
		_, err := db.DB.NewDelete().
			Model(&models.ActivityCode{}).
			Where("TRUE").
			Exec(context.Background())
		if err != nil {
			log.Printf("❌ Mevcut aktivite kodları temizlenirken hata oluştu: %v", err)
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": fmt.Sprintf("Mevcut veriler temizlenirken hata oluştu: %v", err)})
		}
		log.Println("✅ Mevcut aktivite kodları başarıyla temizlendi.")
	}

	_, err = db.DB.NewInsert().
		Model(&activityCodes).
		On("CONFLICT (activity_code) DO UPDATE").
		Set("activity_group_code = EXCLUDED.activity_group_code").
		Set("activity_code_explanation = EXCLUDED.activity_code_explanation").
		Exec(context.Background())
	if err != nil {
		log.Printf("❌ Veritabanına ekleme hatası: %v", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": fmt.Sprintf("Veritabanına ekleme hatası: %v", err)})
	}

	log.Printf("✅ %d adet activity_code kaydı başarıyla eklendi/güncellendi.\n", recordCount)
	return c.Status(fiber.StatusOK).JSON(fiber.Map{"success": recordCount, "failed": 0, "message": fmt.Sprintf("%d kayıt başarıyla eklendi/güncellendi.", recordCount)})
}
