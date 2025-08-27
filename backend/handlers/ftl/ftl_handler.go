// C:\Users\mucah_wi2yyc2\Desktop\mini_CMS_Desktop_App\backend\handlers\ftl\ftl_handler.go

package ftl

import (
	"database/sql"
	"log"
	"sort"
	"time"

	"mini_CMS_Desktop_App/models"
	"mini_CMS_Desktop_App/repositories"
	"mini_CMS_Desktop_App/services"

	"github.com/gofiber/fiber/v2"
)

type FTLHandler struct {
	ftlCalc  *services.FTLCalculator
	tripRepo *repositories.TripRepository
}

func NewFTLHandler(ftlCalc *services.FTLCalculator, tripRepo *repositories.TripRepository) *FTLHandler {
	return &FTLHandler{
		ftlCalc:  ftlCalc,
		tripRepo: tripRepo,
	}
}

// CalculateTripFTLRequest: Frontend'den gelen isteğin yapısı
type CalculateTripFTLRequest struct {
	TripID       string          `json:"trip_id"`
	CrewMemberID string          `json:"crew_member_id"`
	Activities   []models.Actual `json:"activities"`
	// AircraftType alanı kaldırıldı, artık BriefAircraftType ve DebriefAircraftType kullanılacak.
	DutyStartAirport string `json:"duty_start_airport"`
	DutyType         string `json:"duty_type"` // Genel DutyType (hala tutulabilir)
	CrewType         string `json:"crew_type"`
}

// HandleCalculateTripFTL: Tek bir trip için FTL hesaplamalarını yapar ve sonucu döndürür.
func (h *FTLHandler) HandleCalculateTripFTL(c *fiber.Ctx) error {
	var req CalculateTripFTLRequest
	if err := c.BodyParser(&req); err != nil {
		log.Printf("Hata: CalculateTripFTL isteği ayrıştırılamadı: %v", err)
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Geçersiz istek gövdesi", "details": err.Error()})
	}

	if req.TripID == "" || req.CrewMemberID == "" || len(req.Activities) == 0 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "TripID, CrewMemberID ve Activities boş olamaz"})
	}

	sort.Slice(req.Activities, func(i, j int) bool {
		return req.Activities[i].DutyStart.Before(req.Activities[j].DutyStart)
	})

	var firstFlightTime, lastFlightTime time.Time
	var firstFlightFound, lastFlightFound bool
	var firstFLTActivity *models.Actual // BriefTripType için ilk FLT aktivitesi
	var lastFLTActivity *models.Actual  // DebriefTripType için son FLT aktivitesi

	// Aktivitelerdeki AircraftType'ı doldur (bun "-" tag'i nedeniyle DB'den gelmez)
	processedActivities := make([]models.Actual, len(req.Activities))
	for i, act := range req.Activities {
		processedActivities[i] = act
		processedActivities[i].AircraftType = models.GetAircraftTypeFromCmsType(act.PlaneCmsType) // Her aktivitenin kendi AircraftType'ı dolduruluyor

		if act.GroupCode == "FLT" {
			if !firstFlightFound {
				firstFlightTime = act.DepartureTime
				firstFlightFound = true
				firstFLTActivity = &processedActivities[i] // processedActivities'ten referans al
			}
			lastFlightTime = act.ArrivalTime
			lastFlightFound = true
			lastFLTActivity = &processedActivities[i] // processedActivities'ten referans al
		}
	}
	req.Activities = processedActivities // İşlenmiş aktiviteleri geri ata

	if !firstFlightFound || !lastFlightFound {
		log.Printf("Uyarı: Trip %s için uçuş aktivitesi bulunamadı, FirstLegDepartureTime/LastLegArrivalTime belirlenemiyor. DutyStart/DutyEnd kullanılıyor.", req.TripID)
		if len(req.Activities) > 0 {
			firstFlightTime = req.Activities[0].DutyStart
			lastFlightTime = req.Activities[len(req.Activities)-1].DutyEnd
		} else {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Trip için aktivite bulunamadı"})
		}
	}

	// briefTripType ve debriefTripType'ı belirle
	var briefTripType string
	var debriefTripType string

	if firstFLTActivity != nil {
		briefTripType = models.GetDutyTypeFromActual(firstFLTActivity)
	} else if len(req.Activities) > 0 {
		briefTripType = models.GetDutyTypeFromActual(&req.Activities[0])
	} else {
		briefTripType = "BİLİNMİYOR"
	}

	if lastFLTActivity != nil {
		debriefTripType = models.GetDutyTypeFromActual(lastFLTActivity)
	} else if len(req.Activities) > 0 {
		debriefTripType = models.GetDutyTypeFromActual(&req.Activities[len(req.Activities)-1])
	} else {
		debriefTripType = "BİLİNMİYOR"
	}

	// briefAircraftType ve debriefAircraftType'ı belirle
	var briefAircraftType string
	var debriefAircraftType string

	if firstFLTActivity != nil {
		briefAircraftType = models.GetAircraftTypeFromCmsType(firstFLTActivity.PlaneCmsType)
	} else if len(req.Activities) > 0 {
		briefAircraftType = models.GetAircraftTypeFromCmsType(req.Activities[0].PlaneCmsType)
	} else {
		briefAircraftType = "BİLİNMİYOR"
	}

	if lastFLTActivity != nil {
		debriefAircraftType = models.GetAircraftTypeFromCmsType(lastFLTActivity.PlaneCmsType)
	} else if len(req.Activities) > 0 {
		debriefAircraftType = models.GetAircraftTypeFromCmsType(req.Activities[len(req.Activities)-1].PlaneCmsType)
	} else {
		debriefAircraftType = "BİLİNMİYOR"
	}

	// Diğer genel trip özellikleri (req'ten veya ilk aktiviteden türetilir)
	var dutyStartAirport string
	if req.DutyStartAirport != "" {
		dutyStartAirport = req.DutyStartAirport
	} else if len(req.Activities) > 0 {
		dutyStartAirport = req.Activities[0].DeparturePort
	} else {
		dutyStartAirport = "BİLİNMİYOR"
	}

	var dutyType string
	if req.DutyType != "" {
		dutyType = req.DutyType
	} else if len(req.Activities) > 0 {
		dutyType = models.GetDutyTypeFromActual(&req.Activities[0])
	} else {
		dutyType = "BİLİNMİYOR"
	}

	var crewType string
	if req.CrewType != "" {
		crewType = req.CrewType
	} else if len(req.Activities) > 0 {
		crewType = models.GetCrewTypeFromFlightPosition(req.Activities[0].FlightPosition)
	} else {
		crewType = "BİLİNMİYOR"
	}

	// generalAircraftType artık kullanılmıyor, kaldırıldı.

	trip, err := h.tripRepo.GetTripByID(req.TripID)
	if err != nil && err != sql.ErrNoRows {
		log.Printf("Hata: Trip %s veritabanından çekilirken sorun: %v", req.TripID, err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "İç sunucu hatası (trip çekme)"})
	}

	if trip == nil {
		trip = &models.Trip{
			TripID:                req.TripID,
			CrewMemberID:          req.CrewMemberID,
			Activities:            req.Activities, // ProcessedActivities'ten geliyor
			FirstLegDepartureTime: firstFlightTime,
			LastLegArrivalTime:    lastFlightTime,
			// AircraftType:           generalAircraftType, // <<< BU SATIR KALDIRILDI
			DutyStartAirport:    dutyStartAirport,
			DutyType:            dutyType,
			CrewType:            crewType,
			BriefTripType:       briefTripType,
			DebriefTripType:     debriefTripType,
			BriefAircraftType:   briefAircraftType,
			DebriefAircraftType: debriefAircraftType,
		}
	} else {
		trip.CrewMemberID = req.CrewMemberID
		trip.Activities = req.Activities
		trip.FirstLegDepartureTime = firstFlightTime
		trip.LastLegArrivalTime = lastFlightTime
		// trip.AircraftType = generalAircraftType // <<< BU SATIR KALDIRILDI
		trip.DutyStartAirport = dutyStartAirport
		trip.DutyType = dutyType
		trip.CrewType = crewType
		trip.BriefTripType = briefTripType
		trip.DebriefTripType = debriefTripType
		trip.BriefAircraftType = briefAircraftType
		trip.DebriefAircraftType = debriefAircraftType
	}

	allCrewTripsRaw, err := h.tripRepo.GetTripsByCrewMemberID(trip.CrewMemberID, time.Now().AddDate(-1, 0, -28))
	if err != nil {
		log.Printf("Hata: Ekip %s için tüm tripler çekilirken sorun: %v", trip.CrewMemberID, err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "İç sunucu hatası (ekip tripleri çekme)"})
	}

	allCrewTrips := make([]*models.Trip, len(allCrewTripsRaw))
	for i := range allCrewTripsRaw {
		allCrewTrips[i] = &allCrewTripsRaw[i]
	}

	foundCurrentTripInAll := false
	for _, t := range allCrewTrips {
		if t.TripID == trip.TripID {
			foundCurrentTripInAll = true
			break
		}
	}
	if !foundCurrentTripInAll {
		allCrewTrips = append(allCrewTrips, trip)
		sort.Slice(allCrewTrips, func(i, j int) bool {
			return allCrewTrips[i].FirstLegDepartureTime.Before(allCrewTrips[j].FirstLegDepartureTime)
		})
	}

	if err := h.ftlCalc.CalculateFTLForTrip(trip, allCrewTrips); err != nil {
		log.Printf("Hata: Trip %s için FTL hesaplanırken sorun: %v", trip.TripID, err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "FTL hesaplanırken hata"})
	}

	if err := h.tripRepo.SaveTrip(trip); err != nil {
		log.Printf("Hata: Trip %s FTL hesaplaması sonrası kaydedilirken sorun: %v", trip.TripID, err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Hesaplanan FTL kaydedilirken hata"})
	}

	return c.Status(fiber.StatusOK).JSON(trip)
}

type RecalculateCrewScheduleRequest struct {
	CrewMemberID string `json:"crew_member_id"`
}

// HandleRecalculateCrewScheduleFTL: Fiber'a uyumlu hale getirildi
func (h *FTLHandler) HandleRecalculateCrewScheduleFTL(c *fiber.Ctx) error { // <<< BURADA DÜZELTİLDİ: signature değişti
	var req RecalculateCrewScheduleRequest
	if err := c.BodyParser(&req); err != nil {
		log.Printf("Hata: RecalculateCrewSchedule isteği ayrıştırılamadı: %v", err)
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Geçersiz istek gövdesi", "details": err.Error()})
	}

	if req.CrewMemberID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "CrewMemberID boş olamaz"})
	}

	allCrewTripsRaw, err := h.tripRepo.GetTripsByCrewMemberID(req.CrewMemberID, time.Now().AddDate(-1, 0, -28))
	if err != nil {
		log.Printf("Hata: Ekip %s için tüm tripler çekilirken sorun: %v", req.CrewMemberID, err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "İç sunucu hatası çekilirken ekip tripleri"})
	}

	allCrewTrips := make([]*models.Trip, len(allCrewTripsRaw))
	for i := range allCrewTripsRaw {
		allCrewTrips[i] = &allCrewTripsRaw[i]
	}

	sort.Slice(allCrewTrips, func(i, j int) bool {
		return allCrewTrips[i].FirstLegDepartureTime.Before(allCrewTrips[j].FirstLegDepartureTime)
	})

	for i := range allCrewTrips {
		trip := allCrewTrips[i]
		if err := h.ftlCalc.CalculateFTLForTrip(trip, allCrewTrips); err != nil {
			log.Printf("Hata: Toplu yeniden hesaplama sırasında trip %s için FTL hesaplanırken sorun: %v", trip.TripID, err)
		}
		if err := h.tripRepo.SaveTrip(trip); err != nil {
			log.Printf("Hata: Toplu yeniden hesaplama sonrası trip %s kaydedilirken sorun: %v", trip.TripID, err)
		}
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{"message": "Crew schedule FTL recalculation initiated successfully."})
}

func (h *FTLHandler) GetTripsByCrewID(c *fiber.Ctx) error {
	crewID := c.Query("crew_id")
	if crewID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "crew_id parametresi gerekli"})
	}

	fromTime := time.Now().AddDate(-1, 0, -28)
	trips, err := h.tripRepo.GetTripsByCrewMemberID(crewID, fromTime)
	if err != nil {
		log.Printf("Hata: Ekip %s için tripler çekilirken sorun: %v", crewID, err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "İç sunucu hatası"})
	}

	return c.Status(fiber.StatusOK).JSON(trips)
}
