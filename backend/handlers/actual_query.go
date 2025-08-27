// handlers/actual_query.go
package handlers

import (
	"context"
	"encoding/json" // JSON işlemleri için eklendi
	"fmt"
	"mini_CMS_Desktop_App/db"
	"mini_CMS_Desktop_App/models"
	"net/url" // url.Values ve ParseQuery için hala gerekli
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/uptrace/bun"
)

// shortToFullField kısaltmalı komutlar için hala kullanılacak
var shortToFullField = map[string]string{
	"c":  "person_id",
	"s":  "surname",
	"a":  "activity_code",
	"cl": "class",
	"dp": "departure_port",
	"ap": "arrival_port",
	"d":  "date",
	"t":  "trip_id",
	"pt": "plane_tail_name",
	"pc": "plane_cms_type",
	"gc": "group_code",
	"fp": "flight_position",
	"fn": "flight_no",
	"at": "agreement_type",
	"fi": "ucus_id", // Flight ID için
}

// isValidDate fonksiyonu, verilen string'in YYYY-MM-DD formatında geçerli bir tarih olup olmadığını kontrol eder.
func isValidDate(s string) bool {
	_, err := time.Parse("2006-01-02", s)
	return err == nil
}

// parseAbbreviatedQuery kısaltmalı komut stringini url.Values formatına dönüştürür.
// Bu fonksiyon artık QueryActualData içinde çağrılacak.
func parseAbbreviatedQuery(input string) (url.Values, error) {
	tokens := strings.Fields(input)
	values := make(url.Values) // url.Values döndür

	for i := 0; i < len(tokens); {
		key := tokens[i]
		fullKey, ok := shortToFullField[key]
		if !ok {
			return nil, fmt.Errorf("BİLİNMEYEN ALAN KISALTMASI: %s", key)
		}

		// Tarih aralığı desteği: "d 2025-07-01 2025-07-15"
		if key == "d" && i+2 < len(tokens) {
			date1 := tokens[i+1]
			date2 := tokens[i+2]
			if isValidDate(date1) && isValidDate(date2) {
				values.Add(fullKey, date1) // İlk tarihi ekle
				values.Add(fullKey, date2) // İkinci tarihi de aynı key ile ekle
				i += 3
				continue
			}
		}

		// Normal key-value çifti kontrolü
		if i+1 < len(tokens) {
			values.Add(fullKey, tokens[i+1])
			i += 2
		} else {
			return nil, fmt.Errorf("GEÇERSİZ SORGU FORMATI: ÇİFT SAYIDA ANAHTAR-DEĞER BEKLENİYOR")
		}
	}
	return values, nil
}

// QueryActualData, hem URL query parametreleri (kısaltmalı) hem de JSON body (SavedFilter) kabul edebilir.
func QueryActualData(c *fiber.Ctx) error {
	var filterData models.SavedFilter
	var queryParams url.Values
	var err error
	var isJsonRequest bool

	// 1. JSON Body'yi deneme (FilterQueryWindow'dan gelecek)
	// Eğer request Content-Type: application/json ise veya body boş değilse JSON parse etmeyi dene.
	if strings.Contains(c.Get("Content-Type"), "application/json") || len(c.Body()) > 0 {
		err = json.Unmarshal(c.Body(), &filterData)
		if err == nil {
			isJsonRequest = true
		} else if !strings.Contains(err.Error(), "unexpected end of JSON input") && !strings.Contains(err.Error(), "unsupported Media Type") && !strings.Contains(err.Error(), "no content") {
			// Sadece gerçek JSON parsing hatalarını logla/döndür, boş body veya yanlış Content-Type değil
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "JSON filtre verisi okunamadı", "details": err.Error()})
		}
	}

	// 2. Eğer JSON body başarılı değilse veya yoksa, URL Query parametrelerini kontrol et (Timeline search bar'dan gelecek)
	if !isJsonRequest {
		rawInput := c.Query("q") // Örn: "c 109403 a FLT"
		if rawInput != "" {
			queryParams, err = parseAbbreviatedQuery(rawInput)
			if err != nil {
				return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
			}
		}
		// "q" parametresi boşsa veya parseAbbreviatedQuery boş url.Values döndürdüyse
		// filterData.Rows da boş kalacaktır, bu da tüm veriyi veya boş bir seti döndürmeye yol açar.
	}

	// Sorguyu başlat ve varsayılan sıralamayı uygula
	q := db.DB.NewSelect().Model((*models.Actual)(nil))
	q = q.Order("person_id ASC", "departure_time ASC")

	// Eğer "q" parametresiyle gelip herhangi bir filtre yoksa
	if !isJsonRequest && (queryParams == nil || len(queryParams) == 0) {
		// Ancak "q" parametresi bile yoksa veya boşsa, boş bir sonuç döndür (frontend'deki "hiçbir şey bulunamadı" mesajı için)
		// NOT: Eğer bu durumda tüm veriyi döndürmek isterseniz, bu if bloğunu kaldırın.
		return c.Status(fiber.StatusOK).JSON(fiber.Map{
			"total":  0,
			"result": []models.Actual{},
		})
	}

	// Filtreleri uygulama mantığı
	if isJsonRequest {
		// JSON body'den gelen SavedFilter'ı kullan
		applySavedFilterToQuery(q, filterData)
	} else {
		// URL query parametrelerinden gelen filtreleri kullan (tekrar oluşturulur)
		applyQueryParamsToQuery(q, queryParams)
	}

	var results []models.Actual
	err = q.Scan(context.Background(), &results) // err yeniden kullanılıyor
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Sorgu yürütme başarısız", "details": err.Error()})
	}

	total := len(results)

	return c.JSON(fiber.Map{
		"total":  total,
		"result": results,
	})
}

// applySavedFilterToQuery, models.SavedFilter yapısını Bun sorgusuna uygular.
func applySavedFilterToQuery(q *bun.SelectQuery, filterData models.SavedFilter) *bun.SelectQuery {
	if len(filterData.Rows) == 0 {
		return q // Filtre satırı yoksa sorguyu değiştirmeden dön
	}

	if filterData.Logic == "OR" {
		q = q.WhereGroup(" OR ", func(sq *bun.SelectQuery) *bun.SelectQuery {
			for _, row := range filterData.Rows {
				if row.Field == "" || row.Value == "" {
					continue
				}
				sq = applyFilterCondition(sq, row) // Her koşulu WhereOr ile ekle
			}
			return sq
		})
	} else { // Varsayılan olarak "AND"
		for _, row := range filterData.Rows {
			if row.Field == "" || row.Value == "" {
				continue
			}
			q = applyFilterCondition(q, row) // Her koşulu Where ile ekle
		}
	}
	return q
}

// applyQueryParamsToQuery, url.Values yapısını Bun sorgusuna uygular.
// Bu, kısaltmalı sorgudan ayrıştırılan parametreler içindir.
func applyQueryParamsToQuery(q *bun.SelectQuery, params url.Values) *bun.SelectQuery {
	if params == nil || len(params) == 0 {
		return q
	}

	for field, values := range params {
		if len(values) == 0 || values[0] == "" {
			continue // Değer yoksa atla
		}

		if field == "date" {
			if len(values) == 1 {
				// Tek tarih değeri
				if isValidDate(values[0]) {
					q = q.Where("DATE(departure_time) = ?", values[0])
				} else {
					fmt.Printf("URL parametresi: Geçersiz tarih formatı: %s\n", values[0])
				}
			} else if len(values) == 2 {
				// Tarih aralığı değeri (parseAbbreviatedQuery'den)
				if isValidDate(values[0]) && isValidDate(values[1]) {
					q = q.Where("DATE(departure_time) BETWEEN ? AND ?", values[0], values[1])
				} else {
					fmt.Printf("URL parametresi: Geçersiz tarih aralığı formatı: %s TO %s\n", values[0], values[1])
				}
			}
		} else {
			// Diğer alanlar için, parseAbbreviatedQuery sadece "=" operatörüyle gelir
			q = q.Where("? = ?", bun.Ident(field), values[0])
		}
	}
	return q
}

// applyFilterCondition, verilen bir FilterRow'a göre bun.SelectQuery'ye
// uygun SQL WHERE koşulunu ekleyen yardımcı bir fonksiyondur.
// Bu fonksiyon hem applySavedFilterToQuery hem de (gerekirse) applyQueryParamsToQuery tarafından kullanılabilir.
func applyFilterCondition(sq *bun.SelectQuery, row models.FilterRow) *bun.SelectQuery {
	switch row.Field {
	case "date":
		// 'date' alanı için özel işlem: Tek tarih veya tarih aralığı
		// Frontend'den "YYYY-MM-DD" veya "YYYY-MM-DD TO YYYY-MM-DD" formatı bekleniyor.
		if strings.Contains(row.Value, " TO ") {
			dates := strings.Split(row.Value, " TO ")
			if len(dates) == 2 && isValidDate(dates[0]) && isValidDate(dates[1]) {
				sq = sq.Where("DATE(departure_time) BETWEEN ? AND ?", dates[0], dates[1])
			} else {
				fmt.Printf("FilterRow: Geçersiz tarih aralığı formatı: '%s'. 'YYYY-MM-DD TO YYYY-MM-DD' bekleniyor.\n", row.Value)
			}
		} else if isValidDate(row.Value) {
			sq = sq.Where("DATE(departure_time) = ?", row.Value)
		} else {
			fmt.Printf("FilterRow: Geçersiz tarih formatı: '%s'. 'YYYY-MM-DD' bekleniyor.\n", row.Value)
		}

	case "ucus_id":
		// `ucus_id` alanı için özel bir muamele gerekiyorsa buraya ekleyin.
		// Örneğin, büyük/küçük harf duyarsız arama yapmak isterseniz:
		// sq = sq.Where("LOWER(ucus_id) "+row.Operator+" LOWER(?)", row.Value)
		// Şimdilik varsayılan operatör mantığına bırakıyoruz.
		fallthrough // Bu case'den sonra alttaki default switch case'e devam et

	default:
		switch row.Operator {
		case "=":
			sq = sq.Where("? = ?", bun.Ident(row.Field), row.Value)
		case "!=":
			sq = sq.Where("? != ?", bun.Ident(row.Field), row.Value)
		case ">":
			sq = sq.Where("? > ?", bun.Ident(row.Field), row.Value)
		case "<":
			sq = sq.Where("? < ?", bun.Ident(row.Field), row.Value)
		case "LIKE":
			sq = sq.Where("? LIKE ?", bun.Ident(row.Field), "%"+row.Value+"%")
		default:
			fmt.Printf("FilterRow: Bilinmeyen veya desteklenmeyen operatör '%s' alanı: %s\n", row.Operator, row.Field)
		}
	}
	return sq
}

// handlers/actual_query.go

// GetActualsByFlightID, belirli bir flight_id'ye (ucus_id) sahip tüm kişilerin tüm Actual verilerini döndürür.
// Ayrıca bulunan person_id'leri ve aktiviteleri person_id'ye göre gruplanmış olarak döndürür.
func GetActualsByFlightID(c *fiber.Ctx) error {
	flightID := c.Params("flight_id") // URL parametresinden flight_id'yi al

	if flightID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "flight_id parametresi eksik.",
		})
	}

	ctx := context.Background()
	var uniquePersonIDs []string // ucus_id'ye sahip benzersiz person_id'leri saklamak için

	// Adım 1: Belirtilen ucus_id'ye sahip tüm benzersiz person_id'leri bul
	err := db.DB.NewSelect().
		Model((*models.Actual)(nil)).
		Column("person_id"). // Sadece person_id sütununu seç
		Where("ucus_id = ?", flightID).
		Group("person_id").         // Benzersiz person_id'leri almak için grupla
		Scan(ctx, &uniquePersonIDs) // uniquePersonIDs olarak ismini değiştirdim

	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error":   "person_id'ler sorgulanırken bir hata oluştu.",
			"details": err.Error(),
		})
	}

	// Eğer hiçbir person_id bulunamazsa, boş sonuç döndür
	if len(uniquePersonIDs) == 0 {
		return c.Status(fiber.StatusOK).JSON(fiber.Map{
			"total_persons_found": 0,
			"person_ids":          []string{},
			"result":              fiber.Map{}, // Boş bir harita döndür
		})
	}

	var allActuals []models.Actual                        // Tüm ilgili actual kayıtları
	actualsByPersonID := make(map[string][]models.Actual) // person_id'ye göre gruplanmış aktiviteler

	// Adım 2: Bulunan tüm person_id'lere ait tüm Actual kayıtlarını getir
	q := db.DB.NewSelect().
		Model((*models.Actual)(nil)).
		Where("person_id IN (?)", bun.In(uniquePersonIDs)). // IN operatörü ile bulunan person_id'lere göre filtrele
		Order("person_id ASC", "departure_time ASC")        // Sıralama ekle

	err = q.Scan(ctx, &allActuals)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error":   "Tüm actual verileri sorgulanırken bir hata oluştu.",
			"details": err.Error(),
		})
	}

	// Adım 3: Aktiviteleri person_id'ye göre grupla
	for _, actual := range allActuals {
		actualsByPersonID[actual.PersonID] = append(actualsByPersonID[actual.PersonID], actual)
	}

	return c.JSON(fiber.Map{
		"total_persons_found": len(uniquePersonIDs), // Kaç farklı person_id bulunduğu
		"person_ids":          uniquePersonIDs,      // Bulunan person_id'lerin listesi
		"result":              actualsByPersonID,    // Person_id'ye göre gruplanmış aktiviteler
	})
}
