package db

import (
	"fmt"
	"strings"

	"randomcitypicker/models"
)

// SeedIfEmpty seeds France if the cities table is completely empty.
// It is kept for backward compatibility.
func SeedIfEmpty() error {
	var count int
	if err := DB.QueryRow("SELECT COUNT(*) FROM cities").Scan(&count); err != nil {
		return fmt.Errorf("count cities: %w", err)
	}
	if count > 0 {
		return nil
	}
	return SeedCountry("FR")
}

// SeedCountry downloads and inserts all populated places for a given ISO country code.
// It is idempotent: if the country already has rows, it skips.
func SeedCountry(countryCode string) error {
	countryCode = strings.ToUpper(strings.TrimSpace(countryCode))
	if countryCode == "" {
		return fmt.Errorf("empty country code")
	}

	var count int
	if err := DB.QueryRow("SELECT COUNT(*) FROM cities WHERE country_code = ?", countryCode).Scan(&count); err != nil {
		return fmt.Errorf("count cities for %s: %w", countryCode, err)
	}
	if count > 0 {
		fmt.Printf("Country %s already seeded (%d settlements), skipping\n", countryCode, count)
		return nil
	}

	cities, err := DownloadCountry(countryCode)
	if err != nil {
		return fmt.Errorf("download cities for %s: %w", countryCode, err)
	}

	tx, err := DB.Begin()
	if err != nil {
		return fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback()

	stmt, err := tx.Prepare("INSERT INTO cities (name, country_code, population, latitude, longitude) VALUES (?, ?, ?, ?, ?)")
	if err != nil {
		return fmt.Errorf("prepare insert: %w", err)
	}
	defer stmt.Close()

	for _, c := range cities {
		if _, err := stmt.Exec(c.Name, countryCode, c.Population, c.Latitude, c.Longitude); err != nil {
			return fmt.Errorf("insert city %s: %w", c.Name, err)
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit tx: %w", err)
	}

	fmt.Printf("Seeded %d settlements for %s\n", len(cities), countryCode)
	return nil
}

// SeedCountries seeds multiple countries in order.
func SeedCountries(codes []string) error {
	for _, code := range codes {
		if err := SeedCountry(code); err != nil {
			return err
		}
	}
	return nil
}

// GetCountries returns the distinct country codes present in the database.
func GetCountries() ([]string, error) {
	rows, err := DB.Query("SELECT DISTINCT country_code FROM cities ORDER BY country_code")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var codes []string
	for rows.Next() {
		var code string
		if err := rows.Scan(&code); err != nil {
			return nil, err
		}
		codes = append(codes, code)
	}
	return codes, rows.Err()
}

func GetCities(countryCode string, minPop, maxPop int) ([]models.City, error) {
	rows, err := DB.Query(
		"SELECT id, name, country_code, population, latitude, longitude FROM cities WHERE country_code = ? AND population >= ? AND population <= ? ORDER BY name",
		countryCode, minPop, maxPop,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var cities []models.City
	for rows.Next() {
		var c models.City
		if err := rows.Scan(&c.ID, &c.Name, &c.CountryCode, &c.Population, &c.Latitude, &c.Longitude); err != nil {
			return nil, err
		}
		cities = append(cities, c)
	}
	return cities, rows.Err()
}

func PickRandomCity(countryCode string, minPop, maxPop int) (*models.RandomCityResponse, error) {
	var city models.City
	var pickCount int

	err := DB.QueryRow(`
		SELECT c.id, c.name, c.country_code, c.population, c.latitude, c.longitude,
		       COALESCE(cp.pick_count, 0)
		FROM cities c
		LEFT JOIN city_picks cp ON c.id = cp.city_id
		WHERE c.country_code = ? AND c.population >= ? AND c.population <= ?
		ORDER BY COALESCE(cp.pick_count, 0) ASC, RANDOM()
		LIMIT 1
	`, countryCode, minPop, maxPop).Scan(
		&city.ID, &city.Name, &city.CountryCode, &city.Population, &city.Latitude, &city.Longitude,
		&pickCount,
	)
	if err != nil {
		return nil, fmt.Errorf("pick city: %w", err)
	}

	return &models.RandomCityResponse{
		City:      city,
		PickCount: pickCount,
	}, nil
}

func ConfirmPick(cityID int) (int, error) {
	_, err := DB.Exec(`
		INSERT INTO city_picks (city_id, pick_count, first_picked_at, last_picked_at)
		VALUES (?, 1, datetime('now'), datetime('now'))
		ON CONFLICT(city_id) DO UPDATE SET
			pick_count = pick_count + 1,
			last_picked_at = datetime('now')
	`, cityID)
	if err != nil {
		return 0, fmt.Errorf("confirm pick: %w", err)
	}

	var pickCount int
	err = DB.QueryRow("SELECT pick_count FROM city_picks WHERE city_id = ?", cityID).Scan(&pickCount)
	if err != nil {
		return 0, fmt.Errorf("get pick count: %w", err)
	}

	return pickCount, nil
}

func GetPickedCities() ([]models.CityPick, error) {
	rows, err := DB.Query(`
		SELECT cp.city_id, cp.pick_count, cp.first_picked_at, cp.last_picked_at,
		       c.name, c.country_code, c.population, c.latitude, c.longitude
		FROM city_picks cp
		JOIN cities c ON cp.city_id = c.id
		ORDER BY cp.last_picked_at DESC
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var picks []models.CityPick
	for rows.Next() {
		var cp models.CityPick
		var c models.City
		if err := rows.Scan(
			&cp.CityID, &cp.PickCount, &cp.FirstPickedAt, &cp.LastPickedAt,
			&c.Name, &c.CountryCode, &c.Population, &c.Latitude, &c.Longitude,
		); err != nil {
			return nil, err
		}
		c.ID = cp.CityID
		cp.City = &c
		picks = append(picks, cp)
	}
	return picks, rows.Err()
}

func ResetPicks() error {
	_, err := DB.Exec("DELETE FROM city_picks")
	return err
}
