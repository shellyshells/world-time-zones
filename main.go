package main

import (
	"encoding/json"
	"fmt"
	"html/template"
	"io"
	"log"
	"math"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"
)

type HDIData struct {
	HDIRank     int     `json:"hdi_rank"`
	HDIValue    float64 `json:"hdi_value"`
	Category    string  `json:"category"`
	LifeExpect  float64 `json:"life_expectancy"`
	SchoolYears float64 `json:"school_years"`
	GNIPerCap   string  `json:"gni_per_capita"`
}

// Country represents the data structure for a country
type Country struct {
	Name        string `json:"name"`
	TimeZone    string
	Capital     string   `json:"capital"`
	Region      string   `json:"region"`
	Flag        string   `json:"flag"`
	TimeZones   []string `json:"timezones"`
	CurrentTime string
	IsFavorite  bool
	Population  string   `json:"population"`
	Area        float64  `json:"area"`
	Languages   []string `json:"languages"`
	Currency    string   `json:"currency"`
	CallingCode string   `json:"callingCode"`
	DrivingSide string   `json:"drivingSide"`
	Borders     []string `json:"borders"`
	// Add HDI information
	HDI HDIData `json:"hdi"`
}

// PageData represents the data structure for page rendering
type PageData struct {
	Countries    []Country
	Query        string
	Regions      []string
	TimeZones    []string
	CurrentPage  int
	TotalPages   int
	ItemsPerPage int
	Region       string
	TimeZone     string
	TimeRange    string
}

// Favorites represents the data structure for favorites storage
type Favorites struct {
	Countries []string `json:"countries"`
}

const (
	itemsPerPage         = 12
	favoritesFile        = "favorites.json"
	timezonesGeojsonPath = "data"
)

var (
	allCountries  []Country
	favorites     Favorites
	templateFuncs = template.FuncMap{
		"subtract": func(a, b int) int {
			return a - b
		},
		"add": func(a, b int) int {
			return a + b
		},
	}
	standardTimeZones = []string{
		"UTC-12:00", "UTC-11:00", "UTC-10:00", "UTC-09:30", "UTC-09:00",
		"UTC-08:00", "UTC-07:00", "UTC-06:00", "UTC-05:00", "UTC-04:00",
		"UTC-03:30", "UTC-03:00", "UTC-02:00", "UTC-01:00", "UTC",
		"UTC+01:00", "UTC+02:00", "UTC+03:00", "UTC+03:30", "UTC+04:00",
		"UTC+04:30", "UTC+05:00", "UTC+05:30", "UTC+05:45", "UTC+06:00",
		"UTC+06:30", "UTC+07:00", "UTC+08:00", "UTC+08:45", "UTC+09:00",
		"UTC+09:30", "UTC+10:00", "UTC+10:30", "UTC+11:00", "UTC+12:00",
		"UTC+13:00", "UTC+14:00",
	}
)

// loadFavorites loads favorites from JSON file
func loadFavorites() error {
	file, err := os.ReadFile(favoritesFile)
	if err != nil {
		if os.IsNotExist(err) {
			favorites = Favorites{Countries: []string{}}
			return nil
		}
		return err
	}
	return json.Unmarshal(file, &favorites)
}

// saveFavorites saves favorites to JSON file
func saveFavorites() error {
	data, err := json.Marshal(favorites)
	if err != nil {
		return err
	}
	return os.WriteFile(favoritesFile, data, 0644)
}

// fetchCountries fetches country data from the API
// Modify the fetchCountries function to include HDI data
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

// isInTimeRange checks if a time falls within a range
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

// filterCountries filters countries based on region, timezone, and time range
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

// searchCountries searches countries based on query
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

// paginateCountries paginates the country list
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

// getUniqueRegions gets unique regions from countries
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

// contains checks if a string slice contains a string
func contains(slice []string, str string) bool {
	for _, s := range slice {
		if s == str {
			return true
		}
	}
	return false
}

// Add this function to get the absolute path
func getAbsolutePath(relativePath string) string {
	// Get the directory where the executable is running
	execDir, err := os.Getwd()
	if err != nil {
		log.Printf("Error getting working directory: %v", err)
		return relativePath
	}

	// Construct absolute path
	absPath := filepath.Join(execDir, relativePath)

	// Log the path for debugging
	log.Printf("Looking for GeoJSON file at: %s", absPath)

	return absPath
}

func handleTimezoneBorders(w http.ResponseWriter, r *http.Request) {
	log.Printf("Received request for timezone borders")

	// Define the directory where split GeoJSON files are stored
	dataDir := "data"
	files, err := os.ReadDir(dataDir)
	if err != nil {
		log.Printf("Error reading data directory: %v", err)
		http.Error(w, "Timezone data not available", http.StatusInternalServerError)
		return
	}

	// Collect all GeoJSON file contents
	var allFeatures []map[string]interface{}
	for _, file := range files {
		if strings.HasSuffix(file.Name(), ".geojson") {
			filePath := filepath.Join(dataDir, file.Name())
			log.Printf("Processing file: %s", filePath)

			// Open and parse each GeoJSON file
			content, err := os.ReadFile(filePath)
			if err != nil {
				log.Printf("Error reading file %s: %v", filePath, err)
				continue
			}

			var geojson map[string]interface{}
			if err := json.Unmarshal(content, &geojson); err != nil {
				log.Printf("Error parsing GeoJSON from file %s: %v", filePath, err)
				continue
			}

			// Append features from this file
			if features, ok := geojson["features"].([]interface{}); ok {
				for _, feature := range features {
					allFeatures = append(allFeatures, feature.(map[string]interface{}))
				}
			}
		}
	}

	// Prepare combined GeoJSON
	combinedGeoJSON := map[string]interface{}{
		"type":     "FeatureCollection",
		"features": allFeatures,
	}

	// Set headers and send response
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Cache-Control", "public, max-age=86400")

	if err := json.NewEncoder(w).Encode(combinedGeoJSON); err != nil {
		log.Printf("Error encoding combined GeoJSON: %v", err)
		http.Error(w, "Error processing GeoJSON", http.StatusInternalServerError)
	}
}

// Helper function to format numbers with commas
func formatNumber(n int) string {
	in := strconv.Itoa(n)
	numOfDigits := len(in)
	if n < 0 {
		numOfDigits-- // First character is the - sign (not a digit)
	}
	numOfCommas := (numOfDigits - 1) / 3

	out := make([]byte, len(in)+numOfCommas)
	if n < 0 {
		in = in[1:]
		out[0] = '-'
	}

	for i, j, k := len(in)-1, len(out)-1, 0; ; i, j = i-1, j-1 {
		out[j] = in[i]
		if i == 0 {
			break
		}
		k++
		if k == 3 {
			j--
			out[j] = ','
			k = 0
		}
	}
	return string(out)
}

func main() {
	// Load favorites
	if err := loadFavorites(); err != nil {
		log.Fatal("Error loading favorites:", err)
	}

	// Fetch initial countries data
	var err error
	allCountries, err = fetchCountries()
	if err != nil {
		log.Fatal("Error fetching countries:", err)
	}

	// Define routes
	http.HandleFunc("/", handleHome)
	http.HandleFunc("/favorites", handleFavorites)
	http.HandleFunc("/api/favorite", handleFavoriteAPI)
	http.HandleFunc("/about", handleAbout)
	http.HandleFunc("/error", handleError)
	http.HandleFunc("/map", handleMap)
	http.HandleFunc("/api/countries", handleCountriesAPI)
	http.HandleFunc("/api/timezone-borders", handleTimezoneBorders)
	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))

	// Start server
	log.Println("Server starting on http://localhost:8080")
	if err := http.ListenAndServe(":8080", nil); err != nil {
		log.Fatal(err)
	}
}

// Handlers
func handleHome(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query().Get("q")
	region := r.URL.Query().Get("region")
	timezone := r.URL.Query().Get("timezone")
	timeRange := r.URL.Query().Get("timerange")
	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	if page < 1 {
		page = 1
	}

	filteredCountries := filterCountries(allCountries, region, timezone, timeRange, w, r)
	if filteredCountries == nil {
		return
	}
	searchedCountries := searchCountries(filteredCountries, query)
	paginatedCountries, totalPages := paginateCountries(searchedCountries, page)
	regions := getUniqueRegions(allCountries)
	timeZones := getUniqueTimeZones(allCountries)

	data := PageData{
		Countries:    paginatedCountries,
		Query:        query,
		Regions:      regions,
		TimeZones:    timeZones,
		CurrentPage:  page,
		TotalPages:   totalPages,
		Region:       region,
		TimeZone:     timezone,
		TimeRange:    timeRange,
		ItemsPerPage: itemsPerPage,
	}

	tmpl := template.New("home.html").Funcs(templateFuncs)
	tmpl, err := tmpl.ParseFiles("templates/home.html")
	if err != nil {
		log.Printf("Error parsing template: %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	err = tmpl.Execute(w, data)
	if err != nil {
		log.Printf("Error executing template: %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
}

func handleFavorites(w http.ResponseWriter, r *http.Request) {
	var favoriteCountries []Country
	for _, country := range allCountries {
		if contains(favorites.Countries, country.Name) {
			country.IsFavorite = true
			favoriteCountries = append(favoriteCountries, country)
		}
	}

	data := PageData{
		Countries: favoriteCountries,
	}

	tmpl, err := template.ParseFiles("templates/favorites.html")
	if err != nil {
		log.Printf("Error parsing template: %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	err = tmpl.Execute(w, data)
	if err != nil {
		log.Printf("Error executing template: %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
}

// parseHDIData parses the CSV content and returns a map of country names to HDI data
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

func handleError(w http.ResponseWriter, r *http.Request) {
	tmpl, err := template.ParseFiles("templates/error.html")
	if err != nil {
		log.Printf("Error parsing template: %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusNotFound)
	err = tmpl.Execute(w, nil)
	if err != nil {
		log.Printf("Error executing template: %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
}

func handleFavoriteAPI(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var data struct {
		Country string `json:"country"`
		Action  string `json:"action"`
	}

	if err := json.NewDecoder(r.Body).Decode(&data); err != nil {
		http.Error(w, "Bad request", http.StatusBadRequest)
		return
	}

	switch data.Action {
	case "add":
		if !contains(favorites.Countries, data.Country) {
			favorites.Countries = append(favorites.Countries, data.Country)
		}
	case "remove":
		var newFavorites []string
		for _, c := range favorites.Countries {
			if c != data.Country {
				newFavorites = append(newFavorites, c)
			}
		}
		favorites.Countries = newFavorites
	default:
		http.Error(w, "Invalid action", http.StatusBadRequest)
		return
	}

	if err := saveFavorites(); err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func handleAbout(w http.ResponseWriter, r *http.Request) {
	tmpl, err := template.ParseFiles("templates/about.html")
	if err != nil {
		log.Printf("Error parsing template: %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	err = tmpl.Execute(w, nil)
	if err != nil {
		log.Printf("Error executing template: %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
}

func handleMap(w http.ResponseWriter, r *http.Request) {
	tmpl, err := template.ParseFiles("templates/map.html")
	if err != nil {
		log.Printf("Error parsing template: %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	err = tmpl.Execute(w, nil)
	if err != nil {
		log.Printf("Error executing template: %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
}

func handleCountriesAPI(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(allCountries)
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

// test
