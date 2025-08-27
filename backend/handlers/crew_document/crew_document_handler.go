package crew_document

import (
	"context"
	"database/sql" // sql.NullInt64 ve sql.NullString için eklendi
	"encoding/csv"
	"fmt"
	"log"
	"path/filepath" // Dosya uzantısını almak için eklendi
	"strconv"
	"strings"
	"time" // Tarih dönüştürme için time paketi eklendi

	"mini_CMS_Desktop_App/db"
	"mini_CMS_Desktop_App/handlers/progress" // Progress bar için import
	"mini_CMS_Desktop_App/models"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"      // UUID oluşturmak için
	"github.com/xuri/excelize/v2" // ✅ Excelize kütüphanesi eklendi
)

// --- Helper Functions ---

// parseTimestampToNullInt64, string'i sql.NullInt64'e dönüştürür.
// Boş stringler veya parse edilemeyenler için geçerli olmayan (Valid: false) bir sql.NullInt64 döndürür.
// Penalty handler'daki parseDateTimeToUnix gibi formatlı tarih stringlerini de işleyebilir.
func parseTimestampToNullInt64(s string) (sql.NullInt64, error) {
	s = strings.TrimSpace(s)
	if s == "" {
		return sql.NullInt64{Valid: false}, nil // Boş string ise NULL olarak ayarla
	}

	// İlk olarak doğrudan int64 olarak parse etmeye çalış (Unix timestamp ise)
	val, err := strconv.ParseInt(s, 10, 64)
	if err == nil {
		return sql.NullInt64{Int64: val, Valid: true}, nil
	}

	// Eğer doğrudan int64 değilse, tarih/saat stringi olarak parse etmeye çalış
	// "DD/MM/YYYY HH:MM:SS" formatı için (Penalty handler'daki format ile uyumlu)
	const dateTimeLayout = "02/01/2006 15:04:05" // GG/AA/YYYY SS:DD:SS
	t, err := time.Parse(dateTimeLayout, s)
	if err == nil {
		return sql.NullInt64{Int64: t.UnixNano() / int64(time.Millisecond), Valid: true}, nil // Milisaniye cinsinden
	}

	// Başka formatlar da denenebilir, örneğin "YYYY-MM-DD HH:MM:SS"
	const ymdhmsLayout = "2006-01-02 15:04:05"
	t, err = time.Parse(ymdhmsLayout, s)
	if err == nil {
		return sql.NullInt64{Int64: t.UnixNano() / int64(time.Millisecond), Valid: true}, nil
	}

	// Hiçbir format eşleşmezse veya parse edilemezse hata döndür
	return sql.NullInt64{Valid: false}, fmt.Errorf("geçersiz tarih veya timestamp formatı: '%s'", s)
}

// parseBool, string'i boolean'a dönüştürür.
// Belirli Türkçe metinleri anlar: "Calisiyor", "Gecerli", "true", "1" -> true
// "Calismiyor", "Gecerli Degil", "false", "0" -> false
// Boş stringler veya bilinmeyen değerler için false döndürür.
func parseBool(s string) (bool, error) {
	s = strings.TrimSpace(s)
	s = strings.ToLower(s)

	switch s {
	case "calisiyor", "gecerli", "true", "1":
		return true, nil
	case "calismiyor", "gecerli degil", "false", "0":
		return false, nil
	case "": // Boş string ise false kabul et (varsayılan davranış)
		return false, nil
	default:
		return false, fmt.Errorf("bilinmeyen boolean değeri: '%s'", s)
	}
}

// --- Main Handler Function ---

// ImportCrewDocumentData, crew_documents tablosuna hem CSV hem de XLSX verisi aktarır.
// Fonksiyon adı ImportCrewDocumentCSV'den ImportCrewDocumentData olarak değiştirildi.
func ImportCrewDocumentData(c *fiber.Ctx) error {
	log.Println("🔍 ImportCrewDocumentData çağrıldı")

	processID := c.Query("process_id")
	if processID == "" {
		processID = uuid.New().String()
		log.Printf("⚠️ process_id bulunamadı, yeni bir tane oluşturuldu: %s", processID)
	}
	progress.SendProgressUpdate(processID, 0, "Yükleme başlıyor...")

	fileHeader, err := c.FormFile("file")
	if err != nil {
		log.Printf("❌ Dosya alınamadı: %v", err)
		progress.SendProgressUpdate(processID, 0, fmt.Sprintf("Hata: Dosya alınamadı: %v", err))
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": fmt.Sprintf("Dosya alınamadı: %v", err)})
	}

	file, err := fileHeader.Open()
	if err != nil {
		log.Printf("❌ Dosya açılamadı: %v", err)
		progress.SendProgressUpdate(processID, 0, fmt.Sprintf("Hata: Dosya açılamadı: %v", err))
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": fmt.Sprintf("Dosya açılamadı: %v", err)})
	}
	defer file.Close()

	var crewDocuments []models.CrewDocument
	recordCount := 0
	failedCount := 0
	lineNum := 0   // Başlık satırından sonraki veri satırlarını takip etmek için (1'den başlar)
	totalRows := 0 // Toplam satır sayısı tahmini (ilerleme çubuğu için daha doğru bir tahmin)

	fileExtension := strings.ToLower(filepath.Ext(fileHeader.Filename))

	// Dosya boyutuna göre kaba toplam satır sayısını tahmin et
	// Bu, progress bar için daha iyi bir başlangıç tahmini sağlar.
	const avgLineLength = 200 // Ortalama satır uzunluğu tahmini
	estimatedTotalRecords := int(fileHeader.Size) / avgLineLength
	if estimatedTotalRecords == 0 {
		estimatedTotalRecords = 1 // En az 1 kayıt varsay
	}

	switch fileExtension {
	case ".csv":
		log.Println("Handling CSV file for Crew Document...")
		reader := csv.NewReader(file)
		reader.Comma = ';' // Sizin örneğinizde ';' ayracı kullanılmış
		reader.FieldsPerRecord = -1
		reader.LazyQuotes = true

		// Tüm CSV satırlarını önceden okuyarak toplam satır sayısını al
		// Bu, ilerleme çubuğu için daha doğru bir 'totalRows' değeri sağlar.
		allRecords, err := reader.ReadAll()
		if err != nil {
			log.Printf("❌ CSV dosyası tamamı okunamadı: %v", err)
			progress.SendProgressUpdate(processID, 0, fmt.Sprintf("Hata: CSV dosyası okunamadı: %v", err))
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": fmt.Sprintf("CSV dosyası okunamadı: %v", err)})
		}

		if len(allRecords) < 2 { // Başlık satırı + en az bir veri satırı beklenir
			log.Println("⚠️ CSV dosyası boş veya sadece başlık satırı içeriyor.")
			progress.SendProgressUpdate(processID, 100, "CSV dosyası boş veya hiç veri satırı içermiyor.")
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "CSV dosyası boş veya hiç veri satırı içermiyor."})
		}

		header := allRecords[0] // Başlık satırı
		log.Printf("📌 CSV Header: %s\n", strings.Join(header, ";"))

		totalRows = len(allRecords) - 1 // Başlık satırını çıkar
		if totalRows < 0 {
			totalRows = 0
		} // Negatif olmaması için

		expectedColumnCount := 17 // CrewDocument modelindeki alan sayısı
		if len(header) < expectedColumnCount {
			log.Printf("❌ CSV başlık satırı yetersiz sütun içeriyor: %d yerine %d bekleniyor", len(header), expectedColumnCount)
			progress.SendProgressUpdate(processID, 0, fmt.Sprintf("Hata: Yetersiz sütun sayısı: %d yerine %d bekleniyor", len(header), expectedColumnCount))
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": fmt.Sprintf("CSV başlık satırı yetersiz sütun içeriyor: %d yerine %d bekleniyor", len(header), expectedColumnCount)})
		}

		// Veri satırlarını işlemeye başla (başlık hariç)
		for i, record := range allRecords[1:] {
			lineNum = i + 2 // Gerçek dosya satır numarası (başlık+1'den başlar)

			if len(record) < expectedColumnCount {
				log.Printf("⚠️ Satır %d atlandı: Yetersiz sütun sayısı (%d yerine %d bekleniyor)\n", lineNum, len(record), expectedColumnCount)
				failedCount++
				continue
			}

			// --- Veri Türü Dönüşümleri ve TrimSpace ---
			// Personel bilgileri
			personelID := strings.TrimSpace(record[0])
			if personelID == "" {
				log.Printf("❌ Satır %d, 'PersonID' alanı boş. Satır atlandı.\n", lineNum)
				failedCount++
				continue
			}

			gecerlilikBaslangicTarihi, err := parseTimestampToNullInt64(record[9])
			if err != nil {
				log.Printf("❌ Satır %d, 'gecerlilik_baslangic_tarihi' dönüşüm hatası: %v. Satır atlandı.\n", lineNum, err)
				failedCount++
				continue
			}

			gecerlilikBitisTarihi, err := parseTimestampToNullInt64(record[10])
			if err != nil {
				log.Printf("❌ Satır %d, 'gecerlilik_bitis_tarihi' dönüşüm hatası: %v. Satır atlandı.\n", lineNum, err)
				failedCount++
				continue
			}

			endDateLeaveJob, err := parseTimestampToNullInt64(record[13])
			if err != nil {
				log.Printf("❌ Satır %d, 'end_date_leave_job' dönüşüm hatası: %v. Satır atlandı.\n", lineNum, err)
				failedCount++
				continue
			}

			personelThyCalisiyorMu, err := parseBool(record[14])
			if err != nil {
				log.Printf("❌ Satır %d, 'personel_thy_calisiyor_mu' dönüşüm hatası: %v. Satır atlandı.\n", lineNum, err)
				failedCount++
				continue
			}

			dokumanGecerliMi, err := parseBool(record[15])
			if err != nil {
				log.Printf("❌ Satır %d, 'dokuman_gecerli_mi' dönüşüm hatası: %v. Satır atlandı.\n", lineNum, err)
				failedCount++
				continue
			}

			var documentNo sql.NullString
			if strings.TrimSpace(record[11]) != "" {
				documentNo = sql.NullString{String: strings.TrimSpace(record[11]), Valid: true}
			} else {
				documentNo = sql.NullString{Valid: false}
			}

			var dokumaniVeren sql.NullString
			if strings.TrimSpace(record[12]) != "" {
				dokumaniVeren = sql.NullString{String: strings.TrimSpace(record[12]), Valid: true}
			} else {
				dokumaniVeren = sql.NullString{Valid: false}
			}

			crewDocument := models.CrewDocument{
				DataID:                    uuid.New(),
				PersonID:                  personelID,
				PersonSurname:             strings.TrimSpace(record[1]),
				PersonName:                strings.TrimSpace(record[2]),
				CitizenshipNumber:         strings.TrimSpace(record[3]),
				PersonType:                strings.TrimSpace(record[4]),
				UcucuAltTipi:              strings.TrimSpace(record[5]),
				UcucuSinifi:               strings.TrimSpace(record[6]),
				BaseFilo:                  strings.TrimSpace(record[7]),
				DokumanAltTipi:            strings.TrimSpace(record[8]),
				GecerlilikBaslangicTarihi: gecerlilikBaslangicTarihi,
				GecerlilikBitisTarihi:     gecerlilikBitisTarihi,
				DocumentNo:                documentNo,
				DokumaniVeren:             dokumaniVeren,
				EndDateLeaveJob:           endDateLeaveJob,
				PersonelThyCalisiyorMu:    personelThyCalisiyorMu,
				DokumanGecerliMi:          dokumanGecerliMi,
				AgreementType:             strings.TrimSpace(record[16]),
			}
			crewDocuments = append(crewDocuments, crewDocument)
			recordCount++

			// Progress update
			progressPercent := int(float64(recordCount) / float64(totalRows) * 90) // Total read/parse phase
			if progressPercent > 90 {
				progressPercent = 90
			} // Max at 90% for this phase
			progress.SendProgressUpdate(processID, progressPercent, fmt.Sprintf("%d/%d kayıt işlendi...", recordCount, totalRows))
		}

	case ".xlsx":
		log.Println("Handling XLSX file for Crew Document...")
		f, err := excelize.OpenReader(file)
		if err != nil {
			log.Printf("❌ Excel dosyası açılamadı: %v", err)
			progress.SendProgressUpdate(processID, 0, fmt.Sprintf("Hata: Excel dosyası açılamadı: %v", err))
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": fmt.Sprintf("Excel dosyası açılamadı: %v", err)})
		}

		sheetList := f.GetSheetList()
		if len(sheetList) == 0 {
			log.Println("⚠️ Excel dosyasında hiç sayfa bulunamadı.")
			progress.SendProgressUpdate(processID, 100, "Excel dosyasına sayfa bulunamadı.")
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Excel dosyasına sayfa bulunamadı."})
		}
		sheetName := sheetList[0]
		log.Printf("ℹ️ Okunan Excel sayfası: %s\n", sheetName)

		rows, err := f.GetRows(sheetName)
		if err != nil {
			log.Printf("❌ Excel sayfasından satırlar okunamadı: %v", err)
			progress.SendProgressUpdate(processID, 0, fmt.Sprintf("Hata: Excel sayfasından veri okunamadı: %v", err))
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": fmt.Sprintf("Excel sayfasından veri okunamadı: %v", err)})
		}

		if len(rows) < 2 { // Başlık satırı + en az bir veri satırı beklenir
			log.Println("⚠️ Excel dosyası boş veya sadece başlık satırı içeriyor.")
			progress.SendProgressUpdate(processID, 100, "Excel dosyası boş veya hiç veri satırı içermiyor.")
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Excel dosyası boş veya hiç veri satırı içermiyor."})
		}

		header := rows[0]                                            // Başlık satırı
		log.Printf("📌 XLSX Header: %s\n", strings.Join(header, ";")) // CSV gibi ';' ile ayırarak logla

		totalRows = len(rows) - 1 // Başlık satırını çıkar
		if totalRows < 0 {
			totalRows = 0
		}

		expectedColumnCount := 17 // CrewDocument modelindeki alan sayısı
		if len(header) < expectedColumnCount {
			log.Printf("❌ XLSX başlık satırı yetersiz sütun içeriyor: %d yerine %d bekleniyor\n", len(header), expectedColumnCount)
			progress.SendProgressUpdate(processID, 0, fmt.Sprintf("Hata: Yetersiz sütun sayısı: %d yerine %d bekleniyor", len(header), expectedColumnCount))
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

			// --- Veri Türü Dönüşümleri ve TrimSpace ---
			personelID := strings.TrimSpace(row[0])
			if personelID == "" {
				log.Printf("❌ Satır %d, 'PersonID' alanı boş. Satır atlandı.\n", lineNum)
				failedCount++
				continue
			}

			gecerlilikBaslangicTarihi, err := parseTimestampToNullInt64(row[9])
			if err != nil {
				log.Printf("❌ Satır %d, 'gecerlilik_baslangic_tarihi' dönüşüm hatası: %v. Satır atlandı.\n", lineNum, err)
				failedCount++
				continue
			}

			gecerlilikBitisTarihi, err := parseTimestampToNullInt64(row[10])
			if err != nil {
				log.Printf("❌ Satır %d, 'gecerlilik_bitis_tarihi' dönüşüm hatası: %v. Satır atlandı.\n", lineNum, err)
				failedCount++
				continue
			}

			endDateLeaveJob, err := parseTimestampToNullInt64(row[13])
			if err != nil {
				log.Printf("❌ Satır %d, 'end_date_leave_job' dönüşüm hatası: %v. Satır atlandı.\n", lineNum, err)
				failedCount++
				continue
			}

			personelThyCalisiyorMu, err := parseBool(row[14])
			if err != nil {
				log.Printf("❌ Satır %d, 'personel_thy_calisiyor_mu' dönüşüm hatası: %v. Satır atlandı.\n", lineNum, err)
				failedCount++
				continue
			}

			dokumanGecerliMi, err := parseBool(row[15])
			if err != nil {
				log.Printf("❌ Satır %d, 'dokuman_gecerli_mi' dönüşüm hatası: %v. Satır atlandı.\n", lineNum, err)
				failedCount++
				continue
			}

			var documentNo sql.NullString
			if strings.TrimSpace(row[11]) != "" {
				documentNo = sql.NullString{String: strings.TrimSpace(row[11]), Valid: true}
			} else {
				documentNo = sql.NullString{Valid: false}
			}

			var dokumaniVeren sql.NullString
			if strings.TrimSpace(row[12]) != "" {
				dokumaniVeren = sql.NullString{String: strings.TrimSpace(row[12]), Valid: true}
			} else {
				dokumaniVeren = sql.NullString{Valid: false}
			}

			crewDocument := models.CrewDocument{
				DataID:                    uuid.New(),
				PersonID:                  personelID,
				PersonSurname:             strings.TrimSpace(row[1]),
				PersonName:                strings.TrimSpace(row[2]),
				CitizenshipNumber:         strings.TrimSpace(row[3]),
				PersonType:                strings.TrimSpace(row[4]),
				UcucuAltTipi:              strings.TrimSpace(row[5]),
				UcucuSinifi:               strings.TrimSpace(row[6]),
				BaseFilo:                  strings.TrimSpace(row[7]),
				DokumanAltTipi:            strings.TrimSpace(row[8]),
				GecerlilikBaslangicTarihi: gecerlilikBaslangicTarihi,
				GecerlilikBitisTarihi:     gecerlilikBitisTarihi,
				DocumentNo:                documentNo,
				DokumaniVeren:             dokumaniVeren,
				EndDateLeaveJob:           endDateLeaveJob,
				PersonelThyCalisiyorMu:    personelThyCalisiyorMu,
				DokumanGecerliMi:          dokumanGecerliMi,
				AgreementType:             strings.TrimSpace(row[16]),
			}
			crewDocuments = append(crewDocuments, crewDocument)
			recordCount++

			// Progress update
			progressPercent := int(float64(recordCount) / float64(totalRows) * 90)
			if progressPercent > 90 {
				progressPercent = 90
			}
			progress.SendProgressUpdate(processID, progressPercent, fmt.Sprintf("%d/%d kayıt işlendi...", recordCount, totalRows))
		}

	default:
		log.Printf("❌ Desteklenmeyen dosya uzantısı: %s", fileExtension)
		progress.SendProgressUpdate(processID, 0, fmt.Sprintf("Hata: Desteklenmeyen dosya tipi: %s", fileExtension))
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Desteklenmeyen dosya tipi. Lütfen .csv veya .xlsx dosyası yükleyin."})
	}

	if recordCount == 0 {
		log.Println("⚠️ Dosyada işlenecek hiç veri satırı bulunamadı (başlık hariç).")
		progress.SendProgressUpdate(processID, 100, "Dosyada boş veya hiç veri satırı içermiyor.")
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Dosyada boş veya hiç veri satırı içermiyor."})
	}

	log.Printf("🚀 %d adet crew_document kaydı veritabanına ekleniyor...\n", recordCount)

	// Phase: Delete Old Data (90-95%)
	deleteProgressStart := 90
	deleteProgressEnd := 95
	if c.QueryBool("reset", false) {
		log.Println("🚀 'reset=true' parametresi algılandı, mevcut ekip dokümanları temizleniyor...")
		progress.SendProgressUpdate(processID, deleteProgressStart, "Mevcut veriler temizleniyor...")
		_, err := db.DB.NewDelete().
			Model(&models.CrewDocument{}).
			Where("TRUE").
			Exec(context.Background())
		if err != nil {
			log.Printf("❌ Mevcut ekip dokümanları temizlenirken hata oluştu: %v", err)
			progress.SendProgressUpdate(processID, 100, fmt.Sprintf("Hata: Mevcut veriler temizlenirken hata oluştu: %v", err))
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": fmt.Sprintf("Mevcut veriler temizlenirken hata oluştu: %v", err)})
		}
		log.Println("✅ Mevcut ekip dokümanları başarıyla temizlendi.")
		progress.SendProgressUpdate(processID, deleteProgressEnd, "Mevcut veriler temizlendi.")
	} else {
		// Reset yapılmadıysa bu aşamayı atla, progress'i de atla
		deleteProgressStart = 95
		deleteProgressEnd = 95
	}

	// Phase: Insert New Data (95-99%)
	insertProgressStart := deleteProgressEnd
	// İlerleme çubuğunun 99'a kadar gitmesi için yüzde hesaplaması
	insertProgressPerRecord := float64(4) / float64(recordCount) // %4'lük dilim (95'ten 99'a)
	progress.SendProgressUpdate(processID, insertProgressStart, fmt.Sprintf("%d kayıt veritabanına ekleniyor...", recordCount))

	// ⭐ ON CONFLICT ifadesi kaldırıldı
	// DataID hariç unique kısıtlama olmadığı için ON CONFLICT kullanılmaz.
	// Yeni kayıtları ekler.
	tx, err := db.DB.BeginTx(context.Background(), nil) // İşlem başlat
	if err != nil {
		log.Printf("❌ İşlem başlatılırken hata: %v", err)
		progress.SendProgressUpdate(processID, 100, fmt.Sprintf("Hata: İşlem başlatılırken hata: %v", err))
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": fmt.Sprintf("İşlem başlatılırken hata: %v", err)})
	}
	defer tx.Rollback() // Hata olursa geri al

	// Her 1000 kayıtta bir toplu ekleme yap (performans için)
	batchSize := 1000
	for i := 0; i < len(crewDocuments); i += batchSize {
		end := i + batchSize
		if end > len(crewDocuments) {
			end = len(crewDocuments)
		}
		batch := crewDocuments[i:end]

		_, err := tx.NewInsert().
			Model(&batch).
			Exec(context.Background())
		if err != nil {
			log.Printf("❌ Veritabanına toplu ekleme hatası (Batch %d-%d): %v", i, end, err)
			progress.SendProgressUpdate(processID, 100, fmt.Sprintf("Hata: Veritabanına toplu ekleme hatası: %v", err))
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"success": recordCount, "failed": failedCount, "error": fmt.Sprintf("Veritabanına ekleme hatası: %v", err)})
		}
		// Batch progress update
		currentProgress := insertProgressStart + int(float64(i+batchSize)/float64(recordCount)*insertProgressPerRecord)
		if currentProgress > 99 {
			currentProgress = 99
		}
		progress.SendProgressUpdate(processID, currentProgress, fmt.Sprintf("%d/%d kayıt eklendi...", i+batchSize, recordCount))
	}

	if err := tx.Commit(); err != nil { // İşlemi onayla
		log.Printf("❌ İşlem onaylanırken hata: %v", err)
		progress.SendProgressUpdate(processID, 100, fmt.Sprintf("Hata: İşlem onaylanırken hata: %v", err))
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": fmt.Sprintf("İşlem onaylanırken hata: %v", err)})
	}

	progress.SendProgressUpdate(processID, 100, fmt.Sprintf("%d kayıt başarıyla eklendi.", recordCount))
	log.Printf("✅ %d adet crew_document kaydı başarıyla eklendi. %d kayıt atlandı.\n", recordCount, failedCount)
	return c.Status(fiber.StatusOK).JSON(fiber.Map{"success": recordCount, "failed": failedCount, "message": fmt.Sprintf("%d kayıt başarıyla eklendi. %d kayıt atlandı.", recordCount, failedCount)})
}
