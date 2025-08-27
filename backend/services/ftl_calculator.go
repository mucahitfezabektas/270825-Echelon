package services

import (
	"context" // Import the context package
	"fmt"
	"log"
	"sort"
	"time"

	"mini_CMS_Desktop_App/models"
	"mini_CMS_Desktop_App/repositories"
)

// FTLCalculator, SHT-FTL kurallarını uygular ve uçuş/görev/dinlenme sürelerini hesaplar.
type FTLCalculator struct {
	briefDebriefCalc *BriefDebriefCalculator
	tripRepo         *repositories.TripRepository
	actualRepo       *repositories.ActualRepository
	userPrefRepo     *repositories.UserPreferenceRepository
}

// NewFTLCalculator, FTLCalculator'ın yeni bir örneğini oluşturur.
func NewFTLCalculator(
	briefDebriefCalc *BriefDebriefCalculator,
	tripRepo *repositories.TripRepository,
	actualRepo *repositories.ActualRepository,
	userPrefRepo *repositories.UserPreferenceRepository,
) *FTLCalculator {
	return &FTLCalculator{
		briefDebriefCalc: briefDebriefCalc,
		tripRepo:         tripRepo,
		actualRepo:       actualRepo,
		userPrefRepo:     userPrefRepo,
	}
}

// CalculateFTLForTrip (mevcut hali, önceki yanıtta verilmişti)
func (f *FTLCalculator) CalculateFTLForTrip(trip *models.Trip, allCrewTrips []*models.Trip) error {
	trip.FTLViolations = []string{}

	var preferredLocation *time.Location = time.UTC
	userIDForPref := trip.CrewMemberID

	userPref, err := f.userPrefRepo.GetPreferenceByUserID(userIDForPref)
	if err != nil {
		log.Printf("Uyarı: Kullanıcı %s için zaman dilimi tercihi çekilemedi: %v. Varsayılan UTC kullanılıyor.", userIDForPref, err)
	} else if userPref != nil && userPref.TimeZone != "" {
		loc, err := time.LoadLocation(userPref.TimeZone)
		if err != nil {
			log.Printf("Uyarı: Geçersiz zaman dilimi tercihi '%s' için kullanıcı %s: %v. Varsayılan UTC kullanılıyor.", userPref.TimeZone, userIDForPref, err)
		} else {
			preferredLocation = loc
		}
	}

	// Fix: Add context.Background() as the first argument
	briefOnlyMin, _ := f.briefDebriefCalc.GetBriefDebriefDurations(
		context.Background(),
		trip.CrewType,
		trip.BriefTripType,
		trip.BriefAircraftType, // <<<< BURADA KULLANILIYOR
		trip.DutyStartAirport,
	)
	trip.CalculatedBriefDurationMin = briefOnlyMin

	// Fix: Add context.Background() as the first argument
	_, debriefOnlyMin := f.briefDebriefCalc.GetBriefDebriefDurations(
		context.Background(),
		trip.CrewType,
		trip.DebriefTripType,
		trip.DebriefAircraftType, // <<<< BURADA KULLANILIYOR
		trip.DutyStartAirport,
	)
	trip.CalculatedDebriefDurationMin = debriefOnlyMin

	firstLegDepInPrefLoc := trip.FirstLegDepartureTime.In(preferredLocation)
	lastLegArrInPrefLoc := trip.LastLegArrivalTime.In(preferredLocation)

	// trip.CalculatedDutyPeriodStart = firstLegDepInPrefLoc.Add(time.Duration(-briefOnlyMin) * time.Minute)
	// trip.CalculatedDutyPeriodEnd = lastLegArrInPrefLoc.Add(time.Duration(debriefOnlyMin) * time.Minute)

	if len(trip.Activities) > 0 {
		trip.CalculatedDutyPeriodStart = trip.Activities[0].DutyStart.In(preferredLocation)
		trip.CalculatedDutyPeriodEnd = trip.Activities[len(trip.Activities)-1].DutyEnd.In(preferredLocation)
	} else {
		trip.CalculatedDutyPeriodStart = firstLegDepInPrefLoc.Add(time.Duration(-briefOnlyMin) * time.Minute)
		trip.CalculatedDutyPeriodEnd = lastLegArrInPrefLoc.Add(time.Duration(debriefOnlyMin) * time.Minute)
	}

	trip.CalculatedDutyPeriodDurationMin = int(trip.CalculatedDutyPeriodEnd.Sub(trip.CalculatedDutyPeriodStart).Minutes())

	trip.CalculatedFlightDutyPeriodDurationMin = int(lastLegArrInPrefLoc.Sub(trip.CalculatedDutyPeriodStart).Minutes())

	var prevTrip *models.Trip
	for i, t := range allCrewTrips {
		if t.TripID == trip.TripID {
			if i > 0 {
				prevTrip = allCrewTrips[i-1]
			}
			break
		}
	}

	if prevTrip != nil {
		prevTripEndInPrefLoc := prevTrip.CalculatedDutyPeriodEnd.In(preferredLocation)
		currentTripStartInPrefLoc := trip.CalculatedDutyPeriodStart.In(preferredLocation)

		trip.CalculatedRestPeriodStart = prevTripEndInPrefLoc
		trip.CalculatedRestPeriodEnd = currentTripStartInPrefLoc
		trip.CalculatedRestPeriodDurationMin = int(trip.CalculatedRestPeriodEnd.Sub(trip.CalculatedRestPeriodStart).Minutes())

		minRestExpectedMin := 0
		if trip.DutyStartAirport == "IST" || trip.DutyStartAirport == "SAW" || trip.DutyStartAirport == "ISL" {
			minRestExpectedMin = 12 * 60
		} else {
			minRestExpectedMin = 10 * 60
		}

		if trip.CalculatedRestPeriodDurationMin < minRestExpectedMin {
			trip.FTLViolations = append(trip.FTLViolations, fmt.Sprintf("MinRestPeriodViolated: Ekip %s için dinlenme süresi %.0f dk, beklenen minimum %.0f dk",
				trip.CrewMemberID, float64(trip.CalculatedRestPeriodDurationMin), float64(minRestExpectedMin)))
		}
	} else {
		log.Printf("Bilgi: Ekip %s için %s ID'li görevden önce önceki bir görev bulunamadı. Dinlenme süresi hesaplanmadı.", trip.CrewMemberID, trip.TripID)
	}

	nowInPrefLoc := trip.CalculatedDutyPeriodEnd.In(preferredLocation)

	sevenDaysAgo := nowInPrefLoc.AddDate(0, 0, -7)
	current7DayDutyMin := 0
	for _, t := range allCrewTrips {
		// Removed invisible characters here
		if t.CalculatedDutyPeriodEnd.In(preferredLocation).After(sevenDaysAgo) && t.CalculatedDutyPeriodEnd.In(preferredLocation).Before(nowInPrefLoc.Add(1*time.Minute)) {
			current7DayDutyMin += t.CalculatedDutyPeriodDurationMin
		}
	}
	max7DayDutyMin := 60 * 60
	if current7DayDutyMin > max7DayDutyMin {
		trip.FTLViolations = append(trip.FTLViolations, fmt.Sprintf("MaxDutyPeriod7DaysExceeded: Ekip %s için 7 günlük görev süresi %.0f saat, limit %.0f saat",
			trip.CrewMemberID, float64(current7DayDutyMin)/60, float64(max7DayDutyMin)/60))
	}

	fourteenDaysAgo := nowInPrefLoc.AddDate(0, 0, -14)
	current14DayDutyMin := 0
	for _, t := range allCrewTrips {
		// Removed invisible characters here
		if t.CalculatedDutyPeriodEnd.In(preferredLocation).After(fourteenDaysAgo) && t.CalculatedDutyPeriodEnd.In(preferredLocation).Before(nowInPrefLoc.Add(1*time.Minute)) {
			current14DayDutyMin += t.CalculatedDutyPeriodDurationMin
		}
	}
	max14DayDutyMin := 110 * 60
	if current14DayDutyMin > max14DayDutyMin {
		trip.FTLViolations = append(trip.FTLViolations, fmt.Sprintf("MaxDutyPeriod14DaysExceeded: Ekip %s için 14 günlük görev süresi %.0f saat, limit %.0f saat",
			trip.CrewMemberID, float64(current14DayDutyMin)/60, float64(max14DayDutyMin)/60))
	}

	twentyEightDaysAgo := nowInPrefLoc.AddDate(0, 0, -28)
	current28DayDutyMin := 0
	for _, t := range allCrewTrips {
		// Removed invisible characters here
		if t.CalculatedDutyPeriodEnd.In(preferredLocation).After(twentyEightDaysAgo) && t.CalculatedDutyPeriodEnd.In(preferredLocation).Before(nowInPrefLoc.Add(1*time.Minute)) {
			current28DayDutyMin += t.CalculatedDutyPeriodDurationMin
		}
	}
	max28DayDutyMin := 190 * 60
	if current28DayDutyMin > max28DayDutyMin {
		trip.FTLViolations = append(trip.FTLViolations, fmt.Sprintf("MaxDutyPeriod28DaysExceeded: Ekip %s için 28 günlük görev süresi %.0f saat, limit %.0f saat",
			trip.CrewMemberID, float64(current28DayDutyMin)/60, float64(max28DayDutyMin)/60))
	}

	currentYear := nowInPrefLoc.Year()
	currentYearDutyMin := 0
	for _, t := range allCrewTrips {
		if t.CalculatedDutyPeriodEnd.In(preferredLocation).Year() == currentYear {
			currentYearDutyMin += t.CalculatedDutyPeriodDurationMin
		}
	}
	maxYearDutyMin := 2000 * 60
	if currentYearDutyMin > maxYearDutyMin {
		trip.FTLViolations = append(trip.FTLViolations, fmt.Sprintf("MaxDutyPeriodYearExceeded: Ekip %s için yıllık görev süresi %.0f saat, limit %.0f saat",
			trip.CrewMemberID, float64(currentYearDutyMin)/60, float64(maxYearDutyMin)/60))
	}

	current28DayFlightMin := 0
	for _, t := range allCrewTrips {
		// Removed invisible characters here
		if t.LastLegArrivalTime.In(preferredLocation).After(twentyEightDaysAgo) && t.LastLegArrivalTime.In(preferredLocation).Before(nowInPrefLoc.Add(1*time.Minute)) {
			current28DayFlightMin += t.CalculatedFlightDutyPeriodDurationMin
		}
	}
	max28DayFlightMin := 100 * 60
	if current28DayFlightMin > max28DayFlightMin {
		trip.FTLViolations = append(trip.FTLViolations, fmt.Sprintf("MaxFlightTime28DaysExceeded: Ekip %s için 28 günlük uçuş süresi %.0f saat, limit %.0f saat",
			trip.CrewMemberID, float64(current28DayFlightMin)/60, float64(max28DayFlightMin)/60))
	}

	twelveMonthsAgo := nowInPrefLoc.AddDate(0, -12, 0)
	current12MonthFlightMin := 0
	for _, t := range allCrewTrips {
		// Removed invisible characters here
		if t.LastLegArrivalTime.In(preferredLocation).After(twelveMonthsAgo) && t.LastLegArrivalTime.In(preferredLocation).Before(nowInPrefLoc.Add(1*time.Minute)) {
			current12MonthFlightMin += t.CalculatedFlightDutyPeriodDurationMin
		}
	}
	max12MonthFlightMin := 1000 * 60
	if current12MonthFlightMin > max12MonthFlightMin {
		trip.FTLViolations = append(trip.FTLViolations, fmt.Sprintf("MaxFlightTime12MonthsExceeded: Ekip %s için 12 aylık uçuş süresi %.0f saat, limit %.0f saat",
			trip.CrewMemberID, float64(current12MonthFlightMin)/60, float64(max12MonthFlightMin)/60))
	}

	currentYearFlightMin := 0
	for _, t := range allCrewTrips {
		if t.LastLegArrivalTime.In(preferredLocation).Year() == currentYear {
			currentYearFlightMin += t.CalculatedFlightDutyPeriodDurationMin
		}
	}
	maxYearFlightMin := 900 * 60
	if currentYearFlightMin > maxYearFlightMin {
		trip.FTLViolations = append(trip.FTLViolations, fmt.Sprintf("MaxFlightTimeYearExceeded: Ekip %s için yıllık uçuş süresi %.0f saat, limit %.0f saat",
			trip.CrewMemberID, float64(currentYearFlightMin)/60, float64(maxYearFlightMin)/60))
	}

	f.ApplyMaxDailyUGSLimit(trip)

	trip.LastCalculatedAt = time.Now()
	return nil
}

// --- Yardımcı Fonksiyonlar ---

// getMaxDailyUGSTable5, intibak edilmiş ekip üyeleri için Tablo-5'ten azami UGS limitini döndürür.
// Fonksiyonun ilk harfini büyük yaparak public yapıyoruz.
func (f *FTLCalculator) GetMaxDailyUGSTable5(startTime time.Time, numSectors int) (int, error) {
	startHour := startTime.Hour()
	startMinute := startTime.Minute()

	// Tablo-5'deki zaman aralıklarına göre mantık
	if (startHour >= 6 && startHour < 13) || (startHour == 13 && startMinute < 30) { // 06:00-13:29
		switch numSectors {
		case 1, 2:
			return 13 * 60, nil
		case 3:
			return 12*60 + 30, nil
		case 4:
			return 12 * 60, nil
		case 5:
			return 11*60 + 30, nil
		case 6:
			return 11 * 60, nil
		case 7:
			return 10*60 + 30, nil
		case 8:
			return 10 * 60, nil
		case 9:
			return 9*60 + 30, nil
		case 10:
			return 9 * 60, nil
		}
	} else if (startHour == 13 && startMinute >= 30) || (startHour == 13 && startMinute < 60) { // 13:30-13:59
		switch numSectors {
		case 1, 2:
			return 12*60 + 45, nil
		case 3:
			return 12*60 + 15, nil
		case 4:
			return 11*60 + 45, nil
		case 5:
			return 11*60 + 15, nil
		case 6:
			return 10*60 + 45, nil
		case 7:
			return 10*60 + 15, nil
		case 8:
			return 9*60 + 45, nil
		case 9:
			return 9*60 + 15, nil
		case 10:
			return 9 * 60, nil
		}
	} else if (startHour == 14 && startMinute >= 0) || (startHour == 14 && startMinute < 30) { // 14:00-14:29
		switch numSectors {
		case 1, 2:
			return 12*60 + 30, nil
		case 3:
			return 12 * 60, nil
		case 4:
			return 11*60 + 30, nil
		case 5:
			return 11 * 60, nil
		case 6:
			return 10*60 + 30, nil
		case 7:
			return 10 * 60, nil
		case 8:
			return 9*60 + 30, nil
		case 9, 10:
			return 9 * 60, nil
		}
	} else if (startHour == 14 && startMinute >= 30) || (startHour == 14 && startMinute < 60) { // 14:30-14:59
		switch numSectors {
		case 1, 2:
			return 12*60 + 15, nil
		case 3:
			return 11*60 + 45, nil
		case 4:
			return 11*60 + 15, nil
		case 5:
			return 10*60 + 45, nil
		case 6:
			return 10*60 + 15, nil
		case 7:
			return 9*60 + 45, nil
		case 8:
			return 9*60 + 15, nil
		case 9, 10:
			return 9 * 60, nil
		}
	} else if (startHour == 15 && startMinute >= 0) || (startHour == 15 && startMinute < 30) { // 15:00-15:29
		switch numSectors {
		case 1, 2:
			return 12 * 60, nil
		case 3:
			return 11*60 + 30, nil
		case 4:
			return 11 * 60, nil
		case 5:
			return 10*60 + 30, nil
		case 6:
			return 10 * 60, nil
		case 7:
			return 9*60 + 30, nil
		case 8, 9, 10:
			return 9 * 60, nil
		}
	} else if (startHour == 15 && startMinute >= 30) || (startHour == 15 && startMinute < 60) { // 15:30-15:59
		switch numSectors {
		case 1, 2:
			return 11*60 + 45, nil
		case 3:
			return 11*60 + 15, nil
		case 4:
			return 10*60 + 45, nil
		case 5:
			return 10*60 + 15, nil
		case 6:
			return 9*60 + 45, nil
		case 7, 8, 9, 10:
			return 9 * 60, nil
		}
	} else if (startHour == 16 && startMinute >= 0) || (startHour == 16 && startMinute < 30) { // 16:00-16:29
		switch numSectors {
		case 1, 2:
			return 11*60 + 30, nil
		case 3:
			return 11 * 60, nil
		case 4:
			return 10*60 + 30, nil
		case 5:
			return 10 * 60, nil
		case 6, 7, 8, 9, 10:
			return 9 * 60, nil
		}
	} else if (startHour == 16 && startMinute >= 30) || (startHour == 16 && startMinute < 60) { // 16:30-16:59
		switch numSectors {
		case 1, 2:
			return 11*60 + 15, nil
		case 3:
			return 10*60 + 45, nil
		case 4:
			return 10*60 + 15, nil
		case 5:
			return 9*60 + 45, nil
		case 6, 7, 8, 9, 10:
			return 9 * 60, nil
		}
	} else if startHour >= 17 || startHour < 5 { // 17:00-04:59 (Gece görevleri)
		switch numSectors {
		case 1, 2:
			return 11 * 60, nil
		case 3:
			return 10*60 + 30, nil
		case 4:
			return 10 * 60, nil
		case 5:
			return 9*60 + 30, nil
		case 6, 7, 8, 9, 10:
			return 9 * 60, nil
		}
	} else if (startHour == 5 && startMinute >= 0) || (startHour == 5 && startMinute < 15) { // 05:00-05:14
		switch numSectors {
		case 1, 2:
			return 12 * 60, nil
		case 3:
			return 11*60 + 30, nil
		case 4:
			return 11 * 60, nil
		case 5:
			return 10*60 + 30, nil
		case 6:
			return 10 * 60, nil
		case 7:
			return 9*60 + 30, nil
		case 8, 9, 10:
			return 9 * 60, nil
		}
	} else if (startHour == 5 && startMinute >= 15) || (startHour == 5 && startMinute < 30) { // 05:15-05:29
		switch numSectors {
		case 1, 2:
			return 12*60 + 15, nil
		case 3:
			return 11*60 + 45, nil
		case 4:
			return 11*60 + 15, nil
		case 5:
			return 10*60 + 45, nil
		case 6:
			return 10*60 + 15, nil
		case 7:
			return 9*60 + 45, nil
		case 8, 9, 10:
			return 9 * 60, nil
		}
	} else if (startHour == 5 && startMinute >= 30) || (startHour == 5 && startMinute < 45) { // 05:30-05:44
		switch numSectors {
		case 1, 2:
			return 12*60 + 30, nil
		case 3:
			return 12 * 60, nil
		case 4:
			return 11*60 + 30, nil
		case 5:
			return 11 * 60, nil
		case 6:
			return 10*60 + 30, nil
		case 7:
			return 10 * 60, nil
		case 8, 9, 10:
			return 9 * 60, nil
		}
	} else if (startHour == 5 && startMinute >= 45) || (startHour == 5 && startMinute < 60) { // 05:45-05:59
		switch numSectors {
		case 1, 2:
			return 12*60 + 45, nil
		case 3:
			return 12*60 + 15, nil
		case 4:
			return 11*60 + 45, nil
		case 5:
			return 11*60 + 15, nil
		case 6:
			return 10*60 + 45, nil
		case 7, 8, 9, 10:
			return 10*60 + 15, nil
		}
	}
	return 0, fmt.Errorf("geçerli azami UGS Tablo-5 limiti bulunamadı: Başlangıç: %v, Sektör: %d", startTime.Format("15:04"), numSectors)
}

// ApplyMaxDailyUGSLimit kontrolü: Günlük azami UGS limitini uygular.
// Fonksiyonun ilk harfini büyük yaparak public yapıyoruz.
func (f *FTLCalculator) ApplyMaxDailyUGSLimit(trip *models.Trip) {
	numSectors := 0
	for _, activity := range trip.Activities {
		if activity.GroupCode == "FLT" {
			numSectors++
		}
	}
	if numSectors == 0 {
		return
	}

	maxUGSLimitMin, err := f.GetMaxDailyUGSTable5(trip.CalculatedDutyPeriodStart, numSectors)
	if err != nil {
		trip.FTLViolations = append(trip.FTLViolations, fmt.Sprintf("MaxDailyUGSLimitError: %v", err))
		return
	}

	if trip.CalculatedFlightDutyPeriodDurationMin > maxUGSLimitMin {
		trip.FTLViolations = append(trip.FTLViolations, fmt.Sprintf("MaxDailyUGSLimitViolated: Ekip %s için günlük UGS %.0f saat, limit %.0f saat (Sektör: %d)",
			trip.CrewMemberID, float64(trip.CalculatedFlightDutyPeriodDurationMin)/60, float64(maxUGSLimitMin)/60, numSectors))
	}
}

// RecalculateCrewSchedule, belirli bir ekip üyesinin tüm actual kayıtlarını çekip,
// bunları trip_id'ye göre gruplayarak FTL hesaplamalarını yapar ve trips tablosuna kaydeder.
func (f *FTLCalculator) RecalculateCrewSchedule(crewID string) error {
	log.Printf("Bilgi: Ekip %s için tüm program FTL hesaplaması başlatıldı.", crewID)

	// Fix: Add context.Background() and a lookbackDays (e.g., 365 for a year)
	actuals, err := f.actualRepo.GetActualsByPersonID(context.Background(), crewID, time.Now().AddDate(-1, 0, -28), 365)
	if err != nil {
		return fmt.Errorf("ekip %s için actual kayıtları çekilemedi: %w", crewID, err)
	}

	if len(actuals) == 0 {
		log.Printf("Bilgi: Ekip %s için hesaplanacak actual kayıt bulunamadı.", crewID)
		return nil
	}

	tripsMap := make(map[string]*models.Trip)
	for _, act := range actuals {
		if act.TripID == "" {
			log.Printf("Uyarı: TripID'si boş olan aktivite atlandı: %s (PersonID: %s, ActivityCode: %s)", act.DataID.String(), act.PersonID, act.ActivityCode)
			continue
		}

		if _, ok := tripsMap[act.TripID]; !ok {
			trip := &models.Trip{
				TripID:       act.TripID,
				CrewMemberID: act.PersonID,
				Activities:   []models.Actual{},
			}
			tripsMap[act.TripID] = trip
		}
		tripsMap[act.TripID].Activities = append(tripsMap[act.TripID].Activities, act)
	}

	var sortedTrips []*models.Trip
	for _, trip := range tripsMap {
		sort.Slice(trip.Activities, func(i, j int) bool {
			return trip.Activities[i].DutyStart.Before(trip.Activities[j].DutyStart)
		})

		var firstFlightTime, lastFlightTime time.Time
		var firstFlightFound, lastFlightFound bool
		var firstFLTActivity *models.Actual // BriefTripType için ilk FLT aktivitesi
		var lastFLTActivity *models.Actual  // DebriefTripType için son FLT aktivitesi

		for _, act := range trip.Activities {
			if act.GroupCode == "FLT" {
				if !firstFlightFound {
					firstFlightTime = act.DepartureTime
					firstFlightFound = true
					firstFLTActivity = &act
				}
				lastFlightTime = act.ArrivalTime
				lastFlightFound = true
				lastFLTActivity = &act
			}
		}

		if !firstFlightFound || !lastFlightFound {
			log.Printf("Uyarı: Trip %s için uçuş aktivitesi bulunamadı, FirstLegDepartureTime/LastLegArrivalTime belirlenemiyor. DutyStart/DutyEnd kullanılıyor.", trip.TripID)
			if len(trip.Activities) > 0 {
				trip.FirstLegDepartureTime = trip.Activities[0].DutyStart
				trip.LastLegArrivalTime = trip.Activities[len(trip.Activities)-1].DutyEnd
			} else {
				log.Printf("Hata: Trip %s için hiç aktivite bulunamadı, zaman bilgileri doldurulamıyor. Trip atlandı.", trip.TripID)
				continue
			}
		} else {
			trip.FirstLegDepartureTime = firstFlightTime
			trip.LastLegArrivalTime = lastFlightTime
		}

		// briefTripType ve debriefTripType'ı belirle
		if firstFLTActivity != nil {
			trip.BriefTripType = models.GetDutyTypeFromActual(firstFLTActivity)
		} else if len(trip.Activities) > 0 {
			trip.BriefTripType = models.GetDutyTypeFromActual(&trip.Activities[0])
		} else {
			trip.BriefTripType = "BİLİNMİYOR"
		}

		if lastFLTActivity != nil {
			trip.DebriefTripType = models.GetDutyTypeFromActual(lastFLTActivity)
		} else if len(trip.Activities) > 0 {
			trip.DebriefTripType = models.GetDutyTypeFromActual(&trip.Activities[len(trip.Activities)-1])
		} else {
			trip.DebriefTripType = "BİLİNMİYOR"
		}

		// Diğer genel trip özellikleri
		if len(trip.Activities) > 0 {
			firstAct := &trip.Activities[0]
			// trip.AircraftType = models.GetAircraftTypeFromCmsType(firstAct.PlaneCmsType) // <<< Bu satır kaldırılmalı
			trip.DutyStartAirport = firstAct.DeparturePort
			trip.DutyType = models.GetDutyTypeFromActual(firstAct) // Genel DutyType
			trip.CrewType = models.GetCrewTypeFromFlightPosition(firstAct.FlightPosition)
		} else {
			// trip.AircraftType = "BİLİNMİYOR" // <<< Bu satır kaldırılmalı
			trip.DutyStartAirport = "BİLİNMİYOR"
			trip.DutyType = "BİLİNMİYOR"
			trip.CrewType = "BİLİNMİYOR"
		}

		// <<<< BURADA EKLENDİ: BriefAircraftType ve DebriefAircraftType'ı belirle
		// BriefAircraftType
		if firstFLTActivity != nil {
			trip.BriefAircraftType = models.GetAircraftTypeFromCmsType(firstFLTActivity.PlaneCmsType)
		} else if len(trip.Activities) > 0 {
			// Eğer hiç FLT aktivitesi yoksa, ilk aktiviteden genel uçak tipini al
			trip.BriefAircraftType = models.GetAircraftTypeFromCmsType(trip.Activities[0].PlaneCmsType)
		} else {
			trip.BriefAircraftType = "BİLİNMİYOR"
		}

		// DebriefAircraftType
		if lastFLTActivity != nil {
			trip.DebriefAircraftType = models.GetAircraftTypeFromCmsType(lastFLTActivity.PlaneCmsType)
		} else if len(trip.Activities) > 0 {
			// Eğer hiç FLT aktivitesi yoksa, son aktiviteden genel uçak tipini al
			trip.DebriefAircraftType = models.GetAircraftTypeFromCmsType(trip.Activities[len(trip.Activities)-1].PlaneCmsType)
		} else {
			trip.DebriefAircraftType = "BİLİNMİYOR"
		}
		// Trip'in genel AircraftType'ı için, eğer hala isteniyorsa, burada bir değer atanabilir.
		// Örneğin, briefAircraftType veya debriefAircraftType'tan biri seçilerek.
		// trip.AircraftType = trip.BriefAircraftType (veya başka bir mantık)

		sortedTrips = append(sortedTrips, trip)
	}

	sort.Slice(sortedTrips, func(i, j int) bool {
		return sortedTrips[i].FirstLegDepartureTime.Before(sortedTrips[j].FirstLegDepartureTime)
	})

	for _, trip := range sortedTrips {
		if err := f.CalculateFTLForTrip(trip, sortedTrips); err != nil {
			log.Printf("Hata: Trip %s için FTL hesaplanırken sorun: %v", trip.TripID, err)
		}
		if err := f.tripRepo.SaveTrip(trip); err != nil {
			log.Printf("Hata: Trip %s FTL hesaplaması sonrası kaydedilirken sorun: %v", trip.TripID, err)
		}
	}

	log.Printf("✅ Ekip %s için tüm program FTL hesaplaması tamamlandı.", crewID)
	return nil
}
