package crew_info

import (
	"context"
	"encoding/csv"
	"fmt"
	"io"
	"log"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"mini_CMS_Desktop_App/db"
	"mini_CMS_Desktop_App/models"

	"github.com/gofiber/fiber/v2"
	"github.com/xuri/excelize/v2"
)

// parseTimestamp, string veya Excel sayısal formatındaki tarihleri int64 timestamp'e (milisaniye) dönüştürür.
func parseTimestamp(s string) (int64, error) {
	s = strings.TrimSpace(s)
	if s == "" {
		return 0, nil
	}

	val, err := strconv.ParseInt(s, 10, 64)
	if err == nil {
		return val, nil
	}

	fVal, err := strconv.ParseFloat(s, 64)
	if err == nil {
		t, err := excelize.ExcelDateToTime(fVal, false)
		if err == nil {
			return t.UnixNano() / int64(time.Millisecond), nil
		}
	}

	layouts := []string{
		"02/01/2006 15:04:05",
		"02/01/2006",
		"02.01.2006", // GG.AA.YYYY formatı
		"2006-01-02 15:04:05",
		"2006-01-02",
		"01-02-2006 15:04:05",
		"01-02-2006",
		"1/2/2006 15:04:05",
		"1/2/2006",
		"2006/01/02 15:04:05",
		"2006/01/02",
	}

	for _, layout := range layouts {
		t, err := time.Parse(layout, s)
		if err == nil {
			return t.UnixNano() / int64(time.Millisecond), nil
		}
	}

	return 0, fmt.Errorf("geçersiz tarih veya timestamp formatı: '%s'", s)
}

// parseBool, string'i boolean'a dönüştürür.
// ✅ GÜNCELLENDİ: 'Y' ve 'N' değerleri eklendi.
func parseBool(s string) (bool, error) {
	s = strings.TrimSpace(strings.ToLower(s))
	switch s {
	case "true", "1", "calisiyor", "evet", "yes", "gecerli", "y": // 'y' eklendi
		return true, nil
	case "false", "0", "calismiyor", "hayir", "no", "gecerli degil", "n", "": // 'n' ve boş string eklendi
		return false, nil
	default:
		return false, fmt.Errorf("bilinmeyen boolean değeri: '%s'", s)
	}
}

// ImportCrewInfoData, crew_info tablosuna hem CSV hem de XLSX verisi aktarır.
func ImportCrewInfoData(c *fiber.Ctx) error {
	log.Println("🔍 ImportCrewInfoData çağrıldı")

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

	var crewInfoEntries []models.CrewInfo
	recordCount := 0
	failedCount := 0
	lineNum := 0

	fileExtension := strings.ToLower(filepath.Ext(fileHeader.Filename))

	switch fileExtension {
	case ".csv":
		log.Println("Handling CSV file for Crew Info...")
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

		expectedColumnCount := 27 // CrewInfo modelindeki alan sayısı (DataID hariç)
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

			// --- Veri Türü Dönüşümleri ve TrimSpace ---
			dogumTarihi, err := parseTimestamp(strings.TrimSpace(record[6]))
			if err != nil {
				log.Printf("❌ Satır %d, 'dogum_tarihi' dönüşüm hatası: %v. Satır atlandı.\n", lineNum+1, err)
				failedCount++
				continue
			}
			rankChangeDate, err := parseTimestamp(strings.TrimSpace(record[11]))
			if err != nil {
				log.Printf("❌ Satır %d, 'rank_change_date' dönüşüm hatası: %v. Satır atlandı.\n", lineNum+1, err)
				failedCount++
				continue
			}
			jobStartDate, err := parseTimestamp(strings.TrimSpace(record[15]))
			if err != nil {
				log.Printf("❌ Satır %d, 'job_start_date' dönüşüm hatası: %v. Satır atlandı.\n", lineNum+1, err)
				failedCount++
				continue
			}
			jobEndDate, err := parseTimestamp(strings.TrimSpace(record[16]))
			if err != nil {
				log.Printf("❌ Satır %d, 'job_end_date' dönüşüm hatası: %v. Satır atlandı.\n", lineNum+1, err)
				failedCount++
				continue
			}
			marriageDate, err := parseTimestamp(strings.TrimSpace(record[17]))
			if err != nil {
				log.Printf("❌ Satır %d, 'marriage_date' dönüşüm hatası: %v. Satır atlandı.\n", lineNum+1, err)
				failedCount++
				continue
			}

			personThyCalisiyorMu, err := parseBool(strings.TrimSpace(record[21]))
			if err != nil {
				log.Printf("❌ Satır %d, 'person_thy_calisiyor_mu' dönüşüm hatası: %v. Satır atlandı.\n", lineNum+1, err)
				failedCount++
				continue
			}
			serviceUseHomePickup, err := parseBool(strings.TrimSpace(record[24]))
			if err != nil {
				log.Printf("❌ Satır %d, 'service_use_home_pickup' dönüşüm hatası: %v. Satır atlandı.\n", lineNum+1, err)
				failedCount++
				continue
			}
			serviceUseSaw, err := parseBool(strings.TrimSpace(record[25]))
			if err != nil {
				log.Printf("❌ Satır %d, 'service_use_saw' dönüşüm hatası: %v. Satır atlandı.\n", lineNum+1, err)
				failedCount++
				continue
			}
			bridgeUse, err := parseBool(strings.TrimSpace(record[26]))
			if err != nil {
				log.Printf("❌ Satır %d, 'bridge_use' dönüşüm hatası: %v. Satır atlandı.\n", lineNum+1, err)
				failedCount++
				continue
			}

			crewInfo := models.CrewInfo{
				PersonID:                 strings.TrimSpace(record[0]),
				PersonSurname:            strings.TrimSpace(record[1]),
				PersonName:               strings.TrimSpace(record[2]),
				Gender:                   strings.TrimSpace(record[3]),
				Tabiiyet:                 strings.TrimSpace(record[4]),
				BaseFilo:                 strings.TrimSpace(record[5]),
				DogumTarihi:              dogumTarihi,
				BaseLocation:             strings.TrimSpace(record[7]),
				UcucuTipi:                strings.TrimSpace(record[8]),
				OML:                      strings.TrimSpace(record[9]),
				Seniority:                strings.TrimSpace(record[10]),
				RankChangeDate:           rankChangeDate,
				Rank:                     strings.TrimSpace(record[12]),
				AgreementType:            strings.TrimSpace(record[13]),
				AgreementTypeExplanation: strings.TrimSpace(record[14]),
				JobStartDate:             jobStartDate,
				JobEndDate:               jobEndDate,
				MarriageDate:             marriageDate,
				UcucuSinifi:              strings.TrimSpace(record[18]),
				UcucuSinifiLastValid:     strings.TrimSpace(record[19]),
				UcucuAltTipi:             strings.TrimSpace(record[20]),
				PersonThyCalisiyorMu:     personThyCalisiyorMu,
				Birthplace:               strings.TrimSpace(record[22]),
				PeriodInfo:               strings.TrimSpace(record[23]),
				ServiceUseHomePickup:     serviceUseHomePickup,
				ServiceUseSaw:            serviceUseSaw,
				BridgeUse:                bridgeUse,
			}
			crewInfoEntries = append(crewInfoEntries, crewInfo)
			recordCount++
		}

	case ".xlsx":
		log.Println("Handling XLSX file for Crew Info...")
		f, err := excelize.OpenReader(file)
		if err != nil {
			log.Printf("❌ Excel dosyası açılamadı: %v", err)
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": fmt.Sprintf("Excel dosyası açılamadı: %v", err)})
		}

		sheetList := f.GetSheetList()
		if len(sheetList) == 0 {
			log.Println("⚠️ Excel dosyada hiç sayfa bulunamadı.")
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Excel dosyasına sayfa bulunamadı."})
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

		header := rows[0]                                            // Başlık satırı
		log.Printf("📌 XLSX Header: %s\n", strings.Join(header, ",")) // Excel'de virgül ayracı loglanır

		expectedColumnCount := 27
		if len(header) < expectedColumnCount {
			log.Printf("❌ XLSX başlık satırı yetersiz sütun içeriyor: %d yerine %d bekleniyor\n", len(header), expectedColumnCount)
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": fmt.Sprintf("XLSX başlık satırı yetersiz sütun içeriyor: %d yerine %d bekleniyor", len(header), expectedColumnCount)})
		}

		for i, row := range rows {
			if i == 0 { // Başlık satırını atla
				continue
			}
			lineNum = i + 1 // Gerçek Excel satır numarası (1 tabanlı)

			if len(row) == 0 || (len(row) > 0 && strings.Join(row, "") == "") { // Tamamen boş satırları atla
				continue
			}

			record := make([]string, expectedColumnCount)
			copy(record, row) // row'daki mevcut değerleri kopyala. Eksikse boş string olur.

			// --- Veri Türü Dönüşümleri ve TrimSpace ---
			dogumTarihi, err := parseTimestamp(strings.TrimSpace(record[6]))
			if err != nil {
				log.Printf("❌ Satır %d, 'dogum_tarihi' dönüşüm hatası: %v. Satır atlandı.\n", lineNum, err)
				failedCount++
				continue
			}
			rankChangeDate, err := parseTimestamp(strings.TrimSpace(record[11]))
			if err != nil {
				log.Printf("❌ Satır %d, 'rank_change_date' dönüşüm hatası: %v. Satır atlandı.\n", lineNum, err)
				failedCount++
				continue
			}
			jobStartDate, err := parseTimestamp(strings.TrimSpace(record[15]))
			if err != nil {
				log.Printf("❌ Satır %d, 'job_start_date' dönüşüm hatası: %v. Satır atlandı.\n", lineNum, err)
				failedCount++
				continue
			}
			jobEndDate, err := parseTimestamp(strings.TrimSpace(record[16]))
			if err != nil {
				log.Printf("❌ Satır %d, 'job_end_date' dönüşüm hatası: %v. Satır atlandı.\n", lineNum, err)
				failedCount++
				continue
			}
			marriageDate, err := parseTimestamp(strings.TrimSpace(record[17]))
			if err != nil {
				log.Printf("❌ Satır %d, 'marriage_date' dönüşüm hatası: %v. Satır atlandı.\n", lineNum, err)
				failedCount++
				continue
			}

			personThyCalisiyorMu, err := parseBool(strings.TrimSpace(record[21]))
			if err != nil {
				log.Printf("❌ Satır %d, 'person_thy_calisiyor_mu' dönüşüm hatası: %v. Satır atlandı.\n", lineNum, err)
				failedCount++
				continue
			}
			serviceUseHomePickup, err := parseBool(strings.TrimSpace(record[24]))
			if err != nil {
				log.Printf("❌ Satır %d, 'service_use_home_pickup' dönüşüm hatası: %v. Satır atlandı.\n", lineNum, err)
				failedCount++
				continue
			}
			serviceUseSaw, err := parseBool(strings.TrimSpace(record[25]))
			if err != nil {
				log.Printf("❌ Satır %d, 'service_use_saw' dönüşüm hatası: %v. Satır atlandı.\n", lineNum, err)
				failedCount++
				continue
			}
			bridgeUse, err := parseBool(strings.TrimSpace(record[26]))
			if err != nil {
				log.Printf("❌ Satır %d, 'bridge_use' dönüşüm hatası: %v. Satır atlandı.\n", lineNum, err)
				failedCount++
				continue
			}

			crewInfo := models.CrewInfo{
				PersonID:                 strings.TrimSpace(record[0]),
				PersonSurname:            strings.TrimSpace(record[1]),
				PersonName:               strings.TrimSpace(record[2]),
				Gender:                   strings.TrimSpace(record[3]),
				Tabiiyet:                 strings.TrimSpace(record[4]),
				BaseFilo:                 strings.TrimSpace(record[5]),
				DogumTarihi:              dogumTarihi,
				BaseLocation:             strings.TrimSpace(record[7]),
				UcucuTipi:                strings.TrimSpace(record[8]),
				OML:                      strings.TrimSpace(record[9]),
				Seniority:                strings.TrimSpace(record[10]), // ✅ GÜNCELLEDİK: int32 olarak atanıyor
				RankChangeDate:           rankChangeDate,
				Rank:                     strings.TrimSpace(record[12]),
				AgreementType:            strings.TrimSpace(record[13]),
				AgreementTypeExplanation: strings.TrimSpace(record[14]),
				JobStartDate:             jobStartDate,
				JobEndDate:               jobEndDate,
				MarriageDate:             marriageDate,
				UcucuSinifi:              strings.TrimSpace(record[18]),
				UcucuSinifiLastValid:     strings.TrimSpace(record[19]),
				UcucuAltTipi:             strings.TrimSpace(record[20]),
				PersonThyCalisiyorMu:     personThyCalisiyorMu,
				Birthplace:               strings.TrimSpace(record[22]),
				PeriodInfo:               strings.TrimSpace(record[23]),
				ServiceUseHomePickup:     serviceUseHomePickup,
				ServiceUseSaw:            serviceUseSaw,
				BridgeUse:                bridgeUse,
			}
			crewInfoEntries = append(crewInfoEntries, crewInfo)
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

	log.Printf("🚀 %d adet crew_info kaydı veritabanına ekleniyor...\n", recordCount)

	if c.QueryBool("reset", false) {
		log.Println("🚀 'reset=true' parametresi algılandı, mevcut Ekip Bilgileri temizleniyor...")
		_, err := db.DB.NewDelete().
			Model(&models.CrewInfo{}).
			Where("TRUE").
			Exec(context.Background())
		if err != nil {
			log.Printf("❌ Mevcut Ekip Bilgileri temizlenirken hata oluştu: %v", err)
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": fmt.Sprintf("Mevcut veriler temizlenirken hata oluştu: %v", err)})
		}
		log.Println("✅ Mevcut Ekip Bilgileri başarıyla temizlendi.")
	}

	// ON CONFLICT ifadesi kaldırıldı (istek üzerine)
	// DataID hariç unique kısıtlama olmadığı için ON CONFLICT kullanılmaz.
	// Yeni kayıtları ekler.
	_, err = db.DB.NewInsert().
		Model(&crewInfoEntries).
		Exec(context.Background())
	if err != nil {
		log.Printf("❌ Veritabanına ekleme hatası: %v", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"success": recordCount, "failed": failedCount, "error": fmt.Sprintf("Veritabanına ekleme hatası: %v", err)})
	}

	log.Printf("✅ %d adet crew_info kaydı başarıyla eklendi. %d kayıt atlandı.\n", recordCount, failedCount)
	return c.Status(fiber.StatusOK).JSON(fiber.Map{"success": recordCount, "failed": failedCount, "message": fmt.Sprintf("%d kayıt başarıyla eklendi. %d kayıt atlandı.", recordCount, failedCount)})
}
