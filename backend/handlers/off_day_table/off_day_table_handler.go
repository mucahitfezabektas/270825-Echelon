package off_day_table

import (
	"context"
	"encoding/csv" // CSV işlemleri için kalacak
	"fmt"
	"io"
	"log"
	"path/filepath" // Dosya uzantısını almak için eklendi
	"strconv"       // string'den int32'ye dönüşüm için
	"strings"

	"mini_CMS_Desktop_App/db"
	"mini_CMS_Desktop_App/models"

	"github.com/gofiber/fiber/v2"
	"github.com/xuri/excelize/v2" // ✅ Excelize kütüphanesi eklendi
)

// ImportOffDayTableData, off_day_table tablosuna hem CSV hem de XLSX verisi aktarır.
// Fonksiyon adı ImportOffDayTableCSV'den ImportOffDayTableData olarak değiştirildi.
func ImportOffDayTableData(c *fiber.Ctx) error {
	log.Println("🔍 ImportOffDayTableData çağrıldı")

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

	var offDayEntries []models.OffDayTable
	recordCount := 0
	failedCount := 0
	lineNum := 0 // Başlık satırından sonraki veri satırlarını takip etmek için (1'den başlar)

	// Dosya uzantısına göre okuma stratejisi belirle
	fileExtension := strings.ToLower(filepath.Ext(fileHeader.Filename))

	switch fileExtension {
	case ".csv":
		log.Println("Handling CSV file for Off Day Table...")
		reader := csv.NewReader(file)
		reader.Comma = ','
		reader.FieldsPerRecord = -1
		reader.LazyQuotes = true

		header, err := reader.Read() // Başlık satırını oku
		if err != nil {
			if err == io.EOF {
				log.Println("⚠️ CSV dosyası boş veya sadece başlık satırı içeriyor.")
				return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "CSV dosyası boş veya hiç veri satırı içermiyor."})
			}
			log.Printf("❌ CSV başlık satırı okunamadı: %v", err)
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": fmt.Sprintf("CSV başlık satırı okunamadı: %v", err)})
		}
		log.Printf("📌 CSV Header: %s\n", strings.Join(header, ","))

		// Beklenen sütun sayısı (models.OffDayTable'daki work_days, off_day_entitlement, distribution)
		expectedColumnCount := 3
		if len(header) < expectedColumnCount {
			log.Printf("❌ CSV başlık satırı yetersiz sütun içeriyor: %d yerine %d bekleniyor", len(header), expectedColumnCount)
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": fmt.Sprintf("CSV başlık satırı yetersiz sütun içeriyor: %d yerine %d bekleniyor", len(header), expectedColumnCount)})
		}

		for {
			lineNum++ // Veri satırı numarası (1'den başlar)
			record, err := reader.Read()
			if err == io.EOF {
				break
			}
			if err != nil {
				log.Printf("❌ CSV satırı okuma hatası (Satır %d): %v\n", lineNum+1, err) // Log için gerçek satır numarası
				failedCount++
				continue
			}

			// Boş satırları atla
			if len(record) == 0 || strings.Join(record, "") == "" {
				continue
			}

			// Sütun sayısı kontrolü
			if len(record) < expectedColumnCount {
				log.Printf("⚠️ Satır %d atlandı: Yetersiz sütun sayısı (%d yerine %d bekleniyor)\n", lineNum+1, len(record), expectedColumnCount)
				failedCount++
				continue
			}

			// Veri Türü Dönüşümleri
			workDays, err := strconv.ParseInt(strings.TrimSpace(record[0]), 10, 32)
			if err != nil {
				log.Printf("❌ Satır %d, 'work_days' dönüşüm hatası: %v. Satır atlandı.\n", lineNum+1, err)
				failedCount++
				continue
			}

			offDayEntitlement, err := strconv.ParseInt(strings.TrimSpace(record[1]), 10, 32)
			if err != nil {
				log.Printf("❌ Satır %d, 'off_day_entitlement' dönüşüm hatası: %v. Satır atlandı.\n", lineNum+1, err)
				failedCount++
				continue
			}

			offDayEntry := models.OffDayTable{
				WorkDays:          int32(workDays),
				OffDayEntitlement: int32(offDayEntitlement),
				Distribution:      strings.TrimSpace(record[2]),
			}
			offDayEntries = append(offDayEntries, offDayEntry)
			recordCount++
		}

	case ".xlsx":
		log.Println("Handling XLSX file for Off Day Table...")
		f, err := excelize.OpenReader(file)
		if err != nil {
			log.Printf("❌ Excel dosyası açılamadı: %v", err)
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": fmt.Sprintf("Excel dosyası açılamadı: %v", err)})
		}

		// Excel dosyasındaki ilk sayfayı dinamik olarak bul
		sheetList := f.GetSheetList()
		if len(sheetList) == 0 {
			log.Println("⚠️ Excel dosyasında hiç sayfa bulunamadı.")
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Excel dosyasında hiç sayfa bulunamadı."})
		}
		sheetName := sheetList[0]
		log.Printf("ℹ️ Okunan Excel sayfası: %s\n", sheetName)

		rows, err := f.GetRows(sheetName)
		if err != nil {
			log.Printf("❌ Excel sayfasından satırlar okunamadı: %v", err)
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": fmt.Sprintf("Excel sayfasından veri okunamadı: %v", err)})
		}

		if len(rows) < 2 { // Başlık satırı + en az bir veri satırı beklenir
			log.Println("⚠️ Excel dosyası boş veya sadece başlık satırı içeriyor.")
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Excel dosyası boş veya hiç veri satırı içermiyor."})
		}

		header := rows[0] // Başlık satırı
		log.Printf("📌 XLSX Header: %s\n", strings.Join(header, ","))

		expectedColumnCount := 3 // `work_days`, `off_day_entitlement`, `distribution`
		if len(header) < expectedColumnCount {
			log.Printf("❌ XLSX başlık satırı yetersiz sütun içeriyor: %d yerine %d bekleniyor", len(header), expectedColumnCount)
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": fmt.Sprintf("XLSX başlık satırı yetersiz sütun içeriyor: %d yerine %d bekleniyor", len(header), expectedColumnCount)})
		}

		for i, row := range rows {
			if i == 0 { // Başlık satırını atla
				continue
			}
			lineNum = i + 1 // Gerçek Excel satır numarası (1 tabanlı)

			// Boş satırları atla (tüm hücreleri boşsa)
			if len(row) == 0 || strings.Join(row, "") == "" {
				continue
			}

			// Sütun sayısı kontrolü
			if len(row) < expectedColumnCount {
				log.Printf("⚠️ Satır %d atlandı: Yetersiz sütun sayısı (%d yerine %d bekleniyor)\n", lineNum, len(row), expectedColumnCount)
				failedCount++
				continue
			}

			// Veri Türü Dönüşümleri
			workDays, err := strconv.ParseInt(strings.TrimSpace(row[0]), 10, 32)
			if err != nil {
				log.Printf("❌ Satır %d, 'work_days' dönüşüm hatası: %v. Satır atlandı.\n", lineNum, err)
				failedCount++
				continue
			}

			offDayEntitlement, err := strconv.ParseInt(strings.TrimSpace(row[1]), 10, 32)
			if err != nil {
				log.Printf("❌ Satır %d, 'off_day_entitlement' dönüşüm hatası: %v. Satır atlandı.\n", lineNum, err)
				failedCount++
				continue
			}

			offDayEntry := models.OffDayTable{
				WorkDays:          int32(workDays),
				OffDayEntitlement: int32(offDayEntitlement),
				Distribution:      strings.TrimSpace(row[2]),
			}
			offDayEntries = append(offDayEntries, offDayEntry)
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

	log.Printf("🚀 %d adet off_day_table kaydı veritabanına ekleniyor...\n", recordCount)

	// Frontend'den gelen `reset=true` parametresiyle tüm tabloyu temizleme mantığı
	if c.QueryBool("reset", false) {
		log.Println("🚀 'reset=true' parametresi algılandı, mevcut Off Day Tablosu temizleniyor...")
		_, err := db.DB.NewDelete().
			Model(&models.OffDayTable{}).
			Where("TRUE"). // Tüm kayıtları silmek için
			Exec(context.Background())
		if err != nil {
			log.Printf("❌ Mevcut Off Day Tablosu temizlenirken hata oluştu: %v", err)
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": fmt.Sprintf("Mevcut veriler temizlenirken hata oluştu: %v", err)})
		}
		log.Println("✅ Mevcut Off Day Tablosu başarıyla temizlendi.")
	}

	// 🚫 ON CONFLICT ifadesi kaldırıldı
	_, err = db.DB.NewInsert().
		Model(&offDayEntries).
		Exec(context.Background()) // Sadece yeni kayıtları ekler
	if err != nil {
		log.Printf("❌ Veritabanına ekleme hatası: %v", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"success": recordCount, "failed": failedCount, "error": fmt.Sprintf("Veritabanına ekleme hatası: %v", err)})
	}

	log.Printf("✅ %d adet off_day_table kaydı başarıyla eklendi. %d kayıt atlandı.\n", recordCount, failedCount)
	return c.Status(fiber.StatusOK).JSON(fiber.Map{"success": recordCount, "failed": failedCount, "message": fmt.Sprintf("%d kayıt başarıyla eklendi. %d kayıt atlandı.", recordCount, failedCount)})
}
