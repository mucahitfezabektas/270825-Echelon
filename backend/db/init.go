// db/init.go
package db

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"os"
	"time"

	"mini_CMS_Desktop_App/models" // models paketini içe aktardığınızdan emin olun

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/joho/godotenv" // .env dosyasını okumak için
	_ "github.com/lib/pq"      // PostgreSQL sürücüsü
	"github.com/uptrace/bun"
	"github.com/uptrace/bun/dialect/pgdialect"
	"github.com/uptrace/bun/extra/bundebug"
)

var DB *bun.DB
var RawPGConn *pgconn.PgConn

func Init() error {
	// .env dosyasını yükle (uygulama başında bir kez okunur).
	// Eğer .env dosyası yoksa veya yüklenemezse hata döndürmez, sadece loglar.
	if err := godotenv.Load(); err != nil {
		log.Printf("UYARI: .env dosyası yüklenemedi: %v (Ortam değişkenleri zaten ayarlıysa bu normaldir)", err)
	}

	// Ortam değişkeninden DSN'yi al
	dsn := os.Getenv("POSTGRES_DSN")
	if dsn == "" {
		// DSN tanımlı değilse hata döndür
		return fmt.Errorf("POSTGRES_DSN ortam değişkeni tanımlı değil. Lütfen .env dosyanızı veya ortam ayarlarınızı kontrol edin.")
	}

	sqlDB, err := sql.Open("postgres", dsn)
	if err != nil {
		return fmt.Errorf("SQL veritabanı bağlantısı açılamadı: %w", err)
	}

	// 🔧 Performans: Connection Pool ayarları
	sqlDB.SetMaxOpenConns(50)
	sqlDB.SetMaxIdleConns(10)
	sqlDB.SetConnMaxLifetime(time.Hour)

	DB = bun.NewDB(sqlDB, pgdialect.New())

	// Hata ayıklama (debug) için sorgu kancası ekle
	// Sadece development ortamında detaylı ve renkli loglama yap
	if os.Getenv("APP_ENV") == "development" {
		DB.AddQueryHook(bundebug.NewQueryHook(
			bundebug.WithVerbose(true), // Detaylı sorgu loglaması için
		))
	} else {
		// Üretim ortamında daha az detaylı loglama veya farklı bir kanca eklenebilir
		DB.AddQueryHook(bundebug.NewQueryHook()) // Sadece temel hatalar için
	}

	// Veritabanı bağlantısını kontrol et
	if err := DB.PingContext(context.Background()); err != nil {
		return fmt.Errorf("veritabanına bağlanılamadı (Ping hatası): %w", err)
	}
	log.Println("✅ Veritabanına başarıyla bağlanıldı (bun).")

	// 🔌 pgx bağlantısı (COPY FROM gibi özel PostgreSQL işlemleri için)
	// Bu bağlantı, bun'dan farklı olarak düşük seviyeli pgx sürücüsünü doğrudan kullanır.
	rawConn, err := pgx.Connect(context.Background(), dsn)
	if err != nil {
		return fmt.Errorf("pgx bağlantısı kurulamadı: %w", err)
	}
	RawPGConn = rawConn.PgConn() // Ham pgconn bağlantısını sakla
	log.Println("✅ Veritabanına başarıyla bağlanıldı (pgx).")

	// 🧱 Uygulamadaki tüm modeller için veritabanı tablolarını oluştur/kontrol et
	// `models.User` modelini bu listeye ekledik.
	modelsToCreate := []interface{}{
		(*models.Actual)(nil),
		(*models.Publish)(nil),
		(*models.ActivityCode)(nil),
		(*models.CrewDocument)(nil),
		(*models.OffDayTable)(nil),
		(*models.CrewInfo)(nil),
		(*models.Penalty)(nil),
		(*models.AircraftCrewNeed)(nil),
		(*models.Trip)(nil),
		(*models.BriefDebriefRule)(nil),
		(*models.UserPreference)(nil),
		// ✅ Yeni eklenen: Kullanıcılar tablosu için model
		(*models.User)(nil),
	}

	for _, model := range modelsToCreate {
		_, err := DB.NewCreateTable().Model(model).IfNotExists().Exec(context.Background())
		if err != nil {
			// Tablo oluşturma hatasında detaylı bilgi ver
			return fmt.Errorf("'%T' tablosu oluşturulamadı: %w", model, err)
		}
		log.Printf("✔️ Tablo oluşturuldu/kontrol edildi: %T", model)
	}

	// 📦 Brief/Debrief kurallarını başlat (varsa veya boşsa ekle)
	if err := initializeBriefDebriefRules(context.Background(), DB); err != nil {
		log.Printf("❌ Brief/Debrief kuralları başlatılamadı: %v", err)
	}

	return nil
}

// initializeBriefDebriefRules fonksiyonu aynı kalır.
// Bu fonksiyon, BriefDebriefRule modelinin tablo oluşturma mantığına doğrudan etkisi yoktur,
// sadece başlangıç verisi ekler.
func initializeBriefDebriefRules(ctx context.Context, db *bun.DB) error {
	count, err := db.NewSelect().Model((*models.BriefDebriefRule)(nil)).Count(ctx)
	if err != nil {
		return fmt.Errorf("brief_debrief_rules sayılırken hata: %w", err)
	}
	if count > 0 {
		log.Println("Bilgi: brief_debrief_rules tablosunda zaten veri var, başlatma atlandı.")
		return nil
	}

	log.Println("Bilgi: brief_debrief_rules tablosu boş, başlangıç verileri ekleniyor...")

	rules := []models.BriefDebriefRule{
		// --- Uçuş Ekibi (Kargo Hariç) ---
		// Yolculu Uçuşlar (Manuel Tablo-2)
		{ScenarioType: "Yolculu Uçuşlar", AircraftType: "DAR GÖVDE", CrewType: "Uçuş Ekibi", DutyStartAirport: "IST", BriefDurationMin: 75, DebriefDurationMin: 30, Priority: 100},
		{ScenarioType: "Yolculu Uçuşlar", AircraftType: "GENİŞ GÖVDE", CrewType: "Uçuş Ekibi", DutyStartAirport: "IST", BriefDurationMin: 90, DebriefDurationMin: 30, Priority: 100},
		{ScenarioType: "Yolculu Uçuşlar", AircraftType: "DAR GÖVDE", CrewType: "Uçuş Ekibi", DutyStartAirport: "ISL", BriefDurationMin: 60, DebriefDurationMin: 30, Priority: 90},
		{ScenarioType: "Yolculu Uçuşlar", AircraftType: "GENİŞ GÖVDE", CrewType: "Uçuş Ekibi", DutyStartAirport: "ISL", BriefDurationMin: 90, DebriefDurationMin: 30, Priority: 90},
		{ScenarioType: "Yolculu Uçuşlar", AircraftType: "DAR GÖVDE", CrewType: "Uçuş Ekibi", DutyStartAirport: "SAW", BriefDurationMin: 60, DebriefDurationMin: 30, Priority: 90},
		{ScenarioType: "Yolculu Uçuşlar", AircraftType: "GENİŞ GÖVDE", CrewType: "Uçuş Ekibi", DutyStartAirport: "SAW", BriefDurationMin: 90, DebriefDurationMin: 30, Priority: 90},
		{ScenarioType: "Yolculu Uçuşlar", AircraftType: "DAR GÖVDE", CrewType: "Uçuş Ekibi", DutyStartAirport: "Diğer", BriefDurationMin: 60, DebriefDurationMin: 30, Priority: 80},
		{ScenarioType: "Yolculu Uçuşlar", AircraftType: "GENİŞ GÖVDE", CrewType: "Uçuş Ekibi", DutyStartAirport: "Diğer", BriefDurationMin: 60, DebriefDurationMin: 30, Priority: 80},

		// Simülatör (models.GetDutyTypeFromActual'dan "Simülatör" geldiği varsayımıyla)
		// AircraftType "BİLİNMİYOR" olduğunda eşleşmesi için (models.GetAircraftTypeFromCmsType'tan)
		{ScenarioType: "Simülatör", AircraftType: "BİLİNMİYOR", CrewType: "Uçuş Ekibi", DutyStartAirport: "Hepsi", BriefDurationMin: 60, DebriefDurationMin: 60, Priority: 75}, // Yüksek öncelik
		{ScenarioType: "Simülatör", AircraftType: "Hepsi", CrewType: "Uçuş Ekibi", DutyStartAirport: "Hepsi", BriefDurationMin: 60, DebriefDurationMin: 60, Priority: 70},      // Genel kural

		// Konumlandırma (models.GetDutyTypeFromActual'dan "Konumlandırma" geldiği varsayımıyla)
		// AircraftType "BİLİNMİYOR" olduğunda eşleşmesi için (örn. otobüs veya diğer havayolları konumlandırması)
		{ScenarioType: "Konumlandırma", AircraftType: "BİLİNMİYOR", CrewType: "Uçuş Ekibi", DutyStartAirport: "Hepsi", BriefDurationMin: 60, DebriefDurationMin: 0, Priority: 75}, // Yüksek öncelik
		// Konumlandırma (FlightPosition "DH" veya GroupCode="GT" & ActivityCode="BUS"/"OAF" için)
		{ScenarioType: "Konumlandırma", AircraftType: "Hepsi", CrewType: "Uçuş Ekibi", DutyStartAirport: "Hepsi", BriefDurationMin: 60, DebriefDurationMin: 0, Priority: 70}, // Genel kural

		// İntikal Uçuşları-Yolcusuz (models.GetDutyTypeFromActual'dan "Yolculu Uçuşlar" geldiği varsayımıyla, şimdilik aynı)
		{ScenarioType: "Yolculu Uçuşlar", AircraftType: "Hepsi", CrewType: "Uçuş Ekibi", DutyStartAirport: "Hepsi", BriefDurationMin: 60, DebriefDurationMin: 15, Priority: 70}, // İntikal Uçuşları için varsayılan

		// Açık Mesai (models.GetDutyTypeFromActual'dan "Açık Mesai" geldiği varsayımıyla)
		{ScenarioType: "Açık Mesai", AircraftType: "Hepsi", CrewType: "Uçuş Ekibi", DutyStartAirport: "Hepsi", BriefDurationMin: 60, DebriefDurationMin: 15, Priority: 70},

		// --- Kabin Ekibi ---
		// Yolculu Uçuşlar (Kabin Ekibi)
		{ScenarioType: "Yolculu Uçuşlar", AircraftType: "DAR GÖVDE", CrewType: "Kabin Ekibi", DutyStartAirport: "IST", BriefDurationMin: 75, DebriefDurationMin: 30, Priority: 100},
		{ScenarioType: "Yolculu Uçuşlar", AircraftType: "GENİŞ GÖVDE", CrewType: "Kabin Ekibi", DutyStartAirport: "IST", BriefDurationMin: 90, DebriefDurationMin: 30, Priority: 100},
		{ScenarioType: "Yolculu Uçuşlar", AircraftType: "DAR GÖVDE", CrewType: "Kabin Ekibi", DutyStartAirport: "ISL", BriefDurationMin: 60, DebriefDurationMin: 30, Priority: 90},
		{ScenarioType: "Yolculu Uçuşlar", AircraftType: "GENİŞ GÖVDE", CrewType: "Kabin Ekibi", DutyStartAirport: "ISL", BriefDurationMin: 90, DebriefDurationMin: 30, Priority: 90},
		{ScenarioType: "Yolculu Uçuşlar", AircraftType: "DAR GÖVDE", CrewType: "Kabin Ekibi", DutyStartAirport: "SAW", BriefDurationMin: 60, DebriefDurationMin: 30, Priority: 90},
		{ScenarioType: "Yolculu Uçuşlar", AircraftType: "GENİŞ GÖVDE", CrewType: "Kabin Ekibi", DutyStartAirport: "SAW", BriefDurationMin: 90, DebriefDurationMin: 30, Priority: 90},
		{ScenarioType: "Yolculu Uçuşlar", AircraftType: "DAR GÖVDE", CrewType: "Kabin Ekibi", DutyStartAirport: "Diğer", BriefDurationMin: 60, DebriefDurationMin: 30, Priority: 80},
		{ScenarioType: "Yolculu Uçuşlar", AircraftType: "GENİŞ GÖVDE", CrewType: "Kabin Ekibi", DutyStartAirport: "Diğer", BriefDurationMin: 60, DebriefDurationMin: 30, Priority: 80},

		// Simülatör (Kabin Ekibi)
		{ScenarioType: "Simülatör", AircraftType: "BİLİNMİYOR", CrewType: "Kabin Ekibi", DutyStartAirport: "Hepsi", BriefDurationMin: 60, DebriefDurationMin: 60, Priority: 75},
		{ScenarioType: "Simülatör", AircraftType: "Hepsi", CrewType: "Kabin Ekibi", DutyStartAirport: "Hepsi", BriefDurationMin: 60, DebriefDurationMin: 60, Priority: 70},

		// Konumlandırma (Kabin Ekibi)
		{ScenarioType: "Konumlandırma", AircraftType: "BİLİNMİYOR", CrewType: "Kabin Ekibi", DutyStartAirport: "Hepsi", BriefDurationMin: 60, DebriefDurationMin: 0, Priority: 75},
		{ScenarioType: "Konumlandırma", AircraftType: "Hepsi", CrewType: "Kabin Ekibi", DutyStartAirport: "Hepsi", BriefDurationMin: 60, DebriefDurationMin: 0, Priority: 70},

		// İntikal Uçuşları-Yolcusuz (Kabin Ekibi)
		{ScenarioType: "Yolculu Uçuşlar", AircraftType: "Hepsi", CrewType: "Kabin Ekibi", DutyStartAirport: "Hepsi", BriefDurationMin: 60, DebriefDurationMin: 15, Priority: 70}, // İntikal Uçuşları için varsayılan
		{ScenarioType: "Açık Mesai", AircraftType: "Hepsi", CrewType: "Kabin Ekibi", DutyStartAirport: "Hepsi", BriefDurationMin: 60, DebriefDurationMin: 15, Priority: 70},

		// --- Kargo Uçuş Ekibi ---
		// Konumlandırma (Kargo Uçuş Ekibi)
		{ScenarioType: "Konumlandırma", AircraftType: "BİLİNMİYOR", CrewType: "Kargo Uçuş Ekibi", DutyStartAirport: "Hepsi", BriefDurationMin: 60, DebriefDurationMin: 0, Priority: 95},
		{ScenarioType: "Konumlandırma", AircraftType: "Hepsi", CrewType: "Kargo Uçuş Ekibi", DutyStartAirport: "Hepsi", BriefDurationMin: 60, DebriefDurationMin: 0, Priority: 90},

		// İlk sektörü görevli (Kargo Uçuş Ekibi)
		{ScenarioType: "İlk sektörü görevli", AircraftType: "Hepsi", CrewType: "Kargo Uçuş Ekibi", DutyStartAirport: "Hepsi", BriefDurationMin: 60, DebriefDurationMin: 30, Priority: 100},
		{ScenarioType: "Açık Mesai", AircraftType: "Hepsi", CrewType: "Kargo Uçuş Ekibi", DutyStartAirport: "Hepsi", BriefDurationMin: 60, DebriefDurationMin: 15, Priority: 80},
	}

	_, err = db.NewInsert().Model(&rules).Exec(ctx)
	if err != nil {
		return fmt.Errorf("brief/debrief başlangıç verileri eklenirken hata: %w", err)
	}

	log.Println("Bilgi: brief_debrief_rules tablosuna başlangıç verileri başarıyla eklendi.")
	return nil
}
