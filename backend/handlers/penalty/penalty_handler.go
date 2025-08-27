package penalty

import (
	"context"
	"encoding/csv"
	"fmt"
	"io"
	"log"
	"path/filepath"
	"strings"
	"time" // ✅ time paketi eklendi

	"mini_CMS_Desktop_App/db"
	"mini_CMS_Desktop_App/models"

	"github.com/gofiber/fiber/v2"
	"github.com/xuri/excelize/v2"
)

// ImportPenaltyData, penalties tablosuna hem CSV hem de XLSX verisi aktarır.
func ImportPenaltyData(c *fiber.Ctx) error {
	log.Println("🔍 ImportPenaltyData çağrıldı")

	fileHeader, err := c.FormFile("file")
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

	var penaltyEntries []models.Penalty
	recordCount := 0
	failedCount := 0
	lineNum := 0

	fileExtension := strings.ToLower(filepath.Ext(fileHeader.Filename))

	switch fileExtension {
	case ".csv":
		log.Println("Handling CSV file for Penalty...")
		reader := csv.NewReader(file)
		reader.Comma = ','
		reader.FieldsPerRecord = -1
		reader.LazyQuotes = true

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

		expectedColumnCount := 9
		if len(header) < expectedColumnCount {
			log.Printf("❌ CSV başlık satırı yetersiz sütun içeriyor: %d yerine %d bekleniyor", len(header), expectedColumnCount)
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": fmt.Sprintf("CSV başlık satırı yetersiz sütun içeriyor: %d yerine %d bekleniyor", len(header), expectedColumnCount)})
		}

		for {
			lineNum++
			record, err := reader.Read()
			if err == io.EOF {
				break
			}
			if err != nil {
				log.Printf("❌ CSV satırı okuma hatası (Satır %d): %v\n", lineNum+1, err)
				failedCount++
				continue
			}

			if len(record) == 0 || strings.Join(record, "") == "" {
				continue
			}

			if len(record) < expectedColumnCount {
				log.Printf("⚠️ Satır %d atlandı: Yetersiz sütun sayısı (%d yerine %d bekleniyor)\n", lineNum+1, len(record), expectedColumnCount)
				failedCount++
				continue
			}

			penaltyStartDate, err := parseDateTimeToUnix(strings.TrimSpace(record[7])) // ✅ parseDateTimeToUnix kullanıldı
			if err != nil {
				log.Printf("❌ Satır %d, 'penalty_start_date' dönüşüm hatası: %v. Satır atlandı.\n", lineNum+1, err)
				failedCount++
				continue
			}

			penaltyEndDate, err := parseDateTimeToUnix(strings.TrimSpace(record[8])) // ✅ parseDateTimeToUnix kullanıldı
			if err != nil {
				log.Printf("❌ Satır %d, 'penalty_end_date' dönüşüm hatası: %v. Satır atlandı.\n", lineNum+1, err)
				failedCount++
				continue
			}

			penaltyEntry := models.Penalty{
				PersonID:               strings.TrimSpace(record[0]),
				PersonSurname:          strings.TrimSpace(record[1]),
				PersonName:             strings.TrimSpace(record[2]),
				UcucuSinifi:            strings.TrimSpace(record[3]),
				BaseFilo:               strings.TrimSpace(record[4]),
				PenaltyCode:            strings.TrimSpace(record[5]),
				PenaltyCodeExplanation: strings.TrimSpace(record[6]),
				PenaltyStartDate:       penaltyStartDate,
				PenaltyEndDate:         penaltyEndDate,
			}
			penaltyEntries = append(penaltyEntries, penaltyEntry)
			recordCount++
		}

	case ".xlsx":
		log.Println("Handling XLSX file for Penalty...")
		f, err := excelize.OpenReader(file)
		if err != nil {
			log.Printf("❌ Excel dosyası açılamadı: %v", err)
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": fmt.Sprintf("Excel dosyası açılamadı: %v", err)})
		}

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

		if len(rows) < 2 {
			log.Println("⚠️ Excel dosyası boş veya sadece başlık satırı içeriyor.")
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Excel dosyası boş veya hiç veri satırı içermiyor."})
		}

		header := rows[0]
		log.Printf("📌 XLSX Header: %s\n", strings.Join(header, ","))

		expectedColumnCount := 9
		if len(header) < expectedColumnCount {
			log.Printf("❌ XLSX başlık satırı yetersiz sütun içeriyor: %d yerine %d bekleniyor", len(header), expectedColumnCount)
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": fmt.Sprintf("XLSX başlık satırı yetersiz sütun içeriyor: %d yerine %d bekleniyor", len(header), expectedColumnCount)})
		}

		for i, row := range rows {
			if i == 0 {
				continue
			}
			lineNum = i + 1

			if len(row) == 0 || strings.Join(row, "") == "" {
				continue
			}

			if len(row) < expectedColumnCount {
				log.Printf("⚠️ Satır %d atlandı: Yetersiz sütun sayısı (%d yerine %d bekleniyor)\n", lineNum, len(row), expectedColumnCount)
				failedCount++
				continue
			}

			penaltyStartDate, err := parseDateTimeToUnix(strings.TrimSpace(row[7])) // ✅ parseDateTimeToUnix kullanıldı
			if err != nil {
				log.Printf("❌ Satır %d, 'penalty_start_date' dönüşüm hatası: %v. Satır atlandı.\n", lineNum, err)
				failedCount++
				continue
			}

			penaltyEndDate, err := parseDateTimeToUnix(strings.TrimSpace(row[8])) // ✅ parseDateTimeToUnix kullanıldı
			if err != nil {
				log.Printf("❌ Satır %d, 'penalty_end_date' dönüşüm hatası: %v. Satır atlandı.\n", lineNum, err)
				failedCount++
				continue
			}

			penaltyEntry := models.Penalty{
				PersonID:               strings.TrimSpace(row[0]),
				PersonSurname:          strings.TrimSpace(row[1]),
				PersonName:             strings.TrimSpace(row[2]),
				UcucuSinifi:            strings.TrimSpace(row[3]),
				BaseFilo:               strings.TrimSpace(row[4]),
				PenaltyCode:            strings.TrimSpace(row[5]),
				PenaltyCodeExplanation: strings.TrimSpace(row[6]),
				PenaltyStartDate:       penaltyStartDate,
				PenaltyEndDate:         penaltyEndDate,
			}
			penaltyEntries = append(penaltyEntries, penaltyEntry)
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

	log.Printf("🚀 %d adet penalty kaydı veritabanına ekleniyor...\n", recordCount)

	if c.QueryBool("reset", false) {
		log.Println("🚀 'reset=true' parametresi algılandı, mevcut ceza bilgileri temizleniyor...")
		_, err := db.DB.NewDelete().
			Model(&models.Penalty{}).
			Where("TRUE").
			Exec(context.Background())
		if err != nil {
			log.Printf("❌ Mevcut ceza bilgileri temizlenirken hata oluştu: %v", err)
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": fmt.Sprintf("Mevcut veriler temizlenirken hata oluştu: %v", err)})
		}
		log.Println("✅ Mevcut ceza bilgileri başarıyla temizlendi.")
	}

	_, err = db.DB.NewInsert().
		Model(&penaltyEntries).
		Exec(context.Background())
	if err != nil {
		log.Printf("❌ Veritabanına ekleme hatası: %v", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"success": recordCount, "failed": failedCount, "error": fmt.Sprintf("Veritabanına ekleme hatası: %v", err)})
	}

	log.Printf("✅ %d adet penalty kaydı başarıyla eklendi. %d kayıt atlandı.\n", recordCount, failedCount)
	return c.Status(fiber.StatusOK).JSON(fiber.Map{"success": recordCount, "failed": failedCount, "message": fmt.Sprintf("%d kayıt başarıyla eklendi. %d kayıt atlandı.", recordCount, failedCount)})
}

// parseDateTimeToUnix, "DD/MM/YYYY HH:MM:SS" formatındaki string'i Unix timestamp (milisaniye) olarak int64'e dönüştürür.
// Boş stringler için 0 döndürür, hata vermez.
func parseDateTimeToUnix(s string) (int64, error) {
	if s == "" {
		return 0, nil // Boş string ise 0 döndür, hata verme
	}

	// Kabul edilen formatı belirtin
	// "02/01/2006 15:04:05" -> DD/MM/YYYY HH:MM:SS (Go'nun referans tarihi)
	const dateTimeLayout = "02/01/2006 15:04:05"

	t, err := time.Parse(dateTimeLayout, s)
	if err != nil {
		// Tarih ayrıştırma hatası durumunda detaylı loglama
		return 0, fmt.Errorf("tarih formatı '%s' ayrıştırma hatası: %w", s, err)
	}

	// Unix timestamp'i milisaniye cinsinden döndür
	return t.UnixNano() / int64(time.Millisecond), nil // Milisaniye cinsinden
}
