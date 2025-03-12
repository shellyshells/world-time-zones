package src

import (
	"log"
	"net/http"
)

var (
	allCountries []Country
	favorites    Favorites
)

// Run starts the application
func Run() {
	if err := loadFavorites(); err != nil {
		log.Fatal("Error loading favorites:", err)
	}

	var err error
	allCountries, err = fetchCountries()
	if err != nil {
		log.Fatal("Error fetching countries:", err)
	}

	// Register all specific routes first
	http.HandleFunc("/favorites", handleFavorites)
	http.HandleFunc("/api/favorite", handleFavoriteAPI)
	http.HandleFunc("/about", handleAbout)
	http.HandleFunc("/error", handleError)
	http.HandleFunc("/map", handleMap)
	http.HandleFunc("/api/countries", handleCountriesAPI)
	http.HandleFunc("/api/timezone-borders", handleTimezoneBorders)
	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))

	// Register the catch-all handler last
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/" {
			handleHome(w, r)
			return
		}
		handleNotFound(w, r)
	})

	log.Println("Server starting on http://localhost:8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
