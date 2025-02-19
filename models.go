package main

type HDIData struct {
    HDIRank     int     `json:"hdi_rank"`
    HDIValue    float64 `json:"hdi_value"`
    Category    string  `json:"category"`
    LifeExpect  float64 `json:"life_expectancy"`
    SchoolYears float64 `json:"school_years"`
    GNIPerCap   string  `json:"gni_per_capita"`
}

type Country struct {
    Name        string   `json:"name"`
    TimeZone    string   `json:"-"`
    Capital     string   `json:"capital"`
    Region      string   `json:"region"`
    Flag        string   `json:"flag"`
    TimeZones   []string `json:"timezones"`
    CurrentTime string   `json:"-"`
    IsFavorite  bool     `json:"-"`
    Population  string   `json:"population"`
    Area        float64  `json:"area"`
    Languages   []string `json:"languages"`
    Currency    string   `json:"currency"`
    CallingCode string   `json:"callingCode"`
    DrivingSide string   `json:"drivingSide"`
    Borders     []string `json:"borders"`
    HDI         HDIData  `json:"hdi"`
}

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

type Favorites struct {
    Countries []string `json:"countries"`
}