package services

import (
	"context"
	"fmt"
	"strings"

	"mini_CMS_Desktop_App/models"
	"mini_CMS_Desktop_App/repositories"
)

type OpenTripService struct {
	Repo *repositories.OpenTripRepo
}

func NewOpenTripService(repo *repositories.OpenTripRepo) *OpenTripService {
	return &OpenTripService{Repo: repo}
}

func (s *OpenTripService) GetOpenTrips(ctx context.Context, period string) ([]models.OpenTripNeed, error) {
	// 1️⃣ FLT aktivitelerini al
	actuals, err := s.Repo.GetFLTActivities(ctx, period)
	if err != nil {
		return nil, err
	}

	results := []models.OpenTripNeed{}

	// 2️⃣ Flight bazlı grupla
	flightMap := make(map[string][]models.Actual)
	for _, a := range actuals {
		flightMap[a.UçuşID] = append(flightMap[a.UçuşID], a)
	}

	// 3️⃣ Her flight için karşılaştırma yap
	for _, group := range flightMap {
		if len(group) == 0 {
			continue
		}

		aircraftType := strings.TrimSpace(group[0].AircraftType)
		tripID := group[0].TripID

		if aircraftType == "" {
			fmt.Printf("[WARN] FlightKey=%s TripID=%s için aircraftType boş geldi\n", group[0].UçuşID, tripID)
			continue // veya istersen loglayıp atlayabiliriz
		}

		// 4️⃣ İhtiyaçları getir (map olarak)
		required, err := s.Repo.GetNeedsByAircraftType(ctx, aircraftType)
		if err != nil {
			continue
		}

		// 5️⃣ Atanmış ekipleri say
		assigned := make(map[string]int)
		for _, a := range group {
			assigned[a.FlightPosition]++
		}

		// 6️⃣ Farkı hesapla
		diff := make(map[string]int)
		status := "exact"
		for pos, req := range required {
			got := assigned[pos]
			diff[pos] = got - req
			if got < req {
				status = "under"
			} else if got > req && status != "under" {
				status = "over"
			}
		}

		// 7️⃣ Sadece eksik olanları döndür
		if status == "under" {
			results = append(results, models.OpenTripNeed{
				TripID:       tripID,
				FlightKey:    group[0].UçuşID,
				AircraftType: aircraftType,
				Status:       status,
				Assigned:     assigned,
				Required:     required,
				Diff:         diff,
			})
		}
	}

	return results, nil
}
