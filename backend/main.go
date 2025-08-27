package main

import (
	"log"

	"mini_CMS_Desktop_App/db"
	"mini_CMS_Desktop_App/handlers"
	"mini_CMS_Desktop_App/handlers/activity_code"
	"mini_CMS_Desktop_App/handlers/aircraft_crew_need"
	"mini_CMS_Desktop_App/handlers/crew_document"
	"mini_CMS_Desktop_App/handlers/crew_info"
	"mini_CMS_Desktop_App/handlers/off_day_table"
	"mini_CMS_Desktop_App/handlers/open_trip"
	"mini_CMS_Desktop_App/handlers/penalty"
	"mini_CMS_Desktop_App/handlers/progress"

	"mini_CMS_Desktop_App/handlers/ftl"
	"mini_CMS_Desktop_App/handlers/user_preference"
	"mini_CMS_Desktop_App/middleware"
	"mini_CMS_Desktop_App/repositories"
	"mini_CMS_Desktop_App/services"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/websocket/v2"
)

func main() {
	if err := db.Init(); err != nil {
		log.Fatalf("Veritabanı bağlantı hatası: %v", err)
	}
	sqlDB := db.DB // *bun.DB

	app := fiber.New(fiber.Config{
		BodyLimit: 200 * 1024 * 1024,
	})

	app.Use(cors.New(cors.Config{
		AllowOrigins: "http://localhost:1420",
		AllowHeaders: "Origin, Content-Type, Accept, Authorization",
	}))

	// WebSocket
	app.Use("/ws", func(c *fiber.Ctx) error {
		if websocket.IsWebSocketUpgrade(c) {
			c.Locals("allowed", true)
			return c.Next()
		}
		return fiber.ErrUpgradeRequired
	})
	app.Get("/ws/progress", websocket.New(progress.WebSocketHandler))

	// --- Repositories ---
	briefDebriefRuleRepo := repositories.NewBriefDebriefRuleRepository(sqlDB)
	tripRepo := repositories.NewTripRepository(sqlDB)
	actualRepo := repositories.NewActualRepository(sqlDB)
	userPrefRepo := repositories.NewUserPreferenceRepository(sqlDB)
	publishRepo := repositories.NewPublishRepository(sqlDB)
	openTripRepo := repositories.NewOpenTripRepo(sqlDB) // ✅ Tek repo

	// --- Services ---
	briefDebriefCalc := services.NewBriefDebriefCalculator(briefDebriefRuleRepo)
	ftlCalc := services.NewFTLCalculator(briefDebriefCalc, tripRepo, actualRepo, userPrefRepo)
	openTripService := services.NewOpenTripService(openTripRepo) // ✅ Tek parametre

	// --- Handlers ---
	ftlHandler := ftl.NewFTLHandler(ftlCalc, tripRepo)
	actualImportXLSXHandler := handlers.NewActualImportXLSXHandler(actualRepo, ftlCalc, tripRepo, ftlHandler)
	publishImportXLSXHandler := handlers.NewPublishImportXLSXHandler(publishRepo)
	publishQueryHandler := handlers.NewPublishQueryHandler(publishRepo)
	userPrefHandler := user_preference.NewUserPreferenceHandler(userPrefRepo)
	openTripHandler := open_trip.NewOpenTripHandler(openTripService)

	// --- Public Routes ---
	app.Post("/api/register", handlers.RegisterUserHandler)
	app.Post("/api/login", handlers.LoginHandler)

	// --- Protected Routes ---
	protected := app.Group("/api", middleware.JWTMiddleware())

	// ACTUAL
	protected.Get("/actual", handlers.QueryActualData)
	protected.Post("/actual/import-xlsx", actualImportXLSXHandler.ImportActualXLSX)
	protected.Get("/actual/list", handlers.ListActualData)
	protected.All("/actual/query", handlers.QueryActualData)
	protected.Get("/actual/preview", handlers.PreviewActualData)
	protected.Get("/actual/by-flight-id/:flight_id", handlers.GetActualsByFlightID)

	// PUBLISH
	protected.Post("/publish/import-xlsx", publishImportXLSXHandler.ImportPublishXLSX)
	protected.Get("/publish/query", publishQueryHandler.GetPublishesByPersonID)

	// ACTIVITY CODES
	protected.Post("/activity-codes/import-data", activity_code.ImportActivityCodeData)
	protected.Get("/activity-codes/list", activity_code.ListActivityCodes)

	// CREW DOCUMENTS
	protected.Post("/crew-documents/import-data", crew_document.ImportCrewDocumentData)
	protected.Get("/crew-documents", crew_document.ListCrewDocuments)
	protected.Get("/crew-documents/query", crew_document.QueryCrewDocuments)

	// OFF DAY TABLE
	protected.Post("/off-day-table/import-data", off_day_table.ImportOffDayTableData)
	protected.Get("/off-day-table/list", off_day_table.ListOffDayTable)

	// CREW INFO
	protected.Post("/crew-info/import-data", crew_info.ImportCrewInfoData)
	protected.Get("/crew-info/list", crew_info.ListCrewInfo)

	// PENALTIES
	protected.Post("/penalties/import-data", penalty.ImportPenaltyData)
	protected.Get("/penalties/list", penalty.ListPenalties)

	// AIRCRAFT CREW NEED
	protected.Post("/aircraft-crew-need/import-data", aircraft_crew_need.ImportAircraftCrewNeedData)
	protected.Get("/aircraft-crew-need/list", aircraft_crew_need.ListAircraftCrewNeed)

	// FTL
	protected.Post("/ftl/calculate_trip", ftlHandler.HandleCalculateTripFTL)
	protected.Post("/ftl/recalculate_crew_schedule", ftlHandler.HandleRecalculateCrewScheduleFTL)
	protected.Get("/ftl/trips_by_crew_id", ftlHandler.GetTripsByCrewID)

	// USER PREFERENCES
	protected.Post("/user_preferences", userPrefHandler.SetUserPreference)
	protected.Get("/user_preferences", userPrefHandler.GetUserPreference)

	// ✅ OPENTRIP
	protected.Get("/trips/open", openTripHandler.GetOpenTrips)

	// Start server
	log.Println("🚀 Sunucu başlatıldı: http://localhost:8080")
	log.Fatal(app.Listen(":8080"))
}
