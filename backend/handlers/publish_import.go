package handlers

import (
	"fmt"
	"io"
	"log"
	"strings"

	// time paketi gerekli olduğu için eklendi
	"mini_CMS_Desktop_App/db"
	"mini_CMS_Desktop_App/models"
	"mini_CMS_Desktop_App/repositories" // repositories paketi gerekli olduğu için eklendi

	"github.com/gofiber/fiber/v2"
	"github.com/xuri/excelize/v2" // XLSX okuma kütüphanesi
)

// PublishImportXLSXHandler, XLSX dosya yükleme ve işleme mantığını içerir
// Publish verisi için FTL hesaplamaları veya Trip Reposu gibi bağımlılıklara ihtiyaç duymaz.
type PublishImportXLSXHandler struct {
	publishRepo *repositories.PublishRepository // Publish verileri için yeni repository
}

// NewPublishImportXLSXHandler, handler'ın yeni bir örneğini oluşturur
func NewPublishImportXLSXHandler(
	publishRepo *repositories.PublishRepository,
) *PublishImportXLSXHandler {
	return &PublishImportXLSXHandler{
		publishRepo: publishRepo,
	}
}

// ImportPublishXLSX, XLSX dosyasını alır, işler ve 'publishes' tablosuna kaydeder.
// Bu fonksiyon, ayın başında gönderilen planlanmış verileri içe aktarmak için kullanılır.
func (h *PublishImportXLSXHandler) ImportPublishXLSX(c *fiber.Ctx) error {
	periodMonth := c.Query("month")
	reset := c.Query("reset") == "true" // 'reset=true' query parametresi ile tablo sıfırlanabilir

	log.Println("🔍 ImportPublishXLSX çağrıldı")
	log.Printf("ℹ️  Query params -> periodMonth: %s | reset: %v\n", periodMonth, reset)

	if periodMonth == "" {
		log.Println("❌ Eksik periodMonth parametresi")
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "periodMonth parametresi gerekli"})
	}

	var err error

	// Eğer reset parametresi true ise 'publishes' tablosunu sıfırla
	if reset {
		log.Println("⚠️  publishes tablosu sıfırlanıyor...")
		_, err = db.DB.NewTruncateTable().Model((*models.Publish)(nil)).Exec(c.Context())
		if err != nil {
			log.Printf("❌ Publish tablosu sıfırlama hatası: %v\n", err)
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Publish tablosu sıfırlanamadı", "details": err.Error()})
		}
		log.Println("✅ publishes tablosu başarıyla sıfırlandı.")
	}

	// Yüklenen XLSX dosyasını al
	fileHeader, err := c.FormFile("publish_file_xlsx")
	if err != nil {
		log.Printf("❌ XLSX dosya alınamadı: %v", err)
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "XLSX dosya alınamadı", "details": err.Error()})
	}

	// Dosyayı aç
	file, err := fileHeader.Open()
	if err != nil {
		log.Printf("❌ Yüklenen dosya açılamadı: %v", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Yüklenen dosya açılamadı", "details": err.Error()})
	}
	defer file.Close() // Fonksiyon bitiminde dosyayı kapat

	// Excel dosyasını okuyucu ile aç
	f, err := excelize.OpenReader(file)
	if err != nil {
		log.Printf("❌ Excel dosyası okunamadı: %v", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Excel dosyası okunamadı", "details": err.Error()})
	}

	// Aktif çalışma sayfasının adını al
	sheetName := f.GetSheetName(f.GetActiveSheetIndex())
	if sheetName == "" {
		log.Println("❌ Excel dosyasında aktif sayfa bulunamadı.")
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Excel dosyasında aktif sayfa bulunamadı."})
	}

	// Gelen XLSX dosyasındaki sütun başlıkları (beklenen sıra)
	incomingXLSXColumnNames := []string{
		"group_code", "activity_code", "person_id", "surname", "name", "base_filo",
		"class", "flight_position", "flight_no", "departure_port", "arrival_port",
		"departure_time", "arrival_time", "plane_cms_type", "plane_tail_name",
		"checkin_date", "duty_start", "duty_end", "trip_id",
		"excel_original_flight_id", // Bu sütun sadece Excel'den okumak için kullanılır, DB'ye yazılmaz
	}

	// dbColumnOrder: Veritabanına COPY FROM ile yazılacak sütunların sırası (DB şemasına uygun)
	// 'data_id' (UUID, PK, default:gen_random_uuid()) ve 'aircraft_type' (derived) bu listede yer almaz.
	// 'excel_original_flight_id' de burada yer almaz çünkü bizim 'flight_id'miz kod içinde üretiliyor.
	dbColumnOrder := []string{
		"ucus_id",
		"group_code", "activity_code", "person_id", "surname", "name", "base_filo",
		"class", "flight_position", "flight_no", "departure_port", "arrival_port",
		"departure_time", "arrival_time", "plane_cms_type", "plane_tail_name",
		"trip_id",
		"checkin_date",
		"duty_start", "duty_end", "period_month",
	}

	// XLSX sütun adlarını indeksleriyle eşleştiren haritalar oluştur
	xlsxColNameToIndex := make(map[string]int)
	for i, colName := range incomingXLSXColumnNames {
		xlsxColNameToIndex[colName] = i
	}

	// Veritabanı sütun adlarını indeksleriyle eşleştiren harita (kullanılmıyor ama örnek için tutulabilir)
	// dbColNameToIndex := make(map[string]int)
	// for i, colName := range dbColumnOrder {
	// 	dbColNameToIndex[colName] = i
	// }

	// Zaman formatında olan DB sütunlarını işaretle
	timeDBColumns := map[string]bool{
		"departure_time": true,
		"arrival_time":   true,
		"checkin_date":   true,
		"duty_start":     true,
		"duty_end":       true,
	}

	// Pipe oluşturarak Excel verisini PostgreSQL COPY FROM formatına dönüştür
	pr, pw := io.Pipe()
	go func() {
		var writeErr error
		defer pw.Close() // Go rutinden çıkarken pipe'ı kapat

		// COPY FROM için CSV başlık satırını yaz
		newHeaderLine := strings.Join(dbColumnOrder, ";") + "\n"
		_, writeErr = pw.Write([]byte(newHeaderLine))
		if writeErr != nil {
			log.Printf("❌ COPY FROM için başlık satırı Pipe'a yazılamadı: %v", writeErr)
			pw.CloseWithError(writeErr)
			return
		}
		log.Printf("📌 COPY FROM için oluşturulan CSV Header: %s", strings.TrimSpace(newHeaderLine))

		// Excel satırlarını oku
		rows, err := f.Rows(sheetName)
		if err != nil {
			log.Printf("❌ Excel sayfasındaki satırlar okunamadı: %v", err)
			pw.CloseWithError(err)
			return
		}
		defer rows.Close() // Excel satır okuyucuyu kapat

		lineNum := 0
		for rows.Next() {
			lineNum++

			if lineNum <= 3 { // İlk 3 satır genellikle başlık veya boş satır olduğu için atlanıyor
				log.Printf("ℹ️  Excel'in ilk %d satırı atlanıyor (Başlıklar/boş satır).", lineNum)
				continue
			}

			row, err := rows.Columns()
			if err != nil {
				log.Printf("❌ Excel satırı okunamadı (satır %d): %v", lineNum, err)
				continue // Bu satırı atla ve bir sonrakine geç
			}

			// Tamamen boş satırları atla
			if len(row) == 0 || strings.Join(row, "") == "" {
				log.Printf("ℹ️  Excel'de boş satır atlanıyor (satır %d).", lineNum)
				continue
			}

			// Sütun sayısı kontrolü
			if len(row) != len(incomingXLSXColumnNames) {
				log.Printf("UYARI: Satır %d, beklenen %d sütun (%v) yerine %d sütun (%v) içeriyor. Satır atlanıyor.",
					lineNum, len(incomingXLSXColumnNames), incomingXLSXColumnNames, len(row), row)
				continue
			}

			// İşlenmiş kaydı tutacak slice
			processedRecord := make([]string, len(dbColumnOrder))

			// UçuşID oluşturmak için gerekli ham değerleri al
			originalFlightNo := row[xlsxColNameToIndex["flight_no"]]
			arrivalPortRaw := row[xlsxColNameToIndex["arrival_port"]]
			rawDepartureTime := row[xlsxColNameToIndex["departure_time"]]

			// Kalkış zamanını ayrıştır
			depTime, parseDepTimeErr := models.ParseTimeFromDMYHMS(rawDepartureTime)
			if parseDepTimeErr != nil {
				log.Printf("ERROR: Satır %d, 'departure_time' ayrıştırma hatası: %v (Değer: '%s'). Satır atlanıyor.", lineNum, parseDepTimeErr, rawDepartureTime)
				continue
			}

			// Benzersiz UçuşID oluştur
			var finalUcusID string
			if originalFlightNo != "" {
				finalUcusID = fmt.Sprintf("%s-%s-%s",
					strings.TrimSpace(originalFlightNo),
					strings.TrimSpace(arrivalPortRaw),
					depTime.Format("20060102150405"), // YYYYMMDDHHMMSS formatı
				)
			} else {
				finalUcusID = fmt.Sprintf("NO_FLIGHTNO-%s-%s",
					strings.TrimSpace(arrivalPortRaw),
					depTime.Format("20060102150405"),
				)
				log.Printf("UYARI: Satır %d, orijinal 'flight_no' boş. Yerine '%s' UçuşID olarak kullanıldı.", lineNum, finalUcusID)
			}

			// Her bir veritabanı sütunu için değeri hazırla
			for dbIdx, dbColName := range dbColumnOrder {
				var valueToInsert string
				switch dbColName {
				case "ucus_id":
					valueToInsert = finalUcusID
				case "period_month":
					valueToInsert = periodMonth
				default:
					xlsxColIdx, ok := xlsxColNameToIndex[dbColName]
					if !ok || xlsxColIdx >= len(row) {
						// Eğer DB sütunu XLSX'te yoksa veya indeksi hatalıysa boş bırak
						log.Printf("UYARI: Satır %d, DB sütunu '%s' (XLSX'te bekleniyor) veri satırında veya eşlemede bulunamadı. Boş bırakılıyor.", lineNum, dbColName)
						valueToInsert = ""
					} else {
						rawValue := row[xlsxColIdx]
						if timeDBColumns[dbColName] {
							// Zaman sütunlarını PostgreSQL'in beklediği formata dönüştür
							parsedTime, parseErr := models.ParseTimeFromDMYHMS(rawValue)
							if parseErr != nil {
								// Tarih ayrıştırma hatasında boş bırak
								valueToInsert = ""
							} else {
								valueToInsert = parsedTime.Format("2006-01-02 15:04:05") // PostgreSQL TIMESTAMP formatı
							}
						} else {
							// Diğer sütunları doğrudan kullan (boşlukları temizle)
							valueToInsert = strings.TrimSpace(rawValue)
						}
					}
				}
				processedRecord[dbIdx] = valueToInsert
			}

			// İşlenmiş kaydı Pipe'a CSV formatında yaz
			outputLine := strings.Join(processedRecord, ";") + "\n"
			_, writeErr = pw.Write([]byte(outputLine))
			if writeErr != nil {
				log.Printf("❌ İşlenmiş satır Pipe'a yazılamadı (satır %d): %v", lineNum, writeErr)
				pw.CloseWithError(writeErr)
				return
			}
			// İlk birkaç işlenmiş satırı logla (debug amaçlı)
			if lineNum >= 4 && lineNum <= 8 {
				log.Printf("🔹 Satır %d örnek (işlenmiş): %s", lineNum, strings.TrimSpace(outputLine))
			}
		}

		log.Printf("✅ Excel işleme tamamlandı. Toplam işlenmiş veri satırı: %d", lineNum-3)
	}()

	// PostgreSQL COPY FROM komutunu çalıştır
	conn := db.RawPGConn // Ham pgx bağlantısını kullan
	copyStatement := fmt.Sprintf(`
	COPY publishes (
		%s
	) FROM STDIN WITH (FORMAT CSV, HEADER TRUE, DELIMITER ';')
	`, strings.Join(dbColumnOrder, ", "))

	_, err = conn.CopyFrom(c.Context(), pr, copyStatement)

	if err != nil {
		log.Printf("❌ COPY FROM hatası: %v\n", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "COPY FROM başarısız", "details": err.Error()})
	}

	log.Println("✅ COPY FROM başarılı!")

	// Başarılı yanıt dön
	return c.JSON(fiber.Map{"success": 1, "failed": 0, "message": "XLSX Publish import başarılı."})
}
