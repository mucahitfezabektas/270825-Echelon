// C:\Users\mucah_wi2yyc2\Desktop\mini_CMS_Desktop_App\backend\handlers\actual_import.go

package handlers

import (
	"fmt"
	"io"
	"log"
	"strings"

	// time paketi eklendi
	"mini_CMS_Desktop_App/db"
	"mini_CMS_Desktop_App/handlers/ftl" // ftl handler'ı için
	"mini_CMS_Desktop_App/models"
	"mini_CMS_Desktop_App/repositories"
	"mini_CMS_Desktop_App/services"

	"github.com/gofiber/fiber/v2"
	"github.com/xuri/excelize/v2" // XLSX okuma kütüphanesi
)

// ActualImportXLSXHandler, XLSX dosya yükleme ve işleme mantığını içerir
type ActualImportXLSXHandler struct {
	actualRepo *repositories.ActualRepository
	ftlCalc    *services.FTLCalculator
	tripRepo   *repositories.TripRepository
	ftlHandler *ftl.FTLHandler
}

// NewActualImportXLSXHandler, handler'ın yeni bir örneğini oluşturur
func NewActualImportXLSXHandler(
	actualRepo *repositories.ActualRepository,
	ftlCalc *services.FTLCalculator,
	tripRepo *repositories.TripRepository,
	ftlHandler *ftl.FTLHandler,
) *ActualImportXLSXHandler {
	return &ActualImportXLSXHandler{
		actualRepo: actualRepo,
		ftlCalc:    ftlCalc,
		tripRepo:   tripRepo,
		ftlHandler: ftlHandler,
	}
}

// ImportActualXLSX, XLSX dosyasını alır, işler ve veritabanına kaydeder.
func (h *ActualImportXLSXHandler) ImportActualXLSX(c *fiber.Ctx) error {
	periodMonth := c.Query("month")
	reset := c.Query("reset") == "true"

	log.Println("🔍 ImportActualXLSX çağrıldı")
	log.Printf("ℹ️  Query params -> periodMonth: %s | reset: %v\n", periodMonth, reset)

	if periodMonth == "" {
		log.Println("❌ Eksik periodMonth parametresi")
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "periodMonth parametresi gerekli"})
	}

	var err error

	if reset {
		log.Println("⚠️  actuals tablosu sıfırlanıyor...")
		_, err = db.DB.NewTruncateTable().Model((*models.Actual)(nil)).Exec(c.Context())
		if err != nil {
			log.Printf("❌ Tablo sıfırlama hatası: %v\n", err)
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Tablo sıfırlanamadı", "details": err.Error()})
		}
		log.Println("⚠️  trips tablosu sıfırlanıyor...")
		_, err = db.DB.NewTruncateTable().Model((*models.Trip)(nil)).Exec(c.Context())
		if err != nil {
			log.Printf("❌ Trip tablosu sıfırlama hatası: %v\n", err)
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Trip tablosu sıfırlanamadı", "details": err.Error()})
		}
	}

	fileHeader, err := c.FormFile("actual_file_xlsx")
	if err != nil {
		log.Printf("❌ XLSX dosya alınamadı: %v", err)
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "XLSX dosya alınamadı", "details": err.Error()})
	}

	file, err := fileHeader.Open()
	if err != nil {
		log.Printf("❌ Yüklenen dosya açılamadı: %v", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Yüklenen dosya açılamadı", "details": err.Error()})
	}
	defer file.Close()

	f, err := excelize.OpenReader(file)
	if err != nil {
		log.Printf("❌ Excel dosyası okunamadı: %v", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Excel dosyası okunamadı", "details": err.Error()})
	}

	sheetName := f.GetSheetName(f.GetActiveSheetIndex())
	if sheetName == "" {
		log.Println("❌ Excel dosyasında aktif sayfa bulunamadı.")
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Excel dosyasında aktif sayfa bulunamadı."})
	}

	incomingXLSXColumnNames := []string{
		"group_code", "activity_code", "person_id", "surname", "name", "base_filo",
		"class", "flight_position", "flight_no", "departure_port", "arrival_port",
		"departure_time", "arrival_time", "plane_cms_type", "plane_tail_name",
		"checkin_date", "duty_start", "duty_end", "trip_id",
		"excel_original_flight_id",
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

	xlsxColNameToIndex := make(map[string]int)
	for i, colName := range incomingXLSXColumnNames {
		xlsxColNameToIndex[colName] = i
	}

	dbColNameToIndex := make(map[string]int)
	for i, colName := range dbColumnOrder {
		dbColNameToIndex[colName] = i
	}

	timeDBColumns := map[string]bool{
		"departure_time": true,
		"arrival_time":   true,
		"checkin_date":   true,
		"duty_start":     true,
		"duty_end":       true,
	}

	pr, pw := io.Pipe()
	go func() {
		var writeErr error
		defer pw.Close()

		newHeaderLine := strings.Join(dbColumnOrder, ";") + "\n"
		_, writeErr = pw.Write([]byte(newHeaderLine))
		if writeErr != nil {
			log.Printf("❌ COPY FROM için başlık satırı Pipe'a yazılamadı: %v", writeErr)
			pw.CloseWithError(writeErr)
			return
		}
		log.Printf("📌 COPY FROM için oluşturulan CSV Header: %s", strings.TrimSpace(newHeaderLine))

		rows, err := f.Rows(sheetName)
		if err != nil {
			log.Printf("❌ Excel sayfasındaki satırlar okunamadı: %v", err)
			pw.CloseWithError(err)
			return
		}
		defer rows.Close()

		lineNum := 0
		for rows.Next() {
			lineNum++

			if lineNum <= 3 { // İlk 3 satır atlanıyor
				log.Printf("ℹ️  Excel'in ilk %d satırı atlanıyor (Başlıklar/boş satır).", lineNum)
				continue
			}

			row, err := rows.Columns()
			if err != nil {
				log.Printf("❌ Excel satırı okunamadı (satır %d): %v", lineNum, err)
				continue
			}

			if len(row) == 0 || strings.Join(row, "") == "" {
				log.Printf("ℹ️  Excel'de boş satır atlanıyor (satır %d).", lineNum)
				continue
			}

			// --- BURASI GÜNCELLENDİ: Sütun sayısı kontrolü artık doğru ---
			// len(row) artık incomingXLSXColumnNames ile eşleşmeli (20 sütun)
			if len(row) != len(incomingXLSXColumnNames) {
				log.Printf("UYARI: Satır %d, beklenen %d sütun (%v) yerine %d sütun (%v) içeriyor. Satır atlanıyor.",
					lineNum, len(incomingXLSXColumnNames), incomingXLSXColumnNames, len(row), row)
				continue
			}

			processedRecord := make([]string, len(dbColumnOrder))

			// Gerekli ham değerleri al (UçuşID oluşturmak için)
			// Bu alanların incomingXLSXColumnNames içinde doğru indekslerde olduğundan emin olun.
			originalFlightNo := row[xlsxColNameToIndex["flight_no"]]
			arrivalPortRaw := row[xlsxColNameToIndex["arrival_port"]]
			rawDepartureTime := row[xlsxColNameToIndex["departure_time"]]

			depTime, parseDepTimeErr := models.ParseTimeFromDMYHMS(rawDepartureTime)
			if parseDepTimeErr != nil {
				log.Printf("ERROR: Satır %d, 'departure_time' ayrıştırma hatası: %v (Değer: '%s'). Satır atlanıyor.", lineNum, parseDepTimeErr, rawDepartureTime)
				continue
			}

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
						log.Printf("UYARI: Satır %d, DB sütunu '%s' (XLSX'te bekleniyor) veri satırında veya eşlemede bulunamadı. Boş bırakılıyor.", lineNum, dbColName)
						valueToInsert = ""
					} else {
						rawValue := row[xlsxColIdx]
						if timeDBColumns[dbColName] {
							parsedTime, parseErr := models.ParseTimeFromDMYHMS(rawValue)
							if parseErr != nil {
								// log.Printf("UYARI: Satır %d, '%s' alanı için tarih ayrıştırma hatası: %v (Değer: '%s'). Boş bırakılıyor.", lineNum, dbColName, parseErr, rawValue)
								valueToInsert = ""
							} else {
								valueToInsert = parsedTime.Format("2006-01-02 15:04:05") // PostgreSQL formatı
							}
						} else {
							valueToInsert = strings.TrimSpace(rawValue)
						}
					}
				}
				processedRecord[dbIdx] = valueToInsert
			}

			outputLine := strings.Join(processedRecord, ";") + "\n"
			_, writeErr = pw.Write([]byte(outputLine))
			if writeErr != nil {
				log.Printf("❌ İşlenmiş satır Pipe'a yazılamadı (satır %d): %v", lineNum, writeErr)
				pw.CloseWithError(writeErr)
				return
			}
			if lineNum >= 4 && lineNum <= 8 {
				log.Printf("🔹 Satır %d örnek (işlenmiş): %s", lineNum, strings.TrimSpace(outputLine))
			}
		}

		log.Printf("✅ Excel işleme tamamlandı. Toplam işlenmiş veri satırı: %d", lineNum-3)
	}()

	conn := db.RawPGConn

	copyStatement := fmt.Sprintf(`
	COPY actuals (
		%s
	) FROM STDIN WITH (FORMAT CSV, HEADER TRUE, DELIMITER ';')
	`, strings.Join(dbColumnOrder, ", "))

	_, err = conn.CopyFrom(c.Context(), pr, copyStatement)

	if err != nil {
		log.Printf("❌ COPY FROM hatası: %v\n", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "COPY FROM başarısız", "details": err.Error()})
	}

	log.Println("✅ COPY FROM başarılı!")

	// --- FTL Hesaplamalarını Tetikle --- (Yorum satırı olarak kalacak)
	/*
		var affectedCrewIDs []string
		err = db.DB.NewSelect().
			Model((*models.Actual)(nil)).
			ColumnExpr("DISTINCT person_id").
			Where("period_month = ?", periodMonth).
			Scan(c.Context(), &affectedCrewIDs)
		if err != nil {
			log.Printf("❌ Etkilenen ekip üyeleri çekilirken hata: %v", err)
		}

		log.Printf("ℹ️  FTL hesaplaması tetiklenecek ekip üyeleri: %v", affectedCrewIDs)

		for _, crewID := range affectedCrewIDs {
			go func(capturedCrewID string) {
				log.Printf("Bilgi: Ekip %s için FTL hesaplaması arka planda tetikleniyor...", capturedCrewID)
				err := h.ftlCalc.RecalculateCrewSchedule(capturedCrewID)
				if err != nil {
					log.Printf("Hata: Ekip %s için FTL hesaplaması başarısız: %v", capturedCrewID, err)
				} else {
					log.Printf("✅ Ekip %s için FTL hesaplaması tamamlandı.", capturedCrewID)
				}
			}(crewID)
		}
	*/
	// --- FTL Hesaplamalarını Tetikleme kısmı SONU ---

	return c.JSON(fiber.Map{"success": 1, "failed": 0, "message": "XLSX import başarılı."})
}
