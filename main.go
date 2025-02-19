package main

import (
    "log"
    "net/http"
)

var (
    allCountries []Country
    favorites    Favorites
)

func main() {
    if err := loadFavorites(); err != nil {
        log.Fatal("Error loading favorites:", err)
    }

    var err error
    allCountries, err = fetchCountries()
    if err != nil {
        log.Fatal("Error fetching countries:", err)
    }

    http.HandleFunc("/", handleHome)
    http.HandleFunc("/favorites", handleFavorites)
    http.HandleFunc("/api/favorite", handleFavoriteAPI)
    http.HandleFunc("/about", handleAbout)
    http.HandleFunc("/error", handleError)
    http.HandleFunc("/map", handleMap)
    http.HandleFunc("/api/countries", handleCountriesAPI)
    http.HandleFunc("/api/timezone-borders", handleTimezoneBorders)
    http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))

    log.Println("Server starting on http://localhost:8080")
    log.Fatal(http.ListenAndServe(":8080", nil))
}