package models

import "time"

type City struct {
	ID          int     `json:"id"`
	Name        string  `json:"name"`
	CountryCode string  `json:"country_code"`
	Population  int     `json:"population"`
	Latitude    float64 `json:"latitude"`
	Longitude   float64 `json:"longitude"`
}

type CityPick struct {
	CityID         int       `json:"city_id"`
	PickCount      int       `json:"pick_count"`
	FirstPickedAt  time.Time `json:"first_picked_at"`
	LastPickedAt   time.Time `json:"last_picked_at"`
	City           *City     `json:"city,omitempty"`
}

type RandomCityResponse struct {
	City      City   `json:"city"`
	PickCount int    `json:"pick_count"`
	Summary   string `json:"summary,omitempty"`
	ImageURL  string `json:"image_url,omitempty"`
}
