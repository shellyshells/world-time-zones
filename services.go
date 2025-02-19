package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"math"
	"net/http"
	"os"
	"sort"
	"strconv"
	"strings"
	"time"
)

func fetchCountries() ([]Country, error) {
	// Read HDI data from the CSV file
	hdiContent, err := os.ReadFile("HDR23-24_Statistical_Annex_HDI_Table - HDI.csv")
	if err != nil {
		log.Printf("Warning: Could not load HDI data: %v", err)
	}

	hdiMap := parseHDIData(string(hdiContent))

	// Fetch countries from the REST API
	resp, err := http.Get("https://restcountries.com/v3.1/all?fields=name,capital,region,flag,timezones,population,area,languages,currencies,idd,car,borders")
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var rawCountries []struct {
		Name struct {
			Common string `json:"common"`
		} `json:"name"`
		Capital    []string               `json:"capital"`
		Region     string                 `json:"region"`
		Flag       string                 `json:"flag"`
		TimeZones  []string               `json:"timezones"`
		Population int                    `json:"population"`
		Area       float64                `json:"area"`
		Languages  map[string]string      `json:"languages"`
		Currencies map[string]interface{} `json:"currencies"`
		IDD        struct {
			Root     string   `json:"root"`
			Suffixes []string `json:"suffixes"`
		} `json:"idd"`
		Car struct {
			Side string `json:"side"`
		} `json:"car"`
		Borders []string `json:"borders"`
	}

	if err := json.Unmarshal(body, &rawCountries); err != nil {
		return nil, err
	}

	countries := make([]Country, 0, len(rawCountries))
	for _, rc := range rawCountries {
		capital := ""
		if len(rc.Capital) > 0 {
			capital = rc.Capital[0]
		}

		mainTimeZone := "UTC"
		if len(rc.TimeZones) > 0 {
			mainTimeZone = rc.TimeZones[0]
		}

		currentTime := calculateTime(mainTimeZone)

		population := formatNumber(rc.Population)

		var languages []string
		for _, lang := range rc.Languages {
			languages = append(languages, lang)
		}

		currency := ""
		for _, curr := range rc.Currencies {
			if c, ok := curr.(map[string]interface{}); ok {
				if name, ok := c["name"].(string); ok {
					currency = name
					break
				}
			}
		}

		callingCode := ""
		if rc.IDD.Root != "" {
			if len(rc.IDD.Suffixes) > 0 {
				callingCode = rc.IDD.Root + rc.IDD.Suffixes[0]
			} else {
				callingCode = rc.IDD.Root
			}
		}

		// Add HDI data if available
		hdiData, hasHDI := hdiMap[rc.Name.Common]
		if !hasHDI {
			hdiData = HDIData{} // Empty HDI data if not found
		}

		country := Country{
			Name:        rc.Name.Common,
			TimeZone:    mainTimeZone,
			Capital:     capital,
			Region:      rc.Region,
			Flag:        rc.Flag,
			TimeZones:   rc.TimeZones,
			CurrentTime: currentTime,
			IsFavorite:  contains(favorites.Countries, rc.Name.Common),
			Population:  population,
			Area:        rc.Area,
			Languages:   languages,
			Currency:    currency,
			CallingCode: callingCode,
			DrivingSide: strings.Title(rc.Car.Side),
			Borders:     rc.Borders,
			HDI:         hdiData,
		}
		countries = append(countries, country)
	}

	return countries, nil
}

func parseHDIData(csvContent string) map[string]HDIData {
	hdiMap := make(map[string]HDIData)
	lines := strings.Split(csvContent, "\n")

	log.Printf("Starting HDI data parsing with %d lines", len(lines))

	var currentCategory string
	for _, line := range lines {
		// Skip empty lines
		if strings.TrimSpace(line) == "" {
			continue
		}

		// Split the line by comma, handling quoted fields
		fields := strings.Split(line, ",")

		// Clean up each field
		for i := range fields {
			fields[i] = strings.Trim(fields[i], `" `)
		}

		// Check if this is a category line
		if strings.Contains(fields[0], "VERY HIGH HUMAN DEVELOPMENT") ||
			strings.Contains(fields[0], "HIGH HUMAN DEVELOPMENT") ||
			strings.Contains(fields[0], "MEDIUM HUMAN DEVELOPMENT") ||
			strings.Contains(fields[0], "LOW HUMAN DEVELOPMENT") {
			currentCategory = fields[0]
			log.Printf("Found category: %s", currentCategory)
			continue
		}

		// Try to parse rank number to verify if this is a country line
		rank := strings.TrimSpace(fields[0])
		rankNum, err := strconv.Atoi(rank)
		if err != nil {
			continue // Not a country line
		}

		countryName := strings.TrimSpace(fields[1])
		if countryName == "" {
			continue
		}

		// Parse HDI values, with better error handling
		hdiValue, err := strconv.ParseFloat(strings.Trim(fields[2], `" `), 64)
		if err != nil {
			log.Printf("Warning: Could not parse HDI value for %s: %v", countryName, err)
			continue
		}

		lifeExpect, _ := strconv.ParseFloat(strings.Trim(fields[4], `" `), 64)
		schoolYears, _ := strconv.ParseFloat(strings.Trim(fields[6], `" `), 64)
		gniPerCap := strings.Trim(fields[8], `" `)

		log.Printf("Successfully parsed HDI data for country: %s (rank: %d, value: %.3f)",
			countryName, rankNum, hdiValue)

		hdiMap[countryName] = HDIData{
			HDIRank:     rankNum,
			HDIValue:    hdiValue,
			Category:    currentCategory,
			LifeExpect:  lifeExpect,
			SchoolYears: schoolYears,
			GNIPerCap:   gniPerCap,
		}
	}

	log.Printf("Finished parsing HDI data. Found data for %d countries", len(hdiMap))

	// Debug output of all country names we have HDI data for
	log.Printf("Countries with HDI data:")
	for countryName := range hdiMap {
		log.Printf("- %s", countryName)
	}

	return hdiMap
}


func filterCountries(countries []Country, region, timezone, timeRange string, w http.ResponseWriter, r *http.Request) []Country {
	if region == "" && timezone == "" && timeRange == "" {
		return countries
	}

	var filtered []Country
	for _, country := range countries {
		if (region == "" || country.Region == region) &&
			(timezone == "" || contains(country.TimeZones, timezone)) &&
			(timeRange == "" || isInTimeRange(country.CurrentTime, timeRange)) {
			filtered = append(filtered, country)
		}
	}

	if len(filtered) == 0 && (timezone != "" || timeRange != "") {
		http.Redirect(w, r, "/error", http.StatusSeeOther)
		return nil
	}

	return filtered
}

func searchCountries(countries []Country, query string) []Country {
	if query == "" {
		return countries
	}

	query = strings.ToLower(query)
	var results []Country

	for _, country := range countries {
		if strings.Contains(strings.ToLower(country.Name), query) ||
			strings.Contains(strings.ToLower(country.Region), query) ||
			strings.Contains(strings.ToLower(country.Capital), query) {
			results = append(results, country)
		}
	}

	return results
}

func paginateCountries(countries []Country, page int) ([]Country, int) {
	totalPages := int(math.Ceil(float64(len(countries)) / float64(itemsPerPage)))
	if page < 1 {
		page = 1
	}
	if page > totalPages {
		page = totalPages
	}

	start := (page - 1) * itemsPerPage
	end := start + itemsPerPage
	if end > len(countries) {
		end = len(countries)
	}

	if start >= len(countries) {
		return []Country{}, totalPages
	}

	return countries[start:end], totalPages
}

func getUniqueRegions(countries []Country) []string {
	regions := make(map[string]bool)
	for _, country := range countries {
		regions[country.Region] = true
	}

	var uniqueRegions []string
	for region := range regions {
		uniqueRegions = append(uniqueRegions, region)
	}
	sort.Strings(uniqueRegions)
	return uniqueRegions
}


// getUniqueTimeZones gets unique time zones from countries and standard list
func getUniqueTimeZones(countries []Country) []string {
	zones := make(map[string]bool)

	// Add all standard time zones first
	for _, tz := range standardTimeZones {
		zones[tz] = true
	}

	// Add any additional zones from countries
	for _, country := range countries {
		for _, tz := range country.TimeZones {
			cleanTZ := strings.TrimSpace(tz)
			if cleanTZ != "" {
				zones[cleanTZ] = true
			}
		}
	}

	var uniqueZones []string
	for zone := range zones {
		uniqueZones = append(uniqueZones, zone)
	}
	sort.Strings(uniqueZones)
	return uniqueZones
}

func calculateTime(timezone string) string {
	offset := strings.TrimPrefix(timezone, "UTC")
	if offset == "" {
		return time.Now().UTC().Format("15:04")
	}

	var hours, minutes int
	fmt.Sscanf(offset, "%d:%d", &hours, &minutes)

	now := time.Now().UTC()
	localTime := now.Add(time.Duration(hours)*time.Hour + time.Duration(minutes)*time.Minute)
	return localTime.Format("15:04")
}

func isInTimeRange(currentTime, timeRange string) bool {
	if timeRange == "" {
		return true
	}
	// Parse the current time
	hours, err := strconv.Atoi(strings.Split(currentTime, ":")[0])
	if err != nil {
		return false
	}
	switch timeRange {
	case "night":
		return hours >= 0 && hours < 6
	case "morning":
		return hours >= 6 && hours < 12
	case "afternoon":
		return hours >= 12 && hours < 18
	case "evening":
		return hours >= 18 && hours < 24
	default:
		return true
	}
}