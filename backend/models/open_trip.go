package models

type OpenTripNeed struct {
	TripID       string         `json:"trip_id"`
	FlightKey    string         `json:"flight_key"`
	AircraftType string         `json:"aircraft_type"`
	Status       string         `json:"status"` // exact | under | over
	Assigned     map[string]int `json:"assigned"`
	Required     map[string]int `json:"required"`
	Diff         map[string]int `json:"diff"`
}
