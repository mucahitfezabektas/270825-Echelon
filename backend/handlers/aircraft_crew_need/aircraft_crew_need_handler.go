package aircraft_crew_need

import (
	"context"
	"encoding/csv" // CSV işlemleri için kalacak
	"fmt"
	"io"
	"log"
	"path/filepath" // Dosya uzantısını almak için eklendi
	"strconv"       // string'den int32 dönüşümü için
	"strings"

	"mini_CMS_Desktop_App/db"
	"mini_CMS_Desktop_App/models" // models paketini import ettiğinizden emin olun

	"github.com/gofiber/fiber/v2"
	"github.com/xuri/excelize/v2" // ✅ Excelize kütüphanesi eklendi
)

// ImportAircraftCrewNeedData, aircraft_crew_need tablosuna hem CSV hem de XLSX verisi aktarır.
// Fonksiyon adı ImportAircraftCrewNeedCSV'den ImportAircraftCrewNeedData olarak değiştirildi.
func ImportAircraftCrewNeedData(c *fiber.Ctx) error {
	log.Println("🔍 ImportAircraftCrewNeedData çağrıldı")

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

	var aircraftCrewNeedEntries []models.AircraftCrewNeed
	recordCount := 0
	failedCount := 0
	lineNum := 0 // Başlık satırından sonraki veri satırlarını takip etmek için (1'den başlar)

	// Dosya uzantısına göre okuma stratejisi belirle
	fileExtension := strings.ToLower(filepath.Ext(fileHeader.Filename))

	switch fileExtension {
	case ".csv":
		log.Println("Handling CSV file for Aircraft Crew Need...")
		reader := csv.NewReader(file)
		reader.Comma = '|' // Örnek tablonuzda '|' ayracı kullanıldığı için
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
		log.Printf("📌 CSV Header: %s\n", strings.Join(header, "|"))

		expectedColumnCount := 10 // AircraftCrewNeed modelindeki alan sayısı (DataID hariç: actype + 9 pozisyon sayısı)
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

			// Veri Türü Dönüşümleri ve TrimSpace
			actype := strings.TrimSpace(record[0])
			if actype == "" {
				log.Printf("❌ Satır %d, 'actype' alanı boş. Satır atlandı.\n", lineNum+1)
				failedCount++
				continue
			}

			cCount, err := parseInt32(strings.TrimSpace(record[1]))
			if err != nil {
				log.Printf("❌ Satır %d, 'C' dönüşüm hatası: %v. Satır atlandı.\n", lineNum+1, err)
				failedCount++
				continue
			}
			pCount, err := parseInt32(strings.TrimSpace(record[2]))
			if err != nil {
				log.Printf("❌ Satır %d, 'P' dönüşüm hatası: %v. Satır atlandı.\n", lineNum+1, err)
				failedCount++
				continue
			}
			jCount, err := parseInt32(strings.TrimSpace(record[3]))
			if err != nil {
				log.Printf("❌ Satır %d, 'J' dönüşüm hatası: %v. Satır atlandı.\n", lineNum+1, err)
				failedCount++
				continue
			}
			efCount, err := parseInt32(strings.TrimSpace(record[4]))
			if err != nil {
				log.Printf("❌ Satır %d, 'EF' dönüşüm hatası: %v. Satır atlandı.\n", lineNum+1, err)
				failedCount++
				continue
			}
			aCount, err := parseInt32(strings.TrimSpace(record[5]))
			if err != nil {
				log.Printf("❌ Satır %d, 'A' dönüşüm hatası: %v. Satır atlandı.\n", lineNum+1, err)
				failedCount++
				continue
			}
			sCount, err := parseInt32(strings.TrimSpace(record[6]))
			if err != nil {
				log.Printf("❌ Satır %d, 'S' dönüşüm hatası: %v. Satır atlandı.\n", lineNum+1, err)
				failedCount++
				continue
			}
			lCount, err := parseInt32(strings.TrimSpace(record[7]))
			if err != nil {
				log.Printf("❌ Satır %d, 'L' dönüşüm hatası: %v. Satır atlandı.\n", lineNum+1, err)
				failedCount++
				continue
			}
			ecCount, err := parseInt32(strings.TrimSpace(record[8]))
			if err != nil {
				log.Printf("❌ Satır %d, 'EC' dönüşüm hatası: %v. Satır atlandı.\n", lineNum+1, err)
				failedCount++
				continue
			}
			tCount, err := parseInt32(strings.TrimSpace(record[9]))
			if err != nil {
				log.Printf("❌ Satır %d, 'T' dönüşüm hatası: %v. Satır atlandı.\n", lineNum+1, err)
				failedCount++
				continue
			}

			aircraftCrewNeedEntry := models.AircraftCrewNeed{
				Actype:   actype,
				C_Count:  cCount,
				P_Count:  pCount,
				J_Count:  jCount,
				EF_Count: efCount,
				A_Count:  aCount,
				S_Count:  sCount,
				L_Count:  lCount,
				EC_Count: ecCount,
				T_Count:  tCount,
			}
			aircraftCrewNeedEntries = append(aircraftCrewNeedEntries, aircraftCrewNeedEntry)
			recordCount++
		}

	case ".xlsx":
		log.Println("Handling XLSX file for Aircraft Crew Need...")
		f, err := excelize.OpenReader(file)
		if err != nil {
			log.Printf("❌ Excel dosyası açılamadı: %v", err)
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": fmt.Sprintf("Excel dosyası açılamadı: %v", err)})
		}

		sheetList := f.GetSheetList()
		if len(sheetList) == 0 {
			log.Println("⚠️ Excel dosyasında hiç sayfa bulunamadı.")
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
		log.Printf("📌 XLSX Header: %s\n", strings.Join(header, "|")) // CSV gibi "|" ile ayırarak logla

		expectedColumnCount := 10 // AircraftCrewNeed modelindeki alan sayısı
		if len(header) < expectedColumnCount {
			log.Printf("❌ XLSX başlık satırı yetersiz sütun içeriyor: %d yerine %d bekleniyor\n", len(header), expectedColumnCount)
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": fmt.Sprintf("XLSX başlık satırı yetersiz sütun içeriyor: %d yerine %d bekleniyor", len(header), expectedColumnCount)})
		}

		for i, row := range rows {
			if i == 0 { // Başlık satırını atla
				continue
			}
			lineNum = i + 1 // Gerçek Excel satır numarası (1 tabanlı)

			if len(row) == 0 || strings.Join(row, "") == "" {
				continue
			}

			if len(row) < expectedColumnCount {
				log.Printf("⚠️ Satır %d atlandı: Yetersiz sütun sayısı (%d yerine %d bekleniyor)\n", lineNum, len(row), expectedColumnCount)
				failedCount++
				continue
			}

			// Veri Türü Dönüşümleri ve TrimSpace (Excel'den okunanlar da TrimSpace yapılmalı)
			actype := strings.TrimSpace(row[0])
			if actype == "" {
				log.Printf("❌ Satır %d, 'actype' alanı boş. Satır atlandı.\n", lineNum)
				failedCount++
				continue
			}

			cCount, err := parseInt32(strings.TrimSpace(row[1]))
			if err != nil {
				log.Printf("❌ Satır %d, 'C' dönüşüm hatası: %v. Satır atlandı.\n", lineNum, err)
				failedCount++
				continue
			}
			pCount, err := parseInt32(strings.TrimSpace(row[2]))
			if err != nil {
				log.Printf("❌ Satır %d, 'P' dönüşüm hatası: %v. Satır atlandı.\n", lineNum, err)
				failedCount++
				continue
			}
			jCount, err := parseInt32(strings.TrimSpace(row[3]))
			if err != nil {
				log.Printf("❌ Satır %d, 'J' dönüşüm hatası: %v. Satır atlandı.\n", lineNum, err)
				failedCount++
				continue
			}
			efCount, err := parseInt32(strings.TrimSpace(row[4]))
			if err != nil {
				log.Printf("❌ Satır %d, 'EF' dönüşüm hatası: %v. Satır atlandı.\n", lineNum, err)
				failedCount++
				continue
			}
			aCount, err := parseInt32(strings.TrimSpace(row[5]))
			if err != nil {
				log.Printf("❌ Satır %d, 'A' dönüşüm hatası: %v. Satır atlandı.\n", lineNum, err)
				failedCount++
				continue
			}
			sCount, err := parseInt32(strings.TrimSpace(row[6]))
			if err != nil {
				log.Printf("❌ Satır %d, 'S' dönüşüm hatası: %v. Satır atlandı.\n", lineNum, err)
				failedCount++
				continue
			}
			lCount, err := parseInt32(strings.TrimSpace(row[7]))
			if err != nil {
				log.Printf("❌ Satır %d, 'L' dönüşüm hatası: %v. Satır atlandı.\n", lineNum, err)
				failedCount++
				continue
			}
			ecCount, err := parseInt32(strings.TrimSpace(row[8]))
			if err != nil {
				log.Printf("❌ Satır %d, 'EC' dönüşüm hatası: %v. Satır atlandı.\n", lineNum, err)
				failedCount++
				continue
			}
			tCount, err := parseInt32(strings.TrimSpace(row[9]))
			if err != nil {
				log.Printf("❌ Satır %d, 'T' dönüşüm hatası: %v. Satır atlandı.\n", lineNum, err)
				failedCount++
				continue
			}

			aircraftCrewNeedEntry := models.AircraftCrewNeed{
				Actype:   actype,
				C_Count:  cCount,
				P_Count:  pCount,
				J_Count:  jCount,
				EF_Count: efCount,
				A_Count:  aCount,
				S_Count:  sCount,
				L_Count:  lCount,
				EC_Count: ecCount,
				T_Count:  tCount,
			}
			aircraftCrewNeedEntries = append(aircraftCrewNeedEntries, aircraftCrewNeedEntry)
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

	log.Printf("🚀 %d adet Aircraft Crew Need kaydı veritabanına ekleniyor...\n", recordCount)

	if c.QueryBool("reset", false) {
		log.Println("🚀 'reset=true' parametresi algılandı, mevcut Aircraft Crew Need bilgileri temizleniyor...")
		_, err := db.DB.NewDelete().
			Model(&models.AircraftCrewNeed{}).
			Where("TRUE"). // Tüm kayıtları silmek için
			Exec(context.Background())
		if err != nil {
			log.Printf("❌ Mevcut Aircraft Crew Need bilgileri temizlenirken hata oluştu: %v", err)
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": fmt.Sprintf("Mevcut veriler temizlenirken hata oluştu: %v", err)})
		}
		log.Println("✅ Mevcut Aircraft Crew Need bilgileri başarıyla temizlendi.")
	}

	// ⭐⭐⭐ BURADAKİ DEĞİŞİKLİK: ON CONFLICT ifadesi kaldırıldı ⭐⭐⭐
	// DataID hariç unique kısıtlama olmadığı için ON CONFLICT kullanılmaz.
	// Yeni kayıtları ekler.
	_, err = db.DB.NewInsert().
		Model(&aircraftCrewNeedEntries).
		Exec(context.Background())
	if err != nil {
		log.Printf("❌ Veritabanına ekleme hatası: %v", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"success": recordCount, "failed": failedCount, "error": fmt.Sprintf("Veritabanına ekleme hatası: %v", err)})
	}

	log.Printf("✅ %d adet Aircraft Crew Need kaydı başarıyla eklendi. %d kayıt atlandı.\n", recordCount, failedCount)
	return c.Status(fiber.StatusOK).JSON(fiber.Map{"success": recordCount, "failed": failedCount, "message": fmt.Sprintf("%d kayıt başarıyla eklendi. %d kayıt atlandı.", recordCount, failedCount)})
}

// parseInt32, string'i int32'ye dönüştürür. Boş stringler için 0 döndürür.
func parseInt32(s string) (int32, error) {
	if s == "" {
		return 0, nil // Boş string ise 0 döndür, hata verme
	}
	val, err := strconv.ParseInt(s, 10, 32) // 32 bit integer için
	if err != nil {
		return 0, fmt.Errorf("int32 parse hatası: %w", err)
	}
	return int32(val), nil
}
