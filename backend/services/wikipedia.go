package services

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"sync"
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

// --- in-memory cache ---

type cacheEntry struct {
	summary   string
	imageURL  string
	fetchedAt time.Time
}

var (
	summaryCache = make(map[int]cacheEntry)
	cacheMu      sync.RWMutex
	cacheTTL     = 24 * time.Hour
)

func getCachedSummary(cityID int) (summary string, imageURL string, ok bool) {
	cacheMu.RLock()
	entry, exists := summaryCache[cityID]
	cacheMu.RUnlock()
	if !exists || time.Since(entry.fetchedAt) > cacheTTL {
		return "", "", false
	}
	return entry.summary, entry.imageURL, true
}

func setCachedSummary(cityID int, summary, imageURL string) {
	cacheMu.Lock()
	summaryCache[cityID] = cacheEntry{
		summary:   summary,
		imageURL:  imageURL,
		fetchedAt: time.Now(),
	}
	cacheMu.Unlock()
}

// FetchCitySummaryCached returns a cached summary when available; otherwise it
// fetches from Wikipedia, stores the result in memory, and returns it.
func FetchCitySummaryCached(ctx context.Context, cityID int, cityName string, lat, lon float64) (summary string, imageURL string, err error) {
	if s, img, ok := getCachedSummary(cityID); ok {
		return s, img, nil
	}

	summary, imageURL, err = FetchCitySummary(ctx, cityName, lat, lon)
	if err != nil {
		return "", "", err
	}

	setCachedSummary(cityID, summary, imageURL)
	return summary, imageURL, nil
}

// FetchCitySummary queries the English Wikipedia REST API for a page summary
// and image thumbnail for the given city name.
//
// It first tries an exact title match. If that fails or returns a
// disambiguation page, it falls back to Wikipedia's geosearch API using the
// provided coordinates to find the nearest article.
//
// It returns the summary text, image URL, and any error encountered.
// If nothing is found, the error is nil and the strings are empty.
func FetchCitySummary(ctx context.Context, cityName string, lat, lon float64) (summary string, imageURL string, err error) {
	// 1. Try exact title match.
	summary, imageURL, pageType, ok := fetchPageSummary(ctx, cityName)
	if ok && pageType != "disambiguation" && summary != "" {
		return summary, imageURL, nil
	}

	// 2. Fall back to geosearch using coordinates.
	// Try multiple nearby articles and pick the first with a real summary.
	candidates := fetchGeoSearchTitles(ctx, lat, lon, 10)
	for i, title := range candidates {
		if title == "" || title == cityName {
			continue
		}
		// Pace requests to be polite to Wikipedia.
		if i > 0 {
			select {
			case <-ctx.Done():
				return "", "", ctx.Err()
			case <-time.After(100 * time.Millisecond):
			}
		}
		summary, imageURL, pageType, ok := fetchPageSummary(ctx, title)
		if ok && pageType != "disambiguation" && summary != "" {
			return summary, imageURL, nil
		}
	}

	return "", "", nil
}

// fetchPageSummary fetches a single page summary by exact title.
// ok is true when the API returns HTTP 200 and the response can be decoded.
func fetchPageSummary(ctx context.Context, title string) (summary string, imageURL string, pageType string, ok bool) {
	apiURL := fmt.Sprintf("https://en.wikipedia.org/api/rest_v1/page/summary/%s", url.PathEscape(title))
	req, err := http.NewRequest("GET", apiURL, nil)
	if err != nil {
		return "", "", "", false
	}
	req.Header.Set("User-Agent", "RandomCityPicker/1.0")
	req = req.WithContext(ctx)

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
func fetchGeoSearchTitles(ctx context.Context, lat, lon float64, limit int) []string {
	apiURL := fmt.Sprintf(
		"https://en.wikipedia.org/w/api.php?action=query&list=geosearch&gscoord=%.6f|%.6f&gsradius=10000&gslimit=%d&format=json",
		lat, lon, limit,
	)
	req, err := http.NewRequest("GET", apiURL, nil)
	if err != nil {
		return nil
	}
	req.Header.Set("User-Agent", "RandomCityPicker/1.0")
	req = req.WithContext(ctx)

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
