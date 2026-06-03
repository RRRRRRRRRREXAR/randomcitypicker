package db

import (
	"archive/zip"
	"bufio"
	"bytes"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"
)

// SeedCity represents a parsed settlement ready for insertion.
type SeedCity struct {
	Name       string
	Population int
	Latitude   float64
	Longitude  float64
}

// DownloadCountry fetches the GeoNames country dump for the given ISO code.
// It retries up to 3 times with exponential backoff on network or HTTP errors.
func DownloadCountry(countryCode string) ([]SeedCity, error) {
	url := fmt.Sprintf("https://download.geonames.org/export/dump/%s.zip", countryCode)

	var body []byte
	var err error
	for attempt := 0; attempt < 3; attempt++ {
		if attempt > 0 {
			time.Sleep(time.Duration(1<<attempt) * time.Second)
		}
		body, err = downloadURL(url)
		if err == nil {
			break
		}
	}
	if err != nil {
		return nil, fmt.Errorf("download %s after retries: %w", countryCode, err)
	}

	return parseGeoNamesZip(bytes.NewReader(body), int64(len(body)), countryCode)
}

func downloadURL(url string) ([]byte, error) {
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("HTTP %d", resp.StatusCode)
	}
	return io.ReadAll(resp.Body)
}

func parseGeoNamesZip(r io.ReaderAt, size int64, countryCode string) ([]SeedCity, error) {
	zr, err := zip.NewReader(r, size)
	if err != nil {
		return nil, fmt.Errorf("open zip: %w", err)
	}

	var cities []SeedCity
	for _, f := range zr.File {
		if !strings.EqualFold(f.Name, countryCode+".txt") {
			continue
		}

		rc, err := f.Open()
		if err != nil {
			return nil, fmt.Errorf("open %s: %w", f.Name, err)
		}
		defer rc.Close()

		cities, err = parseGeoNamesTSV(rc)
		if err != nil {
			return nil, fmt.Errorf("parse %s: %w", f.Name, err)
		}
		break
	}

	return cities, nil
}

// GeoNames columns (tab-delimited):
// 0:geonameid 1:name 2:asciiname 3:alternatenames 4:latitude 5:longitude
// 6:featureClass 7:featureCode 8:countryCode ... 14:population
func parseGeoNamesTSV(r io.Reader) ([]SeedCity, error) {
	var cities []SeedCity
	scanner := bufio.NewScanner(r)
	for scanner.Scan() {
		line := scanner.Text()
		if line == "" {
			continue
		}
		parts := strings.Split(line, "\t")
		if len(parts) < 15 {
			continue // malformed line
		}
		if parts[6] != "P" {
			continue // not a populated place
		}

		name := parts[1]
		lat, _ := strconv.ParseFloat(parts[4], 64)
		lon, _ := strconv.ParseFloat(parts[5], 64)
		pop, _ := strconv.Atoi(parts[14])

		cities = append(cities, SeedCity{
			Name:       name,
			Population: pop,
			Latitude:   lat,
			Longitude:  lon,
		})
	}
	if err := scanner.Err(); err != nil {
		return nil, err
	}
	return cities, nil
}
