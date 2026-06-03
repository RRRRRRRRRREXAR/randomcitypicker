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
	if ok && pageType != "disambiguation" {
		return summary, imageURL, nil
	}

	// 2. Fall back to geosearch using coordinates.
	bestTitle := fetchGeoSearchTitle(lat, lon)
	if bestTitle != "" && bestTitle != cityName {
		summary, imageURL, _, ok = fetchPageSummary(bestTitle)
		if ok {
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

// fetchGeoSearchTitle queries Wikipedia's geosearch API and returns the title
// of the nearest article to the given coordinates.
func fetchGeoSearchTitle(lat, lon float64) string {
	apiURL := fmt.Sprintf(
		"https://en.wikipedia.org/w/api.php?action=query&list=geosearch&gscoord=%.6f|%.6f&gsradius=10000&gslimit=1&format=json",
		lat, lon,
	)
	req, err := http.NewRequest("GET", apiURL, nil)
	if err != nil {
		return ""
	}
	req.Header.Set("User-Agent", "RandomCityPicker/1.0")

	resp, err := httpClient.Do(req)
	if err != nil {
		return ""
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return ""
	}

	var data geoSearchResponse
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return ""
	}

	if len(data.Query.GeoSearch) == 0 {
		return ""
	}
	return data.Query.GeoSearch[0].Title
}
