package services

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"time"
)

// WikipediaSummaryResponse matches the structure returned by the Wikipedia REST API.
type WikipediaSummaryResponse struct {
	Type      string `json:"type"`
	Title     string `json:"title"`
	Extract   string `json:"extract"`
	Thumbnail *struct {
		Source string `json:"source"`
		Width  int    `json:"width"`
		Height int    `json:"height"`
	} `json:"thumbnail,omitempty"`
}

var httpClient = &http.Client{Timeout: 5 * time.Second}

// FetchCitySummary queries the English Wikipedia REST API for a page summary
// and image thumbnail for the given city name.
//
// It first tries an exact title match. If that fails or returns a
// disambiguation page, it falls back to Wikipedia's geosearch API using the
// provided coordinates to find the nearest article.
//
// It returns the summary text, image URL, and any error encountered.
// If nothing is found, the error is nil and the strings are empty.
func FetchCitySummary(cityName string, lat, lon float64) (summary string, imageURL string, err error) {
	// 1. Try exact title match.
	summary, imageURL, pageType, ok := fetchPageSummary(cityName)
	if ok && pageType != "disambiguation" && summary != "" {
		return summary, imageURL, nil
	}

	// 2. Fall back to geosearch using coordinates.
	// Try multiple nearby articles and pick the first with a real summary.
	candidates := fetchGeoSearchTitles(lat, lon, 10)
	for _, title := range candidates {
		if title == "" || title == cityName {
			continue
		}
		summary, imageURL, pageType, ok := fetchPageSummary(title)
		if ok && pageType != "disambiguation" && summary != "" {
			return summary, imageURL, nil
		}
	}

	return "", "", nil
}

// fetchPageSummary fetches a single page summary by exact title.
// ok is true when the API returns HTTP 200 and the response can be decoded.
func fetchPageSummary(title string) (summary string, imageURL string, pageType string, ok bool) {
	apiURL := fmt.Sprintf("https://en.wikipedia.org/api/rest_v1/page/summary/%s", url.PathEscape(title))
	req, err := http.NewRequest("GET", apiURL, nil)
	if err != nil {
		return "", "", "", false
	}
	req.Header.Set("User-Agent", "RandomCityPicker/1.0")

	resp, err := httpClient.Do(req)
	if err != nil {
		return "", "", "", false
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", "", "", false
	}

	var data WikipediaSummaryResponse
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return "", "", "", false
	}

	if data.Thumbnail != nil {
		imageURL = data.Thumbnail.Source
	}
	return data.Extract, imageURL, data.Type, true
}

type geoSearchResponse struct {
	Query struct {
		GeoSearch []struct {
			PageID int     `json:"pageid"`
			Title  string  `json:"title"`
			Lat    float64 `json:"lat"`
			Lon    float64 `json:"lon"`
			Dist   float64 `json:"dist"`
		} `json:"geosearch"`
	} `json:"query"`
}

// fetchGeoSearchTitles queries Wikipedia's geosearch API and returns up to `limit`
// article titles nearest to the given coordinates, ordered by distance.
func fetchGeoSearchTitles(lat, lon float64, limit int) []string {
	apiURL := fmt.Sprintf(
		"https://en.wikipedia.org/w/api.php?action=query&list=geosearch&gscoord=%.6f|%.6f&gsradius=10000&gslimit=%d&format=json",
		lat, lon, limit,
	)
	req, err := http.NewRequest("GET", apiURL, nil)
	if err != nil {
		return nil
	}
	req.Header.Set("User-Agent", "RandomCityPicker/1.0")

	resp, err := httpClient.Do(req)
	if err != nil {
		return nil
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil
	}

	var data geoSearchResponse
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return nil
	}

	titles := make([]string, 0, len(data.Query.GeoSearch))
	for _, item := range data.Query.GeoSearch {
		titles = append(titles, item.Title)
	}
	return titles
}
